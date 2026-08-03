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
	failForID string
}

func (f *failingHardDeleteUserRepo) HardDelete(ctx context.Context, id string) error {
	if id == f.failForID {
		return errors.New("hard delete failed")
	}
	return f.fakeUserRepo.HardDelete(ctx, id)
}

func newTestAccountUsecase(accountRepo AccountRepository, userRepo UserRepository, kinde KindeClient, enqueuer JobEnqueuer, bookletClient ...BookletClient) *accountUsecase {
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
	userID := "kp_abc123"
	userRepo := newFakeUserRepo()
	userRepo.users[userID] = &domain.User{ID: userID, AccountState: domain.AccountStateActive}
	enqueuer := &mockJobEnqueuer{}
	uc := newTestAccountUsecase(newMinimalAccountRepo(userRepo), userRepo, &fakeKindeClient{}, enqueuer)

	err := uc.MarkForDeletion(context.Background(), userID)

	require.NoError(t, err)
	assert.Equal(t, domain.AccountStatePendingDeletion, userRepo.users[userID].AccountState)
	require.Len(t, enqueuer.insertArgs, 1)
	wipeArgs, ok := enqueuer.insertArgs[0].(AccountWipeArgs)
	require.True(t, ok)
	assert.Equal(t, userID, wipeArgs.UserID)
}

func TestAccountUsecase_MarkForDeletion_NoOpForNonActiveUser(t *testing.T) {
	for _, state := range []domain.AccountState{domain.AccountStatePendingDeletion, domain.AccountStatePurged} {
		t.Run(string(state), func(t *testing.T) {
			userID := "kp_abc123"
			userRepo := newFakeUserRepo()
			userRepo.users[userID] = &domain.User{ID: userID, AccountState: state}
			enqueuer := &mockJobEnqueuer{}
			uc := newTestAccountUsecase(newMinimalAccountRepo(userRepo), userRepo, &fakeKindeClient{}, enqueuer)

			err := uc.MarkForDeletion(context.Background(), userID)

			require.NoError(t, err)
			assert.Equal(t, state, userRepo.users[userID].AccountState, "state must not change")
			assert.Empty(t, enqueuer.insertArgs)
		})
	}
}

func TestAccountUsecase_MarkForDeletion_EnqueuesWipeForUnprovisionedUser(t *testing.T) {
	userRepo := newFakeUserRepo() // user not in DB
	enqueuer := &mockJobEnqueuer{}
	uc := newTestAccountUsecase(newMinimalAccountRepo(userRepo), userRepo, &fakeKindeClient{}, enqueuer)

	err := uc.MarkForDeletion(context.Background(), "kp_unprovisioned")

	require.NoError(t, err)
	require.Len(t, enqueuer.insertArgs, 1)
	wipeArgs, ok := enqueuer.insertArgs[0].(AccountWipeArgs)
	require.True(t, ok)
	assert.Equal(t, "kp_unprovisioned", wipeArgs.UserID)
}

func TestAccountUsecase_MarkForDeletion_EnqueueFailureReturnsNil(t *testing.T) {
	userID := "kp_abc123"
	userRepo := newFakeUserRepo()
	userRepo.users[userID] = &domain.User{ID: userID, AccountState: domain.AccountStateActive}
	enqueuer := &mockJobEnqueuer{err: errors.New("queue unavailable")}
	uc := newTestAccountUsecase(newMinimalAccountRepo(userRepo), userRepo, &fakeKindeClient{}, enqueuer)

	err := uc.MarkForDeletion(context.Background(), userID)

	require.NoError(t, err)
	assert.Equal(t, domain.AccountStatePendingDeletion, userRepo.users[userID].AccountState, "state must still be updated")
}

// --- WipeAccount ---

func TestAccountUsecase_WipeAccount_TransitionsToPurged(t *testing.T) {
	userID := "kp_abc123"
	userRepo := newFakeUserRepo()
	userRepo.users[userID] = &domain.User{ID: userID, AccountState: domain.AccountStatePendingDeletion}
	kinde := &fakeKindeClient{}
	uc := newTestAccountUsecase(newMinimalAccountRepo(userRepo), userRepo, kinde, &mockJobEnqueuer{})

	err := uc.WipeAccount(context.Background(), userID)

	require.NoError(t, err)
	user := userRepo.users[userID]
	assert.Equal(t, domain.AccountStatePurged, user.AccountState)
	require.NotNil(t, user.PurgedAt, "purged_at must be set")
	assert.Equal(t, 1, kinde.deleteUserCalls)
}

func TestAccountUsecase_WipeAccount_SucceedsForUnprovisionedUser(t *testing.T) {
	userRepo := newFakeUserRepo() // no user row
	kinde := &fakeKindeClient{}
	uc := newTestAccountUsecase(newMinimalAccountRepo(userRepo), userRepo, kinde, &mockJobEnqueuer{})

	err := uc.WipeAccount(context.Background(), "kp_unprovisioned")

	require.NoError(t, err, "WipeAccount must succeed even when MarkPurged finds no row")
	assert.Equal(t, 1, kinde.deleteUserSessionsCalls)
	assert.Equal(t, 1, kinde.deleteUserCalls)
}

func TestAccountUsecase_WipeAccount_DeleteUserSessionsCalledBeforeDeleteUser(t *testing.T) {
	userID := "kp_abc123"
	userRepo := newFakeUserRepo()
	userRepo.users[userID] = &domain.User{ID: userID, AccountState: domain.AccountStatePendingDeletion}
	kinde := &fakeKindeClient{deleteUserSessionsErr: errors.New("sessions unavailable")}
	uc := newTestAccountUsecase(newMinimalAccountRepo(userRepo), userRepo, kinde, &mockJobEnqueuer{})

	err := uc.WipeAccount(context.Background(), userID)

	require.ErrorContains(t, err, "sessions unavailable")
	assert.Equal(t, 1, kinde.deleteUserSessionsCalls, "DeleteUserSessions must be called")
	assert.Equal(t, 0, kinde.deleteUserCalls, "DeleteUser must not be called when sessions deletion fails")
}

// --- WipeAccount enqueues R2 jobs ---

func TestAccountUsecase_WipeAccount_EnqueuesR2Jobs(t *testing.T) {
	userID := "kp_abc123"
	userRepo := newFakeUserRepo()
	userRepo.users[userID] = &domain.User{ID: userID, AccountState: domain.AccountStatePendingDeletion}
	thumb := "users/kp_abc123/thumbnails/img1.jpg"
	images := []*domain.Image{
		{ID: uuid.New(), R2Path: "users/kp_abc123/images/img1.jpg", ThumbnailPath: &thumb},
	}
	accountRepo := &fakeAccountRepository{
		images:         &mockImageRepository{images: images},
		folders:        &stubFolderRepo{},
		tags:           &mockTagRepository{},
		pendingUploads: &mockPendingUploadRepository{},
		users:          userRepo,
	}
	enqueuer := &mockJobEnqueuer{}
	uc := newTestAccountUsecase(accountRepo, userRepo, &fakeKindeClient{}, enqueuer)

	err := uc.WipeAccount(context.Background(), userID)

	require.NoError(t, err)
	var r2Args []R2DeleteArgs
	for _, a := range enqueuer.insertArgs {
		if r, ok := a.(R2DeleteArgs); ok {
			r2Args = append(r2Args, r)
		}
	}
	require.Len(t, r2Args, 1)
	assert.Equal(t, "users/kp_abc123/images/img1.jpg", r2Args[0].R2Path)
	require.NotNil(t, r2Args[0].ThumbnailPath)
	assert.Equal(t, thumb, *r2Args[0].ThumbnailPath)
}

// --- ReconcilePendingDeletions ---

func TestAccountUsecase_ReconcilePendingDeletions_EnqueuesJobPerPendingUser(t *testing.T) {
	userRepo := newFakeUserRepo()
	userRepo.users["kp_pending1"] = &domain.User{ID: "kp_pending1", AccountState: domain.AccountStatePendingDeletion}
	userRepo.users["kp_pending2"] = &domain.User{ID: "kp_pending2", AccountState: domain.AccountStatePendingDeletion}
	userRepo.users["kp_active"] = &domain.User{ID: "kp_active", AccountState: domain.AccountStateActive}
	enqueuer := &mockJobEnqueuer{}
	uc := newTestAccountUsecase(&fakeAccountRepository{}, userRepo, &fakeKindeClient{}, enqueuer)

	err := uc.ReconcilePendingDeletions(context.Background())

	require.NoError(t, err)
	var enqueuedUserIDs []string
	for _, args := range enqueuer.insertUniqueArgs {
		if a, ok := args.(AccountWipeArgs); ok {
			enqueuedUserIDs = append(enqueuedUserIDs, a.UserID)
		}
	}
	assert.ElementsMatch(t, []string{"kp_pending1", "kp_pending2"}, enqueuedUserIDs)
}

// --- SweepPurgedAccounts ---

func TestAccountUsecase_SweepPurgedAccounts_HardDeletesExpiredRows(t *testing.T) {
	userRepo := newFakeUserRepo()
	past := time.Now().Add(-26 * time.Hour)
	userRepo.users["kp_purged1"] = &domain.User{ID: "kp_purged1", AccountState: domain.AccountStatePurged, PurgedAt: &past}
	userRepo.users["kp_purged2"] = &domain.User{ID: "kp_purged2", AccountState: domain.AccountStatePurged, PurgedAt: &past}
	uc := newTestAccountUsecase(&fakeAccountRepository{}, userRepo, &fakeKindeClient{}, &mockJobEnqueuer{})

	err := uc.SweepPurgedAccounts(context.Background())

	require.NoError(t, err)
	_, exists1 := userRepo.users["kp_purged1"]
	_, exists2 := userRepo.users["kp_purged2"]
	assert.False(t, exists1, "kp_purged1 must be hard-deleted")
	assert.False(t, exists2, "kp_purged2 must be hard-deleted")
}

func TestAccountUsecase_SweepPurgedAccounts_ContinuesOnHardDeleteFailure(t *testing.T) {
	inner := newFakeUserRepo()
	past := time.Now().Add(-26 * time.Hour)
	inner.users["kp_fail"] = &domain.User{ID: "kp_fail", AccountState: domain.AccountStatePurged, PurgedAt: &past}
	inner.users["kp_ok"] = &domain.User{ID: "kp_ok", AccountState: domain.AccountStatePurged, PurgedAt: &past}
	userRepo := &failingHardDeleteUserRepo{fakeUserRepo: inner, failForID: "kp_fail"}
	uc := newTestAccountUsecase(&fakeAccountRepository{}, userRepo, &fakeKindeClient{}, &mockJobEnqueuer{})

	err := uc.SweepPurgedAccounts(context.Background())

	require.NoError(t, err, "SweepPurgedAccounts must return nil even when a hard-delete fails")
	_, okExists := inner.users["kp_ok"]
	assert.False(t, okExists, "kp_ok must still be hard-deleted")
}

// --- ProcessBookletUserDeletion ---

func TestAccountUsecase_ProcessBookletUserDeletion_SuccessReturnsNil(t *testing.T) {
	userID := "kp_abc123"
	bookletClient := &fakeBookletClient{}
	uc := newTestAccountUsecase(&fakeAccountRepository{}, newFakeUserRepo(), &fakeKindeClient{}, &mockJobEnqueuer{}, bookletClient)

	err := uc.ProcessBookletUserDeletion(context.Background(), userID)

	require.NoError(t, err)
	assert.Equal(t, userID, bookletClient.lastDeletedUserID)
}

func TestAccountUsecase_ProcessBookletUserDeletion_ClientErrorPropagates(t *testing.T) {
	bookletClient := &fakeBookletClient{err: errors.New("booklet unavailable")}
	uc := newTestAccountUsecase(&fakeAccountRepository{}, newFakeUserRepo(), &fakeKindeClient{}, &mockJobEnqueuer{}, bookletClient)

	err := uc.ProcessBookletUserDeletion(context.Background(), "kp_abc123")

	require.ErrorContains(t, err, "booklet unavailable")
}
