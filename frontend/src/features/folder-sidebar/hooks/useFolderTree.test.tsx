import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { getFolders, getFolderSubtreeIds } from '@/lib/folders'
import type { Folder } from '@/lib/folders'
import { useFolderTree } from './useFolderTree'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({ getToken: vi.fn().mockResolvedValue('token') }),
}))

vi.mock('@/lib/folders', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/folders')>()),
  getFolders: vi.fn(),
  getFolderSubtreeIds: vi.fn(),
}))

function makeFolder(id: string, name = `Folder ${id}`): Folder {
  return {
    id,
    name,
    description: null,
    icon: null,
    parent_id: null,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
  }
}

function renderUseFolderTree(activeDragFolderId: string | null = null) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
  const { result } = renderHook(() => useFolderTree(activeDragFolderId), { wrapper })
  return { result }
}

describe('useFolderTree', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('visibleTree contains all folders when folderFilter is empty', async () => {
    vi.mocked(getFolders).mockResolvedValue([makeFolder('a'), makeFolder('b')])

    const { result } = renderUseFolderTree()

    await waitFor(() => expect(result.current.isFoldersPending).toBe(false))
    expect(result.current.visibleTree).toHaveLength(2)
  })

  it('setFolderFilter narrows visibleTree to matching folders', async () => {
    vi.mocked(getFolders).mockResolvedValue([makeFolder('alpha', 'alpha'), makeFolder('beta', 'beta')])

    const { result } = renderUseFolderTree()

    await waitFor(() => expect(result.current.isFoldersPending).toBe(false))
    act(() => result.current.setFolderFilter('alph'))
    expect(result.current.visibleTree).toHaveLength(1)
    expect(result.current.visibleTree[0].id).toBe('alpha')
  })

  it('subtreeIds is empty when activeDragFolderId is null', async () => {
    vi.mocked(getFolders).mockResolvedValue([makeFolder('a')])

    const { result } = renderUseFolderTree(null)

    await waitFor(() => expect(result.current.isFoldersPending).toBe(false))
    expect(result.current.subtreeIds.size).toBe(0)
  })

  it('subtreeIds is populated from getFolderSubtreeIds when activeDragFolderId is set', async () => {
    vi.mocked(getFolders).mockResolvedValue([makeFolder('a')])
    vi.mocked(getFolderSubtreeIds).mockReturnValue(new Set(['a', 'b']))

    const { result } = renderUseFolderTree('a')

    await waitFor(() => expect(result.current.isFoldersPending).toBe(false))
    expect(result.current.subtreeIds).toEqual(new Set(['a', 'b']))
  })
})
