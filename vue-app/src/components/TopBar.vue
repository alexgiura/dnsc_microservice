<script setup lang="ts">
import { User, Bell, LogOut } from 'lucide-vue-next'
import { useRouter } from 'vue-router'
import Button from '@/components/ui/Button.vue'
import DropdownMenu from '@/components/ui/DropdownMenu.vue'
import dnscLogo from '@/assets/dnsc-logo.svg'
import { currentUser, logout as logoutUser } from '@/stores/auth'

const router = useRouter()

async function handleLogout() {
  await logoutUser()
  await router.push('/login')
}

export type TopBarTab = 'dashboard' | 'domains' | 'import'

const props = defineProps<{
  activeTab: TopBarTab
}>()

const emit = defineEmits<{ 'update:activeTab': [tab: TopBarTab] }>()

const tabs: { key: TopBarTab; label: string }[] = [
  { key: 'dashboard', label: 'Dashboard' },
  { key: 'domains', label: 'Domenii' },
  { key: 'import', label: 'Import' },
]

const handleTabChange = (tab: TopBarTab) => {
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
        v-for="tab in tabs"
        :key="tab.key"
        type="button"
        class="px-6 py-2 text-[13px] font-medium rounded-full transition-all duration-200"
        :class="
          props.activeTab === tab.key
            ? 'bg-primary text-primary-foreground shadow-md'
            : 'text-topbar-foreground/60 hover:text-topbar-foreground/90 hover:bg-topbar-foreground/[0.06]'
        "
        @click="handleTabChange(tab.key)"
      >
        {{ tab.label }}
      </button>
    </nav>

    <div class="flex items-center gap-1.5">
      <Button
        variant="ghost"
        size="icon"
        type="button"
        aria-label="Notificări"
        class="relative rounded-full text-topbar-foreground/60 hover:text-topbar-foreground hover:bg-topbar-foreground/10"
      >
        <Bell class="h-[18px] w-[18px]" />
        <span
          class="absolute top-1.5 right-1.5 h-2 w-2 rounded-full bg-destructive ring-2 ring-topbar"
        />
      </Button>

      <DropdownMenu>
        <template #trigger>
          <Button
            variant="ghost"
            type="button"
            aria-label="Cont și deconectare"
            class="h-10 rounded-full px-2.5 gap-2 text-topbar-foreground/80 hover:text-topbar-foreground hover:bg-topbar-foreground/10"
          >
            <User class="h-[18px] w-[18px] shrink-0 opacity-80" />
            <span class="text-sm font-medium max-w-[min(160px,28vw)] truncate">
              {{ currentUser?.username ?? '—' }}
            </span>
          </Button>
        </template>
        <template #content>
          <div class="min-w-[10rem] py-1">
            <button
              type="button"
              class="w-full flex items-center gap-2 px-3 py-2 text-sm text-foreground hover:bg-accent rounded-sm transition-colors"
              @click="handleLogout"
            >
              <LogOut class="h-4 w-4 shrink-0 opacity-70" />
              Deconectare
            </button>
          </div>
        </template>
      </DropdownMenu>
    </div>
  </header>
</template>
