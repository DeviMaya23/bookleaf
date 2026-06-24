package usecase_test

import (
	"testing"
	"time"

	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveSort(t *testing.T) {
	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name             string
		sortField        *string
		direction        *string
		expectedCol      string
		expectedClause   string
		expectedOperator string
		expectedDefDir   string
	}{
		{
			name:             "default when both nil",
			sortField:        nil,
			direction:        nil,
			expectedCol:      "created_at",
			expectedClause:   "created_at DESC, id DESC",
			expectedOperator: "<",
			expectedDefDir:   "desc",
		},
		{
			name:             "created_at descending",
			sortField:        strPtr("created_at"),
			direction:        strPtr("desc"),
			expectedCol:      "created_at",
			expectedClause:   "created_at DESC, id DESC",
			expectedOperator: "<",
			expectedDefDir:   "desc",
		},
		{
			name:             "created_at ascending",
			sortField:        strPtr("created_at"),
			direction:        strPtr("asc"),
			expectedCol:      "created_at",
			expectedClause:   "created_at ASC, id ASC",
			expectedOperator: ">",
			expectedDefDir:   "desc",
		},
		{
			name:             "title ascending",
			sortField:        strPtr("title"),
			direction:        strPtr("asc"),
			expectedCol:      "title",
			expectedClause:   "title ASC, id ASC",
			expectedOperator: ">",
			expectedDefDir:   "asc",
		},
		{
			name:             "title descending",
			sortField:        strPtr("title"),
			direction:        strPtr("desc"),
			expectedCol:      "title",
			expectedClause:   "title DESC, id DESC",
			expectedOperator: "<",
			expectedDefDir:   "asc",
		},
		{
			name:             "title with no direction defaults to asc",
			sortField:        strPtr("title"),
			direction:        nil,
			expectedCol:      "title",
			expectedClause:   "title ASC, id ASC",
			expectedOperator: ">",
			expectedDefDir:   "asc",
		},
		{
			name:             "title with empty direction defaults to asc",
			sortField:        strPtr("title"),
			direction:        strPtr(""),
			expectedCol:      "title",
			expectedClause:   "title ASC, id ASC",
			expectedOperator: ">",
			expectedDefDir:   "asc",
		},
		{
			name:             "deleted_at descending",
			sortField:        strPtr("deleted_at"),
			direction:        strPtr("desc"),
			expectedCol:      "deleted_at",
			expectedClause:   "deleted_at DESC, id DESC",
			expectedOperator: "<",
			expectedDefDir:   "desc",
		},
		{
			name:             "deleted_at ascending",
			sortField:        strPtr("deleted_at"),
			direction:        strPtr("asc"),
			expectedCol:      "deleted_at",
			expectedClause:   "deleted_at ASC, id ASC",
			expectedOperator: ">",
			expectedDefDir:   "desc",
		},
		{
			name:             "deleted_at with no direction defaults to desc",
			sortField:        strPtr("deleted_at"),
			direction:        nil,
			expectedCol:      "deleted_at",
			expectedClause:   "deleted_at DESC, id DESC",
			expectedOperator: "<",
			expectedDefDir:   "desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := usecase.ResolveSort(tt.sortField, tt.direction)
			assert.Equal(t, tt.expectedCol, res.Column)
			assert.Equal(t, tt.expectedClause, res.OrderClause)
			assert.Equal(t, tt.expectedOperator, res.WhereOperator)
			assert.Equal(t, tt.expectedDefDir, res.DefaultDirection)
		})
	}
}

func TestCursorRoundTrip(t *testing.T) {
	t.Run("created_at cursor", func(t *testing.T) {
		orig := &usecase.ImageCursor{
			CreatedAt: time.Now().Truncate(time.Millisecond).UTC(),
			ID:        uuid.New(),
		}
		encoded := usecase.EncodeCursor(orig)
		decoded, err := usecase.DecodeCursor(encoded)
		require.NoError(t, err)
		assert.Equal(t, orig.CreatedAt, decoded.CreatedAt)
		assert.Nil(t, decoded.Title)
		assert.Nil(t, decoded.DeletedAt)
		assert.Equal(t, orig.ID, decoded.ID)
	})

	t.Run("title cursor", func(t *testing.T) {
		title := "my title"
		orig := &usecase.ImageCursor{
			CreatedAt: time.Now().Truncate(time.Millisecond).UTC(),
			Title:     &title,
			ID:        uuid.New(),
		}
		encoded := usecase.EncodeCursor(orig)
		decoded, err := usecase.DecodeCursor(encoded)
		require.NoError(t, err)
		assert.Equal(t, orig.CreatedAt, decoded.CreatedAt)
		require.NotNil(t, decoded.Title)
		assert.Equal(t, *orig.Title, *decoded.Title)
		assert.Nil(t, decoded.DeletedAt)
		assert.Equal(t, orig.ID, decoded.ID)
	})
}
