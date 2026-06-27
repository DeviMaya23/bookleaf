package usecase

import (
	"context"
	"fmt"

	"github.com/devi/bookleaf/internal/agent"
	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"
)

type CategorisationUsecase struct {
	imageRepo    categorisationImageRepository
	folderRepo   categorisationFolderRepository
	agentService *agent.AgentService
	tel          *observability.Telemetry
}

type categorisationImageRepository interface {
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error)
	SetImageFolder(ctx context.Context, imageID uuid.UUID, folderID *uuid.UUID) error
}

type categorisationFolderRepository interface {
	Create(ctx context.Context, folder *domain.Folder) (*domain.Folder, error)
}

func NewCategorisationUsecase(agentService *agent.AgentService,
	imageRepo categorisationImageRepository,
	folderRepo categorisationFolderRepository,
	tel *observability.Telemetry) *CategorisationUsecase {
	return &CategorisationUsecase{
		agentService: agentService,
		imageRepo:    imageRepo,
		folderRepo:   folderRepo,
		tel:          tel,
	}
}

func (u *CategorisationUsecase) CategoriseImage(ctx context.Context, userID string, imageID uuid.UUID) error {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.CategoriseImage")
	defer span.End()

	res, err := u.agentService.GetFolderSuggestion(ctx, userID, imageID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	var folderUUID uuid.UUID

	if res.FolderID != "" {
		folderUUID, err = uuid.Parse(res.FolderID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("parse suggested folder id: %w", err)
		}
	} else if res.NewFolderName != "" {
		var parentFolderUUID uuid.UUID
		if res.NewFolderParentID != "" {
			parentFolderUUID, err = uuid.Parse(res.NewFolderParentID)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return fmt.Errorf("parse suggested parent folder id: %w", err)
			}
		}
		newFolder, err := u.folderRepo.Create(ctx, &domain.Folder{
			UserID:   userID,
			Name:     res.NewFolderName,
			ParentID: &parentFolderUUID,
		})
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		folderUUID = newFolder.ID
	} else {
		return fmt.Errorf("agent returned suggestion with no folder id or new folder name")
	}

	if err := u.imageRepo.SetImageFolder(ctx, imageID, &folderUUID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nil
}
