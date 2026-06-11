import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import type { Folder } from '@/lib/folders'
import DeleteFolderDialog from './DeleteFolderDialog'

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

function makeFolder(overrides: Partial<Folder> = {}): Folder {
  return {
    id: 'folder-1',
    name: 'Vacation Photos',
    description: null,
    parent_id: null,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides,
  }
}

describe('DeleteFolderDialog', () => {
  it('renders with the folder name when folder is non-null', () => {
    render(<DeleteFolderDialog folder={makeFolder()} onCancel={vi.fn()} onConfirm={vi.fn()} />)

    expect(screen.getByRole('heading', { name: 'Delete folder' })).toBeInTheDocument()
    expect(screen.getByText(/"Vacation Photos"/)).toBeInTheDocument()
  })

  it('stays closed when folder is null', () => {
    render(<DeleteFolderDialog folder={null} onCancel={vi.fn()} onConfirm={vi.fn()} />)

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('calls onCancel when Cancel is clicked', async () => {
    const onCancel = vi.fn()
    render(<DeleteFolderDialog folder={makeFolder()} onCancel={onCancel} onConfirm={vi.fn()} />)

    await userEvent.click(screen.getByRole('button', { name: 'Cancel' }))

    expect(onCancel).toHaveBeenCalled()
  })

  it('calls onConfirm when Delete is clicked', async () => {
    const onConfirm = vi.fn()
    render(<DeleteFolderDialog folder={makeFolder()} onCancel={vi.fn()} onConfirm={onConfirm} />)

    await userEvent.click(screen.getByRole('button', { name: 'Delete' }))

    expect(onConfirm).toHaveBeenCalled()
  })
})
