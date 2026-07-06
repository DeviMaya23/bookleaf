package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/devi/bookleaf/internal/agent"
	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"
)

const categorisationMonthlyLimit = 500

type CategorisationUsecase struct {
	imageRepo    categorisationImageRepository
	folderRepo   categorisationFolderRepository
	logRepo      categorisationLogRepository
	agentService CategorisationAgentService
	tel          *observability.Telemetry
}

type CategorisationAgentService interface {
	GetFolderSuggestion(ctx context.Context, userID string, imageID uuid.UUID) (agent.Suggestion, error)
}

type categorisationImageRepository interface {
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error)
	SetImageFolder(ctx context.Context, imageID uuid.UUID, folderID *uuid.UUID) error
}

type categorisationFolderRepository interface {
	Create(ctx context.Context, folder *domain.Folder) (*domain.Folder, error)
}

type categorisationLogRepository interface {
	Create(ctx context.Context, log *domain.CategorisationLog) error
	GetByImageID(ctx context.Context, imageID uuid.UUID) (*domain.CategorisationLog, error)
	CountByUserAndMonth(ctx context.Context, userID string, year, month int) (int, error)
}

func NewCategorisationUsecase(agentService CategorisationAgentService,
	imageRepo categorisationImageRepository,
	folderRepo categorisationFolderRepository,
	logRepo categorisationLogRepository,
	tel *observability.Telemetry) *CategorisationUsecase {
	return &CategorisationUsecase{
		agentService: agentService,
		imageRepo:    imageRepo,
		folderRepo:   folderRepo,
		logRepo:      logRepo,
		tel:          tel,
	}
}

func (u *CategorisationUsecase) CountThisMonth(ctx context.Context, userID string) (int, error) {
	now := time.Now().UTC()
	return u.logRepo.CountByUserAndMonth(ctx, userID, now.Year(), int(now.Month()))
}

func (u *CategorisationUsecase) CategoriseImage(ctx context.Context, userID string, imageID uuid.UUID) error {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.CategoriseImage")
	defer span.End()

	if u.agentService == nil {
		return fmt.Errorf("agent service not configured")
	}

	count, err := u.CountThisMonth(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	if count >= categorisationMonthlyLimit {
		return nil
	}

	var folderID *uuid.UUID
	var newFolderName string
	var newFolderParentID uuid.UUID

	existingLog, err := u.logRepo.GetByImageID(ctx, imageID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	img, err := u.imageRepo.GetByID(ctx, imageID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	if len(img.ImageFolders) > 0 {
		return nil
	}

	if existingLog != nil {
		folderID = existingLog.FolderID
		if existingLog.NewFolderName != nil {
			newFolderName = *existingLog.NewFolderName
		}
	} else {
		res, err := u.agentService.GetFolderSuggestion(ctx, userID, imageID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}

		imgID := imageID
		logEntry := &domain.CategorisationLog{
			ImageID:   &imgID,
			UserID:    userID,
			Reasoning: res.Reasoning,
		}
		if res.FolderID != "" {
			if parsed, parseErr := uuid.Parse(res.FolderID); parseErr == nil {
				logEntry.FolderID = &parsed
				folderID = logEntry.FolderID
			}
		}
		if res.NewFolderName != "" {
			name := res.NewFolderName
			logEntry.NewFolderName = &name
			newFolderName = res.NewFolderName
			if res.NewFolderParentID != "" {
				if parsed, parseErr := uuid.Parse(res.NewFolderParentID); parseErr == nil {
					newFolderParentID = parsed
				}
			}
		}
		if logErr := u.logRepo.Create(ctx, logEntry); logErr != nil {
			span.RecordError(logErr)
			span.SetStatus(codes.Error, logErr.Error())
			return logErr
		}
	}

	if folderID != nil {
		if err := u.imageRepo.SetImageFolder(ctx, imageID, folderID); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
	} else if newFolderName != "" {
		var parentPtr *uuid.UUID
		if newFolderParentID != uuid.Nil {
			parentPtr = &newFolderParentID
		}
		newFolder, err := u.folderRepo.Create(ctx, &domain.Folder{
			UserID:   userID,
			Name:     newFolderName,
			ParentID: parentPtr,
		})
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		if err := u.imageRepo.SetImageFolder(ctx, imageID, &newFolder.ID); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
	} else {
		return fmt.Errorf("agent returned suggestion with no folder id or new folder name")
	}

	return nil
}
