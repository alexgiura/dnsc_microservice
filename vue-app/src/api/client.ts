import { apiBaseUrl } from '@/config/api'

/** URL absolut sau relativ către API (ex. `/auth/me` sau `${apiBaseUrl}/auth/me`). */
export function apiUrl(path: string): string {
  const p = path.startsWith('/') ? path : `/${path}`
  return `${apiBaseUrl}${p}`
}

/** fetch cu cookie de sesiune (HttpOnly) — obligatoriu pentru auth. */
export async function apiFetch(path: string, init?: RequestInit): Promise<Response> {
  return fetch(apiUrl(path), {
    ...init,
    credentials: 'include',
  })
}
