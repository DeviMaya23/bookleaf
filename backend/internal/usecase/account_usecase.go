package usecase

import (
	"context"
	"fmt"

	"github.com/devi/bookleaf/internal/platform/observability"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type accountUsecase struct {
	accountRepo AccountRepository
	userRepo    UserRepository
	kinde       KindeClient
	enqueuer    JobEnqueuer
	tel         *observability.Telemetry
}

func NewAccountUsecase(
	accountRepo AccountRepository,
	userRepo UserRepository,
	kinde KindeClient,
	enqueuer JobEnqueuer,
	tel *observability.Telemetry,
) *accountUsecase {
	return &accountUsecase{
		accountRepo: accountRepo,
		userRepo:    userRepo,
		kinde:       kinde,
		enqueuer:    enqueuer,
		tel:         tel,
	}
}

func (u *accountUsecase) DeleteAccount(ctx context.Context, userID string) error {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.DeleteAccount")
	defer span.End()

	logger := observability.LoggerFromContext(ctx, u.tel.Logger)

	var r2Deletions []R2DeleteArgs

	err := u.accountRepo.Transaction(ctx, func(repos AccountRepos) error {
		if err := repos.Folders.ClearAllParents(ctx, userID); err != nil {
			return fmt.Errorf("clear folder parents: %w", err)
		}

		images, err := repos.Images.ListAllByUserID(ctx, userID)
		if err != nil {
			return fmt.Errorf("list images: %w", err)
		}
		for _, img := range images {
			r2Deletions = append(r2Deletions, R2DeleteArgs{R2Path: img.R2Path, ThumbnailPath: img.ThumbnailPath})
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
			r2Deletions = append(r2Deletions, R2DeleteArgs{R2Path: p.R2Path})
		}
		if err := repos.PendingUploads.DeleteAllByUserID(ctx, userID); err != nil {
			return fmt.Errorf("delete pending uploads: %w", err)
		}

		if err := repos.Users.MarkPendingKindeDeletion(ctx, userID); err != nil {
			return fmt.Errorf("mark pending kinde deletion: %w", err)
		}

		return nil
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("wipe account data: %w", err)
	}

	for _, args := range r2Deletions {
		if err := u.enqueuer.Insert(ctx, args); err != nil {
			logger.Warn("failed to enqueue r2 delete job during account deletion",
				zap.String("event", "account.r2_delete.enqueue_failed"),
				zap.String("user_id", userID),
				zap.String("r2_path", args.R2Path),
				zap.Error(err),
			)
		}
	}

	if err := u.enqueuer.Insert(ctx, AccountKindeDeletionArgs{UserID: userID}); err != nil {
		logger.Warn("failed to enqueue kinde deletion job",
			zap.String("event", "account.kinde_deletion.enqueue_failed"),
			zap.String("user_id", userID),
			zap.Error(err),
		)
	}

	logger.Info("account data wiped",
		zap.String("event", "account.wiped"),
		zap.String("user_id", userID),
	)

	return nil
}

func (u *accountUsecase) ProcessAccountKindeDeletion(ctx context.Context, userID string) error {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.ProcessAccountKindeDeletion")
	defer span.End()

	if err := u.kinde.DeleteUser(ctx, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("delete kinde user: %w", err)
	}

	if err := u.userRepo.HardDelete(ctx, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("hard delete user: %w", err)
	}

	observability.LoggerFromContext(ctx, u.tel.Logger).Info("account kinde identity deleted",
		zap.String("event", "account.kinde_identity_deleted"),
		zap.String("user_id", userID),
	)

	return nil
}

func (u *accountUsecase) ReconcilePendingKindeDeletions(ctx context.Context) error {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.ReconcilePendingKindeDeletions")
	defer span.End()

	users, err := u.userRepo.ListPendingKindeDeletion(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("list pending kinde deletions: %w", err)
	}

	logger := observability.LoggerFromContext(ctx, u.tel.Logger)
	for _, user := range users {
		if err := u.enqueuer.Insert(ctx, AccountKindeDeletionArgs{UserID: user.ID}); err != nil {
			logger.Warn("failed to enqueue reconciled kinde deletion job",
				zap.String("event", "account.kinde_deletion.reconcile_enqueue_failed"),
				zap.String("user_id", user.ID),
				zap.Error(err),
			)
		}
	}

	if len(users) > 0 {
		logger.Info("reconciled pending kinde deletions",
			zap.String("event", "account.kinde_deletion.reconciled"),
			zap.Int("count", len(users)),
		)
	}

	return nil
}
