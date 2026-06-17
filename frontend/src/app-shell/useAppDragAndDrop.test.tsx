import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { DragEndEvent } from '@dnd-kit/core'
import { toast } from 'sonner'
import type { Folder } from '@/lib/folders'
import { handleImageDrop, handleFolderDrop } from './lib/dragHandlers'
import { useAppDragAndDrop } from './useAppDragAndDrop'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({ getToken: vi.fn().mockResolvedValue('token') }),
}))

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

vi.mock('./lib/dragHandlers', async (importOriginal) => {
  const actual = await importOriginal<typeof import('./lib/dragHandlers')>()
  return {
    ...actual,
    handleImageDrop: vi.fn(),
    handleFolderDrop: vi.fn(),
  }
})

const folders: Folder[] = [
  { id: 'folder-1', name: 'Nature', description: null, icon: null, parent_id: null, created_at: '', updated_at: '' },
]

function dragEnd(active: { id: string; type: string; [k: string]: unknown }, over: { id: string; type: string; [k: string]: unknown } | null): DragEndEvent {
  return {
    active: { id: active.id, data: { current: active } },
    over: over ? { id: over.id, data: { current: over } } : null,
  } as unknown as DragEndEvent
}

function renderDnd() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries')
  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
  const { result } = renderHook(() => useAppDragAndDrop(folders), { wrapper })
  return { result, invalidateSpy }
}

describe('useAppDragAndDrop', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('dropping an image on another image sets a sort-end trigger and does not call move handlers', async () => {
    const { result } = renderDnd()

    await act(async () => {
      await result.current.handleDragEnd(dragEnd(
        { id: 'a', type: 'image', imageId: 'a', currentFolderId: null, thumbnailUrl: null },
        { id: 'b', type: 'image', imageId: 'b', currentFolderId: null, thumbnailUrl: null },
      ))
    })

    expect(result.current.sortEndTrigger).toEqual(
      expect.objectContaining({ activeId: 'a', overId: 'b' }),
    )
    expect(handleImageDrop).not.toHaveBeenCalled()
  })

  it('dropping an image on a folder moves it, invalidates images, and toasts success', async () => {
    vi.mocked(handleImageDrop).mockResolvedValue('moved')
    const { result, invalidateSpy } = renderDnd()

    await act(async () => {
      await result.current.handleDragEnd(dragEnd(
        { id: 'a', type: 'image', imageId: 'a', currentFolderId: null, thumbnailUrl: null },
        { id: 'folder-1', type: 'folder', folderId: 'folder-1' },
      ))
    })

    expect(handleImageDrop).toHaveBeenCalled()
    await waitFor(() => expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['images'] }))
    expect(toast.success).toHaveBeenCalledWith('Image moved')
  })

  it('toasts an error and re-invalidates when the image move fails', async () => {
    vi.mocked(handleImageDrop).mockRejectedValue(new Error('network'))
    const { result, invalidateSpy } = renderDnd()

    await act(async () => {
      await result.current.handleDragEnd(dragEnd(
        { id: 'a', type: 'image', imageId: 'a', currentFolderId: null, thumbnailUrl: null },
        { id: 'folder-1', type: 'folder', folderId: 'folder-1' },
      ))
    })

    expect(toast.error).toHaveBeenCalledWith('Failed to move image')
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['images'] })
  })

  it('dropping a folder moves it with the folder list, invalidates folders, and toasts success', async () => {
    vi.mocked(handleFolderDrop).mockResolvedValue('moved')
    const { result, invalidateSpy } = renderDnd()

    await act(async () => {
      await result.current.handleDragEnd(dragEnd(
        { id: 'folder-2', type: 'folder', folderId: 'folder-2', name: 'Folder 2', parentId: null },
        { id: 'folder-1', type: 'folder', folderId: 'folder-1' },
      ))
    })

    expect(handleFolderDrop).toHaveBeenCalledWith(expect.any(Function), expect.anything(), expect.anything(), folders)
    await waitFor(() => expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['folders'] }))
    expect(toast.success).toHaveBeenCalledWith('Folder moved')
  })

  it('does nothing when there is no drop target', async () => {
    const { result } = renderDnd()

    await act(async () => {
      await result.current.handleDragEnd(dragEnd({ id: 'a', type: 'image', imageId: 'a' }, null))
    })

    expect(result.current.sortEndTrigger).toBeNull()
    expect(handleImageDrop).not.toHaveBeenCalled()
  })
})
