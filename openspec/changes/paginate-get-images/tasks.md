## 1. Types and Interfaces

- [ ] 1.1 Add `ImageCursor`, `ListImagesParams`, `ListImagesResult`, `ListTrashedParams`, and `ListTrashedResult` types to `internal/usecase/`
- [ ] 1.2 Add cursor encode/decode helpers (`EncodeCursor`, `DecodeCursor`) in `internal/usecase/`
- [ ] 1.3 Update `ImageRepository.List` interface signature to `List(ctx, userID string, folderID *uuid.UUID, cursor *ImageCursor, limit int) ([]*domain.Image, error)`
- [ ] 1.4 Update `ImageRepository.ListTrashed` interface signature to `ListTrashed(ctx, userID string, cursor *ImageCursor, limit int) ([]*domain.Image, error)`
- [ ] 1.5 Update `ImageUsecase.ListImages` interface signature to `ListImages(ctx, userID string, params ListImagesParams) (*ListImagesResult, error)`
- [ ] 1.6 Update `ImageUsecase.ListTrashed` interface signature to `ListTrashed(ctx, userID string, params ListTrashedParams) (*ListTrashedResult, error)`

## 2. Repository

- [ ] 2.1 Update `imageRepository.List` in `internal/repository/image_repository.go` to accept `cursor *ImageCursor` and `limit int`
- [ ] 2.2 Apply keyset filter `(created_at, id) < (cursor.CreatedAt, cursor.ID)` in `List` when cursor is non-nil
- [ ] 2.3 Order `List` by `created_at DESC, id DESC` and fetch `limit + 1` rows
- [ ] 2.4 Update `imageRepository.ListTrashed` to accept `cursor *ImageCursor` and `limit int`
- [ ] 2.5 Apply keyset filter `(created_at, id) < (cursor.CreatedAt, cursor.ID)` in `ListTrashed` when cursor is non-nil
- [ ] 2.6 Order `ListTrashed` by `created_at DESC, id DESC` and fetch `limit + 1` rows

## 3. Usecase

- [ ] 3.1 Update `imageUsecase.ListImages` to accept `ListImagesParams` and return `*ListImagesResult`
- [ ] 3.2 Apply default limit (50) when `params.Limit == 0`; cap at 200 when `params.Limit > 200` in `ListImages`
- [ ] 3.3 Detect next page in `ListImages` by checking if repository returned `limit + 1` rows; trim slice to `limit`
- [ ] 3.4 Build `NextCursor` from the last item in the trimmed slice in `ListImages` when a next page exists
- [ ] 3.5 Update `imageUsecase.ListTrashed` to accept `ListTrashedParams` and return `*ListTrashedResult`
- [ ] 3.6 Apply default limit (50) when `params.Limit == 0`; cap at 200 when `params.Limit > 200` in `ListTrashed`
- [ ] 3.7 Detect next page in `ListTrashed` by checking if repository returned `limit + 1` rows; trim slice to `limit`
- [ ] 3.8 Build `NextCursor` from the last item in the trimmed slice in `ListTrashed` when a next page exists

## 4. Handler

- [ ] 4.1 Add `listImagesResponse` struct with `Images []imageResponse` and `NextCursor *string` fields (shared by both endpoints)
- [ ] 4.2 Parse `limit` query param in `ListImages` handler (default 50, silently cap at 200)
- [ ] 4.3 Parse `cursor` query param in `ListImages` handler; return `400 Bad Request` on decode failure
- [ ] 4.4 Call updated `imageUsecase.ListImages` with `ListImagesParams` and return `listImagesResponse`
- [ ] 4.5 Parse `limit` query param in `ListTrashed` handler (default 50, silently cap at 200)
- [ ] 4.6 Parse `cursor` query param in `ListTrashed` handler; return `400 Bad Request` on decode failure
- [ ] 4.7 Call updated `imageUsecase.ListTrashed` with `ListTrashedParams` and return `listImagesResponse`

## 5. Unit Tests

- [ ] 5.1 Usecase test — `ListImages` returns correct page and non-nil `NextCursor` when more results exist
- [ ] 5.2 Usecase test — `ListImages` returns nil `NextCursor` on the last page
- [ ] 5.3 Usecase test — `ListTrashed` returns correct page and non-nil `NextCursor` when more results exist
- [ ] 5.4 Usecase test — `ListTrashed` returns nil `NextCursor` on the last page
- [ ] 5.5 Handler test — `ListImages` returns `400` for invalid cursor param
- [ ] 5.6 Handler test — `ListImages` returns paginated envelope with `next_cursor` on success
- [ ] 5.7 Handler test — `ListTrashed` returns `400` for invalid cursor param
- [ ] 5.8 Handler test — `ListTrashed` returns paginated envelope with `next_cursor` on success
