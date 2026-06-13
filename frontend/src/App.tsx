import { Routes, Route, Navigate } from 'react-router-dom'
import AuthGuard from '@/features/auth/components/AuthGuard'
import AppLayout from '@/app-shell/AppLayout'
import LandingPage from './pages/LandingPage'
import CallbackPage from './pages/CallbackPage'

function App() {
  return (
    <Routes>
      <Route path="/" element={<LandingPage />} />
      <Route path="/callback" element={<CallbackPage />} />
      <Route path="/app" element={<AuthGuard />}>
        <Route index element={<AppLayout />} />
        <Route path="unsorted" element={<AppLayout />} />
        <Route path="trash" element={<AppLayout />} />
        <Route path="folders/:folderId" element={<AppLayout />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default App
