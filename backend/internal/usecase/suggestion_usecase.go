package usecase

import (
	"context"

	"github.com/devi/bookleaf/internal/agent"
	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
)

type SuggestionUsecase struct {
	imageRepo    suggestionImageRepository
	folderRepo   suggestionFolderRepository
	agentService *agent.AgentService
}

type suggestionImageRepository interface {
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error)
	SetImageFolder(ctx context.Context, imageID uuid.UUID, folderID *uuid.UUID) error
}

type suggestionFolderRepository interface {
	Create(ctx context.Context, folder *domain.Folder) (*domain.Folder, error)
}

func NewSuggestionUsecase(agentService *agent.AgentService,
	imageRepo suggestionImageRepository,
	folderRepo suggestionFolderRepository) *SuggestionUsecase {
	return &SuggestionUsecase{
		agentService: agentService,
		imageRepo:    imageRepo,
		folderRepo:   folderRepo,
	}
}

func (u *SuggestionUsecase) CategoriseImage(ctx context.Context, userID string, imageID uuid.UUID) (any, error) {

	// TODO check if user has ai helper on
	// maybe check if image belongs to user

	res, err := u.agentService.GetFolderSuggestion(ctx, userID, imageID)
	if err != nil {
		return nil, err
	}
	var folderUUID uuid.UUID

	if res.FolderID != "" {
		// set image folder
		folderUUID, err = uuid.Parse(res.FolderID)
		if err != nil {
			return nil, err
		}
	} else if res.NewFolderName != "" {
		// create new folder and set image folder
		parentFolderUUID, err := uuid.Parse(res.NewFolderParentID)
		if err != nil {
			// TODO log
			parentFolderUUID = uuid.Nil
		}

		newFolder, err := u.folderRepo.Create(ctx, &domain.Folder{
			UserID:   userID,
			Name:     res.NewFolderName,
			ParentID: &parentFolderUUID,
		})
		if err != nil {
			return nil, err
		}
		folderUUID = newFolder.ID
	} else {
		// no folder suggestion, do nothing
		return res, nil
	}

	err = u.imageRepo.SetImageFolder(ctx, imageID, &folderUUID)
	if err != nil {
		return nil, err
	}
	return res, nil

}
