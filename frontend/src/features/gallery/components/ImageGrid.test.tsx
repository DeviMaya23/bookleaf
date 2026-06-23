import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest'
import { DndContext } from '@dnd-kit/core'
import { SortableContext } from '@dnd-kit/sortable'
import ImageGrid from './ImageGrid'
import type { AppView } from '@/lib/view'

vi.mock('@dnd-kit/sortable', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@dnd-kit/sortable')>()
  return {
    ...actual,
    SortableContext: vi.fn(actual.SortableContext),
  }
})

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({ getToken: vi.fn().mockResolvedValue('token') }),
}))

vi.mock('@/lib/images', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/images')>()),
  getImages: vi.fn(),
  getFolderImages: vi.fn(),
  getAllImages: vi.fn(),
  getTrashedImages: vi.fn(),
  deleteImage: vi.fn(),
  hardDeleteImage: vi.fn(),
  restoreImage: vi.fn(),
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
      React.createElement(
        'button',
        { role: 'menuitem', className, onClick },
        children,
      ),
  }
})

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

import { getImages, getFolderImages, getTrashedImages, hardDeleteImage } from '@/lib/images'
import { computeNewPosition } from '@/lib/images'
import type { Image } from '@/lib/images'

function makeImage(overrides?: Partial<Image>): Image {
  return {
    id: '1',
    title: 'Test image',
    description: null,
    mime_type: 'image/jpeg',
    source_url: null,
    folder_ids: [],
    thumbnail_url: 'https://example.com/thumb.jpg',
    width: 100,
    height: 100,
    file_size: 1024,
    position: null,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    tags: [],
    ...overrides,
  }
}

function renderImageGrid(
  view: AppView = { type: 'unsorted' },
  onImageSelect = vi.fn(),
  sortBy: 'manual' | 'created_at' | 'title' = 'manual',
) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <DndContext sensors={[]}>
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>
          <ImageGrid
            view={view}
            searchTerm=""
            debouncedSearchTerm=""
            sortBy={sortBy}
            sortDir={undefined}
            onImageSelect={onImageSelect}
          />
        </MemoryRouter>
      </QueryClientProvider>
    </DndContext>,
  )
}

describe('ImageGrid', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows a loading spinner while images are loading', () => {
    vi.mocked(getImages).mockReturnValue(new Promise(() => {}))

    const { container } = renderImageGrid()

    expect(container.querySelector('.animate-spin')).toBeInTheDocument()
  })

  it('renders image cards when images are returned', async () => {
    vi.mocked(getImages).mockResolvedValue({ images: [makeImage()], next_cursor: null })

    renderImageGrid()

    await waitFor(() => {
      expect(screen.getByText('Test image')).toBeInTheDocument()
    })
  })

  it('shows empty state when no images are returned', async () => {
    vi.mocked(getImages).mockResolvedValue({ images: [], next_cursor: null })

    renderImageGrid()

    await waitFor(() => {
      expect(screen.getByText('No images here yet')).toBeInTheDocument()
    })
  })

  it('calls onImageSelect when a card is clicked', async () => {
    const image = makeImage()
    const onImageSelect = vi.fn()
    vi.mocked(getImages).mockResolvedValue({ images: [image], next_cursor: null })

    renderImageGrid({ type: 'unsorted' }, onImageSelect)

    await waitFor(() => {
      expect(screen.getByText('Test image')).toBeInTheDocument()
    })

    await userEvent.click(screen.getByText('Test image'))

    expect(onImageSelect).toHaveBeenCalledWith(image)
  })
})

describe('ImageGrid pagination', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows Load more button when next_cursor is non-null', async () => {
    vi.mocked(getImages).mockResolvedValue({
      images: [makeImage()],
      next_cursor: 'cursor-abc',
    })

    renderImageGrid()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /load more/i })).toBeInTheDocument()
    })
  })

  it('hides Load more button when next_cursor is null', async () => {
    vi.mocked(getImages).mockResolvedValue({ images: [makeImage()], next_cursor: null })

    renderImageGrid()

    await waitFor(() => {
      expect(screen.getByText('Test image')).toBeInTheDocument()
    })

    expect(screen.queryByRole('button', { name: /load more/i })).not.toBeInTheDocument()
  })
})

describe('computeNewPosition', () => {
  it('returns a key between neighbours for a middle-position move', () => {
    const images = [
      makeImage({ id: 'a', position: 'a0' }),
      makeImage({ id: 'b', position: 'a1' }),
      makeImage({ id: 'c', position: 'a2' }),
    ]
    const newKey = computeNewPosition(images, 1)
    expect(newKey > 'a0').toBe(true)
    expect(newKey < 'a2').toBe(true)
  })

  it('returns a key before the next item when newIndex is 0 (no previous neighbour)', () => {
    const images = [
      makeImage({ id: 'a', position: 'a1' }),
      makeImage({ id: 'b', position: 'a2' }),
    ]
    const newKey = computeNewPosition(images, 0)
    expect(typeof newKey).toBe('string')
    expect(newKey.length).toBeGreaterThan(0)
    // new key must be before the item now at index 1
    expect(newKey < images[1].position!).toBe(true)
  })

  it('does not throw and returns a key after prev when neighbours have duplicate positions', () => {
    const images = [
      makeImage({ id: 'a', position: 'a0' }),
      makeImage({ id: 'b', position: null }),
      makeImage({ id: 'c', position: 'a0' }),
    ]
    expect(() => computeNewPosition(images, 1)).not.toThrow()
    const newKey = computeNewPosition(images, 1)
    expect(newKey > 'a0').toBe(true)
  })

  it('does not throw when a neighbour has an empty-string position', () => {
    const images = [
      makeImage({ id: 'a', position: '' }),
      makeImage({ id: 'b', position: null }),
      makeImage({ id: 'c', position: 'a1' }),
    ]
    expect(() => computeNewPosition(images, 1)).not.toThrow()
  })
})

describe('ImageGrid trash view — permanent delete', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows Restore and Delete permanently items in context menu', async () => {
    vi.mocked(getTrashedImages).mockResolvedValue({ images: [makeImage()], next_cursor: null })

    renderImageGrid({ type: 'trash' })

    await waitFor(() => {
      expect(screen.getByText('Test image')).toBeInTheDocument()
    })

    expect(screen.getByRole('menuitem', { name: /restore/i })).toBeInTheDocument()
    expect(screen.getByRole('menuitem', { name: /delete permanently/i })).toBeInTheDocument()
  })

  it('calls hardDeleteImage after confirming the dialog', async () => {
    vi.mocked(getTrashedImages).mockResolvedValue({ images: [makeImage()], next_cursor: null })
    vi.mocked(hardDeleteImage).mockResolvedValue(undefined)

    renderImageGrid({ type: 'trash' })

    await waitFor(() => {
      expect(screen.getByText('Test image')).toBeInTheDocument()
    })

    await userEvent.click(screen.getByRole('menuitem', { name: /delete permanently/i }))

    const confirmButton = await screen.findByRole('button', { name: /delete permanently/i })
    await userEvent.click(confirmButton)

    await waitFor(() => {
      expect(hardDeleteImage).toHaveBeenCalledWith(expect.any(Function), '1')
    })
  })

  it('does not call hardDeleteImage when the confirmation dialog is cancelled', async () => {
    vi.mocked(getTrashedImages).mockResolvedValue({ images: [makeImage()], next_cursor: null })

    renderImageGrid({ type: 'trash' })

    await waitFor(() => {
      expect(screen.getByText('Test image')).toBeInTheDocument()
    })

    await userEvent.click(screen.getByRole('menuitem', { name: /delete permanently/i }))

    const cancelButton = await screen.findByRole('button', { name: /cancel/i })
    await userEvent.click(cancelButton)

    expect(hardDeleteImage).not.toHaveBeenCalled()
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

describe('ImageGrid — View details context menu item', () => {
  const originalMatchMedia = window.matchMedia

  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    window.matchMedia = originalMatchMedia
  })

  it('shows "View details" in the context menu on a coarse-pointer device', async () => {
    mockPointer(true)
    vi.mocked(getImages).mockResolvedValue({ images: [makeImage()], next_cursor: null })

    renderImageGrid()

    await waitFor(() => {
      expect(screen.getByText('Test image')).toBeInTheDocument()
    })

    expect(screen.getByRole('menuitem', { name: /view details/i })).toBeInTheDocument()
  })

  it('does not show "View details" in the context menu on a fine-pointer device', async () => {
    mockPointer(false)
    vi.mocked(getImages).mockResolvedValue({ images: [makeImage()], next_cursor: null })

    renderImageGrid()

    await waitFor(() => {
      expect(screen.getByText('Test image')).toBeInTheDocument()
    })

    expect(screen.queryByRole('menuitem', { name: /view details/i })).not.toBeInTheDocument()
  })

  it('calls onViewDetails with the correct image when selected', async () => {
    mockPointer(true)
    const image = makeImage()
    const onViewDetails = vi.fn()
    vi.mocked(getImages).mockResolvedValue({ images: [image], next_cursor: null })

    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(
      <DndContext sensors={[]}>
        <QueryClientProvider client={queryClient}>
          <MemoryRouter>
            <ImageGrid
              view={{ type: 'unsorted' }}
              searchTerm=""
              debouncedSearchTerm=""
              sortBy="manual"
              sortDir={undefined}
              onImageSelect={vi.fn()}
              onViewDetails={onViewDetails}
            />
          </MemoryRouter>
        </QueryClientProvider>
      </DndContext>,
    )

    await waitFor(() => {
      expect(screen.getByText('Test image')).toBeInTheDocument()
    })

    await userEvent.click(screen.getByRole('menuitem', { name: /view details/i }))

    expect(onViewDetails).toHaveBeenCalledWith(image)
  })
})

describe('ImageGrid layout', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('wraps the grid in SortableContext for a folder view with manual sort', async () => {
    vi.mocked(getFolderImages).mockResolvedValue({ images: [makeImage()], next_cursor: null })

    renderImageGrid({ type: 'folder', id: 'folder-1' }, vi.fn(), 'manual')

    await waitFor(() => {
      expect(screen.getByText('Test image')).toBeInTheDocument()
    })

    expect(SortableContext).toHaveBeenCalled()
  })

  it('does not wrap the grid in SortableContext for a non-folder view', async () => {
    vi.mocked(getImages).mockResolvedValue({ images: [makeImage()], next_cursor: null })

    renderImageGrid({ type: 'unsorted' })

    await waitFor(() => {
      expect(screen.getByText('Test image')).toBeInTheDocument()
    })

    expect(SortableContext).not.toHaveBeenCalled()
  })

  it('does not wrap the grid in SortableContext for a folder view with an explicit sort', async () => {
    vi.mocked(getFolderImages).mockResolvedValue({ images: [makeImage()], next_cursor: null })

    renderImageGrid({ type: 'folder', id: 'folder-1' }, vi.fn(), 'title')

    await waitFor(() => {
      expect(screen.getByText('Test image')).toBeInTheDocument()
    })

    expect(SortableContext).not.toHaveBeenCalled()
  })

  it('renders the masonry grid with column count derived from the container width', async () => {
    vi.mocked(getImages).mockResolvedValue({
      images: Array.from({ length: 6 }, (_, i) => makeImage({ id: String(i), title: `Image ${i}` })),
      next_cursor: null,
    })

    const originalRO = globalThis.ResizeObserver
    globalThis.ResizeObserver = class {
      private callback: ResizeObserverCallback
      constructor(cb: ResizeObserverCallback) {
        this.callback = cb
      }
      observe(target: Element) {
        this.callback([{ contentRect: { width: 1200 } } as ResizeObserverEntry], this as unknown as ResizeObserver)
        Object.defineProperty(target, 'getBoundingClientRect', {
          configurable: true,
          value: () => ({ width: 1200 }),
        })
      }
      unobserve() {}
      disconnect() {}
    }

    try {
      const { container } = renderImageGrid()

      await waitFor(() => {
        expect(screen.getByText('Image 0')).toBeInTheDocument()
      })

      // numCols = floor(1200 / 220) = 5
      expect(container.querySelectorAll('.flex.flex-col').length).toBe(5)
    } finally {
      globalThis.ResizeObserver = originalRO
    }
  })
})

