## 1. URL Constants

- [x] 1.1 Create `src/lib/downloads.ts` exporting `EXTENSION_FIREFOX_URL` and `EXTENSION_CHROME_URL` pointing to their respective R2 files

## 2. Extensions Public Page

- [x] 2.1 Create `src/pages/ExtensionsPage.tsx` using `SimplePageLayout` with Firefox and Chrome sections, download links from `downloads.ts`, and install instructions for each
- [x] 2.2 Add `/extensions` route to `src/App.tsx`
- [x] 2.3 Add "Extensions" footer link to `src/pages/LandingPage.tsx`

## 3. Settings Modal Extensions Section

- [x] 3.1 Create `src/features/settings/components/ExtensionsSection.tsx` with Firefox and Chrome subsections, download links from `downloads.ts`, and install instructions
- [x] 3.2 Add `{ id: 'extensions', label: 'Extensions' }` to `SECTIONS` in `SettingsModal.tsx` and render `<ExtensionsSection />` for that section

## 4. Build Check

- [x] 4.1 Run `npm run build` from `frontend/` and fix any errors
- [x] 4.2 Run `npm run lint` from `frontend/` and fix any warnings or errors
