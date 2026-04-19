import { createRouter, createWebHashHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/views/layout/AppLayout.vue'),
    redirect: '/home',
    children: [
      { path: '/home', name: 'home', component: () => import('@/views/home/HomeView.vue') },

      // 考勤管理
      {
        path: '/attendance/rule',
        name: 'attendance-rule',
        component: () => import('@/views/attendance/AttendanceRule.vue'),
      },
      {
        path: '/attendance/clock-live',
        name: 'attendance-clock-live',
        component: () => import('@/views/attendance/ClockLive.vue'),
      },
      {
        path: '/attendance/approval',
        name: 'attendance-approval',
        component: () => import('@/views/attendance/AttendanceApproval.vue'),
      },
      {
        path: '/attendance/monthly',
        name: 'attendance-monthly',
        component: () => import('@/views/attendance/AttendanceMonthly.vue'),
      },

      // 员工管理
      {
        path: '/employee/dashboard',
        name: 'employee-dashboard',
        component: () => import('@/views/employee/EmployeeDashboard.vue'),
      },
      {
        path: '/employee/org-chart',
        name: 'employee-org-chart',
        component: () => import('@/views/employee/OrgChart.vue'),
      },
      {
        path: '/employee/registrations',
        name: 'employee-registrations',
        component: () => import('@/views/employee/RegistrationList.vue'),
      },
      {
        path: '/employee',
        name: 'employee',
        component: () => import('@/views/employee/EmployeeList.vue'),
      },
      {
        path: '/employee/create',
        name: 'employee-create',
        component: () => import('@/views/employee/EmployeeCreate.vue'),
      },
      {
        path: '/employee/invitations',
        name: 'employee-invitations',
        component: () => import('@/views/employee/InvitationList.vue'),
      },
      {
        path: '/employee/offboardings',
        name: 'employee-offboardings',
        component: () => import('@/views/employee/OffboardingList.vue'),
      },
      {
        path: '/employee/:id',
        name: 'employee-detail',
        component: () => import('@/views/employee/EmployeeDetail.vue'),
      },

      // 工具
      {
        path: '/tool',
        name: 'tool-overview',
        component: () => import('@/views/tool/ToolOverview.vue'),
      },
      {
        path: '/tool/salary',
        name: 'tool-salary',
        component: () => import('@/views/tool/SalaryTool.vue'),
      },
      {
        path: '/tool/salary/dashboard',
        name: 'tool-salary-dashboard',
        component: () => import('@/views/tool/SalaryDashboard.vue'),
      },
      {
        path: '/tool/salary/slip-send',
        name: 'tool-salary-slip-send',
        component: () => import('@/views/tool/SalarySlipSend.vue'),
      },
      {
        path: '/tool/salary/tax-upload',
        name: 'tool-salary-tax-upload',
        component: () => import('@/views/tool/TaxUpload.vue'),
      },
      {
        path: '/tool/socialinsurance',
        name: 'tool-socialinsurance',
        component: () => import('@/views/tool/SITool.vue'),
      },
      {
        path: '/tool/tax',
        name: 'tool-tax',
        component: () => import('@/views/tool/TaxTool.vue'),
      },

      // 财务
      {
        path: '/finance',
        name: 'finance',
        redirect: '/finance/vouchers',
      },
      {
        path: '/finance/accounts',
        name: 'finance-accounts',
        component: () => import('@/views/finance/AccountTree.vue'),
      },
      {
        path: '/finance/vouchers',
        name: 'finance-vouchers',
        component: () => import('@/views/finance/VoucherList.vue'),
      },
      {
        path: '/finance/vouchers/create',
        name: 'finance-voucher-create',
        component: () => import('@/views/finance/VoucherCreate.vue'),
      },
      {
        path: '/finance/invoices',
        name: 'finance-invoices',
        component: () => import('@/views/finance/InvoiceList.vue'),
      },
      {
        path: '/finance/expenses',
        name: 'finance-expenses',
        component: () => import('@/views/finance/ExpenseApproval.vue'),
      },
      {
        path: '/finance/reports',
        name: 'finance-reports',
        component: () => import('@/views/finance/BookReport.vue'),
      },

      // 我的
      {
        path: '/mine',
        name: 'mine',
        component: () => import('@/views/mine/MineView.vue'),
      },

      // 待办中心
      {
        path: '/todo',
        name: 'todo-list',
        component: () => import('@/views/todo/TodoListView.vue'),
      },
      {
        path: '/carousel/manage',
        name: 'carousel-manage',
        component: () => import('@/views/todo/CarouselManagePage.vue'),
      },
    ],
  },

  // 独立页面（不走 AppLayout）
  {
    path: '/onboarding/org-setup',
    name: 'org-setup',
    component: () => import('@/views/onboarding/OrgSetup.vue'),
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/layout/LoginView.vue'),
  },
  {
    path: '/register/:token',
    name: 'register',
    component: () => import('@/views/employee/RegisterPage.vue'),
  },
  // H5 工资条查看（无需登录）
  {
    path: '/salary/slip/:token',
    name: 'salary-slip-h5',
    component: () => import('@/views/tool/SalarySlipH5.vue'),
  },
  // 协办填写页（无需登录）
  {
    path: '/todo/:id/invite',
    name: 'todo-invite',
    component: () => import('@/views/todo/InviteFillPage.vue'),
  },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

// Auth Guard：未登录访问受保护路由时重定向到 /login
router.beforeEach((to, _from) => {
  const authStore = useAuthStore()

  // /login, /onboarding/org-setup, /register, /salary/slip/ and /todo/*/invite 不做守卫检查
  if (to.path === '/login' || to.path === '/onboarding/org-setup' || to.path.startsWith('/register') || to.path.startsWith('/salary/slip/') || to.path.includes('/invite')) {
    return
  }

  // 受保护路由：/home, /employee, /tool, /finance, /attendance, /mine
  const isProtectedRoute =
    to.path.startsWith('/home') ||
    to.path.startsWith('/employee') ||
    to.path.startsWith('/tool') ||
    to.path.startsWith('/finance') ||
    to.path.startsWith('/attendance') ||
    to.path.startsWith('/mine') ||
    to.path.startsWith('/todo') ||
    to.path.startsWith('/carousel')

  if (isProtectedRoute && !authStore.isLoggedIn) {
    return { path: '/login' }
  }
})

export default router
