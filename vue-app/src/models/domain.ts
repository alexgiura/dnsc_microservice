/** Model al răspunsului BE: GET /api/domains, GET /api/domains/:id */
export interface DomainRecord {
  id: string
  domain_id: string
  ticket_id: string
  description: string
  tags: string[]
  date: string
  source: string
}

/** Aliniat cu BE: core.domains.status */
export type DomainStatusValue = 'whitelist' | 'blacklist' | 'pending' | 'rejected'

export interface DomainStatus {
  id: string
  domain_id: string
  /** BE: whitelist | blacklist | pending | rejected */
  status?: DomainStatusValue
  changed_at: string
  changed_by: string
  notes: string
  /** Răspunsuri vechi BE (înainte de câmpul `status`) */
  whitelist?: boolean
}

export interface WhitelistRequest {
  id: string
  domain_id: string
  first_name: string
  last_name: string
  email: string
  address: string
  phone: string
  reason: string
  created_at: string
}

export interface Domain {
  id: string
  value: string
  type: string
  status: DomainStatusValue
  /** Descriere la nivel de domeniu (primul tichet); opțional */
  description?: string
  records: DomainRecord[]
  status_history?: DomainStatus[]
  whitelist_requests?: WhitelistRequest[]
}
