## Why

The folder sidebar currently displays hardcoded placeholder data with no interactivity. Now that the backend folder endpoints are implemented, the sidebar should reflect real user data and allow users to create, rename, and delete folders without leaving the page.

## What Changes

- Replace the hardcoded folder list in the sidebar with data fetched from `GET /folders`
- Pin an "Unsorted" item permanently at the top of the list, visually separated from API folders by a divider
- Make the "+ New folder" button open a Dialog (shadcn/ui) where the user enters a folder name and confirms
- Add a right-click ContextMenu (shadcn/ui) on each folder with "Rename" and "Delete" options
- Rename reuses the same Dialog mechanism as new-folder creation, pre-populated with the current name
- Delete shows a confirmation Dialog (yes/no); folder is only deleted if the user confirms
- Invalidate the folder list query after any create, rename, or delete mutation so the list stays fresh

## Capabilities

### New Capabilities

- `folder-management`: FE interactions for creating, renaming, and deleting folders via Dialog and ContextMenu

### Modified Capabilities

- `app-shell`: Folder list requirement changes from hardcoded placeholder data to live API data with a pinned "Unsorted" entry

## Impact

- `AppLayout` / sidebar component updated or refactored into a `FolderSidebar` component
- Requires a data-fetching library (TanStack Query / React Query) wired up — check if already installed
- Calls `GET /folders`, `POST /folders`, `PUT /folders/:id`, `DELETE /folders/:id` on the backend
- New shadcn/ui components: `Dialog`, `ContextMenu` (need to be added via shadcn CLI if not present)
