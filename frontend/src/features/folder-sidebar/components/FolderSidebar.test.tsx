import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { getFolderSubtreeIds, getFolders, createFolder, updateFolder } from '@/lib/folders'
import type { Folder } from '@/lib/folders'
import { getMe } from '@/features/auth/lib/me'
import FolderSidebar from './FolderSidebar'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({
    getToken: vi.fn().mockResolvedValue('token'),
    getUserProfile: vi.fn().mockResolvedValue(null),
  }),
}))

vi.mock('@/features/auth/components/ProfileMenu', () => ({
  default: () => null,
}))

vi.mock('@dnd-kit/core', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@dnd-kit/core')>()),
  useDndContext: () => ({ active: null }),
}))

vi.mock('@/lib/folders', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/folders')>()),
  getFolders: vi.fn(),
  createFolder: vi.fn(),
  updateFolder: vi.fn(),
}))

vi.mock('@/lib/images', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/images')>()),
  emptyTrash: vi.fn(),
}))

vi.mock('@/features/auth/lib/me', () => ({
  getMe: vi.fn().mockResolvedValue({ id: 'kp_abc123', vision_enabled: false, folder_icons_enabled: true }),
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
    ContextMenuSub: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
    ContextMenuSubTrigger: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
    ContextMenuSubContent: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
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
    description: null, icon: null,
    parent_id: parentId,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
  }
}

describe('FolderSidebar new folder controls', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('creates a folder via the icon button beside the "Folders" label', async () => {
    vi.mocked(getFolders).mockResolvedValue([makeFolder('1')])
    vi.mocked(createFolder).mockResolvedValue(makeFolder('2'))

    renderSidebar()

    const iconButton = await screen.findByRole('button', { name: 'New folder' })
    await userEvent.click(iconButton)

    const dialogHeading = await screen.findByRole('heading', { name: 'New folder' })
    expect(dialogHeading).toBeInTheDocument()

    await userEvent.type(screen.getByPlaceholderText('Folder name'), 'New Folder Name')
    await userEvent.click(screen.getByRole('button', { name: 'Confirm' }))

    await waitFor(() => {
      expect(createFolder).toHaveBeenCalledWith(expect.any(Function), 'New Folder Name', undefined)
    })
    await waitFor(() => {
      expect(screen.queryByRole('heading', { name: 'New folder' })).not.toBeInTheDocument()
    })
  })

  it('creates a folder via the footer "+ New folder" button when the folder list is empty', async () => {
    vi.mocked(getFolders).mockResolvedValue([])
    vi.mocked(createFolder).mockResolvedValue(makeFolder('1'))

    renderSidebar()

    const footerButton = await screen.findByRole('button', { name: '+ New folder' })
    expect(footerButton).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'New folder' })).toBeInTheDocument()
    await userEvent.click(footerButton)

    const dialogHeading = await screen.findByRole('heading', { name: 'New folder' })
    expect(dialogHeading).toBeInTheDocument()

    await userEvent.type(screen.getByPlaceholderText('Folder name'), 'First Folder')
    await userEvent.click(screen.getByRole('button', { name: 'Confirm' }))

    await waitFor(() => {
      expect(createFolder).toHaveBeenCalledWith(expect.any(Function), 'First Folder', undefined)
    })
    await waitFor(() => {
      expect(screen.queryByRole('heading', { name: 'New folder' })).not.toBeInTheDocument()
    })
  })

  it('hides the footer "+ New folder" button when folders exist', async () => {
    vi.mocked(getFolders).mockResolvedValue([makeFolder('1')])

    renderSidebar()

    await screen.findByText('Folder 1')
    expect(screen.queryByRole('button', { name: '+ New folder' })).not.toBeInTheDocument()
  })
})

describe('FolderSidebar folder icons', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getMe).mockResolvedValue({ id: 'kp_abc123', vision_enabled: false, folder_icons_enabled: true })
  })

  it('renders icons for All, Unsorted, Trash, and user folders when folder_icons_enabled is true', async () => {
    vi.mocked(getFolders).mockResolvedValue([makeFolder('1')])

    renderSidebar()

    const folderRow = (await screen.findByText('Folder 1')).closest('div')
    expect(folderRow?.querySelector('svg')).toBeTruthy()
    expect(screen.getByText('All').closest('div')?.querySelector('svg')).toBeTruthy()
    expect(screen.getByText('Unsorted').closest('div')?.querySelector('svg')).toBeTruthy()
    expect(screen.getByText('Trash').closest('div')?.querySelector('svg')).toBeTruthy()
  })

  it('renders no icons when folder_icons_enabled is false', async () => {
    vi.mocked(getMe).mockResolvedValue({ id: 'kp_abc123', vision_enabled: false, folder_icons_enabled: false })
    vi.mocked(getFolders).mockResolvedValue([makeFolder('1')])

    renderSidebar()

    const folderRow = (await screen.findByText('Folder 1')).closest('div')
    expect(folderRow?.querySelector('svg')).toBeFalsy()
    expect(screen.getByText('All').closest('div')?.querySelector('svg')).toBeFalsy()
    expect(screen.getByText('Unsorted').closest('div')?.querySelector('svg')).toBeFalsy()
    expect(screen.getByText('Trash').closest('div')?.querySelector('svg')).toBeFalsy()
  })

  it('offers a "Change icon" option in a folder\'s context menu but not on Trash', async () => {
    vi.mocked(getFolders).mockResolvedValue([makeFolder('1')])

    renderSidebar()

    await screen.findByText('Folder 1')
    expect(screen.getByText('Change icon')).toBeInTheDocument()
    expect(screen.queryAllByText('Change icon')).toHaveLength(1)
  })

  it('updates a folder\'s icon via the "Change icon" submenu', async () => {
    vi.mocked(getFolders).mockResolvedValue([makeFolder('1')])
    vi.mocked(updateFolder).mockResolvedValue({ ...makeFolder('1'), icon: 'star' })

    renderSidebar()

    await screen.findByText('Folder 1')
    await userEvent.click(screen.getByRole('menuitem', { name: 'star' }))

    await waitFor(() => {
      expect(updateFolder).toHaveBeenCalledWith(expect.any(Function), '1', { icon: 'star' })
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
