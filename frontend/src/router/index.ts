import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'Dashboard',
      component: () => import('@/views/Dashboard.vue')
    },
    {
      path: '/sessions',
      name: 'Sessions',
      component: () => import('@/views/Sessions.vue')
    },
    {
      path: '/messages',
      name: 'Messages',
      component: () => import('@/views/Messages.vue')
    },
    {
      path: '/config',
      name: 'Config',
      component: () => import('@/views/Config.vue')
    }
  ]
})

export default router
