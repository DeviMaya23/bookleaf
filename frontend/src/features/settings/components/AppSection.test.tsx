import { render, screen } from '@testing-library/react'
import { describe, it, expect, beforeEach } from 'vitest'
import { ThemeProvider } from '@/hooks/useTheme'
import AppSection from './AppSection'

describe('AppSection', () => {
  beforeEach(() => {
    localStorage.clear()
    document.documentElement.removeAttribute('data-theme')
  })

  it('shows the Default theme as selected, reflecting useTheme()', () => {
    render(<AppSection />, { wrapper: ThemeProvider })

    expect(screen.getByText('Default')).toBeInTheDocument()
    expect(screen.getByText('Warm parchment')).toBeInTheDocument()
    expect(screen.getByRole('radio', { name: 'Default theme (selected)' })).toBeChecked()
  })
})
