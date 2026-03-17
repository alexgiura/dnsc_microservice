<script setup lang="ts">
import { User } from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import dnscLogo from '@/assets/dnsc-logo.svg'

defineProps<{
  activeTab: 'dashboard' | 'domains'
}>()

const emit = defineEmits<{ 'update:activeTab': ['dashboard' | 'domains'] }>()
</script>

<template>
  <header class="h-16 bg-topbar text-topbar-foreground flex items-center justify-between px-6 shrink-0">
    <div class="flex items-center gap-2">
      <img :src="dnscLogo" alt="DNSC" class="h-11 w-11 rounded-full" />
      <span class="font-semibold text-sm tracking-wide">DNSC</span>
    </div>

    <div class="flex bg-topbar-foreground/10 rounded-md p-0.5">
      <button
        type="button"
        :class="[
          'px-4 py-1.5 text-sm font-medium rounded-sm transition-colors',
          activeTab === 'dashboard'
            ? 'bg-topbar-foreground text-topbar'
            : 'text-topbar-foreground/70 hover:text-topbar-foreground',
        ]"
        @click="emit('update:activeTab', 'dashboard')"
      >
        Dashboard
      </button>
      <button
        type="button"
        :class="[
          'px-4 py-1.5 text-sm font-medium rounded-sm transition-colors',
          activeTab === 'domains'
            ? 'bg-topbar-foreground text-topbar'
            : 'text-topbar-foreground/70 hover:text-topbar-foreground',
        ]"
        @click="emit('update:activeTab', 'domains')"
      >
        Domenii
      </button>
    </div>

    <Button variant="ghost" size="icon" class="text-topbar-foreground/70 hover:text-topbar-foreground hover:bg-topbar-foreground/10">
      <User class="h-5 w-5" />
    </Button>
  </header>
</template>
