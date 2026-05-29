package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/observability"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// --- mock ---

type mockTagRepository struct {
	tag                *domain.Tag
	tags               []*domain.Tag
	err                error
	replaceCalls       int
	lastReplaceImageID uuid.UUID
	lastReplaceTagIDs  []uuid.UUID
}

func (m *mockTagRepository) Create(_ context.Context, tag *domain.Tag) (*domain.Tag, error) {
	return m.tag, m.err
}

func (m *mockTagRepository) ListByUserID(_ context.Context, _ string) ([]*domain.Tag, error) {
	return m.tags, m.err
}

func (m *mockTagRepository) GetByID(_ context.Context, _ uuid.UUID, _ string) (*domain.Tag, error) {
	return m.tag, m.err
}

func (m *mockTagRepository) Update(_ context.Context, _ uuid.UUID, _ string, _ string) (*domain.Tag, error) {
	return m.tag, m.err
}

func (m *mockTagRepository) Delete(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}

func (m *mockTagRepository) ReplaceImageTags(_ context.Context, imageID uuid.UUID, tagIDs []uuid.UUID) error {
	m.replaceCalls++
	m.lastReplaceImageID = imageID
	if tagIDs != nil {
		m.lastReplaceTagIDs = append([]uuid.UUID(nil), tagIDs...)
	} else {
		m.lastReplaceTagIDs = nil
	}
	return m.err
}

func newTagUsecase(repo TagRepository) TagUsecase {
	tel := observability.NewTelemetry(nil, nil, nil)
	return NewTagUsecase(repo, tel)
}

// --- Create ---

func TestTagUsecase_Create_Success(t *testing.T) {
	tag := &domain.Tag{ID: uuid.New(), UserID: "u1", Name: "landscape"}
	uc := newTagUsecase(&mockTagRepository{tag: tag})

	result, err := uc.Create(context.Background(), "u1", "landscape")

	require.NoError(t, err)
	assert.Equal(t, "landscape", result.Name)
}

func TestTagUsecase_Create_EmptyName(t *testing.T) {
	uc := newTagUsecase(&mockTagRepository{})

	_, err := uc.Create(context.Background(), "u1", "   ")

	require.ErrorIs(t, err, ErrInvalidTagName)
}

func TestTagUsecase_Create_DuplicateName(t *testing.T) {
	uc := newTagUsecase(&mockTagRepository{err: errors.New("ERROR: duplicate key value violates unique constraint (SQLSTATE 23505)")})

	_, err := uc.Create(context.Background(), "u1", "nature")

	require.ErrorIs(t, err, ErrDuplicateTagName)
}

func TestTagUsecase_Create_RepositoryError(t *testing.T) {
	uc := newTagUsecase(&mockTagRepository{err: errors.New("db error")})

	_, err := uc.Create(context.Background(), "u1", "travel")

	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrInvalidTagName)
	assert.NotErrorIs(t, err, ErrDuplicateTagName)
}

// --- List ---

func TestTagUsecase_List_Success(t *testing.T) {
	tags := []*domain.Tag{
		{ID: uuid.New(), Name: "art"},
		{ID: uuid.New(), Name: "travel"},
	}
	uc := newTagUsecase(&mockTagRepository{tags: tags})

	result, err := uc.List(context.Background(), "u1")

	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestTagUsecase_List_RepositoryError(t *testing.T) {
	uc := newTagUsecase(&mockTagRepository{err: errors.New("db error")})

	_, err := uc.List(context.Background(), "u1")

	require.Error(t, err)
}

// --- Update ---

func TestTagUsecase_Update_Success(t *testing.T) {
	tag := &domain.Tag{ID: uuid.New(), Name: "newname"}
	uc := newTagUsecase(&mockTagRepository{tag: tag})

	result, err := uc.Update(context.Background(), tag.ID, "u1", "newname")

	require.NoError(t, err)
	assert.Equal(t, "newname", result.Name)
}

func TestTagUsecase_Update_EmptyName(t *testing.T) {
	uc := newTagUsecase(&mockTagRepository{})

	_, err := uc.Update(context.Background(), uuid.New(), "u1", "")

	require.ErrorIs(t, err, ErrInvalidTagName)
}

func TestTagUsecase_Update_DuplicateName(t *testing.T) {
	uc := newTagUsecase(&mockTagRepository{err: errors.New("23505")})

	_, err := uc.Update(context.Background(), uuid.New(), "u1", "existing")

	require.ErrorIs(t, err, ErrDuplicateTagName)
}

func TestTagUsecase_Update_NotFound(t *testing.T) {
	uc := newTagUsecase(&mockTagRepository{err: gorm.ErrRecordNotFound})

	_, err := uc.Update(context.Background(), uuid.New(), "u1", "name")

	require.Error(t, err)
}

// --- Delete ---

func TestTagUsecase_Delete_Success(t *testing.T) {
	uc := newTagUsecase(&mockTagRepository{})

	err := uc.Delete(context.Background(), uuid.New(), "u1")

	require.NoError(t, err)
}

func TestTagUsecase_Delete_NotFound(t *testing.T) {
	uc := newTagUsecase(&mockTagRepository{err: gorm.ErrRecordNotFound})

	err := uc.Delete(context.Background(), uuid.New(), "u1")

	require.Error(t, err)
}
