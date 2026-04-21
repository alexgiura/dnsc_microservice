<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { ShieldAlert, ShieldCheck, CircleDot, Ban } from 'lucide-vue-next'
import Dialog from '@/components/ui/Dialog.vue'
import Button from '@/components/ui/Button.vue'
import Textarea from '@/components/ui/Textarea.vue'
import Badge from '@/components/ui/Badge.vue'
import { domainsApi } from '@/api/domains'
import type { DomainStatusValue } from '@/models/domain'
import { currentUser } from '@/stores/auth'

const props = defineProps<{
  open: boolean
  domainId: string
  domainValue: string
  currentStatus: DomainStatusValue
  targetStatus: DomainStatusValue
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

function statusLabel(s: DomainStatusValue) {
  if (s === 'whitelist') return 'Whitelist'
  if (s === 'blacklist') return 'Blacklist'
  if (s === 'rejected') return 'Respins'
  return 'Pending'
}

function badgeVariant(s: DomainStatusValue): 'trusted' | 'threat' | 'pending' | 'rejected' {
  if (s === 'whitelist') return 'trusted'
  if (s === 'blacklist') return 'threat'
  if (s === 'rejected') return 'rejected'
  return 'pending'
}

const targetBadgeVariant = computed(() => badgeVariant(props.targetStatus))
const currentBadgeVariant = computed(() => badgeVariant(props.currentStatus))

const confirmButtonClasses = computed(() => {
  if (props.targetStatus === 'whitelist') return 'bg-success hover:bg-success/90 text-white'
  if (props.targetStatus === 'blacklist') return 'bg-destructive hover:bg-destructive/90 text-white'
  if (props.targetStatus === 'rejected') return 'bg-secondary hover:bg-secondary/80 text-secondary-foreground'
  return 'bg-secondary hover:bg-secondary/80 text-secondary-foreground'
})

function handleClose(nextOpen: boolean) {
  emit('update:open', nextOpen)
  if (!nextOpen) comment.value = ''
}

async function handleConfirm() {
  const trimmed = comment.value.trim()
  if (!trimmed) return

  const changeBy = currentUser.value?.username?.trim()
  if (!changeBy) {
    error.value = 'Utilizator necunoscut. Reîncearcă autentificarea.'
    return
  }

  error.value = null
  loading.value = true
  try {
    await domainsApi.setDomainStatus(props.domainId, {
      status: props.targetStatus,
      changeBy,
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

      <div class="flex items-center gap-2 text-sm flex-wrap">
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
          <ShieldCheck v-if="targetStatus === 'whitelist'" class="h-4 w-4" />
          <ShieldAlert v-else-if="targetStatus === 'blacklist'" class="h-4 w-4" />
          <Ban v-else-if="targetStatus === 'rejected'" class="h-4 w-4" />
          <CircleDot v-else class="h-4 w-4" />
          Marchează ca {{ statusLabel(targetStatus) }}
        </Button>
      </div>
    </div>
  </Dialog>
</template>
