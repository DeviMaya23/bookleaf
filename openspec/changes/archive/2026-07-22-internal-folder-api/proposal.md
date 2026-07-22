## Why

Another backend application needs to query Bookleaf's folder data (public folder lists, folder contents with presigned image URLs, and folder public status) on behalf of its own users. These queries must bypass Kinde user auth since the caller is a trusted service, not an end user.

## What Changes

- Add a shared-secret middleware (`X-Bookleaf-Internal-Secret` header) for internal route protection
- Add `INTERNAL_API_SECRET` env var to config
- Register a new `/internal` route group gated by the new middleware
- Add three internal-only endpoints:
  - `GET /internal/users/:user_id/public-folders` — lists all folders a user has made public, with their share tokens
  - `GET /internal/folders/:folder_id/contents` — returns folder contents with presigned image URLs (same data shape as `/share/:token`)
  - `GET /internal/folders/:folder_id/status` — returns 200 + token if folder is public, 404 otherwise
- Add `ListByUserID` and `GetByFolderIDWithFolder` methods to `FolderShareRepository`
- Add `GetPublicFoldersByUser` and `GetSharedFolderByFolderID` methods to `shareUsecase`
- Add a new `InternalHandler` in the handler layer

## Capabilities

### New Capabilities

- `internal-folder-api`: Internal service-to-service endpoints for querying public folder data, protected by a shared secret header

### Modified Capabilities

- `folder-sharing`: New repository methods `ListByUserID` and `GetByFolderIDWithFolder` extend the data access layer (no requirement changes, implementation only)

## Impact

- **Backend**: new middleware file, new handler file, new usecase methods, two new repository interface methods and their implementations, config/env change
- **No frontend or extension impact**
- **Bruno**: new collection file for the three internal endpoints
- **Env**: `INTERNAL_API_SECRET` must be provisioned in all environments
