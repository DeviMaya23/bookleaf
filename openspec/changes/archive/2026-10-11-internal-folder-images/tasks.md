# Tasks

## 1. Usecase — folderImagesUsecase

- [x] 1.1 Create `backend/internal/usecase/folder_images_usecase.go` with `FolderImagesUsecase` interface, `FolderImageThumbnail` struct, and `folderImagesUsecase` implementation; verify it compiles with `go build ./...`
- [x] 1.2 Implement `ListFolderImages`: call `FolderRepository.GetByID` for ownership check, call `ImageRepository.ListByFolder`, presign each non-nil `ThumbnailPath` with `presignedGetTTL`, exclude nil-path images; verify the unit tests in task 1.3 pass
- [x] 1.3 Write unit tests in `backend/internal/usecase/folder_images_usecase_test.go` covering: returns thumbnails for owned folder, excludes images without thumbnail path, empty folder returns empty slice, unowned folder returns `gorm.ErrRecordNotFound`; verify `go test ./internal/usecase/...` passes

## 2. Handler — InternalHandler extension

- [x] 2.1 Add `InternalFolderImagesUsecase` interface to `backend/internal/handler/internal.go`, add the field to `InternalHandler`, and update `NewInternalHandler` to accept it; verify `go build ./...` compiles
- [x] 2.2 Implement `ListFolderImages(c echo.Context) error`: parse `:user_id` as IDP subject → resolve via `UserResolver` (404 on `ErrUserNotFound`), parse `:folder_id` UUID (400 on invalid), call usecase (404 on `gorm.ErrRecordNotFound`), return `200` with `{ "images": [...] }`; verify the unit tests in task 2.3 pass
- [x] 2.3 Write unit tests in `backend/internal/handler/internal_test.go` covering: 200 with populated list, 200 with empty list, 404 on `gorm.ErrRecordNotFound`, 404 on `ErrUserNotFound`, 400 on invalid folder UUID; verify `go test ./internal/handler/...` passes

## 3. Route registration and wiring

- [x] 3.1 Instantiate `folderImagesUsecase` in `backend/cmd/server/main.go` (passing `folderRepo`, `imageRepo`, `storageService`, `tel`), pass it as the new argument to `NewInternalHandler`, and add `internalGroup.GET("/users/:user_id/folders/:folder_id/images", internalHandler.ListFolderImages)`; verify `go build ./cmd/server/...` compiles
- [x] 3.2 Create `bruno/internal/list-folder-images.bru` following the existing Bruno file format with `X-Bookleaf-Internal-Secret` header, `{{userId}}` and `{{folderId}}` path vars; verify the file renders correctly in Bruno

## 4. Lint

- [x] 4.1 Run `golangci-lint run ./...` from `backend/` and fix any issues; verify lint exits clean

## Workflow follow-up

- Archive the change once reviewed.
