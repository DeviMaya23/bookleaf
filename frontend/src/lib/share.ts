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
