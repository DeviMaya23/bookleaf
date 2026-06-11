import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { getImages, getFolderImages, getAllImages, getTrashedImages } from '@/lib/images'
import type { Image } from '@/lib/images'
import type { AppView } from '@/lib/view'
import { useGalleryImages, queryKeyFor, fetcherFor, EMPTY_FILTER } from './useGalleryImages'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({ getToken: vi.fn().mockResolvedValue('token') }),
}))

vi.mock('@/lib/images', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/images')>()),
  getImages: vi.fn(),
  getFolderImages: vi.fn(),
  getAllImages: vi.fn(),
  getTrashedImages: vi.fn(),
}))

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

describe('queryKeyFor', () => {
  it('includes a sort key for an explicit sort in the All view', () => {
    expect(queryKeyFor({ view: { type: 'all' }, debouncedSearch: '', sortBy: 'title', sortDir: 'asc', filterTagIds: EMPTY_FILTER, filterMimeTypes: EMPTY_FILTER, filterFolderIds: EMPTY_FILTER }))
      .toEqual(['images', 'all', '', ['title', 'asc'], EMPTY_FILTER, EMPTY_FILTER, EMPTY_FILTER])
  })

  it('omits the sort key when sortBy is manual', () => {
    expect(queryKeyFor({ view: { type: 'unsorted' }, debouncedSearch: '', sortBy: 'manual', sortDir: undefined, filterTagIds: EMPTY_FILTER, filterMimeTypes: EMPTY_FILTER, filterFolderIds: EMPTY_FILTER }))
      .toEqual(['images', 'unsorted', '', undefined, EMPTY_FILTER, EMPTY_FILTER])
  })

  it('keys the trash view without filters', () => {
    expect(queryKeyFor({ view: { type: 'trash' }, debouncedSearch: 'foo', sortBy: 'manual', sortDir: undefined, filterTagIds: EMPTY_FILTER, filterMimeTypes: EMPTY_FILTER, filterFolderIds: EMPTY_FILTER }))
      .toEqual(['images', 'trash', 'foo', undefined])
  })

  it('keys the folder view by folder id and sort, without search/filters', () => {
    expect(queryKeyFor({ view: { type: 'folder', id: 'folder-1' }, debouncedSearch: 'foo', sortBy: 'title', sortDir: 'desc', filterTagIds: ['tag-1'], filterMimeTypes: EMPTY_FILTER, filterFolderIds: EMPTY_FILTER }))
      .toEqual(['images', 'folder', 'folder-1', ['title', 'desc']])
  })
})

describe('fetcherFor', () => {
  const getToken = vi.fn().mockResolvedValue('token')

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('calls getAllImages with sort, direction, and filters for the All view', async () => {
    vi.mocked(getAllImages).mockResolvedValue({ images: [], next_cursor: null })

    const fetcher = fetcherFor({ view: { type: 'all' }, getToken, debouncedSearch: 'foo', sortBy: 'created_at', sortDir: 'asc', filterTagIds: ['tag-1'], filterMimeTypes: ['image/jpeg'], filterFolderIds: ['folder-1'] })
    await fetcher({ pageParam: undefined })

    expect(getAllImages).toHaveBeenCalledWith(getToken, undefined, 'foo', 'created_at', 'asc', ['tag-1'], ['image/jpeg'], ['folder-1'])
  })

  it('calls getImages with no folder ids for the Unsorted view', async () => {
    vi.mocked(getImages).mockResolvedValue({ images: [], next_cursor: null })

    const fetcher = fetcherFor({ view: { type: 'unsorted' }, getToken, debouncedSearch: 'foo', sortBy: 'manual', sortDir: undefined, filterTagIds: ['tag-1'], filterMimeTypes: EMPTY_FILTER, filterFolderIds: ['folder-1'] })
    await fetcher({ pageParam: 'cursor-1' })

    expect(getImages).toHaveBeenCalledWith(getToken, 'cursor-1', 'foo', undefined, undefined, ['tag-1'], EMPTY_FILTER)
  })

  it('calls getTrashedImages with only the search term for the Trash view', async () => {
    vi.mocked(getTrashedImages).mockResolvedValue({ images: [], next_cursor: null })

    const fetcher = fetcherFor({ view: { type: 'trash' }, getToken, debouncedSearch: 'foo', sortBy: 'title', sortDir: 'asc', filterTagIds: EMPTY_FILTER, filterMimeTypes: EMPTY_FILTER, filterFolderIds: EMPTY_FILTER })
    await fetcher({ pageParam: undefined })

    expect(getTrashedImages).toHaveBeenCalledWith(getToken, undefined, 'foo')
  })

  it('calls getFolderImages with the folder id, sort, and direction for the Folder view', async () => {
    vi.mocked(getFolderImages).mockResolvedValue({ images: [], next_cursor: null })

    const fetcher = fetcherFor({ view: { type: 'folder', id: 'folder-1' }, getToken, debouncedSearch: '', sortBy: 'title', sortDir: 'desc', filterTagIds: EMPTY_FILTER, filterMimeTypes: EMPTY_FILTER, filterFolderIds: EMPTY_FILTER })
    await fetcher({ pageParam: undefined })

    expect(getFolderImages).toHaveBeenCalledWith(getToken, 'folder-1', 'title', 'desc')
  })
})

function renderUseGalleryImages(view: AppView, props?: Partial<{
  searchTerm: string
  filterTagIds: string[]
  filterMimeTypes: string[]
  filterFolderIds: string[]
}>) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
  return renderHook(
    ({ searchTerm }: { searchTerm: string }) =>
      useGalleryImages({
        view,
        searchTerm,
        debouncedSearchTerm: searchTerm,
        sortBy: 'manual',
        sortDir: undefined,
        filterTagIds: props?.filterTagIds ?? EMPTY_FILTER,
        filterMimeTypes: props?.filterMimeTypes ?? EMPTY_FILTER,
        filterFolderIds: props?.filterFolderIds ?? EMPTY_FILTER,
      }),
    { wrapper, initialProps: { searchTerm: props?.searchTerm ?? '' } },
  )
}

describe('useGalleryImages', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('filters folder-view images client-side by tag with no additional fetch', async () => {
    const images = [
      makeImage({ id: '1', title: 'First', tags: [{ id: 'tag-1', name: 'Cats' }] }),
      makeImage({ id: '2', title: 'Second', tags: [{ id: 'tag-2', name: 'Dogs' }] }),
    ]
    vi.mocked(getFolderImages).mockResolvedValue({ images, next_cursor: null })

    const { result } = renderUseGalleryImages({ type: 'folder', id: 'folder-1' }, { filterTagIds: ['tag-1'] })

    await waitFor(() => expect(result.current.images.map((i) => i.id)).toEqual(['1']))
    expect(getFolderImages).toHaveBeenCalledTimes(1)
  })

  it('combines an active search term with a tag filter in folder view', async () => {
    const images = [
      makeImage({ id: '1', title: 'Cat photo', tags: [{ id: 'tag-1', name: 'Cats' }] }),
      makeImage({ id: '2', title: 'Cat drawing', tags: [{ id: 'tag-2', name: 'Dogs' }] }),
      makeImage({ id: '3', title: 'Dog photo', tags: [{ id: 'tag-1', name: 'Cats' }] }),
    ]
    vi.mocked(getFolderImages).mockResolvedValue({ images, next_cursor: null })

    const { result } = renderUseGalleryImages({ type: 'folder', id: 'folder-1' }, { filterTagIds: ['tag-1'], searchTerm: 'cat' })

    await waitFor(() => expect(result.current.images.map((i) => i.id)).toEqual(['1']))
  })

  it('returns a referentially stable images array across re-renders when inputs are unchanged', async () => {
    vi.mocked(getImages).mockResolvedValue({ images: [makeImage()], next_cursor: null })

    const { result, rerender } = renderUseGalleryImages({ type: 'unsorted' })

    await waitFor(() => expect(result.current.isLoading).toBe(false))
    const firstImages = result.current.images

    rerender({ searchTerm: '' })

    expect(result.current.images).toBe(firstImages)
  })
})
