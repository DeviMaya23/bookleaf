## Why

The `InitiateUpload` endpoint already accepts a `folder_id` parameter and saves it directly — but it performs no validation. If the provided `folder_id` does not exist or does not belong to the authenticated user, the image is inserted with a dangling or unauthorized reference (or fails at the DB FK constraint). This change makes the behavior safe and predictable by falling back to `null` when the folder cannot be found.

## What Changes

- When `folder_id` is provided in the `InitiateUpload` request, the usecase will look up the folder by ID scoped to the requesting user.
- If the folder is not found (or does not belong to the user), `folder_id` is set to `null` before inserting the image record — no error is returned to the caller.
- If `folder_id` is omitted or explicitly `null`, behavior is unchanged (image created without a folder).

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `image-endpoints`: `InitiateUpload` now validates `folder_id` and silently falls back to `null` if the folder is not found.

## Impact

- **Usecase**: `imageUsecase.InitiateUpload` — adds a `folderRepo.GetByID` call when `folderID != nil`.
- **No API contract change**: request/response shapes are unchanged; the behavior change is internal.
- **Tests**: unit tests for `InitiateUpload` usecase need scenarios for valid folder, not-found folder, and nil folder.
