# Tasks

## 1. Backend — count endpoint

- [x] 1.1 Add a `TagWithCount` struct (embedding `domain.Tag` + `ImageCount int`) to `internal/domain/tag.go` and a `ListByUserIDWithCount(ctx, userID)` method to the `TagRepository` interface in `internal/usecase/tag_repository.go`; verify the interface compiles
- [x] 1.2 Implement `ListByUserIDWithCount` in `internal/repository/tag_repository.go` using a single LEFT JOIN on `image_tags` with GROUP BY; verify the method returns the correct count for a tag applied to known images by running existing repository integration tests
- [x] 1.3 Add `ListWithCount(ctx, userID)` to the `TagUsecase` interface and implement in `internal/usecase/tag_usecase.go`; verify the file compiles
- [x] 1.4 Update `TagHandler.ListTags` to read the `include_count` query param; when present call `tagUsecase.ListWithCount` and include `image_count` in each `tagResponse` (omitempty when absent); verify handler unit tests cover both the no-param path (no `image_count` field) and the `include_count=true` path (count present), plus each error mapping
- [x] 1.5 Add `bruno/tags/list-tags-with-count.bru` for `GET /tags?include_count=true`; verify the file is well-formed by checking it appears in the Bruno collection
- [x] 1.6 Run `golangci-lint ./...` from the repo root and fix any issues it reports

## 2. Frontend — lib, routing, and shell changes

- [x] 2.1 Add `| { type: 'tags' }` to `AppView` in `src/lib/view.ts`; add a `'tags'` case to `filterSectionsForViewType` (returns `[]`) and `defaultSortForViewType` (returns `{ sortBy: 'created_at', sortDir: 'desc' }`) in `useGalleryControls.ts`; verify TypeScript compiles without exhaustiveness errors
- [x] 2.2 Add `/app/tags` route to `src/App.tsx`; update `useAppView.ts` to return `{ type: 'tags' }` for that pathname; add a case to the existing `useAppView` test verifying `/app/tags` maps to `{ type: 'tags' }`; verify the test passes
- [x] 2.3 Add `updateTag(getToken, id, name)` (returns `Tag | null` — null on 409) and `deleteTag(getToken, id)` to `src/lib/tags.ts`; update `getTags` to accept an optional `includeCount?: boolean` param that appends `?include_count=true` when true; verify TypeScript compiles
- [x] 2.4 Add a Tags nav entry to `FolderSidebar.tsx` below Trash, linking to `/app/tags`, with active styling when `view.type === 'tags'`; verify it renders and navigates correctly
- [x] 2.5 Update `useGalleryControls` to accept an `initialFilterTagIds?: string[]` param; seed the `filterTagIds` initial state from it; in the `viewKey` reset effect, reset `filterTagIds` to `[]` (not to `initialFilterTagIds` — the pre-fill is a one-time mount seed, not a persistent default); verify existing filter reset behaviour is unchanged
- [x] 2.6 Update `AppLayout.tsx` to read `useLocation().state?.filterTagIds`; pass it as `initialFilterTagIds` to `useGalleryControls`; immediately clear it with `navigate(location.pathname, { replace: true, state: null })` after reading so it is not re-applied on subsequent navigations; hide `GalleryToolbar` when `view.type === 'tags'`; verify the toolbar is absent in the tags route and present in all other routes

## 3. Frontend — tags feature

- [x] 3.1 Create `src/features/tags/hooks/useTagMutations.ts` with `useRenameMutation` and `useDeleteMutation`; write unit tests covering: rename success (cache updated, row returns to view mode), rename 409 (null returned, inline error surfaced), delete success (tag removed from cache); verify tests pass
- [x] 3.2 Create `src/features/tags/components/TagRow.tsx` implementing `'view' | 'rename' | 'delete-confirm'` mode state; the rename mode renders a pre-filled input with Save/Cancel and an inline error on duplicate name; the delete-confirm mode renders "Remove tag?" with Cancel/Delete buttons; verify modes transition correctly via the action menu
- [x] 3.3 Create `src/features/tags/components/TagsView.tsx` rendering: a local search input at the top; tags fetched with `getTags(getToken, { includeCount: true })`; alphabetically grouped under letter headers showing group counts; each group rendered as `TagRow` instances; empty state when no tags exist; verify the view renders correctly in the `/app/tags` route

## 4. Final checks

- [x] 4.1 Run `npm run build` from `frontend/` and fix any type or build errors
- [x] 4.2 Run `npm run lint` from `frontend/` and fix any lint errors

## Workflow follow-up

- Archive the change once review is complete.
