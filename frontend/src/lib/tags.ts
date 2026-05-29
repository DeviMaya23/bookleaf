import { apiFetch } from './api'

export interface Tag {
  id: string
  name: string
}

type GetToken = () => Promise<string | undefined>

export async function getTags(getToken: GetToken): Promise<Tag[]> {
  const res = await apiFetch('/tags', getToken)
  if (!res.ok) throw new Error('Failed to fetch tags')
  return res.json()
}

export async function createTag(getToken: GetToken, name: string): Promise<Tag> {
  const res = await apiFetch('/tags', getToken, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name }),
  })
  if (!res.ok) throw new Error('Failed to create tag')
  return res.json()
}
