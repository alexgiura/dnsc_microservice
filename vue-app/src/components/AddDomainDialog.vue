<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Plus, X } from 'lucide-vue-next'
import Dialog from '@/components/ui/Dialog.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Textarea from '@/components/ui/Textarea.vue'

const AVAILABLE_TAGS = [
  'malware', 'phishing', 'brute-force', 'ransomware', 'c2', 'apt', 'ddos',
  'dns-tunnel', 'exfiltration', 'port-scan', 'social-engineering', 'ssh',
  'critical', 'false-positive', 'verified', 'internal',
]

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  submit: [payload: { value: string; description: string; ticketId?: string; tags: string[] }]
}>()

const value = ref('')
const description = ref('')
const ticketId = ref('')
const selectedTags = ref<string[]>([])

watch(() => props.open, (isOpen) => {
  if (!isOpen) {
    value.value = ''
    description.value = ''
    ticketId.value = ''
    selectedTags.value = []
  }
})

const isIP = computed(() => /^(\d{1,3}\.){3}\d{1,3}$/.test(value.value.trim()))

function toggleTag(tag: string) {
  if (selectedTags.value.includes(tag)) {
    selectedTags.value = selectedTags.value.filter((t) => t !== tag)
  } else {
    selectedTags.value = [...selectedTags.value, tag]
  }
}

function handleSubmit() {
  if (!value.value.trim() || !description.value.trim()) return

  emit('submit', {
    value: value.value.trim(),
    description: description.value.trim(),
    ticketId: ticketId.value.trim() || undefined,
    tags: [...selectedTags.value],
  })
}
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <div class="grid gap-4 sm:max-w-md">
      <div class="flex flex-col space-y-1.5 text-center sm:text-left">
        <h2 class="text-lg font-semibold leading-none tracking-tight">Adaugă domeniu / IP</h2>
        <p class="text-sm text-muted-foreground">
          Completează datele pentru a adăuga o nouă intrare în lista de monitorizare.
        </p>
      </div>

      <div class="flex flex-col gap-4 py-2">
        <div class="flex flex-col gap-1.5">
          <label for="domain-value" class="text-sm font-medium leading-none">Domeniu sau IP</label>
          <Input
            id="domain-value"
            v-model="value"
            placeholder="ex: 192.168.1.1 sau example.com"
          />
          <span v-if="value.trim()" class="text-xs text-muted-foreground">
            Detectat ca: <strong>{{ isIP ? 'IP' : 'Domeniu' }}</strong>
          </span>
        </div>

        <div class="flex flex-col gap-1.5">
          <label for="domain-desc" class="text-sm font-medium leading-none">Descriere</label>
          <Textarea
            id="domain-desc"
            v-model="description"
            placeholder="Descrie domeniul..."
            class="min-h-[60px]"
          />
        </div>

        <div class="flex flex-col gap-1.5">
          <label for="domain-ticket-id" class="text-sm font-medium leading-none">Ticket ID <span class="text-muted-foreground font-normal">(opțional)</span></label>
          <Input
            id="domain-ticket-id"
            v-model="ticketId"
            placeholder="ex: TKT-001"
          />
        </div>

        <div class="flex flex-col gap-1.5">
          <label class="text-sm font-medium leading-none">Etichete</label>
          <div class="flex flex-wrap gap-1.5">
            <button
              v-for="tag in AVAILABLE_TAGS"
              :key="tag"
              type="button"
              :class="[
                'inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium transition-colors border',
                selectedTags.includes(tag)
                  ? 'bg-primary text-primary-foreground border-primary'
                  : 'bg-muted/50 text-muted-foreground border-border hover:bg-muted',
              ]"
              @click="toggleTag(tag)"
            >
              {{ tag }}
              <X v-if="selectedTags.includes(tag)" class="h-3 w-3" />
            </button>
          </div>
        </div>
      </div>

      <div class="flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2 gap-2">
        <Button variant="outline" @click="emit('update:open', false)">
          Anulează
        </Button>
        <Button
          :disabled="!value.trim() || !description.trim()"
          @click="handleSubmit"
        >
          <Plus class="h-4 w-4" />
          Adaugă
        </Button>
      </div>
    </div>
  </Dialog>
</template>
