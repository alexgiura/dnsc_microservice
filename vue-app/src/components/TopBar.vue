<script setup lang="ts">
import { User, Bell } from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import dnscLogo from '@/assets/dnsc-logo.svg'

const props = defineProps<{
  activeTab: 'dashboard' | 'domains'
}>()

const emit = defineEmits<{ 'update:activeTab': ['dashboard' | 'domains'] }>()

const handleTabChange = (tab: 'dashboard' | 'domains') => {
  emit('update:activeTab', tab)
}
</script>

<template>
  <header
    class="h-16 bg-topbar text-topbar-foreground flex items-center justify-between px-6 shrink-0"
  >
    <div class="flex items-center gap-2.5">
      <img :src="dnscLogo" alt="DNSC" class="h-9 w-9 rounded-full" />
      <span class="font-semibold text-base tracking-wide">DNSC</span>
    </div>

    <nav
      class="flex items-center gap-1 bg-topbar-foreground/[0.06] rounded-full p-1 border border-topbar-foreground/[0.08]"
    >
      <button
        type="button"
        class="px-6 py-2 text-[13px] font-medium rounded-full transition-all duration-200"
        :class="
          props.activeTab === 'dashboard'
            ? 'bg-primary text-primary-foreground shadow-md'
            : 'text-topbar-foreground/60 hover:text-topbar-foreground/90 hover:bg-topbar-foreground/[0.06]'
        "
        @click="handleTabChange('dashboard')"
      >
        Dashboard
      </button>
      <button
        type="button"
        class="px-6 py-2 text-[13px] font-medium rounded-full transition-all duration-200"
        :class="
          props.activeTab === 'domains'
            ? 'bg-primary text-primary-foreground shadow-md'
            : 'text-topbar-foreground/60 hover:text-topbar-foreground/90 hover:bg-topbar-foreground/[0.06]'
        "
        @click="handleTabChange('domains')"
      >
        Domenii
      </button>
    </nav>

    <div class="flex items-center gap-1.5">
      <Button
        variant="ghost"
        size="icon"
        class="relative rounded-full text-topbar-foreground/60 hover:text-topbar-foreground hover:bg-topbar-foreground/10"
      >
        <Bell class="h-[18px] w-[18px]" />
        <span
          class="absolute top-1.5 right-1.5 h-2 w-2 rounded-full bg-destructive ring-2 ring-topbar"
        />
      </Button>
      <Button
        variant="ghost"
        size="icon"
        class="rounded-full text-topbar-foreground/60 hover:text-topbar-foreground hover:bg-topbar-foreground/10"
      >
        <User class="h-[18px] w-[18px]" />
      </Button>
    </div>
  </header>
</template>
