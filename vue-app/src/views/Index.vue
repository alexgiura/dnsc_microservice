<script setup lang="ts">
import { ref } from 'vue'
import TopBar from '@/components/TopBar.vue'
import DomainTable from '@/components/DomainTable.vue'
import Dashboard from '@/components/Dashboard.vue'
import ImportPage from '@/components/ImportPage.vue'

const activeTab = ref<'dashboard' | 'domains' | 'import'>('domains')
</script>

<template>
  <div class="flex flex-col min-h-screen bg-background">
    <TopBar v-model:active-tab="activeTab" />
    <main class="flex-1 p-6 max-w-7xl mx-auto w-full">
      <Dashboard v-if="activeTab === 'dashboard'" />
      <DomainTable v-else-if="activeTab === 'domains'" />
      <ImportPage v-else-if="activeTab === 'import'" />
    </main>
  </div>
</template>
