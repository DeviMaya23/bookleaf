## ADDED Requirements

### Requirement: Feature code is organized under features/<feature>/
Frontend code specific to a single feature SHALL reside under `frontend/src/features/<feature>/`, with `components/`, `hooks/`, and `lib/` subdirectories as needed, rather than alongside unrelated features in top-level `components/`, `hooks/`, or `lib/`.

#### Scenario: Gallery components live under features/gallery
- **WHEN** the codebase is inspected
- **THEN** `frontend/src/features/gallery/components/ImageGrid.tsx` exists
  and `frontend/src/components/ImageGrid.tsx` does not

#### Scenario: Right panel components live under features/right-panel
- **WHEN** the codebase is inspected
- **THEN** `frontend/src/features/right-panel/components/RightPanel.tsx`
  exists and `frontend/src/components/RightPanel.tsx` does not

### Requirement: App shell and cross-cutting orchestration live under app-shell/
The top-level layout component and any logic that coordinates across multiple features (drag-and-drop orchestration, view routing) SHALL reside under `frontend/src/app-shell/`, not in top-level `components/` or `lib/`.

#### Scenario: AppLayout lives under app-shell
- **WHEN** the codebase is inspected
- **THEN** `frontend/src/app-shell/AppLayout.tsx` exists and
  `frontend/src/components/AppLayout.tsx` does not

#### Scenario: Drag-and-drop orchestration lives under app-shell
- **WHEN** the codebase is inspected
- **THEN** `frontend/src/app-shell/lib/dragHandlers.ts` exists and
  `frontend/src/lib/dragHandlers.ts` does not

### Requirement: Shared domain modules remain in lib/
Domain types and API wrappers used across multiple features (`images`, `folders`, `tags`, `thumbnail`, `view`) SHALL remain in `frontend/src/lib/` and SHALL NOT be duplicated or relocated into a single feature's directory.

#### Scenario: Domain modules remain in lib/
- **WHEN** the codebase is inspected
- **THEN** `frontend/src/lib/images.ts`, `frontend/src/lib/folders.ts`,
  `frontend/src/lib/tags.ts`, `frontend/src/lib/thumbnail.ts`, and
  `frontend/src/lib/view.ts` exist

### Requirement: Feature folder names do not collide with shared domain module names
Feature directories whose UI concerns correspond to a shared domain module SHALL use a name distinct from that module (e.g. `folder-sidebar` and `right-panel`, not `folders` or `images`), so the feature directory and the domain module remain separately greppable.

#### Scenario: Folder sidebar feature is named distinctly from the folder domain module
- **WHEN** the codebase is inspected
- **THEN** `frontend/src/features/folder-sidebar/` exists,
  `frontend/src/features/folders/` does not exist, and
  `frontend/src/lib/folders.ts` exists

### Requirement: Generic shared UI, hooks, and pages remain top-level
`frontend/src/components/ui/` (shared design-system primitives), `frontend/src/hooks/` (hooks with no feature-specific dependencies), and `frontend/src/pages/` (thin route entry points) SHALL remain at the top level of `frontend/src/`.

#### Scenario: Shared UI primitives remain in components/ui
- **WHEN** the codebase is inspected
- **THEN** `frontend/src/components/ui/button.tsx` exists
