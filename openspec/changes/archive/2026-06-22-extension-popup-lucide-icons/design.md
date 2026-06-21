## Context

This is a known-library swap, not a new architectural pattern — `lucide-react` is already used throughout `frontend/`. The only thing worth pinning down is visual parity with the current hand-drawn icons.

## Goals / Non-Goals

**Goals:**
- Replace `MoonIcon`/`SunIcon`/`GearIcon` with Lucide's `Moon`/`Sun`/`Settings` at the same visual size/weight as today.

**Non-Goals:**
- No change to the Chrome/Firefox brand icons in `frontend`.
- No change to popup layout, theming logic, or settings behavior.

## Decisions

- **Sizing**: current icons are 14x14 inline SVG. Lucide icons default to 24x24 — pass `size={14}` explicitly to each Lucide icon to match.
- **Color**: current icons render via inherited `currentColor`/stroke from parent CSS (dark/light theme). Lucide icons use `stroke="currentColor"` by default, so no extra color prop needed — verify visually in both themes after swap.
- **Dependency placement**: add `lucide-react` to `extensions/package.json` `dependencies` (not `devDependencies`), matching how `frontend/package.json` declares it.

## Risks / Trade-offs

- [Bundle size] → Lucide icons are imported individually and tree-shaken; adding 3 icons to an extension bundle is negligible (~1-2KB), no mitigation needed beyond confirming via build output.
- [Visual mismatch after swap] → Manually verify icon appearance in both dark and light popup themes before merging.
