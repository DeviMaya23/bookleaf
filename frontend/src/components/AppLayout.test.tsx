import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Routes, Route, Link } from 'react-router-dom'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import AppLayout from './AppLayout'
import { getFolders } from '@/lib/folders'
import { getTags } from '@/lib/tags'

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
  default: () => <div data-testid="folder-sidebar" />,
}))

vi.mock('./ImageViewer', () => ({
  default: () => null,
}))

vi.mock('./UploadModal', () => ({
  default: () => null,
}))

vi.mock('./BatchUploadModal', () => ({
  default: () => null,
}))

vi.mock('./RightPanel', () => ({
  default: () => null,
}))

vi.mock('./ImageGrid', () => ({
  default: ({ sortBy, sortDir, filterTagIds, filterMimeTypes, filterFolderIds }: { sortBy: string; sortDir: string | undefined; filterTagIds?: string[]; filterMimeTypes?: string[]; filterFolderIds?: string[] }) => (
    <div
      data-testid="image-grid"
      data-sort-by={sortBy}
      data-sort-dir={sortDir ?? ''}
      data-filter-tag-ids={(filterTagIds ?? []).join(',')}
      data-filter-mime-types={(filterMimeTypes ?? []).join(',')}
      data-filter-folder-ids={(filterFolderIds ?? []).join(',')}
    />
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

  it('shows Tags, File type, and Folder sections in the All view', async () => {
    vi.mocked(getTags).mockResolvedValue([{ id: 'tag-1', name: 'Cats' }])
    vi.mocked(getFolders).mockResolvedValue([makeFolder('folder-1', 'Vacation')])

    renderApp('/')

    await waitFor(() => expect(screen.getByRole('menuitemcheckbox', { name: 'Cats' })).toBeInTheDocument())
    expect(screen.getByText('Tags')).toBeInTheDocument()
    expect(screen.getByText('File type')).toBeInTheDocument()
    expect(screen.getByText('Folder')).toBeInTheDocument()
    expect(screen.getByRole('menuitemcheckbox', { name: 'JPEG' })).toBeInTheDocument()
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
    await userEvent.click(screen.getByRole('menuitemcheckbox', { name: 'JPEG' }))

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
