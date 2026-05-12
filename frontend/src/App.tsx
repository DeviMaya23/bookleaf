import { Routes, Route, Navigate, Outlet } from 'react-router-dom'
import AuthGuard from './components/AuthGuard'
import LogoutButton from './components/LogoutButton'
import LoginPage from './pages/LoginPage'
import CallbackPage from './pages/CallbackPage'

function AuthenticatedLayout() {
  return (
    <div className="min-h-screen bg-background">
      <div className="flex justify-end p-4">
        <LogoutButton />
      </div>
      <Outlet />
    </div>
  )
}

function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/callback" element={<CallbackPage />} />
      <Route element={<AuthGuard />}>
        <Route element={<AuthenticatedLayout />}>
          <Route path="/" element={null} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Route>
    </Routes>
  )
}

export default App
