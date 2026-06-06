import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { getFolderSubtreeIds, getFolders } from '@/lib/folders'
import { emptyTrash } from '@/lib/images'
import type { Folder } from '@/lib/folders'
import FolderSidebar from './FolderSidebar'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({
    getToken: vi.fn().mockResolvedValue('token'),
    getUserProfile: vi.fn().mockResolvedValue(null),
  }),
}))

vi.mock('./ProfileMenu', () => ({
  default: () => null,
}))

vi.mock('@dnd-kit/core', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@dnd-kit/core')>()),
  useDndContext: () => ({ active: null }),
}))

vi.mock('@/lib/folders', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/folders')>()),
  getFolders: vi.fn(),
}))

vi.mock('@/lib/images', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/images')>()),
  emptyTrash: vi.fn(),
}))

vi.mock('@/components/ui/context-menu', async () => {
  const React = await import('react')
  return {
    ContextMenu: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
    ContextMenuTrigger: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
    ContextMenuContent: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
    ContextMenuSeparator: () => React.createElement('hr'),
    ContextMenuItem: ({
      children,
      onClick,
      className,
    }: {
      children: React.ReactNode
      onClick?: (e: React.MouseEvent) => void
      className?: string
    }) =>
      React.createElement('button', { role: 'menuitem', className, onClick }, children),
  }
})

function renderSidebar() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <FolderSidebar view={{ type: 'trash' }} />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

function makeFolder(id: string, parentId: string | null = null): Folder {
  return {
    id,
    name: `Folder ${id}`,
    description: null,
    parent_id: parentId,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
  }
}

describe('FolderSidebar trash context menu', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getFolders).mockResolvedValue([])
  })

  it('shows Empty trash item when right-clicking the Trash entry', async () => {
    renderSidebar()

    await waitFor(() => {
      expect(screen.getByRole('menuitem', { name: /empty trash/i })).toBeInTheDocument()
    })
  })

  it('calls emptyTrash after confirming the dialog', async () => {
    vi.mocked(emptyTrash).mockResolvedValue(undefined)

    renderSidebar()

    await waitFor(() => {
      expect(screen.getByRole('menuitem', { name: /empty trash/i })).toBeInTheDocument()
    })

    await userEvent.click(screen.getByRole('menuitem', { name: /empty trash/i }))

    const confirmButton = await screen.findByRole('button', { name: /empty trash/i })
    await userEvent.click(confirmButton)

    await waitFor(() => {
      expect(emptyTrash).toHaveBeenCalledWith(expect.any(Function))
    })
  })
})

describe('getFolderSubtreeIds', () => {
  it('returns subtree including self and all descendants', () => {
    const folders = [
      makeFolder('root'),
      makeFolder('child-a', 'root'),
      makeFolder('child-b', 'root'),
      makeFolder('grandchild', 'child-a'),
      makeFolder('other'),
    ]

    const result = getFolderSubtreeIds(folders, 'root')

    expect(result.has('root')).toBe(true)
    expect(result.has('child-a')).toBe(true)
    expect(result.has('child-b')).toBe(true)
    expect(result.has('grandchild')).toBe(true)
    expect(result.has('other')).toBe(false)
  })

  it('returns only self when folder has no descendants', () => {
    const folders = [
      makeFolder('parent'),
      makeFolder('leaf', 'parent'),
    ]

    const result = getFolderSubtreeIds(folders, 'leaf')

    expect(result.has('leaf')).toBe(true)
    expect(result.has('parent')).toBe(false)
    expect(result.size).toBe(1)
  })

  it('blocks self-drop: dragged folder ID is in its own subtree', () => {
    const folders = [makeFolder('folder-a')]
    const result = getFolderSubtreeIds(folders, 'folder-a')
    expect(result.has('folder-a')).toBe(true)
  })

  it('blocks descendant-drop: descendant ID is in subtree', () => {
    const folders = [
      makeFolder('a'),
      makeFolder('b', 'a'),
      makeFolder('c', 'b'),
    ]

    const result = getFolderSubtreeIds(folders, 'a')

    expect(result.has('b')).toBe(true)
    expect(result.has('c')).toBe(true)
  })
})
