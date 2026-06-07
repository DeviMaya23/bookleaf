package usecase

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ImageCursor struct {
	CreatedAt time.Time
	DeletedAt *time.Time // non-nil only for trash cursors
	ID        uuid.UUID
}

type ListImagesParams struct {
	FolderID *uuid.UUID
	Unfiled  bool
	TagID    *uuid.UUID
	Name     *string
	Cursor   *ImageCursor
	Limit    int
}

type ListImagesResult struct {
	Images     []ImageItem
	NextCursor *ImageCursor
}

type ListTrashedParams struct {
	Name   *string
	Cursor *ImageCursor
	Limit  int
}

type ListTrashedResult struct {
	Images     []ImageItem
	NextCursor *ImageCursor
}

type cursorPayload struct {
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	ID        uuid.UUID  `json:"id"`
}

func EncodeCursor(c *ImageCursor) string {
	b, _ := json.Marshal(cursorPayload{CreatedAt: c.CreatedAt, DeletedAt: c.DeletedAt, ID: c.ID})
	return base64.RawURLEncoding.EncodeToString(b)
}

func DecodeCursor(s string) (*ImageCursor, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("decode cursor: %w", err)
	}
	var p cursorPayload
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("unmarshal cursor: %w", err)
	}
	return &ImageCursor{CreatedAt: p.CreatedAt, DeletedAt: p.DeletedAt, ID: p.ID}, nil
}
