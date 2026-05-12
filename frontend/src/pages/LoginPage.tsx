import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { Navigate, useLocation } from 'react-router-dom'

export default function LoginPage() {
  const { isAuthenticated, isLoading, login } = useKindeAuth()
  const location = useLocation()
  const errorMessage = (location.state as { error?: string } | null)?.error

  if (!isLoading && isAuthenticated) return <Navigate to="/" replace />

  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="flex flex-col items-center gap-6">
        {errorMessage && (
          <p className="text-sm text-destructive">{errorMessage}</p>
        )}
        <button
          onClick={() => login()}
          className="rounded-md bg-primary px-6 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          Sign in
        </button>
      </div>
    </div>
  )
}
