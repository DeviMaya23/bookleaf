import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import ImageViewer from './ImageViewer'
import type { Image } from '@/lib/images'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({ getToken: vi.fn().mockResolvedValue('token') }),
}))

vi.mock('@/lib/images', () => ({
  getImage: vi.fn(),
}))

import { getImage } from '@/lib/images'

function makeImage(overrides?: Partial<Image>): Image {
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

function renderViewer(image: Image, onClose = vi.fn()) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ImageViewer image={image} onClose={onClose} />
    </QueryClientProvider>,
  )
}

describe('ImageViewer — thumbnail placeholder', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows thumbnail_url while full-res URL is loading', () => {
    vi.mocked(getImage).mockReturnValue(new Promise(() => {}))

    renderViewer(makeImage())

    expect(screen.getByRole('img')).toHaveAttribute('src', 'https://example.com/thumb.jpg')
  })
})

describe('ImageViewer — full-res image', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows image_url once getImage resolves', async () => {
    vi.mocked(getImage).mockResolvedValue({
      ...makeImage(),
      image_url: 'https://r2.example.com/full.jpg',
      suggested_folder_name: null,
    })

    renderViewer(makeImage())

    await waitFor(() => {
      expect(screen.getByRole('img')).toHaveAttribute('src', 'https://r2.example.com/full.jpg')
    })
  })
})

describe('ImageViewer — Esc key', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getImage).mockReturnValue(new Promise(() => {}))
  })

  it('calls onClose when Esc is pressed', async () => {
    const onClose = vi.fn()
    renderViewer(makeImage(), onClose)

    await userEvent.keyboard('{Escape}')

    expect(onClose).toHaveBeenCalledOnce()
  })
})

describe('ImageViewer — back button', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getImage).mockReturnValue(new Promise(() => {}))
  })

  it('calls onClose when the back button is clicked', async () => {
    const onClose = vi.fn()
    renderViewer(makeImage(), onClose)

    await userEvent.click(screen.getByRole('button', { name: /back/i }))

    expect(onClose).toHaveBeenCalledOnce()
  })
})

describe('ImageViewer — badge', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getImage).mockReturnValue(new Promise(() => {}))
  })

  it('renders the image title and dimensions in the badge', () => {
    renderViewer(makeImage())

    expect(screen.getByText('Sunset photo · 1920 × 1080')).toBeInTheDocument()
  })
})
