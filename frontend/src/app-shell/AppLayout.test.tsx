import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Routes, Route, Link } from 'react-router-dom'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import AppLayout from './AppLayout'
import { getFolders } from '@/lib/folders'
import { getTags } from '@/lib/tags'
import type { Image } from '@/lib/images'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({ getToken: vi.fn().mockResolvedValue('token') }),
}))

vi.mock('@/lib/folders', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/folders')>()),
  getFolders: vi.fn().mockResolvedValue([]),
}))

vi.mock('@/lib/tags', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/tags')>()),
  getTags: vi.fn().mockResolvedValue([]),
}))

vi.mock('./useVisionSuggestion', () => ({
  useVisionSuggestion: () => ({ checkVision: vi.fn() }),
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
  default: ({ onFolderSelect }: { onFolderSelect?: () => void }) => (
    <div data-testid="folder-sidebar">
      <button onClick={onFolderSelect}>Select folder</button>
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
  default: (props: { mode: 'image'; image: Image } | { mode: 'folder'; folder: { name: string } }) => (
    <div data-testid="right-panel" data-mode={props.mode}>
      {props.mode === 'image' ? props.image.title : props.folder.name}
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
  default: ({ sortBy, sortDir, filterTagIds, filterMimeTypes, filterFolderIds, onImageSelect, onImageDoubleClick, onImageDeleted }: { sortBy: string; sortDir: string | undefined; filterTagIds?: string[]; filterMimeTypes?: string[]; filterFolderIds?: string[]; onImageSelect: (img: Image) => void; onImageDoubleClick: (img: Image) => void; onImageDeleted: (id: string) => void }) => (
    <div
      data-testid="image-grid"
      data-sort-by={sortBy}
      data-sort-dir={sortDir ?? ''}
      data-filter-tag-ids={(filterTagIds ?? []).join(',')}
      data-filter-mime-types={(filterMimeTypes ?? []).join(',')}
      data-filter-folder-ids={(filterFolderIds ?? []).join(',')}
    >
      <button onClick={() => onImageSelect(makeTestImage())}>Select image</button>
      <button onDoubleClick={() => onImageDoubleClick(makeTestImage())}>Open image</button>
      <button onClick={() => onImageDeleted('img-1')}>Delete image</button>
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
  return { id, name, description: null, parent_id: null, created_at: '', updated_at: '' }
}

describe('AppLayout focus mode', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getFolders).mockResolvedValue([])
    vi.mocked(getTags).mockResolvedValue([])
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

describe('AppLayout cross-feature integration', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getFolders).mockResolvedValue([])
    vi.mocked(getTags).mockResolvedValue([])
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
