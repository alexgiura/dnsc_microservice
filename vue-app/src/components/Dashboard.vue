<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import {
  ShieldAlert,
  ShieldCheck,
  Activity,
  Clock,
  XCircle,
  Globe,
  Server,
  CalendarDays,
  Tag,
  TrendingUp,
  Loader2,
  BellRing,
  ChevronLeft,
  ChevronRight,
} from 'lucide-vue-next'
import Badge from '@/components/ui/Badge.vue'
import Button from '@/components/ui/Button.vue'
import Select from '@/components/ui/Select.vue'
import BlacklistFollowUpRow from '@/components/BlacklistFollowUpRow.vue'
import { dashboardApi } from '@/api/dashboard'
import type { DashboardResponse } from '@/models/dashboard'

const PAGE_SIZE_OPTIONS = [10, 20, 50] as const

const loading = ref(true)
const error = ref<string | null>(null)
const data = ref<DashboardResponse | null>(null)
/** Paginare tabel Raportări după blacklist; implicit 10 */
const blacklistTablePage = ref(1)
const blacklistTablePageSize = ref<(typeof PAGE_SIZE_OPTIONS)[number]>(10)

onMounted(async () => {
  loading.value = true
  error.value = null
  try {
    data.value = await dashboardApi.get()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Eroare la încărcare'
  } finally {
    loading.value = false
  }
})

const stats = computed(() => {
  const c = data.value?.counts
  if (!c) return []
  return [
    {
      label: 'Total Domenii',
      value: c.total,
      icon: Activity,
      iconColor: 'text-foreground',
      accent: 'border-l-foreground',
    },
    {
      label: 'Blacklist',
      value: c.blacklist,
      icon: ShieldAlert,
      iconColor: 'text-destructive',
      accent: 'border-l-destructive',
    },
    {
      label: 'Whitelist',
      value: c.whitelist,
      icon: ShieldCheck,
      iconColor: 'text-success',
      accent: 'border-l-success',
    },
    {
      label: 'Pending',
      value: c.pending,
      icon: Clock,
      iconColor: 'text-yellow-600 dark:text-yellow-400',
      accent: 'border-l-yellow-500',
    },
    {
      label: 'Rejected',
      value: c.rejected,
      icon: XCircle,
      iconColor: 'text-muted-foreground',
      accent: 'border-l-muted-foreground',
    },
  ] as const
})

function statusBadge(status: string): { label: string; variant: 'trusted' | 'threat' | 'pending' | 'rejected' } {
  switch (status) {
    case 'whitelist':
      return { label: 'Whitelist', variant: 'trusted' }
    case 'blacklist':
      return { label: 'Blacklist', variant: 'threat' }
    case 'rejected':
      return { label: 'Rejected', variant: 'rejected' }
    default:
      return { label: 'Pending', variant: 'pending' }
  }
}

function formatRecordDate(iso: string) {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso.slice(0, 10)
  return d.toLocaleDateString('ro-RO', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

function formatBlacklistAt(iso: string) {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString('ro-RO', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const recentRecords = computed(() => data.value?.recent_records ?? [])
const blacklistFollowUps = computed(() =>
  (data.value?.blacklist_follow_ups ?? []).map((row) => ({
    ...row,
    records: row.records ?? [],
  }))
)

const blacklistTotalPages = computed(() =>
  Math.max(1, Math.ceil(blacklistFollowUps.value.length / blacklistTablePageSize.value))
)

const paginatedBlacklistFollowUps = computed(() => {
  const ps = blacklistTablePageSize.value
  const start = (blacklistTablePage.value - 1) * ps
  return blacklistFollowUps.value.slice(start, start + ps)
})

const blacklistPageRangeStart = computed(() =>
  blacklistFollowUps.value.length === 0
    ? 0
    : (blacklistTablePage.value - 1) * blacklistTablePageSize.value + 1
)
const blacklistPageRangeEnd = computed(() =>
  Math.min(blacklistTablePage.value * blacklistTablePageSize.value, blacklistFollowUps.value.length)
)

const blacklistPaginationSummary = computed(() => {
  if (blacklistFollowUps.value.length === 0) return '0 rezultate'
  return `${blacklistPageRangeStart.value}–${blacklistPageRangeEnd.value} din ${blacklistFollowUps.value.length}`
})

function setBlacklistPageSize(v: number) {
  blacklistTablePageSize.value = v as (typeof PAGE_SIZE_OPTIONS)[number]
}

watch(blacklistTablePageSize, () => {
  blacklistTablePage.value = 1
})

watch(blacklistTotalPages, (tp) => {
  if (blacklistTablePage.value > tp) blacklistTablePage.value = tp
})

const tagsData = computed(() => {
  const tags = data.value?.top_tags ?? []
  return tags.map((t) => ({ name: t.tag, value: t.count }))
})

const maxTagCount = computed(() => {
  const arr = tagsData.value.map((t) => t.value)
  return Math.max(...arr, 1)
})
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex items-end justify-between">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Dashboard</h1>
        <p class="text-sm text-muted-foreground mt-1">
          Privire de ansamblu asupra domeniilor și IP-urilor monitorizate
        </p>
      </div>
    </div>

    <div
      v-if="loading"
      class="flex flex-col items-center justify-center gap-3 py-16 text-muted-foreground"
    >
      <Loader2 class="h-8 w-8 animate-spin" />
      <span class="text-sm">Se încarcă dashboard-ul…</span>
    </div>

    <div v-else-if="error" class="rounded-lg border border-destructive/50 bg-destructive/10 px-4 py-3 text-sm text-destructive">
      {{ error }}
    </div>

    <template v-else>
      <!-- Stat cards -->
      <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4">
        <div
          v-for="stat in stats"
          :key="stat.label"
          :class="[
            'relative overflow-hidden bg-card border border-border border-l-4 rounded-lg p-5 transition-all hover:shadow-md',
            stat.accent,
          ]"
        >
          <component
            :is="stat.icon"
            :class="[
              'absolute right-3 bottom-3 h-16 w-16 opacity-15 pointer-events-none',
              stat.iconColor,
            ]"
          />
          <div class="relative flex flex-col">
            <p class="text-3xl font-bold tracking-tight leading-none">{{ stat.value }}</p>
            <p class="text-xs text-muted-foreground uppercase tracking-wide font-medium mt-2">
              {{ stat.label }}
            </p>
          </div>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Recent records -->
        <div class="lg:col-span-2 bg-card border border-border rounded-lg overflow-hidden">
          <div class="px-5 py-4 border-b border-border flex items-center justify-between">
            <div class="flex items-center gap-2">
              <CalendarDays class="h-4 w-4 text-muted-foreground" />
              <h3 class="text-sm font-semibold">Ultimele domenii adăugate</h3>
            </div>
            <Badge variant="outline" class="text-[10px] uppercase">
              {{ recentRecords.length }} intrări
            </Badge>
          </div>
          <div
            class="grid grid-cols-[1fr_70px_100px_90px] gap-3 px-5 py-2.5 text-[10px] uppercase font-semibold text-muted-foreground border-b border-border bg-muted/40"
          >
            <span>Valoare</span>
            <span class="text-center">Tip</span>
            <span class="text-center">Status</span>
            <span class="text-center">Dată</span>
          </div>
          <div
            v-if="recentRecords.length === 0"
            class="px-5 py-8 text-center text-xs text-muted-foreground"
          >
            Niciun rând de afișat.
          </div>
          <div
            v-for="rec in recentRecords"
            :key="rec.id"
            class="grid grid-cols-[1fr_70px_100px_90px] gap-3 items-center px-5 py-3 border-b border-border last:border-b-0 hover:bg-muted/30 transition-colors"
          >
            <span class="flex items-center gap-2 min-w-0">
              <Server v-if="rec.type === 'IP'" class="h-3.5 w-3.5 text-muted-foreground shrink-0" />
              <Globe v-else class="h-3.5 w-3.5 text-muted-foreground shrink-0" />
              <span class="font-mono text-xs truncate">{{ rec.value }}</span>
            </span>
            <span class="flex justify-center">
              <Badge variant="outline" class="text-[10px] uppercase">
                {{ rec.type }}
              </Badge>
            </span>
            <span class="flex justify-center">
              <Badge :variant="statusBadge(rec.status).variant" class="text-[10px] uppercase">
                {{ statusBadge(rec.status).label }}
              </Badge>
            </span>
            <span class="text-xs text-muted-foreground text-center font-medium">
              {{ formatRecordDate(rec.date) }}
            </span>
          </div>
        </div>

        <!-- Top tags -->
        <div class="bg-card border border-border rounded-lg overflow-hidden">
          <div class="px-5 py-4 border-b border-border flex items-center gap-2">
            <TrendingUp class="h-4 w-4 text-muted-foreground" />
            <h3 class="text-sm font-semibold">Top etichete</h3>
          </div>
          <div class="p-5 flex flex-col gap-4">
            <p v-if="tagsData.length === 0" class="text-xs text-muted-foreground text-center py-6">
              Nu există etichete
            </p>
            <div v-for="(tag, idx) in tagsData" :key="tag.name" class="flex flex-col gap-1.5">
              <div class="flex items-center justify-between gap-2">
                <div class="flex items-center gap-2 min-w-0">
                  <span class="text-[10px] font-mono text-muted-foreground w-4">
                    {{ String(idx + 1).padStart(2, '0') }}
                  </span>
                  <Tag class="h-3 w-3 text-muted-foreground shrink-0" />
                  <span class="text-xs font-medium truncate">{{ tag.name }}</span>
                </div>
                <span class="text-xs font-semibold tabular-nums">{{ tag.value }}</span>
              </div>
              <div class="h-1.5 bg-muted rounded-full overflow-hidden">
                <div
                  class="h-full bg-foreground/80 rounded-full transition-all"
                  :style="{ width: `${(tag.value / maxTagCount) * 100}%` }"
                />
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Blacklist: raportări după ultima trecere în blacklist -->
      <div class="bg-card border border-border rounded-lg overflow-hidden">
        <div class="px-5 py-4 border-b border-border flex items-center justify-between gap-3">
          <div class="flex items-center gap-2 min-w-0">
            <BellRing class="h-4 w-4 text-destructive shrink-0" />
            <h3 class="text-sm font-semibold">Raportări după blacklist</h3>
          </div>
          <Badge variant="outline" class="text-[10px] uppercase shrink-0">
            {{ blacklistFollowUps.length }} domenii
          </Badge>
        </div>
        <div
          class="grid grid-cols-[36px_1fr_72px_100px_minmax(0,150px)_100px] gap-3 px-5 py-2.5 text-[10px] uppercase font-semibold text-muted-foreground border-b border-border bg-muted/40"
        >
          <span />
          <span>Valoare</span>
          <span class="text-center">Tip</span>
          <span class="text-center">Status</span>
          <span class="text-center">Blacklistat la</span>
          <span class="text-center">Raportări noi</span>
        </div>
        <div
          v-if="blacklistFollowUps.length === 0"
          class="px-5 py-8 text-center text-xs text-muted-foreground"
        >
          Nicio intrare: nu există domenii blacklist cu raportări după ultima trecere în blacklist.
        </div>
        <div v-if="blacklistFollowUps.length > 0">
          <BlacklistFollowUpRow
            v-for="row in paginatedBlacklistFollowUps"
            :key="row.domain_id"
            :row="row"
            :status-label="statusBadge(row.status).label"
            :status-variant="statusBadge(row.status).variant"
            :format-blacklist-at="formatBlacklistAt"
          />

          <div
            class="flex items-center justify-between gap-4 border-t border-border bg-muted/30 px-5 py-3"
          >
            <div class="flex items-center gap-2 text-sm text-muted-foreground">
              <span>Rânduri per pagină:</span>
              <Select
                :model-value="blacklistTablePageSize"
                :options="PAGE_SIZE_OPTIONS"
                @update:model-value="setBlacklistPageSize"
              />
            </div>

            <div class="flex items-center gap-2 text-sm text-muted-foreground">
              <span class="tabular-nums">{{ blacklistPaginationSummary }}</span>
              <div class="flex gap-1">
                <Button
                  variant="outline"
                  size="icon"
                  type="button"
                  class="h-8 w-8 shrink-0"
                  :disabled="blacklistTablePage <= 1"
                  aria-label="Pagina anterioară"
                  @click="blacklistTablePage--"
                >
                  <ChevronLeft class="h-4 w-4" />
                </Button>
                <Button
                  variant="outline"
                  size="icon"
                  type="button"
                  class="h-8 w-8 shrink-0"
                  :disabled="blacklistTablePage >= blacklistTotalPages"
                  aria-label="Pagina următoare"
                  @click="blacklistTablePage++"
                >
                  <ChevronRight class="h-4 w-4" />
                </Button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
