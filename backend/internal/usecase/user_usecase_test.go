package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/devi/bookleaf/internal/domain"
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

func TestUserUsecase_GetOrProvision_HappyPath(t *testing.T) {
	repo := &mockUserRepository{
		user: &domain.User{ID: "kp_abc123", VisionEnabled: false},
	}
	uc := NewUserUsecase(repo)

	user, err := uc.GetOrProvision(context.Background(), "kp_abc123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if user.ID != "kp_abc123" {
		t.Fatalf("expected user ID kp_abc123, got: %s", user.ID)
	}
}

func TestUserUsecase_GetOrProvision_ErrorPath(t *testing.T) {
	repo := &mockUserRepository{
		err: errors.New("db error"),
	}
	uc := NewUserUsecase(repo)

	_, err := uc.GetOrProvision(context.Background(), "kp_abc123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUserUsecase_GetByID_HappyPath(t *testing.T) {
	repo := &mockUserRepository{
		user: &domain.User{ID: "kp_abc123", VisionEnabled: false},
	}
	uc := NewUserUsecase(repo)

	user, err := uc.GetByID(context.Background(), "kp_abc123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if user.ID != "kp_abc123" {
		t.Fatalf("expected user ID kp_abc123, got: %s", user.ID)
	}
}

func TestUserUsecase_GetByID_ErrorPath(t *testing.T) {
	repo := &mockUserRepository{
		err: errors.New("db error"),
	}
	uc := NewUserUsecase(repo)

	_, err := uc.GetByID(context.Background(), "kp_abc123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
