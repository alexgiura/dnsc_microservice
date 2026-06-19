<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { Search, Plus, Loader2, ChevronLeft, ChevronRight, ShieldAlert, ShieldCheck, CircleDot, Ban, X } from 'lucide-vue-next'
import Input from '@/components/ui/Input.vue'
import Button from '@/components/ui/Button.vue'
import Select from '@/components/ui/Select.vue'
import Checkbox from '@/components/ui/Checkbox.vue'
import DomainRow from '@/components/DomainRow.vue'
import AddDomainDialog from '@/components/AddDomainDialog.vue'
import EditDomainDialog from '@/components/EditDomainDialog.vue'
import StatusChangeDialog from '@/components/StatusChangeDialog.vue'
import { domainsApi } from '@/api/domains'
import type { Domain, DomainStatusValue } from '@/models/domain'
import { useDomainStatusPermissions } from '@/composables/useDomainStatusPermissions'

type FilterTab = 'all' | 'blacklist' | 'whitelist' | 'pending' | 'rejected'

const PAGE_SIZE_OPTIONS = [10, 20, 50] as const

const { canBulkBlacklistDomain, canBulkWhitelistDomain, canBulkPendingDomain, canBulkRejectDomain } =
  useDomainStatusPermissions()

const domains = ref<Domain[]>([])
const currentPage = ref(1)
/** Implicit 10; poate fi schimbat din selectorul de sub tabel */
const pageSize = ref<(typeof PAGE_SIZE_OPTIONS)[number]>(10)
const loading = ref(true)
const error = ref<string | null>(null)
const search = ref('')
const activeFilter = ref<FilterTab>('all')
const dialogOpen = ref(false)
const editOpen = ref(false)
const editDomain = ref<Domain | null>(null)
const selectedIds = ref<Set<string>>(new Set())
const statusDialog = ref<null | {
  domainId: string
  domainValue: string
  currentStatus: DomainStatusValue
  targetStatus: DomainStatusValue
}>(null)
const bulkDialog = ref<null | {
  targetStatus: DomainStatusValue
  domainIds: string[]
}>(null)

async function fetchDomains() {
  loading.value = true
  error.value = null
  try {
    const list = await domainsApi.list()
    const normalized = Array.isArray(list) ? list : []
    domains.value = normalized.map((d) => ({
      ...d,
      status: d.status ?? 'pending',
      records: d.records ?? [],
      status_history: d.status_history ?? [],
      whitelist_requests: d.whitelist_requests ?? [],
    }))
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Eroare la încărcare'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchDomains()
})

async function addDomain(payload: { value: string; description: string; ticketId?: string; tags: string[] }) {
  error.value = null
  try {
    await domainsApi.save({
      value: payload.value.trim(),
      status: 'pending',
      records: [
        {
          ticket_id: payload.ticketId?.trim() ?? null,
          description: payload.description.trim(),
          tags: payload.tags,
          date: new Date().toISOString(),
          source: 'Manual Entry',
        },
      ],
    })
    dialogOpen.value = false
    await fetchDomains()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Eroare la salvare'
  }
}

function setStatus(id: string, targetStatus: DomainStatusValue) {
  const domain = domains.value.find((d) => d.id === id)
  if (!domain) return

  statusDialog.value = {
    domainId: id,
    domainValue: domain.value,
    currentStatus: domain.status,
    targetStatus,
  }
}

function toggleSelect(id: string) {
  const next = new Set(selectedIds.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  selectedIds.value = next
}

function clearSelection() {
  selectedIds.value = new Set()
}

const selectedDomains = computed(() =>
  domains.value.filter((d) => selectedIds.value.has(d.id))
)

/** Intersecție: acțiunea bulk apare doar dacă TOATE selecțiile o permit. */
const bulkCanBlacklist = computed(
  () => selectedDomains.value.length > 0 && selectedDomains.value.every(canBulkBlacklistDomain)
)
const bulkCanWhitelist = computed(
  () => selectedDomains.value.length > 0 && selectedDomains.value.every(canBulkWhitelistDomain)
)
const bulkCanPending = computed(
  () => selectedDomains.value.length > 0 && selectedDomains.value.every(canBulkPendingDomain)
)
const bulkCanReject = computed(
  () => selectedDomains.value.length > 0 && selectedDomains.value.every(canBulkRejectDomain)
)
const bulkHasAnyAction = computed(
  () =>
    bulkCanBlacklist.value ||
    bulkCanWhitelist.value ||
    bulkCanPending.value ||
    bulkCanReject.value
)

const allPageSelected = computed(
  () =>
    paginatedFiltered.value.length > 0 &&
    paginatedFiltered.value.every((d) => selectedIds.value.has(d.id))
)

function toggleSelectAllPage(checked: boolean) {
  const next = new Set(selectedIds.value)
  for (const d of paginatedFiltered.value) {
    if (checked) next.add(d.id)
    else next.delete(d.id)
  }
  selectedIds.value = next
}

function openBulkDialog(targetStatus: DomainStatusValue) {
  if (selectedDomains.value.length === 0) return
  bulkDialog.value = {
    targetStatus,
    domainIds: selectedDomains.value.map((d) => d.id),
  }
}

function onBulkUpdated() {
  bulkDialog.value = null
  clearSelection()
  fetchDomains()
}

function openEdit(domain: Domain) {
  editDomain.value = domain
  editOpen.value = true
}

function onEditOpen(open: boolean) {
  editOpen.value = open
  if (!open) editDomain.value = null
}

const filtered = computed(() =>
  domains.value.filter((d) => {
    const matchesSearch = d.value.toLowerCase().includes(search.value.toLowerCase())
    const matchesFilter =
      activeFilter.value === 'all' ||
      (activeFilter.value === 'whitelist' && d.status === 'whitelist') ||
      (activeFilter.value === 'blacklist' && d.status === 'blacklist') ||
      (activeFilter.value === 'pending' && d.status === 'pending') ||
      (activeFilter.value === 'rejected' && d.status === 'rejected')
    return matchesSearch && matchesFilter
  })
)

const totalPages = computed(() =>
  Math.max(1, Math.ceil(filtered.value.length / pageSize.value))
)

const paginatedFiltered = computed(() => {
  const ps = pageSize.value
  const start = (currentPage.value - 1) * ps
  return filtered.value.slice(start, start + ps)
})

watch([search, activeFilter], () => {
  currentPage.value = 1
})

watch(pageSize, () => {
  currentPage.value = 1
})

watch(totalPages, (tp) => {
  if (currentPage.value > tp) currentPage.value = tp
})

const pageRangeStart = computed(() =>
  filtered.value.length === 0 ? 0 : (currentPage.value - 1) * pageSize.value + 1
)
const pageRangeEnd = computed(() =>
  Math.min(currentPage.value * pageSize.value, filtered.value.length)
)

/** Text în dreapta: „0 rezultate” sau „X–Y din Z” (ca în designul React). */
const paginationSummary = computed(() => {
  if (filtered.value.length === 0) return '0 rezultate'
  return `${pageRangeStart.value}–${pageRangeEnd.value} din ${filtered.value.length}`
})

function setPageSize(v: number) {
  pageSize.value = v as (typeof PAGE_SIZE_OPTIONS)[number]
}

const pendingCount = computed(() => domains.value.filter((d) => d.status === 'pending').length)
const blacklistCount = computed(() => domains.value.filter((d) => d.status === 'blacklist').length)
const whitelistCount = computed(() => domains.value.filter((d) => d.status === 'whitelist').length)
const rejectedCount = computed(() => domains.value.filter((d) => d.status === 'rejected').length)

const tabs = computed(() => [
  { key: 'all' as const, label: 'Toate', count: domains.value.length },
  { key: 'blacklist' as const, label: 'Blacklist', count: blacklistCount.value },
  { key: 'whitelist' as const, label: 'Whitelist', count: whitelistCount.value },
  { key: 'pending' as const, label: 'Pending', count: pendingCount.value },
  { key: 'rejected' as const, label: 'Respinse', count: rejectedCount.value },
])
</script>

<template>
  <div class="flex flex-col gap-3">
    <div class="flex items-center justify-between">
      <div class="flex flex-wrap gap-1">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          type="button"
          :class="[
            'px-3 py-1.5 text-sm font-medium rounded-md transition-colors',
            activeFilter === tab.key ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:bg-muted',
          ]"
          @click="activeFilter = tab.key"
        >
          {{ tab.label }}
          <span class="ml-1.5 text-xs opacity-70">{{ tab.count }}</span>
        </button>
      </div>
      <Button size="sm" @click="dialogOpen = true">
        <Plus class="h-4 w-4" />
        Adaugă
      </Button>
    </div>

    <div
      class="bg-card rounded-lg border border-border overflow-hidden relative"
      :class="loading && 'min-h-[200px]'"
    >
      <!-- Loading overlay -->
      <div
        v-if="loading"
        class="absolute inset-0 z-10 flex flex-col items-center justify-center gap-3 bg-background/80 rounded-lg"
      >
        <Loader2 class="h-8 w-8 animate-spin text-muted-foreground" />
        <span class="text-sm text-muted-foreground">Se încarcă domeniile...</span>
      </div>

      <template v-else>
        <div v-if="error" class="px-4 py-3 text-sm text-destructive bg-destructive/10">
          {{ error }}
        </div>

        <div v-if="selectedIds.size > 0" class="flex items-center justify-between px-4 py-2.5 border-b border-border bg-primary/5">
          <div class="flex items-center gap-3">
            <Button variant="ghost" size="icon" class="h-7 w-7" aria-label="Anulează selecția" @click="clearSelection">
              <X class="h-4 w-4" />
            </Button>
            <span class="text-sm font-medium">{{ selectedIds.size }} selectate</span>
          </div>
          <div class="flex items-center gap-2">
            <Button
              v-if="bulkCanBlacklist"
              size="sm"
              variant="outline"
              class="border-destructive/40 text-destructive hover:bg-destructive/10 hover:text-destructive"
              @click="openBulkDialog('blacklist')"
            >
              <ShieldAlert class="h-4 w-4" />
              Marchează Blacklist
            </Button>
            <Button
              v-if="bulkCanWhitelist"
              size="sm"
              variant="outline"
              class="border-success/40 text-success hover:bg-success/10 hover:text-success"
              @click="openBulkDialog('whitelist')"
            >
              <ShieldCheck class="h-4 w-4" />
              Marchează Whitelist
            </Button>
            <Button
              v-if="bulkCanPending"
              size="sm"
              variant="outline"
              class="border-muted-foreground/40 text-muted-foreground hover:bg-muted hover:text-foreground"
              @click="openBulkDialog('pending')"
            >
              <CircleDot class="h-4 w-4" />
              Marchează Pending
            </Button>
            <Button
              v-if="bulkCanReject"
              size="sm"
              variant="outline"
              class="border-muted-foreground/40 text-muted-foreground hover:bg-muted hover:text-foreground"
              @click="openBulkDialog('rejected')"
            >
              <Ban class="h-4 w-4" />
              Respinge
            </Button>
            <span v-if="!bulkHasAnyAction" class="text-xs text-muted-foreground">
              Nicio acțiune comună pentru selecție
            </span>
          </div>
        </div>
        <div v-else class="flex items-center px-4 py-3 border-b border-border">
          <div class="relative w-80">
            <Search class="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              v-model="search"
              placeholder="Caută domeniu sau IP..."
              class="pl-9 h-9 text-sm"
            />
          </div>
        </div>

        <div class="grid grid-cols-[36px_minmax(0,1fr)_minmax(0,1fr)_80px_100px_80px_44px] gap-4 px-4 py-2.5 text-[10px] uppercase font-semibold text-muted-foreground border-b border-border bg-muted/50 items-center">
          <span class="flex justify-center">
            <Checkbox
              :checked="allPageSelected"
              aria-label="Selectează toate"
              @update:checked="toggleSelectAllPage($event)"
            />
          </span>
          <span>Valoare</span>
          <span class="text-left min-w-0">Descriere</span>
          <span class="text-left">Tip</span>
          <span class="text-center">Status</span>
          <span class="text-center">Raportări</span>
          <span class="text-center">Acțiuni</span>
        </div>

        <div
          v-if="filtered.length === 0"
          class="px-4 py-8 text-center text-sm text-muted-foreground"
        >
          Niciun domeniu găsit.
        </div>
        <!-- Wrapper: ultimul DomainRow e last-child → last:border-b-0; nu se mai dublează cu border-t la paginare -->
        <div v-if="filtered.length > 0">
          <DomainRow
            v-for="domain in paginatedFiltered"
            :key="domain.id"
            :domain="domain"
            :selected="selectedIds.has(domain.id)"
            @set-status="setStatus"
            @edit="openEdit"
            @toggle-select="toggleSelect"
          />
        </div>

        <div
          v-if="filtered.length > 0"
          class="flex items-center justify-between gap-4 border-t border-border bg-muted/30 px-4 py-3"
        >
          <div class="flex items-center gap-2 text-sm text-muted-foreground">
            <span>Rânduri per pagină:</span>
            <Select
              :model-value="pageSize"
              :options="PAGE_SIZE_OPTIONS"
              @update:model-value="setPageSize"
            />
          </div>

          <div class="flex items-center gap-2 text-sm text-muted-foreground">
            <span class="tabular-nums">{{ paginationSummary }}</span>
            <div class="flex gap-1">
              <Button
                variant="outline"
                size="icon"
                type="button"
                class="h-8 w-8 shrink-0"
                :disabled="currentPage <= 1"
                aria-label="Pagina anterioară"
                @click="currentPage--"
              >
                <ChevronLeft class="h-4 w-4" />
              </Button>
              <Button
                variant="outline"
                size="icon"
                type="button"
                class="h-8 w-8 shrink-0"
                :disabled="currentPage >= totalPages"
                aria-label="Pagina următoare"
                @click="currentPage++"
              >
                <ChevronRight class="h-4 w-4" />
              </Button>
            </div>
          </div>
        </div>
      </template>
    </div>

    <AddDomainDialog
      :open="dialogOpen"
      @update:open="dialogOpen = $event"
      @submit="addDomain"
    />

    <EditDomainDialog
      :open="editOpen"
      :domain="editDomain"
      @update:open="onEditOpen"
      @saved="fetchDomains"
    />

    <StatusChangeDialog
      v-if="statusDialog"
      :open="true"
      :domain-value="statusDialog.domainValue"
      :domain-id="statusDialog.domainId"
      :current-status="statusDialog.currentStatus"
      :target-status="statusDialog.targetStatus"
      @update:open="
        (val) => {
          if (!val) statusDialog = null
        }
      "
      @updated="fetchDomains"
    />

    <StatusChangeDialog
      v-if="bulkDialog"
      :open="true"
      :target-status="bulkDialog.targetStatus"
      :bulk-domain-ids="bulkDialog.domainIds"
      :bulk-count="bulkDialog.domainIds.length"
      @update:open="
        (val) => {
          if (!val) bulkDialog = null
        }
      "
      @updated="onBulkUpdated"
    />
  </div>
</template>
