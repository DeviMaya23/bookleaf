package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type categorisationLogRepository struct {
	db *gorm.DB
}

func NewCategorisationLogRepository(db *gorm.DB) *categorisationLogRepository {
	return &categorisationLogRepository{db: db}
}

func (r *categorisationLogRepository) Create(ctx context.Context, log *domain.CategorisationLog) error {
	if err := r.db.WithContext(ctx).Table("ai_categorisation_logs").Create(log).Error; err != nil {
		return fmt.Errorf("insert categorisation log: %w", err)
	}
	return nil
}

func (r *categorisationLogRepository) GetByImageID(ctx context.Context, imageID uuid.UUID) (*domain.CategorisationLog, error) {
	var log domain.CategorisationLog
	err := r.db.WithContext(ctx).
		Table("ai_categorisation_logs").
		Where("image_id = ?", imageID).
		Order("created_at DESC").
		Limit(1).
		First(&log).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get categorisation log by image id: %w", err)
	}
	return &log, nil
}

func (r *categorisationLogRepository) CountByUserAndMonth(ctx context.Context, userID uuid.UUID, year, month int) (int, error) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	var count int64
	err := r.db.WithContext(ctx).
		Table("ai_categorisation_logs").
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, start, end).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("count categorisation logs by user and month: %w", err)
	}
	return int(count), nil
}
