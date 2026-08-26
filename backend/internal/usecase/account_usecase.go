package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

const purgedAccountTTL = 25 * time.Hour

type BookletClient interface {
	DeleteUser(ctx context.Context, userID string) error
}

type accountJobEnqueuer interface {
	EnqueueAccountWipe(ctx context.Context, userID string) error
	EnqueueAccountWipeUnique(ctx context.Context, userID string) error
	EnqueueBookletUserDeletion(ctx context.Context, userID string) error
	EnqueueR2Delete(ctx context.Context, r2Path string, thumbnailPath *string) error
}

type accountUsecase struct {
	accountRepo   AccountRepository
	userRepo      UserRepository
	kinde         KindeClient
	enqueuer      accountJobEnqueuer
	tel           *observability.Telemetry
	bookletClient BookletClient
}

func NewAccountUsecase(
	accountRepo AccountRepository,
	userRepo UserRepository,
	kinde KindeClient,
	enqueuer accountJobEnqueuer,
	tel *observability.Telemetry,
	bookletClient BookletClient,
) *accountUsecase {
	return &accountUsecase{
		accountRepo:   accountRepo,
		userRepo:      userRepo,
		kinde:         kinde,
		enqueuer:      enqueuer,
		tel:           tel,
		bookletClient: bookletClient,
	}
}

func (u *accountUsecase) MarkForDeletion(ctx context.Context, userID string) error {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.MarkForDeletion")
	defer span.End()

	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		if !errors.Is(err, ErrUserNotFound) {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("get user: %w", err)
		}
		// User has no Bookleaf row — enqueue wipe job for Kinde-only cleanup.
		if err := u.enqueuer.EnqueueAccountWipe(ctx, userID); err != nil {
			observability.LoggerFromContext(ctx, u.tel.Logger).Warn("failed to enqueue account wipe job for unprovisioned user",
				zap.String("event", "account.wipe.enqueue_failed_unprovisioned"),
				zap.String("user_id", userID),
				zap.Error(err),
			)
		}
		return nil
	}

	if user.AccountState != domain.AccountStateActive {
		return nil
	}

	if err := u.userRepo.SetAccountState(ctx, userID, domain.AccountStatePendingDeletion); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("set account state pending deletion: %w", err)
	}

	if err := u.enqueuer.EnqueueAccountWipe(ctx, userID); err != nil {
		observability.LoggerFromContext(ctx, u.tel.Logger).Warn("failed to enqueue account wipe job",
			zap.String("event", "account.wipe.enqueue_failed"),
			zap.String("user_id", userID),
			zap.Error(err),
		)
	}

	return nil
}

func (u *accountUsecase) WipeAccount(ctx context.Context, userID string) error {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.WipeAccount")
	defer span.End()

	logger := observability.LoggerFromContext(ctx, u.tel.Logger)
	if err := u.kinde.DeleteUserSessions(ctx, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("delete user sessions: %w", err)
	}

	if err := u.kinde.DeleteUser(ctx, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("delete kinde user: %w", err)
	}

	type r2Item struct {
		path      string
		thumbnail *string
	}
	var r2Deletions []r2Item

	err := u.accountRepo.Transaction(ctx, func(repos AccountRepos) error {
		if err := repos.Folders.ClearAllParents(ctx, userID); err != nil {
			return fmt.Errorf("clear folder parents: %w", err)
		}

		images, err := repos.Images.ListAllByUserID(ctx, userID)
		if err != nil {
			return fmt.Errorf("list images: %w", err)
		}
		for _, img := range images {
			r2Deletions = append(r2Deletions, r2Item{path: img.R2Path, thumbnail: img.ThumbnailPath})
		}
		if err := repos.Images.HardDeleteAllByUserID(ctx, userID); err != nil {
			return fmt.Errorf("hard delete images: %w", err)
		}

		if err := repos.Folders.DeleteAllByUserID(ctx, userID); err != nil {
			return fmt.Errorf("delete folders: %w", err)
		}

		if err := repos.Tags.DeleteAllByUserID(ctx, userID); err != nil {
			return fmt.Errorf("delete tags: %w", err)
		}

		pendingUploads, err := repos.PendingUploads.ListAllByUserID(ctx, userID)
		if err != nil {
			return fmt.Errorf("list pending uploads: %w", err)
		}
		for _, p := range pendingUploads {
			r2Deletions = append(r2Deletions, r2Item{path: p.R2Path})
		}
		if err := repos.PendingUploads.DeleteAllByUserID(ctx, userID); err != nil {
			return fmt.Errorf("delete pending uploads: %w", err)
		}

		return nil
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("wipe account data: %w", err)
	}

	for _, item := range r2Deletions {
		if err := u.enqueuer.EnqueueR2Delete(ctx, item.path, item.thumbnail); err != nil {
			logger.Warn("failed to enqueue r2 delete job during account wipe",
				zap.String("event", "account.r2_delete.enqueue_failed"),
				zap.String("user_id", userID),
				zap.String("r2_path", item.path),
				zap.Error(err),
			)
		}
	}

	if err := u.userRepo.MarkPurged(ctx, userID, time.Now()); err != nil && !errors.Is(err, ErrUserNotFound) {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("mark purged: %w", err)
	}

	if u.bookletClient != nil {
		if err := u.enqueuer.EnqueueBookletUserDeletion(ctx, userID); err != nil {
			logger.Warn("failed to enqueue booklet user deletion job",
				zap.String("event", "account.booklet_deletion.enqueue_failed"),
				zap.String("user_id", userID),
				zap.Error(err),
			)
		}
	}

	logger.Info("account wiped",
		zap.String("event", "account.wiped"),
		zap.String("user_id", userID),
	)

	return nil
}

func (u *accountUsecase) ReconcilePendingDeletions(ctx context.Context) error {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.ReconcilePendingDeletions")
	defer span.End()

	users, err := u.userRepo.ListByAccountState(ctx, domain.AccountStatePendingDeletion)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("list pending deletion users: %w", err)
	}

	logger := observability.LoggerFromContext(ctx, u.tel.Logger)
	for _, user := range users {
		if err := u.enqueuer.EnqueueAccountWipeUnique(ctx, user.ID); err != nil {
			logger.Warn("failed to enqueue reconcile account wipe job",
				zap.String("event", "account.wipe.reconcile_enqueue_failed"),
				zap.String("user_id", user.ID),
				zap.Error(err),
			)
		}
	}

	if len(users) > 0 {
		logger.Info("reconciled pending deletions",
			zap.String("event", "account.deletion.reconciled"),
			zap.Int("count", len(users)),
		)
	}

	return nil
}

func (u *accountUsecase) SweepPurgedAccounts(ctx context.Context) error {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.SweepPurgedAccounts")
	defer span.End()

	cutoff := time.Now().Add(-purgedAccountTTL)
	users, err := u.userRepo.ListPurgedBefore(ctx, cutoff)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("list purged users: %w", err)
	}

	logger := observability.LoggerFromContext(ctx, u.tel.Logger)
	for _, user := range users {
		if err := u.userRepo.HardDelete(ctx, user.ID); err != nil {
			logger.Warn("failed to hard delete purged user",
				zap.String("event", "account.sweep.hard_delete_failed"),
				zap.String("user_id", user.ID),
				zap.Error(err),
			)
		}
	}

	return nil
}

func (u *accountUsecase) ProcessBookletUserDeletion(ctx context.Context, userID string) error {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.ProcessBookletUserDeletion")
	defer span.End()

	if err := u.bookletClient.DeleteUser(ctx, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("delete booklet user: %w", err)
	}

	observability.LoggerFromContext(ctx, u.tel.Logger).Info("booklet account deleted",
		zap.String("event", "account.booklet_account_deleted"),
		zap.String("user_id", userID),
	)

	return nil
}
