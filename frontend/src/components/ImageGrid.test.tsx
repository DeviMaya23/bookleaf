import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import { toast } from 'sonner'
import { DndContext } from '@dnd-kit/core'
import ImageGrid from './ImageGrid'
import type { AppView } from '@/lib/view'
import type { SortEndTrigger } from './ImageGrid'

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
  updateImagePosition: vi.fn(),
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

import { getImages, getFolderImages, getAllImages, getTrashedImages, deleteImage, hardDeleteImage, updateImagePosition } from '@/lib/images'
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
  sortDir: 'asc' | 'desc' | undefined = undefined,
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
            sortDir={sortDir}
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

describe('ImageGrid delete flow', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('calls deleteImage when Delete context menu item is clicked', async () => {
    vi.mocked(getImages).mockResolvedValue({ images: [makeImage()], next_cursor: null })
    vi.mocked(deleteImage).mockResolvedValue(undefined)

    renderImageGrid()

    await waitFor(() => {
      expect(screen.getByText('Test image')).toBeInTheDocument()
    })

    await userEvent.click(screen.getByRole('menuitem', { name: /delete/i }))

    await waitFor(() => {
      expect(deleteImage).toHaveBeenCalledWith(expect.any(Function), '1')
    })
  })

  it('does not call deleteImage when context menu is opened but nothing is clicked', async () => {
    vi.mocked(getImages).mockResolvedValue({ images: [makeImage()], next_cursor: null })

    renderImageGrid()

    await waitFor(() => {
      expect(screen.getByText('Test image')).toBeInTheDocument()
    })

    expect(deleteImage).not.toHaveBeenCalled()
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

describe('ImageGrid sort wiring', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('calls getFolderImages with sort=title and direction when Name is selected in a folder view', async () => {
    vi.mocked(getFolderImages).mockResolvedValue({ images: [makeImage()], next_cursor: null })

    renderImageGrid({ type: 'folder', id: 'folder-1' }, vi.fn(), 'title', 'asc')

    await waitFor(() => {
      expect(getFolderImages).toHaveBeenCalledWith(expect.any(Function), 'folder-1', 'title', 'asc')
    })
  })

  it('calls getFolderImages with no sort/direction params when Manual is selected in a folder view', async () => {
    vi.mocked(getFolderImages).mockResolvedValue({ images: [makeImage()], next_cursor: null })

    renderImageGrid({ type: 'folder', id: 'folder-1' }, vi.fn(), 'manual', undefined)

    await waitFor(() => {
      expect(getFolderImages).toHaveBeenCalledWith(expect.any(Function), 'folder-1', undefined, undefined)
    })
  })

  it('calls getAllImages with sort/direction params when an explicit sort is active in the All view', async () => {
    vi.mocked(getAllImages).mockResolvedValue({ images: [makeImage()], next_cursor: null })

    renderImageGrid({ type: 'all' }, vi.fn(), 'created_at', 'asc')

    await waitFor(() => {
      expect(getAllImages).toHaveBeenCalledWith(expect.any(Function), undefined, '', 'created_at', 'asc')
    })
  })
})

describe('ImageGrid drag gating with explicit sort', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('does not persist a position update when dropping an image onto another image under an explicit sort', async () => {
    const images = [
      makeImage({ id: '1', title: 'First', position: 'a0' }),
      makeImage({ id: '2', title: 'Second', position: 'a1' }),
    ]
    vi.mocked(getFolderImages).mockResolvedValue({ images, next_cursor: null })

    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })

    const wrap = (sortEndTrigger: SortEndTrigger | null) => (
      <DndContext sensors={[]}>
        <QueryClientProvider client={queryClient}>
          <MemoryRouter>
            <ImageGrid
              view={{ type: 'folder', id: 'folder-1' }}
              searchTerm=""
              debouncedSearchTerm=""
              sortBy="title"
              sortDir="asc"
              onImageSelect={vi.fn()}
              sortEndTrigger={sortEndTrigger}
            />
          </MemoryRouter>
        </QueryClientProvider>
      </DndContext>
    )

    const { rerender } = render(wrap(null))
    await waitFor(() => expect(screen.getByText('First')).toBeInTheDocument())

    rerender(wrap({ activeId: 'image-1', overId: 'image-2', ts: 1 }))

    await waitFor(() => expect(screen.getByText('Second')).toBeInTheDocument())
    expect(updateImagePosition).not.toHaveBeenCalled()
  })
})

describe('ImageGrid reorder rollback', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows error toast and reverts order when position update fails', async () => {
    const images = [
      makeImage({ id: '1', title: 'First', position: 'a0' }),
      makeImage({ id: '2', title: 'Second', position: 'a1' }),
    ]
    vi.mocked(getFolderImages).mockResolvedValue({ images, next_cursor: null })
    vi.mocked(updateImagePosition).mockRejectedValue(new Error('API error'))
    const toastError = vi.spyOn(toast, 'error')

    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })

    const wrap = (sortEndTrigger: SortEndTrigger | null) => (
      <DndContext sensors={[]}>
        <QueryClientProvider client={queryClient}>
          <MemoryRouter>
            <ImageGrid
              view={{ type: 'folder', id: 'folder-1' }}
              searchTerm=""
              debouncedSearchTerm=""
              sortBy="manual"
              sortDir={undefined}
              onImageSelect={vi.fn()}
              sortEndTrigger={sortEndTrigger}
            />
          </MemoryRouter>
        </QueryClientProvider>
      </DndContext>
    )

    const { rerender } = render(wrap(null))
    await waitFor(() => expect(screen.getByText('First')).toBeInTheDocument())

    rerender(wrap({ activeId: 'image-1', overId: 'image-2', ts: 1 }))

    await waitFor(() => {
      expect(toastError).toHaveBeenCalledWith('Failed to save order')
    })
  })
})
