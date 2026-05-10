## 1. Vite Project Init

- [ ] 1.1 Scaffold `frontend/` with `npm create vite@latest` using React + TypeScript template
- [ ] 1.2 Add `.nvmrc` pinned to `20` and `engines: { node: ">=20" }` to `package.json`
- [ ] 1.3 Verify `npm run dev` starts cleanly and serves on `localhost:5173`

## 2. TypeScript Config

- [ ] 2.1 Enable strict mode in `tsconfig.json`
- [ ] 2.2 Add `@/` path alias pointing to `src/` in `tsconfig.json` and `vite.config.ts`
- [ ] 2.3 Verify `npx tsc --noEmit` passes with no errors

## 3. Tailwind CSS

- [ ] 3.1 Install `tailwindcss`, `postcss`, and `autoprefixer` as dev dependencies
- [ ] 3.2 Generate `tailwind.config.ts` and `postcss.config.js` via `npx tailwindcss init -p`
- [ ] 3.3 Set content paths in `tailwind.config.ts` to cover `src/**/*.{ts,tsx}`
- [ ] 3.4 Replace `src/index.css` contents with the three Tailwind directives
- [ ] 3.5 Verify a utility class applied in `App.tsx` renders correctly in the browser

## 4. shadcn/ui Init

- [ ] 4.1 Install shadcn peer dependencies (`clsx`, `tailwind-merge`, `class-variance-authority`)
- [ ] 4.2 Run `npx shadcn init` and configure `components.json` (style: default, base colour: neutral, CSS variables: yes, alias: `@/`)
- [ ] 4.3 Confirm `src/lib/utils.ts` exists and exports the `cn` helper
- [ ] 4.4 Confirm `src/components/ui/` directory exists
- [ ] 4.5 Smoke-test by running `npx shadcn add button` and verifying `src/components/ui/button.tsx` is created

## 5. Additional Dependencies

- [ ] 5.1 Install `react-router-dom`
- [ ] 5.2 Install `@kinde-oss/kinde-auth-react`
- [ ] 5.3 Install `@tanstack/react-query`
- [ ] 5.4 Verify `npm install` completes without errors and all three packages appear in `node_modules/`

## 6. Environment Variables

- [ ] 6.1 Create `frontend/.env.example` with placeholder values for `VITE_API_BASE_URL`, `VITE_KINDE_CLIENT_ID`, `VITE_KINDE_ISSUER_URL`, `VITE_KINDE_REDIRECT_URL`, and `VITE_KINDE_LOGOUT_REDIRECT_URL`
- [ ] 6.2 Add `frontend/.env` and `frontend/.env.local` to `.gitignore`
