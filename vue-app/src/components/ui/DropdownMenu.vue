<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { cn } from '@/lib/utils'

const open = ref(false)
const triggerEl = ref<HTMLElement | null>(null)
const contentEl = ref<HTMLElement | null>(null)

function onClickOutside(e: MouseEvent) {
  const target = e.target as Node
  if (
    open.value &&
    triggerEl.value &&
    !triggerEl.value.contains(target) &&
    contentEl.value &&
    !contentEl.value.contains(target)
  ) {
    open.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', onClickOutside)
})
onUnmounted(() => {
  document.removeEventListener('click', onClickOutside)
})
</script>

<template>
  <div class="relative inline-block">
    <div ref="triggerEl" @click.stop="open = !open">
      <slot name="trigger" />
    </div>
    <div
      v-if="open"
      ref="contentEl"
      :class="cn(
        'absolute right-0 top-full z-50 mt-1 min-w-[8rem] overflow-hidden rounded-md border bg-popover p-1 text-popover-foreground shadow-md'
      )"
      @click.stop="open = false"
    >
      <slot name="content" />
    </div>
  </div>
</template>
