import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, it, expect } from 'vitest'
import AiNotesPage from './AiNotesPage'

describe('AiNotesPage', () => {
  it('renders the title and section headings', () => {
    render(
      <MemoryRouter>
        <AiNotesPage />
      </MemoryRouter>,
    )

    expect(screen.getByRole('heading', { level: 1, name: 'AI Notes' })).toBeInTheDocument()
    expect(
      screen.getByRole('heading', { level: 2, name: 'How AI is used here' }),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('heading', { level: 2, name: 'What gets sent to Google Vision API' }),
    ).toBeInTheDocument()
    expect(screen.getByRole('heading', { level: 2, name: 'Opting out' })).toBeInTheDocument()
  })

  it('links to the Privacy Policy page', () => {
    render(
      <MemoryRouter>
        <AiNotesPage />
      </MemoryRouter>,
    )

    expect(screen.getByRole('link', { name: 'Privacy Policy' })).toHaveAttribute('href', '/privacy')
  })
})
