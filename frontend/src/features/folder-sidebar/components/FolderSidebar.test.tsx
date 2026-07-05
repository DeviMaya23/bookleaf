import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
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
  getMe: vi.fn().mockResolvedValue({ id: 'kp_abc123', vision_enabled: false, folder_icons_enabled: true, ai_categorisation_enabled: false, ai_categorisation_count_this_month: 0 }),
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

function renderSidebar(props: Partial<{ mobileOpen: boolean; onMobileClose: () => void; onFolderViewDetails: () => void }> = {}) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <FolderSidebar view={{ type: 'trash' }} {...props} />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

function mockPointer(isCoarse: boolean) {
  window.matchMedia = (query: string) =>
    ({
      matches: isCoarse,
      media: query,
      onchange: null,
      addEventListener: () => {},
      removeEventListener: () => {},
      addListener: () => {},
      removeListener: () => {},
      dispatchEvent: () => false,
    }) as MediaQueryList
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
    vi.mocked(getMe).mockResolvedValue({ id: 'kp_abc123', vision_enabled: false, folder_icons_enabled: true, ai_categorisation_enabled: false, ai_categorisation_count_this_month: 0 })
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
    vi.mocked(getMe).mockResolvedValue({ id: 'kp_abc123', vision_enabled: false, folder_icons_enabled: false, ai_categorisation_enabled: false, ai_categorisation_count_this_month: 0 })
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

describe('FolderSidebar mobile drawer', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getFolders).mockResolvedValue([])
  })

  it('is off-canvas by default and slides in when mobileOpen is true', async () => {
    const { rerender } = renderSidebar({ mobileOpen: false })
    const aside = await screen.findByRole('complementary')
    const classesWhenClosed = aside.className.split(/\s+/)
    expect(classesWhenClosed).toContain('-translate-x-full')
    expect(classesWhenClosed).not.toContain('translate-x-0')
    expect(classesWhenClosed).toContain('sm:translate-x-0')

    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    rerender(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>
          <FolderSidebar view={{ type: 'trash' }} mobileOpen />
        </MemoryRouter>
      </QueryClientProvider>,
    )
    const classesWhenOpen = aside.className.split(/\s+/)
    expect(classesWhenOpen).toContain('translate-x-0')
    expect(classesWhenOpen).not.toContain('-translate-x-full')
  })

  it('calls onMobileClose when a navigation entry (All) is selected', async () => {
    const onMobileClose = vi.fn()
    renderSidebar({ mobileOpen: true, onMobileClose })

    await userEvent.click(screen.getByText('All'))

    expect(onMobileClose).toHaveBeenCalled()
  })

  it('calls onMobileClose when a folder entry is selected', async () => {
    const onMobileClose = vi.fn()
    vi.mocked(getFolders).mockResolvedValue([makeFolder('1')])
    renderSidebar({ mobileOpen: true, onMobileClose })

    await userEvent.click(await screen.findByText('Folder 1'))

    expect(onMobileClose).toHaveBeenCalled()
  })
})

describe('FolderSidebar — onViewDetails threading', () => {
  const originalMatchMedia = window.matchMedia

  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    window.matchMedia = originalMatchMedia
  })

  it('threads onFolderViewDetails through to a root folder\'s "View details" item', async () => {
    mockPointer(true)
    vi.mocked(getFolders).mockResolvedValue([makeFolder('1')])
    const onFolderViewDetails = vi.fn()

    renderSidebar({ onFolderViewDetails })

    await screen.findByText('Folder 1')
    await userEvent.click(screen.getByRole('menuitem', { name: 'View details' }))

    expect(onFolderViewDetails).toHaveBeenCalled()
  })

  it('threads onFolderViewDetails through to a nested child folder\'s "View details" item', async () => {
    mockPointer(true)
    vi.mocked(getFolders).mockResolvedValue([makeFolder('parent'), makeFolder('child', 'parent')])
    const onFolderViewDetails = vi.fn()

    renderSidebar({ onFolderViewDetails })

    await screen.findByText('Folder child')
    const viewDetailsItems = screen.getAllByRole('menuitem', { name: 'View details' })
    await userEvent.click(viewDetailsItems[1])

    expect(onFolderViewDetails).toHaveBeenCalled()
  })

  it('navigates to a non-active folder when "View details" is selected', async () => {
    mockPointer(true)
    vi.mocked(getFolders).mockResolvedValue([makeFolder('1')])

    render(
      <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
        <MemoryRouter initialEntries={['/app']}>
          <FolderSidebar view={{ type: 'all' }} />
          <Routes>
            <Route path="/app/folders/:id" element={<div data-testid="folder-route" />} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>,
    )

    await screen.findByText('Folder 1')
    await userEvent.click(screen.getByRole('menuitem', { name: 'View details' }))

    expect(await screen.findByTestId('folder-route')).toBeInTheDocument()
  })

  it('does not navigate when "View details" is selected on the already-active folder', async () => {
    mockPointer(true)
    vi.mocked(getFolders).mockResolvedValue([makeFolder('1')])
    const onFolderViewDetails = vi.fn()

    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={['/app/folders/1']}>
          <FolderSidebar view={{ type: 'folder', id: '1' }} onFolderViewDetails={onFolderViewDetails} />
        </MemoryRouter>
      </QueryClientProvider>,
    )

    await screen.findByText('Folder 1')
    await userEvent.click(screen.getByRole('menuitem', { name: 'View details' }))

    expect(onFolderViewDetails).toHaveBeenCalled()
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
