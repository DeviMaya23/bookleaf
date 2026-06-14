## 1. Shared layout

- [x] 1.1 Create `frontend/src/components/SimplePageLayout.tsx`: nav with "Bookleaf" wordmark linking to `/`, centered max-width scrolling column, `title` prop rendered as `<h1>`, `children` for body content
- [x] 1.2 Add `SimplePageLayout.test.tsx`: renders nav/wordmark link to `/`, title, and children

## 2. Content pages

- [x] 2.1 Create `frontend/src/pages/AboutPage.tsx` using `SimplePageLayout`, with section headings + lorem-ipsum body placeholders
- [x] 2.2 Create `frontend/src/pages/PrivacyPolicyPage.tsx` using `SimplePageLayout`, with section headings (data collection, usage, user choices) + lorem-ipsum placeholders, and a link to `/ai-notes`
- [x] 2.3 Create `frontend/src/pages/AiNotesPage.tsx` using `SimplePageLayout`, with section headings (how AI features work, what is sent to Google Vision API, opting out) + lorem-ipsum placeholders, and a link to `/privacy`
- [x] 2.4 Add tests for `AboutPage`, `PrivacyPolicyPage`, `AiNotesPage`: each renders its title/section headings; Privacy and AI Notes pages render their cross-links

## 3. Routing

- [x] 3.1 Add `/about`, `/privacy`, `/ai-notes` routes to `App.tsx`, alongside `/` and `/callback`

## 4. Landing page footer

- [x] 4.1 Update `LandingPage.tsx` footer: replace the centered "Bookleaf" wordmark with links to "About" (`/about`), "Privacy" (`/privacy`), and "AI Notes" (`/ai-notes`)
- [x] 4.2 Update `LandingPage.test.tsx` to assert the new footer links

## 5. Final checks

- [x] 5.1 Run `npm run build` and fix any issues
- [x] 5.2 Run `npm run lint` and fix any issues
