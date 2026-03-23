<script setup lang="ts">
import { computed } from 'vue'
import { User, ArrowRight, MessageSquare, CalendarDays } from 'lucide-vue-next'
import Badge from '@/components/ui/Badge.vue'
import type { DomainStatus, ThreatStatus } from '@/models/domain'

type TimelineEntry = {
  id: string
  fromStatus: ThreatStatus
  toStatus: ThreatStatus
  changedBy: string
  changedAt: string
  notes: string
}

const props = defineProps<{
  history?: DomainStatus[]
}>()

// Demo pentru UI (dacă BE încă nu întoarce `status_history`)
const demoHistory: DomainStatus[] = [
  {
    id: 'demo-1',
    domain_id: 'demo-domain',
    whitelist: true,
    changed_at: '2026-03-16T10:45:00+02:00',
    changed_by: 'security@example.com',
    notes: 'A re-apărut comportament suspect; se revine la Threat.',
  },
  {
    id: 'demo-2',
    domain_id: 'demo-domain',
    whitelist: false,
    changed_at: '2026-03-16T10:15:00+02:00',
    changed_by: 'admin@example.com',
    notes: 'A fost verificat ca fiind benign după analiza internă.',
  },
]

const effectiveHistory = computed(() =>
  props.history && props.history.length > 0 ? props.history : demoHistory
)

function statusLabel(s: ThreatStatus) {
  return s === 'trusted' ? 'Whitelist' : 'Blacklist'
}

function badgeVariantFrom(s: ThreatStatus) {
  return s === 'trusted' ? 'trusted' : 'threat'
}

function badgeVariantTo(s: ThreatStatus) {
  return s === 'threat' ? 'threat' : 'trusted'
}

function formatDateTime(v: string) {
  // Dacă e ISO: 2026-03-16T10:15:00Z -> 2026-03-16 10:15
  const s = (v ?? '').replace('T', ' ')
  return s.slice(0, 16)
}

function whitelistToStatus(whitelist: boolean): ThreatStatus {
  return whitelist ? 'trusted' : 'threat'
}

// BE trimite status_history ordonat DESC (nou -> vechi).
// Afișăm timeline tot în aceeași ordine și derivăm "fromStatus" din intrarea următoare (mai veche).
const timeline = computed<TimelineEntry[]>(() =>
  effectiveHistory.value.map((entry, idx) => {
    const toStatus = whitelistToStatus(entry.whitelist)
    const prev = effectiveHistory.value[idx + 1]
    const fromStatus = prev ? whitelistToStatus(prev.whitelist) : toStatus

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
</script>

<template>
  <div v-if="effectiveHistory.length === 0" class="p-3 pl-10 py-6 text-center text-xs text-muted-foreground">
    Nicio schimbare de status înregistrată.
  </div>

  <div v-else class="p-3 pl-10">
    <div class="relative ml-3">
      <!-- Vertical line -->
      <div class="absolute left-0 top-2 bottom-2 w-px bg-border" />

      <div v-for="entry in timeline" :key="entry.id" class="relative pl-6 pb-3 last:pb-1">
        <!-- Node dot (bulina plină + pe linie) -->
        <div
          class="absolute left-0 top-[18px] h-2.5 w-2.5 rounded-full -translate-x-[4.5px] ring-2 ring-background"
          :class="entry.toStatus === 'threat' ? 'bg-destructive' : 'bg-success'"
        />

        <!-- Card (mic, ca TicketList) -->
        <div
          class="bg-card border border-border rounded-md p-3 shadow-sm hover:shadow-md transition-shadow flex flex-col w-full max-w-md"
        >
          <!-- Row 1: Status vechi -> Status nou | date -->
          <div class="flex items-center justify-between gap-2 mb-2 shrink-0">
            <span class="flex items-center gap-1.5">
              <Badge :variant="badgeVariantFrom(entry.fromStatus)" class="text-[10px] uppercase px-1.5 py-0">
                {{ statusLabel(entry.fromStatus) }}
              </Badge>
              <ArrowRight class="h-3 w-3 text-muted-foreground" />
              <Badge :variant="badgeVariantTo(entry.toStatus)" class="text-[10px] uppercase px-1.5 py-0">
                {{ statusLabel(entry.toStatus) }}
              </Badge>
            </span>

            <span class="text-xs text-muted-foreground uppercase tracking-wide flex items-center gap-1">
              <CalendarDays class="h-3 w-3" />
              {{ formatDateTime(entry.changedAt) }}
            </span>
          </div>

          <!-- Row 2: user (icon + nume aliniate pe mijloc) -->
          <span class="flex items-center gap-1 text-[13px] text-muted-foreground shrink-0">
            <User class="h-3.5 w-3.5 text-muted-foreground self-center" />
            <span class="font-normal text-foreground self-center">{{ entry.changedBy }}</span>
          </span>

          <!-- Row 3: notes (icon + text aliniate pe mijloc) -->
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