import { apiFetch } from '@/lib/api'

export interface Me {
  id: string
  vision_enabled: boolean
  folder_icons_enabled: boolean
  ai_categorisation_enabled: boolean
}

type GetToken = () => Promise<string | undefined>

export async function getMe(getToken: GetToken): Promise<Me> {
  const res = await apiFetch('/me', getToken)
  if (!res.ok) throw new Error('Failed to fetch user profile')
  return res.json()
}

export async function deleteMe(getToken: GetToken): Promise<void> {
  const res = await apiFetch('/me', getToken, { method: 'DELETE' })
  if (!res.ok) throw new Error('Failed to delete account')
}

export interface UpdateMeParams {
  vision_enabled?: boolean
  folder_icons_enabled?: boolean
}

export async function updateMe(getToken: GetToken, params: UpdateMeParams): Promise<Me> {
  const res = await apiFetch('/me', getToken, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(params),
  })
  if (!res.ok) throw new Error('Failed to update user profile')
  return res.json()
}
