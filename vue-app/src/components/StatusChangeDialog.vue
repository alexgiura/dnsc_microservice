<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { ShieldAlert, ShieldCheck } from 'lucide-vue-next'
import Dialog from '@/components/ui/Dialog.vue'
import Button from '@/components/ui/Button.vue'
import Textarea from '@/components/ui/Textarea.vue'
import Badge from '@/components/ui/Badge.vue'
import { domainsApi } from '@/api/domains'
import type { ThreatStatus } from '@/models/domain'

const props = defineProps<{
  open: boolean
  domainId: string
  domainValue: string
  currentStatus: ThreatStatus
  targetStatus: ThreatStatus
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  updated: []
}>()

const comment = ref('')
const error = ref<string | null>(null)
const loading = ref(false)

watch(
  () => props.open,
  (isOpen) => {
    if (!isOpen) comment.value = ''
  }
)

const isThreatTarget = computed(() => props.targetStatus === 'threat')
const targetBadgeVariant = computed(() => (isThreatTarget.value ? 'threat' : 'trusted'))

const currentBadgeVariant = computed(() => (props.currentStatus === 'threat' ? 'threat' : 'trusted'))

function statusLabel(status: ThreatStatus) {
  return status === 'trusted' ? 'Whitelist' : 'Blacklist'
}

const confirmButtonClasses = computed(() => {
  // Vue project nu are clase bg-trusted/bg-threat, așa că mapăm pe succes/dezasttrus
  return props.targetStatus === 'trusted'
    ? 'bg-success hover:bg-success/90 text-white'
    : 'bg-destructive hover:bg-destructive/90 text-white'
})

function handleClose(nextOpen: boolean) {
  emit('update:open', nextOpen)
  if (!nextOpen) comment.value = ''
}

async function handleConfirm() {
  const trimmed = comment.value.trim()
  if (!trimmed) return

  error.value = null
  loading.value = true
  try {
    const whitelist = props.targetStatus === 'trusted'
    await domainsApi.whitelist(props.domainId, {
      whitelist,
      changeBy: 'user@example.com',
      notes: trimmed,
    })

    emit('updated')
    handleClose(false)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Eroare la schimbarea statusului'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="handleClose">
    <div class="grid gap-4 sm:max-w-md">
      <div class="flex flex-col space-y-1.5 text-center sm:text-left">
        <h2 class="text-base leading-none">Schimbare status</h2>
        <p class="text-xs text-muted-foreground">Adaugă un motiv pentru schimbarea statusului.</p>
      </div>

      <div class="flex items-center gap-2 text-sm">
        <span class="font-mono text-xs text-muted-foreground">{{ domainValue }}</span>
        <span class="text-muted-foreground">:</span>
        <Badge :variant="currentBadgeVariant" class="text-[10px] uppercase">
          {{ statusLabel(currentStatus) }}
        </Badge>
        <span class="text-muted-foreground">→</span>
        <Badge :variant="targetBadgeVariant" class="text-[10px] uppercase">
          {{ statusLabel(targetStatus) }}
        </Badge>
      </div>

      <div class="flex flex-col gap-1.5">
        <Textarea
          id="status-change-notes"
          v-model="comment"
          placeholder="Motivul schimbării statusului... (obligatoriu)"
          class="min-h-[80px] text-sm"
          autofocus
        />
        <p v-if="error" class="text-xs text-destructive">{{ error }}</p>
      </div>

      <div class="flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2 gap-2">
        <Button variant="outline" size="sm" @click="handleClose(false)">
          Anulează
        </Button>
        <Button
          size="sm"
          :disabled="!comment.trim() || loading"
          @click="handleConfirm"
          :class="confirmButtonClasses"
        >
          <ShieldCheck v-if="targetStatus === 'trusted'" class="h-4 w-4" />
          <ShieldAlert v-else class="h-4 w-4" />
          {{ targetStatus === 'trusted' ? 'Marchează ca Whitelist' : 'Marchează ca Blacklist' }}
        </Button>
      </div>
    </div>
  </Dialog>
</template>

