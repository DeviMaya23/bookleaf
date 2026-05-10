## Context

The backend is Go + Echo with Kinde JWT auth. The frontend needs to be a standalone Vite app living at `frontend/` in the same repo. This change covers only the project scaffold — getting to a clean `npm run dev` with no wiring code. All library choices must be compatible with Kinde's React SDK and the existing backend auth model.

## Goals / Non-Goals

**Goals:**
- A `frontend/` directory that starts cleanly with `npm install && npm run dev`
- Tailwind CSS configured and working
- shadcn/ui initialised and ready to add components via CLI
- React Router, Kinde React SDK, TanStack Query present as installed (but unwired) dependencies
- `.env.example` documenting all required `VITE_*` vars

**Non-Goals:**
- Any page components, routing logic, or provider wiring
- API client implementation
- Authentication flow implementation
- Any backend changes

## Decisions

**Vite over CRA / Next.js**
Vite is the right choice for an SPA backed by a separate Go API. CRA is deprecated; Next.js adds SSR complexity that provides no benefit when the backend is a separate service.

**TanStack Query for async state**
Installed now (unwired) so it's available when feature work begins. It handles mutations + cache invalidation cleanly, which will be needed for upload, folder management, and gallery — better fit than SWR for a write-heavy app.

**shadcn/ui over a component library**
shadcn copies components into `components/ui/` rather than importing from a package, so components are fully owned and customisable. The CLI-based add workflow means only used components live in the repo.

**`src/` flat structure for now**
Subdirectories (`hooks/`, `components/`, `pages/`, `lib/`, `types/`) will be created as features are added. Forcing them empty at scaffold time adds noise.

**`components.json` at repo root of `frontend/`**
shadcn's CLI expects `components.json` at the package root. Placing it there keeps `npx shadcn add <component>` working without flags.

**Tailwind v4 + shadcn v4**
shadcn v4 (current stable) requires Tailwind v4 — the two are coupled. Tailwind v4 drops `tailwind.config.ts` in favour of CSS-based configuration; content paths are auto-detected. shadcn's init generates all required CSS variables and base layer rules directly into `src/index.css`. No `tailwind.config.ts` is needed for the scaffold.

## Risks / Trade-offs

[Tailwind v4 CSS-based config] → No `tailwind.config.ts` theme extension at scaffold time; custom tokens are added via CSS variables in `index.css` instead. This is the idiomatic v4 approach and is fully supported by shadcn v4.

[Node version drift] → Add `.nvmrc` / `engines` field in `package.json` pinned to Node 20 LTS to keep CI and local environments aligned.

[Monorepo tooling] → No Turborepo or workspace setup for MVP. `frontend/` is a self-contained package; the backend is Go. Running both means two terminal tabs. Acceptable for MVP scope.
