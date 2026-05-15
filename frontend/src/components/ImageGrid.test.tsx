import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import ImageGrid from './ImageGrid'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({ getToken: vi.fn().mockResolvedValue('token') }),
}))

vi.mock('@/lib/images', () => ({
  getImages: vi.fn(),
  deleteImage: vi.fn(),
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
    ContextMenuItem: ({
      children,
      onSelect,
      className,
    }: {
      children: React.ReactNode
      onSelect?: (e: Event) => void
      className?: string
    }) =>
      React.createElement(
        'button',
        { role: 'menuitem', className, onClick: () => onSelect?.(new Event('select')) },
        children,
      ),
  }
})

import { getImages, deleteImage } from '@/lib/images'

function makeImage(overrides?: object) {
  return {
    id: '1',
    title: 'Test image',
    description: null,
    mime_type: 'image/jpeg',
    source_url: null,
    folder_id: null,
    thumbnail_url: 'https://example.com/thumb.jpg',
    width: 100,
    height: 100,
    file_size: 1024,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    ...overrides,
  }
}

function renderImageGrid(folderId: string | null = null) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <ImageGrid folderId={folderId} />
      </MemoryRouter>
    </QueryClientProvider>,
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

describe('ImageGrid delete flow', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('calls deleteImage and invalidates on confirm', async () => {
    vi.mocked(getImages).mockResolvedValue({ images: [makeImage()], next_cursor: null })
    vi.mocked(deleteImage).mockResolvedValue(undefined)

    renderImageGrid()

    await waitFor(() => {
      expect(screen.getByText('Test image')).toBeInTheDocument()
    })

    await userEvent.click(screen.getByRole('menuitem', { name: /delete/i }))

    const confirmBtn = await screen.findByRole('button', { name: /^delete$/i })
    await userEvent.click(confirmBtn)

    await waitFor(() => {
      expect(deleteImage).toHaveBeenCalledWith(expect.any(Function), '1')
    })
  })

  it('does not call deleteImage when cancel is clicked', async () => {
    vi.mocked(getImages).mockResolvedValue({ images: [makeImage()], next_cursor: null })

    renderImageGrid()

    await waitFor(() => {
      expect(screen.getByText('Test image')).toBeInTheDocument()
    })

    await userEvent.click(screen.getByRole('menuitem', { name: /delete/i }))

    const cancelBtn = await screen.findByRole('button', { name: /cancel/i })
    await userEvent.click(cancelBtn)

    expect(deleteImage).not.toHaveBeenCalled()
  })
})
