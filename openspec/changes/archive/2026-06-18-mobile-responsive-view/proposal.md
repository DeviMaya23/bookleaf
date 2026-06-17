## Why

The app is currently unusable on small viewports: the left sidebar and main content area use fixed pixel widths (240px sidebar + unconditional `ml-[240px]` offset on `<main>`), which leaves almost no usable width for the gallery on a phone-sized screen. This is a bare-minimum pass to make the gallery viewable and navigable on mobile — not a full design-parity port of the mobile mockup.

## What Changes

- The left sidebar becomes an off-canvas drawer below a small-viewport breakpoint: hidden by default, opened via a hamburger button in a new mobile-only top bar, with a tap-to-dismiss backdrop. Above the breakpoint, the sidebar's existing fixed/always-visible behavior is unchanged.
- A new mobile-only top bar (hamburger + centered "Bookleaf" wordmark) renders below the breakpoint; nothing changes above it.
- The main content area fills the full viewport width below the breakpoint instead of always reserving 240px for the sidebar.
- The right panel (image/folder detail panel) does not open below the breakpoint — **BREAKING** (mobile-only): selecting an image or folder no longer opens any detail panel on small viewports. This is an explicit, temporary scope cut for this pass.
- The focus-mode toggle is hidden below the breakpoint (the sidebar is already collapsed by default there, so the toggle is redundant).
- The sidebar's profile menu dropdown shows only "Sign out" below the breakpoint (the "Settings" item is omitted), since the settings modal itself is not being made responsive in this pass.
- The gallery toolbar's sort control and the existing "+Image" split button are hidden below the breakpoint.
- A new floating action button (FAB), bottom-right, becomes the mobile upload entry point below the breakpoint, opening the same upload modal as the existing "+Image" button.
- The masonry gallery grid, image cards, and rich pan/zoom image viewer are unchanged — masonry is already container-width-driven and will naturally reflow.

## Capabilities

### New Capabilities
- `fe-mobile-shell`: the mobile-only top bar (hamburger + wordmark), the sidebar drawer open/close/backdrop interaction, and the floating upload button — net-new UI not covered by any existing capability.

### Modified Capabilities
- `app-shell`: the "Two-panel application shell" requirement changes from "sidebar is always fixed and visible, main is always offset" to "sidebar is off-canvas by default and main is full-width below the breakpoint; unchanged above it."
- `focus-mode`: the focus toggle's visibility requirement gains an exception — not rendered below the breakpoint.
- `fe-gallery-sort`: the "Sort control in gallery toolbar" requirement gains an exception — not rendered below the breakpoint.
- `fe-right-panel`: the "Right panel opens when an image card is clicked" and "Right panel opens or updates when a folder is selected" requirements gain an exception — selection does not open the panel below the breakpoint.
- `user-profile-menu`: the "ProfileMenu component" requirement gains an exception — the dropdown shows only Sign out (no Settings item) below the breakpoint.

## Impact

- `frontend/src/app-shell/AppLayout.tsx`: responsive margin/offset, drawer-open state, renders the new mobile top bar and FAB, gates right-panel rendering by breakpoint.
- `frontend/src/features/folder-sidebar/components/FolderSidebar.tsx`: off-canvas transform + open/close props, driven by state lifted to `AppLayout`.
- `frontend/src/features/gallery/components/GalleryToolbar.tsx`: hide sort control and upload split button below the breakpoint.
- `frontend/src/features/right-panel/components/RightPanel.tsx`: not rendered below the breakpoint.
- `frontend/src/features/auth/components/ProfileMenu.tsx`: dropdown shows only "Sign out" below the breakpoint.
- New components: mobile top bar, floating upload button.
- No backend, extension, database, or dependency changes. No changes to `MasonryLayout`, `ImageGrid`, or the rich `ImageViewer`.
