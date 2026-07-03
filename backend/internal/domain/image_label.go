package domain

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ImageLabel struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey"`
	ImageID uuid.UUID `gorm:"column:image_id;not null;index"`
	Label   string    `gorm:"column:label;not null"`
	Score   float32   `gorm:"column:score;not null"`
}

func (il *ImageLabel) BeforeCreate(*gorm.DB) error {
	if il.ID == uuid.Nil {
		il.ID = uuid.New()
	}

	return nil
}
