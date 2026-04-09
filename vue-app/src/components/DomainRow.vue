<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ChevronDown, ChevronRight, Globe, Server, MoreVertical, ShieldCheck, ShieldAlert } from 'lucide-vue-next'
import Badge from '@/components/ui/Badge.vue'
import Button from '@/components/ui/Button.vue'
import DropdownMenu from '@/components/ui/DropdownMenu.vue'
import TicketList from '@/components/TicketList.vue'
import StatusHistory from '@/components/StatusHistory.vue'
import WhitelistRequestList from '@/components/WhitelistRequestList.vue'
import type { Domain, DomainRecord, DomainStatusValue } from '@/models/domain'

const props = defineProps<{
  domain: Domain
}>()

const emit = defineEmits<{ setStatus: [id: string, status: DomainStatusValue] }>()

const expanded = ref(false)
const status = computed(() => props.domain.status)
const recordsList = computed(() => props.domain.records ?? [])
const activeTab = ref<'tickets' | 'history' | 'whitelist'>('tickets')
const historyCount = computed(() => props.domain.status_history?.length ?? 0)
const whitelistCount = computed(() => props.domain.whitelist_requests?.length ?? 0)

watch(expanded, (val) => {
  if (val) activeTab.value = 'tickets'
})

function setStatus(next: DomainStatusValue) {
  emit('setStatus', props.domain.id, next)
}

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

/** O singură acțiune: pending/whitelist → Blacklist; blacklist → Whitelist */
const menuAction = computed(() => {
  if (status.value === 'blacklist') {
    return { next: 'whitelist' as const, label: 'Marchează ca Whitelist' }
  }
  return { next: 'blacklist' as const, label: 'Marchează ca Blacklist' }
})

/** Map BE records to the ticket shape expected by TicketList */
function recordsAsTickets(records: DomainRecord[]) {
  return records.map((r) => ({
    ticketId: r.ticket_id,
    description: r.description,
    tags: r.tags,
    date: r.date,
    source: r.source,
  }))
}
</script>

<template>
  <div class="border-b border-border last:border-b-0">
    <button
      type="button"
      class="w-full grid grid-cols-[1fr_100px_100px_100px_50px] gap-4 items-center px-4 py-3 hover:bg-muted/50 transition-colors text-left"
      @click="expanded = !expanded"
    >
      <span class="flex items-center gap-2">
        <span class="text-muted-foreground">
          <ChevronDown v-if="expanded" class="h-4 w-4" />
          <ChevronRight v-else class="h-4 w-4" />
        </span>
        <Server v-if="domain.type === 'IP'" class="h-3.5 w-3.5 text-muted-foreground" />
        <Globe v-else class="h-3.5 w-3.5 text-muted-foreground" />
        <span class="font-mono text-xs">{{ domain.value }}</span>
      </span>

      <span class="flex justify-start">
        <Badge variant="outline" class="justify-center text-[10px] uppercase">
          {{ domain.type }}
        </Badge>
      </span>

      <span class="flex justify-center">
        <Badge
          :variant="badgeVariant(status)"
          class="justify-center text-[10px] uppercase"
        >
          {{ statusLabel(status) }}
        </Badge>
      </span>

      <span class="text-xs text-muted-foreground text-center flex justify-center">
        {{ recordsList.length }}
      </span>

      <span class="flex justify-center" @click.stop>
        <DropdownMenu>
          <template #trigger>
            <Button variant="ghost" size="icon" class="h-8 w-8">
              <MoreVertical class="h-4 w-4" />
            </Button>
          </template>
          <template #content>
            <button
              type="button"
              class="relative flex w-full cursor-default select-none items-center rounded-sm px-2 py-1.5 text-sm outline-none transition-colors hover:bg-accent hover:text-accent-foreground whitespace-nowrap"
              @click="setStatus(menuAction.next)"
            >
              <ShieldCheck
                v-if="menuAction.next === 'whitelist'"
                class="h-3.5 w-3.5 mr-2 shrink-0 text-success"
              />
              <ShieldAlert v-else class="h-3.5 w-3.5 mr-2 shrink-0 text-destructive" />
              {{ menuAction.label }}
            </button>
          </template>
        </DropdownMenu>
      </span>
    </button>

    <div v-if="expanded" class="animate-slide-down bg-muted/30 border-t border-border">
      <div class="flex gap-2 px-4 py-3">
        <button
          type="button"
          :class="[
            'px-3 py-1.5 text-[11px] font-medium rounded-md transition-colors',
            activeTab === 'tickets'
              ? 'bg-primary text-primary-foreground'
              : 'text-muted-foreground hover:bg-muted',
          ]"
          @click.stop="activeTab = 'tickets'"
        >
          Raportări
          <span v-if="recordsList.length > 0" class="ml-1.5 text-xs opacity-70">
            {{ recordsList.length }}
          </span>
        </button>

        <button
          type="button"
          :class="[
            'px-3 py-1.5 text-[11px] font-medium rounded-md transition-colors',
            activeTab === 'history'
              ? 'bg-primary text-primary-foreground'
              : 'text-muted-foreground hover:bg-muted',
          ]"
          @click.stop="activeTab = 'history'"
        >
          Istoric Status
          <span v-if="historyCount > 0" class="ml-1.5 text-xs opacity-70">
            {{ historyCount }}
          </span>
        </button>

        <button
          type="button"
          :class="[
            'px-3 py-1.5 text-[11px] font-medium rounded-md transition-colors',
            activeTab === 'whitelist'
              ? 'bg-primary text-primary-foreground'
              : 'text-muted-foreground hover:bg-muted',
          ]"
          @click.stop="activeTab = 'whitelist'"
        >
          Cereri Whitelistare
          <span v-if="whitelistCount > 0" class="ml-1.5 text-xs opacity-70">
            {{ whitelistCount }}
          </span>
        </button>
      </div>

      <div v-if="activeTab === 'tickets'">
        <TicketList :tickets="recordsAsTickets(recordsList)" />
      </div>
      <div v-else-if="activeTab === 'history'">
        <StatusHistory :history="domain.status_history ?? []" />
      </div>
      <div v-else>
        <WhitelistRequestList :requests="domain.whitelist_requests ?? []" />
      </div>
    </div>
  </div>
</template>
