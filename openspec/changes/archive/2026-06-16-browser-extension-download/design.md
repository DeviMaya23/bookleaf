## Context

The extension builds are ready but not yet store-approved. Two surfaces need download links: a new public `/extensions` page (no auth required) and the existing Settings modal (auth'd app). Both surfaces point to the same files, so the URLs need to be defined once and shared.

Extension files are hosted on a new public Cloudflare R2 bucket (`bookleaf-public`) with a custom domain (e.g. `downloads.bookleaf.app`). Files use fixed key names (`bookleaf-extension.xpi`, `bookleaf-extension.zip`) so the URLs never change between releases — new builds overwrite the same keys.

The Firefox `.xpi` is signed via `web-ext sign --channel=unlisted`, meaning Mozilla signs it without listing it in the store. It installs on regular Firefox without developer mode. The Chrome `.zip` requires the user to load it unpacked in developer mode — this friction is acknowledged and documented explicitly in the UI.

## Goals / Non-Goals

**Goals:**
- Single source of truth for download URLs, shared between the public page and settings section
- Public page follows the existing `SimplePageLayout` pattern (no new layout components)
- Settings section follows the existing section component pattern (`AppSection`, `AccountSection`, etc.)
- Install instructions are honest about friction (especially for Chrome)

**Non-Goals:**
- Auto-update mechanism for installed extensions
- Version checking or "new version available" UI
- Any backend involvement — these are static download links
- R2 bucket creation or wrangler upload scripts — infrastructure setup is out of scope for this change

## Decisions

### URL constants in `src/lib/downloads.ts`

The same two URLs appear in two separate parts of the app (public page + settings modal). Inlining them would scatter a detail that changes together. A small dedicated module in `src/lib/` — the existing home for cross-cutting utilities — is the lightest abstraction that solves this.

**Alternative considered:** Defining the URLs as props passed down from a parent. Rejected — there's no common parent between `LandingPage`/`ExtensionsPage` and `SettingsModal`, so props would require threading through unrelated components.

### Extensions as a new `fe-content-pages` route, not a standalone spec

The Extensions page is structurally identical to About, Privacy, and AI Notes — it uses `SimplePageLayout`, lives in `src/pages/`, and has a route in `App.tsx`. It belongs in the same spec rather than a separate one. The only new thing is the content (download links + instructions).

### Extensions as a new Settings modal section

The modal's section pattern is already established: add an entry to `SECTIONS`, create a `*Section.tsx` component, render it conditionally. Adding `extensions` follows this pattern exactly — no structural change to `SettingsModal.tsx` beyond the new entry and render branch.

### No new npm dependencies

Install instructions are plain text + anchor tags. No download-tracking, no version-fetching, no progress UI. Everything needed is already in the codebase.

## Risks / Trade-offs

- **Stale URLs if the R2 domain changes** → Both URLs are defined in one file (`src/lib/downloads.ts`), so a domain change is a one-line update per URL. Low risk.
- **Chrome install friction may discourage users** → Accepted. The alternative (hiding the steps) would cause more frustration at install time. Being explicit is the right call until store approval lands.
- **Firefox auto-install prompt depends on browser behaviour** → When the `.xpi` is served with `Content-Type: application/x-xpinstall`, Firefox should trigger the install dialog on download. If it doesn't (e.g. browser settings differ), the fallback manual steps cover it. The instructions cover both paths.
