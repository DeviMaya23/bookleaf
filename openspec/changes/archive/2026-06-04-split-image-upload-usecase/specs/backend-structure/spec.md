## ADDED Requirements

### Requirement: Upload lifecycle lives in its own usecase
The upload lifecycle (initiate, complete, accept suggestion, stale cleanup) SHALL reside in a dedicated `imageUploadUsecase` and not be part of `imageUsecase`.

#### Scenario: Upload usecase is separate from image usecase
- **WHEN** the codebase is built
- **THEN** `internal/usecase/image_upload_usecase.go` exists and `imageUsecase` contains no upload methods

### Requirement: Upload handler lives in its own handler struct
The upload lifecycle endpoints SHALL be handled by a dedicated `UploadHandler` and not by `ImageHandler`.

#### Scenario: Upload handler is separate from image handler
- **WHEN** the codebase is built
- **THEN** `internal/handler/image_upload.go` exists and `ImageHandler` contains no upload endpoint methods

### Requirement: Each usecase defines its own repository interfaces
Where two usecases depend on the same underlying repository type, each SHALL declare its own interface scoped to only the methods it needs. The same concrete repository type satisfies both interfaces at wire-up time.

#### Scenario: Upload usecase defines its own ImageRepository interface
- **WHEN** the codebase is built
- **THEN** `imageUploadUsecase` depends on an `ImageRepository` interface declared within `image_upload_usecase.go`, not the one declared in `image_usecase.go`
