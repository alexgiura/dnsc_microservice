<script setup lang="ts">
import { computed } from 'vue'
import { cn } from '@/lib/utils'

type Variant = 'default' | 'secondary' | 'destructive' | 'success' | 'outline' | 'tag' | 'trusted' | 'threat'

const props = withDefaults(
  defineProps<{
    variant?: Variant
    class?: string
  }>(),
  { variant: 'default' }
)

const variantClasses: Record<Variant, string> = {
  default: 'border-transparent bg-primary text-primary-foreground',
  secondary: 'border-transparent bg-secondary text-secondary-foreground',
  destructive: 'border-transparent bg-destructive text-destructive-foreground',
  success: 'border-transparent bg-success text-success-foreground',
  outline: 'text-foreground',
  tag: 'border-border bg-muted text-muted-foreground font-normal',
  trusted: 'border-transparent bg-success text-success-foreground',
  threat: 'border-transparent bg-destructive text-destructive-foreground',
}

const classes = computed(() =>
  cn(
    'inline-flex items-center rounded-sm border px-2 py-0.5 text-xs font-semibold transition-colors',
    variantClasses[props.variant],
    props.class
  )
)
</script>

<template>
  <div :class="classes">
    <slot />
  </div>
</template>
