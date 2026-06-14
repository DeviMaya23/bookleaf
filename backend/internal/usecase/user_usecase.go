package usecase

import (
	"context"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type userUsecase struct {
	userRepo UserRepository
	tel      *observability.Telemetry
}

func NewUserUsecase(userRepo UserRepository, tel *observability.Telemetry) *userUsecase {
	return &userUsecase{
		userRepo: userRepo,
		tel:      tel,
	}
}

func (u *userUsecase) GetOrProvision(ctx context.Context, kindeID string) (*domain.User, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.GetOrProvision")
	defer span.End()

	user, err := u.userRepo.GetByID(ctx, kindeID)
	if err == nil {
		return user, nil
	}

	createdUser, err := u.userRepo.GetOrCreate(ctx, kindeID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	observability.LoggerFromContext(ctx, u.tel.Logger).Info(
		"user persisted",
		zap.String("event", "user.created"),
		zap.String("user_id", createdUser.ID),
	)
	return createdUser, nil
}

func (u *userUsecase) GetByID(ctx context.Context, kindeID string) (*domain.User, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.GetByID")
	defer span.End()

	user, err := u.userRepo.GetByID(ctx, kindeID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return user, nil
}

func (u *userUsecase) UpdateVisionEnabled(ctx context.Context, id string, enabled bool) (*domain.User, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.UpdateVisionEnabled")
	defer span.End()

	fields := map[string]any{"vision_enabled": enabled}

	user, err := u.userRepo.Update(ctx, id, fields)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return user, nil
}
