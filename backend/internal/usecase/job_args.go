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

type CategoriseImageArgs struct {
	ImageID uuid.UUID `json:"image_id"`
	UserID  string    `json:"user_id"`
}

func (CategoriseImageArgs) Kind() string     { return "categorise_image" }
func (CategoriseImageArgs) MaxAttempts() int { return 3 }

type BookletUserDeletionArgs struct {
	UserID string `json:"user_id"`
}

func (BookletUserDeletionArgs) Kind() string     { return "booklet_user_deletion" }
func (BookletUserDeletionArgs) MaxAttempts() int { return 10 }

type AccountWipeArgs struct {
	UserID string `json:"user_id"`
}

func (AccountWipeArgs) Kind() string     { return "account_wipe" }
func (AccountWipeArgs) MaxAttempts() int { return 5 }

type AccountWipeReconcileArgs struct{}

func (AccountWipeReconcileArgs) Kind() string     { return "account_wipe_reconcile" }
func (AccountWipeReconcileArgs) MaxAttempts() int { return 3 }

type PurgedAccountSweepArgs struct{}

func (PurgedAccountSweepArgs) Kind() string     { return "purged_account_sweep" }
func (PurgedAccountSweepArgs) MaxAttempts() int { return 3 }
