import { describe, it, expect, vi } from 'vitest'
import { act, renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useRenameMutation, useDeleteMutation } from './useTagMutations'
import type { Tag } from '@/lib/tags'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({ getToken: vi.fn().mockResolvedValue('token') }),
}))

vi.mock('@/lib/tags', () => ({
  updateTag: vi.fn(),
  deleteTag: vi.fn(),
}))

import { updateTag, deleteTag } from '@/lib/tags'

function makeWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

function makeTag(id: string, name: string): Tag {
  return { id, name }
}

describe('useRenameMutation', () => {
  it('invalidates tags cache and calls onSuccess on successful rename', async () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } })
    vi.mocked(updateTag).mockResolvedValueOnce(makeTag('t-1', 'New Name'))
    const invalidate = vi.spyOn(queryClient, 'invalidateQueries')

    const onSuccess = vi.fn()
    const { result } = renderHook(() => useRenameMutation(onSuccess), { wrapper: makeWrapper(queryClient) })

    act(() => { result.current.mutate('t-1', 'New Name') })

    await waitFor(() => expect(onSuccess).toHaveBeenCalled())
    expect(invalidate).toHaveBeenCalledWith({ queryKey: ['tags'] })
    expect(result.current.duplicateError).toBe(false)
  })

  it('surfaces duplicateError and does not call onSuccess when updateTag returns null (409)', async () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } })
    vi.mocked(updateTag).mockResolvedValueOnce(null)

    const onSuccess = vi.fn()
    const { result } = renderHook(() => useRenameMutation(onSuccess), { wrapper: makeWrapper(queryClient) })

    act(() => { result.current.mutate('t-1', 'Duplicate') })

    await waitFor(() => expect(result.current.duplicateError).toBe(true))
    expect(onSuccess).not.toHaveBeenCalled()
  })
})

describe('useDeleteMutation', () => {
  it('invalidates tags cache and calls onSuccess on successful delete', async () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } })
    vi.mocked(deleteTag).mockResolvedValueOnce(undefined)
    const invalidate = vi.spyOn(queryClient, 'invalidateQueries')

    const onSuccess = vi.fn()
    const { result } = renderHook(() => useDeleteMutation(onSuccess), { wrapper: makeWrapper(queryClient) })

    act(() => { result.current.mutate('t-1') })

    await waitFor(() => expect(onSuccess).toHaveBeenCalled())
    expect(invalidate).toHaveBeenCalledWith({ queryKey: ['tags'] })
  })
})
