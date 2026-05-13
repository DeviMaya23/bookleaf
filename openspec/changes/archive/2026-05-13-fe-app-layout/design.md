## Context

The frontend project (`fe-project-setup`) is already scaffolded with Tailwind CSS and shadcn/ui. We need an application shell — the persistent chrome that all pages live inside — before any feature screens can be built. The target aesthetic is Raindrop.io / Eagle: a narrow fixed sidebar for collection navigation and a wide scrollable main area for content.

## Goals / Non-Goals

**Goals:**
- Deliver a two-panel layout shell: fixed 240 px sidebar + fluid main content area
- Render a static placeholder folder list and an empty image grid
- Use only Tailwind utility classes and shadcn/ui primitives (no bespoke CSS files)
- Keep components purely presentational — no state management, no API calls

**Non-Goals:**
- Interactive folder selection or routing
- Real image data or API integration
- Responsive / mobile breakpoints (desktop-first shell only for now)
- Authentication gate around the layout

## Decisions

**1. Single `AppLayout` wrapper component**
A single `AppLayout` component wraps the two panels and is mounted at the root route. This is the simplest entry point and avoids premature route-based code splitting.
- Alternative: per-route layout via file-based routing — adds indirection with no benefit at this stage.

**2. Fixed sidebar via CSS `fixed` + `inset-y-0`**
The sidebar is `position: fixed` so it never scrolls with content. The main area gets `ml-[240px]` to compensate.
- Alternative: CSS Grid with a frozen column — equally valid but Tailwind flex/fixed is more legible for a narrow sidebar pattern.

**3. shadcn/ui `ScrollArea` for the main content**
Wrapping the image grid in shadcn's `ScrollArea` keeps overflow handling consistent with the rest of the design system.

**4. Placeholder data hardcoded in component**
Folder names are a static array literal inside the component. No context, store, or prop drilling needed until real data is wired up.

## Risks / Trade-offs

- [Hardcoded placeholder content] → Easy to remove; it's all in one place and clearly marked. No migration risk.
- [Fixed sidebar breaks at very narrow viewports] → Acceptable for now; mobile layout is an explicit non-goal.
