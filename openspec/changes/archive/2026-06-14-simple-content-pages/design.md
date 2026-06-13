## Context

`LandingPage.tsx` (just implemented) is a single non-scrollable (`h-dvh overflow-hidden`) viewport with a footer that currently shows only a centered "Bookleaf" wordmark. `App.tsx` has three top-level routes (`/`, `/callback`, `/app/*`) plus a catch-all. There's no existing pattern for a plain text-content page — every page so far is either a tightly laid-out marketing page or an app shell.

## Goals / Non-Goals

**Goals:**
- A small, reusable `SimplePageLayout` for text-content pages, visually consistent with the warm theme (serif headings, sans body, existing color tokens), that scrolls normally
- Three new pages (`/about`, `/privacy`, `/ai-notes`) built on it, with real section headings but lorem-ipsum body text
- Privacy Policy and AI Notes cross-link to each other
- `LandingPage` footer links to all three

**Non-Goals:**
- Final copy — all body text is lorem ipsum; only section headings are real (so the page has the right shape when the user writes the actual content later)
- `<title>`/meta tag management — not addressed, consistent with the rest of the app (no existing pattern for it)
- Any content beyond the three listed pages

## Decisions

### 1. `SimplePageLayout` shape

A single component at `frontend/src/components/SimplePageLayout.tsx`:
- `<nav>`: "Bookleaf" wordmark (font-serif, same styling as `LandingPage`'s nav), as a `<Link to="/">`. This doubles as the "back" affordance — no separate "← Back" element, to avoid two competing links to the same destination in a small nav bar.
- `<main>`: centered column (`max-w-2xl mx-auto`), normal vertical scroll (`min-h-dvh`, no `overflow-hidden`), `font-sans` body text via the existing global `body` rule, `font-serif` for headings
- Takes a `title` prop (rendered as the page's `<h1>`) and `children` for the body content

Alternative considered: give each page its own full nav/layout (copy-pasted from `LandingPage`). Rejected — three near-identical layouts is exactly the "three similar lines" case where a shared component pays for itself immediately.

### 2. Skeleton content: real headings, lorem-ipsum bodies

Each page gets 2-4 `<h2>` section headings that describe what the section will actually cover, with lorem-ipsum `<p>` placeholders underneath. This gives the page its intended structure (so the user can see/edit the shape) without writing real copy now.

- **About**: headings like "What Bookleaf is", "Why it exists" — generic, since the user is hand-writing this themselves
- **Privacy Policy**: headings like "What we collect", "How we use it", "Your choices"
- **AI Notes**: headings reflecting the known fact from the proposal — e.g. "How AI is used here", "What gets sent to Google Vision API", "Opting out" — bodies still lorem ipsum

Alternative considered: fully lorem-ipsum including headings. Rejected — the headings are "free" structural information the user asked for implicitly by listing what each page should cover, and they cost nothing to get right now versus needing to be restructured later.

### 3. Cross-link via inline `<Link>`

`PrivacyPolicyPage` and `AiNotesPage` each end with a short paragraph containing a `<Link>` to the other (e.g. "See also: AI Notes" / "See also: Privacy Policy"). Plain React Router `<Link>`, no new component.

### 4. Footer: links replace wordmark

`LandingPage`'s footer changes from:
```tsx
<footer className="flex items-center justify-center border-t border-border px-8 py-4">
  <span className="font-serif text-sm font-semibold">Bookleaf</span>
</footer>
```
to a `justify-center` (or `justify-between` if a wordmark were kept — it's not) row of three `<Link>`s: "About", "Privacy", "AI Notes", separated by `·` or spacing, using `text-sm text-muted-foreground` consistent with other secondary text on the page.

### 5. Routes: flat, public, alongside `/` and `/callback`

```tsx
<Route path="/about" element={<AboutPage />} />
<Route path="/privacy" element={<PrivacyPolicyPage />} />
<Route path="/ai-notes" element={<AiNotesPage />} />
```
No nesting/layout route needed — `SimplePageLayout` is a component each page renders internally, not a route-level layout, since these three routes don't share any other route-level concern (no guard, no shared loader).

## Risks / Trade-offs

- **[Risk]** An authenticated user clicking "About"/"Privacy"/"AI Notes" from somewhere, then clicking the wordmark to go "back" to `/`, will be bounced to `/app` (per `fe-landing-page`'s existing authenticated-redirect rule) → **Mitigation**: this is existing, intentional behavior for `/`; not specially handled here, just noted as expected.
- **[Risk]** Lorem-ipsum body text could ship to production if forgotten → **Mitigation**: out of scope to prevent here (no build-time lorem-ipsum linter); relies on the user replacing it before launch, same as the landing page's placeholder copy.

## Migration Plan

Purely additive (new component, new pages, new routes) plus one footer edit to `LandingPage`. No data/backend changes. Rollback is a straight revert.

## Open Questions

None outstanding — naming, cross-links, and content approach were settled during exploration.
