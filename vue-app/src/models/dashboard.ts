/** GET /api/dashboard */
export interface DashboardCounts {
  total: number
  blacklist: number
  whitelist: number
  pending: number
  rejected: number
}

export interface DashboardRecentRecord {
  id: string
  domain_id: string
  value: string
  type: string
  status: string
  ticket_id: string
  date: string
  source?: string
}

export interface DashboardTagCount {
  tag: string
  count: number
}

/** Domenii blacklist cu raportări (tickete) după ultima blacklistare */
export interface DashboardBlacklistFollowUp {
  domain_id: string
  value: string
  type: string
  status: string
  blacklisted_at: string
  reports_after_blacklist: number
}

export interface DashboardResponse {
  counts: DashboardCounts
  recent_records: DashboardRecentRecord[]
  top_tags: DashboardTagCount[]
  blacklist_follow_ups: DashboardBlacklistFollowUp[]
}
