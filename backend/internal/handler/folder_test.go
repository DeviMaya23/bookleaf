package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	authmw "github.com/devi/bookleaf/internal/middleware"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type mockFolderUsecase struct {
	folder          *domain.Folder
	folders         []*domain.Folder
	err             error
	createdUserID   string
	createdName     string
	createdParentID *uuid.UUID
	listedUserID    string
	getID           uuid.UUID
	getUserID       string
	updateID        uuid.UUID
	updateUserID    string
	updateName      string
	updateParentID  *uuid.UUID
	deleteID        uuid.UUID
	deleteUserID    string
}

func (m *mockFolderUsecase) Create(_ context.Context, userID, name string, parentID *uuid.UUID) (*domain.Folder, error) {
	m.createdUserID = userID
	m.createdName = name
	m.createdParentID = parentID
	return m.folder, m.err
}

func (m *mockFolderUsecase) List(_ context.Context, userID string) ([]*domain.Folder, error) {
	m.listedUserID = userID
	return m.folders, m.err
}

func (m *mockFolderUsecase) GetByID(_ context.Context, id uuid.UUID, userID string) (*domain.Folder, error) {
	m.getID = id
	m.getUserID = userID
	return m.folder, m.err
}

func (m *mockFolderUsecase) Update(_ context.Context, id uuid.UUID, userID, name string, parentID *uuid.UUID) (*domain.Folder, error) {
	m.updateID = id
	m.updateUserID = userID
	m.updateName = name
	m.updateParentID = parentID
	return m.folder, m.err
}

func (m *mockFolderUsecase) Delete(_ context.Context, id uuid.UUID, userID string) error {
	m.deleteID = id
	m.deleteUserID = userID
	return m.err
}

func TestFolderHandler_CreateFolder_HappyPath(t *testing.T) {
	parentID := uuid.New()
	folderID := uuid.New()
	now := time.Now().UTC()
	mockUC := &mockFolderUsecase{
		folder: &domain.Folder{
			ID:        folderID,
			UserID:    "kp_abc123",
			Name:      "travel",
			ParentID:  &parentID,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	h := NewFolderHandler(mockUC)

	body := []byte(`{"name":"travel","parent_id":"` + parentID.String() + `"}`)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/folders", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(string(authmw.AuthenticatedUserIDContextKey), "kp_abc123")

	if err := h.CreateFolder(c); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp["id"] != folderID.String() {
		t.Fatalf("expected id %s, got %v", folderID, resp["id"])
	}
	if mockUC.createdUserID != "kp_abc123" {
		t.Fatalf("expected create user id kp_abc123, got %s", mockUC.createdUserID)
	}
}

func TestFolderHandler_CreateFolder_ErrorPath(t *testing.T) {
	mockUC := &mockFolderUsecase{
		err: usecase.ErrInvalidFolderName,
	}
	h := NewFolderHandler(mockUC)

	body := []byte(`{"name":""}`)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/folders", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(string(authmw.AuthenticatedUserIDContextKey), "kp_abc123")

	err := h.CreateFolder(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, httpErr.Code)
	}
}

func TestFolderHandler_ListFolders_HappyPath(t *testing.T) {
	mockUC := &mockFolderUsecase{
		folders: []*domain.Folder{
			{
				ID:        uuid.New(),
				UserID:    "kp_abc123",
				Name:      "travel",
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
		},
	}
	h := NewFolderHandler(mockUC)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/folders", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(string(authmw.AuthenticatedUserIDContextKey), "kp_abc123")

	if err := h.ListFolders(c); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("expected 1 folder, got %d", len(resp))
	}
	if mockUC.listedUserID != "kp_abc123" {
		t.Fatalf("expected list user id kp_abc123, got %s", mockUC.listedUserID)
	}
}

func TestFolderHandler_ListFolders_ErrorPath(t *testing.T) {
	mockUC := &mockFolderUsecase{
		err: errors.New("db error"),
	}
	h := NewFolderHandler(mockUC)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/folders", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(string(authmw.AuthenticatedUserIDContextKey), "kp_abc123")

	err := h.ListFolders(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, httpErr.Code)
	}
}

func TestFolderHandler_GetFolder_HappyPath(t *testing.T) {
	folderID := uuid.New()
	mockUC := &mockFolderUsecase{
		folder: &domain.Folder{
			ID:        folderID,
			UserID:    "kp_abc123",
			Name:      "travel",
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
	}
	h := NewFolderHandler(mockUC)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/folders/"+folderID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/folders/:id")
	c.SetParamNames("id")
	c.SetParamValues(folderID.String())
	c.Set(string(authmw.AuthenticatedUserIDContextKey), "kp_abc123")

	if err := h.GetFolder(c); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if mockUC.getID != folderID {
		t.Fatalf("expected get id %s, got %s", folderID, mockUC.getID)
	}
}

func TestFolderHandler_GetFolder_ErrorPath(t *testing.T) {
	folderID := uuid.New()
	mockUC := &mockFolderUsecase{
		err: gorm.ErrRecordNotFound,
	}
	h := NewFolderHandler(mockUC)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/folders/"+folderID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/folders/:id")
	c.SetParamNames("id")
	c.SetParamValues(folderID.String())
	c.Set(string(authmw.AuthenticatedUserIDContextKey), "kp_abc123")

	err := h.GetFolder(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, httpErr.Code)
	}
}

func TestFolderHandler_UpdateFolder_HappyPath(t *testing.T) {
	folderID := uuid.New()
	parentID := uuid.New()
	mockUC := &mockFolderUsecase{
		folder: &domain.Folder{
			ID:        folderID,
			UserID:    "kp_abc123",
			Name:      "updated",
			ParentID:  &parentID,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
	}
	h := NewFolderHandler(mockUC)

	body := []byte(`{"name":"updated","parent_id":"` + parentID.String() + `"}`)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/folders/"+folderID.String(), bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/folders/:id")
	c.SetParamNames("id")
	c.SetParamValues(folderID.String())
	c.Set(string(authmw.AuthenticatedUserIDContextKey), "kp_abc123")

	if err := h.UpdateFolder(c); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if mockUC.updateID != folderID {
		t.Fatalf("expected update id %s, got %s", folderID, mockUC.updateID)
	}
}

func TestFolderHandler_UpdateFolder_ErrorPath(t *testing.T) {
	folderID := uuid.New()
	mockUC := &mockFolderUsecase{
		err: usecase.ErrInvalidFolderName,
	}
	h := NewFolderHandler(mockUC)

	body := []byte(`{"name":""}`)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/folders/"+folderID.String(), bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/folders/:id")
	c.SetParamNames("id")
	c.SetParamValues(folderID.String())
	c.Set(string(authmw.AuthenticatedUserIDContextKey), "kp_abc123")

	err := h.UpdateFolder(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, httpErr.Code)
	}
}

func TestFolderHandler_DeleteFolder_HappyPath(t *testing.T) {
	folderID := uuid.New()
	mockUC := &mockFolderUsecase{}
	h := NewFolderHandler(mockUC)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/folders/"+folderID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/folders/:id")
	c.SetParamNames("id")
	c.SetParamValues(folderID.String())
	c.Set(string(authmw.AuthenticatedUserIDContextKey), "kp_abc123")

	if err := h.DeleteFolder(c); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
	if mockUC.deleteID != folderID {
		t.Fatalf("expected delete id %s, got %s", folderID, mockUC.deleteID)
	}
}

func TestFolderHandler_DeleteFolder_ErrorPath(t *testing.T) {
	folderID := uuid.New()
	mockUC := &mockFolderUsecase{
		err: gorm.ErrRecordNotFound,
	}
	h := NewFolderHandler(mockUC)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/folders/"+folderID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/folders/:id")
	c.SetParamNames("id")
	c.SetParamValues(folderID.String())
	c.Set(string(authmw.AuthenticatedUserIDContextKey), "kp_abc123")

	err := h.DeleteFolder(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, httpErr.Code)
	}
}
