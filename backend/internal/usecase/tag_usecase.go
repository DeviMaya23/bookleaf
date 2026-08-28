package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"
)

var (
	ErrInvalidTagName   = errors.New("tag name is required")
	ErrDuplicateTagName = errors.New("tag name already exists")
)

type tagUsecase struct {
	tagRepo TagRepository
	tel     *observability.Telemetry
}

func NewTagUsecase(tagRepo TagRepository, tel *observability.Telemetry) *tagUsecase {
	return &tagUsecase{tagRepo: tagRepo, tel: tel}
}

func (u *tagUsecase) Create(ctx context.Context, userID uuid.UUID, name string) (*domain.Tag, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.CreateTag")
	defer span.End()

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidTagName
	}

	tag, err := u.tagRepo.Create(ctx, &domain.Tag{UserID: userID, Name: name})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if isUniqueConstraintViolation(err) {
			return nil, ErrDuplicateTagName
		}
		return nil, fmt.Errorf("create tag: %w", err)
	}

	return tag, nil
}

func (u *tagUsecase) List(ctx context.Context, userID uuid.UUID) ([]*domain.Tag, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.ListTags")
	defer span.End()

	tags, err := u.tagRepo.ListByUserID(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("list tags: %w", err)
	}

	return tags, nil
}

func (u *tagUsecase) Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, name string) (*domain.Tag, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.UpdateTag")
	defer span.End()

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidTagName
	}

	tag, err := u.tagRepo.Update(ctx, id, userID, name)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if isUniqueConstraintViolation(err) {
			return nil, ErrDuplicateTagName
		}
		return nil, fmt.Errorf("update tag: %w", err)
	}

	return tag, nil
}

func (u *tagUsecase) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.DeleteTag")
	defer span.End()

	if err := u.tagRepo.Delete(ctx, id, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("delete tag: %w", err)
	}

	return nil
}

// isUniqueConstraintViolation detects a Postgres unique constraint violation.
func isUniqueConstraintViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "23505")
}

