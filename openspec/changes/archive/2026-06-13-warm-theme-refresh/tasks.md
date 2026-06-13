## 1. Fonts

- [x] 1.1 Remove `@fontsource-variable/geist` from `frontend/package.json`; add `@fontsource-variable/dm-sans` and `@fontsource-variable/lora` (`^5.2.8`)
- [x] 1.2 In `index.css`, replace the `@import "@fontsource-variable/geist"` with imports for DM Sans and Lora
- [x] 1.3 Update `--font-sans` to `'DM Sans Variable', sans-serif`; add `--font-serif: 'Lora Variable', Georgia, serif`
- [x] 1.4 Apply `font-serif` to the "Bookleaf" wordmark in `FolderSidebar` (the only place Lora is used)

## 2. Design tokens

- [x] 2.1 Add `[data-theme="warm"]` block to `index.css` with the warm cream palette values from the handover (`--background`, `--foreground`, `--card`, `--muted`, `--muted-foreground`, `--accent`, `--accent-foreground`, `--secondary`, `--secondary-foreground`, `--primary`, `--primary-foreground`, `--destructive`, `--border`, `--input`, `--ring`, `--sidebar`, `--radius`)
- [x] 2.2 Add `--popover`, `--popover-foreground`, `--card-foreground` to the `[data-theme="warm"]` block, set equal to `--card`/`--foreground` per design.md
- [x] 2.3 Mirror the `[data-theme="warm"]` block's values into `:root` as the pre-hydration fallback
- [x] 2.4 Remove the existing `.dark { ... }` block entirely
- [x] 2.5 Add `@custom-variant dark (&:where([data-theme="dark"], [data-theme="dark"] *));` (no `[data-theme="dark"]` block yet — intentionally inert)
- [x] 2.6 Leave `--sidebar-primary*`, `--sidebar-accent*`, `--sidebar-border`, `--sidebar-ring`, `--chart-1..5` untouched

## 3. ThemeProvider

- [x] 3.1 Create `frontend/src/hooks/useTheme.tsx`: `type Theme = 'warm'`, `ThemeProvider` component, `useTheme()` hook (`{ theme, setTheme }`, throws if used outside provider)
- [x] 3.2 `ThemeProvider`: lazy-init state from `localStorage['bookleaf-theme']` (fallback `'warm'`); effect syncs `document.documentElement.dataset.theme` and `localStorage` on change
- [x] 3.3 Mount `ThemeProvider` as the outermost provider in `frontend/src/main.tsx`
- [x] 3.4 Write `frontend/src/hooks/useTheme.test.tsx` covering: default theme with no stored preference, restoring a stored preference, `setTheme` updates `data-theme` + `localStorage`, `useTheme()` outside `ThemeProvider` throws

## 4. Verification

- [x] 4.1 Manual visual pass over dropdown menus, context menus, and dialogs to confirm `--popover`/`--popover-foreground`/`--card-foreground` and the alpha-based `--accent`/`--border`/`--input`/`--ring` tokens render correctly against the new palette
- [x] 4.2 Run `npm run build` and fix any resulting issues
