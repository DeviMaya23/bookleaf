## Why

Bookleaf has no informational pages — no About, no Privacy Policy, and nothing explaining that images are sent to Google Vision API for AI-assisted folder suggestions (and future AI features). These are needed for basic transparency, and there's currently no shared layout for plain text-content pages to be built on.

## What Changes

- New shared `SimplePageLayout` component: nav (Bookleaf wordmark + back link), centered max-width column, normal scrolling, styled with the existing warm theme (`font-serif` headings, `font-sans` body, existing color tokens)
- Three new public pages built on `SimplePageLayout`, each with lorem-ipsum skeleton content (section headings + placeholder paragraphs) for the user to fill in later:
  - `AboutPage` at `/about`
  - `PrivacyPolicyPage` at `/privacy`
  - `AiNotesPage` at `/ai-notes` — explains that images are processed via Google Vision API for folder suggestions and future AI features
- Privacy Policy and AI Notes pages cross-link to each other
- `LandingPage` footer updated: the centered "Bookleaf" wordmark is removed and replaced with links to "About", "Privacy", and "AI Notes"
- All three new routes are public (outside `AuthGuard`), alongside `/` and `/callback`

## Capabilities

### New Capabilities
- `fe-content-pages`: Shared `SimplePageLayout` and the three informational pages (About, Privacy Policy, AI Notes), including the Privacy ↔ AI Notes cross-link

### Modified Capabilities
- `fe-landing-page`: Footer gains navigation links to About, Privacy, and AI Notes, replacing the centered wordmark

## Impact

- `frontend/src/components/SimplePageLayout.tsx` (new)
- `frontend/src/pages/AboutPage.tsx`, `PrivacyPolicyPage.tsx`, `AiNotesPage.tsx` (new) + tests
- `frontend/src/pages/LandingPage.tsx` — footer update + test
- `frontend/src/App.tsx` — new public routes for `/about`, `/privacy`, `/ai-notes`
