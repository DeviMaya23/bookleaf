import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import MobileDrawerShell from './MobileDrawerShell'

describe('MobileDrawerShell', () => {
  it('renders its children', () => {
    render(
      <MobileDrawerShell open={true} onClose={vi.fn()}>
        <p>Drawer content</p>
      </MobileDrawerShell>,
    )

    expect(screen.getByText('Drawer content')).toBeInTheDocument()
  })

  it('calls onClose when the backdrop is tapped', async () => {
    const onClose = vi.fn()
    render(
      <MobileDrawerShell open={true} onClose={onClose}>
        <p>Drawer content</p>
      </MobileDrawerShell>,
    )

    await userEvent.click(screen.getByTestId('mobile-drawer-shell-backdrop'))

    expect(onClose).toHaveBeenCalledTimes(1)
  })
})
