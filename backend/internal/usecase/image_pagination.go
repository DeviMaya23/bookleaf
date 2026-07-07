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
	Title     *string    // non-nil only when the active sort orders by title
	DeletedAt *time.Time // non-nil only for trash cursors
	ID        uuid.UUID
}

type ListImagesParams struct {
	Unfiled      bool
	FolderIDs    []uuid.UUID
	TagIDs       []uuid.UUID
	MIMETypes    []string
	Name         *string
	SearchLabels bool
	Sort         *string
	Direction    *string
	Cursor       *ImageCursor
	Limit        int
}

type ListImagesResult struct {
	Images     []ImageItem
	NextCursor *ImageCursor
}

type ListTrashedParams struct {
	Name      *string
	Sort      *string
	Direction *string
	Cursor    *ImageCursor
	Limit     int
}

type ListTrashedResult struct {
	Images     []ImageItem
	NextCursor *ImageCursor
}

type cursorPayload struct {
	CreatedAt time.Time  `json:"created_at"`
	Title     *string    `json:"title,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	ID        uuid.UUID  `json:"id"`
}

func EncodeCursor(c *ImageCursor) string {
	b, _ := json.Marshal(cursorPayload{CreatedAt: c.CreatedAt, Title: c.Title, DeletedAt: c.DeletedAt, ID: c.ID})
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
	return &ImageCursor{CreatedAt: p.CreatedAt, Title: p.Title, DeletedAt: p.DeletedAt, ID: p.ID}, nil
}

// SortDispatchResult contains the SQL fragments needed for pagination and ordering.
type SortDispatchResult struct {
	Column           string
	OrderClause      string
	WhereOperator    string
	DefaultDirection string
}

// ResolveSort determines the SQL components for the given sort field and direction.
// It maps the allow-listed (field, direction) pair to its column, ORDER BY clause,
// keyset operator (> for asc, < for desc) and resolves the default direction if unspecified.
func ResolveSort(sortField *string, direction *string) SortDispatchResult {
	var field, dir string
	if sortField != nil {
		field = *sortField
	}
	if direction != nil {
		dir = *direction
	}

	switch field {
	case "title":
		if dir == "" {
			dir = "asc"
		}
		if dir == "asc" {
			return SortDispatchResult{Column: "title", OrderClause: "title ASC, id ASC", WhereOperator: ">", DefaultDirection: "asc"}
		}
		return SortDispatchResult{Column: "title", OrderClause: "title DESC, id DESC", WhereOperator: "<", DefaultDirection: "asc"}
	case "deleted_at":
		// Trash-only field (GET /images/trash); GET /images never passes this.
		if dir == "" {
			dir = "desc"
		}
		if dir == "asc" {
			return SortDispatchResult{Column: "deleted_at", OrderClause: "deleted_at ASC, id ASC", WhereOperator: ">", DefaultDirection: "desc"}
		}
		return SortDispatchResult{Column: "deleted_at", OrderClause: "deleted_at DESC, id DESC", WhereOperator: "<", DefaultDirection: "desc"}
	default:
		// Default to created_at DESC
		if dir == "" {
			dir = "desc"
		}
		if dir == "asc" {
			return SortDispatchResult{Column: "created_at", OrderClause: "created_at ASC, id ASC", WhereOperator: ">", DefaultDirection: "desc"}
		}
		return SortDispatchResult{Column: "created_at", OrderClause: "created_at DESC, id DESC", WhereOperator: "<", DefaultDirection: "desc"}
	}
}
