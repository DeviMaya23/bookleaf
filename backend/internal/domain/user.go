package domain

import (
	"time"

	"gorm.io/gorm"
)

type AccountState string

const (
	AccountStateActive          AccountState = "active"
	AccountStatePendingDeletion AccountState = "pending_deletion"
	AccountStatePurged          AccountState = "purged"
)

type User struct {
	ID                      string         `gorm:"type:text;primaryKey"`
	VisionEnabled           bool           `gorm:"column:vision_enabled;default:false"`
	AICategorisationEnabled bool           `gorm:"column:ai_categorisation_enabled;default:false"`
	AccountState            AccountState   `gorm:"column:account_state;default:active"`
	PurgedAt                *time.Time     `gorm:"column:purged_at"`
	FolderIconsEnabled      bool           `gorm:"column:folder_icons_enabled;default:true"`
	CreatedAt               time.Time      `gorm:"column:created_at"`
	UpdatedAt               time.Time      `gorm:"column:updated_at"`
	DeletedAt               gorm.DeletedAt `gorm:"column:deleted_at;index"`
}
