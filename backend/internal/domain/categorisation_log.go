package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategorisationLog struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ImageID       *uuid.UUID `gorm:"column:image_id;type:uuid"`
	UserID        string     `gorm:"column:user_id;not null"`
	Reasoning     string     `gorm:"column:reasoning;not null"`
	FolderID      *uuid.UUID `gorm:"column:folder_id;type:uuid"`
	NewFolderName *string    `gorm:"column:new_folder_name"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
}

func (c *CategorisationLog) BeforeCreate(*gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
