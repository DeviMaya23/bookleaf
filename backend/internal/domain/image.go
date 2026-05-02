package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Image struct {
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey"`
	UserID        string          `gorm:"column:user_id;type:text;not null;index"`
	FolderID      *uuid.UUID      `gorm:"column:folder_id;type:uuid;index"`
	Title         string          `gorm:"column:title;not null"`
	SourceURL     *string         `gorm:"column:source_url"`
	R2Path        string          `gorm:"column:r2_path;not null"`
	ThumbnailPath *string         `gorm:"column:thumbnail_path"`
	MIMEType      string          `gorm:"column:mime_type;not null"`
	AILabels      json.RawMessage `gorm:"column:ai_labels;type:jsonb"`
	CreatedAt     time.Time       `gorm:"column:created_at"`
	UpdatedAt     time.Time       `gorm:"column:updated_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"column:deleted_at;index"`

	User   User    `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Folder *Folder `gorm:"foreignKey:FolderID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

func (i *Image) BeforeCreate(*gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}

	return nil
}
