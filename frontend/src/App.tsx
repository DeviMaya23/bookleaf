import { Routes, Route, Navigate } from 'react-router-dom'
import AuthGuard from '@/features/auth/components/AuthGuard'
import PublicThemeLock from '@/components/PublicThemeLock'
import AppLayout from '@/app-shell/AppLayout'
import LandingPage from './pages/LandingPage'
import CallbackPage from './pages/CallbackPage'
import AboutPage from './pages/AboutPage'
import PrivacyPolicyPage from './pages/PrivacyPolicyPage'
import AiNotesPage from './pages/AiNotesPage'
import SharePage from './features/share-viewer/components/SharePage'

function App() {
  return (
    <Routes>
      <Route element={<PublicThemeLock />}>
        <Route path="/" element={<LandingPage />} />
        <Route path="/about" element={<AboutPage />} />
        <Route path="/privacy" element={<PrivacyPolicyPage />} />
        <Route path="/ai-notes" element={<AiNotesPage />} />
        <Route path="/share/:token" element={<SharePage />} />
      </Route>
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
