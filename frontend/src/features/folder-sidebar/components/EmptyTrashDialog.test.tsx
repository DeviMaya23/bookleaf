import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import EmptyTrashDialog from './EmptyTrashDialog'

vi.mock('@/components/ui/dialog', async () => {
  const React = await import('react')
  return {
    Dialog: ({ open, children }: { open: boolean; onOpenChange?: (v: boolean) => void; children: React.ReactNode }) =>
      open ? React.createElement(React.Fragment, null, children) : null,
    DialogContent: ({ children }: { children: React.ReactNode }) =>
      React.createElement('div', { role: 'dialog' }, children),
    DialogHeader: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
    DialogTitle: ({ children }: { children: React.ReactNode }) =>
      React.createElement('h2', null, children),
    DialogFooter: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
  }
})

describe('EmptyTrashDialog', () => {
  it('renders when open is true', () => {
    render(<EmptyTrashDialog open={true} onCancel={vi.fn()} onConfirm={vi.fn()} />)

    expect(screen.getByRole('heading', { name: 'Empty trash?' })).toBeInTheDocument()
  })

  it('does not render when open is false', () => {
    render(<EmptyTrashDialog open={false} onCancel={vi.fn()} onConfirm={vi.fn()} />)

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('calls onCancel when Cancel is clicked', async () => {
    const onCancel = vi.fn()
    render(<EmptyTrashDialog open={true} onCancel={onCancel} onConfirm={vi.fn()} />)

    await userEvent.click(screen.getByRole('button', { name: 'Cancel' }))

    expect(onCancel).toHaveBeenCalled()
  })

  it('calls onConfirm when Empty trash is clicked', async () => {
    const onConfirm = vi.fn()
    render(<EmptyTrashDialog open={true} onCancel={vi.fn()} onConfirm={onConfirm} />)

    await userEvent.click(screen.getByRole('button', { name: 'Empty trash' }))

    expect(onConfirm).toHaveBeenCalled()
  })
})
