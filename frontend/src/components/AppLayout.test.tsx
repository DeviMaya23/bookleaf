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

vi.mock('@/hooks/useVisionSuggestion', () => ({
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

vi.mock('./FolderSidebar', () => ({
  default: ({ onFolderSelect }: { onFolderSelect?: () => void }) => (
    <div data-testid="folder-sidebar">
      <button onClick={onFolderSelect}>Select folder</button>
    </div>
  ),
}))

vi.mock('./ImageViewer', () => ({
  default: ({ focusMode }: { focusMode: boolean }) => (
    <div data-testid="image-viewer" data-focus-mode={String(focusMode)} />
  ),
}))

vi.mock('./UploadModal', () => ({
  default: () => null,
}))

vi.mock('./BatchUploadModal', () => ({
  default: () => null,
}))

vi.mock('./RightPanel', () => ({
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

vi.mock('./ImageGrid', () => ({
  default: ({ sortBy, sortDir, filterTagIds, filterMimeTypes, filterFolderIds, onImageSelect, onImageDoubleClick }: { sortBy: string; sortDir: string | undefined; filterTagIds?: string[]; filterMimeTypes?: string[]; filterFolderIds?: string[]; onImageSelect: (img: Image) => void; onImageDoubleClick: (img: Image) => void }) => (
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
    </div>
  ),
}))

function renderApp(initialPath: string) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[initialPath]}>
        <nav>
          <Link to="/">All</Link>
          <Link to="/folders/folder-1">Folder 1</Link>
          <Link to="/folders/folder-2">Folder 2</Link>
        </nav>
        <Routes>
          <Route path="/" element={<AppLayout />} />
          <Route path="/unsorted" element={<AppLayout />} />
          <Route path="/trash" element={<AppLayout />} />
          <Route path="/folders/:folderId" element={<AppLayout />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

function imageGrid() {
  return screen.getByTestId('image-grid')
}

describe('AppLayout sort control', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('switching from a folder with an explicit sort to All resets to Date added / newest first', async () => {
    renderApp('/folders/folder-1')

    await waitFor(() => expect(imageGrid()).toHaveAttribute('data-sort-by', 'manual'))

    await userEvent.click(screen.getByRole('menuitemradio', { name: 'Name' }))

    await waitFor(() => expect(imageGrid()).toHaveAttribute('data-sort-by', 'title'))

    await userEvent.click(screen.getByText('All'))

    await waitFor(() => {
      expect(imageGrid()).toHaveAttribute('data-sort-by', 'created_at')
      expect(imageGrid()).toHaveAttribute('data-sort-dir', 'desc')
    })

    expect(screen.getByRole('menuitemradio', { name: 'Date added' })).toHaveAttribute('aria-checked', 'true')
    expect(screen.getByText('Newest first')).toBeInTheDocument()
  })

  it('switching between two folders resets sort to Manual', async () => {
    renderApp('/folders/folder-1')

    await waitFor(() => expect(imageGrid()).toHaveAttribute('data-sort-by', 'manual'))

    await userEvent.click(screen.getByRole('menuitemradio', { name: 'Name' }))
    await waitFor(() => expect(imageGrid()).toHaveAttribute('data-sort-by', 'title'))

    await userEvent.click(screen.getByText('Folder 2'))

    await waitFor(() => {
      expect(imageGrid()).toHaveAttribute('data-sort-by', 'manual')
      expect(imageGrid()).toHaveAttribute('data-sort-dir', '')
    })

    expect(screen.getByRole('menuitemradio', { name: 'Manual' })).toHaveAttribute('aria-checked', 'true')
  })

  it('hides the direction toggle for Manual and shows it with field-appropriate labels otherwise', async () => {
    renderApp('/folders/folder-1')
    await waitFor(() => expect(imageGrid()).toHaveAttribute('data-sort-by', 'manual'))

    expect(screen.queryByText(/oldest first|newest first|a → z|z → a/i)).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('menuitemradio', { name: 'Date added' }))
    await waitFor(() => expect(imageGrid()).toHaveAttribute('data-sort-by', 'created_at'))
    expect(screen.getByText('Newest first')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('menuitemradio', { name: 'Name' }))
    await waitFor(() => expect(imageGrid()).toHaveAttribute('data-sort-by', 'title'))
    expect(screen.getByText('A → Z')).toBeInTheDocument()
  })

  it('shows the sort trigger as active only when the selection differs from the view default', async () => {
    renderApp('/folders/folder-1')
    await waitFor(() => expect(imageGrid()).toHaveAttribute('data-sort-by', 'manual'))

    const trigger = screen.getByRole('button', { name: /sort/i })
    expect(trigger.className).not.toMatch(/bg-primary/)

    await userEvent.click(screen.getByRole('menuitemradio', { name: 'Name' }))
    await waitFor(() => expect(imageGrid()).toHaveAttribute('data-sort-by', 'title'))

    expect(trigger.className).toMatch(/bg-primary/)
  })
})

function makeFolder(id: string, name: string) {
  return { id, name, description: null, parent_id: null, created_at: '', updated_at: '' }
}

describe('AppLayout filter controls', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getFolders).mockResolvedValue([])
    vi.mocked(getTags).mockResolvedValue([])
  })

  it('hides the Filters button in Trash', async () => {
    renderApp('/trash')

    await waitFor(() => expect(imageGrid()).toBeInTheDocument())
    expect(screen.queryByRole('button', { name: /filters/i })).not.toBeInTheDocument()
  })

  it('shows File type, Tags, and Folder sections, in that order, in the All view', async () => {
    vi.mocked(getTags).mockResolvedValue([{ id: 'tag-1', name: 'Cats' }])
    vi.mocked(getFolders).mockResolvedValue([makeFolder('folder-1', 'Vacation')])

    renderApp('/')

    await waitFor(() => expect(screen.getByRole('menuitemcheckbox', { name: 'Cats' })).toBeInTheDocument())
    expect(screen.getByText('Tags')).toBeInTheDocument()
    expect(screen.getByText('File type')).toBeInTheDocument()
    expect(screen.getByText('Folder')).toBeInTheDocument()
    const filterPanel = screen.getAllByTestId('dropdown-content').find((el) => el.textContent?.includes('File type'))
    const panelText = filterPanel?.textContent ?? ''
    expect(panelText.indexOf('File type')).toBeLessThan(panelText.indexOf('Tags'))
    expect(panelText.indexOf('Tags')).toBeLessThan(panelText.indexOf('Folder'))
    expect(screen.getByRole('button', { name: 'JPEG' })).toBeInTheDocument()
    await waitFor(() => expect(screen.getByRole('menuitemcheckbox', { name: 'Vacation' })).toBeInTheDocument())
  })

  it('shows only Tags and File type sections in the Unsorted view', async () => {
    vi.mocked(getTags).mockResolvedValue([{ id: 'tag-1', name: 'Cats' }])
    vi.mocked(getFolders).mockResolvedValue([makeFolder('folder-1', 'Vacation')])

    renderApp('/unsorted')

    await waitFor(() => expect(screen.getByRole('button', { name: /filters/i })).toBeInTheDocument())
    expect(screen.getByText('Tags')).toBeInTheDocument()
    expect(screen.getByText('File type')).toBeInTheDocument()
    expect(screen.queryByText('Folder')).not.toBeInTheDocument()
    expect(screen.queryByRole('menuitemcheckbox', { name: 'Vacation' })).not.toBeInTheDocument()
  })

  it('shows only Tags and File type sections in a folder view', async () => {
    vi.mocked(getTags).mockResolvedValue([{ id: 'tag-1', name: 'Cats' }])
    vi.mocked(getFolders).mockResolvedValue([makeFolder('folder-1', 'Vacation')])

    renderApp('/folders/folder-1')

    await waitFor(() => expect(screen.getByRole('button', { name: /filters/i })).toBeInTheDocument())
    expect(screen.getByText('Tags')).toBeInTheDocument()
    expect(screen.getByText('File type')).toBeInTheDocument()
    expect(screen.queryByText('Folder')).not.toBeInTheDocument()
  })

  it('shows a badge with the total filter count and switches to the active variant', async () => {
    vi.mocked(getTags).mockResolvedValue([
      { id: 'tag-1', name: 'Cats' },
      { id: 'tag-2', name: 'Dogs' },
    ])

    renderApp('/')

    await waitFor(() => expect(screen.getByRole('menuitemcheckbox', { name: 'Cats' })).toBeInTheDocument())
    const filtersButton = screen.getByRole('button', { name: /filters/i })
    expect(filtersButton.className).not.toMatch(/bg-primary/)
    expect(screen.queryByText('3')).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('menuitemcheckbox', { name: 'Cats' }))
    await userEvent.click(screen.getByRole('menuitemcheckbox', { name: 'Dogs' }))
    await userEvent.click(screen.getByRole('button', { name: 'JPEG' }))

    await waitFor(() => expect(screen.getByText('3')).toBeInTheDocument())
    expect(filtersButton.className).toMatch(/bg-primary/)
  })

  it('clears active filters when switching views', async () => {
    vi.mocked(getTags).mockResolvedValue([{ id: 'tag-1', name: 'Cats' }])

    renderApp('/folders/folder-1')

    await waitFor(() => expect(screen.getByRole('menuitemcheckbox', { name: 'Cats' })).toBeInTheDocument())
    await userEvent.click(screen.getByRole('menuitemcheckbox', { name: 'Cats' }))

    await waitFor(() => expect(imageGrid()).toHaveAttribute('data-filter-tag-ids', 'tag-1'))
    expect(screen.getByRole('button', { name: 'Remove filter Cats' })).toBeInTheDocument()

    await userEvent.click(screen.getByText('All'))

    await waitFor(() => expect(imageGrid()).toHaveAttribute('data-filter-tag-ids', ''))
    expect(screen.queryByRole('button', { name: 'Remove filter Cats' })).not.toBeInTheDocument()
    expect(screen.queryByText('Clear all')).not.toBeInTheDocument()
  })

  it('searching tags filters the tag list without affecting the folder list', async () => {
    vi.mocked(getTags).mockResolvedValue([
      { id: 'tag-1', name: 'Cats' },
      { id: 'tag-2', name: 'Dogs' },
    ])
    vi.mocked(getFolders).mockResolvedValue([makeFolder('folder-1', 'Vacation')])

    renderApp('/')

    await waitFor(() => expect(screen.getByRole('menuitemcheckbox', { name: 'Cats' })).toBeInTheDocument())
    await waitFor(() => expect(screen.getByRole('menuitemcheckbox', { name: 'Vacation' })).toBeInTheDocument())

    await userEvent.type(screen.getByPlaceholderText('Search tags…'), 'Ca')

    expect(screen.getByRole('menuitemcheckbox', { name: 'Cats' })).toBeInTheDocument()
    expect(screen.queryByRole('menuitemcheckbox', { name: 'Dogs' })).not.toBeInTheDocument()
    expect(screen.getByRole('menuitemcheckbox', { name: 'Vacation' })).toBeInTheDocument()
  })

  it('searching folders filters the folder list without affecting the tag list', async () => {
    vi.mocked(getTags).mockResolvedValue([{ id: 'tag-1', name: 'Cats' }])
    vi.mocked(getFolders).mockResolvedValue([
      makeFolder('folder-1', 'Vacation'),
      makeFolder('folder-2', 'Work'),
    ])

    renderApp('/')

    await waitFor(() => expect(screen.getByRole('menuitemcheckbox', { name: 'Cats' })).toBeInTheDocument())
    await waitFor(() => expect(screen.getByRole('menuitemcheckbox', { name: 'Vacation' })).toBeInTheDocument())

    await userEvent.type(screen.getByPlaceholderText('Search folders…'), 'Vac')

    expect(screen.getByRole('menuitemcheckbox', { name: 'Vacation' })).toBeInTheDocument()
    expect(screen.queryByRole('menuitemcheckbox', { name: 'Work' })).not.toBeInTheDocument()
    expect(screen.getByRole('menuitemcheckbox', { name: 'Cats' })).toBeInTheDocument()
  })

  it('hides all tags when the search matches none, even a checked tag, while keeping it active', async () => {
    vi.mocked(getTags).mockResolvedValue([
      { id: 'tag-1', name: 'Cats' },
      { id: 'tag-2', name: 'Dogs' },
    ])

    renderApp('/')

    await waitFor(() => expect(screen.getByRole('menuitemcheckbox', { name: 'Cats' })).toBeInTheDocument())
    await userEvent.click(screen.getByRole('menuitemcheckbox', { name: 'Cats' }))
    await waitFor(() => expect(imageGrid()).toHaveAttribute('data-filter-tag-ids', 'tag-1'))

    await userEvent.type(screen.getByPlaceholderText('Search tags…'), 'zzz')

    expect(screen.queryByRole('menuitemcheckbox', { name: 'Cats' })).not.toBeInTheDocument()
    expect(screen.queryByRole('menuitemcheckbox', { name: 'Dogs' })).not.toBeInTheDocument()
    expect(imageGrid()).toHaveAttribute('data-filter-tag-ids', 'tag-1')
    expect(screen.getByRole('button', { name: 'Remove filter Cats' })).toBeInTheDocument()
  })

  it('typing in the tag search input does not close the filter panel', async () => {
    vi.mocked(getTags).mockResolvedValue([{ id: 'tag-1', name: 'Cats' }])

    renderApp('/')

    await waitFor(() => expect(screen.getByRole('menuitemcheckbox', { name: 'Cats' })).toBeInTheDocument())

    const searchInput = screen.getByPlaceholderText('Search tags…')
    await userEvent.type(searchInput, 'C')

    expect(searchInput).toHaveValue('C')
    expect(screen.getByRole('menuitemcheckbox', { name: 'Cats' })).toBeInTheDocument()
  })

  it('removes a single filter via its chip and clears all via Clear all', async () => {
    vi.mocked(getTags).mockResolvedValue([
      { id: 'tag-1', name: 'Cats' },
      { id: 'tag-2', name: 'Dogs' },
    ])

    renderApp('/')

    await waitFor(() => expect(screen.getByRole('menuitemcheckbox', { name: 'Cats' })).toBeInTheDocument())
    await userEvent.click(screen.getByRole('menuitemcheckbox', { name: 'Cats' }))
    await userEvent.click(screen.getByRole('menuitemcheckbox', { name: 'Dogs' }))

    await waitFor(() => expect(imageGrid()).toHaveAttribute('data-filter-tag-ids', 'tag-1,tag-2'))

    await userEvent.click(screen.getByRole('button', { name: 'Remove filter Cats' }))
    await waitFor(() => expect(imageGrid()).toHaveAttribute('data-filter-tag-ids', 'tag-2'))

    await userEvent.click(screen.getByRole('button', { name: 'Clear all' }))
    await waitFor(() => expect(imageGrid()).toHaveAttribute('data-filter-tag-ids', ''))
    expect(screen.queryByText('Clear all')).not.toBeInTheDocument()
  })
})

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
    renderApp('/')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    expect(screen.getByTestId('folder-sidebar')).toBeInTheDocument()
    expect(document.querySelector('main')!.className).toMatch(/ml-\[240px\]/)

    await userEvent.click(focusToggle())

    expect(screen.queryByTestId('folder-sidebar')).not.toBeInTheDocument()
    expect(document.querySelector('main')!.className).not.toMatch(/ml-\[240px\]/)
  })

  it('enabling focus mode hides an open RightPanel in image mode', async () => {
    renderApp('/')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.click(screen.getByRole('button', { name: 'Select image' }))
    expect(screen.getByTestId('right-panel')).toHaveAttribute('data-mode', 'image')

    await userEvent.click(focusToggle())

    expect(screen.queryByTestId('right-panel')).not.toBeInTheDocument()
  })

  it('enabling focus mode hides an open RightPanel in folder mode', async () => {
    vi.mocked(getFolders).mockResolvedValue([makeFolder('folder-1', 'Vacation')])
    renderApp('/folders/folder-1')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.click(screen.getByRole('button', { name: 'Select folder' }))
    await waitFor(() => expect(screen.getByTestId('right-panel')).toHaveAttribute('data-mode', 'folder'))

    await userEvent.click(focusToggle())

    expect(screen.queryByTestId('right-panel')).not.toBeInTheDocument()
  })

  it('clicking an image card while focus mode is active updates selection without rendering RightPanel, revealed once focus mode is disabled', async () => {
    renderApp('/')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.click(focusToggle())
    expect(screen.queryByTestId('folder-sidebar')).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Select image' }))
    expect(screen.queryByTestId('right-panel')).not.toBeInTheDocument()

    await userEvent.click(focusToggle())
    expect(screen.getByTestId('right-panel')).toHaveAttribute('data-mode', 'image')
  })

  it('double-clicking an image card while focus mode is active opens the viewer at full width with no RightPanel', async () => {
    renderApp('/')
    await waitFor(() => expect(imageGrid()).toBeInTheDocument())

    await userEvent.click(focusToggle())
    await userEvent.dblClick(screen.getByRole('button', { name: 'Open image' }))

    expect(screen.getByTestId('image-viewer')).toHaveAttribute('data-focus-mode', 'true')
    expect(screen.queryByTestId('right-panel')).not.toBeInTheDocument()
    expect(document.querySelector('main')!.className).not.toMatch(/ml-\[240px\]/)
  })
})
