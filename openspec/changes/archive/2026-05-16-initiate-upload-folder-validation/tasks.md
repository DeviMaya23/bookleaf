## 1. Usecase — Folder Validation in InitiateUpload

- [x] 1.1 In `imageUsecase.InitiateUpload`, when `folderID != nil`, call `folderRepo.GetByID(ctx, *folderID, userID)`; if the folder is not found (any error), set `folderID = nil` before creating the image record

## 2. Unit Tests — Usecase

- [x] 2.1 Add unit test: `InitiateUpload` with a valid `folder_id` — folder repo returns the folder → image is created with that `folder_id`
- [x] 2.2 Add unit test: `InitiateUpload` with a `folder_id` that is not found — folder repo returns error → image is created with `folder_id = null`
