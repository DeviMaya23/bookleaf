## Context

`PUT /folders/:id` currently does a blind full-replace: `folder_repository.go` fetches the existing row, overwrites `name`/`parent_id`/`description` with whatever's in the request struct, and runs `Select(...).Updates(existing)`. Every real caller (`renameFolder`, `moveFolder`, `updateFolderDetails` in `lib/folders.ts`) sends a partial body, so any field it omits gets written as `NULL` — silently un-parenting a folder on rename, or clearing its description on move.

`PATCH /images/:id` already solves this exact problem and is the only other "partial update" contract in the codebase:

```
updateImageRequest{ Title *string; FolderIDs json.RawMessage; SourceURL json.RawMessage; ... }
        ↓ handler decides: absent / "null" / value
UpdateImageParams{ Title *string; SourceURL **string; FolderIDs *[]uuid.UUID; ... }
        ↓ usecase builds selective column map
fields := map[string]any{}; only populated for provided keys
        ↓
imageRepo.Update(ctx, id, userID, fields)  -- Model(&domain.Image{}).Where(...).Updates(fields)
```

The folder endpoint is the one resource that doesn't follow this house pattern. The original `folder-endpoints` spec already mixed contracts (`name` "required"/PUT-like, `parent_id` "optional"/PATCH-like) under a `PUT` verb — the implementation resolved that tension with a naive overwrite. The FE's three-way wrapper split (`renameFolder`/`moveFolder`/`updateFolderDetails`) exists only because each call site had to hand-curate which fields it dared send to avoid wiping the others under the full-replace contract.

The folders FE is the only consumer of this endpoint — the browser extension never calls `/folders/*`.

## Goals / Non-Goals

**Goals:**
- Replace `PUT /folders/:id` full-replace with `PATCH /folders/:id` partial-merge semantics, structurally mirroring the `PATCH /images/:id` contract at every layer (DTO, usecase params, repository update)
- Eliminate the data-integrity bug where omitted fields are nulled
- Consolidate the three FE wrappers into one `updateFolder` mirroring `updateImage`/`UpdateImageParams`, and update all call sites to send minimal partial bodies

**Non-Goals:**
- No changes to `POST /folders`, `GET /folders`, `GET /folders/:id`, or `DELETE /folders/:id`
- No changes to the `folders` table schema or `domain.Folder`
- No new validation beyond what merge-correctness requires (e.g., this does not add parent-cycle detection — that's pre-existing absent behavior, unchanged by this proposal)
- No changes to `PATCH /images/:id` itself — it's the reference, not a target
- No retroactive repair of folders whose `parent_id`/`description` may have already been nulled by the existing bug (see Open Questions)

## Decisions

### 1. Change the verb: `PUT` → `PATCH`
Partial-merge semantics under `PUT` would be technically possible but dishonest — `PUT` implies full-resource replacement by convention, and leaving it as `PUT` while behaving like `PATCH` perpetuates the exact contract ambiguity that caused this bug (the original spec's "required name / optional parent_id under PUT" contradiction). `PATCH` is both the correct REST verb for "send only what changes" and matches the sibling `PATCH /images/:id`, so callers and reviewers already have the right mental model.

**Alternative considered**: keep `PUT`, just fix the merge logic underneath. Rejected — it would leave folders as the only `PUT`-with-merge-semantics endpoint in the API, a worse inconsistency than the one being fixed.

### 2. Mirror the images contract structurally, not just behaviorally
Three layers, copied directly from the proven `image` implementation:
- **Handler DTO**: `json.RawMessage` per optional field (`Name`, `Description`, `ParentID`) so the handler can distinguish "key absent" / `"null"` / a real value — same technique as `updateImageRequest.SourceURL`/`FolderIDs`
- **Usecase params**: a dedicated `UpdateFolderParams` struct using `*string`/`**string`/`*uuid.UUID`/`**uuid.UUID` (mirroring `UpdateImageParams.Title`/`SourceURL`), so "not provided" (`nil` outer pointer) is distinguishable from "explicitly cleared" (`non-nil` outer pointing to `nil` inner)
- **Repository**: replace `Update(ctx, folder *domain.Folder)` with `Update(ctx, id uuid.UUID, userID string, fields map[string]any) (*domain.Folder, error)`, doing `Model(&domain.Folder{}).Where("id = ? AND user_id = ?", ...).Updates(fields)` and a `RowsAffected == 0 → ErrRecordNotFound` check — mirroring `imageRepository.Update` exactly, including dropping the pre-fetch (`GetByID`) in favor of the affected-rows check

**Alternative considered**: invent a folder-specific optional-field convention (e.g., a generic `Option[T]` wrapper type). Rejected per CLAUDE.md's decision-boundary rule — that would be a *new* pattern requiring sign-off, whereas copying the image convention introduces nothing new to the codebase and keeps the two "update with partial body" endpoints consistent for future readers.

### 3. Validation: `name` optional, but non-blank if provided
If `name` is present in the request and is blank/whitespace-only, reject with the existing `ErrInvalidFolderName` (400) — exactly mirroring `title`'s validation in `UpdateImage` (`req.Title != nil && TrimSpace(*req.Title) == "" → 400`). This was an explicit decision to favor consistency with images over a "silently ignore blank values" alternative, since silent no-ops are harder to debug than a clear rejection.

### 4. Repository interface signature change
`FolderRepository.Update` changes from `Update(ctx, folder *domain.Folder) (*domain.Folder, error)` to `Update(ctx, id uuid.UUID, userID string, fields map[string]any) (*domain.Folder, error)`. This is a breaking change to an internal interface (one implementation, no external consumers), justified by making the selective-update intent explicit at the type level — a `*domain.Folder` parameter invites exactly the "just overwrite everything" mistake this proposal is fixing.

## Risks / Trade-offs

- **[Risk] Breaking change to a shipped, tested endpoint** → Mitigated: the FE is the sole consumer (confirmed — the browser extension never calls `/folders/*`), and this proposal updates BE and FE together in lockstep, so there's no window where a stale client hits the new contract or vice versa.
- **[Risk] Repository interface signature change ripples into existing tests** → `TestFolderRepository_Update_PersistsFields` currently asserts full-replace behavior and must be rewritten to assert partial-merge (omitted-field-preserved vs. explicitly-nulled cases); handler/usecase tests for `UpdateFolder` need equivalent updates. This is expected churn, not a hidden cost — called out explicitly in tasks.
- **[Trade-off] Dropping the pre-update `GetByID` fetch** → The current repo fetches-then-updates; mirroring images means relying on `RowsAffected == 0` for the not-found/not-owned check instead. This is strictly an alignment with the proven image pattern, not a novel risk — `imageRepository.Update` has run this way in production already.

## Migration Plan

No database migration — this is a contract and code-shape change only. Deploy is a single coordinated BE+FE release (same as any other shared-contract change in this codebase): merge the PR that lands both sides together. Rollback is a straight revert of that merge; no data backfill is required to roll back since no schema or stored data changes.

## Open Questions

- **Should already-corrupted folders be repaired?** If the live bug has already nulled `parent_id` on any production folder via a rename, fixing the endpoint going forward won't restore that lost relationship. Worth a quick production data check (`SELECT * FROM folders WHERE parent_id IS NULL` cross-referenced with whatever audit trail exists) to decide whether a one-off data-repair script is warranted — proposed as a follow-up decision once this lands, not a blocker for the contract fix itself.
