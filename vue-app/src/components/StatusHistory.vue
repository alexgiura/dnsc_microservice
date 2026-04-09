<script setup lang="ts">
import { computed } from 'vue'
import { User, ArrowRight, MessageSquare, CalendarDays } from 'lucide-vue-next'
import Badge from '@/components/ui/Badge.vue'
import type { DomainStatus, DomainStatusValue } from '@/models/domain'

type TimelineEntry = {
  id: string
  fromStatus: DomainStatusValue
  toStatus: DomainStatusValue
  changedBy: string
  changedAt: string
  notes: string
}

const props = defineProps<{
  history?: DomainStatus[]
}>()

/** Preferă `status` din BE; fallback la `whitelist` pentru răspunsuri vechi. */
function statusFromEntry(e: DomainStatus): DomainStatusValue {
  const s = e.status
  if (s === 'whitelist' || s === 'blacklist' || s === 'pending') return s
  if (typeof e.whitelist === 'boolean') return e.whitelist ? 'whitelist' : 'blacklist'
  return 'pending'
}

const historyList = computed(() => props.history ?? [])

/** Tranziția from→to are sens doar cu ≥2 înregistrări în istoric. */
const showStatusTransition = computed(() => historyList.value.length > 1)

// BE trimite status_history ordonat DESC (nou -> vechi).
const timeline = computed<TimelineEntry[]>(() =>
  historyList.value.map((entry, idx) => {
    const toStatus = statusFromEntry(entry)
    const prev = historyList.value[idx + 1]
    const fromStatus = prev ? statusFromEntry(prev) : toStatus

    return {
      id: entry.id,
      fromStatus,
      toStatus,
      changedBy: entry.changed_by,
      changedAt: entry.changed_at,
      notes: entry.notes,
    }
  })
)

function statusLabel(s: DomainStatusValue) {
  if (s === 'whitelist') return 'Whitelist'
  if (s === 'blacklist') return 'Blacklist'
  return 'Pending'
}

function badgeVariant(s: DomainStatusValue): 'trusted' | 'threat' | 'pending' {
  if (s === 'whitelist') return 'trusted'
  if (s === 'blacklist') return 'threat'
  return 'pending'
}

function formatDateTime(v: string) {
  const s = (v ?? '').replace('T', ' ')
  return s.slice(0, 16)
}

function dotClass(s: DomainStatusValue) {
  if (s === 'blacklist') return 'bg-destructive'
  if (s === 'whitelist') return 'bg-success'
  return 'bg-muted-foreground'
}
</script>

<template>
  <div v-if="!historyList.length" class="p-3 pl-10 py-6 text-center text-xs text-muted-foreground">
    Nicio schimbare de status înregistrată.
  </div>

  <div v-else class="p-3 pl-10">
    <div class="relative ml-3">
      <div class="absolute left-0 top-2 bottom-2 w-px bg-border" />

      <div v-for="entry in timeline" :key="entry.id" class="relative pl-6 pb-3 last:pb-1">
        <div
          class="absolute left-0 top-[18px] h-2.5 w-2.5 rounded-full -translate-x-[4.5px] ring-2 ring-background"
          :class="dotClass(entry.toStatus)"
        />

        <div
          class="bg-card border border-border rounded-md p-3 shadow-sm hover:shadow-md transition-shadow flex flex-col w-full max-w-md"
        >
          <div class="flex items-center justify-between gap-2 mb-2 shrink-0">
            <span v-if="showStatusTransition" class="flex items-center gap-1.5">
              <Badge :variant="badgeVariant(entry.fromStatus)" class="text-[10px] uppercase px-1.5 py-0">
                {{ statusLabel(entry.fromStatus) }}
              </Badge>
              <ArrowRight class="h-3 w-3 text-muted-foreground" />
              <Badge :variant="badgeVariant(entry.toStatus)" class="text-[10px] uppercase px-1.5 py-0">
                {{ statusLabel(entry.toStatus) }}
              </Badge>
            </span>
            <span v-else class="flex items-center gap-1.5">
              <Badge :variant="badgeVariant(entry.toStatus)" class="text-[10px] uppercase px-1.5 py-0">
                {{ statusLabel(entry.toStatus) }}
              </Badge>
            </span>

            <span class="text-xs text-muted-foreground uppercase tracking-wide flex items-center gap-1">
              <CalendarDays class="h-3 w-3" />
              {{ formatDateTime(entry.changedAt) }}
            </span>
          </div>

          <span class="flex items-center gap-1 text-[13px] text-muted-foreground shrink-0">
            <User class="h-3.5 w-3.5 text-muted-foreground self-center" />
            <span class="font-normal text-foreground self-center">{{ entry.changedBy }}</span>
          </span>

          <div class="h-[4px]" />

          <div class="flex items-start gap-1 text-[13px] text-foreground shrink-0">
            <MessageSquare class="h-3.5 w-3.5 text-muted-foreground shrink-0 mt-0.5" />
            <p class="font-normal text-foreground">{{ entry.notes }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
