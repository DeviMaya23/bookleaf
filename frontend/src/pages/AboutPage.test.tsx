import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, it, expect } from 'vitest'
import AboutPage from './AboutPage'

describe('AboutPage', () => {
  it('renders the title and section headings', () => {
    render(
      <MemoryRouter>
        <AboutPage />
      </MemoryRouter>,
    )

    expect(screen.getByRole('heading', { level: 1, name: 'About Bookleaf' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { level: 2, name: 'What Bookleaf is' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { level: 2, name: 'Why it exists' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { level: 2, name: "Who's behind it" })).toBeInTheDocument()
  })
})
