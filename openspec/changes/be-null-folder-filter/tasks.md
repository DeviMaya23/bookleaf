## 1. Usecase Types

- [ ] 1.1 Add `Unfiled bool` field to `ListImagesParams` in `internal/usecase/`

## 2. Repository

- [ ] 2.1 Update `imageRepository.List` WHERE logic: when `Unfiled = true` emit `folder_id IS NULL` and skip `FolderID`; existing `FolderID` path unchanged

## 3. Handler

- [ ] 3.1 In `ListImages` handler, read `unfiled` query param — set `params.Unfiled = true` when value is `"true"`

## 4. Unit Tests

- [ ] 4.1 Handler test: `unfiled=true` → `ListImagesParams{Unfiled: true}` passed to usecase (success)
- [ ] 4.2 Handler test: `unfiled` absent → `ListImagesParams{Unfiled: false}` passed to usecase (success)
- [ ] 4.3 Usecase test: `Unfiled: true` passes unfiled filter to repo, returns only unfoldered images (success)
- [ ] 4.4 Usecase test: `Unfiled: false` with no `FolderID` passes no filter to repo (success)
