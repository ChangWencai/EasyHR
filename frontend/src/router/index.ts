import { createRouter, createWebHashHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/views/layout/AppLayout.vue'),
    redirect: '/home',
    children: [
      { path: '/home', name: 'home', component: () => import('@/views/home/HomeView.vue') },
      { path: '/employee', name: 'employee', component: () => import('@/views/layout/PlaceholderView.vue') },
      { path: '/tool', name: 'tool', component: () => import('@/views/layout/PlaceholderView.vue') },
      { path: '/finance', name: 'finance', component: () => import('@/views/layout/PlaceholderView.vue') },
      { path: '/mine', name: 'mine', component: () => import('@/views/layout/PlaceholderView.vue') },
    ],
  },
  { path: '/login', name: 'login', component: () => import('@/views/layout/PlaceholderView.vue') },
]

export default createRouter({
  history: createWebHashHistory(),
  routes,
})
