package usecase

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
)

type ImageCursor struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

type ListImagesParams struct {
	FolderID *uuid.UUID
	Cursor   *ImageCursor
	Limit    int
}

type ListImagesResult struct {
	Images     []*domain.Image
	NextCursor *ImageCursor
}

type ListTrashedParams struct {
	Cursor *ImageCursor
	Limit  int
}

type ListTrashedResult struct {
	Images     []*domain.Image
	NextCursor *ImageCursor
}

type cursorPayload struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uuid.UUID `json:"id"`
}

func EncodeCursor(c *ImageCursor) string {
	b, _ := json.Marshal(cursorPayload{CreatedAt: c.CreatedAt, ID: c.ID})
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
	return &ImageCursor{CreatedAt: p.CreatedAt, ID: p.ID}, nil
}
