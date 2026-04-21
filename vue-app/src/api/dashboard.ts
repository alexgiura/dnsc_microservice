import { apiFetch } from '@/api/client'
import type { DashboardResponse } from '@/models/dashboard'

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || `HTTP ${res.status}`)
  }
  return res.json()
}

export const dashboardApi = {
  async get(): Promise<DashboardResponse> {
    const res = await apiFetch('/api/dashboard')
    return handleResponse<DashboardResponse>(res)
  },
}
