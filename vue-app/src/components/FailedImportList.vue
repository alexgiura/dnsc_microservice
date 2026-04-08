<script setup lang="ts">
import { AlertTriangle, Loader2, RotateCcw } from 'lucide-vue-next'
import type { FailedImport } from '@/data/mockData'

withDefaults(
  defineProps<{
    imports: FailedImport[]
    /** Ticket ID currently calling POST .../reimport */
    retryingTicketId: string | null
    /** Eroare la GET import-errors — afișată în corpul listei, nu deasupra search */
    loadError?: string | null
  }>(),
  { loadError: null }
)

const emit = defineEmits<{ retry: [ticketId: string] }>()

const onRetry = (ticketId: string) => {
  emit('retry', ticketId)
}
</script>

<template>
  <div v-if="imports.length === 0" class="px-4 py-8 text-center text-sm">
    <span v-if="loadError" class="text-destructive">{{ loadError }}</span>
    <span v-else class="text-muted-foreground">Nu există importuri eșuate.</span>
  </div>

  <div v-else class="flex flex-col">
    <div
      class="grid grid-cols-[80px_130px_1fr_140px_140px_44px] gap-4 px-4 py-2.5 text-[10px] uppercase font-semibold text-muted-foreground border-b border-border bg-muted/50"
    >
      <span>Ticket ID</span>
      <span>Sursă</span>
      <span>Eroare</span>
      <span>Data ticket</span>
      <span>Ultima sincr.</span>
      <span class="text-center">Acțiuni</span>
    </div>

    <div
      v-for="item in imports"
      :key="item.ticketId"
      class="grid grid-cols-[80px_130px_1fr_140px_140px_44px] gap-4 items-center px-4 py-2.5 border-b border-border last:border-b-0 hover:bg-muted/30 transition-colors"
    >
      <span class="text-sm font-medium text-foreground">{{ item.ticketId }}</span>

      <span class="text-xs text-muted-foreground truncate">{{ item.source }}</span>

      <span class="text-xs font-semibold text-destructive flex items-center gap-1.5">
        <AlertTriangle class="h-3 w-3 flex-shrink-0" />
        {{ item.errorMessage }}
      </span>

      <span class="text-xs text-muted-foreground">{{ item.date }}</span>

      <span class="text-xs text-muted-foreground">{{ item.lastSyncTry }}</span>

      <div class="flex justify-center">
        <button
          type="button"
          :disabled="retryingTicketId !== null"
          class="h-7 w-7 inline-flex items-center justify-center rounded-md text-muted-foreground hover:text-foreground hover:bg-muted transition-colors disabled:opacity-50 disabled:pointer-events-none"
          title="Reîncearcă sincronizarea (retry)"
          @click="onRetry(item.ticketId)"
        >
          <Loader2
            v-if="retryingTicketId === item.ticketId"
            class="h-3.5 w-3.5 animate-spin"
          />
          <RotateCcw v-else class="h-3.5 w-3.5" />
        </button>
      </div>
    </div>
  </div>
</template>
