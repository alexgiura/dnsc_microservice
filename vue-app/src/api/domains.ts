import { apiBaseUrl } from '@/config/api'
import type { Domain, DomainRecord } from '@/models/domain'

const base = () => `${apiBaseUrl}/api/domains`

export type { Domain, DomainRecord }

export interface SaveDomainPayload {
  value: string
  whitelist: boolean
  records?: Array<{
    ticket_id: string | null
    description: string
    tags: string[]
    date: string
    source: string
  }>
}

export interface UpdateDomainPayload {
  value?: string
  whitelist?: boolean
}

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || `HTTP ${res.status}`)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

export const domainsApi = {
  /** POST /api/domains */
  async save(payload: SaveDomainPayload): Promise<Domain> {
    const res = await fetch(base(), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
    return handleResponse<Domain>(res)
  },

  /** GET /api/domains */
  async list(): Promise<Domain[]> {
    const res = await fetch(base())
    return handleResponse<Domain[]>(res)
  },

  /** GET /api/domains/:id */
  async getById(id: string): Promise<Domain> {
    const res = await fetch(`${base()}/${id}`)
    return handleResponse<Domain>(res)
  },

  /** PATCH /api/domains/:id */
  async update(id: string, payload: UpdateDomainPayload): Promise<Domain> {
    const res = await fetch(`${base()}/${id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
    return handleResponse<Domain>(res)
  },

  /** POST /api/domains/:id/whitelist */
  async whitelist(id: string): Promise<void> {
    const res = await fetch(`${base()}/${id}/whitelist`, { method: 'POST' })
    return handleResponse<void>(res)
  },
}
