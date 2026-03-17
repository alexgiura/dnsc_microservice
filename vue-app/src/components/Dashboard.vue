<script setup lang="ts">
import { computed } from 'vue'
import {
  ShieldAlert,
  ShieldCheck,
  Activity,
  AlertTriangle,
  Globe,
  Server,
  CalendarDays,
} from 'lucide-vue-next'
import { mockDomains } from '@/data/mockData'
import Badge from '@/components/ui/Badge.vue'

const blacklistCount = computed(() => mockDomains.filter((d) => d.status === 'blacklist').length)
const whitelistCount = computed(() => mockDomains.filter((d) => d.status === 'whitelist').length)
const totalTickets = computed(() => mockDomains.reduce((sum, d) => sum + d.tickets.length, 0))

const stats = computed(() => [
  { label: 'Total Domenii', value: mockDomains.length, icon: Activity, color: 'text-foreground' },
  { label: 'Blacklisted', value: blacklistCount.value, icon: ShieldAlert, color: 'text-destructive' },
  { label: 'Whitelisted', value: whitelistCount.value, icon: ShieldCheck, color: 'text-success' },
  { label: 'Total Raportări', value: totalTickets.value, icon: AlertTriangle, color: 'text-muted-foreground' },
])

const recentDomains = computed(() =>
  [...mockDomains]
    .sort((a, b) => b.addedDate.localeCompare(a.addedDate))
    .slice(0, 5)
)
</script>

<template>
  <div class="flex flex-col gap-6">
    <!-- Stat cards -->
    <div class="grid grid-cols-4 gap-4">
      <div
        v-for="stat in stats"
        :key="stat.label"
        class="bg-card border border-border rounded-lg p-5 flex items-center gap-4"
      >
        <div :class="stat.color">
          <component :is="stat.icon" class="h-8 w-8" />
        </div>
        <div>
          <p class="text-2xl font-bold">{{ stat.value }}</p>
          <p class="text-xs text-muted-foreground">{{ stat.label }}</p>
        </div>
      </div>
    </div>

    <!-- Recent domains table -->
    <div class="bg-card border border-border rounded-lg overflow-hidden">
      <div class="px-5 py-4 border-b border-border">
        <h3 class="text-sm font-semibold">Ultimele domenii adăugate</h3>
      </div>
      <div class="grid grid-cols-[1fr_80px_100px_80px_100px] gap-4 px-5 py-2.5 text-[10px] uppercase font-semibold text-muted-foreground border-b border-border bg-muted/50">
        <span>Valoare</span>
        <span class="text-center">Tip</span>
        <span class="text-center">Status</span>
        <span class="text-center">Țară</span>
        <span class="text-center">Dată</span>
      </div>
      <div
        v-for="d in recentDomains"
        :key="d.id"
        class="grid grid-cols-[1fr_80px_100px_80px_100px] gap-4 items-center px-5 py-3 border-b border-border last:border-b-0 hover:bg-muted/30 transition-colors"
      >
        <span class="flex items-center gap-2">
          <Server v-if="d.type === 'IP'" class="h-3.5 w-3.5 text-muted-foreground" />
          <Globe v-else class="h-3.5 w-3.5 text-muted-foreground" />
          <span class="font-mono text-xs">{{ d.value }}</span>
        </span>
        <span class="flex justify-center">
          <Badge variant="outline" class="text-[10px] uppercase justify-center">
            {{ d.type }}
          </Badge>
        </span>
        <span class="flex justify-center">
          <Badge
            :variant="d.status === 'whitelist' ? 'success' : 'destructive'"
            class="text-[10px] uppercase justify-center"
          >
            {{ d.status === 'whitelist' ? 'Whitelist' : 'Blacklist' }}
          </Badge>
        </span>
        <span class="text-xs text-muted-foreground text-center font-medium">
          {{ d.country || '—' }}
        </span>
        <span class="flex items-center justify-center gap-1 text-xs text-muted-foreground">
          <CalendarDays class="h-3 w-3" />
          {{ d.addedDate }}
        </span>
      </div>
    </div>
  </div>
</template>
