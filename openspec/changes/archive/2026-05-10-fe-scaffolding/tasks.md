## 1. Vite Project Init

- [x] 1.1 Scaffold `frontend/` with `npm create vite@latest` using React + TypeScript template
- [x] 1.2 Add `.nvmrc` pinned to `20` and `engines: { node: ">=20" }` to `package.json`
- [x] 1.3 Verify `npm run dev` starts cleanly and serves on `localhost:5173`

## 2. TypeScript Config

- [x] 2.1 Enable strict mode in `tsconfig.json`
- [x] 2.2 Add `@/` path alias pointing to `src/` in `tsconfig.json` and `vite.config.ts`
- [x] 2.3 Verify `npx tsc --noEmit` passes with no errors

## 3. Tailwind CSS

- [x] 3.1 Install `tailwindcss`, `postcss`, and `autoprefixer` as dev dependencies
- [x] 3.2 Generate `tailwind.config.ts` and `postcss.config.js` via `npx tailwindcss init -p`
- [x] 3.3 Set content paths in `tailwind.config.ts` to cover `src/**/*.{ts,tsx}`
- [x] 3.4 Replace `src/index.css` contents with the three Tailwind directives
- [x] 3.5 Verify a utility class applied in `App.tsx` renders correctly in the browser

## 4. shadcn/ui Init

- [x] 4.1 Install shadcn peer dependencies (`clsx`, `tailwind-merge`, `class-variance-authority`)
- [x] 4.2 Run `npx shadcn init` and configure `components.json` (style: default, base colour: neutral, CSS variables: yes, alias: `@/`)
- [x] 4.3 Confirm `src/lib/utils.ts` exists and exports the `cn` helper
- [x] 4.4 Confirm `src/components/ui/` directory exists
- [x] 4.5 Smoke-test by running `npx shadcn add button` and verifying `src/components/ui/button.tsx` is created

## 5. Additional Dependencies

- [x] 5.1 Install `react-router-dom`
- [x] 5.2 Install `@kinde-oss/kinde-auth-react`
- [x] 5.3 Install `@tanstack/react-query`
- [x] 5.4 Verify `npm install` completes without errors and all three packages appear in `node_modules/`

## 6. Environment Variables

- [x] 6.1 Create `frontend/.env.example` with placeholder values for `VITE_API_BASE_URL`, `VITE_KINDE_CLIENT_ID`, `VITE_KINDE_ISSUER_URL`, `VITE_KINDE_REDIRECT_URL`, and `VITE_KINDE_LOGOUT_REDIRECT_URL`
- [x] 6.2 Add `frontend/.env` and `frontend/.env.local` to `.gitignore`
