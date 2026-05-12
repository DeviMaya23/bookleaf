const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? ''

export async function apiFetch(
  path: string,
  getToken: () => Promise<string | undefined>,
  options?: RequestInit,
): Promise<Response> {
  const token = await getToken()
  return fetch(`${BASE_URL}${path}`, {
    ...options,
    headers: {
      ...options?.headers,
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
  })
}
