package usecase

import "github.com/google/uuid"

type ThumbnailUploadArgs struct {
	ImageID      uuid.UUID `json:"image_id"`
	UserID       string    `json:"user_id"`
	R2Path       string    `json:"r2_path"`
	ThumbnailKey string    `json:"thumbnail_key"`
}

func (ThumbnailUploadArgs) Kind() string        { return "thumbnail_upload" }
func (ThumbnailUploadArgs) MaxAttempts() int    { return 5 }

type VisionArgs struct {
	ImageID uuid.UUID `json:"image_id"`
	UserID  string    `json:"user_id"`
	R2Path  string    `json:"r2_path"`
}

func (VisionArgs) Kind() string     { return "vision_labelling" }
func (VisionArgs) MaxAttempts() int { return 3 }
