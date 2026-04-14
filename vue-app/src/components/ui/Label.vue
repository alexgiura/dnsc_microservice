<script setup lang="ts">
import { cn } from '@/lib/utils'

const props = defineProps<{
  /** maps to HTML `for` (reserved in TS) */
  htmlFor?: string
  class?: string
}>()
</script>

<template>
  <label
    :for="props.htmlFor"
    :class="cn('text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70', props.class)"
  >
    <slot />
  </label>
</template>
