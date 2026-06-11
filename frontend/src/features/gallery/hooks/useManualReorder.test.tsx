import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useDndMonitor } from '@dnd-kit/core'
import { toast } from 'sonner'
import { updateImagePosition } from '@/lib/images'
import type { Image } from '@/lib/images'
import type { AppView } from '@/lib/view'
import type { SortEndTrigger } from '../components/ImageGrid'
import { useManualReorder } from './useManualReorder'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({ getToken: vi.fn().mockResolvedValue('token') }),
}))

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

vi.mock('@dnd-kit/core', () => ({
  useDndMonitor: vi.fn(),
}))

vi.mock('@/lib/images', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/images')>()),
  updateImagePosition: vi.fn(),
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

const folderView: AppView = { type: 'folder', id: 'folder-1' }

function renderUseManualReorder(images: Image[], sortEndTrigger: SortEndTrigger | null = null, sortBy: 'manual' | 'created_at' | 'title' = 'manual') {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
  return renderHook(
    ({ images, sortEndTrigger }: { images: Image[]; sortEndTrigger: SortEndTrigger | null }) =>
      useManualReorder(folderView, sortBy, true, images, sortEndTrigger),
    { wrapper, initialProps: { images, sortEndTrigger } },
  )
}

describe('useManualReorder', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('syncs orderedImages from the images input', async () => {
    const images = [makeImage({ id: 'a' }), makeImage({ id: 'b' })]
    const { result } = renderUseManualReorder(images)

    await waitFor(() => expect(result.current.orderedImages).toEqual(images))
    expect(result.current.sortableItems).toEqual(['image-a', 'image-b'])
  })

  it('reorders on a sortEndTrigger and persists the new position', async () => {
    vi.mocked(updateImagePosition).mockResolvedValue(undefined)
    const images = [
      makeImage({ id: 'a', position: 'a0' }),
      makeImage({ id: 'b', position: 'a1' }),
    ]
    const { result, rerender } = renderUseManualReorder(images)
    await waitFor(() => expect(result.current.orderedImages).toEqual(images))

    rerender({ images, sortEndTrigger: { activeId: 'image-a', overId: 'image-b', ts: 1 } })

    await waitFor(() => expect(updateImagePosition).toHaveBeenCalled())
    expect(updateImagePosition).toHaveBeenCalledWith(expect.any(Function), 'a', 'folder-1', expect.any(String))
    expect(result.current.orderedImages.map((i) => i.id)).toEqual(['b', 'a'])
  })

  it('does not persist a position update when sortBy is not manual', async () => {
    const images = [
      makeImage({ id: 'a', position: 'a0' }),
      makeImage({ id: 'b', position: 'a1' }),
    ]
    const { result, rerender } = renderUseManualReorder(images, null, 'title')
    await waitFor(() => expect(result.current.orderedImages).toEqual(images))

    rerender({ images, sortEndTrigger: { activeId: 'image-a', overId: 'image-b', ts: 1 } })

    await waitFor(() => expect(result.current.orderedImages).toEqual(images))
    expect(updateImagePosition).not.toHaveBeenCalled()
  })

  it('rolls back the order and shows an error toast when persisting fails', async () => {
    vi.mocked(updateImagePosition).mockRejectedValue(new Error('boom'))
    const images = [
      makeImage({ id: 'a', position: 'a0' }),
      makeImage({ id: 'b', position: 'a1' }),
    ]
    const { result, rerender } = renderUseManualReorder(images)
    await waitFor(() => expect(result.current.orderedImages).toEqual(images))

    rerender({ images, sortEndTrigger: { activeId: 'image-a', overId: 'image-b', ts: 1 } })

    await waitFor(() => expect(toast.error).toHaveBeenCalledWith('Failed to save order'))
    expect(result.current.orderedImages).toEqual(images)
  })

  it('ignores a sortEndTrigger whose timestamp was already processed', async () => {
    vi.mocked(updateImagePosition).mockResolvedValue(undefined)
    const images = [
      makeImage({ id: 'a', position: 'a0' }),
      makeImage({ id: 'b', position: 'a1' }),
    ]
    const { result, rerender } = renderUseManualReorder(images)
    await waitFor(() => expect(result.current.orderedImages).toEqual(images))

    rerender({ images, sortEndTrigger: { activeId: 'image-a', overId: 'image-b', ts: 1 } })
    await waitFor(() => expect(updateImagePosition).toHaveBeenCalledTimes(1))

    // Same ts, different object reference - should be ignored.
    rerender({ images, sortEndTrigger: { activeId: 'image-a', overId: 'image-b', ts: 1 } })

    expect(updateImagePosition).toHaveBeenCalledTimes(1)
  })

  it('removeImage filters the image out of orderedImages', async () => {
    const images = [makeImage({ id: 'a' }), makeImage({ id: 'b' })]
    const { result } = renderUseManualReorder(images)
    await waitFor(() => expect(result.current.orderedImages).toEqual(images))

    act(() => result.current.removeImage('a'))

    expect(result.current.orderedImages.map((i) => i.id)).toEqual(['b'])
  })

  it('updates dragOverId via the useDndMonitor handlers', async () => {
    const images = [makeImage({ id: 'a' }), makeImage({ id: 'b' })]
    const { result } = renderUseManualReorder(images)
    await waitFor(() => expect(result.current.orderedImages).toEqual(images))

    const listener = vi.mocked(useDndMonitor).mock.calls[0][0]

    act(() => listener.onDragStart?.({ active: { id: 'image-a' } } as Parameters<NonNullable<typeof listener.onDragStart>>[0]))
    act(() => listener.onDragOver?.({
      active: { id: 'image-a', data: { current: { type: 'image' } } },
      over: { id: 'image-b' },
    } as unknown as Parameters<NonNullable<typeof listener.onDragOver>>[0]))

    expect(result.current.dragOverId).toBe('image-b')

    act(() => listener.onDragEnd?.({} as Parameters<NonNullable<typeof listener.onDragEnd>>[0]))

    expect(result.current.dragOverId).toBeNull()
  })
})
