import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { vi, describe, it, expect } from 'vitest'
import LoginPage from './LoginPage'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: vi.fn(),
}))

import { useKindeAuth } from '@kinde-oss/kinde-auth-react'

function renderLogin(locationState?: { error: string }) {
  const initialEntries = locationState
    ? [{ pathname: '/login', state: locationState }]
    : ['/login']

  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/" element={<div>Home</div>} />
      </Routes>
    </MemoryRouter>,
  )
}

describe('LoginPage', () => {
  it('renders sign-in button for unauthenticated user', () => {
    vi.mocked(useKindeAuth).mockReturnValue({
      isAuthenticated: false,
      isLoading: false,
      login: vi.fn(),
    } as unknown as ReturnType<typeof useKindeAuth>)

    renderLogin()

    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
  })

  it('displays error message when location state contains an error', () => {
    vi.mocked(useKindeAuth).mockReturnValue({
      isAuthenticated: false,
      isLoading: false,
      login: vi.fn(),
    } as unknown as ReturnType<typeof useKindeAuth>)

    renderLogin({ error: 'User denied access' })

    expect(screen.getByText('User denied access')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
  })
})
