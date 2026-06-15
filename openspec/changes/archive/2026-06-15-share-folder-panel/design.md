## Context

`FolderPanelContent` (`frontend/src/features/right-panel/components/FolderPanelContent.tsx`) already follows a query + mutation pattern via `@tanstack/react-query` (see `folderDetail` query and `saveMutation`), and the codebase has an established confirm-dialog pattern (`DeleteFolderDialog`) built on the generic `Dialog` primitives. The three backend endpoints this change wires up already exist and are unchanged.

## Goals / Non-Goals

**Goals:**
- Surface folder share state (on/off + link) in the right panel, following existing query/mutation conventions.
- Keep the UI's notion of "shared" derived directly from server state — no separate local on/off flag that could drift.

**Non-Goals:**
- Building the public `/share/:token` viewer page.
- Any change to backend behavior or contracts.

## Decisions

### Share state is a single query, `null` means "off"

`getFolderShare(getToken, folderId)` calls `GET /folders/:id/share` and returns `{ token: string } | null` — a 404 response is mapped to `null` inside the lib function, not thrown. This lets `useQuery` represent "not shared" as a normal successful result (`data === null`) rather than an error state, so the UI doesn't need to distinguish "loading vs. errored vs. not shared" — only loading vs. `data` (null | `{ token }`).

Query key: `['folder-share', folder.id]`.

### Switch reflects server state directly, both mutations invalidate the query

- `checked = !!shareData` (loading → switch disabled, not a guessed state).
- Turning the switch **on** (`shareData === null` → checked): calls `createFolderShare` (`POST /folders/:id/share`, idempotent) directly — no confirmation needed for enabling.
- Turning the switch **off** (`shareData` present → unchecked): does **not** mutate immediately. It opens a confirm dialog (mirroring `DeleteFolderDialog`'s `Dialog` + `DialogFooter` structure). Confirming calls `deleteFolderShare` (`DELETE /folders/:id/share`); cancelling leaves the switch in its current (on) state.
- Both mutations call `queryClient.invalidateQueries({ queryKey: ['folder-share', folder.id] })` on success, so the switch/link UI always reflects the latest server-confirmed state rather than an optimistic guess.

Alternative considered: optimistic toggle with rollback on error. Rejected — the extra complexity isn't justified for a low-frequency toggle, and re-fetching after mutation is cheap and simpler to reason about.

### Share link is derived, not stored separately

When `shareData` is present, the link is computed inline as `` `${window.location.origin}/share/${shareData.token}` ``. No separate state for the URL.

### Copy-to-clipboard

A small icon button calls `navigator.clipboard.writeText(url)`, with `toast.success('Link copied')` / `toast.error('Failed to copy link')` mirroring the existing toast usage in this file (`saveMutation`'s `onSuccess`/`onError`).

### Confirm dialog as a new sibling component

A new `DisableShareDialog` component (alongside `FolderPanelContent`, modeled on `DeleteFolderDialog`) takes `open`, `onCancel`, `onConfirm` props and explains that the current link will stop working and that re-enabling generates a new one.

## Risks / Trade-offs

- **Re-enabling after disabling generates a new token** (existing backend behavior — `DeleteShare` removes the row, so `CreateShare` mints a fresh token next time). → Mitigated by the confirm dialog's copy explicitly stating this.
- **Clipboard API requires a secure context** (HTTPS or localhost). The app is served over HTTPS in deployed environments and localhost in dev, so this is not expected to be an issue, but `navigator.clipboard` may be `undefined` in unusual embeddings — `writeText` call is wrapped in try/catch with the existing error-toast pattern.
- **Copied link points to a page that doesn't exist yet** (`/share/:token` viewer is out of scope). Accepted per proposal — this change only wires the authenticated control surface.
