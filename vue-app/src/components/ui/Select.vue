<script setup lang="ts">
import { ref, watch, nextTick, onMounted, onUnmounted, computed } from 'vue'
import { ChevronDown, Check } from 'lucide-vue-next'
import { cn } from '@/lib/utils'
import { computeFlipPopoverY } from '@/lib/popoverFlip'
import { sharedOpenMenuId, nextDropdownMenuId } from '@/components/ui/dropdownMenuShared'

/** Select tip shadcn: trigger + listă în popover (închide alte DropdownMenu-uri deschise). */
const menuId = nextDropdownMenuId()
const open = ref(false)
const triggerEl = ref<HTMLElement | null>(null)
const contentEl = ref<HTMLElement | null>(null)

const contentStyle = ref<Record<string, string>>({})

const props = withDefaults(
  defineProps<{
    modelValue: number
    options: readonly number[]
    disabled?: boolean
    /** clase suplimentare pe trigger (ex. w-[70px]) */
    triggerClass?: string
  }>(),
  { disabled: false, triggerClass: '' }
)

const emit = defineEmits<{ 'update:modelValue': [value: number] }>()

const display = computed(() => String(props.modelValue))

function updatePosition() {
  requestAnimationFrame(() => {
    const trigger = triggerEl.value
    const content = contentEl.value
    if (!trigger) return
    const r = trigger.getBoundingClientRect()
    const vw = window.innerWidth

    if (!content) {
      void nextTick(() => updatePosition())
      return
    }

    const contentH = content.offsetHeight
    if (contentH === 0 && open.value) {
      void nextTick(() => updatePosition())
      return
    }

    const { top, maxHeight } = computeFlipPopoverY(r, contentH)
    let leftPx = Math.round(r.left)
    const width = Math.round(r.width)
    if (leftPx + width > vw - 8) {
      leftPx = Math.max(8, vw - width - 8)
    }

    contentStyle.value = {
      position: 'fixed',
      top: `${Math.round(top)}px`,
      left: `${leftPx}px`,
      width: `${width}px`,
      zIndex: '50',
      maxHeight: `${Math.round(maxHeight)}px`,
      overflowX: 'hidden',
      overflowY: 'auto',
    }
  })
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

function select(n: number) {
  emit('update:modelValue', n)
  open.value = false
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
    <button
      ref="triggerEl"
      type="button"
      role="combobox"
      :aria-expanded="open"
      :disabled="disabled"
      :class="
        cn(
          'flex h-8 w-[70px] items-center justify-between gap-1 rounded-md border border-input bg-background px-2 text-sm text-foreground shadow-sm ring-offset-background',
          'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
          'disabled:cursor-not-allowed disabled:opacity-50',
          open && 'ring-2 ring-ring ring-offset-2',
          triggerClass
        )
      "
      @click.stop="open = !open"
    >
      <span class="truncate tabular-nums">{{ display }}</span>
      <ChevronDown
        class="h-4 w-4 shrink-0 opacity-50 transition-transform"
        :class="open && 'rotate-180'"
      />
    </button>
    <Teleport to="body">
      <div
        v-if="open"
        ref="contentEl"
        :style="contentStyle"
        :class="
          cn(
            'rounded-md border bg-popover p-1 text-popover-foreground shadow-md'
          )
        "
        @click.stop
      >
        <button
          v-for="n in options"
          :key="n"
          type="button"
          role="option"
          :aria-selected="n === modelValue"
          class="relative flex w-full cursor-pointer select-none items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-none hover:bg-accent hover:text-accent-foreground"
          @click="select(n)"
        >
          <span class="flex-1 truncate text-left tabular-nums">{{ n }}</span>
          <Check v-if="n === modelValue" class="h-4 w-4 shrink-0 opacity-100" />
        </button>
      </div>
    </Teleport>
  </div>
</template>
