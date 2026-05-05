## MODIFIED Requirements

### Requirement: Image Repository Interface

The system SHALL define an `ImageRepository` interface in `internal/usecase/` that the SQL repository implements.

Methods:
- `Create(ctx, image *domain.Image) (*domain.Image, error)`
- `List(ctx, userID string, folderID *uuid.UUID) ([]*domain.Image, error)` — returns non-deleted images; `folderID` nil means no filter
- `GetByID(ctx, id uuid.UUID, userID string) (*domain.Image, error)` — returns non-deleted images only
- `GetDeletedByID(ctx, id uuid.UUID, userID string) (*domain.Image, error)` — returns soft-deleted images only
- `UpdateThumbnailPath(ctx, id uuid.UUID, thumbnailPath string) error` — updates `thumbnail_path`; no ownership check (called internally by goroutine)
- `Update(ctx, id uuid.UUID, userID string, fields map[string]any) (*domain.Image, error)` — selectively updates the supplied fields for the image owned by `userID`
- `SoftDelete(ctx, id uuid.UUID, userID string) error`
- `Restore(ctx, id uuid.UUID, userID string) error`
- `ListTrashed(ctx, userID string) ([]*domain.Image, error)`

#### Scenario: Repository interface is satisfied by SQL implementation

- **WHEN** the Go package is compiled
- **THEN** `imageRepository` in `internal/repository/` implements `usecase.ImageRepository` without compilation errors

---

### Requirement: Image Routes Wiring

The system SHALL register image routes on the protected Echo group in `main.go`.

Routes:
- `POST /images`
- `POST /images/:id/complete`
- `GET /images`
- `GET /images/:id`
- `PATCH /images/:id`
- `DELETE /images/:id`
- `GET /images/trash`
- `POST /images/:id/restore`

#### Scenario: Image routes are registered under auth middleware

- **WHEN** the server starts
- **THEN** all `/images` routes require a valid Kinde Bearer token
- **AND** unauthenticated requests return `401 Unauthorized`

---

### Requirement: Image Usecase Interface

The system SHALL define an `ImageUsecase` interface in `internal/usecase/` with methods corresponding to the HTTP operations, including `UpdateImage`.

#### Scenario: Usecase interface is satisfied by concrete implementation

- **WHEN** the Go package is compiled
- **THEN** `imageUsecase` implements `usecase.ImageUsecase` without compilation errors
