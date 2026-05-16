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

export interface InitiateUploadParams {
  title: string
  mimeType: string
  folderId?: string
}

export interface InitiateUploadResult {
  id: string
  upload_url: string
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
  const res = await apiFetch('/images', getToken, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error('Failed to initiate upload')
  return res.json()
}

export async function putToR2(uploadUrl: string, file: File): Promise<void> {
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
  warning?: string
}

export async function completeUpload(
  getToken: GetToken,
  id: string,
): Promise<CompleteUploadResult> {
  const res = await apiFetch(`/images/${id}/complete`, getToken, { method: 'POST' })
  if (!res.ok) throw new Error('Failed to complete upload')
  return res.json()
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
