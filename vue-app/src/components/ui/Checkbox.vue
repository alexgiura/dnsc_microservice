<script setup lang="ts">
import { Check } from 'lucide-vue-next'
import { cn } from '@/lib/utils'

const props = withDefaults(
  defineProps<{
    checked?: boolean
    class?: string
    ariaLabel?: string
    disabled?: boolean
  }>(),
  { checked: false, disabled: false }
)

const emit = defineEmits<{
  'update:checked': [value: boolean]
}>()

function toggle() {
  if (props.disabled) return
  emit('update:checked', !props.checked)
}
</script>

<template>
  <button
    type="button"
    role="checkbox"
    :aria-checked="checked"
    :aria-label="ariaLabel"
    :disabled="disabled"
    :class="
      cn(
        'inline-flex h-4 w-4 shrink-0 items-center justify-center rounded-sm border border-primary ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50',
        checked ? 'bg-primary text-primary-foreground' : 'bg-background',
        props.class
      )
    "
    @click.stop="toggle"
  >
    <Check v-if="checked" class="h-3 w-3" />
  </button>
</template>
