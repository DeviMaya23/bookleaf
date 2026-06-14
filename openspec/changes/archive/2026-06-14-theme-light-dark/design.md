## Context

`fe-theme` currently defines one theme (`warm`) as a `[data-theme="warm"]` CSS variable block mirrored into `:root`, a `Theme = 'warm'` single-member union in `useTheme.tsx`, and an unused `@custom-variant dark (&:where([data-theme="dark"], ...))`. `AppSection.tsx` (Settings → App) already calls `useTheme()` and renders one non-interactive radio row — built as the seam for a future picker but not wired to `setTheme`.

Three existing components use `dark:` utilities (`avatar.tsx`, `dropdown-menu.tsx`, `context-menu.tsx`) but they're inert today since nothing sets `data-theme="dark"`.

Public/marketing routes (`/`, `/about`, `/privacy`, `/ai-notes`) and the authenticated app (`/app/*`) currently share one global `data-theme` on `<html>`, set by `ThemeProvider` from `localStorage`.

## Goals / Non-Goals

**Goals:**
- Add `lumen` (bright/neutral light) and `sunless` (near-black dark) palettes, matching the token set already defined for `warm`.
- Make the Settings theme row a real, accessible 3-way picker.
- Re-scope `dark:` utilities to the `sunless` theme.
- Keep public/marketing pages visually on `warm` regardless of the signed-in user's preference.
- Remove the last hardcoded color (`LandingPage.tsx`'s `#F0EBE3`).
- Prevent a flash of the wrong theme on initial load.

**Non-Goals:**
- A "system"/`prefers-color-scheme`-following option — three explicit named themes only.
- Per-section or per-component theme overrides beyond the public/app split.
- Changing `--radius` or typography per theme — both stay constant across `warm`/`lumen`/`sunless`.
- Server-persisted theme preference (still `localStorage` only).

## Decisions

### 1. `lumen` palette values

```css
[data-theme="lumen"] {
  --background: #FFFFFF;
  --foreground: #1A1A1A;
  --card: #FFFFFF;
  --card-foreground: #1A1A1A;
  --popover: #FFFFFF;
  --popover-foreground: #1A1A1A;
  --primary: #1A1A1A;
  --primary-foreground: #FFFFFF;
  --secondary: #F1F1F1;
  --secondary-foreground: #1A1A1A;
  --muted: #F5F5F5;
  --muted-foreground: #8A8A8A;
  --accent: rgba(0, 0, 0, 0.05);
  --accent-foreground: #1A1A1A;
  --destructive: oklch(0.577 0.245 27.325);
  --border: rgba(0, 0, 0, 0.10);
  --input: rgba(0, 0, 0, 0.07);
  --ring: #B0B0B0;
  --radius: 0.5rem;
  --sidebar: #F5F5F5;
}
```

Same structure as `[data-theme="warm"]`, with the warm/beige hues swapped for neutral gray/white/black. `--destructive` is unchanged from `warm` since `lumen` is also a light-background theme (same contrast requirements).

### 2. `sunless` palette values

```css
[data-theme="sunless"] {
  --background: #121212;
  --foreground: #EDEDED;
  --card: #1A1A1A;
  --card-foreground: #EDEDED;
  --popover: #1A1A1A;
  --popover-foreground: #EDEDED;
  --primary: #EDEDED;
  --primary-foreground: #121212;
  --secondary: #2A2A2A;
  --secondary-foreground: #EDEDED;
  --muted: #1F1F1F;
  --muted-foreground: #9A9A9A;
  --accent: rgba(255, 255, 255, 0.07);
  --accent-foreground: #EDEDED;
  --destructive: oklch(0.704 0.191 22.216);
  --border: rgba(255, 255, 255, 0.11);
  --input: rgba(255, 255, 255, 0.08);
  --ring: #5A5A5A;
  --radius: 0.5rem;
  --sidebar: #1A1A1A;
}
```

`--destructive` uses shadcn's standard dark-mode destructive token (lighter/more saturated red, needed for contrast against `#121212`) rather than the `warm`/`lumen` value.

**Note**: these are a structural starting point (mirroring `warm`'s token relationships with neutral hues). Exact values should get a quick visual pass in-browser once implemented — easy to tweak without affecting the mechanism.

### 3. `@custom-variant dark` re-pointed to `sunless`

```css
@custom-variant dark (&:where([data-theme="sunless"], [data-theme="sunless"] *));
```

One-line change from the unused `[data-theme="dark"]`. The three existing `dark:` usages (avatar border blend mode, destructive-action focus backgrounds in context/dropdown menus) become live under `sunless` — no component changes needed.

### 4. `AppSection` picker

`THEME_OPTIONS` becomes a 3-entry `Record<Theme, { label, sub, swatches }>`:

| theme id | label | sub | swatches (background / secondary / foreground) |
|---|---|---|---|
| `warm` | Default | Warm parchment | `#FAF8F4` / `#E5DED6` / `#2D2A26` *(unchanged)* |
| `lumen` | Lumen | Bright and clean | `#FFFFFF` / `#F1F1F1` / `#1A1A1A` |
| `sunless` | Sunless | Mostly black | `#121212` / `#2A2A2A` / `#EDEDED` |

The component maps over `Object.entries(THEME_OPTIONS)`, rendering each as a `<label>`-wrapped radio row (reusing the existing card markup/classes). Each radio gets `name="theme"`, `checked={theme === key}`, `onChange={() => setTheme(key as Theme)}`; `disabled`/`readOnly` are removed.

**Accessibility**: `aria-label="${option.label} theme (selected)"` (meaningful only for a single static, always-selected item) becomes `aria-label="${option.label} theme"` for all three — the radio's native `checked` state already conveys selection to assistive tech, and a real radio group shouldn't bake "(selected)" into every label.

### 5. Public-route theme lock: `PublicThemeLock` layout route

New `frontend/src/components/PublicThemeLock.tsx`, mirroring the existing `AuthGuard` layout-route pattern (`Outlet`-based):

```tsx
import { Outlet } from 'react-router-dom'

export default function PublicThemeLock() {
  return (
    <div data-theme="warm">
      <Outlet />
    </div>
  )
}
```

In `App.tsx`, `/`, `/about`, `/privacy`, and `/ai-notes` become children of a `<Route element={<PublicThemeLock />}>` wrapper. CSS custom properties cascade by DOM position: `[data-theme="warm"]` on this wrapper sets `--background`/`--foreground`/etc. for itself and all descendants, overriding whatever `<html data-theme="lumen|sunless">` set above it. `/app/*` routes are untouched and continue to read the global `data-theme`.

**Alternative considered**: setting `data-theme="warm"` individually in `LandingPage`/`SimplePageLayout`'s root `<div>`. Rejected — four call sites vs. one layout route, and `App.tsx` already uses the layout-route pattern for `AuthGuard`.

### 6. FOUC prevention: inline script in `index.html`

```html
<script>
  document.documentElement.dataset.theme = localStorage.getItem('bookleaf-theme') || 'warm'
</script>
```

Placed first in `<head>`, before any stylesheet `<link>`, so `data-theme` is set on `<html>` before the browser paints with the loaded CSS.

No validation against the known theme set: an unrecognized/stale stored value simply matches no `[data-theme="..."]` block, so `:root` (which mirrors `warm`) applies — same visual result as defaulting to `warm`, without duplicating the valid-theme list from `useTheme.tsx`. `ThemeProvider`'s mount effect (`readStoredTheme` → `setTheme`) still runs afterward and corrects `data-theme`/`localStorage` if the stored value was invalid; in the normal case the values already match, so there's no visible second flash.

### 7. `LandingPage.tsx`'s `#F0EBE3` → `bg-secondary`

The hardcoded section background (`#F0EBE3`, between `warm`'s `--background` `#FAF8F4` and `--secondary` `#E5DED6`) becomes `className="... bg-secondary"` (dropping the `style` prop entirely). `--secondary` is already defined for all three themes and is the closest existing token to the current color — no new token needed. The shade shifts slightly (`#F0EBE3` → `#E5DED6`) but stays within the same warm-beige family; this section is inside `PublicThemeLock` so it always renders in `warm` regardless of the viewer's theme preference anyway.

## Risks / Trade-offs

- **Palette values are a judgment call** — `lumen`/`sunless` numbers above are a reasonable starting point but should be eyeballed in-browser (contrast, hover states, borders) before considering this "done." Adjusting them is a pure CSS-value change, no structural impact.
- **`aria-label` change is test-visible** — `AppSection.test.tsx`'s existing assertion (`'Default theme (selected)'`) needs updating to `'Default theme'`, plus new assertions for `lumen`/`sunless` rows and click → `setTheme` behavior.
- **FOUC script and `readStoredTheme()` both read `localStorage['bookleaf-theme']`** — intentional light duplication of *reading* the key, but not of the valid-theme list (see Decision 6). If the storage key itself ever changes, both need updating.

## Migration Plan

Purely client-side. Existing `localStorage['bookleaf-theme'] = 'warm'` values remain valid (id unchanged). No backend/extension changes, no feature flag. Ships as a normal frontend deploy; rollback is a plain revert.
