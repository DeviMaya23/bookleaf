package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FolderShare struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	FolderID  uuid.UUID `gorm:"column:folder_id;type:uuid;not null;uniqueIndex"`
	Token     string    `gorm:"column:token;type:text;not null;uniqueIndex"`
	CreatedAt time.Time `gorm:"column:created_at"`

	Folder Folder `gorm:"foreignKey:FolderID;references:ID;constraint:OnDelete:CASCADE"`
}

func (s *FolderShare) BeforeCreate(*gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}

	return nil
}
