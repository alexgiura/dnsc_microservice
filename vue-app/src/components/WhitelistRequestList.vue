<script setup lang="ts">
import { MapPin, Phone, CalendarDays, FileText, Mail } from 'lucide-vue-next'
import type { WhitelistRequest } from '@/models/domain'

const props = defineProps<{
  requests: WhitelistRequest[]
}>()

function formatDateTime(v: string): string {
  const s = (v ?? '').replace('T', ' ')
  return s.slice(0, 16)
}
</script>

<template>
  <div v-if="props.requests.length === 0" class="p-3 pl-10 py-6 text-center text-xs text-muted-foreground">
    Nicio cerere de whitelistare.
  </div>

  <div v-else class="p-3 pl-10 grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
    <div
      v-for="req in props.requests"
      :key="req.id"
      class="bg-card border border-border rounded-md shadow-sm hover:shadow-md transition-shadow flex flex-col overflow-hidden"
    >
      <!-- Colored header -->
      <div class="bg-muted/60 px-3 py-2.5 flex items-center justify-between border-b border-border">
        <span class="text-xs font-semibold text-foreground">
          {{ req.last_name }} {{ req.first_name }}
        </span>
        <span class="text-xs text-muted-foreground uppercase tracking-wide flex items-center gap-1">
          <CalendarDays class="h-3 w-3" />
          {{ formatDateTime(req.created_at) }}
        </span>
      </div>

      <!-- Body -->
      <div class="p-3 flex flex-col gap-2">
        <span class="flex items-center gap-1.5 text-xs text-muted-foreground">
          <Mail class="h-3.5 w-3.5" />
          {{ req.email }}
        </span>
        <span class="flex items-center gap-1.5 text-xs text-muted-foreground">
          <Phone class="h-3.5 w-3.5" />
          {{ req.phone }}
        </span>
        <span class="flex items-center gap-1.5 text-xs text-muted-foreground">
          <MapPin class="h-3.5 w-3.5" />
          {{ req.address }}
        </span>
      </div>

      <div class="px-3 pb-3 pt-1 border-t border-border">
        <p class="text-xs text-foreground leading-snug flex items-start gap-1.5 pt-2">
          <FileText class="h-3 w-3 text-muted-foreground shrink-0 mt-0.5" />
          {{ req.reason }}
        </p>
      </div>
    </div>
  </div>
</template>
