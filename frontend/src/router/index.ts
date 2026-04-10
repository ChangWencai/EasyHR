import { createRouter, createWebHashHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/views/layout/AppLayout.vue'),
    redirect: '/home',
    children: [
      { path: '/home', name: 'home', component: () => import('@/views/home/HomeView.vue') },
      { path: '/employee', name: 'employee', component: () => import('@/views/employee/EmployeeList.vue') },
      { path: '/employee/create', name: 'employee-create', component: () => import('@/views/employee/EmployeeCreate.vue') },
      { path: '/employee/invitations', name: 'employee-invitations', component: () => import('@/views/employee/InvitationList.vue') },
      { path: '/employee/offboardings', name: 'employee-offboardings', component: () => import('@/views/employee/OffboardingList.vue') },
      { path: '/employee/:id', name: 'employee-detail', component: () => import('@/views/employee/EmployeeDetail.vue') },
      { path: '/employee/:id/edit', name: 'employee-edit', component: () => import('@/views/employee/EmployeeCreate.vue') },
      {
        path: '/tool',
        name: 'tool',
        component: () => import('@/views/tool/ToolHome.vue'),
      },
      {
        path: '/finance',
        name: 'finance',
        component: () => import('@/views/finance/FinanceHome.vue'),
      },
      {
        path: '/mine',
        name: 'mine',
        component: () => import('@/views/mine/MineView.vue'),
      },
    ],
  },
  {
    path: '/onboarding/org-setup',
    name: 'org-setup',
    component: () => import('@/views/onboarding/OrgSetup.vue'),
  },
  { path: '/login', name: 'login', component: () => import('@/views/layout/PlaceholderView.vue') },
]

export default createRouter({
  history: createWebHashHistory(),
  routes,
})
