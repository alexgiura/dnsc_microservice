<script setup lang="ts">
import { ref, computed } from 'vue'
import { ChevronDown, ChevronRight, Globe, Server, MoreVertical, ShieldCheck, ShieldAlert } from 'lucide-vue-next'
import Badge from '@/components/ui/Badge.vue'
import Button from '@/components/ui/Button.vue'
import DropdownMenu from '@/components/ui/DropdownMenu.vue'
import TicketList from '@/components/TicketList.vue'
import type { Domain, DomainRecord } from '@/models/domain'

const props = defineProps<{
  domain: Domain
}>()

const emit = defineEmits<{ setStatus: [id: string, status: 'trusted' | 'threat'] }>()

const expanded = ref(false)
const isTrusted = computed(() => props.domain.whitelist)

function setStatus(status: 'trusted' | 'threat') {
  emit('setStatus', props.domain.id, status)
}

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
          :variant="isTrusted ? 'trusted' : 'threat'"
          class="justify-center text-[10px] uppercase"
        >
          {{ isTrusted ? 'Trusted' : 'Threat' }}
        </Badge>
      </span>

      <span class="text-xs text-muted-foreground text-center flex justify-center">
        {{ domain.records.length }}
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
              v-if="isTrusted"
              type="button"
              class="relative flex w-full cursor-default select-none items-center rounded-sm px-2 py-1.5 text-sm outline-none transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground whitespace-nowrap"
              @click="setStatus('threat')"
            >
              <ShieldAlert class="h-3.5 w-3.5 mr-2 shrink-0 text-destructive" />
              Marchează ca Threat
            </button>
            <button
              v-else
              type="button"
              class="relative flex w-full cursor-default select-none items-center rounded-sm px-2 py-1.5 text-sm outline-none transition-colors hover:bg-accent hover:text-accent-foreground whitespace-nowrap"
              @click="setStatus('trusted')"
            >
              <ShieldCheck class="h-3.5 w-3.5 mr-2 shrink-0 text-success" />
              Marchează ca Trusted
            </button>
          </template>
        </DropdownMenu>
      </span>
    </button>

    <div v-if="expanded" class="animate-slide-down bg-muted/30 border-t border-border">
      <TicketList :tickets="recordsAsTickets(domain.records)" />
    </div>
  </div>
</template>
