# Phase 9: 待办中心 - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-19
**Phase:** 09-待办中心
**Areas discussed:** 完成率环形图, 限时任务引擎, 首页轮播图+快捷入口, 协办邀请+终止任务

---

## 完成率环形图

| Option | Description | Selected |
|--------|-------------|----------|
| ECharts 环形图（推荐） | 使用 ECharts pie chart（radius=['40%','70%']），配色蓝色系，两张图并排 | ✓ |
| el-progress 简单环形 | 使用 Element Plus el-progress type="circle"，两张环形进度并排 | |
| CSS 自绘环形图 | 纯 CSS + SVG 绘制，不引入图表库依赖 | |

**User's choice:** ECharts 环形图（推荐）
**Notes:** ECharts 已在 Phase 5 使用，且环形图是 PRD 明确要求，无需引入新依赖

| Option | Description | Selected |
|--------|-------------|----------|
| HomeView 顶部（推荐） | 环形图紧接页面标题区下方（标题 → 环形图 → 待办事项 → 快捷入口） | ✓ |
| 独立待办中心页面 | 独立路由页面（/todo-center），通过侧边栏菜单进入 | |
| 混合（HomeView+独立页） | 环形图放 HomeView，完整列表和筛选器放在独立页面 | |

**User's choice:** HomeView 顶部（推荐）
**Notes:** 保持所有信息在首页，减少跳转，符合"3步完成"原则

---

## 限时任务引擎

| Option | Description | Selected |
|--------|-------------|----------|
| 扩展 TodoItem（推荐） | 不新建表，扩展 TodoItem：新增 deadline/is_time_limited/urgency_status | ✓ |
| 独立 TimeLimitedTask 表 | 新建 TimeLimitedTask 表，TodoItem 仅作为关联展示层 | |

**User's choice:** 扩展 TodoItem（推荐）
**Notes:** 与现有待办列表统一，不需要维护两张表

| Option | Description | Selected |
|--------|-------------|----------|
| 1-7天超期=超时，15天+=失效（推荐） | 剩余 1-7 天 → 超时（红色警告）；超过截止日期 15 天以上 → 失效（灰色） | ✓ |
| 超期直接失效，不单独显示超时 | 超期即失效，失效任务从列表移除，仅在归档中可查 | |

**User's choice:** 1-7天超期=超时，15天+=失效（推荐）
**Notes:** 与 PRD TODO-18 一致（"超时状态为超时（1-7日），超过15日状态为失效"）

---

## 首页轮播图+快捷入口

| Option | Description | Selected |
|--------|-------------|----------|
| 管理员配置轮播图（推荐） | 管理员可在后台上传1-3张图片+跳转链接，图片存OSS，asynq定时任务同步生效 | ✓ |
| 固定文案轮播图 | 内容固定写死：入职季/发薪日等固定文案+对应功能链接 | |

**User's choice:** 管理员配置轮播图（推荐）
**Notes:** 更灵活，管理员可自定义内容

| Option | Description | Selected |
|--------|-------------|----------|
| 保留现有6个+追加新入口（推荐） | 追加新入口（新入职/调薪/考勤）到现有网格，支持横向滚动或换行展示 | ✓ |
| 全部替换为PRD指定5入口 | 全部替换为 PRD 指定的 5 个入口，去掉其他入口 | |
| 管理员自定义快捷入口 | 管理员可自定义快捷入口（增删改顺序），存 org_id 级别配置 | |

**User's choice:** 保留现有6个+追加新入口（推荐）
**Notes:** 保持现有功能不变，同时满足 PRD 要求

| Option | Description | Selected |
|--------|-------------|----------|
| CarouselItem 表（推荐） | 新建 CarouselItem 表（id/org_id/image_url/link_url/sort_order/active/start_at/end_at） | ✓ |
| Organization.JSONB 扩展 | 复用 Organization 表的 JSON 字段，不新建表 | |

**User's choice:** CarouselItem 表（推荐）
**Notes:** 独立表更清晰，支持多张轮播图和生效时间段管理

---

## 协办邀请+终止任务

| Option | Description | Selected |
|--------|-------------|----------|
| 纯填写（推荐） | 仅可填写数据（补充员工信息/提交假勤申请），不能查看企业敏感数据 | ✓ |
| 查看详情+填写 | 协办人可查看待办详情+填写数据，类似员工信息登记的 Token 模式 | |

**User's choice:** 纯填写（推荐）
**Notes:** 降低数据泄露风险，符合最小权限原则

| Option | Description | Selected |
|--------|-------------|----------|
| 复用 Token 机制（推荐） | 复用 Phase 5 员工信息登记 RegisterPage.vue 模式，生成 /todo/:id/invite?token=xxx | ✓ |
| 需要注册/登录 | 协办人需先注册/登录才能填写 | |

**User's choice:** 复用 Token 机制（推荐）
**Notes:** Phase 5 已验证，无需额外注册流程，体验流畅

| Option | Description | Selected |
|--------|-------------|----------|
| 保留数据+标记终止（推荐） | 终止后保留数据，仅状态变为"已终止"，管理员仍可在筛选中看到 | ✓ |
| 软删除+归档 | 终止后从列表移除，仅在归档/审计日志中可查 | |

**User's choice:** 保留数据+标记终止（推荐）
**Notes:** 保留数据便于追溯，适用于临时暂停场景

---

## Claude's Discretion

- 环形图的具体配色（蓝色系 vs 品牌主色）
- 轮播图的切换动画（淡入淡出 vs 滑动）
- 快捷入口的具体图标选择（Element Plus icons）
- 限时任务 7 种的具体生成触发时机（asynq cron 还是各模块直接创建）
- 协办填写页的具体字段和布局（由各待办类型决定）
- TodoItem 列表的分页大小（20/50/100）

## Deferred Ideas

None
