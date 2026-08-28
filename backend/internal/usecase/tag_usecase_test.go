package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- spy ---

type stubTagRepo struct {
	tag *domain.Tag
	err error
}

func (s *stubTagRepo) Create(_ context.Context, _ *domain.Tag) (*domain.Tag, error) {
	return s.tag, s.err
}
func (s *stubTagRepo) ListByUserID(_ context.Context, _ uuid.UUID) ([]*domain.Tag, error) {
	return nil, s.err
}
func (s *stubTagRepo) GetByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*domain.Tag, error) {
	return s.tag, s.err
}
func (s *stubTagRepo) Update(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ string) (*domain.Tag, error) {
	return s.tag, s.err
}
func (s *stubTagRepo) Delete(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return s.err
}
func (s *stubTagRepo) ReplaceImageTags(_ context.Context, _ uuid.UUID, _ []uuid.UUID) error {
	return s.err
}
func (s *stubTagRepo) DeleteAllByUserID(_ context.Context, _ uuid.UUID) error {
	return s.err
}

func newTestTagUsecase(repo TagRepository) *tagUsecase {
	return NewTagUsecase(repo, observability.NewTelemetry(nil, nil, nil))
}

// --- tests ---

func TestTagUsecase_Create_BlankName(t *testing.T) {
	uc := newTestTagUsecase(&stubTagRepo{})

	_, err := uc.Create(context.Background(), uuid.New(), "   ")

	require.ErrorIs(t, err, ErrInvalidTagName)
}

func TestTagUsecase_Create_DuplicateName(t *testing.T) {
	uc := newTestTagUsecase(&stubTagRepo{err: errors.New("23505")})

	_, err := uc.Create(context.Background(), uuid.New(), "nature")

	require.ErrorIs(t, err, ErrDuplicateTagName)
}

func TestTagUsecase_Create_GenericRepoError(t *testing.T) {
	uc := newTestTagUsecase(&stubTagRepo{err: errors.New("db error")})

	_, err := uc.Create(context.Background(), uuid.New(), "nature")

	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrDuplicateTagName)
}

func TestTagUsecase_Update_BlankName(t *testing.T) {
	uc := newTestTagUsecase(&stubTagRepo{})

	_, err := uc.Update(context.Background(), uuid.New(), uuid.New(), "   ")

	require.ErrorIs(t, err, ErrInvalidTagName)
}

func TestTagUsecase_Update_DuplicateName(t *testing.T) {
	uc := newTestTagUsecase(&stubTagRepo{err: errors.New("23505")})

	_, err := uc.Update(context.Background(), uuid.New(), uuid.New(), "existing")

	require.ErrorIs(t, err, ErrDuplicateTagName)
}
