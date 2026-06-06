## Why

Users currently have no way to permanently delete trashed images on demand — they must wait up to 30 days for the automatic purge to run. This is a gap in control for users who want to free up storage or simply clean up immediately.

## What Changes

- New endpoint: `DELETE /images/trash` — permanently deletes all trashed images for the authenticated user (empty trash)
- New endpoint: `DELETE /images/trash/:id` — permanently deletes a single trashed image by ID for the authenticated user

Both endpoints perform the same full deletion sequence as the existing purge worker: delete R2 object, delete thumbnail if present, hard-delete DB record.

## Capabilities

### New Capabilities

- `manual-trash-delete`: Exposes user-initiated permanent deletion of trashed images via two new API endpoints — empty all trash and delete a single trashed item.

### Modified Capabilities

- `trash-purge`: `HardDelete` on `ImageRepository` is already defined here. No requirement changes — implementation is reused, not modified.

## Impact

- **Backend**: New usecase methods (`EmptyTrash`, `DeleteFromTrash`), new repository method (`HardDeleteByUser` or reuse of `HardDelete`), two new handler functions, two new routes registered in `main.go`
- **API**: Two new authenticated endpoints under `/images/trash`
- **Storage**: R2 objects (image + thumbnail) are permanently deleted on call — irreversible
- **No frontend changes** in this proposal — BE only
