<script setup lang="ts">
import { X } from 'lucide-vue-next'
import { cn } from '@/lib/utils'

defineProps<{
  open: boolean
}>()

const emit = defineEmits<{ 'update:open': [value: boolean] }>()
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="fixed inset-0 z-50">
      <div
        class="fixed inset-0 z-50 bg-black/80"
        aria-hidden="true"
        @click="emit('update:open', false)"
      />
      <div
        role="dialog"
        :class="cn(
          'fixed left-[50%] top-[50%] z-50 grid w-full max-w-lg -translate-x-1/2 -translate-y-1/2 gap-4 border bg-background p-6 shadow-lg sm:rounded-lg'
        )"
        @click.stop
      >
        <button
          type="button"
          class="absolute right-4 top-4 rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
          @click="emit('update:open', false)"
        >
          <X class="h-4 w-4" />
          <span class="sr-only">Close</span>
        </button>
        <slot />
      </div>
    </div>
  </Teleport>
</template>
