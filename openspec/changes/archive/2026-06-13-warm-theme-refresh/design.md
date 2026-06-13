## Context

Today `frontend/src/index.css` defines shadcn's default grayscale tokens in `:root` plus a `.dark { ... }` block that nothing ever activates (no `@custom-variant dark` is defined anywhere, including the bundled `shadcn/tailwind.css`, so `dark:` utilities fall back to Tailwind v4's built-in `@media (prefers-color-scheme: dark)`). Only 3 components use `dark:` utilities (`avatar.tsx`, `dropdown-menu.tsx`, `context-menu.tsx`), and the effect is visually negligible.

The handover design (`/Users/devi/Downloads/applayouthandodver/project/App Layout v2.html`) defines a warm cream palette as a drop-in replacement for ~17 of the ~27 CSS variables referenced by `@theme inline`. The remaining variables split into "actively used, needs a value" (`--popover`, `--popover-foreground`, `--card-foreground` — consumed by `dropdown-menu.tsx`, `context-menu.tsx`, `dialog.tsx`) and "unused, leave alone" (`--sidebar-primary*`, `--sidebar-accent*`, `--sidebar-border`, `--sidebar-ring`, `--chart-1..5`).

This design also establishes the mechanism for a multi-theme system per the proposal: "warm" is the first of N named themes, not "light" with "dark" as the only alternative.

## Goals / Non-Goals

**Goals:**
- Apply the warm cream palette + DM Sans/Lora fonts from the handover, with no visual regression in components the handover doesn't explicitly cover (popovers, dialogs, context menus).
- Stand up a `ThemeProvider`/`useTheme()` whose underlying mechanism (a `data-theme` attribute + per-theme CSS blocks) scales to any number of named themes by adding a CSS block + a union member — no structural rework when theme #2 arrives.
- Keep today's behavior visually identical to "warm" being the only theme (provider is real but inert).

**Non-Goals:**
- Designing or implementing token values for any theme other than "warm" (no dark palette, no second light theme).
- Any theme-switcher UI.
- Font-size preferences.
- Server-persisted / account-level preferences (client `localStorage` only, per prior discussion).
- FOUC (flash-of-wrong-theme) handling — not relevant while only one theme exists; revisit when theme #2 ships.

## Decisions

### 1. Token structure: `[data-theme="warm"]` block, with `:root` mirroring it as a fallback
All warm-palette values live in `index.css` under `[data-theme="warm"] { ... }`. `:root` carries the same values so the page renders correctly before `ThemeProvider`'s effect runs (no FOUC, since "warm" is both the default and the only theme).

**Alternative considered**: put everything directly in `:root` (simplest). Rejected — it doesn't express "this is one of several named themes" and would need restructuring the moment a second theme is added, which is exactly the rework this change is meant to avoid.

### 2. `@custom-variant dark` scoped to `[data-theme="dark"]`, not `.dark`
```css
@custom-variant dark (&:where([data-theme="dark"], [data-theme="dark"] *));
```
This makes the existing `dark:` utilities (avatar, dropdown-menu, context-menu) key off the "dark" *theme* specifically, consistent with the data-theme model — rather than a generic `.dark` class or `prefers-color-scheme`. No `[data-theme="dark"]` CSS block is defined yet, so these utilities remain inert (matching their current negligible-impact state).

The stale `.dark { ... }` grayscale block is deleted entirely rather than ported — it doesn't match the new palette and nothing depends on its current values.

### 3. `ThemeProvider` / `useTheme()` location: `frontend/src/hooks/useTheme.tsx`
A single new file exporting `ThemeProvider` (context + effect that sets `document.documentElement.dataset.theme` and persists to `localStorage`) and `useTheme()` (returns `{ theme, setTheme }`, throws if called outside the provider).

**Alternatives considered**:
- New top-level `providers/` directory — introduces a new layer boundary not in the current directory structure (CLAUDE.md requires confirming new layer boundaries up front); unnecessary for a single provider.
- `lib/theme.ts` — `lib/` is reserved for shared domain modules (types + API wrappers like `images.ts`/`folders.ts`); a stateful context provider doesn't fit that shape.
- `hooks/` already holds generic, dependency-free hooks (`useDebouncedValue.ts`) and is the closest existing fit for a cross-cutting, feature-agnostic hook + its provider.

### 4. `Theme` type starts as a single-member union: `type Theme = 'warm'`
Adding a theme later = add a union member + a `[data-theme="<name>"]` CSS block. No changes to `ThemeProvider`, `useTheme`, persistence, or the custom variant.

**Alternative considered**: pre-declare `type Theme = 'warm' | 'light' | 'dark'` now. Rejected as speculative — TypeScript would allow selecting themes with no CSS behind them, and CLAUDE.md/CONVENTIONS.md both discourage building for hypothetical future requirements. The mechanism (decisions 1–3) is what's reusable; the type union is intentionally the cheap, mechanical part left for later.

### 5. Persistence: `localStorage` key `bookleaf-theme`
Read via a lazy `useState` initializer (falls back to `'warm'` if missing/unrecognized), written via an effect on change. Single source of truth is React state; `data-theme` and `localStorage` are both derived from it.

### 6. Provider placement: outermost in `main.tsx`
`ThemeProvider` wraps `KindeProvider`/`QueryClientProvider`/`BrowserRouter` — it has no dependency on auth or data state, and applying `data-theme` to `<html>` is a pure side effect of the theme preference.

### 7. Fill-in values for `--popover`, `--popover-foreground`, `--card-foreground`
Set equal to `--card` and `--foreground` respectively, mirroring how the current default theme treats popover/card surfaces (same background/foreground as the base card surface). `--sidebar-*` extended tokens and `--chart-*` are left at their current (unused) values.

### 8. Fonts: `@fontsource-variable/dm-sans` + `@fontsource-variable/lora`
Both confirmed available at `5.2.8`. DM Sans becomes `--font-sans` (body/UI, replacing Geist Variable). Lora is added as a new `--font-serif` token, applied via a `font-serif` utility class only on the "Bookleaf" wordmark in the sidebar — no other components switch fonts.

## Risks / Trade-offs

- **Alpha-based tokens** (`--accent`, `--border`, `--input`, `--ring` become `rgba(...)` instead of solid colors) composite differently depending on the surface behind them → Mitigation: manual visual pass over dropdowns/dialogs/context menus after the swap (no automated visual regression suite exists to lean on).
- **`--radius` changes 0.625rem → 0.5rem**, subtly tightening corners app-wide → accepted as part of the intentional redesign.
- **The 3 existing `dark:` usages stop responding to OS `prefers-color-scheme`** once `@custom-variant dark` is scoped to `[data-theme="dark"]` (which nothing sets) → already-negligible visual effect today; confirmed during exploration that this is imperceptible in normal use.
- **`useTheme()` outside `ThemeProvider`** → guarded with a thrown error (standard context-hook pattern), surfaced immediately in development rather than silently defaulting.
- **Trade-off**: `Theme = 'warm'` as a single-member union means TypeScript won't catch an attempt to call `setTheme('dark')` until that member exists — intentional; this is the seam where the next theme gets added later.

## Migration Plan

Purely client-side visual + scaffolding change — no data migration, no feature flag, no backend/extension changes. Ships as a normal frontend deploy. Rollback is a plain revert; `localStorage['bookleaf-theme']` only ever holds `'warm'` until a second theme exists, so there's no stale-value compatibility concern.
