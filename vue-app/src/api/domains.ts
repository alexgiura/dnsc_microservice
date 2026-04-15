import { apiFetch } from '@/api/client'
import type { Domain, DomainRecord, DomainStatusValue } from '@/models/domain'

const base = () => `/api/domains`

export type { Domain, DomainRecord }

export interface SaveDomainPayload {
  value: string
  /** whitelist | blacklist | pending; omit sau gol → pending pe server */
  status?: DomainStatusValue
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
  type?: string
  status?: DomainStatusValue
  description?: string
}

export interface WhitelistDomainPayload {
  status: DomainStatusValue
  changeBy: string
  notes?: string
  domainId?: string
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
    const res = await apiFetch(base(), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
    return handleResponse<Domain>(res)
  },

  /** GET /api/domains */
  async list(): Promise<Domain[]> {
    const res = await apiFetch(base())
    return handleResponse<Domain[]>(res)
  },

  /** GET /api/domains/:id */
  async getById(id: string): Promise<Domain> {
    const res = await apiFetch(`${base()}/${id}`)
    return handleResponse<Domain>(res)
  },

  /** PATCH /api/domains/:id */
  async update(id: string, payload: UpdateDomainPayload): Promise<Domain> {
    const res = await apiFetch(`${base()}/${id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
    return handleResponse<Domain>(res)
  },

  /** POST /api/domains/:id/whitelist — body: { status, changeBy, notes } */
  async setDomainStatus(id: string, payload: WhitelistDomainPayload): Promise<void> {
    const res = await apiFetch(`${base()}/${id}/whitelist`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        status: payload.status,
        changeBy: payload.changeBy,
        notes: payload.notes ?? undefined,
        domainId: payload.domainId ?? undefined,
      }),
    })
    return handleResponse<void>(res)
  },
}
