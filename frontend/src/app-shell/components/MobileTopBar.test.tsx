import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import MobileTopBar from './MobileTopBar'

describe('MobileTopBar', () => {
  it('renders a hamburger button and the Bookleaf wordmark', () => {
    render(<MobileTopBar onMenuClick={vi.fn()} />)

    expect(screen.getByRole('button', { name: 'Open menu' })).toBeInTheDocument()
    expect(screen.getByText('Bookleaf')).toBeInTheDocument()
  })

  it('calls onMenuClick when the hamburger button is clicked', async () => {
    const onMenuClick = vi.fn()
    render(<MobileTopBar onMenuClick={onMenuClick} />)

    await userEvent.click(screen.getByRole('button', { name: 'Open menu' }))

    expect(onMenuClick).toHaveBeenCalledOnce()
  })
})
