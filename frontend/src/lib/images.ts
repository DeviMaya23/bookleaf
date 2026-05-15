import { apiFetch } from './api'

export interface Image {
  id: string
  title: string
  description: string | null
  mime_type: string
  source_url: string | null
  folder_id: string | null
  thumbnail_url: string | null
  width: number | null
  height: number | null
  file_size: number | null
  created_at: string
  updated_at: string
}

export interface ImagesPage {
  images: Image[]
  next_cursor: string | null
}

type GetToken = () => Promise<string | undefined>

export async function getImages(
  getToken: GetToken,
  folderId: string | null,
  cursor?: string,
): Promise<ImagesPage> {
  const params = new URLSearchParams()
  if (folderId === null) {
    params.set('unfiled', 'true')
  } else {
    params.set('folder_id', folderId)
  }
  if (cursor) params.set('cursor', cursor)
  const res = await apiFetch(`/images?${params}`, getToken)
  if (!res.ok) throw new Error('Failed to fetch images')
  return res.json()
}

export async function deleteImage(getToken: GetToken, id: string): Promise<void> {
  const res = await apiFetch(`/images/${id}`, getToken, { method: 'DELETE' })
  if (!res.ok) throw new Error('Failed to delete image')
}
