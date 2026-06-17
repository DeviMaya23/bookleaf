import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { toast } from 'sonner'
import { createFolder, deleteFolder } from '@/lib/folders'
import type { Folder } from '@/lib/folders'
import { useFolderMutations } from './useFolderMutations'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({ getToken: vi.fn().mockResolvedValue('token') }),
}))

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

vi.mock('@/lib/folders', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/folders')>()),
  createFolder: vi.fn(),
  updateFolder: vi.fn(),
  deleteFolder: vi.fn(),
}))

function makeFolder(id: string): Folder {
  return {
    id,
    name: `Folder ${id}`,
    description: null, icon: null,
    parent_id: null,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
  }
}

function renderUseFolderMutations() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries')
  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
  const { result } = renderHook(() => useFolderMutations(), { wrapper })
  return { result, invalidateSpy }
}

describe('useFolderMutations', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('createMutation creates a folder and invalidates the folders query', async () => {
    const created = makeFolder('new')
    vi.mocked(createFolder).mockResolvedValue(created)

    const { result, invalidateSpy } = renderUseFolderMutations()
    result.current.createMutation.mutate({ name: 'Photos', parentId: 'parent-1' })

    await waitFor(() => expect(result.current.createMutation.isSuccess).toBe(true))
    expect(createFolder).toHaveBeenCalledWith(expect.any(Function), 'Photos', 'parent-1')
    expect(result.current.createMutation.data).toEqual(created)
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['folders'] })
  })

  it('deleteMutation invalidates folders and shows a success toast on success', async () => {
    vi.mocked(deleteFolder).mockResolvedValue(undefined)

    const { result, invalidateSpy } = renderUseFolderMutations()
    result.current.deleteMutation.mutate('folder-1')

    await waitFor(() => expect(result.current.deleteMutation.isSuccess).toBe(true))
    expect(deleteFolder).toHaveBeenCalledWith(expect.any(Function), 'folder-1')
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['folders'] })
    expect(toast.success).toHaveBeenCalledWith('Folder deleted')
  })

  it('deleteMutation shows an error toast when deletion fails', async () => {
    vi.mocked(deleteFolder).mockRejectedValue(new Error('boom'))

    const { result } = renderUseFolderMutations()
    result.current.deleteMutation.mutate('folder-1')

    await waitFor(() => expect(result.current.deleteMutation.isError).toBe(true))
    expect(toast.error).toHaveBeenCalledWith('Failed to delete folder')
  })
})
