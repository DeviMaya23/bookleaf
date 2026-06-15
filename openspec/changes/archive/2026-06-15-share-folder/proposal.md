## Why

Users want to share the contents of a folder with people who don't have a Bookleaf account, via a stable link that shows a read-only gallery view. Today there is no mechanism to expose any folder data without authentication.

## What Changes

- Add a new `folder_shares` table linking a folder to an opaque, randomly-generated share token (one active share per folder).
- Add owner-facing endpoints (authenticated) to manage sharing for a folder:
  - `POST /folders/:id/share` — create (or return existing) share token for the folder
  - `GET /folders/:id/share` — return the folder's share token if shared, 404 if not
  - `DELETE /folders/:id/share` — revoke sharing for the folder
- Add a public endpoint (no authentication required):
  - `GET /share/:token` — resolve a share token to its folder and return the folder's name, notes (description), and its images (title, thumbnail URL, full-resolution URL), ordered the same way as the folder's normal image listing. Only the folder's direct images are included; child folders are out of scope.
- Add a new `ShareHandler` / `ShareUsecase` / `FolderShareRepository`, following the existing per-domain handler convention (mirrors `trash`).

## Capabilities

### New Capabilities
- `folder-sharing`: Creating, retrieving, and revoking a folder share link (owner-facing), and the public read-only endpoint that serves a shared folder's name, notes, and images via its share token.

### Modified Capabilities
(none — no existing capability's requirements change)

## Impact

- New migration `000015_create_folder_shares` (up/down) adding the `folder_shares` table.
- New `domain.FolderShare` type (`backend/internal/domain/folder_share.go` or added to `folder.go`).
- New `backend/internal/usecase/share_usecase.go` and `share_repository.go` (repository interface), `backend/internal/repository/folder_share_repository.go` (implementation).
- New `backend/internal/handler/share.go` (`ShareHandler`).
- New routes registered in `backend/cmd/server/main.go`:
  - `protected.POST/GET/DELETE /folders/:id/share`
  - `e.GET /share/:token` (public, outside the `protected` group)
- New Bruno requests for all four endpoints.
- `ShareUsecase` defines its own narrow repository interfaces (per the "interfaces defined by consumer" convention) — e.g. a folder lookup and an ordered image-listing interface — satisfied implicitly by the existing `folder_repository.go` / image-folder repository implementations. No changes to the existing `FolderRepository` / `FolderImageRepository` interfaces defined by `folder_usecase.go` / `image_usecase.go`.
