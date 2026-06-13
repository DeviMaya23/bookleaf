import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, it, expect } from 'vitest'
import PrivacyPolicyPage from './PrivacyPolicyPage'

describe('PrivacyPolicyPage', () => {
  it('renders the title and section headings', () => {
    render(
      <MemoryRouter>
        <PrivacyPolicyPage />
      </MemoryRouter>,
    )

    expect(screen.getByRole('heading', { level: 1, name: 'Privacy Policy' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { level: 2, name: 'What we collect' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { level: 2, name: 'How we use it' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { level: 2, name: 'Your choices' })).toBeInTheDocument()
  })

  it('links to the AI Notes page', () => {
    render(
      <MemoryRouter>
        <PrivacyPolicyPage />
      </MemoryRouter>,
    )

    expect(screen.getByRole('link', { name: 'AI Notes' })).toHaveAttribute('href', '/ai-notes')
  })
})
