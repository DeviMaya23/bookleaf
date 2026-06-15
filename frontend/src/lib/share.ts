import { apiFetch } from './api'

type GetToken = () => Promise<string | undefined>

export interface FolderShare {
  token: string
}

export async function getFolderShare(getToken: GetToken, folderId: string): Promise<FolderShare | null> {
  const res = await apiFetch(`/folders/${folderId}/share`, getToken)
  if (res.status === 404) return null
  if (!res.ok) throw new Error('Failed to fetch folder share')
  return res.json()
}

export async function createFolderShare(getToken: GetToken, folderId: string): Promise<FolderShare> {
  const res = await apiFetch(`/folders/${folderId}/share`, getToken, { method: 'POST' })
  if (!res.ok) throw new Error('Failed to create folder share')
  return res.json()
}

export async function deleteFolderShare(getToken: GetToken, folderId: string): Promise<void> {
  const res = await apiFetch(`/folders/${folderId}/share`, getToken, { method: 'DELETE' })
  if (!res.ok) throw new Error('Failed to delete folder share')
}

export interface SharedImage {
  title: string
  thumbnail_url: string | null
  full_res_url: string
  download_url: string
  width: number | null
  height: number | null
}

export interface SharedFolderResponse {
  folder: { name: string; notes: string | null }
  images: SharedImage[]
}

export async function getSharedFolder(token: string): Promise<SharedFolderResponse> {
  const res = await apiFetch(`/share/${token}`, () => Promise.resolve(undefined))
  if (!res.ok) throw new Error('Failed to fetch shared folder')
  return res.json()
}

export async function exportSharedFolder(token: string): Promise<Blob> {
  const res = await apiFetch(`/share/${token}/export`, () => Promise.resolve(undefined))
  if (!res.ok) throw new Error('Failed to export shared folder')
  return res.blob()
}
