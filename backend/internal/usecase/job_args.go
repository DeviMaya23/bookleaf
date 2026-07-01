package usecase

import "github.com/google/uuid"

type VisionArgs struct {
	ImageID uuid.UUID `json:"image_id"`
	UserID  string    `json:"user_id"`
}

func (VisionArgs) Kind() string     { return "vision_labelling" }
func (VisionArgs) MaxAttempts() int { return 3 }

type R2DeleteArgs struct {
	R2Path        string  `json:"r2_path"`
	ThumbnailPath *string `json:"thumbnail_path"`
}

func (R2DeleteArgs) Kind() string     { return "r2_delete" }
func (R2DeleteArgs) MaxAttempts() int { return 5 }

type AccountKindeDeletionArgs struct {
	UserID string `json:"user_id"`
}

func (AccountKindeDeletionArgs) Kind() string     { return "account_kinde_deletion" }
func (AccountKindeDeletionArgs) MaxAttempts() int { return 5 }

type CategoriseImageArgs struct {
	ImageID uuid.UUID `json:"image_id"`
	UserID  string    `json:"user_id"`
}

func (CategoriseImageArgs) Kind() string     { return "categorise_image" }
func (CategoriseImageArgs) MaxAttempts() int { return 3 }
