/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_KINDE_ISSUER_URL: string
  readonly VITE_KINDE_CLIENT_ID: string
  readonly VITE_KINDE_AUDIENCE: string
  readonly VITE_API_BASE_URL: string
  readonly VITE_APP_URL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
