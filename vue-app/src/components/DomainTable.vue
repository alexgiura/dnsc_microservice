<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Search, Plus, Loader2 } from 'lucide-vue-next'
import Input from '@/components/ui/Input.vue'
import Button from '@/components/ui/Button.vue'
import DomainRow from '@/components/DomainRow.vue'
import AddDomainDialog from '@/components/AddDomainDialog.vue'
import { domainsApi } from '@/api/domains'
import type { Domain } from '@/models/domain'

type FilterTab = 'all' | 'threat' | 'trusted'

const domains = ref<Domain[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const search = ref('')
const activeFilter = ref<FilterTab>('all')
const dialogOpen = ref(false)

async function fetchDomains() {
  loading.value = true
  error.value = null
  try {
    domains.value = await domainsApi.list()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Eroare la încărcare'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchDomains()
})

async function addDomain(payload: { value: string; description: string; ticketId?: string; tags: string[] }) {
  error.value = null
  try {
    await domainsApi.save({
      value: payload.value.trim(),
      whitelist: false,
      records: [
        {
          ticket_id: payload.ticketId?.trim() ?? null,
          description: payload.description.trim(),
          tags: payload.tags,
          date: new Date().toISOString(),
          source: 'Manual Entry',
        },
      ],
    })
    dialogOpen.value = false
    await fetchDomains()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Eroare la salvare'
  }
}

async function setStatus(id: string, status: 'trusted' | 'threat') {
  const whitelist = status === 'trusted'
  try {
    await domainsApi.update(id, { whitelist })
    domains.value = domains.value.map((d) =>
      d.id === id ? { ...d, whitelist } : d
    )
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Eroare la actualizare'
  }
}

const filtered = computed(() =>
  domains.value.filter((d) => {
    const matchesSearch = d.value.toLowerCase().includes(search.value.toLowerCase())
    const matchesFilter =
      activeFilter.value === 'all' ||
      (activeFilter.value === 'trusted' && d.whitelist) ||
      (activeFilter.value === 'threat' && !d.whitelist)
    return matchesSearch && matchesFilter
  })
)

const threatCount = computed(() => domains.value.filter((d) => !d.whitelist).length)
const trustedCount = computed(() => domains.value.filter((d) => d.whitelist).length)

const tabs = computed(() => [
  { key: 'all' as const, label: 'Toate', count: domains.value.length },
  { key: 'threat' as const, label: 'Threat', count: threatCount.value },
  { key: 'trusted' as const, label: 'Trusted', count: trustedCount.value },
])
</script>

<template>
  <div class="flex flex-col gap-3">
    <div class="flex items-center justify-between">
      <div class="flex gap-1">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          type="button"
          :class="[
            'px-3 py-1.5 text-sm font-medium rounded-md transition-colors',
            activeFilter === tab.key ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:bg-muted',
          ]"
          @click="activeFilter = tab.key"
        >
          {{ tab.label }}
          <span class="ml-1.5 text-xs opacity-70">{{ tab.count }}</span>
        </button>
      </div>
      <Button size="sm" @click="dialogOpen = true">
        <Plus class="h-4 w-4" />
        Adaugă
      </Button>
    </div>

    <div class="bg-card rounded-lg border border-border overflow-hidden relative min-h-[200px]">
      <!-- Loading overlay -->
      <div
        v-if="loading"
        class="absolute inset-0 z-10 flex flex-col items-center justify-center gap-3 bg-background/80 rounded-lg"
      >
        <Loader2 class="h-8 w-8 animate-spin text-muted-foreground" />
        <span class="text-sm text-muted-foreground">Se încarcă domeniile...</span>
      </div>

      <template v-else>
        <div v-if="error" class="px-4 py-3 text-sm text-destructive bg-destructive/10">
          {{ error }}
        </div>

        <div class="flex items-center px-4 py-3 border-b border-border">
          <div class="relative w-80">
            <Search class="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              v-model="search"
              placeholder="Caută domeniu sau IP..."
              class="pl-9 h-9 text-sm"
            />
          </div>
        </div>

        <div class="grid grid-cols-[1fr_100px_100px_100px_50px] gap-4 px-4 py-2.5 text-[10px] uppercase font-semibold text-muted-foreground border-b border-border bg-muted/50 items-center">
          <span class="pl-6">Valoare</span>
          <span class="text-left">Tip</span>
          <span class="text-center">Status</span>
          <span class="text-center">Raportări</span>
          <span class="text-center">Acțiuni</span>
        </div>

        <div
          v-if="filtered.length === 0"
          class="px-4 py-8 text-center text-sm text-muted-foreground"
        >
          Niciun domeniu găsit.
        </div>
        <DomainRow
          v-for="domain in filtered"
          :key="domain.id"
          :domain="domain"
          @set-status="setStatus"
        />
      </template>
    </div>

    <AddDomainDialog
      :open="dialogOpen"
      @update:open="dialogOpen = $event"
      @submit="addDomain"
    />
  </div>
</template>
