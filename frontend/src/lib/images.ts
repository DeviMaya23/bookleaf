import { apiFetch } from './api'
import { KeyBetween } from './fracdex'

export interface Image {
  id: string
  title: string
  description: string | null
  mime_type: string
  source_url: string | null
  folder_ids: string[]
  thumbnail_url: string | null
  width: number | null
  height: number | null
  file_size: number | null
  tags: { id: string; name: string }[]
  position: string | null
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
  cursor?: string,
  name?: string,
  sort?: 'created_at' | 'title',
  direction?: 'asc' | 'desc',
): Promise<ImagesPage> {
  const params = new URLSearchParams()
  params.set('unfiled', 'true')
  if (cursor) params.set('cursor', cursor)
  if (name) params.set('name', name)
  if (sort) params.set('sort', sort)
  if (direction) params.set('direction', direction)
  const res = await apiFetch(`/images?${params}`, getToken)
  if (!res.ok) throw new Error('Failed to fetch images')
  return res.json()
}

export async function getFolderImages(
  getToken: GetToken,
  folderId: string,
  sort?: 'created_at' | 'title',
  direction?: 'asc' | 'desc',
): Promise<ImagesPage> {
  const params = new URLSearchParams()
  if (sort) params.set('sort', sort)
  if (direction) params.set('direction', direction)
  const res = await apiFetch(`/images/in-folder/${folderId}?${params}`, getToken)
  if (!res.ok) throw new Error('Failed to fetch folder images')
  const images: Image[] = await res.json()
  return { images, next_cursor: null }
}

export async function getAllImages(
  getToken: GetToken,
  cursor?: string,
  name?: string,
  sort?: 'created_at' | 'title',
  direction?: 'asc' | 'desc',
): Promise<ImagesPage> {
  const params = new URLSearchParams()
  if (cursor) params.set('cursor', cursor)
  if (name) params.set('name', name)
  if (sort) params.set('sort', sort)
  if (direction) params.set('direction', direction)
  const res = await apiFetch(`/images?${params}`, getToken)
  if (!res.ok) throw new Error('Failed to fetch images')
  return res.json()
}

export async function getTrashedImages(getToken: GetToken, cursor?: string, name?: string): Promise<ImagesPage> {
  const params = new URLSearchParams()
  if (cursor) params.set('cursor', cursor)
  if (name) params.set('name', name)
  const res = await apiFetch(`/images/trash?${params}`, getToken)
  if (!res.ok) throw new Error('Failed to fetch trashed images')
  return res.json()
}

export async function restoreImage(getToken: GetToken, id: string): Promise<Image> {
  const res = await apiFetch(`/images/${id}/restore`, getToken, { method: 'POST' })
  if (!res.ok) throw new Error('Failed to restore image')
  return res.json()
}

export async function deleteImage(getToken: GetToken, id: string): Promise<void> {
  const res = await apiFetch(`/images/${id}`, getToken, { method: 'DELETE' })
  if (!res.ok) throw new Error('Failed to delete image')
}

export async function hardDeleteImage(getToken: GetToken, id: string): Promise<void> {
  const res = await apiFetch(`/images/trash/${id}`, getToken, { method: 'DELETE' })
  if (!res.ok) throw new Error('Failed to permanently delete image')
}

export async function emptyTrash(getToken: GetToken): Promise<void> {
  const res = await apiFetch('/images/trash', getToken, { method: 'DELETE' })
  if (!res.ok) throw new Error('Failed to empty trash')
}

export interface InitiateUploadParams {
  title: string
  mimeType: string
  folderId?: string
  description?: string
  sourceUrl?: string
}

export interface InitiateUploadResult {
  id: string
  upload_url: string
  thumbnail_upload_url: string
  r2_path: string
}

export async function initiateUpload(
  getToken: GetToken,
  params: InitiateUploadParams,
): Promise<InitiateUploadResult> {
  const body: Record<string, string> = {
    title: params.title,
    mime_type: params.mimeType,
  }
  if (params.folderId) body.folder_id = params.folderId
  if (params.description) body.description = params.description
  if (params.sourceUrl) body.source_url = params.sourceUrl
  const res = await apiFetch('/images', getToken, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error('Failed to initiate upload')
  return res.json()
}

export async function putToR2(uploadUrl: string, file: Blob | File): Promise<void> {
  const res = await fetch(uploadUrl, {
    method: 'PUT',
    headers: { 'Content-Type': file.type },
    body: file,
  })
  if (!res.ok) throw new Error('Failed to upload file to storage')
}

export interface CompleteUploadResult {
  image_id: string
  suggested_folder_name: string | null
}

export interface CompleteUploadParams {
  width?: number
  height?: number
  fileSize?: number
}

export async function completeUpload(
  getToken: GetToken,
  id: string,
  params?: CompleteUploadParams,
): Promise<CompleteUploadResult> {
  const body: Record<string, number> = {}
  if (params?.width !== undefined) body.width = params.width
  if (params?.height !== undefined) body.height = params.height
  if (params?.fileSize !== undefined) body.file_size = params.fileSize
  const res = await apiFetch(`/images/${id}/complete`, getToken, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error('Failed to complete upload')
  return res.json()
}

export interface ImageDetail extends Image {
  image_url: string
  suggested_folder_name: string | null
}

export async function getImage(getToken: GetToken, id: string): Promise<ImageDetail> {
  const res = await apiFetch(`/images/${id}`, getToken)
  if (!res.ok) throw new Error('Failed to fetch image')
  return res.json()
}

export interface UpdateImageParams {
  title?: string
  description?: string | null
  source_url?: string | null
  folder_ids?: string[] | null
  tags?: string[]
}

export async function updateImage(getToken: GetToken, id: string, params: UpdateImageParams): Promise<Image> {
  const res = await apiFetch(`/images/${id}`, getToken, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(params),
  })
  if (!res.ok) throw new Error('Failed to update image')
  return res.json()
}

export async function moveImageFolder(
  getToken: GetToken,
  imageId: string,
  fromFolderId: string | null,
  toFolderId: string | null,
): Promise<void> {
  const res = await apiFetch(`/images/${imageId}/move-folder`, getToken, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ from_folder_id: fromFolderId, to_folder_id: toFolderId }),
  })
  if (!res.ok) throw new Error('Failed to move image folder')
}

export async function downloadImage(getToken: GetToken, id: string): Promise<string> {
  const res = await apiFetch(`/images/${id}/download`, getToken)
  if (!res.ok) throw new Error('Failed to get download URL')
  const data: { download_url: string } = await res.json()
  return data.download_url
}

export async function updateImagePosition(
  getToken: GetToken,
  imageId: string,
  folderId: string,
  position: string,
): Promise<void> {
  const res = await apiFetch(`/images/${imageId}/position`, getToken, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ folder_id: folderId, position }),
  })
  if (!res.ok) throw new Error('Failed to update image position')
}

export function computeNewPosition(orderedImages: Image[], newIndex: number): string {
  const prevPos = orderedImages[newIndex - 1]?.position || null  // || treats '' as null
  const nextPos = orderedImages[newIndex + 1]?.position || null  // || treats '' as null
  if (prevPos !== null && nextPos !== null && prevPos >= nextPos) {
    // Degenerate case: duplicate or inverted positions (e.g. multiple legacy null-position
    // images that all got 'a0' from earlier KeyBetween(null,null) calls). Fall back to
    // inserting after prevPos rather than throwing.
    return KeyBetween(prevPos, null)
  }
  return KeyBetween(prevPos, nextPos)
}

export async function acceptSuggestion(
  getToken: GetToken,
  id: string,
  suggestedFolderName: string,
): Promise<void> {
  const res = await apiFetch(`/images/${id}/accept-suggestion`, getToken, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ suggested_folder_name: suggestedFolderName }),
  })
  if (!res.ok) throw new Error('Failed to accept folder suggestion')
}
