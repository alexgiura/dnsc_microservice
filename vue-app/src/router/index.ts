import { createRouter, createWebHistory } from 'vue-router'
import Index from '@/views/Index.vue'
import NotFound from '@/views/NotFound.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'index', component: Index },
    { path: '/:pathMatch(.*)*', name: 'not-found', component: NotFound },
  ],
})
