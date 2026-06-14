import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import SettingsModal from './SettingsModal'

vi.mock('@/components/ui/dialog', async () => {
  const React = await import('react')
  const DialogContext = React.createContext<(open: boolean) => void>(() => {})
  return {
    Dialog: ({
      open,
      onOpenChange,
      children,
    }: {
      open: boolean
      onOpenChange: (open: boolean) => void
      children: React.ReactNode
    }) =>
      open
        ? React.createElement(DialogContext.Provider, { value: onOpenChange }, children)
        : null,
    DialogContent: ({ children, className }: { children: React.ReactNode; className?: string }) =>
      React.createElement('div', { role: 'dialog', className }, children),
    DialogTitle: ({ children, className }: { children: React.ReactNode; className?: string }) =>
      React.createElement('h2', { className }, children),
    DialogClose: ({
      children,
      render,
    }: {
      children: React.ReactNode
      render: React.ReactElement<{ onClick?: () => void }>
    }) => {
      const onOpenChange = React.useContext(DialogContext)
      return React.cloneElement(render, { onClick: () => onOpenChange(false) }, children)
    },
  }
})

vi.mock('./AccountSection', () => ({
  default: () => <div>Account content</div>,
}))
vi.mock('./AppSection', () => ({
  default: () => <div>App content</div>,
}))
vi.mock('./AdvancedSection', () => ({
  default: () => <div>Advanced content</div>,
}))

describe('SettingsModal', () => {
  it('defaults to the Account section, highlighted as active', () => {
    render(<SettingsModal open onOpenChange={vi.fn()} />)

    expect(screen.getByText('Account content')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Account' })).toHaveAttribute('aria-current', 'true')
    expect(screen.getByRole('button', { name: 'App' })).not.toHaveAttribute('aria-current')
  })

  it('switches sections and updates the active nav highlight', async () => {
    render(<SettingsModal open onOpenChange={vi.fn()} />)

    await userEvent.click(screen.getByRole('button', { name: 'App' }))

    expect(screen.getByText('App content')).toBeInTheDocument()
    expect(screen.queryByText('Account content')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'App' })).toHaveAttribute('aria-current', 'true')

    await userEvent.click(screen.getByRole('button', { name: 'Advanced' }))

    expect(screen.getByText('Advanced content')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Advanced' })).toHaveAttribute('aria-current', 'true')
  })

  it('closes the modal via the close control', async () => {
    const onOpenChange = vi.fn()
    render(<SettingsModal open onOpenChange={onOpenChange} />)

    await userEvent.click(screen.getByRole('button', { name: 'Close' }))

    expect(onOpenChange).toHaveBeenCalledWith(false)
  })
})
