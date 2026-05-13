## Why

The frontend needs a foundational application shell before any feature UI can be built. Establishing the two-panel layout now (sidebar + main content area) gives a consistent visual frame modelled after tools like Raindrop.io and Eagle that users already recognise for image/collection management.

## What Changes

- Add a fixed 240 px left sidebar that renders a placeholder folder list (Unfiled, Nature, Travel) with a "+ New folder" affordance
- Add a scrollable right content area that renders an empty image grid shell
- Wire the two panels together into a single root layout component
- No real data or API calls — placeholder content only

## Capabilities

### New Capabilities

- `app-shell`: Root two-panel layout (sidebar + main content area) using Tailwind and shadcn/ui components

### Modified Capabilities

<!-- none -->

## Impact

- New React component files under the frontend source tree
- Tailwind CSS and shadcn/ui must already be configured (see `fe-project-setup` spec)
- No backend or API changes
