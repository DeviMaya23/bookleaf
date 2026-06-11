import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { toast } from 'sonner'
import { deleteImage, hardDeleteImage, restoreImage } from '@/lib/images'
import type { Image } from '@/lib/images'
import { useImageLifecycle } from './useImageLifecycle'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({ getToken: vi.fn().mockResolvedValue('token') }),
}))

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

vi.mock('@/lib/images', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/images')>()),
  deleteImage: vi.fn(),
  hardDeleteImage: vi.fn(),
  restoreImage: vi.fn(),
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

function renderUseImageLifecycle(isTrash: boolean, removeImage = vi.fn(), onImageDeleted = vi.fn()) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries')
  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
  const { result } = renderHook(() => useImageLifecycle(isTrash, removeImage, onImageDeleted), { wrapper })
  return { result, removeImage, onImageDeleted, invalidateSpy }
}

describe('useImageLifecycle', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('handleAction calls deleteImage when not in trash', async () => {
    vi.mocked(deleteImage).mockResolvedValue(undefined)
    const { result } = renderUseImageLifecycle(false)

    act(() => result.current.handleAction(makeImage()))

    await waitFor(() => expect(deleteImage).toHaveBeenCalledWith(expect.any(Function), '1'))
    expect(restoreImage).not.toHaveBeenCalled()
  })

  it('handleAction calls restoreImage when in trash', async () => {
    vi.mocked(restoreImage).mockResolvedValue(makeImage())
    const { result } = renderUseImageLifecycle(true)

    act(() => result.current.handleAction(makeImage()))

    await waitFor(() => expect(restoreImage).toHaveBeenCalledWith(expect.any(Function), '1'))
    expect(deleteImage).not.toHaveBeenCalled()
  })

  it('on delete success, removes the image, invalidates trash, toasts, and notifies onImageDeleted', async () => {
    vi.mocked(deleteImage).mockResolvedValue(undefined)
    const { result, removeImage, onImageDeleted, invalidateSpy } = renderUseImageLifecycle(false)

    act(() => result.current.handleAction(makeImage()))

    await waitFor(() => expect(removeImage).toHaveBeenCalledWith('1'))
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['images', 'trash'] })
    expect(toast.success).toHaveBeenCalledWith('Image moved to trash')
    expect(onImageDeleted).toHaveBeenCalledWith('1')
  })

  it('on delete error, shows an error toast and does not remove the image', async () => {
    vi.mocked(deleteImage).mockRejectedValue(new Error('boom'))
    const { result, removeImage } = renderUseImageLifecycle(false)

    act(() => result.current.handleAction(makeImage()))

    await waitFor(() => expect(toast.error).toHaveBeenCalledWith('Failed to delete image'))
    expect(removeImage).not.toHaveBeenCalled()
  })

  it('on restore success, toasts and does not call removeImage (D4)', async () => {
    vi.mocked(restoreImage).mockResolvedValue(makeImage())
    const { result, removeImage, invalidateSpy } = renderUseImageLifecycle(true)

    act(() => result.current.handleAction(makeImage()))

    await waitFor(() => expect(toast.success).toHaveBeenCalledWith('Image restored'))
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['images', 'trash'] })
    expect(removeImage).not.toHaveBeenCalled()
  })

  it('on restore error, shows an error toast', async () => {
    vi.mocked(restoreImage).mockRejectedValue(new Error('boom'))
    const { result } = renderUseImageLifecycle(true)

    act(() => result.current.handleAction(makeImage()))

    await waitFor(() => expect(toast.error).toHaveBeenCalledWith('Failed to restore image'))
  })

  it('openDeleteConfirm sets confirmDeleteImage and closeDeleteConfirm clears it without calling hardDeleteImage', () => {
    const { result } = renderUseImageLifecycle(true)
    const image = makeImage()

    act(() => result.current.openDeleteConfirm(image))
    expect(result.current.confirmDeleteImage).toEqual(image)

    act(() => result.current.closeDeleteConfirm())
    expect(result.current.confirmDeleteImage).toBeNull()
    expect(hardDeleteImage).not.toHaveBeenCalled()
  })

  it('confirmDelete calls hardDeleteImage, removes the image, clears the dialog, and notifies onImageDeleted', async () => {
    vi.mocked(hardDeleteImage).mockResolvedValue(undefined)
    const { result, removeImage, onImageDeleted, invalidateSpy } = renderUseImageLifecycle(true)
    const image = makeImage()

    act(() => result.current.openDeleteConfirm(image))
    act(() => result.current.confirmDelete())

    expect(result.current.confirmDeleteImage).toBeNull()
    await waitFor(() => expect(hardDeleteImage).toHaveBeenCalledWith(expect.any(Function), '1'))
    expect(removeImage).toHaveBeenCalledWith('1')
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['images', 'trash'] })
    expect(toast.success).toHaveBeenCalledWith('Image permanently deleted')
    expect(onImageDeleted).toHaveBeenCalledWith('1')
  })

  it('on hard-delete error, shows an error toast', async () => {
    vi.mocked(hardDeleteImage).mockRejectedValue(new Error('boom'))
    const { result } = renderUseImageLifecycle(true)
    const image = makeImage()

    act(() => result.current.openDeleteConfirm(image))
    act(() => result.current.confirmDelete())

    await waitFor(() => expect(toast.error).toHaveBeenCalledWith('Failed to permanently delete image'))
  })
})
