import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import type { AppView } from '@/lib/view'
import type { Folder } from '@/lib/folders'
import { useFolderSidebarDialogs } from './useFolderSidebarDialogs'

const mockNavigate = vi.fn()
vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}))

const mockCreateMutate = vi.fn()
const mockRenameMutate = vi.fn()
const mockDeleteMutate = vi.fn()

vi.mock('./useFolderMutations', () => ({
  useFolderMutations: () => ({
    createMutation: { mutate: mockCreateMutate },
    renameMutation: { mutate: mockRenameMutate },
    deleteMutation: { mutate: mockDeleteMutate },
    changeIconMutation: { mutate: vi.fn() },
  }),
}))

function makeFolder(id: string): Folder {
  return {
    id,
    name: `Folder ${id}`,
    description: null,
    icon: null,
    parent_id: null,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
  }
}

const allView: AppView = { type: 'all' }

describe('useFolderSidebarDialogs', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('handleNameDialogSubmit with create-root calls createMutation with name only', () => {
    const { result } = renderHook(() => useFolderSidebarDialogs(allView))

    act(() => result.current.setNameDialog({ mode: 'create-root' }))
    act(() => result.current.handleNameDialogSubmit('Photos'))

    expect(mockCreateMutate).toHaveBeenCalledWith({ name: 'Photos' })
  })

  it('handleNameDialogSubmit with create-sub calls createMutation with name and parentId', () => {
    const parent = makeFolder('parent-1')
    const { result } = renderHook(() => useFolderSidebarDialogs(allView))

    act(() => result.current.setNameDialog({ mode: 'create-sub', parent }))
    act(() => result.current.handleNameDialogSubmit('Sub'))

    expect(mockCreateMutate).toHaveBeenCalledWith({ name: 'Sub', parentId: 'parent-1' })
  })

  it('handleNameDialogSubmit with rename calls renameMutation with id and name', () => {
    const target = makeFolder('folder-1')
    const { result } = renderHook(() => useFolderSidebarDialogs(allView))

    act(() => result.current.setNameDialog({ mode: 'rename', target }))
    act(() => result.current.handleNameDialogSubmit('New Name'))

    expect(mockRenameMutate).toHaveBeenCalledWith({ id: 'folder-1', name: 'New Name' })
  })

  it('handleDelete navigates to /app when the deleted folder is currently viewed', () => {
    const folder = makeFolder('folder-1')
    const folderView: AppView = { type: 'folder', id: 'folder-1' }
    const { result } = renderHook(() => useFolderSidebarDialogs(folderView))

    act(() => result.current.setDeleteTarget(folder))
    act(() => result.current.handleDelete())

    const onSuccess = vi.mocked(mockDeleteMutate).mock.calls[0][1]?.onSuccess
    act(() => onSuccess?.())
    expect(mockNavigate).toHaveBeenCalledWith('/app')
  })

  it('handleDelete does not navigate when a different folder is viewed', () => {
    const folder = makeFolder('folder-1')
    const otherView: AppView = { type: 'folder', id: 'other' }
    const { result } = renderHook(() => useFolderSidebarDialogs(otherView))

    act(() => result.current.setDeleteTarget(folder))
    act(() => result.current.handleDelete())

    const onSuccess = vi.mocked(mockDeleteMutate).mock.calls[0][1]?.onSuccess
    act(() => onSuccess?.())
    expect(mockNavigate).not.toHaveBeenCalled()
  })
})
