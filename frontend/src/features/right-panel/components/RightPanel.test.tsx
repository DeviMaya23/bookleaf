import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { vi, describe, it, expect, beforeEach } from 'vitest'
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

import { downloadImage, updateImage } from '@/lib/images'
import { resolveOrCreateTags } from '@/lib/tags'
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

function renderPanel(image: Image, onClose = vi.fn()) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <RightPanel mode="image" image={image} onClose={onClose} />
    </QueryClientProvider>,
  )
}

describe('RightPanel — success scenario', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(updateImage).mockResolvedValue(makeImage())
  })

  it('renders title, notes, source URL, details, and download button', () => {
    renderPanel(makeImage())

    expect(screen.getByDisplayValue('Sunset photo')).toBeInTheDocument()
    expect(screen.getByDisplayValue('A nice sunset')).toBeInTheDocument()
    expect(screen.getByDisplayValue('https://example.com')).toBeInTheDocument()
    expect(screen.getByText('1920 × 1080')).toBeInTheDocument()
    expect(screen.getByText('2.0 MB')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /download image/i })).toBeInTheDocument()
  })

  it('calls downloadImage when the download button is clicked', async () => {
    vi.mocked(downloadImage).mockResolvedValue('https://r2.example.com/download?sig=abc')

    renderPanel(makeImage())

    await userEvent.click(screen.getByRole('button', { name: /download image/i }))

    await waitFor(() => {
      expect(downloadImage).toHaveBeenCalledWith(expect.any(Function), 'img-1')
    })
  })
})

describe('RightPanel — failure scenario', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('re-enables button and does not crash when downloadImage rejects', async () => {
    vi.mocked(downloadImage).mockRejectedValue(new Error('Network error'))

    renderPanel(makeImage())

    const button = screen.getByRole('button', { name: /download image/i })
    await userEvent.click(button)

    await waitFor(() => {
      expect(downloadImage).toHaveBeenCalledWith(expect.any(Function), 'img-1')
    })

    // Button returns to enabled state after error
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /download image/i })).not.toBeDisabled()
    })
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
    vi.mocked(getFolders).mockResolvedValue([{ id: 'folder-1', name: 'Nature', description: null, parent_id: null, created_at: '', updated_at: '' }])
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

describe('RightPanel folders — failure scenario', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getFolders).mockResolvedValue([{ id: 'folder-1', name: 'Nature', description: null, parent_id: null, created_at: '', updated_at: '' }])
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
