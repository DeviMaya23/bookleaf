package usecase

import (
	"context"
	"errors"
	"testing"

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
	err             error
	deleteUserCalls int
	lastUserID      string
}

func (f *fakeKindeClient) DeleteUser(_ context.Context, userID string) error {
	f.deleteUserCalls++
	f.lastUserID = userID
	return f.err
}

func newTestAccountUsecase(accountRepo AccountRepository, userRepo UserRepository, kinde KindeClient, enqueuer JobEnqueuer) *accountUsecase {
	return NewAccountUsecase(accountRepo, userRepo, kinde, enqueuer, noopTel())
}

// --- DeleteAccount ---

func TestAccountUsecase_DeleteAccount_EnqueuesR2AndKindeDeletionJobs(t *testing.T) {
	userID := "kp_abc123"
	userRepo := newFakeUserRepo()
	userRepo.users[userID] = &domain.User{ID: userID}

	thumb := "users/kp_abc123/thumbnails/img1.jpg"
	images := []*domain.Image{
		{ID: uuid.New(), R2Path: "users/kp_abc123/images/img1.jpg", ThumbnailPath: &thumb},
		{ID: uuid.New(), R2Path: "users/kp_abc123/images/img2.jpg"},
	}
	pendingUploads := []*domain.PendingUpload{
		{ID: uuid.New(), R2Path: "users/kp_abc123/images/pending1.jpg"},
	}

	accountRepo := &fakeAccountRepository{
		images:         &mockImageRepository{images: images},
		folders:        &stubFolderRepo{},
		tags:           &mockTagRepository{},
		pendingUploads: &mockPendingUploadRepository{pendings: pendingUploads},
		users:          userRepo,
	}
	enqueuer := &mockJobEnqueuer{}
	uc := newTestAccountUsecase(accountRepo, userRepo, &fakeKindeClient{}, enqueuer)

	err := uc.DeleteAccount(context.Background(), userID)

	require.NoError(t, err)

	var r2Args []R2DeleteArgs
	var kindeArgs []AccountKindeDeletionArgs
	for _, args := range enqueuer.insertArgs {
		switch a := args.(type) {
		case R2DeleteArgs:
			r2Args = append(r2Args, a)
		case AccountKindeDeletionArgs:
			kindeArgs = append(kindeArgs, a)
		}
	}

	require.Len(t, r2Args, 3)
	assert.Equal(t, "users/kp_abc123/images/img1.jpg", r2Args[0].R2Path)
	require.NotNil(t, r2Args[0].ThumbnailPath)
	assert.Equal(t, thumb, *r2Args[0].ThumbnailPath)
	assert.Equal(t, "users/kp_abc123/images/img2.jpg", r2Args[1].R2Path)
	assert.Nil(t, r2Args[1].ThumbnailPath)
	assert.Equal(t, "users/kp_abc123/images/pending1.jpg", r2Args[2].R2Path)
	assert.Nil(t, r2Args[2].ThumbnailPath)

	require.Len(t, kindeArgs, 1)
	assert.Equal(t, userID, kindeArgs[0].UserID)

	require.True(t, userRepo.users[userID].PendingKindeDeletion, "user row must be tombstoned")
}

func TestAccountUsecase_DeleteAccount_TransactionErrorEnqueuesNoJobs(t *testing.T) {
	userID := "kp_abc123"
	accountRepo := &fakeAccountRepository{transactionErr: errors.New("db unavailable")}
	enqueuer := &mockJobEnqueuer{}
	uc := newTestAccountUsecase(accountRepo, newFakeUserRepo(), &fakeKindeClient{}, enqueuer)

	err := uc.DeleteAccount(context.Background(), userID)

	require.ErrorContains(t, err, "db unavailable")
	assert.Empty(t, enqueuer.insertArgs)
}

// --- ProcessAccountKindeDeletion ---

func TestAccountUsecase_ProcessAccountKindeDeletion_Success(t *testing.T) {
	userID := "kp_abc123"
	userRepo := newFakeUserRepo()
	userRepo.users[userID] = &domain.User{ID: userID, PendingKindeDeletion: true}
	kinde := &fakeKindeClient{}
	uc := newTestAccountUsecase(&fakeAccountRepository{}, userRepo, kinde, &mockJobEnqueuer{})

	err := uc.ProcessAccountKindeDeletion(context.Background(), userID)

	require.NoError(t, err)
	assert.Equal(t, 1, kinde.deleteUserCalls)
	assert.Equal(t, userID, kinde.lastUserID)
	_, stillExists := userRepo.users[userID]
	assert.False(t, stillExists, "user row must be hard-deleted on success")
}

func TestAccountUsecase_ProcessAccountKindeDeletion_KindeErrorLeavesUserRow(t *testing.T) {
	userID := "kp_abc123"
	userRepo := newFakeUserRepo()
	userRepo.users[userID] = &domain.User{ID: userID, PendingKindeDeletion: true}
	kinde := &fakeKindeClient{err: errors.New("kinde unavailable")}
	uc := newTestAccountUsecase(&fakeAccountRepository{}, userRepo, kinde, &mockJobEnqueuer{})

	err := uc.ProcessAccountKindeDeletion(context.Background(), userID)

	require.ErrorContains(t, err, "kinde unavailable")
	_, stillExists := userRepo.users[userID]
	assert.True(t, stillExists, "user row must not be deleted when kinde deletion fails")
}

// --- ReconcilePendingKindeDeletions ---

func TestAccountUsecase_ReconcilePendingKindeDeletions_EnqueuesJobPerPendingUser(t *testing.T) {
	userRepo := newFakeUserRepo()
	userRepo.users["kp_pending1"] = &domain.User{ID: "kp_pending1", PendingKindeDeletion: true}
	userRepo.users["kp_pending2"] = &domain.User{ID: "kp_pending2", PendingKindeDeletion: true}
	userRepo.users["kp_active"] = &domain.User{ID: "kp_active"}
	enqueuer := &mockJobEnqueuer{}
	uc := newTestAccountUsecase(&fakeAccountRepository{}, userRepo, &fakeKindeClient{}, enqueuer)

	err := uc.ReconcilePendingKindeDeletions(context.Background())

	require.NoError(t, err)
	var enqueuedUserIDs []string
	for _, args := range enqueuer.insertArgs {
		if a, ok := args.(AccountKindeDeletionArgs); ok {
			enqueuedUserIDs = append(enqueuedUserIDs, a.UserID)
		}
	}
	assert.ElementsMatch(t, []string{"kp_pending1", "kp_pending2"}, enqueuedUserIDs)
}
