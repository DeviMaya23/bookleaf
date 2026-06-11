import { describe, it, expect, vi, beforeEach } from 'vitest'
import { QueryClient } from '@tanstack/react-query'
import { apiFetch } from './api'
import { resolveOrCreateTags, type Tag } from './tags'

vi.mock('./api', () => ({ apiFetch: vi.fn() }))

const getToken = vi.fn().mockResolvedValue('token')

function jsonResponse(body: unknown, status = 200): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  } as Response
}

function makeClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } })
}

describe('resolveOrCreateTags', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('reuses an existing tag by case-insensitive name without creating it', async () => {
    const client = makeClient()
    const allTags: Tag[] = [{ id: 'tag-abc', name: 'nature' }]

    const result = await resolveOrCreateTags(getToken, [{ id: '', name: 'Nature' }], allTags, client)

    expect(result).toEqual([{ id: 'tag-abc', name: 'nature' }])
    expect(apiFetch).not.toHaveBeenCalled()
  })

  it('creates a new tag, appends it to the cache, and returns it', async () => {
    const client = makeClient()
    client.setQueryData<Tag[]>(['tags'], [])
    vi.mocked(apiFetch).mockResolvedValueOnce(jsonResponse({ id: 'tag-new', name: 'concept' }))

    const result = await resolveOrCreateTags(getToken, [{ id: '', name: 'concept' }], [], client)

    expect(result).toEqual([{ id: 'tag-new', name: 'concept' }])
    expect(client.getQueryData<Tag[]>(['tags'])).toEqual([{ id: 'tag-new', name: 'concept' }])
  })

  it('on a 409 race, re-fetches tags and resolves to the existing one', async () => {
    const client = makeClient()
    vi.mocked(apiFetch)
      .mockResolvedValueOnce(jsonResponse(null, 409))
      .mockResolvedValueOnce(jsonResponse([{ id: 'tag-race', name: 'concept' }]))

    const result = await resolveOrCreateTags(getToken, [{ id: '', name: 'concept' }], [], client)

    expect(result).toEqual([{ id: 'tag-race', name: 'concept' }])
    expect(client.getQueryData<Tag[]>(['tags'])).toEqual([{ id: 'tag-race', name: 'concept' }])
  })

  it('throws a tag-specific message when creation fails', async () => {
    const client = makeClient()
    vi.mocked(apiFetch).mockResolvedValueOnce(jsonResponse(null, 500))

    await expect(
      resolveOrCreateTags(getToken, [{ id: '', name: 'concept' }], [], client),
    ).rejects.toThrow('Failed to create tag "concept"')
  })

  it('throws a resolve message when a 409 re-fetch still cannot find the tag', async () => {
    const client = makeClient()
    vi.mocked(apiFetch)
      .mockResolvedValueOnce(jsonResponse(null, 409))
      .mockResolvedValueOnce(jsonResponse([]))

    await expect(
      resolveOrCreateTags(getToken, [{ id: '', name: 'concept' }], [], client),
    ).rejects.toThrow('Failed to resolve tag "concept"')
  })
})
