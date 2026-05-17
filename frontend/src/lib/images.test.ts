import { vi, describe, it, expect, beforeEach } from 'vitest'
import { getImage } from './images'

vi.mock('./api', () => ({
  apiFetch: vi.fn(),
}))

import { apiFetch } from './api'

const getToken = vi.fn().mockResolvedValue('token')

describe('getImage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('returns ImageDetail with image_url on success', async () => {
    const imageDetail = {
      id: 'abc',
      title: 'My Photo',
      description: null,
      mime_type: 'image/jpeg',
      source_url: null,
      folder_id: null,
      thumbnail_url: null,
      width: 1920,
      height: 1080,
      file_size: 204800,
      image_url: 'https://r2.example.com/presigned-url',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }
    vi.mocked(apiFetch).mockResolvedValue(new Response(JSON.stringify(imageDetail), { status: 200 }))

    const result = await getImage(getToken, 'abc')

    expect(result.image_url).toBe('https://r2.example.com/presigned-url')
    expect(result.id).toBe('abc')
  })

  it('throws on non-ok response', async () => {
    vi.mocked(apiFetch).mockResolvedValue(new Response(null, { status: 404 }))

    await expect(getImage(getToken, 'missing')).rejects.toThrow('Failed to fetch image')
  })
})
