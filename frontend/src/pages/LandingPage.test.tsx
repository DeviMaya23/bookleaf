import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { vi, describe, it, expect } from 'vitest'
import LandingPage from './LandingPage'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: vi.fn(),
}))

import { useKindeAuth } from '@kinde-oss/kinde-auth-react'

function renderLanding(locationState?: { error: string }) {
  const initialEntries = locationState ? [{ pathname: '/', state: locationState }] : ['/']

  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <Routes>
        <Route path="/" element={<LandingPage />} />
        <Route path="/app" element={<div>App</div>} />
      </Routes>
    </MemoryRouter>,
  )
}

describe('LandingPage', () => {
  it('renders nav, hero, features, and footer for an unauthenticated user', () => {
    vi.mocked(useKindeAuth).mockReturnValue({
      isAuthenticated: false,
      isLoading: false,
      login: vi.fn(),
    } as unknown as ReturnType<typeof useKindeAuth>)

    renderLanding()

    expect(screen.getByText('Bookleaf')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /get started/i })).toBeInTheDocument()
    expect(screen.getByRole('img', { name: /bookleaf app screenshot/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'About' })).toHaveAttribute('href', '/about')
    expect(screen.getByRole('link', { name: 'Privacy' })).toHaveAttribute('href', '/privacy')
    expect(screen.getByRole('link', { name: 'AI Notes' })).toHaveAttribute('href', '/ai-notes')
  })

  it('calls login() when "Sign in" is clicked', async () => {
    const login = vi.fn()
    vi.mocked(useKindeAuth).mockReturnValue({
      isAuthenticated: false,
      isLoading: false,
      login,
    } as unknown as ReturnType<typeof useKindeAuth>)

    renderLanding()

    await userEvent.click(screen.getByRole('button', { name: /sign in/i }))

    expect(login).toHaveBeenCalled()
  })

  it('calls login() when "Get started" is clicked', async () => {
    const login = vi.fn()
    vi.mocked(useKindeAuth).mockReturnValue({
      isAuthenticated: false,
      isLoading: false,
      login,
    } as unknown as ReturnType<typeof useKindeAuth>)

    renderLanding()

    await userEvent.click(screen.getByRole('button', { name: /get started/i }))

    expect(login).toHaveBeenCalled()
  })

  it('redirects an already-authenticated user to /app', () => {
    vi.mocked(useKindeAuth).mockReturnValue({
      isAuthenticated: true,
      isLoading: false,
      login: vi.fn(),
    } as unknown as ReturnType<typeof useKindeAuth>)

    renderLanding()

    expect(screen.getByText('App')).toBeInTheDocument()
  })

  it('displays an error message passed via location state', () => {
    vi.mocked(useKindeAuth).mockReturnValue({
      isAuthenticated: false,
      isLoading: false,
      login: vi.fn(),
    } as unknown as ReturnType<typeof useKindeAuth>)

    renderLanding({ error: 'Sign-in failed. Please try again.' })

    expect(screen.getByText('Sign-in failed. Please try again.')).toBeInTheDocument()
  })
})
