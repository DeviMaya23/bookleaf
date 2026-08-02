package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- fake ---

type fakeUserRepo struct {
	users       map[string]*domain.User
	getOrCreate error
	updateErr   error
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[string]*domain.User)}
}

func (f *fakeUserRepo) GetByID(_ context.Context, id string) (*domain.User, error) {
	user, ok := f.users[id]
	if !ok {
		return nil, ErrUserNotFound
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

func (f *fakeUserRepo) SetAccountState(_ context.Context, id string, state domain.AccountState) error {
	user, ok := f.users[id]
	if !ok {
		return ErrUserNotFound
	}
	user.AccountState = state
	return nil
}

func (f *fakeUserRepo) MarkPurged(_ context.Context, id string, purgedAt time.Time) error {
	user, ok := f.users[id]
	if !ok {
		return ErrUserNotFound
	}
	user.AccountState = domain.AccountStatePurged
	user.PurgedAt = &purgedAt
	return nil
}

func (f *fakeUserRepo) ListByAccountState(_ context.Context, state domain.AccountState) ([]*domain.User, error) {
	var users []*domain.User
	for _, u := range f.users {
		if u.AccountState == state {
			users = append(users, u)
		}
	}
	return users, nil
}

func (f *fakeUserRepo) ListPurgedBefore(_ context.Context, threshold time.Time) ([]*domain.User, error) {
	var users []*domain.User
	for _, u := range f.users {
		if u.AccountState == domain.AccountStatePurged && u.PurgedAt != nil && u.PurgedAt.Before(threshold) {
			users = append(users, u)
		}
	}
	return users, nil
}

func (f *fakeUserRepo) HardDelete(_ context.Context, id string) error {
	if _, ok := f.users[id]; !ok {
		return errors.New("record not found")
	}
	delete(f.users, id)
	return nil
}

func (f *fakeUserRepo) Update(_ context.Context, id string, fields map[string]any) (*domain.User, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	user, ok := f.users[id]
	if !ok {
		return nil, errors.New("record not found")
	}
	if enabled, ok := fields["vision_enabled"].(bool); ok {
		user.VisionEnabled = enabled
	}
	if enabled, ok := fields["folder_icons_enabled"].(bool); ok {
		user.FolderIconsEnabled = enabled
	}
	if enabled, ok := fields["ai_categorisation_enabled"].(bool); ok {
		user.AICategorisationEnabled = enabled
	}
	return user, nil
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

func TestUserUsecase_UpdatePreferences_VisionEnabledSuccess(t *testing.T) {
	repo := newFakeUserRepo()
	repo.users["kp_abc123"] = &domain.User{ID: "kp_abc123", VisionEnabled: false}
	uc := newTestUserUsecase(repo)
	enabled := true

	user, err := uc.UpdatePreferences(context.Background(), "kp_abc123", UpdateUserPreferencesParams{VisionEnabled: &enabled})

	require.NoError(t, err)
	assert.True(t, user.VisionEnabled)
}

func TestUserUsecase_UpdatePreferences_FolderIconsEnabledSuccess(t *testing.T) {
	repo := newFakeUserRepo()
	repo.users["kp_abc123"] = &domain.User{ID: "kp_abc123", FolderIconsEnabled: true}
	uc := newTestUserUsecase(repo)
	disabled := false

	user, err := uc.UpdatePreferences(context.Background(), "kp_abc123", UpdateUserPreferencesParams{FolderIconsEnabled: &disabled})

	require.NoError(t, err)
	assert.False(t, user.FolderIconsEnabled)
}

func TestUserUsecase_UpdatePreferences_AICategorisationEnabledSuccess(t *testing.T) {
	repo := newFakeUserRepo()
	repo.users["kp_abc123"] = &domain.User{ID: "kp_abc123", AICategorisationEnabled: false}
	uc := newTestUserUsecase(repo)
	enabled := true

	user, err := uc.UpdatePreferences(context.Background(), "kp_abc123", UpdateUserPreferencesParams{AICategorisationEnabled: &enabled})

	require.NoError(t, err)
	assert.True(t, user.AICategorisationEnabled)
}

func TestUserUsecase_UpdatePreferences_RepoError(t *testing.T) {
	repo := newFakeUserRepo()
	repo.updateErr = errors.New("db error")
	uc := newTestUserUsecase(repo)
	enabled := true

	_, err := uc.UpdatePreferences(context.Background(), "kp_abc123", UpdateUserPreferencesParams{VisionEnabled: &enabled})

	require.Error(t, err)
}
