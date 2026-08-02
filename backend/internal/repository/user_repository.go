package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, usecase.ErrUserNotFound
		}
		return nil, fmt.Errorf("select user: %w", err)
	}

	return &user, nil
}

func (r *userRepository) SetAccountState(ctx context.Context, id string, state domain.AccountState) error {
	result := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", id).
		Update("account_state", state)
	if result.Error != nil {
		return fmt.Errorf("set account state: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("set account state: %w", usecase.ErrUserNotFound)
	}
	return nil
}

func (r *userRepository) MarkPurged(ctx context.Context, id string, purgedAt time.Time) error {
	result := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", id).
		Updates(map[string]any{"account_state": domain.AccountStatePurged, "purged_at": purgedAt})
	if result.Error != nil {
		return fmt.Errorf("mark purged: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("mark purged: %w", usecase.ErrUserNotFound)
	}
	return nil
}

func (r *userRepository) ListByAccountState(ctx context.Context, state domain.AccountState) ([]*domain.User, error) {
	var users []*domain.User
	if err := r.db.WithContext(ctx).
		Where("account_state = ?", state).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list users by account state: %w", err)
	}
	return users, nil
}

func (r *userRepository) ListPurgedBefore(ctx context.Context, threshold time.Time) ([]*domain.User, error) {
	var users []*domain.User
	if err := r.db.WithContext(ctx).
		Where("account_state = ? AND purged_at < ?", domain.AccountStatePurged, threshold).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list purged users before threshold: %w", err)
	}
	return users, nil
}

func (r *userRepository) HardDelete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		Delete(&domain.User{})
	if result.Error != nil {
		return fmt.Errorf("hard delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("hard delete user: %w", gorm.ErrRecordNotFound)
	}
	return nil
}

func (r *userRepository) Update(ctx context.Context, id string, fields map[string]any) (*domain.User, error) {
	result := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, fmt.Errorf("update user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("update user: %w", gorm.ErrRecordNotFound)
	}

	return r.GetByID(ctx, id)
}
