import { apiFetch } from './api'

export interface Me {
  id: string
  vision_enabled: boolean
}

type GetToken = () => Promise<string | undefined>

export async function getMe(getToken: GetToken): Promise<Me> {
  const res = await apiFetch('/me', getToken)
  if (!res.ok) throw new Error('Failed to fetch user profile')
  return res.json()
}
