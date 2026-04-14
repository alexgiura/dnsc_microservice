import { createRouter, createWebHistory } from 'vue-router'
import Index from '@/views/Index.vue'
import Login from '@/views/Login.vue'
import NotFound from '@/views/NotFound.vue'
import { currentUser, fetchMe } from '@/stores/auth'

let authBootstrapped = false

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: Login, meta: { public: true } },
    { path: '/', name: 'index', component: Index },
    { path: '/:pathMatch(.*)*', name: 'not-found', component: NotFound },
  ],
})

// După fetchMe: cu sesiune validă (cookie) → currentUser setat → acces la app.
// Fără sesiune → currentUser null → /login (LoginPage); de la /login cu user → redirect la /.
router.beforeEach(async (to) => {
  if (!authBootstrapped) {
    await fetchMe()
    authBootstrapped = true
  }

  if (to.name === 'login') {
    if (currentUser.value) {
      return { name: 'index' }
    }
    return true
  }

  if (!currentUser.value) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  return true
})
