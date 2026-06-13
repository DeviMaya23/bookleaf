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
        <Route path="/app" element={<div>App</div>} />
        <Route path="/" element={<div>Landing page</div>} />
      </Routes>
    </MemoryRouter>,
  )
}

describe('CallbackPage', () => {
  it('navigates to /app after successful authentication', () => {
    vi.mocked(useKindeAuth).mockReturnValue({
      isAuthenticated: true,
      isLoading: false,
    } as ReturnType<typeof useKindeAuth>)

    renderCallback()

    expect(screen.getByText('App')).toBeInTheDocument()
  })

  it('redirects to / with error message when Kinde returns an error', () => {
    vi.mocked(useKindeAuth).mockReturnValue({
      isAuthenticated: false,
      isLoading: false,
    } as ReturnType<typeof useKindeAuth>)

    renderCallback('?error=access_denied&error_description=User+denied+access')

    expect(screen.getByText('Landing page')).toBeInTheDocument()
  })
})
