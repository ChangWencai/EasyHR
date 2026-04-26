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
      {
        path: '/attendance/overtime',
        name: 'attendance-overtime',
        component: () => import('@/views/attendance/ComplianceOvertime.vue'),
      },
      {
        path: '/attendance/leave',
        name: 'attendance-leave',
        component: () => import('@/views/attendance/ComplianceLeave.vue'),
      },
      {
        path: '/attendance/anomaly',
        name: 'attendance-anomaly',
        component: () => import('@/views/attendance/ComplianceAnomaly.vue'),
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
        path: '/employee/positions',
        name: 'employee-positions',
        component: () => import('@/views/employee/PositionManage.vue'),
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
      {
        path: '/employee/:id/edit',
        name: 'employee-edit',
        component: () => import('@/views/employee/EmployeeCreate.vue'),
      },

      // 薪资管理
      {
        path: '/salary',
        name: 'salary',
        component: () => import('@/views/tool/SalaryTool.vue'),
      },
      {
        path: '/salary/dashboard',
        name: 'salary-dashboard',
        component: () => import('@/views/tool/SalaryDashboard.vue'),
      },
      {
        path: '/salary/slip-send',
        name: 'salary-slip-send',
        component: () => import('@/views/tool/SalarySlipSend.vue'),
      },
      {
        path: '/salary/tax-upload',
        name: 'salary-tax-upload',
        component: () => import('@/views/tool/TaxUpload.vue'),
      },
      {
        path: '/salary/tax',
        name: 'salary-tax',
        component: () => import('@/views/tool/TaxTool.vue'),
      },

      // 社保管理
      {
        path: '/social-insurance',
        name: 'social-insurance',
        component: () => import('@/views/tool/SITool.vue'),
      },

      // 工具（保留旧路由兼容）
      {
        path: '/tool',
        redirect: '/salary',
      },
      {
        path: '/tool/salary',
        redirect: '/salary',
      },
      {
        path: '/tool/salary/dashboard',
        redirect: '/salary/dashboard',
      },
      {
        path: '/tool/salary/slip-send',
        redirect: '/salary/slip-send',
      },
      {
        path: '/tool/salary/tax-upload',
        redirect: '/salary/tax-upload',
      },
      {
        path: '/tool/socialinsurance',
        redirect: '/social-insurance',
      },
      {
        path: '/tool/email-templates',
        redirect: '/hr-tools/email-templates',
      },
      {
        path: '/tool/tax',
        redirect: '/salary/tax',
      },

      // 人事工具
      {
        path: '/hr-tools/email-templates',
        name: 'hr-tools-email-templates',
        component: () => import('@/views/tool/EmailTemplateList.vue'),
      },
      {
        path: '/hr-tools/sms-templates',
        name: 'hr-tools-sms-templates',
        component: () => import('@/views/tool/SmsTemplateList.vue'),
      },
      {
        path: '/hr-tools/todo',
        name: 'hr-tools-todo',
        component: () => import('@/views/todo/TodoListView.vue'),
      },
      {
        path: '/hr-tools/performance',
        name: 'hr-tools-performance',
        component: () => import('@/views/tool/PerformanceCoefficient.vue'),
      },

      // 旧路由兼容
      {
        path: '/todo',
        redirect: '/hr-tools/todo',
      },
      {
        path: '/carousel/manage',
        name: 'carousel-manage',
        component: () => import('@/views/todo/CarouselManagePage.vue'),
      },

      // 财务
      {
        path: '/finance',
        name: 'finance',
        redirect: '/finance/vouchers',
      },
      {
        path: '/finance/overview',
        name: 'finance-overview',
        component: () => import('@/views/finance/FinanceOverview.vue'),
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
        path: '/finance/accounts',
        name: 'finance-accounts',
        component: () => import('@/views/finance/AccountTree.vue'),
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
  // 合同签署页（员工端，无需登录）
  {
    path: '/sign/:contractId',
    name: 'contract-sign',
    component: () => import('@/views/sign/SignPage.vue'),
  },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

// Auth Guard：未登录访问受保护路由时重定向到 /login
router.beforeEach((to, _from) => {
  const authStore = useAuthStore()

  // /login, /onboarding/org-setup, /register, /salary/slip/, /todo/*/invite, /sign/ 不做守卫检查
  if (
    to.path === '/login' ||
    to.path === '/onboarding/org-setup' ||
    to.path.startsWith('/register') ||
    to.path.startsWith('/salary/slip/') ||
    to.path.includes('/invite') ||
    to.path.startsWith('/sign/')
  ) {
    return
  }

  // 受保护路由：/home, /employee, /salary, /social-insurance, /hr-tools, /finance, /attendance, /mine
  const isProtectedRoute =
    to.path.startsWith('/home') ||
    to.path.startsWith('/employee') ||
    to.path.startsWith('/salary') ||
    to.path.startsWith('/social-insurance') ||
    to.path.startsWith('/hr-tools') ||
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
