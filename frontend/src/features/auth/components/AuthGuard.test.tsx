import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { vi, describe, it, expect } from 'vitest'
import AuthGuard from './AuthGuard'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: vi.fn(),
}))

import { useKindeAuth } from '@kinde-oss/kinde-auth-react'

function renderWithRouter(initialEntry = '/protected') {
  return render(
    <MemoryRouter initialEntries={[initialEntry]}>
      <Routes>
        <Route path="/" element={<div>Landing page</div>} />
        <Route element={<AuthGuard />}>
          <Route path="/protected" element={<div>Protected content</div>} />
        </Route>
      </Routes>
    </MemoryRouter>,
  )
}

describe('AuthGuard', () => {
  it('redirects to / when unauthenticated', () => {
    vi.mocked(useKindeAuth).mockReturnValue({
      isAuthenticated: false,
      isLoading: false,
    } as ReturnType<typeof useKindeAuth>)

    renderWithRouter()

    expect(screen.getByText('Landing page')).toBeInTheDocument()
    expect(screen.queryByText('Protected content')).not.toBeInTheDocument()
  })

  it('renders outlet when authenticated', () => {
    vi.mocked(useKindeAuth).mockReturnValue({
      isAuthenticated: true,
      isLoading: false,
    } as ReturnType<typeof useKindeAuth>)

    renderWithRouter()

    expect(screen.getByText('Protected content')).toBeInTheDocument()
    expect(screen.queryByText('Landing page')).not.toBeInTheDocument()
  })
})
