import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { Navigate, Outlet } from 'react-router-dom'

export default function AuthGuard() {
  const { isAuthenticated, isLoading } = useKindeAuth()

  if (isLoading) return null

  if (!isAuthenticated) return <Navigate to="/login" replace />

  return <Outlet />
}
