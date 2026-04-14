import { ref } from 'vue'
import { apiFetch } from '@/api/client'

export interface AuthUser {
  id: string
  username: string
}

export const currentUser = ref<AuthUser | null>(null)
export const authLoading = ref(false)

/** Verifică sesiunea (cookie HttpOnly) prin GET /auth/me. Fără sesiune validă → currentUser null. */
export async function fetchMe(): Promise<boolean> {
  authLoading.value = true
  try {
    const res = await apiFetch('/auth/me')
    if (!res.ok) {
      currentUser.value = null
      return false
    }
    const ct = res.headers.get('content-type') ?? ''
    if (!ct.includes('application/json')) {
      // ex. SPA fallback dacă /auth nu e proxy-at — nu crăpa la JSON
      currentUser.value = null
      return false
    }
    const text = await res.text()
    try {
      const data = JSON.parse(text) as { user?: AuthUser }
      if (data?.user) {
        currentUser.value = data.user
        return true
      }
    } catch {
      /* corp non-JSON */
    }
    currentUser.value = null
    return false
  } catch {
    currentUser.value = null
    return false
  } finally {
    authLoading.value = false
  }
}

export async function login(username: string, password: string): Promise<void> {
  const res = await apiFetch('/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  })
  if (!res.ok) {
    const t = await res.text()
    let msg = 'Utilizator sau parolă incorectă'
    try {
      const j = JSON.parse(t) as { message?: string }
      if (j.message) msg = j.message
    } catch {
      if (t && t.length < 400 && !t.trim().startsWith('<')) msg = t
    }
    throw new Error(msg)
  }
  const ct = res.headers.get('content-type') ?? ''
  if (!ct.includes('application/json')) {
    throw new Error('Răspuns invalid de la server. Verifică proxy-ul /auth către backend.')
  }
  const text = await res.text()
  let data: { user: AuthUser }
  try {
    data = JSON.parse(text) as { user: AuthUser }
  } catch {
    throw new Error('Răspuns invalid de la server.')
  }
  currentUser.value = data.user
}

export async function logout(): Promise<void> {
  await apiFetch('/auth/logout', { method: 'POST' })
  currentUser.value = null
}
