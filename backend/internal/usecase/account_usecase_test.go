package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- test doubles ---

type fakeAccountRepository struct {
	images         ImageRepository
	folders        FolderRepository
	tags           TagRepository
	pendingUploads PendingUploadRepository
	users          UserRepository
	transactionErr error
}

func (f *fakeAccountRepository) Transaction(_ context.Context, fn func(AccountRepos) error) error {
	if f.transactionErr != nil {
		return f.transactionErr
	}
	return fn(AccountRepos{
		Images:         f.images,
		Folders:        f.folders,
		Tags:           f.tags,
		PendingUploads: f.pendingUploads,
		Users:          f.users,
	})
}

type fakeKindeClient struct {
	deleteUserErr           error
	deleteUserSessionsErr   error
	deleteUserCalls         int
	deleteUserSessionsCalls int
	lastUserID              string
}

func (f *fakeKindeClient) DeleteUser(_ context.Context, userID string) error {
	f.deleteUserCalls++
	f.lastUserID = userID
	return f.deleteUserErr
}

func (f *fakeKindeClient) DeleteUserSessions(_ context.Context, userID string) error {
	f.deleteUserSessionsCalls++
	f.lastUserID = userID
	return f.deleteUserSessionsErr
}

type fakeBookletClient struct {
	err               error
	lastDeletedUserID string
}

func (f *fakeBookletClient) DeleteUser(_ context.Context, userID string) error {
	f.lastDeletedUserID = userID
	return f.err
}

// failingHardDeleteUserRepo wraps fakeUserRepo and returns an error for one specific user ID on HardDelete.
type failingHardDeleteUserRepo struct {
	*fakeUserRepo
	failForID uuid.UUID
}

func (f *failingHardDeleteUserRepo) HardDelete(ctx context.Context, id uuid.UUID) error {
	if id == f.failForID {
		return errors.New("hard delete failed")
	}
	return f.fakeUserRepo.HardDelete(ctx, id)
}

// --- account enqueuer mock ---

type mockAccountJobEnqueuer struct {
	err                    error
	accountWipeCalls       []string
	accountWipeUniqueCalls []string
	bookletDeletionCalls   []string
	r2DeleteCalls          []struct {
		path      string
		thumbnail *string
	}
}

func (m *mockAccountJobEnqueuer) EnqueueAccountWipe(_ context.Context, idpSubject string) error {
	m.accountWipeCalls = append(m.accountWipeCalls, idpSubject)
	return m.err
}
func (m *mockAccountJobEnqueuer) EnqueueAccountWipeUnique(_ context.Context, idpSubject string) error {
	m.accountWipeUniqueCalls = append(m.accountWipeUniqueCalls, idpSubject)
	return m.err
}
func (m *mockAccountJobEnqueuer) EnqueueBookletUserDeletion(_ context.Context, idpSubject string) error {
	m.bookletDeletionCalls = append(m.bookletDeletionCalls, idpSubject)
	return m.err
}
func (m *mockAccountJobEnqueuer) EnqueueR2Delete(_ context.Context, r2Path string, thumbnailPath *string) error {
	m.r2DeleteCalls = append(m.r2DeleteCalls, struct {
		path      string
		thumbnail *string
	}{path: r2Path, thumbnail: thumbnailPath})
	return m.err
}

func newTestAccountUsecase(accountRepo AccountRepository, userRepo UserRepository, kinde KindeClient, enqueuer accountJobEnqueuer, bookletClient ...BookletClient) *accountUsecase {
	var bc BookletClient
	if len(bookletClient) > 0 {
		bc = bookletClient[0]
	}
	return NewAccountUsecase(accountRepo, userRepo, kinde, enqueuer, noopTel(), bc)
}

func newMinimalAccountRepo(userRepo UserRepository) *fakeAccountRepository {
	return &fakeAccountRepository{
		images:         &mockImageRepository{},
		folders:        &stubFolderRepo{},
		tags:           &mockTagRepository{},
		pendingUploads: &mockPendingUploadRepository{},
		users:          userRepo,
	}
}

// --- MarkForDeletion ---

func TestAccountUsecase_MarkForDeletion_SetsStateAndEnqueuesJob(t *testing.T) {
	idpSubject := "kp_abc123"
	userRepo := newFakeUserRepo()
	user := &domain.User{ID: uuid.New(), IDPSubject: idpSubject, AccountState: domain.AccountStateActive}
	userRepo.seed(user)
	enqueuer := &mockAccountJobEnqueuer{}
	uc := newTestAccountUsecase(newMinimalAccountRepo(userRepo), userRepo, &fakeKindeClient{}, enqueuer)

	err := uc.MarkForDeletion(context.Background(), idpSubject)

	require.NoError(t, err)
	assert.Equal(t, domain.AccountStatePendingDeletion, userRepo.byIDPSubject[idpSubject].AccountState)
	require.Len(t, enqueuer.accountWipeCalls, 1)
	assert.Equal(t, idpSubject, enqueuer.accountWipeCalls[0])
}

func TestAccountUsecase_MarkForDeletion_NoOpForNonActiveUser(t *testing.T) {
	for _, state := range []domain.AccountState{domain.AccountStatePendingDeletion, domain.AccountStatePurged} {
		t.Run(string(state), func(t *testing.T) {
			idpSubject := "kp_abc123"
			userRepo := newFakeUserRepo()
			user := &domain.User{ID: uuid.New(), IDPSubject: idpSubject, AccountState: state}
			userRepo.seed(user)
			enqueuer := &mockAccountJobEnqueuer{}
			uc := newTestAccountUsecase(newMinimalAccountRepo(userRepo), userRepo, &fakeKindeClient{}, enqueuer)

			err := uc.MarkForDeletion(context.Background(), idpSubject)

			require.NoError(t, err)
			assert.Equal(t, state, userRepo.byIDPSubject[idpSubject].AccountState, "state must not change")
			assert.Empty(t, enqueuer.accountWipeCalls)
		})
	}
}

func TestAccountUsecase_MarkForDeletion_EnqueuesWipeForUnprovisionedUser(t *testing.T) {
	userRepo := newFakeUserRepo() // user not in DB
	enqueuer := &mockAccountJobEnqueuer{}
	uc := newTestAccountUsecase(newMinimalAccountRepo(userRepo), userRepo, &fakeKindeClient{}, enqueuer)

	err := uc.MarkForDeletion(context.Background(), "kp_unprovisioned")

	require.NoError(t, err)
	require.Len(t, enqueuer.accountWipeCalls, 1)
	assert.Equal(t, "kp_unprovisioned", enqueuer.accountWipeCalls[0])
}

func TestAccountUsecase_MarkForDeletion_EnqueueFailureReturnsNil(t *testing.T) {
	idpSubject := "kp_abc123"
	userRepo := newFakeUserRepo()
	user := &domain.User{ID: uuid.New(), IDPSubject: idpSubject, AccountState: domain.AccountStateActive}
	userRepo.seed(user)
	enqueuer := &mockAccountJobEnqueuer{err: errors.New("queue unavailable")}
	uc := newTestAccountUsecase(newMinimalAccountRepo(userRepo), userRepo, &fakeKindeClient{}, enqueuer)

	err := uc.MarkForDeletion(context.Background(), idpSubject)

	require.NoError(t, err)
	assert.Equal(t, domain.AccountStatePendingDeletion, userRepo.byIDPSubject[idpSubject].AccountState, "state must still be updated")
}

// --- WipeAccount ---

func TestAccountUsecase_WipeAccount_TransitionsToPurged(t *testing.T) {
	idpSubject := "kp_abc123"
	userRepo := newFakeUserRepo()
	user := &domain.User{ID: uuid.New(), IDPSubject: idpSubject, AccountState: domain.AccountStatePendingDeletion}
	userRepo.seed(user)
	kinde := &fakeKindeClient{}
	uc := newTestAccountUsecase(newMinimalAccountRepo(userRepo), userRepo, kinde, &mockAccountJobEnqueuer{})

	err := uc.WipeAccount(context.Background(), idpSubject)

	require.NoError(t, err)
	u := userRepo.byIDPSubject[idpSubject]
	assert.Equal(t, domain.AccountStatePurged, u.AccountState)
	require.NotNil(t, u.PurgedAt, "purged_at must be set")
	assert.Equal(t, 1, kinde.deleteUserCalls)
}

func TestAccountUsecase_WipeAccount_SucceedsForUnprovisionedUser(t *testing.T) {
	userRepo := newFakeUserRepo() // no user row
	kinde := &fakeKindeClient{}
	uc := newTestAccountUsecase(newMinimalAccountRepo(userRepo), userRepo, kinde, &mockAccountJobEnqueuer{})

	err := uc.WipeAccount(context.Background(), "kp_unprovisioned")

	require.NoError(t, err, "WipeAccount must succeed even when MarkPurged finds no row")
	assert.Equal(t, 1, kinde.deleteUserSessionsCalls)
	assert.Equal(t, 1, kinde.deleteUserCalls)
}

func TestAccountUsecase_WipeAccount_DeleteUserSessionsCalledBeforeDeleteUser(t *testing.T) {
	idpSubject := "kp_abc123"
	userRepo := newFakeUserRepo()
	user := &domain.User{ID: uuid.New(), IDPSubject: idpSubject, AccountState: domain.AccountStatePendingDeletion}
	userRepo.seed(user)
	kinde := &fakeKindeClient{deleteUserSessionsErr: errors.New("sessions unavailable")}
	uc := newTestAccountUsecase(newMinimalAccountRepo(userRepo), userRepo, kinde, &mockAccountJobEnqueuer{})

	err := uc.WipeAccount(context.Background(), idpSubject)

	require.ErrorContains(t, err, "sessions unavailable")
	assert.Equal(t, 1, kinde.deleteUserSessionsCalls, "DeleteUserSessions must be called")
	assert.Equal(t, 0, kinde.deleteUserCalls, "DeleteUser must not be called when sessions deletion fails")
}

// --- WipeAccount enqueues R2 jobs ---

func TestAccountUsecase_WipeAccount_EnqueuesR2Jobs(t *testing.T) {
	idpSubject := "kp_abc123"
	userID := uuid.New()
	userRepo := newFakeUserRepo()
	user := &domain.User{ID: userID, IDPSubject: idpSubject, AccountState: domain.AccountStatePendingDeletion}
	userRepo.seed(user)
	r2Path := "users/" + userID.String() + "/images/img1.jpg"
	thumb := "users/" + userID.String() + "/thumbnails/img1.jpg"
	images := []*domain.Image{
		{ID: uuid.New(), R2Path: r2Path, ThumbnailPath: &thumb},
	}
	accountRepo := &fakeAccountRepository{
		images:         &mockImageRepository{images: images},
		folders:        &stubFolderRepo{},
		tags:           &mockTagRepository{},
		pendingUploads: &mockPendingUploadRepository{},
		users:          userRepo,
	}
	enqueuer := &mockAccountJobEnqueuer{}
	uc := newTestAccountUsecase(accountRepo, userRepo, &fakeKindeClient{}, enqueuer)

	err := uc.WipeAccount(context.Background(), idpSubject)

	require.NoError(t, err)
	require.Len(t, enqueuer.r2DeleteCalls, 1)
	assert.Equal(t, r2Path, enqueuer.r2DeleteCalls[0].path)
	require.NotNil(t, enqueuer.r2DeleteCalls[0].thumbnail)
	assert.Equal(t, thumb, *enqueuer.r2DeleteCalls[0].thumbnail)
}

// --- ReconcilePendingDeletions ---

func TestAccountUsecase_ReconcilePendingDeletions_EnqueuesJobPerPendingUser(t *testing.T) {
	userRepo := newFakeUserRepo()
	userRepo.seed(&domain.User{ID: uuid.New(), IDPSubject: "kp_pending1", AccountState: domain.AccountStatePendingDeletion})
	userRepo.seed(&domain.User{ID: uuid.New(), IDPSubject: "kp_pending2", AccountState: domain.AccountStatePendingDeletion})
	userRepo.seed(&domain.User{ID: uuid.New(), IDPSubject: "kp_active", AccountState: domain.AccountStateActive})
	enqueuer := &mockAccountJobEnqueuer{}
	uc := newTestAccountUsecase(&fakeAccountRepository{}, userRepo, &fakeKindeClient{}, enqueuer)

	err := uc.ReconcilePendingDeletions(context.Background())

	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"kp_pending1", "kp_pending2"}, enqueuer.accountWipeUniqueCalls)
}

// --- SweepPurgedAccounts ---

func TestAccountUsecase_SweepPurgedAccounts_HardDeletesExpiredRows(t *testing.T) {
	userRepo := newFakeUserRepo()
	past := time.Now().Add(-26 * time.Hour)
	id1 := uuid.New()
	id2 := uuid.New()
	userRepo.seed(&domain.User{ID: id1, IDPSubject: "kp_purged1", AccountState: domain.AccountStatePurged, PurgedAt: &past})
	userRepo.seed(&domain.User{ID: id2, IDPSubject: "kp_purged2", AccountState: domain.AccountStatePurged, PurgedAt: &past})
	uc := newTestAccountUsecase(&fakeAccountRepository{}, userRepo, &fakeKindeClient{}, &mockAccountJobEnqueuer{})

	err := uc.SweepPurgedAccounts(context.Background())

	require.NoError(t, err)
	_, exists1 := userRepo.byID[id1]
	_, exists2 := userRepo.byID[id2]
	assert.False(t, exists1, "kp_purged1 must be hard-deleted")
	assert.False(t, exists2, "kp_purged2 must be hard-deleted")
}

func TestAccountUsecase_SweepPurgedAccounts_ContinuesOnHardDeleteFailure(t *testing.T) {
	inner := newFakeUserRepo()
	past := time.Now().Add(-26 * time.Hour)
	failID := uuid.New()
	okID := uuid.New()
	inner.seed(&domain.User{ID: failID, IDPSubject: "kp_fail", AccountState: domain.AccountStatePurged, PurgedAt: &past})
	inner.seed(&domain.User{ID: okID, IDPSubject: "kp_ok", AccountState: domain.AccountStatePurged, PurgedAt: &past})
	userRepo := &failingHardDeleteUserRepo{fakeUserRepo: inner, failForID: failID}
	uc := newTestAccountUsecase(&fakeAccountRepository{}, userRepo, &fakeKindeClient{}, &mockAccountJobEnqueuer{})

	err := uc.SweepPurgedAccounts(context.Background())

	require.NoError(t, err, "SweepPurgedAccounts must return nil even when a hard-delete fails")
	_, okExists := inner.byID[okID]
	assert.False(t, okExists, "kp_ok must still be hard-deleted")
}

// --- ProcessBookletUserDeletion ---

func TestAccountUsecase_ProcessBookletUserDeletion_SuccessReturnsNil(t *testing.T) {
	idpSubject := "kp_abc123"
	bookletClient := &fakeBookletClient{}
	uc := newTestAccountUsecase(&fakeAccountRepository{}, newFakeUserRepo(), &fakeKindeClient{}, &mockAccountJobEnqueuer{}, bookletClient)

	err := uc.ProcessBookletUserDeletion(context.Background(), idpSubject)

	require.NoError(t, err)
	assert.Equal(t, idpSubject, bookletClient.lastDeletedUserID)
}

func TestAccountUsecase_ProcessBookletUserDeletion_ClientErrorPropagates(t *testing.T) {
	bookletClient := &fakeBookletClient{err: errors.New("booklet unavailable")}
	uc := newTestAccountUsecase(&fakeAccountRepository{}, newFakeUserRepo(), &fakeKindeClient{}, &mockAccountJobEnqueuer{}, bookletClient)

	err := uc.ProcessBookletUserDeletion(context.Background(), "kp_abc123")

	require.ErrorContains(t, err, "booklet unavailable")
}
