## Why

The extension popup's Moon/Sun/Settings (gear) icons are hand-rolled inline SVGs, while the rest of the codebase (frontend) has already standardized on `lucide-react` for generic UI icons. Swapping to Lucide removes one-off SVG maintenance and keeps icon usage consistent across the project.

## What Changes

- Add `lucide-react` as a dependency of the `extensions` package.
- Replace the inline `MoonIcon`, `SunIcon`, and `GearIcon` components in `extensions/src/popup/App.tsx` with Lucide's `Moon`, `Sun`, and `Settings` icons, sized/styled to match current visual appearance (14x14, theme-aware stroke color).
- Update `extensions/src/popup/Settings.tsx` to import the Lucide icons in place of the removed local ones.
- Out of scope: the Chrome/Firefox brand icons in `frontend/src/features/settings/components/ExtensionsSection.tsx` stay as inline SVG — they are brand marks with no Lucide equivalent.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `extension-popup-settings`: the "Settings entry point" requirement currently names the entry control as a "gear icon" — wording updates to "Settings icon" to reflect the Lucide `Settings` icon. No behavior change.

## Impact

- `extensions/package.json` — new dependency (`lucide-react`).
- `extensions/src/popup/App.tsx` — removes `MoonIcon`, `SunIcon`, `GearIcon`; updates usages.
- `extensions/src/popup/Settings.tsx` — updates icon import source.
- No backend, API, or cross-layer contract changes.
