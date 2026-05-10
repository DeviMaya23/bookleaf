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

## Risks / Trade-offs

[Tailwind v4 vs v3 API differences] → Pin to Tailwind v3 for now; shadcn's stable config tooling targets v3. Upgrade path is documented when shadcn ships full v4 support.

[Node version drift] → Add `.nvmrc` / `engines` field in `package.json` pinned to Node 20 LTS to keep CI and local environments aligned.

[Monorepo tooling] → No Turborepo or workspace setup for MVP. `frontend/` is a self-contained package; the backend is Go. Running both means two terminal tabs. Acceptable for MVP scope.
