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

export interface Domain {
  id: string
  value: string
  type: string
  whitelist: boolean
  records: DomainRecord[]
}
