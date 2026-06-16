## Context

`SharePage` uses a fixed horizontal flex layout: a `flex-1 overflow-y-auto` gallery container on the left and `SharedFolderPanel` (fixed `w-[280px]`, `h-full`, `border-l`, sticky-footer export button) on the right. This layout assumes the viewport is wide enough to accommodate both columns. At 375px, the panel consumes 280px and leaves ~95px for the masonry gallery — one very narrow column of 95px-wide thumbnails.

The masonry layout itself (`computeMasonryLayout`) already degrades to a single column via `Math.max(1, Math.floor(containerWidth / 220))`. The gallery needs no changes — only the surrounding layout and the panel component need to become responsive.

This is the first use of Tailwind responsive prefixes in the codebase. The convention established here (mobile-first base styles, `sm:` overrides for desktop) will serve as precedent if `RightPanel` in `AppLayout` eventually needs the same treatment.

## Goals / Non-Goals

**Goals:**
- Make `SharePage` usable on phone-width viewports (≥320px)
- Panel stacks above gallery on mobile; side-by-side layout restored at `sm:` (640px)
- Establish mobile-first Tailwind convention for this codebase

**Non-Goals:**
- Tablet/mid-range breakpoints (`md:`, `lg:`) — out of scope for this change
- Responsive treatment for `AppLayout` or `RightPanel` — separate concern
- Any touch-specific interactions (swipe-to-dismiss, etc.)

## Decisions

### D1: Stacked column layout (Option C) over a toggle/drawer

The panel content is lightweight (folder name, image count, notes, export button). A drawer would require a new primitive not present in the codebase. Stacking the panel above the gallery as plain flow content achieves the same outcome without new components or interaction state. If `RightPanel` (heavier content) later needs mobile support, a drawer approach can be evaluated then.

### D2: Mobile-first Tailwind (base = mobile, `sm:` = desktop)

Tailwind's canonical idiom is mobile-first. Writing `sm:` overrides for the current desktop layout rather than adding `hidden sm:block` toggles results in cleaner, more maintainable markup. The entire codebase uses Tailwind v4 (no config file, CSS-based), so responsive prefixes work out of the box.

### D3: CSS `order` for panel placement over DOM reordering

Panel is last in the DOM (`SharePage`: gallery first, panel second). On mobile, the panel should appear above the gallery visually. Moving it first in the DOM would place it on the left at desktop (breaking the current design). Duplicating markup is worse. Using `order-first sm:order-last` on the panel achieves the desired visual order on both breakpoints without touching DOM structure.

### D4: Single scrolling column on mobile via parent `overflow-y-auto`

Desktop layout: outer div has `overflow-hidden`, each child manages its own scroll. Mobile layout: outer div becomes `overflow-y-auto flex-col` — one unified scroll surface containing both the panel and the gallery as natural-height flow content. `SharedFolderPanel`'s `h-full` and internal `flex-1 overflow-y-auto` (which assume a fixed-height column) are suppressed on mobile and restored at `sm:`.

### D5: `border-b` replaces `border-l` on the panel for mobile

On desktop the panel uses `border-l` to separate it from the gallery. When stacked, a `border-b` at the bottom of the panel is the appropriate separator. Implemented as `border-b sm:border-b-0 sm:border-l` on the panel's outer div.

## Risks / Trade-offs

- **ResizeObserver width at mobile**: The gallery container loses `flex-1` on mobile (it becomes auto-height, full-width flow content). The ResizeObserver still measures `clientWidth` correctly since the container fills its parent's width — but this should be verified during implementation.
- **Export button no longer always-visible on mobile**: The sticky footer export button becomes inline content at the top of the scroll. Users who scroll deep into the gallery must scroll back to export. Acceptable for a one-time action; can revisit if UX feedback suggests otherwise.
- **`sm:` breakpoint at 640px**: Includes small tablets in portrait. At 640px, 280px panel + 360px gallery is workable. If edge cases arise, the breakpoint can be raised to `md:` (768px) without structural change.
