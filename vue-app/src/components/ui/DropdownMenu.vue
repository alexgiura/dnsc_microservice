<script setup lang="ts">
import { ref, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { cn } from '@/lib/utils'
import { sharedOpenMenuId, nextDropdownMenuId } from '@/components/ui/dropdownMenuShared'

const menuId = nextDropdownMenuId()
const open = ref(false)
const triggerEl = ref<HTMLElement | null>(null)
const contentEl = ref<HTMLElement | null>(null)

/** fixed + align end la trigger — nu e tăiat de overflow pe părinți */
const contentStyle = ref<Record<string, string>>({
  top: '0px',
  right: '0px',
})

function updatePosition() {
  const el = triggerEl.value
  if (!el) return
  const r = el.getBoundingClientRect()
  contentStyle.value = {
    position: 'fixed',
    top: `${Math.round(r.bottom + 4)}px`,
    right: `${Math.round(window.innerWidth - r.right)}px`,
    zIndex: '50',
    minWidth: '8rem',
  }
}

watch(open, async (v) => {
  if (v) {
    sharedOpenMenuId.value = menuId
    await nextTick()
    updatePosition()
  } else if (sharedOpenMenuId.value === menuId) {
    sharedOpenMenuId.value = null
  }
})

watch(sharedOpenMenuId, (id) => {
  if (id !== menuId && open.value) {
    open.value = false
  }
})

function onScrollOrResize() {
  if (open.value) updatePosition()
}

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
  window.addEventListener('scroll', onScrollOrResize, true)
  window.addEventListener('resize', onScrollOrResize)
})
onUnmounted(() => {
  document.removeEventListener('click', onClickOutside)
  window.removeEventListener('scroll', onScrollOrResize, true)
  window.removeEventListener('resize', onScrollOrResize)
})
</script>

<template>
  <div class="relative inline-block">
    <div ref="triggerEl" @click.stop="open = !open">
      <slot name="trigger" />
    </div>
    <Teleport to="body">
      <div
        v-if="open"
        ref="contentEl"
        :style="contentStyle"
        :class="cn(
          'overflow-hidden rounded-md border bg-popover p-1 text-popover-foreground shadow-md'
        )"
        @click.stop="open = false"
      >
        <slot name="content" />
      </div>
    </Teleport>
  </div>
</template>
