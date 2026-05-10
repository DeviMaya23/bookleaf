## Why

The backend is now feature-complete for the MVP; a frontend is needed to expose those capabilities to users. Establishing the project scaffold first — before any feature work — ensures a consistent foundation (tooling, auth wiring, routing, API client) that all future UI work builds on.

## What Changes

- New `frontend/` directory at repo root with a Vite + React + TypeScript project
- Tailwind CSS installed and configured (`tailwind.config.ts`, `postcss.config.js`, base CSS import)
- shadcn/ui initialised (`components.json`, `lib/utils.ts`, `components/ui/` drop zone ready)
- React Router, Kinde React SDK, and TanStack Query installed as dependencies (no wiring code)
- `.env.example` documenting the required `VITE_*` environment variables

## Capabilities

### New Capabilities

- `fe-project-setup`: All project config files and dependency installs that produce a clean, runnable empty app — `vite.config.ts`, `tsconfig.json`, `tailwind.config.ts`, `postcss.config.js`, `components.json`, `package.json`, `index.html`, `src/main.tsx`, `src/App.tsx`, `src/index.css`, `lib/utils.ts`

### Modified Capabilities

## Impact

- New top-level `frontend/` directory; no changes to `backend/`
- Adds Node.js / npm as a repo dependency (frontend only)
- `.env.example` documents 5 new `VITE_*` vars: `VITE_API_BASE_URL`, `VITE_KINDE_CLIENT_ID`, `VITE_KINDE_ISSUER_URL`, `VITE_KINDE_REDIRECT_URL`, `VITE_KINDE_LOGOUT_REDIRECT_URL`
