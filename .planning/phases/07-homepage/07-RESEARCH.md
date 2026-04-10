# Phase 7: 首页工作台 - Research

**Gathered:** 2026-04-10
**Status:** Complete

## User Constraints (from CONTEXT.md)

All decisions below are already locked by context or prior phases — do NOT revisit:

| ID | Decision | Source |
|----|----------|--------|
| D-01 | H5 管理后台先行，APP 原生 V2.0 实现 | CONTEXT D-01 |
| D-02 | 前端项目建于 `frontend/` 目录，独立部署 | CONTEXT D-02 |
| D-03 | Vue 3 + Element Plus + Vite + TypeScript + Pinia + Vue Router | CONTEXT D-03 |
| D-04 | 5 Tab 底部导航（首页/员工/工具/财务/我的） | CONTEXT D-04 |
| D-05 | 首页布局：顶部栏 → 待办卡片 → 5宫格 → 数据概览 → Tab栏 | CONTEXT D-05 |
| D-06 | 6张待办卡片，优先级顺序锁定 | CONTEXT D-06 |
| D-07 | 卡片点击后跳转，完成后自动消失 | CONTEXT D-07 |
| D-08 | 无待办时显示空状态组件 | CONTEXT D-08 |
| D-09 | 5宫格入口：员工/社保/工资/个税/财务 | CONTEXT D-09 |
| D-10 | Element Plus 内置图标（ep:*），主色 #1677FF | CONTEXT D-10 |
| D-11 | 数据概览：员工数/本月入离职/社保总额/工资总额 | CONTEXT D-11 |
| D-12 | 数据概览可折叠/展开 | CONTEXT D-12 |
| D-13 | 数据仅进入时加载一次，无轮询 | CONTEXT D-13 |
| D-14 | Tab 栏 5 个 tab 固定 | CONTEXT D-14 |
| D-15 | Tab 图标+文字双标识，首页默认选中 | CONTEXT D-15 |
| D-16 | Go 新建 `internal/dashboard/` 模块 | CONTEXT D-16 |
| D-17 | DashboardService 聚合各模块待办 + 数据概览 | CONTEXT D-17 |
| D-18 | org_id 从 JWT middleware 注入，无需前端传参 | CONTEXT D-18 |
| D-19 | API: `GET /api/v1/dashboard` 返回 todos + overview | CONTEXT D-19 |
| D-20 | UI 规范遵循 ui-ux.md §1.3 | CONTEXT D-20 |

---

## Research Area 1: Vue 3 H5 Project Setup

### Stack Verification (from CLAUDE.md + package.json)

| 包 | 版本 | 来源 |
|----|------|------|
| Vue | 3.5.32 | CLAUDE.md |
| Element Plus | 2.13.6 | CLAUDE.md |
| Vite | 8.0.3 | CLAUDE.md |
| TypeScript | 6.0.2 | CLAUDE.md |
| Pinia | 3.0.4 | CLAUDE.md |
| Vue Router | 5.0.4 | CLAUDE.md |
| @vueuse/core | 14.2.1 | CLAUDE.md |
| Axios | 1.14.0 | CLAUDE.md |
| dayjs | 1.11.20 | CLAUDE.md |
| ECharts | 6.0.0 | CLAUDE.md |
| xlsx | 0.18.5 | CLAUDE.md |
| ESLint | 10.2.0 | CLAUDE.md |
| Prettier | 3.8.1 | CLAUDE.md |
| Sass | 1.99.0 | CLAUDE.md |
| unplugin-auto-import | 21.0.0 | CLAUDE.md |
| unplugin-vue-components | 32.0.0 | CLAUDE.md |

### Project Initialization Commands

```bash
# Step 1: 创建 Vite + Vue3 + TS 项目
cd /Users/wencai/github/EasyHR
npm create vite@latest frontend -- --template vue-ts

# Step 2: 进入目录安装依赖
cd frontend
npm install vue vue-router@5 pinia axios dayjs

# Element Plus + icons
npm install element-plus @element-plus/icons-vue
npm install -D unplugin-auto-import unplugin-vue-components

# Dev tools
npm install -D sass eslint prettier eslint-plugin-vue

# Step 3: 配置 vite.config.ts (关键配置)
# 需配置: auto-import, vue-components, proxy到后端API
```

### Vite Config Key Points

```typescript
// vite.config.ts — 必须配置项
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

export default defineConfig({
  plugins: [
    vue(),
    AutoImport({
      resolvers: [ElementPlusResolver()],
      imports: ['vue', 'vue-router', 'pinia'],
      dts: 'src/auto-imports.d.ts',
    }),
    Components({
      resolvers: [ElementPlusResolver()],
      dts: 'src/components.d.ts',
    }),
  ],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
```

### Mobile Viewport Setup

```html
<!-- index.html — 必须的移动端视口 -->
<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no, viewport-fit=cover">
<!-- iOS 安全区域 -->
<meta name="apple-mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-status-bar-style" content="default">
<!-- 去除电话/email/地址识别 -->
<meta name="format-detection" content="telephone=no">
```

### Recommended Directory Structure

```
frontend/src/
├── api/               # Axios 实例 + API 模块
│   ├── index.ts       # axios 实例（baseURL, interceptor）
│   ├── auth.ts        # 认证相关 API
│   └── dashboard.ts   # Dashboard API
├── assets/            # 静态资源
├── components/        # 公共组件
│   └── common/        # 通用组件
├── router/
│   └── index.ts       # Vue Router 配置 + 路由守卫
├── stores/             # Pinia stores
│   ├── index.ts       # store 初始化
│   ├── auth.ts        # 认证状态（token, user info）
│   └── dashboard.ts   # Dashboard 数据
├── styles/
│   ├── variables.scss # Element Plus 主题变量覆盖
│   └── global.scss    # 全局样式
├── utils/
│   └── request.ts    # Axios 封装（与 api/index.ts 合并）
├── views/              # 页面组件
│   ├── layout/
│   │   └── AppLayout.vue    # 底部 Tab 容器布局
│   ├── home/
│   │   └── HomeView.vue     # 首页工作台
│   ├── employee/
│   ├── tool/
│   ├── finance/
│   └── mine/
├── App.vue
└── main.ts
```

---

## Research Area 2: Element Plus Mobile Adaptation

### iOS Safe Area Handling

CSS `env(safe-area-inset-*)` 支持 iOS 刘海屏/圆角屏底部导航栏：

```scss
// styles/global.scss
:root {
  --safe-area-bottom: env(safe-area-inset-bottom, 0px);
  --tab-bar-height: calc(56px + var(--safe-area-bottom));
}

.tab-bar {
  padding-bottom: var(--safe-area-bottom);
  height: var(--tab-bar-height);
}
```

### Element Plus Mobile Considerations

- **el-card**: 待办卡片使用，添加 `shadow="hover"` 提升交互感
- **el-grid**: 5宫格使用 `el-row` + `el-col :span="8"` (3列布局)
- **el-icon**: 使用 `<component :is="'IconName'" />` 动态渲染
- **el-badge**: 待办数量角标（卡片右上角）
- **el-collapse**: 数据概览折叠使用 `el-collapse`
- **el-button**: 主色 `#1677FF` via CSS 变量覆盖
- **el-tag**: 待办优先级用不同颜色 tag 标识

### CSS Variable Overrides

```scss
// styles/variables.scss
:root {
  --el-color-primary: #1677FF;
  --el-color-success: #52C41A;
  --el-color-danger: #FF4D4F;
  --el-font-size-base: 14px;
  --el-font-size-medium: 16px;
  --el-font-size-large: 18px;
}
```

---

## Research Area 3: Pinia + Dashboard State

### Auth Store

从 Phase 1 JWT 实现推断：

```typescript
// stores/auth.ts
interface AuthState {
  token: string | null
  userId: number
  orgId: number
  role: 'OWNER' | 'ADMIN' | 'MEMBER'
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    token: localStorage.getItem('token'),
    userId: 0,
    orgId: 0,
    role: 'MEMBER',
  }),
  getters: {
    isLoggedIn: (state) => !!state.token,
  },
  actions: {
    setToken(token: string) {
      this.token = token
      localStorage.setItem('token', token)
    },
    logout() {
      this.token = null
      localStorage.removeItem('token')
    },
  },
})
```

### Dashboard Store

```typescript
// stores/dashboard.ts
import { defineStore } from 'pinia'
import { fetchDashboard } from '@/api/dashboard'

export interface TodoItem {
  type: string
  title: string
  count: number
  deadline?: string
  priority: number
}

export interface DashboardOverview {
  employee_count: number
  joined_this_month: number
  left_this_month: number
  social_insurance_total: string
  payroll_total: string
}

export interface DashboardState {
  todos: TodoItem[]
  overview: DashboardOverview | null
  loading: boolean
  overviewExpanded: boolean
}

export const useDashboardStore = defineStore('dashboard', {
  state: (): DashboardState => ({
    todos: [],
    overview: null,
    loading: false,
    overviewExpanded: true,
  }),
  actions: {
    async load() {
      this.loading = true
      try {
        const data = await fetchDashboard()
        this.todos = data.todos
        this.overview = data.overview
      } finally {
        this.loading = false
      }
    },
    toggleOverview() {
      this.overviewExpanded = !this.overviewExpanded
    },
    removeTodo(type: string) {
      this.todos = this.todos.filter(t => t.type !== type)
    },
  },
})
```

### Dashboard API Module

```typescript
// api/dashboard.ts
import axios from '@/utils/request'  // 封装的 axios 实例

export interface DashboardResponse {
  todos: Array<{
    type: string
    title: string
    count: number
    deadline?: string
    priority: number
  }>
  overview: {
    employee_count: number
    joined_this_month: number
    left_this_month: number
    social_insurance_total: string
    payroll_total: string
  }
}

export function fetchDashboard(): Promise<DashboardResponse> {
  return axios.get('/api/v1/dashboard')
}
```

---

## Research Area 4: Go Dashboard Service

### Package Structure

```
internal/
├── dashboard/
│   ├── handler.go       # HTTP handler
│   ├── service.go       # DashboardService（聚合逻辑）
│   ├── repository.go    # 数据查询
│   ├── router.go        # 路由注册
│   └── model.go         # 数据结构
```

### DashboardService Design

关键设计：聚合 Phase 1-6 各模块的待办和概览数据。

```go
// internal/dashboard/service.go

type DashboardService struct {
    db              *gorm.DB
    employeeRepo    *employee.Repository
    socialRepo      *social.Repository
    payrollRepo     *payroll.Repository
    financeRepo     *finance.Repository
}

type DashboardResult struct {
    Todos    []TodoItem
    Overview Overview
}

type TodoItem struct {
    Type      string  `json:"type"`
    Title     string  `json:"title"`
    Count     int     `json:"count"`
    Deadline  string  `json:"deadline,omitempty"`
    Priority  int     `json:"priority"`
}

type Overview struct {
    EmployeeCount         int    `json:"employee_count"`
    JoinedThisMonth       int    `json:"joined_this_month"`
    LeftThisMonth         int    `json:"left_this_month"`
    SocialInsuranceTotal  string `json:"social_insurance_total"`
    PayrollTotal          string `json:"payroll_total"`
}

// GetDashboard returns aggregated dashboard for the authenticated org.
func (s *DashboardService) GetDashboard(ctx context.Context, orgID uint) (*DashboardResult, error) {
    // 1. 并行获取各模块数据（goroutine）
    // 2. 按优先级排序合并待办
    // 3. 返回聚合结果
}
```

### Priority Order (from D-06)

| Priority | Type | Title | Source |
|----------|------|-------|--------|
| 1 | social_insurance | 社保缴费提醒 | SOCL-03 |
| 2 | tax | 个税申报提醒 | TAX-03 |
| 3 | employee | 员工入离职待审核 | EMPL-01/05 |
| 4 | contract | 合同到期提醒 | EMPL-08 |
| 5 | expense | 费用报销待审批 | FINC-09 |
| 6 | voucher | 凭证待审核 | FINC-03 |

### Handler

```go
// internal/dashboard/handler.go
func RegisterDashboardRouter(r *gin.RouterGroup, svc *DashboardService, authMw gin.HandlerFunc) {
    r := r.Group("/dashboard")
    r.Use(authMw)
    r.GET("", svc.GetDashboard)
}

// JWT org_id extraction (参考 Phase 1 middleware 模式)
// claims 中的 org_id 通过 TenantScope middleware 注入
```

---

## Research Area 5: Bottom Tab Navigation

### Vue Router + Tab Bar Integration

```vue
<!-- views/layout/AppLayout.vue -->
<template>
  <div class="app-container">
    <div class="main-content">
      <router-view />
    </div>
    <van-tabbar
      v-model="active"
      :fixed="true"
      :placeholder="true"
      route
      active-color="#1677FF"
      inactive-color="#999"
    >
      <van-tabbar-item
        v-for="tab in tabs"
        :key="tab.path"
        :to="tab.path"
        :icon="tab.icon"
      >
        {{ tab.label }}
      </van-tabbar-item>
    </van-tabbar>
  </div>
</template>
```

**Note:** 虽然 CLAUDE.md 推荐使用 Vue + Element Plus，但移动端 Tab 导航用 **Vant 4** 的 Tabbar 组件体验更好。需评估：
- 方案A：纯 Element Plus 底部导航（`el-menu` 无法做底部固定）
- 方案B：使用 Vant 4 的 Tabbar 组件（移动端体验最佳）
- 方案C：自研 CSS 固定底部导航栏

**Recommendation:** Phase 7 使用 **方案C（自研 CSS TabBar）**，保持技术栈一致性（只用 Element Plus），避免引入 Vant 依赖增加包体积：

```vue
<!-- 自研底部 Tab 栏 -->
<template>
  <div class="tab-bar-wrapper">
    <router-link
      v-for="tab in tabs"
      :key="tab.path"
      :to="tab.path"
      class="tab-item"
      :class="{ active: route.path.startsWith(tab.path) }"
    >
      <el-icon><component :is="tab.icon" /></el-icon>
      <span>{{ tab.label }}</span>
    </router-link>
  </div>
</template>

<style scoped>
.tab-bar-wrapper {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  background: #fff;
  border-top: 1px solid #eee;
  padding-bottom: env(safe-area-inset-bottom);
  z-index: 1000;
}
.tab-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 8px 0;
  text-decoration: none;
  color: #999;
  font-size: 10px;
}
.tab-item.active {
  color: #1677FF;
}
</style>
```

### Router Configuration

```typescript
// router/index.ts
import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/views/layout/AppLayout.vue'),
    redirect: '/home',
    children: [
      { path: '/home', component: () => import('@/views/home/HomeView.vue') },
      { path: '/employee', component: () => import('@/views/employee/EmployeeListView.vue') },
      { path: '/tool', component: () => import('@/views/tool/ToolView.vue') },
      { path: '/finance', component: () => import('@/views/finance/FinanceView.vue') },
      { path: '/mine', component: () => import('@/views/mine/MineView.vue') },
    ],
  },
]
```

---

## Research Area 6: Axios + JWT Interceptor

### Request Interceptor (Token Injection)

```typescript
// api/request.ts
import axios, { AxiosError } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 10000,
})

// 请求拦截：注入 JWT
request.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截：401 处理
request.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      ElMessage.error('登录已过期，请重新登录')
      router.push('/login')
    }
    return Promise.reject(error)
  }
)

export default request
```

---

## Research Area 7: Validation Architecture

### Go Backend Verification

1. **Build**: `cd backend && go build ./...` 无编译错误
2. **Unit Tests**: `internal/dashboard/` service 有单元测试覆盖聚合逻辑
3. **API Integration**: `curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/dashboard`
4. **JWT auth**: 返回 401 如果无 token 或 token 无效

### Frontend Verification

1. **Build**: `cd frontend && npm run build` 无错误
2. **Dev Server**: `npm run dev` 启动正常
3. **TypeScript**: `npm run type-check` 无类型错误
4. **ESLint**: `npm run lint` 无错误
5. **Runtime**: 首页加载后 dashboard API 调用成功，数据正确展示

### Cross-Phase Integration Points

| 模块 | 依赖 Phase | 数据来源 |
|------|-----------|---------|
| 员工待办 | Phase 2 | employee module (待审核入离职) |
| 社保待办 | Phase 3 | social_insurance module (缴费提醒) |
| 个税待办 | Phase 4 | tax module (申报提醒) |
| 工资数据 | Phase 5 | payroll module (本月工资总额) |
| 报销待办 | Phase 8 (小程序) | expense module (待审批报销) |
| 凭证待办 | Phase 6 | voucher module (待审核凭证) |

---

## Implementation Risk Assessment

| 风险 | 级别 | 缓解措施 |
|------|------|---------|
| Phase 3-6 尚未执行，dashboard 调用的接口可能不存在 | 中 | P07 plan 中设计好接口签名，Phase 3-6 执行时按签名实现 |
| Element Plus 移动端体验不如 Vant | 低 | Phase 7 聚焦功能，Vant 迁移 V2.0 考虑 |
| H5 认证状态与 Go JWT 同步 | 中 | 统一使用 Bearer token，前端 localStorage 存储 |
| 多模块并行查询 DB 性能 | 低 | DashboardService 并发 goroutine 聚合，DB 已有索引 |

---

## Validation Architecture

**Requirement:** Every phase with nyquist_validation=true needs a VALIDATION.md with 8 dimensions.

For Phase 7, the validation dimensions are:

| Dim | 内容 | 验证方法 |
|-----|------|---------|
| 1 | 功能正确性 | Dashboard API 返回正确 todos + overview 数据 |
| 2 | 数据完整性 | 所有 6 种待办类型都有数据，优先级正确排序 |
| 3 | 移动端适配 | 视口配置正确，底部 Tab 在各机型正常显示 |
| 4 | 认证集成 | JWT token 正确注入，401 时跳转登录页 |
| 5 | 路由导航 | 5 个 Tab 切换正常，首页默认选中 |
| 6 | 状态管理 | Pinia store 正确管理 dashboard 数据和展开状态 |
| 7 | 错误处理 | API 失败时显示友好错误提示，不是白屏 |
| 8 | 性能 | Dashboard API 响应 ≤ 500ms，无不必要重复请求 |

