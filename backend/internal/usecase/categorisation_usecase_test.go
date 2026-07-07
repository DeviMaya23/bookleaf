package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/devi/bookleaf/internal/agent"
	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- spies (inline: single-step operations, used only in this file) ---

type spyCategorisationAgent struct {
	suggestion agent.Suggestion
	err        error
	calls      int
}

func (s *spyCategorisationAgent) GetFolderSuggestion(_ context.Context, _ string, _ uuid.UUID) (agent.Suggestion, error) {
	s.calls++
	return s.suggestion, s.err
}

type spyCategorisationImageRepo struct {
	image             *domain.Image
	setFolderCalls    int
	setFolderFolderID *uuid.UUID
}

func (s *spyCategorisationImageRepo) GetByID(_ context.Context, _ uuid.UUID, _ string) (*domain.Image, error) {
	if s.image != nil {
		return s.image, nil
	}
	return &domain.Image{}, nil
}

func (s *spyCategorisationImageRepo) SetImageFolder(_ context.Context, _ uuid.UUID, folderID *uuid.UUID) error {
	s.setFolderCalls++
	s.setFolderFolderID = folderID
	return nil
}

type spyCategorisationFolderRepo struct {
	createCalls   int
	createdFolder *domain.Folder
}

func (s *spyCategorisationFolderRepo) Create(_ context.Context, folder *domain.Folder) (*domain.Folder, error) {
	s.createCalls++
	folder.ID = uuid.New()
	s.createdFolder = folder
	return folder, nil
}

// --- tests ---

func TestCategorisationUsecase_CategoriseImage_IdempotencyReusesPriorLog(t *testing.T) {
	imageID := uuid.New()
	folderID := uuid.New()
	logRepo := &fakeCategorisationLogRepo{
		existing: &domain.CategorisationLog{ImageID: &imageID, FolderID: &folderID},
	}
	agentSvc := &spyCategorisationAgent{}
	imageRepo := &spyCategorisationImageRepo{}
	uc := NewCategorisationUsecase(agentSvc, imageRepo, &spyCategorisationFolderRepo{}, logRepo, noopTel())

	err := uc.CategoriseImage(context.Background(), "kp_abc123", imageID)

	require.NoError(t, err)
	assert.Equal(t, 0, agentSvc.calls, "agent must not be called when a prior log entry exists")
	require.NotNil(t, imageRepo.setFolderFolderID)
	assert.Equal(t, folderID, *imageRepo.setFolderFolderID)
}

func TestCategorisationUsecase_CategoriseImage_FolderIDAssigned(t *testing.T) {
	imageID := uuid.New()
	folderID := uuid.New()
	logRepo := &fakeCategorisationLogRepo{}
	agentSvc := &spyCategorisationAgent{
		suggestion: agent.Suggestion{FolderID: folderID.String(), Reasoning: "matches existing folder"},
	}
	imageRepo := &spyCategorisationImageRepo{}
	uc := NewCategorisationUsecase(agentSvc, imageRepo, &spyCategorisationFolderRepo{}, logRepo, noopTel())

	err := uc.CategoriseImage(context.Background(), "kp_abc123", imageID)

	require.NoError(t, err)
	require.NotNil(t, logRepo.created)
	assert.Equal(t, "matches existing folder", logRepo.created.Reasoning)
	require.NotNil(t, imageRepo.setFolderFolderID)
	assert.Equal(t, folderID, *imageRepo.setFolderFolderID)
}

func TestCategorisationUsecase_CategoriseImage_NewFolderCreatedAndAssigned(t *testing.T) {
	imageID := uuid.New()
	logRepo := &fakeCategorisationLogRepo{}
	agentSvc := &spyCategorisationAgent{
		suggestion: agent.Suggestion{NewFolderName: "Nature", Reasoning: "no matching folder"},
	}
	imageRepo := &spyCategorisationImageRepo{}
	folderRepo := &spyCategorisationFolderRepo{}
	uc := NewCategorisationUsecase(agentSvc, imageRepo, folderRepo, logRepo, noopTel())

	err := uc.CategoriseImage(context.Background(), "kp_abc123", imageID)

	require.NoError(t, err)
	require.NotNil(t, logRepo.created)
	assert.Equal(t, 1, folderRepo.createCalls)
	assert.Equal(t, "Nature", folderRepo.createdFolder.Name)
	require.NotNil(t, imageRepo.setFolderFolderID)
	assert.Equal(t, folderRepo.createdFolder.ID, *imageRepo.setFolderFolderID)
}

func TestCategorisationUsecase_CategoriseImage_AlreadyInFolder_ReturnsNil(t *testing.T) {
	imageID := uuid.New()
	folderID := uuid.New()
	imageRepo := &spyCategorisationImageRepo{
		image: &domain.Image{ImageFolders: []domain.ImageFolder{{FolderID: folderID}}},
	}
	agentSvc := &spyCategorisationAgent{}
	uc := NewCategorisationUsecase(agentSvc, imageRepo, &spyCategorisationFolderRepo{}, &fakeCategorisationLogRepo{}, noopTel())

	err := uc.CategoriseImage(context.Background(), "kp_abc123", imageID)

	require.NoError(t, err)
	assert.Equal(t, 0, agentSvc.calls, "agent must not be called when image is already in a folder")
	assert.Equal(t, 0, imageRepo.setFolderCalls, "SetImageFolder must not be called when image is already in a folder")
}

func TestCategorisationUsecase_CategoriseImage_EmptySuggestionReturnsError(t *testing.T) {
	imageID := uuid.New()
	logRepo := &fakeCategorisationLogRepo{}
	agentSvc := &spyCategorisationAgent{suggestion: agent.Suggestion{Reasoning: "indecisive"}}
	uc := NewCategorisationUsecase(agentSvc, &spyCategorisationImageRepo{}, &spyCategorisationFolderRepo{}, logRepo, noopTel())

	err := uc.CategoriseImage(context.Background(), "kp_abc123", imageID)

	require.ErrorContains(t, err, "agent returned suggestion with no folder id or new folder name")
}

func TestCategorisationUsecase_CategoriseImage_MonthlyLimitReached_SkipsAgent(t *testing.T) {
	imageID := uuid.New()
	logRepo := &fakeCategorisationLogRepo{monthCount: 50}
	agentSvc := &spyCategorisationAgent{}
	uc := NewCategorisationUsecase(agentSvc, &spyCategorisationImageRepo{}, &spyCategorisationFolderRepo{}, logRepo, noopTel())

	err := uc.CategoriseImage(context.Background(), "kp_abc123", imageID)

	require.NoError(t, err)
	assert.Equal(t, 0, agentSvc.calls, "agent must not be called when monthly limit is reached")
	assert.Nil(t, logRepo.created, "no log entry must be created when monthly limit is reached")
}

func TestCategorisationUsecase_CategoriseImage_CountQueryError_ReturnsError(t *testing.T) {
	imageID := uuid.New()
	countErr := errors.New("db error")
	logRepo := &fakeCategorisationLogRepo{countErr: countErr}
	agentSvc := &spyCategorisationAgent{}
	uc := NewCategorisationUsecase(agentSvc, &spyCategorisationImageRepo{}, &spyCategorisationFolderRepo{}, logRepo, noopTel())

	err := uc.CategoriseImage(context.Background(), "kp_abc123", imageID)

	require.ErrorContains(t, err, "db error")
	assert.Equal(t, 0, agentSvc.calls, "agent must not be called when count query errors")
}
