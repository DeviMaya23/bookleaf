import { apiFetch } from './api'

export interface Folder {
  id: string
  name: string
  description: string | null
  parent_id: string | null
  created_at: string
  updated_at: string
}

type GetToken = () => Promise<string | undefined>

export async function getFolders(getToken: GetToken): Promise<Folder[]> {
  const res = await apiFetch('/folders', getToken)
  if (!res.ok) throw new Error('Failed to fetch folders')
  return res.json()
}

export async function createFolder(getToken: GetToken, name: string): Promise<Folder> {
  const res = await apiFetch('/folders', getToken, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name }),
  })
  if (!res.ok) throw new Error('Failed to create folder')
  return res.json()
}

export async function renameFolder(getToken: GetToken, id: string, name: string): Promise<Folder> {
  const res = await apiFetch(`/folders/${id}`, getToken, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name }),
  })
  if (!res.ok) throw new Error('Failed to rename folder')
  return res.json()
}

export async function deleteFolder(getToken: GetToken, id: string): Promise<void> {
  const res = await apiFetch(`/folders/${id}`, getToken, { method: 'DELETE' })
  if (!res.ok) throw new Error('Failed to delete folder')
}
