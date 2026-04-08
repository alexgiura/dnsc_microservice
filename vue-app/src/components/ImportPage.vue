<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Search } from 'lucide-vue-next'
import Input from '@/components/ui/Input.vue'
import FailedImportList from '@/components/FailedImportList.vue'
import { rtirApi } from '@/api/rtir'
import type { FailedImport } from '@/data/mockData'
import { formatDateTimeInAppTz } from '@/utils/formatDateTime'

const failedImports = ref<FailedImport[]>([])
const loading = ref(true)
const loadError = ref<string | null>(null)
const retryingTicketId = ref<string | null>(null)
const search = ref('')

const filtered = computed(() =>
  failedImports.value.filter((item) =>
    item.ticketId.toLowerCase().includes(search.value.toLowerCase())
  )
)

async function loadErrors(options?: { silent?: boolean }) {
  const silent = options?.silent ?? false
  if (!silent) {
    loading.value = true
    loadError.value = null
  }
  try {
    const rows = await rtirApi.getImportErrors()
    failedImports.value = rows.map((r) => ({
      ticketId: r.ticket_id,
      source: r.source,
      errorMessage: r.error,
      date: formatDateTimeInAppTz(r.date),
      lastSyncTry: formatDateTimeInAppTz(r.last_sync_try_at),
    }))
  } catch (e) {
    loadError.value = e instanceof Error ? e.message : 'Eroare la încărcare'
  } finally {
    if (!silent) loading.value = false
  }
}

onMounted(() => {
  void loadErrors()
})

const handleRetry = async (ticketId: string) => {
  retryingTicketId.value = ticketId
  try {
    await rtirApi.retryImport(ticketId)
  } catch {
    // Eșecul e înregistrat în BE în rtir_import_errors; nu afișăm mesajul din răspunsul HTTP.
  } finally {
    await loadErrors({ silent: true })
    retryingTicketId.value = null
  }
}
</script>

<template>
  <div class="flex flex-col gap-3">
    <div class="flex items-center justify-between">
      <div class="flex gap-1">
        <button
          type="button"
          class="px-3 py-1.5 text-sm font-medium rounded-md bg-destructive text-destructive-foreground"
        >
          Erori Import
          <span class="ml-1.5 text-xs opacity-70">{{ failedImports.length }}</span>
        </button>
      </div>
    </div>

    <div class="bg-card rounded-lg border border-border overflow-hidden">
      <div class="flex items-center px-4 py-3 border-b border-border">
        <div class="relative w-64">
          <Search
            class="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none"
          />
          <Input
            v-model="search"
            placeholder="Caută ticket..."
            class="pl-9 h-9 text-sm"
          />
        </div>
      </div>

      <div v-if="loading" class="px-4 py-8 text-center text-sm text-muted-foreground">
        Se încarcă…
      </div>
      <FailedImportList
        v-else
        :imports="filtered"
        :load-error="loadError"
        :retrying-ticket-id="retryingTicketId"
        @retry="handleRetry"
      />
    </div>
  </div>
</template>
