## Context

`imageUsecase` currently owns both image management (CRUD, listing, folders, trash) and the upload lifecycle (initiate → complete → accept suggestion, stale cleanup). The upload flow has its own dependency set — `pendingUploadRepo`, `thumbnails`, `visionService`, `userRepo`, and three OTel metrics — none of which are needed by image management. The handler mirrors this conflation: `ImageHandler` handles both image management routes and the three upload endpoints.

Current route ownership in `ImageHandler`:
- Upload: `POST /images`, `POST /images/:id/complete`, `POST /images/:id/accept-suggestion`
- Management: `GET /images`, `GET /images/:id`, `PATCH /images/:id`, `DELETE /images/:id`, etc.

## Goals / Non-Goals

**Goals:**
- Split `imageUsecase` into `imageUsecase` (management) and `imageUploadUsecase` (upload lifecycle)
- Split `ImageHandler` into `ImageHandler` (management) and `UploadHandler` (upload lifecycle)
- Keep all route paths unchanged — no client-visible impact
- Keep test coverage for both usecases and both handlers

**Non-Goals:**
- Changing upload behavior, response shapes, or error handling
- Moving or renaming any route paths
- Touching the frontend or extension

## Decisions

### Usecase split boundary

Upload usecase takes: `pendingUploadRepo`, `imageRepo`, `folderRepo`, `userRepo`, `store`, `thumbnails`, `visionService`, `tel`.
Image usecase takes: `imageRepo`, `tagRepo`, `folderRepo`, `store`, `tel`.

Each usecase defines its own interface for every dependency it holds — including `ImageRepository` and `FolderRepository`. The upload usecase's `ImageRepository` interface declares only the methods it needs (`Create`, `SetImageFolder`, `GetByID`); the image usecase's declares the full set it needs for CRUD and listing. Both interfaces are satisfied by the same concrete `*repository.ImageRepository` at wire-up time in `main.go`. No shared state, no shared interface — just two independent contracts that happen to be fulfilled by the same type.

Upload metrics (`r2.upload.count`, `r2.thumbnail.duration`, `r2.thumbnail.count`) move entirely to `imageUploadUsecase`. `imageUsecase` needs no metrics.

### Result types

`UploadInitResult` and `CompleteUploadResult` are upload-specific. They move to `image_upload_usecase.go`. `ImageDetail`, `ImageItem`, `UpdateImageParams`, `ImageCursor` stay in `image_usecase.go`.

The handler-layer interfaces reference these types via the `usecase` package — no change needed there since they remain in the same package.

### Handler split

New `UploadHandler` struct in `internal/handler/image_upload.go`. It holds a single `UploadUsecase` interface (the three upload methods). `ImageHandler` loses those three methods and the corresponding interface entries.

```
// image_upload.go
type UploadUsecase interface {
    InitiateUpload(...) (*usecase.UploadInitResult, error)
    CompleteUpload(...) (*usecase.CompleteUploadResult, error)
    AcceptSuggestion(...) error
}

type UploadHandler struct {
    uploadUsecase UploadUsecase
    tel           *observability.Telemetry
}
```

### Router wiring

`main.go` constructs both usecases and both handlers, then registers routes:

```go
imageUsecase  := usecase.NewImageUsecase(...)
uploadUsecase := usecase.NewImageUploadUsecase(...)

imageHandler  := handler.NewImageHandler(imageUsecase, tel)
uploadHandler := handler.NewUploadHandler(uploadUsecase, tel)

// Upload routes
protected.POST("/images", uploadHandler.InitiateUpload)
protected.POST("/images/:id/complete", uploadHandler.CompleteUpload)
protected.POST("/images/:id/accept-suggestion", uploadHandler.AcceptSuggestion)

// Management routes (unchanged)
protected.GET("/images", imageHandler.ListImages)
// ...
```

`CleanupStaleUploads` is called from the background job in `main.go`. It moves to `imageUploadUsecase` — the job wiring simply switches from `imageUsecase` to `uploadUsecase`.

### AcceptSuggestion placement

`AcceptSuggestion` mechanically resembles folder assignment (which image management does), but semantically it is a response to the upload vision flow — it only exists because `CompleteUpload` may return a `SuggestedFolderName`. It belongs in the upload usecase alongside the flow it responds to.

## Risks / Trade-offs

- **Both usecases need image/folder repo methods** → not a risk; each defines its own interface scoped to what it needs, both satisfied by the same concrete type at wire-up. Standard convention.
- **More wiring in main.go** → minor increase in constructor call count. Acceptable given the clarity gained.

## Open Questions

None. The boundary is clear and the implementation is straightforward.
