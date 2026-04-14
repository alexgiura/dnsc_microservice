import { apiFetch } from '@/api/client'

export interface RTIRImportErrorDTO {
  id: string
  ticket_id: string
  source: string
  error: string
  date: string
  last_sync_try_at: string
}

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const text = await res.text()
    let msg = text || `HTTP ${res.status}`
    try {
      const body = JSON.parse(text) as { message?: string; details?: string }
      if (body.message) msg = body.details ? `${body.message}: ${body.details}` : body.message
    } catch {
      /* plain text */
    }
    throw new Error(msg)
  }
  return res.json()
}

export const rtirApi = {
  /** GET /api/rtir/import-errors */
  async getImportErrors(): Promise<RTIRImportErrorDTO[]> {
    const res = await apiFetch('/api/rtir/import-errors')
    return handleResponse<RTIRImportErrorDTO[]>(res)
  },

  /**
   * Retry sync for a failed ticket (POST /api/rtir/tickets/{ticketId}/reimport).
   */
  async retryImport(ticketId: string): Promise<{ ticket_id: string; status: string }> {
    const res = await apiFetch(`/api/rtir/tickets/${encodeURIComponent(ticketId)}/reimport`, {
      method: 'POST',
    })
    return handleResponse(res)
  },
}
