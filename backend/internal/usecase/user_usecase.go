package usecase

import (
	"context"

	"github.com/devi/bookleaf/internal/domain"
)

type UserUsecase interface {
	GetOrProvision(ctx context.Context, kindeID string) (*domain.User, error)
	GetByID(ctx context.Context, kindeID string) (*domain.User, error)
}

type userUsecase struct {
	userRepo UserRepository
}

func NewUserUsecase(userRepo UserRepository) UserUsecase {
	return &userUsecase{
		userRepo: userRepo,
	}
}

func (u *userUsecase) GetOrProvision(ctx context.Context, kindeID string) (*domain.User, error) {
	return u.userRepo.GetOrCreate(ctx, kindeID)
}

func (u *userUsecase) GetByID(ctx context.Context, kindeID string) (*domain.User, error) {
	return u.userRepo.GetByID(ctx, kindeID)
}
