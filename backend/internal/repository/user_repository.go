package repository

import (
	"context"
	"fmt"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/usecase"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) usecase.UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) GetOrCreate(ctx context.Context, id string) (*domain.User, error) {
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&domain.User{ID: id}).
		Error
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("select user: %w", err)
	}

	return &user, nil
}
