import { renderHook, act, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { createElement } from 'react'
import { useSSEEvents } from './useSSEEvents'

type SSEOptions = {
  onmessage?: (msg: { data: string }) => void
  onclose?: () => void
  onerror?: (err: unknown) => void
}

let capturedOnMessage: ((msg: { data: string }) => void) | undefined

vi.mock('@microsoft/fetch-event-source', () => ({
  fetchEventSource: (_url: string, opts: SSEOptions) => {
    capturedOnMessage = opts.onmessage
  },
}))

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({ getToken: vi.fn() }),
}))

const mockGetMe = vi.fn()

vi.mock('@/features/auth/lib/me', () => ({
  getMe: () => mockGetMe(),
}))

function makeWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return {
    queryClient,
    wrapper: ({ children }: { children: React.ReactNode }) =>
      createElement(QueryClientProvider, { client: queryClient }, children),
  }
}

describe('useSSEEvents', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    capturedOnMessage = undefined
  })

  it('invalidates folders, images, image, folder, and me caches on categorisation_complete', async () => {
    mockGetMe.mockResolvedValue({
      id: 'kp_1',
      ai_categorisation_enabled: true,
      vision_enabled: true,
      ai_categorisation_count_this_month: 3,
    })

    const { queryClient, wrapper } = makeWrapper()
    const invalidate = vi.spyOn(queryClient, 'invalidateQueries')

    renderHook(() => useSSEEvents(), { wrapper })

    // Wait for the me query to resolve so the effect fires and fetchEventSource is called
    await waitFor(() => expect(capturedOnMessage).toBeDefined())

    act(() => {
      capturedOnMessage!({
        data: JSON.stringify({
          type: 'categorisation_complete',
          payload: { image_id: 'img-abc' },
        }),
      })
    })

    expect(invalidate).toHaveBeenCalledWith({ queryKey: ['folders'] })
    expect(invalidate).toHaveBeenCalledWith({ queryKey: ['images'] })
    expect(invalidate).toHaveBeenCalledWith({ queryKey: ['image', 'img-abc'] })
    expect(invalidate).toHaveBeenCalledWith({ queryKey: ['folder'], exact: false })
    expect(invalidate).toHaveBeenCalledWith({ queryKey: ['me'] })
  })
})
