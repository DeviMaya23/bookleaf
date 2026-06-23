import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import ImageLightbox from './ImageLightbox'
import type { Image } from '@/lib/images'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({ getToken: vi.fn().mockResolvedValue('token') }),
}))

vi.mock('@/lib/images', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/images')>()),
  getImage: vi.fn(),
}))

import { getImage } from '@/lib/images'

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

function renderLightbox(props: Partial<React.ComponentProps<typeof ImageLightbox>> = {}) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const image = props.image ?? makeImage()
  return render(
    <QueryClientProvider client={queryClient}>
      <ImageLightbox image={image} onClose={props.onClose ?? vi.fn()} />
    </QueryClientProvider>,
  )
}

describe('ImageLightbox', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows the thumbnail before the full-res image loads', async () => {
    vi.mocked(getImage).mockReturnValue(new Promise(() => {}))

    renderLightbox()

    expect(screen.getByTestId('lightbox-thumbnail')).toBeInTheDocument()
    expect(screen.queryByTestId('lightbox-full-image')).not.toBeInTheDocument()
  })

  it('shows the full-res image once the URL resolves', async () => {
    vi.mocked(getImage).mockResolvedValue({ ...makeImage(), image_url: 'https://example.com/full.jpg' } as never)

    renderLightbox()

    await waitFor(() => {
      expect(screen.getByTestId('lightbox-full-image')).toBeInTheDocument()
    })
  })

  it('renders no zoom, rotate, or flip controls', async () => {
    vi.mocked(getImage).mockResolvedValue({ ...makeImage(), image_url: 'https://example.com/full.jpg' } as never)

    renderLightbox()

    expect(screen.queryByLabelText(/rotate/i)).not.toBeInTheDocument()
    expect(screen.queryByLabelText(/flip/i)).not.toBeInTheDocument()
    expect(screen.queryByRole('slider')).not.toBeInTheDocument()
  })

  it('calls onClose when the close control is activated', async () => {
    vi.mocked(getImage).mockResolvedValue({ ...makeImage(), image_url: 'https://example.com/full.jpg' } as never)
    const onClose = vi.fn()

    renderLightbox({ onClose })

    await userEvent.click(screen.getByLabelText('Close'))

    expect(onClose).toHaveBeenCalledTimes(1)
  })
})
