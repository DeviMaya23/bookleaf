import { apiFetch } from './api'

export interface Folder {
  id: string
  name: string
  description: string | null
  icon: string | null
  parent_id: string | null
  created_at: string
  updated_at: string
}

export interface FolderDetail extends Folder {
  image_count: number
}

type GetToken = () => Promise<string | undefined>

export async function getFolders(getToken: GetToken): Promise<Folder[]> {
  const res = await apiFetch('/folders', getToken)
  if (!res.ok) throw new Error('Failed to fetch folders')
  return res.json()
}

export async function getFolder(getToken: GetToken, id: string): Promise<FolderDetail> {
  const res = await apiFetch(`/folders/${id}`, getToken)
  if (!res.ok) throw new Error('Failed to fetch folder')
  return res.json()
}

export async function createFolder(
  getToken: GetToken,
  name: string,
  parentId?: string,
  icon?: string | null,
): Promise<Folder> {
  const body: Record<string, string | null> = { name }
  if (parentId) body.parent_id = parentId
  if (icon !== undefined) body.icon = icon
  const res = await apiFetch('/folders', getToken, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error('Failed to create folder')
  return res.json()
}

export interface UpdateFolderParams {
  name?: string
  description?: string | null
  icon?: string | null
  parent_id?: string | null
}

export async function updateFolder(getToken: GetToken, id: string, params: UpdateFolderParams): Promise<Folder> {
  const res = await apiFetch(`/folders/${id}`, getToken, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(params),
  })
  if (!res.ok) throw new Error('Failed to update folder')
  return res.json()
}

export async function deleteFolder(getToken: GetToken, id: string): Promise<void> {
  const res = await apiFetch(`/folders/${id}`, getToken, { method: 'DELETE' })
  if (!res.ok) throw new Error('Failed to delete folder')
}

export async function exportFolder(getToken: GetToken, id: string): Promise<Blob> {
  const res = await apiFetch(`/folders/${id}/export`, getToken)
  if (!res.ok) throw new Error('Failed to export folder')
  return res.blob()
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
