package usecase

import "github.com/google/uuid"

type VisionArgs struct {
	ImageID uuid.UUID `json:"image_id"`
	UserID  string    `json:"user_id"`
	R2Path  string    `json:"r2_path"`
}

func (VisionArgs) Kind() string     { return "vision_labelling" }
func (VisionArgs) MaxAttempts() int { return 3 }
