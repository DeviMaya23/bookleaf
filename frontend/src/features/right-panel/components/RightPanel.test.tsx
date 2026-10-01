import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest'
import RightPanel from './RightPanel'
import type { Image } from '@/lib/images'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({ getToken: vi.fn().mockResolvedValue('token') }),
}))

vi.mock('@/lib/images', () => ({
  getImage: vi.fn(),
  updateImage: vi.fn(),
  downloadImage: vi.fn(),
}))

vi.mock('@/lib/folders', () => ({
  getFolders: vi.fn().mockResolvedValue([{ id: 'folder-1', name: 'Nature' }]),
}))

vi.mock('@/lib/tags', () => ({
  getTags: vi.fn().mockResolvedValue([]),
  resolveOrCreateTags: vi.fn(),
}))

import { updateImage } from '@/lib/images'
import { getTags, resolveOrCreateTags } from '@/lib/tags'
import { getFolders } from '@/lib/folders'

function makeImage(overrides?: Partial<Image>): Image {
  return {
    id: 'img-1',
    title: 'Sunset photo',
    description: 'A nice sunset',
    mime_type: 'image/jpeg',
    source_url: 'https://example.com',
    folder_ids: [],
    thumbnail_url: 'https://example.com/thumb.jpg',
    width: 1920,
    height: 1080,
    file_size: 2 * 1024 * 1024,
    tags: [],
    position: null,
    created_at: '2026-05-01T00:00:00Z',
    updated_at: '2026-05-01T00:00:00Z',
    ...overrides,
  }
}

function renderPanel(image: Image) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <RightPanel panelContent={{ mode: 'image', image }} focusMode={false} />
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

describe('RightPanel — success scenario', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(updateImage).mockResolvedValue(makeImage())
  })

  it('renders title, notes, and source URL', () => {
    renderPanel(makeImage())

    expect(screen.getByDisplayValue('Sunset photo')).toBeInTheDocument()
    expect(screen.getByDisplayValue('A nice sunset')).toBeInTheDocument()
    expect(screen.getByDisplayValue('https://example.com')).toBeInTheDocument()
  })

  it('is hidden below sm breakpoint', () => {
    renderPanel(makeImage())

    expect(screen.getByRole('complementary').className).toMatch(/hidden sm:flex/)
  })
})

describe('RightPanel — pointer-capability shell', () => {
  const originalMatchMedia = window.matchMedia

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(updateImage).mockResolvedValue(makeImage())
  })

  afterEach(() => {
    window.matchMedia = originalMatchMedia
  })

  it('renders the sidebar shell on a fine-pointer device', () => {
    mockPointer(false)

    renderPanel(makeImage())

    expect(screen.getByRole('complementary').className).toMatch(/hidden sm:flex/)
    expect(screen.queryByTestId('mobile-drawer-shell-backdrop')).not.toBeInTheDocument()
  })

  it('renders the drawer shell on a coarse-pointer device', () => {
    mockPointer(true)

    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(
      <QueryClientProvider client={queryClient}>
        <RightPanel panelContent={{ mode: 'image', image: makeImage() }} focusMode={false} mobileOpen={true} />
      </QueryClientProvider>,
    )

    expect(screen.queryByRole('complementary')).not.toBeInTheDocument()
    expect(screen.getByTestId('mobile-drawer-shell-backdrop')).toBeInTheDocument()
  })

  it('renders the same content component in both shells', () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })

    mockPointer(true)
    render(
      <QueryClientProvider client={queryClient}>
        <RightPanel panelContent={{ mode: 'image', image: makeImage() }} focusMode={false} mobileOpen={true} />
      </QueryClientProvider>,
    )
    expect(screen.getByDisplayValue('Sunset photo')).toBeInTheDocument()

    mockPointer(false)
    renderPanel(makeImage())
    expect(screen.getAllByDisplayValue('Sunset photo').length).toBeGreaterThan(0)
  })
})

describe('RightPanel — neutral mode', () => {
  beforeEach(() => vi.clearAllMocks())

  it('renders the view label for All', () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(
      <QueryClientProvider client={queryClient}>
        <RightPanel panelContent={{ mode: 'neutral', viewLabel: 'All' }} focusMode={false} />
      </QueryClientProvider>,
    )
    expect(screen.getByText('All')).toBeInTheDocument()
  })

  it('renders the view label for Trash', () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(
      <QueryClientProvider client={queryClient}>
        <RightPanel panelContent={{ mode: 'neutral', viewLabel: 'Trash' }} focusMode={false} />
      </QueryClientProvider>,
    )
    expect(screen.getByText('Trash')).toBeInTheDocument()
  })

  it('returns null while focus mode is active', () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(
      <QueryClientProvider client={queryClient}>
        <RightPanel panelContent={{ mode: 'neutral', viewLabel: 'All' }} focusMode={true} />
      </QueryClientProvider>,
    )
    expect(screen.queryByRole('complementary')).not.toBeInTheDocument()
  })
})

describe('RightPanel — chevron toggle and collapsed state', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
  })

  it('starts expanded by default', () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(
      <QueryClientProvider client={queryClient}>
        <RightPanel panelContent={{ mode: 'neutral', viewLabel: 'All' }} focusMode={false} />
      </QueryClientProvider>,
    )
    expect(screen.getByRole('complementary').style.width).toBe('20rem')
  })

  it('collapses when the chevron button is clicked and persists to localStorage', async () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(
      <QueryClientProvider client={queryClient}>
        <RightPanel panelContent={{ mode: 'neutral', viewLabel: 'All' }} focusMode={false} />
      </QueryClientProvider>,
    )

    await userEvent.click(screen.getByRole('button', { name: 'Collapse panel' }))

    expect(screen.getByRole('complementary').style.width).toBe('2rem')
    expect(localStorage.getItem('bookleaf-right-panel-collapsed')).toBe('true')
  })

  it('reads collapsed state from localStorage on mount', () => {
    localStorage.setItem('bookleaf-right-panel-collapsed', 'true')
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(
      <QueryClientProvider client={queryClient}>
        <RightPanel panelContent={{ mode: 'neutral', viewLabel: 'All' }} focusMode={false} />
      </QueryClientProvider>,
    )
    expect(screen.getByRole('complementary').style.width).toBe('2rem')
  })
})

describe('RightPanel tags — success scenario', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(updateImage).mockResolvedValue(makeImage())
  })

  it('patches the image with the IDs returned by resolveOrCreateTags', async () => {
    vi.mocked(resolveOrCreateTags).mockResolvedValue([{ id: 'tag-new', name: 'concept' }])

    renderPanel(makeImage())

    const input = await screen.findByPlaceholderText('Add tags…')
    await userEvent.type(input, 'concept{Enter}')

    await waitFor(() => {
      expect(resolveOrCreateTags).toHaveBeenCalled()
    })
    await waitFor(() => {
      expect(updateImage).toHaveBeenCalledWith(
        expect.any(Function),
        'img-1',
        expect.objectContaining({ tags: ['tag-new'] }),
      )
    })
  })
})

describe('RightPanel tags — failure scenario', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('does not patch the image when tag resolution fails', async () => {
    vi.mocked(resolveOrCreateTags).mockRejectedValue(new Error('Failed to create tag "concept"'))

    renderPanel(makeImage())

    const input = await screen.findByPlaceholderText('Add tags…')
    await userEvent.type(input, 'concept{Enter}')

    await waitFor(() => {
      expect(resolveOrCreateTags).toHaveBeenCalled()
    })
    expect(updateImage).not.toHaveBeenCalledWith(
      expect.any(Function),
      'img-1',
      expect.objectContaining({ tags: expect.anything() }),
    )
  })
})

describe('RightPanel folders — success scenario', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getFolders).mockResolvedValue([{ id: 'folder-1', name: 'Nature', description: null, icon: null, parent_id: null, created_at: '', updated_at: '' }])
    vi.mocked(updateImage).mockResolvedValue(makeImage())
  })

  it('adding a folder calls PATCH with updated folder_ids', async () => {
    renderPanel(makeImage())

    const input = await screen.findByPlaceholderText('Add to folder…')
    await userEvent.type(input, 'nat')

    const option = await screen.findByText('Nature')
    await userEvent.click(option)

    await waitFor(() => {
      expect(updateImage).toHaveBeenCalledWith(
        expect.any(Function),
        'img-1',
        expect.objectContaining({ folder_ids: ['folder-1'] }),
      )
    })
  })
})

describe('RightPanel selection mode', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getFolders).mockResolvedValue([{ id: 'folder-1', name: 'Nature', description: null, icon: null, parent_id: null, created_at: '', updated_at: '' }])
  })

  function renderSelectionPanel(props: Partial<{ selectedCount: number; onApply: (params: { folderIds: string[]; tags: { id: string; name: string }[] }) => void; onMoveToTrash: () => void; onExitSelectMode: () => void; onDownloadZip: () => Promise<void> }> = {}) {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    return render(
      <QueryClientProvider client={queryClient}>
        <RightPanel
          panelContent={{
            mode: 'selection',
            selectedCount: props.selectedCount ?? 1,
            onApply: props.onApply ?? vi.fn(),
            onMoveToTrash: props.onMoveToTrash ?? vi.fn(),
            onExitSelectMode: props.onExitSelectMode ?? vi.fn(),
            onDownloadZip: props.onDownloadZip ?? vi.fn().mockResolvedValue(undefined),
          }}
          focusMode={false}
        />
      </QueryClientProvider>,
    )
  }

  it('renders the selected count', () => {
    renderSelectionPanel({ selectedCount: 1 })

    expect(screen.getByText('1 selected')).toBeInTheDocument()
  })

  it('calls onApply with folder id when a folder is added and Apply is clicked', async () => {
    const onApply = vi.fn()
    renderSelectionPanel({ onApply })

    await userEvent.type(screen.getByPlaceholderText('Add to folder…'), 'nat')
    await userEvent.click(await screen.findByText('Nature'))
    await userEvent.click(screen.getByRole('button', { name: /^apply$/i }))

    expect(onApply).toHaveBeenCalledWith({ folderIds: ['folder-1'], tags: [] })
  })

  it('calls onApply with tag when a tag is added and Apply is clicked', async () => {
    vi.mocked(getFolders).mockResolvedValue([])
    vi.mocked(getTags).mockResolvedValue([{ id: 'tag-1', name: 'nature' }])
    const onApply = vi.fn()
    renderSelectionPanel({ onApply })

    await userEvent.type(screen.getByPlaceholderText('Add tags…'), 'nat')
    await userEvent.click(await screen.findByText('nature'))
    await userEvent.click(screen.getByRole('button', { name: /^apply$/i }))

    expect(onApply).toHaveBeenCalledWith({ folderIds: [], tags: [{ id: 'tag-1', name: 'nature' }] })
  })

  it('calls onApply with both folder and tag when Apply is clicked', async () => {
    vi.mocked(getTags).mockResolvedValue([{ id: 'tag-1', name: 'summer' }])
    const onApply = vi.fn()
    renderSelectionPanel({ onApply })

    await userEvent.type(screen.getByPlaceholderText('Add to folder…'), 'nat')
    await userEvent.click(await screen.findByText('Nature'))
    await userEvent.type(screen.getByPlaceholderText('Add tags…'), 'sum')
    await userEvent.click(await screen.findByText('summer'))
    await userEvent.click(screen.getByRole('button', { name: /^apply$/i }))

    expect(onApply).toHaveBeenCalledWith({ folderIds: ['folder-1'], tags: [{ id: 'tag-1', name: 'summer' }] })
  })

  it('filters the folder list as the user types in the search input', async () => {
    vi.mocked(getFolders).mockResolvedValue([
      { id: 'folder-1', name: 'Nature', description: null, icon: null, parent_id: null, created_at: '', updated_at: '' },
      { id: 'folder-2', name: 'Work', description: null, icon: null, parent_id: null, created_at: '', updated_at: '' },
    ])
    renderSelectionPanel()

    await userEvent.type(screen.getByPlaceholderText('Add to folder…'), 'nat')

    expect(screen.getByText('Nature')).toBeInTheDocument()
    expect(screen.queryByText('Work')).not.toBeInTheDocument()
  })

  it('calls onMoveToTrash when the trash action is clicked', async () => {
    const onMoveToTrash = vi.fn()
    renderSelectionPanel({ onMoveToTrash })

    await userEvent.click(screen.getByRole('button', { name: /move to trash/i }))
    const dialog = await screen.findByRole('dialog')
    await userEvent.click(within(dialog).getByRole('button', { name: /move to trash/i }))

    expect(onMoveToTrash).toHaveBeenCalled()
  })

  it('renders the sidebar shell on a fine-pointer device, taking priority over image mode', () => {
    renderSelectionPanel({ selectedCount: 3 })

    expect(screen.getByRole('complementary')).toBeInTheDocument()
    expect(screen.getByText('3 selected')).toBeInTheDocument()
  })
})

describe('RightPanel folders — failure scenario', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getFolders).mockResolvedValue([{ id: 'folder-1', name: 'Nature', description: null, icon: null, parent_id: null, created_at: '', updated_at: '' }])
    vi.mocked(updateImage).mockRejectedValue(new Error('Server error'))
  })

  it('shows an error toast when PATCH fails after folder change', async () => {
    renderPanel(makeImage())

    const input = await screen.findByPlaceholderText('Add to folder…')
    await userEvent.type(input, 'nat')

    const option = await screen.findByText('Nature')
    await userEvent.click(option)

    await waitFor(() => {
      expect(updateImage).toHaveBeenCalledWith(
        expect.any(Function),
        'img-1',
        expect.objectContaining({ folder_ids: ['folder-1'] }),
      )
    })
  })
})
