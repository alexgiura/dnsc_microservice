<script setup lang="ts">
import { ref } from 'vue'
import { ChevronDown, ChevronRight, Globe, Server } from 'lucide-vue-next'
import Badge from '@/components/ui/Badge.vue'
import TicketList from '@/components/TicketList.vue'
import type { DashboardBlacklistFollowUp } from '@/models/dashboard'

defineProps<{
  row: DashboardBlacklistFollowUp
  statusLabel: string
  statusVariant: 'trusted' | 'threat' | 'pending' | 'rejected'
  formatBlacklistAt: (iso: string) => string
}>()

const expanded = ref(false)

function recordsAsTickets(records: DashboardBlacklistFollowUp['records']) {
  return records.map((r) => ({
    ticketId: r.ticket_id,
    description: r.description,
    tags: r.tags ?? [],
    date: r.date,
    source: r.source ?? '',
  }))
}
</script>

<template>
  <div class="border-b border-border last:border-b-0">
    <div
      class="grid grid-cols-[36px_1fr_72px_100px_minmax(0,150px)_100px] gap-3 items-center px-5 py-3 hover:bg-muted/30 transition-colors cursor-pointer text-left"
      @click="expanded = !expanded"
    >
      <span class="flex justify-center text-muted-foreground">
        <ChevronDown v-if="expanded" class="h-4 w-4" />
        <ChevronRight v-else class="h-4 w-4" />
      </span>
      <span class="flex items-center gap-2 min-w-0 select-text" @click.stop>
        <Server v-if="row.type === 'IP'" class="h-3.5 w-3.5 text-muted-foreground shrink-0" />
        <Globe v-else class="h-3.5 w-3.5 text-muted-foreground shrink-0" />
        <span class="font-mono text-xs truncate" :title="row.value">{{ row.value }}</span>
      </span>
      <span class="flex justify-center">
        <Badge variant="outline" class="text-[10px] uppercase">
          {{ row.type }}
        </Badge>
      </span>
      <span class="flex justify-center">
        <Badge :variant="statusVariant" class="text-[10px] uppercase">
          {{ statusLabel }}
        </Badge>
      </span>
      <span
        class="text-[11px] text-muted-foreground text-center tabular-nums leading-tight px-0.5"
        :title="row.blacklisted_at"
      >
        {{ formatBlacklistAt(row.blacklisted_at) }}
      </span>
      <span class="text-xs font-semibold text-center tabular-nums text-destructive">
        {{ row.reports_after_blacklist }}
      </span>
    </div>

    <div
      v-if="expanded"
      class="animate-slide-down bg-muted/30 border-t border-border"
    >
      <div class="px-5 py-2 text-[10px] uppercase font-semibold text-muted-foreground">
        Raportări după blacklist
        <span v-if="row.records.length > 0" class="ml-1.5 opacity-70">{{ row.records.length }}</span>
      </div>
      <TicketList v-if="row.records.length > 0" :tickets="recordsAsTickets(row.records)" />
      <p v-else class="px-5 pb-4 text-xs text-muted-foreground">Nicio raportare de afișat.</p>
    </div>
  </div>
</template>
