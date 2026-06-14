import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, beforeEach } from 'vitest'
import { ThemeProvider } from '@/hooks/useTheme'
import AppSection from './AppSection'

describe('AppSection', () => {
  beforeEach(() => {
    localStorage.clear()
    document.documentElement.removeAttribute('data-theme')
  })

  it('shows the Parchment theme as selected, reflecting useTheme()', () => {
    render(<AppSection />, { wrapper: ThemeProvider })

    expect(screen.getByText('Parchment')).toBeInTheDocument()
    expect(screen.getByText('Warm parchment')).toBeInTheDocument()
    expect(screen.getByRole('radio', { name: 'Parchment theme' })).toBeChecked()
  })

  it('renders all three theme rows with the correct selected state', () => {
    render(<AppSection />, { wrapper: ThemeProvider })

    expect(screen.getByRole('radio', { name: 'Parchment theme' })).toBeChecked()
    expect(screen.getByRole('radio', { name: 'Lumen theme' })).not.toBeChecked()
    expect(screen.getByRole('radio', { name: 'Sunless theme' })).not.toBeChecked()
  })

  it('calls setTheme when an unselected row is clicked', async () => {
    const user = userEvent.setup()
    render(<AppSection />, { wrapper: ThemeProvider })

    await user.click(screen.getByRole('radio', { name: 'Lumen theme' }))

    expect(screen.getByRole('radio', { name: 'Lumen theme' })).toBeChecked()
    expect(screen.getByRole('radio', { name: 'Parchment theme' })).not.toBeChecked()
  })
})
