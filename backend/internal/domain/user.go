package domain

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                   string         `gorm:"type:text;primaryKey"`
	VisionEnabled              bool           `gorm:"column:vision_enabled;default:false"`
	AICategorisationEnabled    bool           `gorm:"column:ai_categorisation_enabled;default:false"`
	PendingKindeDeletion       bool           `gorm:"column:pending_kinde_deletion;default:false"`
	FolderIconsEnabled   bool           `gorm:"column:folder_icons_enabled;default:true"`
	CreatedAt            time.Time      `gorm:"column:created_at"`
	UpdatedAt            time.Time      `gorm:"column:updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"column:deleted_at;index"`
}
