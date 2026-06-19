import { computed } from 'vue'
import { currentUser } from '@/stores/auth'
import type { Domain } from '@/models/domain'

export const PUBLIC_BLACKLIST_DESCRIPTIONS = ['Phishing', 'SMiShing', 'Scam', 'Impersonation'] as const

export const PRIVILEGED_BLACKLIST_USER_IDS = new Set([
  '00000000-0000-0000-0000-000000000001',
  '00000000-0000-0000-0000-000000000002',
  '00000000-0000-0000-0000-000000000003',
  '00000000-0000-0000-0000-000000000004',
])

export function useDomainStatusPermissions() {
  const isPrivilegedUser = computed(() => {
    const userId = currentUser.value?.id
    return userId != null && PRIVILEGED_BLACKLIST_USER_IDS.has(userId)
  })

  function canBlacklistDomain(domain: Domain): boolean {
    const desc = domain.description?.trim() ?? ''
    if ((PUBLIC_BLACKLIST_DESCRIPTIONS as readonly string[]).includes(desc)) {
      return true
    }
    return isPrivilegedUser.value
  }

  /** Eligibil pentru bulk blacklist (pending/whitelist/rejected). */
  function canBulkBlacklistDomain(domain: Domain): boolean {
    if (domain.status === 'blacklist') return false
    if (domain.status === 'rejected') return isPrivilegedUser.value
    if (domain.status === 'pending' || domain.status === 'whitelist') {
      return canBlacklistDomain(domain)
    }
    return false
  }

  function canBulkWhitelistDomain(domain: Domain): boolean {
    return domain.status === 'blacklist'
  }

  /** Eligibil pentru bulk pending (rejected → pending, doar useri privilegiați). */
  function canBulkPendingDomain(domain: Domain): boolean {
    return domain.status === 'rejected' && isPrivilegedUser.value
  }

  /** Eligibil pentru bulk reject (pending → rejected, orice user). */
  function canBulkRejectDomain(domain: Domain): boolean {
    return domain.status === 'pending'
  }

  return {
    isPrivilegedUser,
    canBlacklistDomain,
    canBulkBlacklistDomain,
    canBulkWhitelistDomain,
    canBulkPendingDomain,
    canBulkRejectDomain,
  }
}
