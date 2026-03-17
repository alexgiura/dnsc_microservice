<script setup lang="ts">
import { Info, ExternalLink } from 'lucide-vue-next'
import type { Ticket } from '@/data/mockData'

const tagColors: Record<string, string> = {
  CRITICAL: 'border-destructive text-destructive',
  INTERNAL: 'border-success text-success',
  DEVOPS: 'border-primary text-primary',
  'UX RESEARCH': 'border-success text-success',
  'HIGH PRIORITY': 'border-[hsl(25,95%,53%)] text-[hsl(25,95%,53%)]',
  DATABASE: 'border-destructive text-destructive',
  PERFORMANCE: 'border-primary text-primary',
}

function getTagColor(tag: string): string {
  const upper = tag.toUpperCase()
  return tagColors[upper] || 'border-muted-foreground text-muted-foreground'
}

defineProps<{
  tickets: Ticket[]
}>()
</script>

<template>
  <div class="p-3 pl-10 grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
    <div
      v-for="(ticket, index) in tickets"
      :key="index"
      class="bg-card border border-border rounded-md p-3 shadow-sm hover:shadow-md transition-shadow flex flex-col"
    >
      <div class="flex items-center justify-between mb-2 shrink-0">
        <span class="font-mono text-xs text-muted-foreground bg-muted px-1.5 py-0.5 rounded-sm">
          {{ ticket.ticketId || '###' }}
        </span>
        <span class="text-xs text-muted-foreground uppercase tracking-wide">
          {{ ticket.date.slice(0, 10) }}
        </span>
      </div>

      <p class="text-sm text-foreground leading-snug mb-2 shrink-0">
        {{ ticket.description }}
      </p>

      <div class="flex gap-1.5 flex-wrap shrink-0">
        <span
          v-for="tag in ticket.tags"
          :key="tag"
          :class="['text-[10px] font-semibold uppercase tracking-wider border rounded-sm px-1.5 py-px', getTagColor(tag)]"
        >
          {{ tag }}
        </span>
      </div>

      <div class="min-h-3 flex-1 shrink-0" aria-hidden="true" />

      <div class="flex items-center justify-between pt-2 mt-auto border-t border-border shrink-0">
        <span class="flex items-center gap-1 text-xs text-muted-foreground">
          <Info class="h-3.5 w-3.5" />
          <span class="font-medium text-foreground">{{ ticket.source }}</span>
        </span>
        <button type="button" class="text-xs font-medium text-success hover:underline flex items-center gap-1">
          View Details
          <ExternalLink class="h-3 w-3" />
        </button>
      </div>
    </div>
  </div>
</template>
