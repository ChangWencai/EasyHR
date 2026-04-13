# H5 管理后台 - 页面模块与路由

> 本文档描述 易人事 H5 管理后台（Vue 3 + Element Plus）的全部页面模块、路由结构和功能说明。
>
> 最后更新：2026-04-11

---

## 一、整体架构

```
AppLayout（根布局）
├── Sidebar（桌面端左侧导航栏 / 移动端抽屉）
└── MainContent（主内容区）
    ├── HomeView（首页 / 工作台）
    ├── EmployeeModule（员工管理）
    │   ├── EmployeeList
    │   ├── EmployeeCreate / EmployeeEdit
    │   ├── EmployeeDetail
    │   ├── InvitationList
    │   └── OffboardingList
    ├── ToolModule（人事工具）
    │   ├── ToolHome（子菜单）
    │   ├── SalaryTool
    │   ├── SITool
    │   └── TaxTool
    ├── FinanceModule（财务记账）
    │   ├── FinanceHome（子菜单）
    │   ├── AccountTree
    │   ├── VoucherList
    │   ├── VoucherCreate
    │   ├── InvoiceList
    │   ├── ExpenseApproval
    │   └── BookReport
    └── MineView（我的 / 个人中心）
```

---

## 二、路由表

### 2.1 首页

| 路径 | 组件 | 说明 |
|------|------|------|
| `/home` | `HomeView.vue` | 工作台首页，含待办卡片、快捷入口、数据概览 |

### 2.2 员工管理

| 路径 | 组件 | 说明 |
|------|------|------|
| `/employee` | `EmployeeList.vue` | 员工列表，支持搜索/筛选/分页 |
| `/employee/create` | `EmployeeCreate.vue` | 新增员工（新建模式） |
| `/employee/:id/edit` | `EmployeeCreate.vue` | 编辑员工（复用新建组件，传入编辑模式） |
| `/employee/:id` | `EmployeeDetail.vue` | 员工详情，含档案/合同/社保/工资等信息 |
| `/employee/invitations` | `InvitationList.vue` | 入职邀请管理，发送邀请链接/查看邀请状态 |
| `/employee/offboardings` | `OffboardingList.vue` | 离职管理，提交离职申请/老板审批 |

### 2.3 人事工具

| 路径 | 组件 | 说明 |
|------|------|------|
| `/tool` | `ToolHome.vue` | 工具首页（三栏布局：概览/薪资/社保/个税） |
| `/tool/salary` | `SalaryTool.vue` | 薪资管理，薪资结构配置、工资核算 |
| `/tool/socialinsurance` | `SITool.vue` | 社保管理，参保城市基数配置 |
| `/tool/tax` | `TaxTool.vue` | 个税申报，专项附加扣除、申报管理 |

### 2.4 财务记账

| 路径 | 组件 | 说明 |
|------|------|------|
| `/finance` | `FinanceHome.vue` | 财务首页（重定向到 `/finance/vouchers`） |
| `/finance/vouchers` | `VoucherList.vue` | 凭证列表，支持按期间/凭证号筛选 |
| `/finance/vouchers/create` | `VoucherCreate.vue` | 填制凭证，录入借方/贷方分录 |
| `/finance/accounts` | `AccountTree.vue` | 科目管理，会计科目体系树形结构 |
| `/finance/invoices` | `InvoiceList.vue` | 发票管理，销项/进项发票登记 |
| `/finance/expenses` | `ExpenseApproval.vue` | 费用报销，提交/审批报销单 |
| `/finance/reports` | `BookReport.vue` | 账簿报表，科目余额表/财务报表/期间结账 |

### 2.5 我的

| 路径 | 组件 | 说明 |
|------|------|------|
| `/mine` | `MineView.vue` | 个人中心，企业信息/账号设置/退出登录 |

### 2.6 独立页面（不经过 AppLayout）

| 路径 | 组件 | 说明 |
|------|------|------|
| `/login` | `PlaceholderView.vue` | 登录页（目前由后端 API 驱动） |
| `/onboarding/org-setup` | `OrgSetup.vue` | 企业初始化配置（首次注册引导） |

---

## 三、布局结构

### 3.1 桌面端（≥ 768px）

```
┌──────────┬────────────────────────────────────────┐
│          │  [页面标题区]                            │
│  侧边栏   │  [内容区：表格/表单/卡片/图表]           │
│  220px   │                                        │
│  可折叠   │                                        │
│  64px    │                                        │
└──────────┴────────────────────────────────────────┘
```

- **侧边栏**：`AppLayout.vue`，固定宽度 220px，可折叠至 64px（仅图标）
- **子菜单**：财务和工具模块内有二级侧边菜单（宽度 180px，sticky）
- **内容区**：自适应剩余宽度，padding 20px 24px

### 3.2 移动端（< 768px）

```
┌──────────────────────────────────┐
│ [≡]    页面标题           [图标] │  ← 顶部栏
├──────────────────────────────────┤
│                                  │
│         内容区（全宽）             │
│                                  │
│                                  │
└──────────────────────────────────┘
```

- 汉堡菜单触发 `el-drawer` 抽屉导航
- 子菜单区域改为 `el-tabs` 标签栏

---

## 四、响应式断点

| 断点 | 宽度 | 布局变化 |
|------|------|----------|
| 4K | ≥ 2560px | 内容全宽，快捷入口 6 列，待办卡片 4 列 |
| 桌面 | 1200px - 2560px | 内容全宽，快捷入口 6 列，待办卡片 3-4 列 |
| 笔记本 | 900px - 1200px | 快捷入口/数据概览上下分行 |
| iPad | 768px - 900px | 待办卡片 2 列，快捷入口 3 列 |
| 手机 | < 768px | 移动端布局，汉堡菜单，底部无固定导航 |

---

## 五、API 模块对应

| 前端页面 | API 模块 | 主要接口 |
|----------|----------|----------|
| `HomeView` | `dashboard` | `GET /dashboard` |
| `EmployeeList/Create/Detail` | `employee` | `GET/POST /employees`, `GET/PUT /employees/:id` |
| `InvitationList` | `employee` | `POST /invitations`, `GET /invitations` |
| `OffboardingList` | `employee` | `POST /offboardings`, `PUT /offboardings/:id` |
| `SalaryTool` | `salary` | `GET/POST /salary-structures`, `POST /payroll/calculate` |
| `SITool` | `socialinsurance` | `GET /social-policies`, `POST /social-enrollments` |
| `TaxTool` | `tax` | `GET /tax-brackets`, `GET /special-deductions` |
| `VoucherList/Create` | `finance` | `GET/POST /vouchers` |
| `AccountTree` | `finance` | `GET/POST /accounts` |
| `InvoiceList` | `finance` | `GET/POST /invoices` |
| `ExpenseApproval` | `finance` | `GET /expenses`, `PUT /expenses/:id` |
| `BookReport` | `finance` | `GET /books/trial-balance`, `GET /books/ledger` |
| `OrgSetup` | `auth` | `POST /org-setup` |

---

## 六、技术栈

- **框架**：Vue 3.5 + TypeScript 6 + Vite 8
- **UI 组件**：Element Plus 2.13
- **状态管理**：Pinia 3.0（`useAuthStore`、`useDashboardStore`、`useUserStore`）
- **路由**：Vue Router 5（Hash 模式，`createWebHashHistory`）
- **HTTP**：Axios 1.14，统一 `request.ts` 实例（JWT 拦截器、401 重定向）
- **样式**：SCSS，Element Plus 变量覆盖，响应式断点 via `@media`

---

## 七、已知限制

- `/login` 目前为 `PlaceholderView`，登录流程完全由后端 API 驱动，后续接入真实登录页
- 微信小程序（`miniprogram/`）为独立项目，不走此 H5 路由体系
- 集成测试需要 `docker-compose up` 启动 PostgreSQL 和 Redis
