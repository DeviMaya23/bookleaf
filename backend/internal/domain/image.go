package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ImageFolder struct {
	ImageID  uuid.UUID `gorm:"primaryKey;column:image_id"`
	FolderID uuid.UUID `gorm:"primaryKey;column:folder_id"`
	Position string    `gorm:"column:position;not null;default:''"`
	Folder   Folder    `gorm:"foreignKey:FolderID;references:ID"`
}

type Image struct {
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey"`
	UserID        uuid.UUID       `gorm:"column:user_id;type:uuid;not null;index"`
	Title         string          `gorm:"column:title;not null"`
	Description   *string         `gorm:"column:description"`
	SourceURL     *string         `gorm:"column:source_url"`
	R2Path        string          `gorm:"column:r2_path;not null"`
	ThumbnailPath *string         `gorm:"column:thumbnail_path"`
	MIMEType      string          `gorm:"column:mime_type;not null"`
	Width         *int            `gorm:"column:width"`
	Height        *int            `gorm:"column:height"`
	FileSize      *int64          `gorm:"column:file_size"`
	PHash         *string         `gorm:"column:phash;type:bit(64)"`
	AILabels      json.RawMessage `gorm:"column:ai_labels;type:jsonb"`
	CreatedAt     time.Time       `gorm:"column:created_at"`
	UpdatedAt     time.Time       `gorm:"column:updated_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"column:deleted_at;index"`

	User         User          `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	ImageFolders []ImageFolder `gorm:"foreignKey:ImageID"`
	Tags         []Tag         `gorm:"many2many:image_tags;foreignKey:ID;joinForeignKey:ImageID;References:ID;joinReferences:TagID"`
}

func (i *Image) BeforeCreate(*gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}

	return nil
}
