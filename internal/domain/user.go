package domain

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        string         `gorm:"type:text;primaryKey"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}
