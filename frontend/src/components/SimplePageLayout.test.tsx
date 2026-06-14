import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, it, expect } from 'vitest'
import SimplePageLayout from './SimplePageLayout'

describe('SimplePageLayout', () => {
  it('renders the nav wordmark linking to /, title, and children', () => {
    render(
      <MemoryRouter>
        <SimplePageLayout title="Example Page">
          <p>Example content</p>
        </SimplePageLayout>
      </MemoryRouter>,
    )

    expect(screen.getByRole('link', { name: 'Bookleaf' })).toHaveAttribute('href', '/')
    expect(screen.getByRole('heading', { level: 1, name: 'Example Page' })).toBeInTheDocument()
    expect(screen.getByText('Example content')).toBeInTheDocument()
  })
})
