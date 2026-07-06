import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Routes, Route, Link } from 'react-router-dom'
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest'
import AppLayout from './AppLayout'
import { getFolders } from '@/lib/folders'
import { getTags } from '@/lib/tags'
import { setMaintenanceActive } from '@/lib/maintenanceStore'
import { bulkAddImagesToFolder, bulkTrashImages } from '@/lib/images'
import type { Image, BulkActionResult } from '@/lib/images'
import { toast } from 'sonner'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({ getToken: vi.fn().mockResolvedValue('token') }),
}))

vi.mock('@/features/auth/lib/me', () => ({
  getMe: vi.fn().mockResolvedValue({
    id: 'kp_test',
    vision_enabled: false,
    folder_icons_enabled: false,
    ai_categorisation_enabled: false,
    ai_categorisation_count_this_month: 0,
  }),
  updateMe: vi.fn(),
}))

vi.mock('@/lib/folders', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/folders')>()),
  getFolders: vi.fn().mockResolvedValue([]),
}))

vi.mock('@/lib/tags', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/tags')>()),
  getTags: vi.fn().mockResolvedValue([]),
}))

vi.mock('@/lib/images', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/images')>()),
  bulkAddImagesToFolder: vi.fn(),
  bulkTrashImages: vi.fn(),
}))

vi.mock('sonner', async (importOriginal) => ({
  ...(await importOriginal<typeof import('sonner')>()),
  toast: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
}))

vi.mock('@/components/ui/dropdown-menu', async () => {
  const React = await import('react')
  const RadioGroupContext = React.createContext<{ value?: string; onValueChange?: (value: string) => void }>({})

  return {
    DropdownMenu: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
    DropdownMenuTrigger: ({ children, className, ...props }: { children: React.ReactNode; className?: string }) =>
      React.createElement('button', { className, ...props }, children),
    DropdownMenuContent: ({ children }: { children: React.ReactNode }) =>
      React.createElement('div', { 'data-testid': 'dropdown-content' }, children),
    DropdownMenuGroup: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
    DropdownMenuItem: ({ children, onClick }: { children: React.ReactNode; onClick?: () => void }) =>
      React.createElement('button', { role: 'menuitem', onClick }, children),
    DropdownMenuSeparator: () => React.createElement('hr'),
    DropdownMenuRadioGroup: ({ children, value, onValueChange }: { children: React.ReactNode; value?: string; onValueChange?: (value: string) => void }) =>
      React.createElement(RadioGroupContext.Provider, { value: { value, onValueChange } }, children),
    DropdownMenuRadioItem: ({ children, value }: { children: React.ReactNode; value: string }) => {
      const ctx = React.useContext(RadioGroupContext)
      return React.createElement('button', {
        role: 'menuitemradio',
        'aria-checked': ctx.value === value,
        onClick: () => ctx.onValueChange?.(value),
      }, children)
    },
    DropdownMenuLabel: ({ children }: { children: React.ReactNode }) =>
      React.createElement('div', null, children),
    DropdownMenuCheckboxItem: ({ children, checked, onCheckedChange }: { children: React.ReactNode; checked?: boolean; onCheckedChange?: (checked: boolean) => void }) =>
      React.createElement('button', {
        role: 'menuitemcheckbox',
        'aria-checked': !!checked,
        onClick: () => onCheckedChange?.(!checked),
      }, children),
  }
})

vi.mock('@/components/ui/scroll-area', async () => {
  const React = await import('react')
  return {
    ScrollArea: ({ children, className }: { children: React.ReactNode; className?: string }) =>
      React.createElement('div', { className }, children),
  }
})

vi.mock('@/features/folder-sidebar/components/FolderSidebar', () => ({
  default: ({ onFolderSelect, onFolderViewDetails }: { onFolderSelect?: () => void; onFolderViewDetails?: () => void }) => (
    <div data-testid="folder-sidebar">
      <button onClick={onFolderSelect}>Select folder</button>
      <button onClick={() => onFolderViewDetails?.()}>View folder details</button>
    </div>
  ),
}))

vi.mock('@/features/viewer/components/ImageViewer', () => ({
  default: ({ focusMode, onClose }: { focusMode: boolean; onClose: () => void }) => (
    <div data-testid="image-viewer" data-focus-mode={String(focusMode)}>
      <button onClick={onClose}>Close viewer</button>
    </div>
  ),
}))

vi.mock('@/features/upload/components/UploadModal', () => ({
  default: () => null,
}))

vi.mock('@/features/upload/components/BatchUploadModal', () => ({
  default: () => null,
}))

vi.mock('@/features/right-panel/components/RightPanel', () => ({
  default: (
    props:
      | { mode: 'image'; image: Image }
      | { mode: 'folder'; folder: { name: string } }
      | { mode: 'selection'; selectedCount: number; onAddToFolder: (folderId: string) => void; onMoveToTrash: () => void; onClose: () => void },
  ) => (
    <div data-testid="right-panel" data-mode={props.mode}>
      {props.mode === 'image' ? (
        props.image.title
      ) : props.mode === 'folder' ? (
        props.folder.name
      ) : (
        <>
          {props.selectedCount} selected
          <button onClick={() => props.onAddToFolder('folder-1')}>Add selection to folder</button>
          <button onClick={() => props.onMoveToTrash()}>Move selection to trash</button>
          <button onClick={() => props.onClose()}>Close selection panel</button>
        </>
      )}
    </div>
  ),
}))

function makeTestImage(overrides?: Partial<Image>): Image {
  return {
    id: 'img-1',
    title: 'Sunset photo',
    description: null,
    mime_type: 'image/jpeg',
    source_url: null,
    folder_ids: [],
    thumbnail_url: 'https://example.com/thumb.jpg',
    width: 1920,
    height: 1080,
    file_size: null,
    tags: [],
    position: null,
    created_at: '2026-05-01T00:00:00Z',
    updated_at: '2026-05-01T00:00:00Z',
    ...overrides,
  }
}

vi.mock('@/features/gallery/components/ImageGrid', () => ({
  default: ({ sortBy, sortDir, filterTagIds, filterMimeTypes, filterFolderIds, onImageSelect, onImageDoubleClick, onImageDeleted, selectMode, onSelectionChange }: { sortBy: string; sortDir: string | undefined; filterTagIds?: string[]; filterMimeTypes?: string[]; filterFolderIds?: string[]; onImageSelect: (img: Image) => void; onImageDoubleClick: (img: Image) => void; onImageDeleted: (id: string) => void; selectMode?: boolean; onSelectionChange?: (ids: Set<string>, anchorId: string | null) => void }) => (
    <div
      data-testid="image-grid"
      data-sort-by={sortBy}
      data-sort-dir={sortDir ?? ''}
      data-filter-tag-ids={(filterTagIds ?? []).join(',')}
      data-filter-mime-types={(filterMimeTypes ?? []).join(',')}
      data-filter-folder-ids={(filterFolderIds ?? []).join(',')}
      data-select-mode={String(!!selectMode)}
    >
      <button onClick={() => onImageSelect(makeTestImage())}>Select image</button>
      <button onDoubleClick={() => onImageDoubleClick(makeTestImage())}>Open image</button>
      <button onClick={() => onImageDeleted('img-1')}>Delete image</button>
      <button onClick={() => onSelectionChange?.(new Set(['img-1']), 'img-1')}>Select img-1 in select mode</button>
    </div>
  ),
}))

function renderApp(initialPath: string) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[initialPath]}>
        <nav>
          <Link to="/app">All</Link>
          <Link to="/app/folders/folder-1">Folder 1</Link>
          <Link to="/app/folders/folder-2">Folder 2</Link>
        </nav>
        <Routes>
          <Route path="/app" element={<AppLayout />} />
          <Route path="/app/unsorted" element={<AppLayout />} />
          <Route path="/app/trash" element={<AppLayout />} />
          <Route path="/app/folders/:folderId" element={<AppLayout />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

function imageGrid() {
  return screen.getByTestId('image-grid')
}

function makeFolder(id: string, name: string) {
  return { id, name, description: null, icon: null, parent_id: null, created_at: '', updated_at: '' }
}

describe('AppLayout focus mode', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getFolders).mockResolvedValue([])
    vi.mocked(getTags).mockResolvedValue([])
    setMaintenanceActive(false)
  })

  function focusToggle() {
    return screen.getByRole('button', { name: /focus mode/i })
  }

  it('enabling focus mode hides FolderSidebar and removes the ml-[240px] class from main', async () => {
    renderApp('/app')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    expect(screen.getByTestId('folder-sidebar')).toBeInTheDocument()
    expect(document.querySelector('main')!.className).toMatch(/ml-\[240px\]/)

    await userEvent.click(focusToggle())

    expect(screen.queryByTestId('folder-sidebar')).not.toBeInTheDocument()
    expect(document.querySelector('main')!.className).not.toMatch(/ml-\[240px\]/)
  })

  it('enabling focus mode hides an open RightPanel in image mode', async () => {
    renderApp('/app')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.click(screen.getByRole('button', { name: 'Select image' }))
    expect(screen.getByTestId('right-panel')).toHaveAttribute('data-mode', 'image')

    await userEvent.click(focusToggle())

    expect(screen.queryByTestId('right-panel')).not.toBeInTheDocument()
  })

  it('enabling focus mode hides an open RightPanel in folder mode', async () => {
    vi.mocked(getFolders).mockResolvedValue([makeFolder('folder-1', 'Vacation')])
    renderApp('/app/folders/folder-1')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.click(screen.getByRole('button', { name: 'Select folder' }))
    await waitFor(() => expect(screen.getByTestId('right-panel')).toHaveAttribute('data-mode', 'folder'))

    await userEvent.click(focusToggle())

    expect(screen.queryByTestId('right-panel')).not.toBeInTheDocument()
  })

  it('clicking an image card while focus mode is active updates selection without rendering RightPanel, revealed once focus mode is disabled', async () => {
    renderApp('/app')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.click(focusToggle())
    expect(screen.queryByTestId('folder-sidebar')).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Select image' }))
    expect(screen.queryByTestId('right-panel')).not.toBeInTheDocument()

    await userEvent.click(focusToggle())
    expect(screen.getByTestId('right-panel')).toHaveAttribute('data-mode', 'image')
  })

  it('double-clicking an image card while focus mode is active opens the viewer at full width with no RightPanel', async () => {
    renderApp('/app')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.click(focusToggle())
    await userEvent.dblClick(screen.getByRole('button', { name: 'Open image' }))

    expect(screen.getByTestId('image-viewer')).toHaveAttribute('data-focus-mode', 'true')
    expect(screen.queryByTestId('right-panel')).not.toBeInTheDocument()
    expect(document.querySelector('main')!.className).not.toMatch(/ml-\[240px\]/)
  })
})

describe('AppLayout mobile responsiveness', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getFolders).mockResolvedValue([])
    vi.mocked(getTags).mockResolvedValue([])
    setMaintenanceActive(false)
  })

  it('wraps the focus toggle in a hidden sm:flex container', async () => {
    renderApp('/app')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    const toggle = screen.getByRole('button', { name: /focus mode/i })
    expect(toggle.parentElement?.className).toMatch(/hidden sm:flex/)
  })

  it('gives main the responsive margin classes instead of an unconditional ml-[240px]', async () => {
    renderApp('/app')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    const main = document.querySelector('main')!
    expect(main.className).toMatch(/ml-0 sm:ml-\[240px\]/)
  })

  it('opens the drawer backdrop via the mobile top bar hamburger and closes it on backdrop tap', async () => {
    renderApp('/app')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    expect(screen.queryByTestId('mobile-drawer-backdrop')).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Open menu' }))
    expect(screen.getByTestId('mobile-drawer-backdrop')).toBeInTheDocument()

    await userEvent.click(screen.getByTestId('mobile-drawer-backdrop'))
    expect(screen.queryByTestId('mobile-drawer-backdrop')).not.toBeInTheDocument()
  })

  it('does not render the mobile top bar while focus mode is active', async () => {
    renderApp('/app')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    expect(screen.getByRole('button', { name: 'Open menu' })).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: /focus mode/i }))

    expect(screen.queryByRole('button', { name: 'Open menu' })).not.toBeInTheDocument()
  })
})

describe('AppLayout cross-feature integration', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getFolders).mockResolvedValue([])
    vi.mocked(getTags).mockResolvedValue([])
    setMaintenanceActive(false)
  })

  it('double-clicking an image opens the viewer and keeps it selected so the right panel shows once the viewer closes', async () => {
    renderApp('/app')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.dblClick(screen.getByRole('button', { name: 'Open image' }))
    expect(screen.getByTestId('image-viewer')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Close viewer' }))

    expect(screen.queryByTestId('image-viewer')).not.toBeInTheDocument()
    expect(screen.getByTestId('right-panel')).toHaveAttribute('data-mode', 'image')
  })

  it('deleting the active image clears its right-panel selection', async () => {
    renderApp('/app')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.click(screen.getByRole('button', { name: 'Select image' }))
    expect(screen.getByTestId('right-panel')).toHaveAttribute('data-mode', 'image')

    await userEvent.click(screen.getByRole('button', { name: 'Delete image' }))

    expect(screen.queryByTestId('right-panel')).not.toBeInTheDocument()
  })
})

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

describe('AppLayout folder selection — pointer-capability gating', () => {
  const originalMatchMedia = window.matchMedia

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getFolders).mockResolvedValue([makeFolder('folder-1', 'Vacation')])
    vi.mocked(getTags).mockResolvedValue([])
    setMaintenanceActive(false)
  })

  afterEach(() => {
    window.matchMedia = originalMatchMedia
  })

  it('selecting a folder still opens the right panel on a fine-pointer device', async () => {
    mockPointer(false)
    renderApp('/app/folders/folder-1')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.click(screen.getByRole('button', { name: 'Select folder' }))

    await waitFor(() => expect(screen.getByTestId('right-panel')).toHaveAttribute('data-mode', 'folder'))
  })

  it('selecting a folder does not open the right panel on a coarse-pointer device', async () => {
    mockPointer(true)
    renderApp('/app/folders/folder-1')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.click(screen.getByRole('button', { name: 'Select folder' }))

    expect(screen.queryByTestId('right-panel')).not.toBeInTheDocument()
  })

  it('"View folder details" opens the right panel on a coarse-pointer device', async () => {
    mockPointer(true)
    renderApp('/app/folders/folder-1')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.click(screen.getByRole('button', { name: 'View folder details' }))

    await waitFor(() => expect(screen.getByTestId('right-panel')).toHaveAttribute('data-mode', 'folder'))
  })

  it('an already-open panel updates to the newly selected folder on a coarse-pointer device without closing', async () => {
    mockPointer(true)
    vi.mocked(getFolders).mockResolvedValue([makeFolder('folder-1', 'Vacation'), makeFolder('folder-2', 'Work')])
    renderApp('/app/folders/folder-1')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.click(screen.getByRole('button', { name: 'View folder details' }))
    await waitFor(() => expect(screen.getByTestId('right-panel')).toHaveTextContent('Vacation'))

    await userEvent.click(screen.getByRole('link', { name: 'Folder 2' }))

    await waitFor(() => expect(screen.getByTestId('right-panel')).toHaveTextContent('Work'))
    expect(screen.getByTestId('right-panel')).toBeInTheDocument()
  })
})

describe('AppLayout maintenance mode', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getFolders).mockResolvedValue([])
    vi.mocked(getTags).mockResolvedValue([])
    setMaintenanceActive(false)
  })

  it('renders MaintenancePage when maintenance is active', async () => {
    setMaintenanceActive(true)

    renderApp('/app')

    expect(await screen.findByTestId('maintenance-page')).toBeInTheDocument()
    expect(screen.queryByTestId('image-grid')).not.toBeInTheDocument()
  })

  it('renders the normal shell when maintenance is inactive', async () => {
    renderApp('/app')

    await waitFor(() => expect(imageGrid()).toBeInTheDocument())
    expect(screen.queryByTestId('maintenance-page')).not.toBeInTheDocument()
  })

  it('switches back to the normal shell when maintenance ends', async () => {
    setMaintenanceActive(true)
    renderApp('/app')
    expect(await screen.findByTestId('maintenance-page')).toBeInTheDocument()

    setMaintenanceActive(false)

    await waitFor(() => expect(imageGrid()).toBeInTheDocument())
    expect(screen.queryByTestId('maintenance-page')).not.toBeInTheDocument()
  })
})

function selectViaGrid() {
  return screen.getByRole('button', { name: 'Select img-1 in select mode' })
}

describe('AppLayout selection mode — right panel priority and visibility', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getFolders).mockResolvedValue([])
    vi.mocked(getTags).mockResolvedValue([])
    setMaintenanceActive(false)
  })

  it('shows the selection panel once an image is selected via onSelectionChange', async () => {
    renderApp('/app')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.click(selectViaGrid())

    expect(screen.getByTestId('right-panel')).toHaveAttribute('data-mode', 'selection')
    expect(screen.getByText('1 selected')).toBeInTheDocument()
  })

  it('takes priority over the image panel when a selectedImage is also set', async () => {
    renderApp('/app')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.click(screen.getByRole('button', { name: 'Select image' }))
    expect(screen.getByTestId('right-panel')).toHaveAttribute('data-mode', 'image')

    await userEvent.click(selectViaGrid())

    expect(screen.getByTestId('right-panel')).toHaveAttribute('data-mode', 'selection')
  })

  it('remains visible while focus mode is active', async () => {
    renderApp('/app')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.click(selectViaGrid())
    expect(screen.getByTestId('right-panel')).toHaveAttribute('data-mode', 'selection')

    await userEvent.click(screen.getByRole('button', { name: /focus mode/i }))

    expect(screen.getByTestId('right-panel')).toHaveAttribute('data-mode', 'selection')
  })

  it('clears the selection when navigating to a different view', async () => {
    renderApp('/app')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.click(selectViaGrid())
    expect(screen.getByTestId('right-panel')).toHaveAttribute('data-mode', 'selection')

    await userEvent.click(screen.getByRole('link', { name: 'Folder 1' }))

    await waitFor(() => expect(screen.queryByTestId('right-panel')).not.toBeInTheDocument())
  })

  it('entering select mode closes an open image panel, and it does not resurface once the selection panel is closed', async () => {
    renderApp('/app')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.click(screen.getByRole('button', { name: 'Select image' }))
    expect(screen.getByTestId('right-panel')).toHaveAttribute('data-mode', 'image')

    await userEvent.click(screen.getByRole('button', { name: 'Select mode' }))
    expect(screen.queryByTestId('right-panel')).not.toBeInTheDocument()

    await userEvent.click(selectViaGrid())
    expect(screen.getByTestId('right-panel')).toHaveAttribute('data-mode', 'selection')

    await userEvent.click(screen.getByRole('button', { name: 'Close selection panel' }))

    expect(screen.queryByTestId('right-panel')).not.toBeInTheDocument()
  })

  it('entering select mode closes an open folder panel', async () => {
    vi.mocked(getFolders).mockResolvedValue([makeFolder('folder-1', 'Vacation')])
    renderApp('/app/folders/folder-1')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.click(screen.getByRole('button', { name: 'Select folder' }))
    await waitFor(() => expect(screen.getByTestId('right-panel')).toHaveAttribute('data-mode', 'folder'))

    await userEvent.click(screen.getByRole('button', { name: 'Select mode' }))

    expect(screen.queryByTestId('right-panel')).not.toBeInTheDocument()
  })

  it('closing the selection panel turns select mode off entirely', async () => {
    renderApp('/app')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.click(screen.getByRole('button', { name: 'Select mode' }))
    await userEvent.click(selectViaGrid())
    expect(screen.getByTestId('right-panel')).toHaveAttribute('data-mode', 'selection')

    await userEvent.click(screen.getByRole('button', { name: 'Close selection panel' }))

    expect(screen.queryByTestId('right-panel')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Select mode' })).toHaveAttribute('aria-pressed', 'false')
  })

  it('turns select mode off entirely when navigating to a different view, not just clearing the selection', async () => {
    renderApp('/app')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.click(screen.getByRole('button', { name: 'Select mode' }))
    expect(screen.getByRole('button', { name: 'Select mode' })).toHaveAttribute('aria-pressed', 'true')

    await userEvent.click(screen.getByRole('link', { name: 'Folder 1' }))

    await waitFor(() => expect(screen.getByRole('button', { name: 'Select mode' })).toHaveAttribute('aria-pressed', 'false'))
  })

  it('does not render the select-mode toggle while viewing the trash', async () => {
    renderApp('/app/trash')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    expect(screen.queryByRole('button', { name: 'Select mode' })).not.toBeInTheDocument()
  })
})

describe('AppLayout bulk selection actions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getFolders).mockResolvedValue([])
    vi.mocked(getTags).mockResolvedValue([])
    setMaintenanceActive(false)
  })

  it('exits select mode entirely after a successful add-to-folder', async () => {
    vi.mocked(bulkAddImagesToFolder).mockResolvedValue({ succeeded_count: 1 } as BulkActionResult)
    renderApp('/app')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.click(screen.getByRole('button', { name: 'Select mode' }))
    await userEvent.click(selectViaGrid())
    await userEvent.click(screen.getByRole('button', { name: 'Add selection to folder' }))

    expect(bulkAddImagesToFolder).toHaveBeenCalledWith(expect.any(Function), ['img-1'], 'folder-1')
    await waitFor(() => expect(screen.queryByTestId('right-panel')).not.toBeInTheDocument())
    expect(imageGrid()).toHaveAttribute('data-select-mode', 'false')
  })

  it('exits select mode entirely after a successful move-to-trash', async () => {
    vi.mocked(bulkTrashImages).mockResolvedValue({ succeeded_count: 1 } as BulkActionResult)
    renderApp('/app')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.click(screen.getByRole('button', { name: 'Select mode' }))
    await userEvent.click(selectViaGrid())
    await userEvent.click(screen.getByRole('button', { name: 'Move selection to trash' }))

    expect(bulkTrashImages).toHaveBeenCalledWith(expect.any(Function), ['img-1'])
    await waitFor(() => expect(screen.queryByTestId('right-panel')).not.toBeInTheDocument())
    expect(imageGrid()).toHaveAttribute('data-select-mode', 'false')
  })

  it('shows a partial-success toast when succeeded_count is less than the selection size', async () => {
    vi.mocked(bulkTrashImages).mockResolvedValue({ succeeded_count: 0 } as BulkActionResult)
    renderApp('/app')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.click(selectViaGrid())
    await userEvent.click(screen.getByRole('button', { name: 'Move selection to trash' }))

    await waitFor(() => expect(toast.warning).toHaveBeenCalledWith('Moved 0 of 1 images to trash'))
  })
})
