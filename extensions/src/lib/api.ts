import { getAuth } from "./storage";

export async function apiFetch(
  path: string,
  options: RequestInit = {},
): Promise<Response> {
  const auth = await getAuth();
  const baseUrl = import.meta.env.VITE_API_BASE_URL;

  const headers = new Headers(options.headers);
  if (auth) {
    headers.set("Authorization", `Bearer ${auth.accessToken}`);
  }

  return fetch(`${baseUrl}${path}`, { ...options, headers });
}

export async function getFolders(): Promise<Array<{ id: string; name: string }>> {
  const res = await apiFetch("/folders");
  if (!res.ok) throw new Error(`GET /folders failed: ${res.status}`);
  return res.json() as Promise<Array<{ id: string; name: string }>>;
}
