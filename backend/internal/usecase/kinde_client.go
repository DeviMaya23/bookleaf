package usecase

import "context"

// KindeClient manages a user's identity in Kinde.
type KindeClient interface {
	// DeleteUser permanently deletes the given user from Kinde. A response
	// indicating the user no longer exists is treated as success.
	DeleteUser(ctx context.Context, kindeUserID string) error
}
