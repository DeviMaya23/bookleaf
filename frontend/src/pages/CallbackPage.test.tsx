import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { vi, describe, it, expect } from 'vitest'
import CallbackPage from './CallbackPage'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: vi.fn(),
}))

import { useKindeAuth } from '@kinde-oss/kinde-auth-react'

function renderCallback(search = '') {
  return render(
    <MemoryRouter initialEntries={[`/callback${search}`]}>
      <Routes>
        <Route path="/callback" element={<CallbackPage />} />
        <Route path="/" element={<div>Home</div>} />
        <Route path="/login" element={<div>Login page</div>} />
      </Routes>
    </MemoryRouter>,
  )
}

describe('CallbackPage', () => {
  it('navigates to / after successful authentication', () => {
    vi.mocked(useKindeAuth).mockReturnValue({
      isAuthenticated: true,
      isLoading: false,
    } as ReturnType<typeof useKindeAuth>)

    renderCallback()

    expect(screen.getByText('Home')).toBeInTheDocument()
  })

  it('redirects to /login with error message when Kinde returns an error', () => {
    vi.mocked(useKindeAuth).mockReturnValue({
      isAuthenticated: false,
      isLoading: false,
    } as ReturnType<typeof useKindeAuth>)

    renderCallback('?error=access_denied&error_description=User+denied+access')

    expect(screen.getByText('Login page')).toBeInTheDocument()
  })
})
