import { renderHook, act, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import { useSelectionMode } from './useSelectionMode'
import { bulkTrashImages } from '@/lib/images'
import { toast } from 'sonner'
import type { BulkActionResult } from '@/lib/images'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({ getToken: vi.fn().mockResolvedValue('token') }),
}))

vi.mock('@/lib/images', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/images')>()),
  bulkTrashImages: vi.fn(),
}))

vi.mock('@/lib/tags', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/tags')>()),
}))

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
}))

function makeWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return {
    queryClient,
    wrapper: ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    ),
  }
}

describe('useSelectionMode', () => {
  beforeEach(() => { vi.clearAllMocks() })

  it('handleSelectModeToggle entering select mode calls clearImageSelection', () => {
    const clearImageSelection = vi.fn()
    const { wrapper } = makeWrapper()
    const { result } = renderHook(
      () => useSelectionMode({ viewKey: 'all', clearImageSelection }),
      { wrapper },
    )
    act(() => result.current.handleSelectModeToggle(true))
    expect(clearImageSelection).toHaveBeenCalled()
    expect(result.current.selectMode).toBe(true)
  })

  it('handleSelectModeToggle exiting select mode resets selectedIds and mainSelectedId', () => {
    const { wrapper } = makeWrapper()
    const { result } = renderHook(
      () => useSelectionMode({ viewKey: 'all', clearImageSelection: vi.fn() }),
      { wrapper },
    )
    act(() => result.current.handleSelectModeToggle(true))
    act(() => result.current.handleSelectionChange(new Set(['img-1']), 'img-1'))
    act(() => result.current.handleSelectModeToggle(false))
    expect(result.current.selectMode).toBe(false)
    expect(result.current.selectedIds.size).toBe(0)
    expect(result.current.mainSelectedId).toBeNull()
  })

  it('exitSelectMode resets selectMode, selectedIds, and mainSelectedId', () => {
    const { wrapper } = makeWrapper()
    const { result } = renderHook(
      () => useSelectionMode({ viewKey: 'all', clearImageSelection: vi.fn() }),
      { wrapper },
    )
    act(() => result.current.handleSelectModeToggle(true))
    act(() => result.current.handleSelectionChange(new Set(['img-1']), 'img-1'))
    act(() => result.current.exitSelectMode())
    expect(result.current.selectMode).toBe(false)
    expect(result.current.selectedIds.size).toBe(0)
    expect(result.current.mainSelectedId).toBeNull()
  })

  it('resets all selection state on viewKey change', () => {
    const { wrapper } = makeWrapper()
    const { result, rerender } = renderHook(
      ({ vk }) => useSelectionMode({ viewKey: vk, clearImageSelection: vi.fn() }),
      { wrapper, initialProps: { vk: 'all' } },
    )
    act(() => result.current.handleSelectModeToggle(true))
    act(() => result.current.handleSelectionChange(new Set(['img-1']), 'img-1'))
    rerender({ vk: 'folder:abc' })
    expect(result.current.selectMode).toBe(false)
    expect(result.current.selectedIds.size).toBe(0)
    expect(result.current.mainSelectedId).toBeNull()
  })

  it('handleMoveSelectionToTrash calls bulkTrashImages with selected ids and exits select mode on success', async () => {
    vi.mocked(bulkTrashImages).mockResolvedValue({ succeeded_count: 2 } as BulkActionResult)
    const { wrapper } = makeWrapper()
    const { result } = renderHook(
      () => useSelectionMode({ viewKey: 'all', clearImageSelection: vi.fn() }),
      { wrapper },
    )
    act(() => result.current.handleSelectModeToggle(true))
    act(() => result.current.handleSelectionChange(new Set(['img-1', 'img-2']), 'img-1'))
    act(() => result.current.handleMoveSelectionToTrash())
    await waitFor(() => expect(result.current.selectMode).toBe(false))
    expect(bulkTrashImages).toHaveBeenCalledWith(expect.any(Function), ['img-1', 'img-2'])
    expect(toast.success).toHaveBeenCalledWith('Moved 2 images to trash')
    expect(result.current.selectedIds.size).toBe(0)
  })

  it('handleMoveSelectionToTrash shows warning when some images fail to trash', async () => {
    vi.mocked(bulkTrashImages).mockResolvedValue({ succeeded_count: 1 } as BulkActionResult)
    const { wrapper } = makeWrapper()
    const { result } = renderHook(
      () => useSelectionMode({ viewKey: 'all', clearImageSelection: vi.fn() }),
      { wrapper },
    )
    act(() => result.current.handleSelectionChange(new Set(['img-1', 'img-2']), 'img-1'))
    act(() => result.current.handleMoveSelectionToTrash())
    await waitFor(() => expect(toast.warning).toHaveBeenCalled())
    expect(toast.warning).toHaveBeenCalledWith('Moved 1 of 2 images to trash')
  })
})
