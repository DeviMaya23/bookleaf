import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import MobileDrawerShell from './MobileDrawerShell'

describe('MobileDrawerShell', () => {
  it('renders its children', () => {
    render(
      <MobileDrawerShell onClose={vi.fn()}>
        <p>Drawer content</p>
      </MobileDrawerShell>,
    )

    expect(screen.getByText('Drawer content')).toBeInTheDocument()
  })

  it('calls onClose when the backdrop is tapped', () => {
    const onClose = vi.fn()
    render(
      <MobileDrawerShell onClose={onClose}>
        <p>Drawer content</p>
      </MobileDrawerShell>,
    )

    fireEvent.click(screen.getByTestId('mobile-drawer-shell-backdrop'))

    expect(onClose).toHaveBeenCalledTimes(1)
  })
})
