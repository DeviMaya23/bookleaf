package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- fake ---

type fakeUserRepo struct {
	byIDPSubject map[string]*domain.User
	byID         map[uuid.UUID]*domain.User
	getOrCreate  error
	updateErr    error
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		byIDPSubject: make(map[string]*domain.User),
		byID:         make(map[uuid.UUID]*domain.User),
	}
}

func (f *fakeUserRepo) seed(u *domain.User) {
	f.byID[u.ID] = u
	f.byIDPSubject[u.IDPSubject] = u
}

func (f *fakeUserRepo) GetOrCreate(_ context.Context, idpSubject string) (*domain.User, error) {
	if f.getOrCreate != nil {
		return nil, f.getOrCreate
	}
	if u, ok := f.byIDPSubject[idpSubject]; ok {
		return u, nil
	}
	u := &domain.User{ID: uuid.New(), IDPSubject: idpSubject}
	f.seed(u)
	return u, nil
}

func (f *fakeUserRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (f *fakeUserRepo) GetByIDPSubject(_ context.Context, idpSubject string) (*domain.User, error) {
	u, ok := f.byIDPSubject[idpSubject]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (f *fakeUserRepo) SetAccountState(_ context.Context, id uuid.UUID, state domain.AccountState) error {
	u, ok := f.byID[id]
	if !ok {
		return ErrUserNotFound
	}
	u.AccountState = state
	return nil
}

func (f *fakeUserRepo) MarkPurged(_ context.Context, id uuid.UUID, purgedAt time.Time) error {
	u, ok := f.byID[id]
	if !ok {
		return ErrUserNotFound
	}
	u.AccountState = domain.AccountStatePurged
	u.PurgedAt = &purgedAt
	return nil
}

func (f *fakeUserRepo) ListByAccountState(_ context.Context, state domain.AccountState) ([]*domain.User, error) {
	var users []*domain.User
	for _, u := range f.byID {
		if u.AccountState == state {
			users = append(users, u)
		}
	}
	return users, nil
}

func (f *fakeUserRepo) ListPurgedBefore(_ context.Context, threshold time.Time) ([]*domain.User, error) {
	var users []*domain.User
	for _, u := range f.byID {
		if u.AccountState == domain.AccountStatePurged && u.PurgedAt != nil && u.PurgedAt.Before(threshold) {
			users = append(users, u)
		}
	}
	return users, nil
}

func (f *fakeUserRepo) HardDelete(_ context.Context, id uuid.UUID) error {
	u, ok := f.byID[id]
	if !ok {
		return errors.New("record not found")
	}
	delete(f.byID, id)
	delete(f.byIDPSubject, u.IDPSubject)
	return nil
}

func (f *fakeUserRepo) Update(_ context.Context, id uuid.UUID, fields map[string]any) (*domain.User, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	u, ok := f.byID[id]
	if !ok {
		return nil, errors.New("record not found")
	}
	if enabled, ok := fields["vision_enabled"].(bool); ok {
		u.VisionEnabled = enabled
	}
	if enabled, ok := fields["folder_icons_enabled"].(bool); ok {
		u.FolderIconsEnabled = enabled
	}
	if enabled, ok := fields["ai_categorisation_enabled"].(bool); ok {
		u.AICategorisationEnabled = enabled
	}
	return u, nil
}

func newTestUserUsecase(repo UserRepository) *userUsecase {
	return NewUserUsecase(repo, observability.NewTelemetry(nil, nil, nil))
}

// --- tests ---

func TestUserUsecase_GetOrProvision_ExistingUser(t *testing.T) {
	repo := newFakeUserRepo()
	existing := &domain.User{ID: uuid.New(), IDPSubject: "kp_abc123"}
	repo.seed(existing)
	uc := newTestUserUsecase(repo)

	user, err := uc.GetOrProvision(context.Background(), "kp_abc123")

	require.NoError(t, err)
	assert.Equal(t, existing.ID, user.ID)
	assert.Equal(t, "kp_abc123", user.IDPSubject)
}

func TestUserUsecase_GetOrProvision_NewUser(t *testing.T) {
	repo := newFakeUserRepo()
	uc := newTestUserUsecase(repo)

	user, err := uc.GetOrProvision(context.Background(), "kp_new123")

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.Equal(t, "kp_new123", user.IDPSubject)
	_, exists := repo.byIDPSubject["kp_new123"]
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
	userID := uuid.New()
	repo.seed(&domain.User{ID: userID, IDPSubject: "kp_abc123", VisionEnabled: false})
	uc := newTestUserUsecase(repo)
	enabled := true

	user, err := uc.UpdatePreferences(context.Background(), userID, UpdateUserPreferencesParams{VisionEnabled: &enabled})

	require.NoError(t, err)
	assert.True(t, user.VisionEnabled)
}

func TestUserUsecase_UpdatePreferences_FolderIconsEnabledSuccess(t *testing.T) {
	repo := newFakeUserRepo()
	userID := uuid.New()
	repo.seed(&domain.User{ID: userID, IDPSubject: "kp_abc123", FolderIconsEnabled: true})
	uc := newTestUserUsecase(repo)
	disabled := false

	user, err := uc.UpdatePreferences(context.Background(), userID, UpdateUserPreferencesParams{FolderIconsEnabled: &disabled})

	require.NoError(t, err)
	assert.False(t, user.FolderIconsEnabled)
}

func TestUserUsecase_UpdatePreferences_AICategorisationEnabledSuccess(t *testing.T) {
	repo := newFakeUserRepo()
	userID := uuid.New()
	repo.seed(&domain.User{ID: userID, IDPSubject: "kp_abc123", AICategorisationEnabled: false})
	uc := newTestUserUsecase(repo)
	enabled := true

	user, err := uc.UpdatePreferences(context.Background(), userID, UpdateUserPreferencesParams{AICategorisationEnabled: &enabled})

	require.NoError(t, err)
	assert.True(t, user.AICategorisationEnabled)
}

func TestUserUsecase_UpdatePreferences_RepoError(t *testing.T) {
	repo := newFakeUserRepo()
	repo.updateErr = errors.New("db error")
	uc := newTestUserUsecase(repo)
	enabled := true

	_, err := uc.UpdatePreferences(context.Background(), uuid.New(), UpdateUserPreferencesParams{VisionEnabled: &enabled})

	require.Error(t, err)
}
