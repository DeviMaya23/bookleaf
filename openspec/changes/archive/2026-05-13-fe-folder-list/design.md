## Context

The sidebar in `AppLayout` currently renders a hardcoded folder list. TanStack Query v5 is installed but not yet wired up — no `QueryClientProvider` exists. An `apiFetch` helper in `src/lib/api.ts` already handles auth token injection, so all API calls will go through it. The backend exposes `GET /folders`, `POST /folders`, `PUT /folders/:id`, and `DELETE /folders/:id` on a protected route group.

## Goals / Non-Goals

**Goals:**
- Wire `QueryClientProvider` at the app root
- Fetch and display the real folder list in the sidebar via `useQuery`
- Pin "Unsorted" permanently at the top with a visual divider below it
- Create, rename, and delete folders via `useMutation` + `invalidateQueries`
- Use shadcn `Dialog` for name input (create + rename) and delete confirmation
- Use shadcn `ContextMenu` for the right-click menu on each folder row

**Non-Goals:**
- Folder nesting / tree UI (flat list only)
- Optimistic updates (refetch-after-mutation is sufficient for now)
- Drag-and-drop reordering
- Clicking a folder to filter the image grid (wired in a future change)

## Decisions

**1. Extract sidebar into a `FolderSidebar` component**
`AppLayout` currently owns the sidebar JSX. With query logic, dialog state, and context menu state all coming in, the sidebar warrants its own component. `AppLayout` becomes a thin shell that composes `FolderSidebar` and the main content area.

**2. Single shared Dialog component for create and rename**
Both operations need the same UI: a text input and a confirm button. A single `FolderNameDialog` component accepts `initialValue` (empty for create, current name for rename) and an `onSubmit` callback. This avoids duplicating dialog markup.

**3. Delete confirmation via a separate Dialog, not ContextMenu submenu**
shadcn's ContextMenu doesn't compose well with an inline confirm step. A dedicated confirmation Dialog (triggered after selecting "Delete" from the context menu) is cleaner and matches common patterns.

**4. `invalidateQueries` after every mutation, no optimistic updates**
The folder list is small and mutations are infrequent. Refetching after each mutation is simple and always correct. Optimistic updates add complexity and edge-case handling that isn't warranted yet.

**5. Query key: `['folders']`**
A single flat key is sufficient — all mutations invalidate it. If per-folder queries are needed later, the key can be namespaced then.

**6. `apiFetch` for all API calls, Kinde `getToken` passed via hook**
The existing `apiFetch` handles auth headers. Query/mutation functions will call `useKindeAuth()` to get `getToken` and pass it through. This keeps auth concerns out of the query layer.

## Risks / Trade-offs

- [QueryClientProvider not yet present] → Must be added to `main.tsx` before any `useQuery` call works. First task in implementation.
- [Dialog state and ContextMenu state co-located in FolderSidebar] → Manageable at this scale; if it grows, extract into a custom hook later.
- [Refetch latency after mutation] → On slow connections the list may lag briefly after a create/rename/delete. Acceptable trade-off given the simplicity of the approach.
