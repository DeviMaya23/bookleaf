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

export async function createFolder(getToken: GetToken, name: string, parentId?: string): Promise<Folder> {
  const body: Record<string, string> = { name }
  if (parentId) body.parent_id = parentId
  const res = await apiFetch('/folders', getToken, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
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

export async function moveFolder(getToken: GetToken, id: string, name: string, parentId: string | null): Promise<Folder> {
  const body: Record<string, string | null> = { name, parent_id: parentId }
  const res = await apiFetch(`/folders/${id}`, getToken, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error('Failed to move folder')
  return res.json()
}

export async function deleteFolder(getToken: GetToken, id: string): Promise<void> {
  const res = await apiFetch(`/folders/${id}`, getToken, { method: 'DELETE' })
  if (!res.ok) throw new Error('Failed to delete folder')
}

export function getFolderSubtreeIds(folders: Folder[], folderId: string): Set<string> {
  const ids = new Set<string>([folderId])
  let changed = true
  while (changed) {
    changed = false
    for (const f of folders) {
      if (f.parent_id && ids.has(f.parent_id) && !ids.has(f.id)) {
        ids.add(f.id)
        changed = true
      }
    }
  }
  return ids
}
