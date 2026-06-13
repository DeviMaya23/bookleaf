package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- fake ---

type fakeUserRepo struct {
	users       map[string]*domain.User
	getOrCreate error
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[string]*domain.User)}
}

func (f *fakeUserRepo) GetByID(_ context.Context, id string) (*domain.User, error) {
	user, ok := f.users[id]
	if !ok {
		return nil, errors.New("record not found")
	}
	return user, nil
}

func (f *fakeUserRepo) GetOrCreate(_ context.Context, id string) (*domain.User, error) {
	if f.getOrCreate != nil {
		return nil, f.getOrCreate
	}
	user := &domain.User{ID: id}
	f.users[id] = user
	return user, nil
}

func (f *fakeUserRepo) MarkPendingKindeDeletion(_ context.Context, id string) error {
	user, ok := f.users[id]
	if !ok {
		return errors.New("record not found")
	}
	user.PendingKindeDeletion = true
	return nil
}

func (f *fakeUserRepo) HardDelete(_ context.Context, id string) error {
	if _, ok := f.users[id]; !ok {
		return errors.New("record not found")
	}
	delete(f.users, id)
	return nil
}

func (f *fakeUserRepo) ListPendingKindeDeletion(_ context.Context) ([]*domain.User, error) {
	var users []*domain.User
	for _, u := range f.users {
		if u.PendingKindeDeletion {
			users = append(users, u)
		}
	}
	return users, nil
}

func newTestUserUsecase(repo UserRepository) *userUsecase {
	return NewUserUsecase(repo, observability.NewTelemetry(nil, nil, nil))
}

// --- tests ---

func TestUserUsecase_GetOrProvision_ExistingUser(t *testing.T) {
	repo := newFakeUserRepo()
	repo.users["kp_abc123"] = &domain.User{ID: "kp_abc123"}
	uc := newTestUserUsecase(repo)

	user, err := uc.GetOrProvision(context.Background(), "kp_abc123")

	require.NoError(t, err)
	assert.Equal(t, "kp_abc123", user.ID)
}

func TestUserUsecase_GetOrProvision_NewUser(t *testing.T) {
	repo := newFakeUserRepo()
	uc := newTestUserUsecase(repo)

	user, err := uc.GetOrProvision(context.Background(), "kp_new123")

	require.NoError(t, err)
	assert.Equal(t, "kp_new123", user.ID)
	_, exists := repo.users["kp_new123"]
	assert.True(t, exists)
}

func TestUserUsecase_GetOrProvision_ProvisionFails(t *testing.T) {
	repo := newFakeUserRepo()
	repo.getOrCreate = errors.New("db error")
	uc := newTestUserUsecase(repo)

	_, err := uc.GetOrProvision(context.Background(), "kp_new123")

	require.Error(t, err)
}
