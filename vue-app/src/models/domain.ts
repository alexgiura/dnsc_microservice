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

export type ThreatStatus = 'threat' | 'trusted'

export interface DomainStatus {
  id: string
  domain_id: string
  whitelist: boolean
  changed_at: string
  changed_by: string
  notes: string
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
  whitelist: boolean
  records: DomainRecord[]
  status_history?: DomainStatus[]
  whitelist_requests?: WhitelistRequest[]
}
