import { useEffect } from 'react'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { useNavigate, useSearchParams } from 'react-router-dom'

export default function CallbackPage() {
  const { isAuthenticated, isLoading } = useKindeAuth()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()

  useEffect(() => {
    const error = searchParams.get('error')
    if (error) {
      const description = searchParams.get('error_description') ?? 'Sign-in failed. Please try again.'
      navigate('/login', { state: { error: description }, replace: true })
      return
    }

    if (!isLoading && isAuthenticated) {
      navigate('/', { replace: true })
    }
  }, [isAuthenticated, isLoading, navigate, searchParams])

  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <span className="text-sm text-muted-foreground">Signing you in…</span>
    </div>
  )
}
