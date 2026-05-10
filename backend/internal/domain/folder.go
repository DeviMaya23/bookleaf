package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Folder struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID      string     `gorm:"column:user_id;type:text;not null;index"`
	ParentID    *uuid.UUID `gorm:"column:parent_id;type:uuid;index"`
	Name        string     `gorm:"column:name;not null"`
	Description *string    `gorm:"column:description"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`

	User   User    `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Parent *Folder `gorm:"foreignKey:ParentID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

func (f *Folder) BeforeCreate(*gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}

	return nil
}
