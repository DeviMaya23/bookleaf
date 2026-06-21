## 1. Dependency

- [x] 1.1 Add `lucide-react` to `extensions/package.json` dependencies (match version used in `frontend/package.json`)
- [x] 1.2 Install and confirm lockfile updates cleanly

## 2. Icon swap

- [x] 2.1 Remove `MoonIcon`, `SunIcon`, `GearIcon` from `extensions/src/popup/App.tsx`
- [x] 2.2 Import `Moon`, `Sun`, `Settings` from `lucide-react` in `App.tsx`, sized `14` to match current icons
- [x] 2.3 Update `extensions/src/popup/Settings.tsx` to import `Moon`/`Sun` from `lucide-react` instead of from `./App`
- [x] 2.4 Update all usages of the removed icon components across both files

## 3. Spec sync

- [x] 3.1 Confirm `extension-popup-settings` spec wording ("gear icon" → "Settings icon") matches implementation after swap

## 4. Verification

- [x] 4.1 Manually verify icon appearance in both light and dark popup themes (main view toggle, Settings view toggle, header Settings icon)
- [x] 4.2 Run `npm run build` in `extensions/` and fix any issues
- [x] 4.3 Run `npm run lint` in `extensions/` and fix any issues (if a lint script exists) — no lint script exists in `extensions/package.json`; ran `npm run type-check` instead, no errors
- [x] 4.4 Run `npm run test` (vitest) in `extensions/` to confirm `Settings.test.ts` still passes
