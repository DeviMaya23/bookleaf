## Context

Folders currently have no visual differentiator beyond name and nesting depth. The sidebar (`FolderItem.tsx`, `FolderSidebar.tsx`, `UnsortedEntry.tsx`, `TrashEntry.tsx`) renders plain text rows in a `flex items-center gap-1` layout with no fixed-width slots, so adding an icon span is layout-neutral.

The backend already has an established pattern for adding an optional folder field end-to-end (the `description` field, migration `000007`): a migration, a domain struct field, `json.RawMessage`-based partial-update parsing in the handler, and a `map[string]any` dynamic update in the repository. The frontend already has an established pattern for a per-user boolean preference (`vision_enabled`): a column on `users`, exposed via `GET /me` / `PATCH /me`, and a `Switch` wired with `useMutation` in a settings section.

This design reuses both patterns rather than introducing new ones.

## Goals / Non-Goals

**Goals:**
- Let users assign one icon (from a fixed 55-key allowlist) to each of their own folders.
- Validate the icon key server-side on write, so the DB never holds an unrecognized value.
- Show icons next to folder names in the sidebar, including fixed (non-editable) icons for All, Unsorted, and Trash.
- Let users globally hide all folder icons via a settings toggle, without altering stored data.

**Non-Goals:**
- No icon upload/custom image support — lucide allowlist only.
- No per-folder color customization.
- No "More icons..." expansion UI (explicitly deferred; the submenu is fixed to the 55-key list for now).
- No shared BE/FE manifest or codegen for the allowlist — each side maintains its own list manually (see Decisions).

## Decisions

**1. Icon stored as nullable `TEXT` column on `folders`, not a separate table.**
Mirrors the `description` field exactly. One folder has at most one icon; no need for a join table. `NULL` means "use default icon" (`folder`), keeping existing folders valid with no backfill required.

**2. Allowlist enforced server-side, defined as a Go `map[string]struct{}` (or slice) constant in the folder package.**
Without server-side validation, any client could write an arbitrary string into `icon`, forcing the frontend to defensively handle invalid keys on every render. Validating once at the write boundary (consistent with existing conventions) means the frontend can trust `icon` is always either `null` or a known key.
*Alternative considered*: DB-level `CHECK` constraint enum. Rejected — Go-level validation is simpler to extend and matches how other request validation already works in this handler layer (e.g. non-empty `name`).

**3. Allowlist source of truth lives in the backend; frontend maintains its own hand-written mirror (icon key → lucide component map), not a shared/generated manifest.**
A single shared manifest (e.g. JSON file read by both layers, or a codegen step) would be a new cross-layer tooling pattern. Per the project's decision-boundary rule, introducing such a pattern requires separate confirmation — not something to fold into this change. With 55 static, rarely-changing entries, two hand-maintained lists are an acceptable one-time-sync cost.

**4. Update flow reuses the existing `json.RawMessage` partial-update pattern in `PATCH /folders/:id`.**
`icon` becomes a third optional field alongside `name` and `description` in `updateFolderRequest`, parsed the same way (presence in the JSON body vs. omitted vs. explicit `null`), and validated against the allowlist before being added to the `fields map[string]any` passed to the repository's `Update`.

**5. Icon visibility toggle (`folder_icons_enabled`) stored as a boolean column on `users`, exposed via `GET /me` / `PATCH /me`, mirroring `vision_enabled` exactly.**
*Alternative considered*: a generic `preferences` JSON blob column on `users` to hold multiple future toggles. Rejected for now — there is no existing preferences-blob pattern in the codebase, and introducing one is a new abstraction outside this change's scope. A second flat boolean column is consistent with how `vision_enabled` was added and avoids a premature generalization for a single flag.

**6. Toggle-off behavior: the icon `<span>` is conditionally not rendered at all (not rendered as blank/hidden).**
`FolderItem.tsx`, `UnsortedEntry.tsx`, `TrashEntry.tsx`, and the "All" entry in `FolderSidebar.tsx` all use plain flex rows with `gap-*` and no fixed-width slots, so omitting the icon span lets the name shift left naturally with no layout compensation needed.

**7. Icon picker UI: a `ContextMenuSub` flyout (submenu) listing all 55 icons, opened from "Change icon" in the existing folder context menu.**
Matches the existing `ContextMenuSub`/`ContextMenuSubTrigger`/`ContextMenuSubContent` pattern already used elsewhere in `context-menu.tsx`. A searchable dialog was considered but is unnecessary at this list size and is the natural place to add a "More..." escape hatch later if the allowlist grows.

## Risks / Trade-offs

- **[Allowlist drift]** Backend and frontend allowlists could fall out of sync (e.g. a key removed on one side but not the other) → Mitigation: keep the list small and treat it as rarely-changing; if an unknown key is ever encountered on the frontend (e.g. legacy data), fall back to the default `folder` icon rather than erroring.
- **[Submenu length]** A flat 55-item submenu may be unwieldy to scan → Mitigation: explicitly accepted as a v1 trade-off per requirements; "More..." grouping/search is a deferred future option, not in scope.
- **[Default icon ambiguity]** Existing folders will have `icon = NULL` after migration → Mitigation: NULL is treated as "use default (`folder`)" everywhere it's read, so no backfill migration step is needed.

## Migration Plan

1. Add migration `000017_add_folder_icon` (`folders.icon TEXT`, nullable) and `000018_add_user_folder_icons_enabled` (`users.folder_icons_enabled BOOLEAN NOT NULL DEFAULT true`).
2. Ship backend changes (domain fields, handler/usecase validation and wiring) — backward compatible, no frontend dependency.
3. Ship frontend changes (icon map, sidebar rendering, context submenu, settings toggle) once backend is deployed.
4. Rollback: drop both columns via the corresponding `.down.sql` migrations; no data loss beyond the icon assignments themselves.

## Open Questions

None outstanding — allowlist, defaults, and fixed system icons were finalized in proposal.md based on `icons.txt`.
