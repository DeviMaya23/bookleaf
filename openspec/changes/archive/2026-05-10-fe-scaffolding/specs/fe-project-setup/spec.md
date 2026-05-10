## ADDED Requirements

### Requirement: Vite project initialised
A `frontend/` directory SHALL exist at the repo root containing a Vite + React + TypeScript project. The project MUST start successfully with `npm install && npm run dev` and serve the default page on `localhost:5173`.

#### Scenario: Dev server starts
- **WHEN** developer runs `npm run dev` inside `frontend/`
- **THEN** Vite starts without errors and serves the app at `localhost:5173`

#### Scenario: Production build succeeds
- **WHEN** developer runs `npm run build` inside `frontend/`
- **THEN** Vite produces a `dist/` directory without errors

### Requirement: TypeScript configured
The project SHALL include a `tsconfig.json` and `tsconfig.node.json` with strict mode enabled and path aliases configured (`@/` → `src/`).

#### Scenario: Type check passes on clean scaffold
- **WHEN** developer runs `npx tsc --noEmit` inside `frontend/`
- **THEN** TypeScript reports no errors

### Requirement: Tailwind CSS configured
Tailwind CSS v3 SHALL be installed and configured. `tailwind.config.ts` MUST include content paths covering `src/**/*.{ts,tsx}`. A base `src/index.css` MUST import the three Tailwind directives (`@tailwind base`, `@tailwind components`, `@tailwind utilities`).

#### Scenario: Tailwind utility classes apply
- **WHEN** a utility class (e.g. `bg-red-500`) is used in `App.tsx`
- **THEN** the compiled CSS includes that utility and applies the style in the browser

### Requirement: shadcn/ui initialised
`components.json` SHALL exist at `frontend/` root with the project's shadcn configuration (style, base colour, CSS variables, path aliases). `src/lib/utils.ts` MUST export the `cn` helper. The `src/components/ui/` directory MUST exist and be ready to receive components via `npx shadcn add`.

#### Scenario: shadcn component can be added
- **WHEN** developer runs `npx shadcn add button` inside `frontend/`
- **THEN** `src/components/ui/button.tsx` is created without errors

### Requirement: Dependencies installed
`package.json` SHALL declare the following dependencies: `react-router-dom`, `@kinde-oss/kinde-auth-react`, `@tanstack/react-query`. These MUST be present in `node_modules` after `npm install` but require no wiring code in this change.

#### Scenario: All dependencies resolve
- **WHEN** developer runs `npm install` inside `frontend/`
- **THEN** all packages install without errors and `node_modules/` is populated

### Requirement: Environment variables documented
A `.env.example` file SHALL exist at `frontend/` root listing all required `VITE_*` variables with placeholder values and inline comments describing their purpose.

#### Scenario: Example env file is present
- **WHEN** developer clones the repo and navigates to `frontend/`
- **THEN** `.env.example` exists and lists `VITE_API_BASE_URL`, `VITE_KINDE_CLIENT_ID`, `VITE_KINDE_ISSUER_URL`, `VITE_KINDE_REDIRECT_URL`, and `VITE_KINDE_LOGOUT_REDIRECT_URL`

### Requirement: Node version pinned
A `.nvmrc` file SHALL exist at `frontend/` root pinning Node 20 LTS. `package.json` MUST include an `engines` field specifying `"node": ">=20"`.

#### Scenario: Node version is discoverable
- **WHEN** developer runs `cat frontend/.nvmrc`
- **THEN** the file outputs `20`
