## Context

`AppLayout.tsx` renders a `flex h-screen` shell: `FolderSidebar` is a `fixed inset-y-0 left-0 w-[240px]` aside, and `<main>` carries an unconditional `ml-[240px]` (or `ml-0` only in focus mode). `RightPanel` is a `w-80` aside rendered as a sibling of `<main>` when an image or folder is selected. None of this branches on viewport size today — on a phone-width screen, the fixed sidebar margin alone consumes most of the available width.

The codebase already has one working responsive pattern: `SharePage`/`SharedFolderPanel` (share-viewer feature) uses plain Tailwind `sm:` breakpoint classes to reflow from a stacked mobile layout to a side-by-side desktop layout, with no JS viewport-detection hook. This design reuses that same pattern rather than introducing a new one.

`MasonryLayout` already computes its column count from `containerWidth` (via `Math.floor(containerWidth / 220)`), so it requires no changes — it will naturally yield fewer columns once the sidebar stops permanently consuming width on mobile.

## Goals / Non-Goals

**Goals:**
- Make the gallery viewable and navigable on small viewports using the existing Tailwind-breakpoint convention.
- Keep desktop behavior (≥ breakpoint) pixel-identical to today.
- Touch only the files listed in the proposal's Impact section; no new dependencies, no new architectural pattern.

**Non-Goals:**
- Matching the mobile design mockup pixel-for-pixel (full-screen image viewer, bottom sheet, 2-column forced grid, dark full-screen view).
- Image/folder detail viewing on mobile (right panel is inaccessible this pass).
- Touch-correct drag-and-drop (manual reorder, drag-to-folder). `dnd-kit` sensors remain mouse-oriented; touch drag behavior is unaddressed and out of scope.
- A responsive `SettingsModal`. It stays as-is; whether its entry point is reachable from the mobile drawer is an open question below.

## Decisions

**Breakpoint: Tailwind's `sm` (640px), not a JS `useIsMobile` hook.**
The design mockup used a JS resize-listener hook (`useIsMobile(680)`) because it swapped entire component subtrees (drawer vs. fixed sidebar, bottom sheet vs. aside). Since this pass cuts the bottom-sheet/full-screen-viewer scope entirely, the remaining changes are all "hide/show/transform via CSS" — exactly what `SharePage` already does with `sm:` classes. Reusing that convention avoids introducing a new pattern and avoids a resize-listener re-render on every breakpoint crossing. Alternative considered: a `useIsMobile` hook matching the mockup — rejected as unnecessary architecture for what's now a CSS-only problem, and inconsistent with the one responsive precedent already in the codebase.

**Sidebar: off-canvas transform, state lifted to `AppLayout`.**
`FolderSidebar`'s root stays `fixed inset-y-0 left-0 w-[240px]`; it gains two new optional props, `mobileOpen` and `onMobileClose`, and its className becomes transform-driven: translated fully off-screen by default, translated in when `mobileOpen` is true, and always `translate-x-0` at `sm:` and up (so desktop is unaffected regardless of the boolean's value). The boolean itself is `useState` in `AppLayout`, mirroring the existing `folderPanelOpen` pattern already there — no new state-management approach. A backdrop (`fixed inset-0 bg-black/35`, `sm:hidden`, only rendered while `mobileOpen`) closes the drawer on tap. Alternative considered: a CSS-only checkbox/`:checked` sibling-selector toggle (no JS state at all) — rejected because `AppLayout` already manages comparable boolean UI state imperatively (e.g. `folderPanelOpen`, `uploadOpen`), and a `useState` boolean is more consistent with that and easier to wire to the new top bar's hamburger button.

**New mobile top bar, rendered only below `sm`.**
A small new component (hamburger button + centered "Bookleaf" wordmark, fixed top, `sm:hidden`) gives the only way to reach the drawer once it's off-canvas by default. Built as plain Tailwind markup consistent with the existing `SharePage` header (`h-12`/`h-11` flex row, border-b), not ported from the mockup's inline-style React component.

**Upload entry point: FAB, not a second mobile toolbar.**
Already decided in conversation: a small additive `FloatingUploadButton` (`sm:hidden`, fixed bottom-right) wired to the same `setUploadOpen(true)` already in `AppLayout`, rather than forking the search/toolbar into a parallel mobile-only component. This avoids duplicating the search input's markup/state and keeps `GalleryToolbar` as the single source of truth for toolbar behavior — it only gains two `hidden sm:flex` wrappers (around the existing sort dropdown and the existing `uploadActions` div).

**Right panel: hidden via className, not via branching the selection handlers.**
`RightPanel`'s root `<aside>` gains `hidden sm:flex` (replacing `flex`). `AppLayout`'s existing `onImageSelect`/`onFolderSelect` handlers are left untouched — they still set `selectedImage`/`folderPanelOpen` state on tap, it just renders nothing visible below `sm`. Alternative considered: branch the handlers themselves to no-op on mobile — rejected as unnecessary complexity; the state being set with no visible effect is harmless, and keeping the handlers unbranched means zero risk of accidentally changing desktop behavior.

**Focus-mode toggle: hidden via className, logic untouched.**
The toggle's `Toggle` element in `GalleryToolbar`'s `focusToggle` slot gets wrapped in `hidden sm:flex` at the call site in `AppLayout`. `focusMode` state and its effect on `<main>`'s margin are unchanged; the toggle is just not reachable below `sm` (consistent with the sidebar already being off-canvas by default there).

**z-index scale.**
Reusing the same layering the mockup already settled on, since it's a reasonable scale and avoids picking new numbers: top bar `z-20`, backdrop `z-25`, drawer `z-30`, existing drag-and-drop overlay stays `z-50`, `SettingsModal` stays `z-100`. No existing z-index in the codebase needs to change.

## Risks / Trade-offs

- [Risk] Tapping an image card on mobile sets `selectedImage` but renders no visible panel — a tap with no feedback. → Mitigation: explicitly accepted per proposal scope (right panel is out of scope this pass); a follow-up change can add a lightweight mobile affordance later.
- [Risk] `dnd-kit`'s pointer sensors are mouse-oriented; manual image reordering and drag-to-folder may behave poorly or not at all on touch. → Mitigation: out of scope, called out explicitly rather than silently broken; no attempt to fix sensor config this pass.
- [Risk] `FolderSidebar`'s internal profile menu (`position: absolute`, opens upward from the footer) is unaffected by the parent's `translate-x` transform in theory, but should be visually verified once the drawer is built, since CSS transforms create a new containing block for descendant `fixed`-position elements (not `absolute`-position ones, which this is — low risk, but worth a manual check). → Mitigation: visual check during implementation; no `fixed`-positioned descendants exist inside `FolderSidebar` today.
- [Risk] 640px (`sm`) is narrower than the mockup's 680px cutoff. → Mitigation: accepted; matching the existing codebase's only responsive precedent (`SharePage`) takes priority over matching the mockup's exact number, and the difference is cosmetic at the boundary, not a functional risk.

**Resolved:** below `sm`, the `ProfileMenu` dropdown renders only "Sign out" — the "Settings" item is omitted entirely, consistent with `SettingsModal` being out of scope this pass (same reasoning as the right panel being inaccessible). Tapping an image card on mobile is an accepted silent no-op — no acknowledgment UI is added.

## Migration Plan

Pure frontend, additive/CSS-driven change — no data migration, no backend or extension changes, no feature flag. Ships as a normal frontend deploy. Rollback is a plain revert (no state to unwind), since nothing persists beyond in-memory `useState`.

## Open Questions

None outstanding.
