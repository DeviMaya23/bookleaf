# Design

## Context

The existing `/internal` Echo group in `main.go` is already wired with `NewInternalSecretMiddleware` and hosts four routes served by `InternalHandler`. That handler has three injected dependencies today: `InternalShareUsecase`, `InternalAccountUsecase`, and `InternalUserResolver`.

The new route needs ownership-checked folder access and thumbnail presigning — neither of which belongs in `shareUsecase` (which is scoped to public/token-based sharing). All the required building blocks already exist: `folderRepo.GetByID` enforces ownership, `imageRepo.ListByFolder` lists images, and `store.GeneratePresignedGetURL` produces thumbnail URLs.

## Goals / Non-Goals

**Goals:**
- New route `GET /internal/users/:user_id/folders/:folder_id/images` in the existing internal group
- Ownership-checked access: folder not owned by the resolved user → 404
- Returns presigned thumbnail URLs for images that have one; silently excludes images without a thumbnail path
- Fits cleanly into the existing handler/usecase/repository layering

**Non-Goals:**
- Pagination — Booklet consumes this as a thumbnail preview; full pagination is not needed now
- Full-res or download URLs — thumbnail only
- TTL negotiation — reuses the existing `presignedGetTTL` (24 h)

## Decisions

### New focused usecase, not an addition to shareUsecase

`shareUsecase` is tightly coupled to `FolderShareRepository` and the concept of public sharing. Adding an ownership-checked, non-share route there would introduce a dependency mismatch and obscure intent. A new `folderImagesUsecase` (or equivalent) takes only what it needs: `FolderRepository`, `ImageRepository`, and `StorageService` — all interfaces already defined in the usecase package.

**Alternative considered**: reuse `shareUsecase` by adding a new method. Rejected because the new method has no conceptual relation to sharing and would require the usecase to depend on folder ownership, which is outside its current responsibility.

### Fourth dependency on InternalHandler

`InternalHandler` gains a fourth injected interface — `InternalFolderImagesUsecase`. This keeps the handler consistent with the existing pattern (all internal-facing usecase interfaces are declared in `handler/internal.go` and injected via the constructor). `NewInternalHandler` signature in `main.go` gains one more argument.

**Alternative considered**: a separate handler type. Rejected — the internal group is small and cohesive; splitting it adds indirection without benefit.

### IDP subject resolution — reuse existing UserResolver

`:user_id` is a Kinde IDP subject, consistent with all other `/internal/users/:user_id/…` routes. The already-injected `InternalUserResolver.GetByIDPSubject` resolves it to an internal UUID. On `ErrUserNotFound` the handler returns 404 (unlike `ListPublicFolders`, which returns 200 with an empty list — here an unknown user cannot own a folder, so 404 is the correct signal).

### Thumbnail-only, exclude nil paths

Images without a `ThumbnailPath` are silently excluded. Returning a null URL for them would require Booklet to handle a nullable field; excluding them keeps the response uniform and avoids serving partial data for images still being processed.

## Risks / Trade-offs

- **Large folders**: no pagination. If a folder has thousands of images, the handler presigns all of them in one request. This is acceptable for the initial use case (thumbnail preview) but may need revisiting if Booklet adopts it for large collections. → Mitigation: the endpoint is internal-only; Booklet controls call frequency.
- **Presign latency**: each image thumbnail requires one R2 presign call. The existing share endpoints have the same pattern and have not been a bottleneck. → No mitigation needed now; can be parallelised later if profiling shows a problem.
