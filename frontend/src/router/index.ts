import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'Dashboard',
      component: () => import('@/views/Dashboard.vue'),
      meta: { public: true }
    },
    {
      path: '/sessions',
      name: 'Sessions',
      component: () => import('@/views/Sessions.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/messages',
      name: 'Messages',
      component: () => import('@/views/Messages.vue'),
      meta: { public: true }
    },
    {
      path: '/config',
      name: 'Config',
      component: () => import('@/views/Config.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/send',
      name: 'Send',
      component: () => import('@/views/Send.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue'),
      meta: { public: true }
    }
  ]
})

// Navigation guard
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next({ name: 'Login', query: { redirect: to.fullPath } })
  } else if (to.name === 'Login' && authStore.isAuthenticated) {
    next({ name: 'Dashboard' })
  } else {
    next()
  }
})

export default router
