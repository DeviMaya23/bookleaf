import { vi, describe, it, expect, beforeEach } from 'vitest'
import { apiFetch } from './api'

describe('apiFetch', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response()))
  })

  it('attaches Authorization header with bearer token', async () => {
    const getToken = vi.fn().mockResolvedValue('test-token')

    await apiFetch('/images', getToken)

    expect(fetch).toHaveBeenCalledWith(
      expect.stringContaining('/images'),
      expect.objectContaining({
        headers: expect.objectContaining({
          Authorization: 'Bearer test-token',
        }),
      }),
    )
  })

  it('sends request without Authorization header when token is undefined', async () => {
    const getToken = vi.fn().mockResolvedValue(undefined)

    await apiFetch('/images', getToken)

    const [, options] = vi.mocked(fetch).mock.calls[0]
    expect((options?.headers as Record<string, string>)?.Authorization).toBeUndefined()
  })
})
