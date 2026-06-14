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

func (r *userRepository) MarkPendingKindeDeletion(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", id).
		Update("pending_kinde_deletion", true)
	if result.Error != nil {
		return fmt.Errorf("mark pending kinde deletion: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("mark pending kinde deletion: %w", gorm.ErrRecordNotFound)
	}
	return nil
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

func (r *userRepository) ListPendingKindeDeletion(ctx context.Context) ([]*domain.User, error) {
	var users []*domain.User
	if err := r.db.WithContext(ctx).
		Where("pending_kinde_deletion = ?", true).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list pending kinde deletion users: %w", err)
	}
	return users, nil
}
