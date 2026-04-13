# Phase 7: 首页工作台 - Context

**Gathered:** 2026-04-10
**Status:** Ready for planning

<domain>
## Phase Boundary

老板打开 APP/H5 后第一时间看到首页工作台，一眼看清待办事项，一键进入核心功能。覆盖 DASH-01 ~ DASH-10 全部需求（待定）。Phase 7 是**首次引入 Vue 3 H5 前端**的项目里程碑。

</domain>

<decisions>
## Implementation Decisions

### 前端范围与策略
- **D-01:** H5 管理后台先行，APP 原生（Android/iOS）留在 V2.0 实现。Phase 7 不做移动端 APP 开发。
- **D-02:** H5 前端项目建于 `frontend/` 目录（项目根目录下），独立于 Go 后端。独立部署（`frontend/dist/` 静态资源由 Nginx 托管）。
- **D-03:** 技术栈：Vue 3 + Element Plus + Vite + TypeScript + Pinia + Vue Router。完全按 CLAUDE.md 推荐栈。
- **D-04:** H5 导航结构与 APP 设计规范（ui-ux.md）一致：底部 5 Tab（首页/员工/工具/财务/我的）。Tab 栏固定在页面底部。

### 首页布局（ui-ux.md §2.2 已锁定）
- **D-05:** 首页布局从上到下：
  1. **顶部栏**：企业名称 + 「我的」入口（右上角）
  2. **待办提醒区**（首屏核心）：卡片式展示，最多 6 张，按紧急程度排序
  3. **核心功能入口区**：5 宫格（大图标 + 标题）
  4. **数据概览区**（可选折叠）：极简数字卡片
  5. **底部 Tab 导航栏**（固定）

### 待办卡片（全部 6 张，无数据则显示空状态）
- **D-06:** 待办卡片按以下顺序展示（优先级从高到低）：
  1. **社保缴费提醒**（来源：SOCL-03）— 截止日期临近时出现
  2. **个税申报提醒**（来源：TAX-03）— 申报截止前 3 天出现
  3. **员工入离职待审核**（来源：EMPL-01/EMPL-05）— 有待审核申请时出现
  4. **合同到期提醒**（来源：EMPL-08）— 合同到期前 30 天出现
  5. **费用报销待审批**（来源：FINC-09）— 有待审批报销时出现
  6. **凭证待审核**（来源：FINC-03）— 有已提交待审核凭证时出现
- **D-07:** 卡片点击后直接跳转到对应功能模块的待处理列表页，完成后卡片自动消失（下次加载时不再显示）
- **D-08:** 无待办时：显示「暂无待办事项，轻松搞定人事~」的空状态提示（非静态文案，是空状态组件）

### 功能入口（5 宫格）
- **D-09:** 5 个功能入口：
  1. **员工管理** — 入职/离职/档案
  2. **社保管理** — 参保/缴费/变更
  3. **工资管理** — 核算/工资单/发放
  4. **个税申报** — 计算/申报/记录
  5. **财务管理** — 凭证/发票/报表
- **D-10:** 图标采用 Element Plus 内置图标（ep:*），颜色跟随主色 #1677FF

### 数据概览
- **D-11:** 默认展开，显示：
  - 在职员工数 + 本月入职/离职人数
  - 本月社保总金额（个人+单位）
  - 本月工资总发放额
- **D-12:** 支持点击折叠/展开，折叠后仅显示在职人数一行
- **D-13:** 数据仅在进入首页时加载一次（进入时请求，无轮询，无下拉刷新）

### 底部 Tab 导航
- **D-14:** 5 个 Tab：
  - 首页（工作台）
  - 员工（员工列表）
  - 工具（社保/工资/个税快捷入口）
  - 财务（凭证/发票/报表/结账）
  - 我的（个人中心/设置）
- **D-15:** 图标 + 文字双标识，首页 Tab 默认选中

### Go 后端 Dashboard Service
- **D-16:** 新建 `internal/dashboard/` 模块（Go package），提供 `GET /api/v1/dashboard` 聚合接口
- **D-17:** DashboardService 聚合以下数据：
  - 各模块待办数量（调用 Phase 1~6 各 Service）
  - 数据概览数字（员工数、社保总额、工资总额）
  - 当前登录用户的 org_id 通过 JWT middleware 注入，无需前端传参
- **D-18:** API 返回格式：
  ```json
  {
    "todos": [
      {"type": "social_insurance", "title": "社保缴费提醒", "count": 2, "deadline": "2026-04-15", "priority": 1},
      {"type": "tax", "title": "个税申报提醒", "count": 1, "deadline": "2026-04-20", "priority": 2},
      ...
    ],
    "overview": {
      "employee_count": 15,
      "joined_this_month": 2,
      "left_this_month": 0,
      "social_insurance_total": "12345.00",
      "payroll_total": "67890.00"
    }
  }
  ```
- **D-19:** 各待办卡片的**具体跳转目标页面**（如社保缴费→社保缴费列表页）由 Phase 7 前端 plan 定义，不需要在后端决策

### UI 视觉规范（ui-ux.md 已锁定，D-20 引用）
- **D-20:** 遵循 ui-ux.md §1.3 视觉规范：
  - 主色：#1677FF（Element Plus 默认蓝色）
  - 成功色：#52C41A（Element Plus success）
  - 危险色：#FF4D4F（Element Plus danger）
  - 字体：标题 18px / 正文 16px / 辅助文字 14px
  - 图标：线性极简，Element Plus 内置图标
  - 间距：8px / 16px 层级

### Claude's Discretion
- `frontend/` 目录的初始化（`npm create vite@latest` + Element Plus 配置）可以由 planner 在 plan 中决定具体命令，不需要在 context 中锁定
- Tab 栏中「工具」Tab 的内容（社保/工资/个税三合一还是独立展示）由前端 plan 决定
- 响应式策略（H5 在桌面浏览器访问时 Tab 栏是否变成侧边栏）→ V2.0 考虑，Phase 7 专注移动端

</decisions>

<deferred>
## Deferred to V2.0 or Later

- **APP 原生开发**（Android Kotlin / iOS SwiftUI）— V2.0
- **H5 桌面端响应式布局**（侧边栏替代底部 Tab）— V2.0
- **深度数据可视化**（图表/趋势图）— ui-ux.md 已注明 V2.0 人事报表
- **暗黑模式** — PROJECT.md 已标注 Out of Scope
- **桌面端管理后台**（H5 已覆盖基本功能，V2.0 可考虑 PC 专版）
</deferred>

<prior_decisions>
## From Prior Phases

### Phase 1: 基础框架
- JWT token 通过 `Authorization: Bearer <token>` header 传递
- 多租户隔离：所有查询必须包含 `org_id`（TenantScope middleware）
- RBAC 三级：OWNER / ADMIN / MEMBER

### Phase 2: 员工管理
- 员工列表 API：`GET /api/v1/employees?status=active`
- 入职邀请：通过 H5 链接 + 短信验证

### Phase 3: 社保管理
- SOCL-03 提醒通过本地定时任务（gocron）触发，不需要外部队列

### Phase 5: 工资核算
- 工资表状态机：draft → calculated → confirmed → paid
- H5 工资单推送：token + 短信验证模式

### Phase 6: 财务记账
- 凭证状态机：draft → submitted → audited → closed
- decimal.Decimal 全链路使用

### Project-Level (PROJECT.md)
- Core Value: 简单、好用、省时间 — 老板打开APP第一时间知道要做什么，3步完成核心人事操作
- 核心功能操作步骤 ≤ 3步是产品约束，不可违背
</prior_decisions>
