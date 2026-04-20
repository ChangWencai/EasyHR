---
phase: "10"
plan: "03"
type: "execute"
status: "completed"
completed: "2026-04-20"
duration_seconds: 60
tasks_completed: 4
files_created: 1
files_modified: 4
commits:
  - "bd5f71c"
  - "7f0fbf5"
  - "919e542"
  - "24b7837"
requirements:
  - "UX-03"
  - "UX-05"
  - "UX-06"
  - "UX-08"
  - "UX-09"
---

# Phase 10 Plan 03 Summary: 首次引导 + API错误处理 + 工具提示

**One-liner:** 首次使用Tour引导(3步遮罩气泡)+request.ts统一错误映射+关键页面el-tooltip

## Objective

实现首次使用Tour引导(UX-03)、统一API错误处理和重试机制(UX-05/UX-06)、以及关键页面工具提示(UX-09)，同时补充request.ts使用useMessage。

## Tasks Completed

| # | Task | Commit | Files |
|---|------|--------|-------|
| 1 | 创建 TourOverlay.vue 组件 | `bd5f71c` | TourOverlay.vue |
| 2 | 在 HomeView.vue 集成 TourOverlay | `7f0fbf5` | HomeView.vue |
| 3 | 更新 request.ts 错误映射 + useMessage | `919e542` | request.ts |
| 4 | 添加 el-tooltip 到关键列表页 | `24b7837` | EmployeeList.vue, SalaryList.vue |

## Key Decisions

1. **Tour高亮方式**: `.tour-highlight` 使用非scoped `<style>` 全局规则，通过 `!important` + `z-index` 确保覆盖目标元素的原始样式
2. **gridItems dataTour属性**: 在 gridItems 数组每个对象中添加 `dataTour` 属性，v-for 时绑定到 `:data-tour`，而非直接操作DOM
3. **ElMessage→useMessage**: request.ts 直接导入 useMessage()，移除 ElMessage import（Plan 10-01 已创建 useMessage.ts）
4. **el-tooltip header slot**: 使用 `<template #header>` 包裹列标题，el-tooltip 包住 `<span>标签文字</span>`，保留原生列头交互

## TourOverlay.vue 实现细节

- **Props**: `steps: TourStep[]` (title/body/target) + `visible: boolean` (v-model)
- **Emits**: `update:visible`, `close`, `complete`
- **状态**: `currentStep` ref，localStorage `hasSeenTour` 持久化
- **Teleport**: 渲染到 `<body>`，z-index=9999/10000/10001 三层
- **定位**: `getTooltipStyle(target?)` 查询 DOM getBoundingClientRect，默认居中
- **高亮**: watch currentStep 调用 `el.classList.add('tour-highlight')` + `scrollIntoView`
- **清理**: onUnmounted 移除所有 tour-highlight class

## HomeView.vue 集成

- 导入 `TourOverlay` + `TourStep` 类型
- `showTour = ref(!localStorage.getItem(TOUR_DONE_KEY))`
- `tourSteps` 为 computed，动态使用 `todoCount.value`
- 3个引导点: 新增员工 → 待办事项 → 快速上手
- 新入职 gridItem 添加 `dataTour: 'new-employee'`，v-for 绑定 `:data-tour="item.dataTour"`
- `<section class="todo-section" data-tour="todo-section">`

## request.ts 错误映射

| 状态码 | 用户消息 | 处理 |
|--------|---------|------|
| 400 | 请求参数错误，请检查输入 | $msg.error() |
| 401 | 登录已过期，请重新登录 | redirect /login |
| 403 | 您没有权限进行此操作 | $msg.error() |
| 404 | 请求的数据不存在 | $msg.error() |
| 409 | 数据冲突，请刷新后重试 | $msg.error() |
| 422 | 数据验证失败，请检查输入 | $msg.error() |
| 500 | 服务器异常，请稍后重试 | $msg.error({showActions:true}) |
| 502/503 | 服务暂时不可用，请稍后重试 | $msg.error({showActions:true}) |
| 无响应+timeout | 请求超时，请稍后重试 | $msg.error({showActions:true}) |
| 无响应+网络错误 | 网络连接失败，请检查网络后重试 | $msg.error({showActions:true}) |

## el-tooltip 添加详情

| 页面 | 列 | 内容 |
|------|-----|------|
| EmployeeList.vue | 姓名 | 点击查看员工详情 |
| EmployeeList.vue | 操作 | 编辑、查看员工信息 |
| SalaryList.vue | 员工 | 点击查看员工详情 |
| SalaryList.vue | 状态 | 草稿→已核算→已确认→已发放 |
| SalaryList.vue | 实发 | 点击查看月度工资详情 |

所有 tooltip: `effect="dark"`, `placement="top"`, `:show-after="500"`

## Deviations from Plan

**Rule 2 [Auto-fix]: gridItems dataTour 属性方案**
- 原计划通过条件判断在模板中设置 `data-tour` 属性，但操作DOM属性繁琐
- 改为在 gridItems 数组每个对象添加 `dataTour?: string` 属性，v-for 中直接绑定 `:data-tour="item.dataTour"`
- 仅新入职项有值，其他为 undefined（Vue 自动忽略无效 data 属性）

## Known Stubs

无。

## Threat Flags

无新增安全面。

## Self-Check

- [x] TourOverlay.vue 包含 TourStep 接口、props、emits
- [x] TourOverlay 使用 Teleport to body
- [x] localStorage.hasSeenTour 持久化逻辑存在
- [x] skip/prev/next/complete 导航完整
- [x] HomeView.vue 导入 TourOverlay、showTour ref、tourSteps computed
- [x] TourOverlay 在模板中使用 v-model:visible + @complete
- [x] data-tour="new-employee" 和 data-tour="todo-section" 存在
- [x] request.ts 导入 useMessage，移除 ElMessage
- [x] ERROR_MESSAGES 映射表存在
- [x] 401 redirect 行为保持
- [x] EmployeeList.vue 姓名列 + 操作列 tooltip
- [x] SalaryList.vue 员工列 + 状态列 + 实发列 tooltip
- [x] 4个 commit 全部存在

## TDD Gate Compliance

N/A — 此计划为功能实现，非 TDD 计划。

## Commits

- `bd5f71c` feat(10-03): 创建 TourOverlay 首次使用引导组件
- `7f0fbf5` feat(10-03): 在首页集成 TourOverlay 首次引导
- `919e542` feat(10-03): 扩展 request.ts API 错误映射与 useMessage
- `24b7837` feat(10-03): 添加 el-tooltip 到关键列表页 (UX-09)
