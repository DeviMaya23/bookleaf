import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { KindeProvider } from '@kinde-oss/kinde-auth-react'
import './index.css'
import App from './App.tsx'

const kindeVars = {
  VITE_KINDE_CLIENT_ID: import.meta.env.VITE_KINDE_CLIENT_ID,
  VITE_KINDE_ISSUER_URL: import.meta.env.VITE_KINDE_ISSUER_URL,
  VITE_KINDE_REDIRECT_URL: import.meta.env.VITE_KINDE_REDIRECT_URL,
  VITE_KINDE_LOGOUT_REDIRECT_URL: import.meta.env.VITE_KINDE_LOGOUT_REDIRECT_URL,
}

Object.entries(kindeVars).forEach(([key, value]) => {
  if (!value) console.warn(`[bookleaf] Missing required env var: ${key}`)
})

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <KindeProvider
      clientId={kindeVars.VITE_KINDE_CLIENT_ID}
      domain={kindeVars.VITE_KINDE_ISSUER_URL}
      redirectUri={kindeVars.VITE_KINDE_REDIRECT_URL}
      logoutUri={kindeVars.VITE_KINDE_LOGOUT_REDIRECT_URL}
    >
      <BrowserRouter>
        <App />
      </BrowserRouter>
    </KindeProvider>
  </StrictMode>,
)
