package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Tag struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"column:user_id;type:uuid;not null;index"`
	Name      string    `gorm:"column:name;not null"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`

	User   User    `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Images []Image `gorm:"many2many:image_tags;foreignKey:ID;joinForeignKey:TagID;References:ID;joinReferences:ImageID"`
}

func (t *Tag) BeforeCreate(*gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
