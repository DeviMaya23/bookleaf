package usecase

import "context"

// KindeClient manages a user's identity in Kinde.
type KindeClient interface {
	// DeleteUser permanently deletes the given user from Kinde. A response
	// indicating the user no longer exists is treated as success.
	DeleteUser(ctx context.Context, kindeUserID string) error
	// DeleteUserSessions revokes all active sessions for the given user. A
	// response indicating the user has no sessions or does not exist is treated
	// as success.
	DeleteUserSessions(ctx context.Context, kindeUserID string) error
}
