package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/observability"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type mockTagUsecase struct {
	tag  *domain.Tag
	tags []*domain.Tag
	err  error
}

func (m *mockTagUsecase) Create(_ context.Context, _ string, _ string) (*domain.Tag, error) {
	return m.tag, m.err
}

func (m *mockTagUsecase) List(_ context.Context, _ string) ([]*domain.Tag, error) {
	return m.tags, m.err
}

func (m *mockTagUsecase) Update(_ context.Context, _ uuid.UUID, _ string, _ string) (*domain.Tag, error) {
	return m.tag, m.err
}

func (m *mockTagUsecase) Delete(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}

func TestTagHandler_CreateTag(t *testing.T) {
	tagID := uuid.New()
	now := time.Now().UTC()

	tests := []struct {
		name          string
		body          string
		mockUC        *mockTagUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name: "creates tag and returns 201",
			body: `{"name":"landscape"}`,
			mockUC: &mockTagUsecase{
				tag: &domain.Tag{
					ID:        tagID,
					Name:      "landscape",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:          "returns 400 on invalid name",
			body:          `{"name":""}`,
			mockUC:        &mockTagUsecase{err: usecase.ErrInvalidTagName},
			wantErrStatus: http.StatusBadRequest,
		},
		{
			name:          "returns 409 on duplicate name",
			body:          `{"name":"landscape"}`,
			mockUC:        &mockTagUsecase{err: usecase.ErrDuplicateTagName},
			wantErrStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTagHandler(tt.mockUC, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodPost, "/tags", tt.body)

			err := h.CreateTag(c)

			if tt.wantErrStatus != 0 {
				assertHTTPError(t, err, tt.wantErrStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)

			var resp map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, tagID.String(), resp["id"])
			assert.Equal(t, "landscape", resp["name"])
			_, hasCreatedAt := resp["created_at"]
			_, hasUpdatedAt := resp["updated_at"]
			assert.True(t, hasCreatedAt)
			assert.True(t, hasUpdatedAt)
		})
	}
}

func TestTagHandler_ListTags(t *testing.T) {
	tests := []struct {
		name          string
		mockUC        *mockTagUsecase
		wantStatus    int
		wantLen       int
		wantErrStatus int
	}{
		{
			name: "returns tags list",
			mockUC: &mockTagUsecase{
				tags: []*domain.Tag{
					{ID: uuid.New(), Name: "travel"},
					{ID: uuid.New(), Name: "design"},
				},
			},
			wantStatus: http.StatusOK,
			wantLen:    2,
		},
		{
			name:          "returns 500 on usecase error",
			mockUC:        &mockTagUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTagHandler(tt.mockUC, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodGet, "/tags", "")

			err := h.ListTags(c)

			if tt.wantErrStatus != 0 {
				assertHTTPError(t, err, tt.wantErrStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)

			var resp []map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Len(t, resp, tt.wantLen)
		})
	}
}

func TestTagHandler_UpdateTag(t *testing.T) {
	tagID := uuid.New()
	now := time.Now().UTC()

	tests := []struct {
		name          string
		body          string
		mockUC        *mockTagUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name: "updates tag and returns 200",
			body: `{"name":"renamed"}`,
			mockUC: &mockTagUsecase{
				tag: &domain.Tag{
					ID:        tagID,
					Name:      "renamed",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:          "returns 404 when tag not found",
			body:          `{"name":"renamed"}`,
			mockUC:        &mockTagUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
		{
			name:          "returns 409 on duplicate name",
			body:          `{"name":"travel"}`,
			mockUC:        &mockTagUsecase{err: usecase.ErrDuplicateTagName},
			wantErrStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTagHandler(tt.mockUC, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodPut, "/tags/"+tagID.String(), tt.body)
			c.SetPath("/tags/:id")
			c.SetParamNames("id")
			c.SetParamValues(tagID.String())

			err := h.UpdateTag(c)

			if tt.wantErrStatus != 0 {
				assertHTTPError(t, err, tt.wantErrStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)

			var resp map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, tagID.String(), resp["id"])
			assert.Equal(t, "renamed", resp["name"])
		})
	}
}

func TestTagHandler_DeleteTag(t *testing.T) {
	tagID := uuid.New()

	tests := []struct {
		name          string
		mockUC        *mockTagUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name:       "deletes tag and returns 204",
			mockUC:     &mockTagUsecase{},
			wantStatus: http.StatusNoContent,
		},
		{
			name:          "returns 404 when tag not found",
			mockUC:        &mockTagUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTagHandler(tt.mockUC, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodDelete, "/tags/"+tagID.String(), "")
			c.SetPath("/tags/:id")
			c.SetParamNames("id")
			c.SetParamValues(tagID.String())

			err := h.DeleteTag(c)

			if tt.wantErrStatus != 0 {
				assertHTTPError(t, err, tt.wantErrStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
