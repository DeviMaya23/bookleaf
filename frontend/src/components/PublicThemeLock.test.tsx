import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { describe, it, expect } from 'vitest'
import PublicThemeLock from './PublicThemeLock'

describe('PublicThemeLock', () => {
  it('renders outlet content inside a data-theme="warm" wrapper', () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <Routes>
          <Route element={<PublicThemeLock />}>
            <Route path="/" element={<div>Public content</div>} />
          </Route>
        </Routes>
      </MemoryRouter>,
    )

    const content = screen.getByText('Public content')
    expect(content).toBeInTheDocument()
    expect(content.closest('[data-theme="warm"]')).not.toBeNull()
  })
})
