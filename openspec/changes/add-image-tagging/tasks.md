## 1. Database Migration

- [ ] 1.1 Write `000009_create_tags.up.sql` — create `tags` table with `UNIQUE(user_id, name)` and `image_tags` junction table with composite PK and CASCADE deletes on both FKs
- [ ] 1.2 Write `000009_create_tags.down.sql` — drop `image_tags` then `tags`

## 2. Tag Domain & Repository

- [ ] 2.1 Create `internal/domain/tag.go` — `Tag` struct with GORM tags, `User` belongs-to association, `Images []Image` many2many association, and `BeforeCreate` UUID hook
- [ ] 2.2 Add `Tags []Tag` many2many association field to `Image` struct in `internal/domain/image.go`
- [ ] 2.3 Create `internal/usecase/tag_repository.go` — `TagRepository` interface with `Create`, `ListByUserID`, `GetByID`, `Update`, `Delete`, `ReplaceImageTags` methods
- [ ] 2.4 Create `internal/repository/tag_repository.go` — SQL implementation of `TagRepository`; `ReplaceImageTags` deletes existing rows then inserts new ones in a single transaction
- [ ] 2.5 Write unit tests for `tagRepository` — skip (integration test only, per convention)
- [ ] 2.6 Write integration tests for `tagRepository` in `internal/repository/tag_repository_integration_test.go`

## 3. Image Repository Updates

- [ ] 3.1 Update `ImageRepository` interface in `internal/usecase/image_repository.go` — add `tagID *uuid.UUID` param to `List`; document that `List`, `GetByID`, and `Update` return images with `Tags` preloaded
- [ ] 3.2 Update `imageRepository.List` in `internal/repository/image_repository.go` — add `tagID *uuid.UUID` param, add JOIN on `image_tags` when non-nil, add `.Preload("Tags")`
- [ ] 3.3 Update `imageRepository.GetByID` in `internal/repository/image_repository.go` — add `.Preload("Tags")`
- [ ] 3.4 Update `imageRepository.Update` in `internal/repository/image_repository.go` — add `.Preload("Tags")` to the re-fetch after update
- [ ] 3.5 Update image repository integration tests to cover tag preloading and `tag_id` filter

## 4. Tag Usecase

- [ ] 4.1 Create `internal/usecase/tag_usecase.go` — `TagUsecase` interface and `tagUsecase` implementation; define `ErrInvalidTagName` and `ErrDuplicateTagName`; detect unique constraint violations from the DB error to return `ErrDuplicateTagName`
- [ ] 4.2 Write unit tests for `tagUsecase` in `internal/usecase/tag_usecase_test.go` — success and failure scenarios for `Create`, `Update`, `Delete`, `List`

## 5. Image Usecase Updates

- [ ] 5.1 Add `TagID *uuid.UUID` to `ListImagesParams` in `internal/usecase/image_pagination.go` (or wherever `ListImagesParams` is defined)
- [ ] 5.2 Add `Tags *[]uuid.UUID` to `UpdateImageParams` in `internal/usecase/image_usecase.go`
- [ ] 5.3 Inject `TagRepository` into `imageUsecase` — add to struct, constructor, and `NewImageUsecase` signature
- [ ] 5.4 Update `imageUsecase.ListImages` to pass `params.TagID` down to `imageRepo.List`
- [ ] 5.5 Update `imageUsecase.UpdateImage` — after scalar field update, if `params.Tags != nil` call `tagRepo.ReplaceImageTags`
- [ ] 5.6 Update image usecase unit tests to cover new `TagID` filter and tag replacement behaviour

## 6. Tag Handler

- [ ] 6.1 Create `internal/handler/tag.go` — `TagHandler` with `CreateTag`, `ListTags`, `UpdateTag`, `DeleteTag`; map `ErrInvalidTagName` → 400, `ErrDuplicateTagName` → 409, `gorm.ErrRecordNotFound` → 404
- [ ] 6.2 Write unit tests for `TagHandler` in `internal/handler/tag_test.go` — success and failure scenarios for all four endpoints

## 7. Image Handler Updates

- [ ] 7.1 Add `tagResponse` struct to `internal/handler/image.go` (or a shared handler types file)
- [ ] 7.2 Add `Tags []tagResponse` field to `imageResponse` and `imageDetailResponse`
- [ ] 7.3 Update `toImageResponse` to map `item.Image.Tags` to `[]tagResponse`; return empty slice when nil
- [ ] 7.4 Add `Tags json.RawMessage` field to `updateImageRequest`; parse it in `UpdateImage` handler and populate `params.Tags`
- [ ] 7.5 Add `tag_id` query param parsing to `ListImages` handler; pass as `params.TagID`
- [ ] 7.6 Update image handler unit tests to cover tags in responses, tag update parsing, and tag filter

## 8. Wiring

- [ ] 8.1 Instantiate `tagRepository`, `tagUsecase`, and `tagHandler` in `cmd/server/main.go`
- [ ] 8.2 Register tag routes on the protected Echo group: `POST /tags`, `GET /tags`, `PUT /tags/:id`, `DELETE /tags/:id`
- [ ] 8.3 Pass `tagRepository` to `NewImageUsecase`

## 9. Bruno Files

- [ ] 9.1 Create `bruno/tags/create-tag.bru`
- [ ] 9.2 Create `bruno/tags/list-tags.bru`
- [ ] 9.3 Create `bruno/tags/update-tag.bru`
- [ ] 9.4 Create `bruno/tags/delete-tag.bru`
