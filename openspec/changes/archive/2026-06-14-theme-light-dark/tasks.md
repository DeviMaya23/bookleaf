## 1. Design tokens (`index.css`)

- [x] 1.1 Add `[data-theme="lumen"]` block with the full token set per design.md Decision 1 (`background`, `foreground`, `card`, `card-foreground`, `popover`, `popover-foreground`, `primary`, `primary-foreground`, `secondary`, `secondary-foreground`, `muted`, `muted-foreground`, `accent`, `accent-foreground`, `destructive`, `border`, `input`, `ring`, `radius`, `sidebar`)
- [x] 1.2 Add `[data-theme="sunless"]` block with the full token set per design.md Decision 2
- [x] 1.3 Re-point `@custom-variant dark` from `[data-theme="dark"]` to `[data-theme="sunless"]`

## 2. Theme type & provider (`useTheme.tsx`)

- [x] 2.1 Extend `Theme` to `'warm' | 'lumen' | 'sunless'`
- [x] 2.2 Update `readStoredTheme()` to validate against all three values, defaulting to `'warm'`
- [x] 2.3 Update `useTheme.test.tsx`: add cases for restoring `lumen`/`sunless` from `localStorage` and `setTheme('lumen' | 'sunless')` updating `data-theme` + `localStorage`

## 3. Settings theme picker (`AppSection.tsx`)

- [x] 3.1 Update `THEME_OPTIONS` to a 3-entry `Record<Theme, {label, sub, swatches}>` per design.md Decision 4 table (warm/lumen/sunless labels, sub-copy, swatches)
- [x] 3.2 Replace the single static row with a `.map()` over `THEME_OPTIONS`: each row is a radio (`name="theme"`, `checked={theme === key}`, `onChange={() => setTheme(key as Theme)}`), remove `disabled`/`readOnly`
- [x] 3.3 Update `aria-label` from `"${label} theme (selected)"` to `"${label} theme"`
- [x] 3.4 Update `AppSection.test.tsx`: fix the existing aria-label assertion, add cases asserting all three rows render with correct selected state, and that clicking an unselected row calls `setTheme` with that theme's id

## 4. Public-route theme lock

- [x] 4.1 Create `frontend/src/components/PublicThemeLock.tsx` — `Outlet`-based layout-route component rendering `<div data-theme="warm"><Outlet /></div>`, mirroring `AuthGuard`'s pattern
- [x] 4.2 In `App.tsx`, wrap `/`, `/about`, `/privacy`, `/ai-notes` as children of `<Route element={<PublicThemeLock />}>`
- [x] 4.3 Add `PublicThemeLock.test.tsx` (pattern per `AuthGuard.test.tsx`): renders `<Outlet />` content inside a `data-theme="warm"` wrapper

## 5. Landing page hardcoded color

- [x] 5.1 In `LandingPage.tsx`, remove `style={{ backgroundColor: '#F0EBE3' }}` and add `bg-secondary` to the section's `className`

## 6. FOUC prevention (`index.html`)

- [x] 6.1 Add inline script as the first element in `<head>`, before any stylesheet `<link>`: `document.documentElement.dataset.theme = localStorage.getItem('bookleaf-theme') || 'warm'`

## 7. Verification

- [x] 7.1 Manual visual pass over `lumen` and `sunless`: cards, dropdown/context menus, dialogs, avatar borders, destructive-action states — confirm contrast and adjust token values from design.md if needed
- [x] 7.2 Manual check: with `lumen`/`sunless` selected in Settings, confirm `/`, `/about`, `/privacy`, `/ai-notes` still render in `warm` while `/app` reflects the selected theme
- [x] 7.3 Run `npm run build` and `npm run lint`; fix any resulting issues
