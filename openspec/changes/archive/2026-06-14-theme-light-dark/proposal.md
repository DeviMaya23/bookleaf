## Why

The `fe-theme` capability was deliberately built to scale to multiple named themes (a `Theme` union + a `[data-theme="..."]` CSS block per theme), but only "warm" has ever existed — the `Theme` type is a single-member union, `@custom-variant dark` points at an empty/unused `[data-theme="dark"]` block, and the Settings modal's theme row (`AppSection.tsx`) renders a single non-interactive entry with no `setTheme` call. The original warm-theme-refresh proposal explicitly deferred a real picker and FOUC handling until a second theme existed.

This change adds two new themes — "lumen" (a brighter, more neutral light theme) and "sunless" (a near-black dark theme) — turns the Settings theme row into a real 3-way picker, and closes two gaps that only matter once multiple themes exist: theme preference leaking onto public/marketing pages, and a flash of the wrong theme on load.

## What Changes

- Add two new named themes to `fe-theme`: `lumen` and `sunless`, each with a full `[data-theme="..."]` CSS variable block (`background`, `foreground`, `card`, `popover`, `muted`, `accent`, `secondary`, `primary`, `border`, `input`, `ring`, `sidebar`, `+ -foreground` counterparts, and `destructive` for `sunless`).
- Extend `Theme` from `'warm'` to `'warm' | 'lumen' | 'sunless'`; `readStoredTheme()` validates against all three and continues to default to `'warm'`.
- Re-point `@custom-variant dark` from the unused `[data-theme="dark"]` to `[data-theme="sunless"]`, so the existing `dark:` utilities (`avatar.tsx`, `dropdown-menu.tsx`, `context-menu.tsx`) activate under the `sunless` theme.
- Turn `AppSection`'s theme row into a real 3-option picker: render `warm`, `lumen`, and `sunless` as selectable radio cards with swatches, wired to `setTheme`.
- Lock all public/marketing routes (`/`, `/about`, `/privacy`, `/ai-notes`) to `data-theme="warm"`, independent of the user's stored preference.
- Replace the hardcoded `style={{ backgroundColor: '#F0EBE3' }}` in `LandingPage.tsx` with a theme-token-based class.
- Add a FOUC-prevention inline script to `index.html` that applies the stored `data-theme` (or `'warm'` default) before first paint.

## Capabilities

### Modified Capabilities
- `fe-theme`: adds `lumen` and `sunless` theme palettes, re-scopes the `dark:` variant to `sunless`, adds a public-route theme lock, and adds initial-load FOUC prevention.

### New Capabilities
(none)

## Impact

- `frontend/src/hooks/useTheme.tsx`: `Theme` union gains `lumen` | `sunless`; `readStoredTheme` validation updated.
- `frontend/src/index.css`: new `[data-theme="lumen"]` and `[data-theme="sunless"]` blocks; `@custom-variant dark` re-pointed to `sunless`.
- `frontend/src/features/settings/components/AppSection.tsx`: theme row becomes a 3-option picker (new swatches/labels for `lumen`/`sunless`, `onChange` wired to `setTheme`).
- `frontend/src/pages/LandingPage.tsx`: hardcoded `#F0EBE3` background replaced with a token.
- `frontend/index.html`: inline FOUC-prevention script.
- `frontend/src/App.tsx` (or a new wrapper component): public routes (`/`, `/about`, `/privacy`, `/ai-notes`) locked to `data-theme="warm"`.
- `frontend/src/hooks/useTheme.test.tsx` and `frontend/src/features/settings/components/AppSection.test.tsx`: updated for the 3-theme set and picker behavior.
- Visual-only change for the authenticated app (`/app/*`); public pages render identically to today regardless of the user's theme preference. No backend or extension changes.
