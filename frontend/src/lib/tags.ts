import type { QueryClient } from '@tanstack/react-query'
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

// Returns null on 409 (tag already exists for this user — caller resolves via re-fetch)
export async function createTag(getToken: GetToken, name: string): Promise<Tag | null> {
  const res = await apiFetch('/tags', getToken, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name }),
  })
  if (res.status === 409) return null
  if (!res.ok) throw new Error('Failed to create tag')
  return res.json()
}

// Resolves a list of tags (some possibly newly typed, with no id) to tags that
// all have real ids, creating any that don't yet exist. The ['tags'] cache is
// kept in sync as tags are created or re-fetched. Throws with a tag-specific
// message if a tag can't be resolved or created — caller surfaces it.
export async function resolveOrCreateTags(
  getToken: GetToken,
  incoming: Tag[],
  allTags: Tag[],
  queryClient: QueryClient,
): Promise<Tag[]> {
  const resolved: Tag[] = []
  for (const t of incoming) {
    if (t.id) {
      resolved.push(t)
      continue
    }
    const existing = allTags.find((a) => a.name.toLowerCase() === t.name.toLowerCase())
    if (existing) {
      resolved.push(existing)
      continue
    }
    try {
      const created = await createTag(getToken, t.name)
      if (created === null) {
        // 409: race — another tab created the same tag; re-fetch to get real ID
        const fresh = await getTags(getToken)
        queryClient.setQueryData<Tag[]>(['tags'], fresh)
        const found = fresh.find((a) => a.name.toLowerCase() === t.name.toLowerCase())
        if (found) {
          resolved.push(found)
          continue
        }
        throw new Error(`Failed to resolve tag "${t.name}"`)
      }
      queryClient.setQueryData<Tag[]>(['tags'], (prev = []) => [...prev, created])
      resolved.push(created)
    } catch (err) {
      if (err instanceof Error && err.message.startsWith('Failed to resolve tag')) throw err
      throw new Error(`Failed to create tag "${t.name}"`, { cause: err })
    }
  }
  return resolved
}
