package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
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

// --- CreateTag ---

func TestTagHandler_CreateTag(t *testing.T) {
	tagID := uuid.New()
	now := time.Now().UTC()

	tests := []struct {
		name          string
		body          string
		uc            *mockTagUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name: "returns 201 with tag body",
			body: `{"name":"landscape"}`,
			uc: &mockTagUsecase{
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
			name:          "returns 400 for invalid name",
			body:          `{"name":""}`,
			uc:            &mockTagUsecase{err: usecase.ErrInvalidTagName},
			wantErrStatus: http.StatusBadRequest,
		},
		{
			name:          "returns 409 on duplicate name",
			body:          `{"name":"landscape"}`,
			uc:            &mockTagUsecase{err: usecase.ErrDuplicateTagName},
			wantErrStatus: http.StatusConflict,
		},
		{
			name:          "returns 500 on generic error",
			body:          `{"name":"landscape"}`,
			uc:            &mockTagUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTagHandler(tt.uc, observability.NewTelemetry(nil, nil, nil))
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

func TestTagHandler_CreateTag_MalformedJSON(t *testing.T) {
	h := NewTagHandler(&mockTagUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPost, "/tags", `{not valid json}`)

	err := h.CreateTag(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

// --- ListTags ---

func TestTagHandler_ListTags_ReturnsList(t *testing.T) {
	h := NewTagHandler(&mockTagUsecase{
		tags: []*domain.Tag{
			{ID: uuid.New(), Name: "travel"},
			{ID: uuid.New(), Name: "design"},
		},
	}, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodGet, "/tags", "")

	err := h.ListTags(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)
}

func TestTagHandler_ListTags_GenericError(t *testing.T) {
	h := NewTagHandler(&mockTagUsecase{err: errors.New("db error")}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/tags", "")

	err := h.ListTags(c)

	assertHTTPError(t, err, http.StatusInternalServerError)
}

// --- UpdateTag ---

func TestTagHandler_UpdateTag(t *testing.T) {
	tagID := uuid.New()
	now := time.Now().UTC()

	tests := []struct {
		name          string
		body          string
		uc            *mockTagUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name:       "returns 200 with updated tag",
			body:       `{"name":"renamed"}`,
			uc:         &mockTagUsecase{tag: &domain.Tag{ID: tagID, Name: "renamed", CreatedAt: now, UpdatedAt: now}},
			wantStatus: http.StatusOK,
		},
		{
			name:          "returns 400 for invalid name",
			body:          `{"name":""}`,
			uc:            &mockTagUsecase{err: usecase.ErrInvalidTagName},
			wantErrStatus: http.StatusBadRequest,
		},
		{
			name:          "returns 409 on duplicate name",
			body:          `{"name":"travel"}`,
			uc:            &mockTagUsecase{err: usecase.ErrDuplicateTagName},
			wantErrStatus: http.StatusConflict,
		},
		{
			name:          "returns 404 when tag not found",
			body:          `{"name":"renamed"}`,
			uc:            &mockTagUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
		{
			name:          "returns 500 on generic error",
			body:          `{"name":"renamed"}`,
			uc:            &mockTagUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTagHandler(tt.uc, observability.NewTelemetry(nil, nil, nil))
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

func TestTagHandler_UpdateTag_InvalidUUID(t *testing.T) {
	h := NewTagHandler(&mockTagUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPut, "/tags/not-a-uuid", `{"name":"renamed"}`)
	c.SetPath("/tags/:id")
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.UpdateTag(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestTagHandler_UpdateTag_MalformedJSON(t *testing.T) {
	tagID := uuid.New()
	h := NewTagHandler(&mockTagUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPut, "/tags/"+tagID.String(), `{not valid json}`)
	c.SetPath("/tags/:id")
	c.SetParamNames("id")
	c.SetParamValues(tagID.String())

	err := h.UpdateTag(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

// --- DeleteTag ---

func TestTagHandler_DeleteTag(t *testing.T) {
	tagID := uuid.New()

	tests := []struct {
		name          string
		uc            *mockTagUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name:       "returns 204 on success",
			uc:         &mockTagUsecase{},
			wantStatus: http.StatusNoContent,
		},
		{
			name:          "returns 404 when tag not found",
			uc:            &mockTagUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
		{
			name:          "returns 500 on generic error",
			uc:            &mockTagUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTagHandler(tt.uc, observability.NewTelemetry(nil, nil, nil))
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

func TestTagHandler_DeleteTag_InvalidUUID(t *testing.T) {
	h := NewTagHandler(&mockTagUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodDelete, "/tags/not-a-uuid", "")
	c.SetPath("/tags/:id")
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.DeleteTag(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}
