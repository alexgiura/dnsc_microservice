<script setup lang="ts">
import { ref, watch } from 'vue'
import { Pencil } from 'lucide-vue-next'
import Dialog from '@/components/ui/Dialog.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Textarea from '@/components/ui/Textarea.vue'
import Label from '@/components/ui/Label.vue'
import { domainsApi } from '@/api/domains'
import type { Domain, DomainStatusValue } from '@/models/domain'

const props = defineProps<{
  open: boolean
  domain: Domain | null
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  saved: []
}>()

const description = ref('')
const status = ref<DomainStatusValue>('pending')
const saving = ref(false)
const err = ref<string | null>(null)

const statusOptions: { value: DomainStatusValue; label: string }[] = [
  { value: 'whitelist', label: 'Whitelist' },
  { value: 'blacklist', label: 'Blacklist' },
  { value: 'pending', label: 'Pending' },
]

watch(
  () => [props.open, props.domain] as const,
  ([isOpen, d]) => {
    if (isOpen && d) {
      description.value = d.description ?? ''
      status.value = d.status
      err.value = null
    }
  },
  { immediate: true }
)

async function submit() {
  if (!props.domain) return
  saving.value = true
  err.value = null
  try {
    await domainsApi.update(props.domain.id, {
      description: description.value.trim(),
      status: status.value,
    })
    emit('saved')
    emit('update:open', false)
  } catch (e) {
    err.value = e instanceof Error ? e.message : 'Eroare la salvare'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <Dialog :open="open && !!domain" @update:open="emit('update:open', $event)">
    <div v-if="domain" class="grid gap-4 sm:max-w-md">
      <div class="flex flex-col space-y-1.5 text-center sm:text-left pr-6">
        <h2 class="text-lg font-semibold leading-none tracking-tight">Editează domeniu</h2>
        <p class="text-sm text-muted-foreground">
          Modifică descrierea și statusul domeniului.
        </p>
      </div>

      <div class="flex flex-col gap-4 py-2">
        <div class="flex flex-col gap-1.5">
          <Label html-for="edit-domain-value">Valoare</Label>
          <Input id="edit-domain-value" :model-value="domain.value" disabled class="font-mono text-xs" />
          <span class="text-xs text-muted-foreground">Valoarea nu poate fi modificată din acest ecran.</span>
        </div>

        <div class="flex flex-col gap-1.5">
          <Label html-for="edit-domain-desc">Descriere</Label>
          <Textarea
            id="edit-domain-desc"
            v-model="description"
            placeholder="Descriere…"
            class="min-h-[80px] text-sm"
          />
        </div>

        <div class="flex flex-col gap-1.5">
          <Label html-for="edit-domain-status">Status</Label>
          <select
            id="edit-domain-status"
            v-model="status"
            class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            <option v-for="opt in statusOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </option>
          </select>
        </div>

        <p v-if="err" class="text-sm text-destructive">{{ err }}</p>
      </div>

      <div class="flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2 gap-2">
        <Button variant="outline" type="button" :disabled="saving" @click="emit('update:open', false)">
          Anulează
        </Button>
        <Button type="button" :disabled="saving" @click="submit">
          <Pencil class="h-4 w-4" />
          Salvează
        </Button>
      </div>
    </div>
  </Dialog>
</template>
