import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { getImage } from '@/lib/images'
import { getFolders } from '@/lib/folders'
import { getTags } from '@/lib/tags'
import type { Image } from '@/lib/images'
import type { Folder } from '@/lib/folders'
import { useImageDetailsData } from './useImageDetailsData'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({ getToken: vi.fn().mockResolvedValue('token') }),
}))

vi.mock('@/lib/images', () => ({ getImage: vi.fn() }))
vi.mock('@/lib/folders', () => ({ getFolders: vi.fn() }))
vi.mock('@/lib/tags', () => ({ getTags: vi.fn() }))

function makeFolder(id: string, name: string): Folder {
  return { id, name, description: null, parent_id: null, created_at: '', updated_at: '' }
}

function makeImage(overrides?: Partial<Image>): Image {
  return {
    id: 'img-1',
    title: 'Title',
    description: null,
    mime_type: 'image/jpeg',
    source_url: null,
    folder_ids: [],
    thumbnail_url: 'https://example.com/t.jpg',
    width: 10,
    height: 10,
    file_size: 100,
    tags: [],
    position: null,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides,
  }
}

function renderData(image: Image) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
  return renderHook(({ img }) => useImageDetailsData(img), {
    wrapper,
    initialProps: { img: image },
  })
}

describe('useImageDetailsData', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getImage).mockResolvedValue({
      ...makeImage(),
      image_url: 'https://example.com/full.jpg',
      suggested_folder_name: null,
    })
    vi.mocked(getTags).mockResolvedValue([{ id: 'tag-1', name: 'nature' }])
    vi.mocked(getFolders).mockResolvedValue([makeFolder('f1', 'Nature'), makeFolder('f2', 'Travel')])
  })

  it('resolves selectedFolders from the image folder_ids once folders load', async () => {
    const { result } = renderData(makeImage({ folder_ids: ['f2'] }))

    await waitFor(() => {
      expect(result.current.selectedFolders).toEqual([makeFolder('f2', 'Travel')])
    })
    expect(result.current.allTags).toEqual([{ id: 'tag-1', name: 'nature' }])
  })

  it('re-syncs selectedFolders when the image changes', async () => {
    const { result, rerender } = renderData(makeImage({ folder_ids: ['f1'] }))

    await waitFor(() => {
      expect(result.current.selectedFolders).toEqual([makeFolder('f1', 'Nature')])
    })

    rerender({ img: makeImage({ id: 'img-2', folder_ids: ['f2'] }) })

    await waitFor(() => {
      expect(result.current.selectedFolders).toEqual([makeFolder('f2', 'Travel')])
    })
  })
})
