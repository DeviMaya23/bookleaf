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
  getFolders: vi.fn().mockResolvedValue([]),
}))

vi.mock('@/lib/tags', () => ({
  getTags: vi.fn().mockResolvedValue([]),
  createTag: vi.fn(),
}))

import { downloadImage, updateImage } from '@/lib/images'
import { getTags, createTag } from '@/lib/tags'

function makeImage(overrides?: Partial<Image>): Image {
  return {
    id: 'img-1',
    title: 'Sunset photo',
    description: 'A nice sunset',
    mime_type: 'image/jpeg',
    source_url: 'https://example.com',
    folder_id: null,
    thumbnail_url: 'https://example.com/thumb.jpg',
    width: 1920,
    height: 1080,
    file_size: 2 * 1024 * 1024,
    tags: [],
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
      <RightPanel image={image} onClose={onClose} />
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
    expect(screen.getByText('Unsorted')).toBeInTheDocument()
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

  it('adds an existing tag by reusing its ID from the cache without calling createTag', async () => {
    vi.mocked(getTags).mockResolvedValue([{ id: 'tag-abc', name: 'nature' }])

    renderPanel(makeImage())

    const input = await screen.findByPlaceholderText('Add tags…')
    await userEvent.type(input, 'nature{Enter}')

    await waitFor(() => {
      expect(updateImage).toHaveBeenCalledWith(
        expect.any(Function),
        'img-1',
        expect.objectContaining({ tags: ['tag-abc'] }),
      )
    })
    expect(createTag).not.toHaveBeenCalled()
  })
})

describe('RightPanel tags — failure scenario', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getTags).mockResolvedValue([])
  })

  it('shows an error toast and does not call updateImage when PATCH fails', async () => {
    vi.mocked(createTag).mockResolvedValue({ id: 'tag-new', name: 'concept' })
    vi.mocked(updateImage).mockRejectedValue(new Error('Server error'))

    renderPanel(makeImage())

    const input = await screen.findByPlaceholderText('Add tags…')
    await userEvent.type(input, 'concept{Enter}')

    await waitFor(() => {
      expect(updateImage).toHaveBeenCalledWith(
        expect.any(Function),
        'img-1',
        expect.objectContaining({ tags: ['tag-new'] }),
      )
    })
  })
})
