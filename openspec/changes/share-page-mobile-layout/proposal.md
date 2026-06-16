## Why

`SharePage` is the only public-facing page in Bookleaf, and share links are commonly opened on mobile devices. `SharedFolderPanel`'s fixed 280px width squeezes the masonry gallery to ~95px on a 375px phone, making the page effectively unusable at phone resolutions.

## What Changes

- `SharePage` outer layout changes from a fixed horizontal row to a mobile-first responsive layout: stacked column by default, side-by-side at `sm:` (640px) and above.
- `SharedFolderPanel` gets a second layout mode: full-width with inline content flow on mobile (auto height, `border-b` divider, no sticky footer), restoring to its current fixed-width side-panel form at `sm:`.
- The panel visually stacks above the gallery on mobile via CSS `order-first sm:order-last` — DOM order is unchanged (gallery first in source).
- No new primitives, components, or dependencies are introduced. Pure responsive class changes.

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `fe-share-viewer`: Layout requirements extended to cover mobile viewport behavior. `SharedFolderPanel` now has a defined stacked layout at narrow widths; `SharePage` scroll container adapts per breakpoint.

## Impact

- `frontend/src/features/share-viewer/components/SharePage.tsx`
- `frontend/src/features/share-viewer/components/SharedFolderPanel.tsx`
- `openspec/specs/fe-share-viewer/spec.md` — delta spec to add mobile layout requirements

No backend changes. No API changes. No other components touched.
