package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserRepository struct {
	user *domain.User
	err  error
}

func (m *mockUserRepository) GetOrCreate(context.Context, string) (*domain.User, error) {
	return m.user, m.err
}

func (m *mockUserRepository) GetByID(context.Context, string) (*domain.User, error) {
	return m.user, m.err
}

func TestUserUsecase_GetOrProvision(t *testing.T) {
	tests := []struct {
		name    string
		repo    *mockUserRepository
		wantID  string
		wantErr bool
	}{
		{
			name:   "returns provisioned user",
			repo:   &mockUserRepository{user: &domain.User{ID: "kp_abc123"}},
			wantID: "kp_abc123",
		},
		{
			name:    "propagates repository error",
			repo:    &mockUserRepository{err: errors.New("db error")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewUserUsecase(tt.repo)

			user, err := uc.GetOrProvision(context.Background(), "kp_abc123")

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, user.ID)
		})
	}
}

func TestUserUsecase_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		repo    *mockUserRepository
		wantID  string
		wantErr bool
	}{
		{
			name:   "returns user by id",
			repo:   &mockUserRepository{user: &domain.User{ID: "kp_abc123"}},
			wantID: "kp_abc123",
		},
		{
			name:    "propagates repository error",
			repo:    &mockUserRepository{err: errors.New("db error")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewUserUsecase(tt.repo)

			user, err := uc.GetByID(context.Background(), "kp_abc123")

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, user.ID)
		})
	}
}
