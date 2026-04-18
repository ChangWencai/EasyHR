# Phase 9: 待办中心 - Context

**Gathered:** 2026-04-19
**Status:** Ready for planning

<domain>
## Phase Boundary

管理员拥有统一的事项聚合入口，系统自动生成限时任务，快捷办事直达核心操作。

具体包含：
- 待办事项汇总列表（搜索/筛选/置顶/导出）
- 首页完成率环形图（全部事项 / 限时任务）
- 首页轮播图（管理员配置 1-3 张）+ 快捷入口增强
- 7 种限时任务自动生成（合同/个税/社保公积金）+ 超时/失效状态
- 协办邀请（Token 链接，纯填写）+ 终止任务

**Scope:** 仅 H5 管理后台，后端 API 配合新增。
**Depends on:** Phase 5（员工数据）、Phase 6（考勤审批数据）、Phase 7（薪资任务）、Phase 8（社保任务）

</domain>

<decisions>
## Implementation Decisions

### 完成率环形图
- **D-09-01:** 使用 ECharts pie chart（radius=['40%','70%']），蓝色系配色，两张图并排展示（全部事项完成率 / 限时任务完成率）
- **D-09-02:** 环形图位于 HomeView.vue 顶部，页面标题区下方，与待办事项列表临近（同一页面无需跳转）

### 限时任务引擎
- **D-09-03:** 不新建表，扩展 TodoItem 模型：新增 `deadline`（截止日期）、`is_time_limited`（是否限时任务）、`urgency_status`（normal/overdue/expired）
- **D-09-04:** 超时/失效判定规则：剩余 1-7 天 → 超时（红色警告）；超过截止日期 15 天以上 → 失效（灰色，显示已失效）；1-7 天超期且 < 15 天仍为超时态
- **D-09-05:** asynq 定时任务每日凌晨扫描所有限时任务，更新 urgency_status；7 种限时任务由各模块（attendance/salary/socialinsurance/employee）触发创建

### 首页轮播图
- **D-09-06:** 新建 `CarouselItem` 表（id/org_id/image_url/link_url/sort_order/active/start_at/end_at），图片存阿里云 OSS；管理员在个人中心/设置页配置 1-3 张轮播图
- **D-09-07:** asynq 定时任务在轮播图 start_at/end_at 时间段内自动激活/停用

### 快捷入口
- **D-09-08:** 保留现有 6 个入口（员工管理/薪资管理/社保管理/个税申报/凭证管理/发票管理），追加 3 个新入口（新入职/调薪/考勤），支持横向滚动或换行展示

### 协办邀请 + 终止任务
- **D-09-09:** 协办人收到链接后仅可填写数据（补充员工信息/提交假勤申请），不能查看企业敏感数据；Token 独立验证，无需登录
- **D-09-10:** 复用 Token 机制（类似 Phase 5 员工信息登记 RegisterPage.vue 模式）：生成 `/todo/:id/invite?token=xxx` 链接，Token 验证后进入填写页
- **D-09-11:** 终止后的待办保留数据，仅状态变为"已终止"（terminated），管理员仍可在筛选中看到，适用于临时暂停场景；不软删除

### Claude's Discretion
- 环形图的具体配色（蓝色系 vs 品牌主色）
- 轮播图的切换动画（淡入淡出 vs 滑动）
- 快捷入口的具体图标选择（Element Plus icons）
- 限时任务 7 种的具体生成触发时机（asynq cron 还是各模块直接创建）
- 协办填写页的具体字段和布局（由各待办类型决定）
- TodoItem 列表的分页大小（20/50/100）

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements
- `.planning/REQUIREMENTS.md` §1 — Phase 9 需求定义（TODO-01~TODO-20）
- `.planning/ROADMAP.md` §Phase 9 — 阶段目标、成功标准、依赖关系

### Existing Code Patterns
- `frontend/src/views/home/HomeView.vue` — 首页现有结构（待办事项/快捷入口/数据概览），环形图和轮播图在此基础上扩展
- `frontend/src/stores/dashboard.ts` — DashboardStore.TodoItem 已有字段结构，需扩展 deadline/is_time_limited/urgency_status
- `internal/dashboard/service.go` — Dashboard 聚合模式，环形图统计复用 errgroup 并发
- `frontend/src/views/employee/RegisterPage.vue` — Token 链接模式，复用于协办邀请
- `internal/employee/invitation_model.go` — Invitation Token 生成和验证机制
- `internal/socialinsurance/scheduler.go` — asynq 定时任务框架，复用于限时任务状态扫描
- `.planning/phases/08-社保公积金增强/08-CONTEXT.md` — Phase 08 决策（asynq 定时任务模式）
- `.planning/phases/05-员工管理增强-组织架构基础/05-CONTEXT.md` — Phase 05 决策（Token 机制、RegisterPage）

### Project Decisions
- `.planning/PROJECT.md` — Key Decisions 表（多租户 org_id、RBAC、加密策略）
- `.planning/PROJECT.md` — Core Value：3步内完成核心操作，零学习成本
- `.planning/STATE.md` — 项目状态（Phase 05-08 已完成，Phase 07 部分完成）

### Go Module Dependency
- `asynq`（已安装）— 定时任务扫描 + 批量操作
- `ECharts`（前端已安装）— 环形统计图
- `excelize`（已安装）— Excel 导出

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `DashboardStore.TodoItem`: 已有 title/type/count/deadline/type 字段，扩展 deadline/is_time_limited/urgency_status
- `HomeView.vue`: 已有待办事项卡片区域（DashboardStore.todos）和快捷入口网格（gridItems），可直接扩展环形图和轮播图
- `RegisterPage.vue`: Token 验证+填写页模式，复用于协办邀请
- `asynq scheduler.go`: 定时任务框架，复用于限时任务状态扫描

### Established Patterns
- Handler → Service → Repository 三层架构（所有模块统一）
- org_id 逻辑多租户隔离，GORM Scope 自动注入
- Token 链接邀请（无需登录），Phase 5 员工信息登记已验证
- asynq 定时任务每日凌晨运行（与 Phase 8 一致）
- 环形图并排展示两张（全部事项 / 限时任务）

### Integration Points
- attendance → todo: 考勤相关待办（假勤申请/补卡审批）由 Phase 6 审批流触发
- salary → todo: 个税申报/发工资条由 Phase 7 薪资模块触发
- socialinsurance → todo: 社保缴费/增减员由 Phase 8 触发
- employee → todo: 合同新签/续签由 employee 模块触发
- home → todo: HomeView 展示环形图和轮播图
- CarouselItem → home: 轮播图数据由 CarouselItem 表提供

### Critical Anti-Patterns to Avoid
- ❌ 新建独立限时任务表 → 直接扩展 TodoItem 字段，与现有待办列表统一展示
- ❌ 浮点计算百分比 → 全部使用整数计算（已完成/（已完成+待办）×100）
- ❌ 轮播图写死文案 → 管理员可配置 1-3 张，支持设置生效时间段

</code_context>

<specifics>
## Specific Ideas

- 轮播图支持设置开始/结束时间，asynq 定时任务自动激活/停用
- 快捷入口追加：新入职（/employee/create）、调薪（/tool/salary）、考勤（/attendance/clock-live）
- 协办填写页布局参考 RegisterPage.vue，字段由各待办类型决定
- 终止任务保留数据，仅在列表筛选器中增加"已终止"状态筛选

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 09-待办中心*
*Context gathered: 2026-04-19*
