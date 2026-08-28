package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PendingUpload struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID  `gorm:"column:user_id;type:uuid;not null"`
	Title       string     `gorm:"column:title;not null"`
	Description *string    `gorm:"column:description"`
	SourceURL   *string    `gorm:"column:source_url"`
	R2Path      string     `gorm:"column:r2_path;not null"`
	MIMEType    string     `gorm:"column:mime_type;not null"`
	FolderID    *uuid.UUID `gorm:"column:folder_id"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
}

func (p *PendingUpload) BeforeCreate(*gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}

	return nil
}
