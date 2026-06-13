## Why

The app currently uses shadcn's default grayscale palette and the Geist font, which doesn't match the warm, editorial look defined in the "App Layout v2" design handover. We also have no mechanism for a user-selectable color theme — `dark:` utilities exist in a handful of components but only respond to OS `prefers-color-scheme`, and the `.dark` CSS variable block is unused dead code. This change applies the new visual design now, and puts a real (if currently inert) theme-preference mechanism in place so a future dark theme can be added without re-architecting.

## What Changes

- Introduce a named-theme model: the new cream palette is the **"warm"** theme — the first of several named themes (warm, light, dark, ...), not "light mode". A separate white/light theme may be added later.
- Replace the current `:root` design tokens in `frontend/src/index.css` with a `[data-theme="warm"]` block containing the warm cream palette from the handover (`--background`, `--foreground`, `--card`, `--muted`, `--accent`, `--secondary`, `--primary`, `--border`, `--input`, `--ring`, `--sidebar`, `--radius`, etc.). `:root` keeps a matching fallback so styling is correct before the provider applies `data-theme`.
- Fill in values for tokens the handover doesn't specify but that are actively used (`--popover`, `--popover-foreground`, `--card-foreground`), by analogy with the new `--card`/`--foreground` values.
- Leave unused tokens (`--sidebar-primary*`, `--sidebar-accent*`, `--sidebar-border`, `--sidebar-ring`, `--chart-1..5`) as-is.
- Replace the Geist Variable font with DM Sans (body/UI) and Lora (serif, used only for the "Bookleaf" wordmark in the sidebar).
- Add `@custom-variant dark (&:where([data-theme="dark"], [data-theme="dark"] *));` to `index.css` so `dark:` utilities key off the "dark" theme specifically, rather than `prefers-color-scheme`.
- Remove the existing (unused, mismatched) `.dark { ... }` token block; a `[data-theme="dark"]` block is left empty/omitted so selecting it is currently a no-op until real dark tokens are designed.
- Add a `ThemeProvider` (React context) that:
  - Tracks a `theme` preference (named theme, e.g. `'warm' | 'dark'`, extensible), persisted to `localStorage`.
  - Applies the active theme via a `data-theme` attribute on `<html>`.
  - Exposes a `useTheme()` hook for reading/setting the preference.
  - Is mounted in `main.tsx` alongside the existing provider stack.
  - Defaults to `'warm'`; no UI control is exposed yet, so behavior is currently inert (only "warm" is selectable/defined today).

## Capabilities

### New Capabilities
- `fe-theme`: Design token palette/fonts and the client-side theme-preference provider (persistence, `data-theme` attribute application, `useTheme()` hook).

### Modified Capabilities
(none — no existing capability specs define color tokens, fonts, or theme state)

## Impact

- `frontend/src/index.css`: token values, font imports, `@custom-variant dark`, removal of stale `.dark` block.
- `frontend/package.json`: swap `@fontsource-variable/geist` for DM Sans + Lora font packages.
- `frontend/src/main.tsx`: new `ThemeProvider` wrapping the app.
- New files: `ThemeProvider` component + `useTheme` hook (location TBD in design.md).
- Visual-only change for end users today (new color/font scheme); no behavioral change to existing features. No backend or extension changes.
