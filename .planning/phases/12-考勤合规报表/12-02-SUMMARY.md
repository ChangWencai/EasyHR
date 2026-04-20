# Phase 12 Plan 02: 考勤合规报表前端基础设施

## One-liner

考勤合规报表前端基础设施: API层、路由、侧边栏菜单和可复用UI组件，为 Plan 12-03 页面提供基础。

## Commits

| # | Task | Commit | Files |
|---|------|--------|-------|
| 1 | API 函数和类型定义 | `77eb279` | attendance.ts |
| 2 | 路由和认证守卫 | `e4370aa` | router/index.ts |
| 3 | 侧边栏菜单 | `bc0df62` | AppLayout.vue |
| 4 | 可复用组件 | `17d8749` | ComplianceStatCard.vue, ComplianceTable.vue |
| -- | 合规报表页面骨架 | `307ea95` | ComplianceOvertime/Leave/Anomaly/Monthly.vue |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] 缺少合规页面占位符导致构建失败**

- **Found during:** Task 2 verification (npm run build)
- **Issue:** vue-tsc 报错 `Cannot find module '@/views/compliance/ComplianceOvertime.vue'`，因 router 注册了 4 个合规页面但文件不存在
- **Fix:** 在 `src/views/compliance/` 下创建 4 个完整的页面骨架文件，包含统计卡片、数据表格、部门筛选和导出功能，与 Plan 12-03 完整实现保持一致
- **Files modified:** 新增 ComplianceOvertime.vue, ComplianceLeave.vue, ComplianceAnomaly.vue, ComplianceMonthly.vue
- **Commit:** `307ea95`

## Artifacts

### API Routes (5 endpoints consumed)

| Method | Path | Description |
|--------|------|-------------|
| GET | /attendance/compliance/overtime | 加班统计报表 |
| GET | /attendance/compliance/leave | 请假合规报表 |
| GET | /attendance/compliance/anomaly | 出勤异常报表 |
| GET | /attendance/compliance/monthly | 月度综合合规报表 |
| GET | /attendance/compliance/monthly/export | 月度考勤汇总 Excel 导出 |

### Key Implementation Details

**API 层 (attendance.ts):**
- 新增 5 个合规报表 API 函数
- 新增 13 个 TypeScript 类型定义（OvertimeItem, LeaveItem, AnomalyItem, MonthlyComplianceItem 等）
- 所有 API 支持 dept_ids 多部门筛选参数

**路由 (router/index.ts):**
- 注册 4 个合规页面路由: /compliance/overtime, /compliance/leave, /compliance/anomaly, /compliance/monthly
- /compliance/* 加入 isProtectedRoute 守卫

**侧边栏 (AppLayout.vue):**
- Desktop 侧边栏: 在考勤管理后新增合规报表子菜单（Document 图标）
- Mobile 抽屉菜单: 同步新增合规报表子菜单
- pageTitleMap 添加 4 个合规页面标题映射

**可复用组件:**
- ComplianceStatCard: 可复用统计卡片，支持 icon + value + label + iconClass
- ComplianceTable: 可复用合规表格（el-table + 分页 + 异常行样式）

### Compliance Pages

每个页面包含:
- 页面标题 + 月份选择器
- 部门多选筛选
- 统计卡片概览区域
- 数据表格 + 分页
- 异常行红色高亮（ComplianceAnomaly, ComplianceMonthly）
- Excel 导出（ComplianceMonthly）

## Verification

```
vue-tsc -b  # 4 个合规路由错误已修复，剩余错误均为预存问题
vite build  # 成功，dist/ 目录已生成
```

预存 TypeScript 错误（不在本次 plan 范围内）:
- StepWizard.vue: unused import
- ContractStatusBadge.vue: type error
- useMessage.ts: unused import
- ContractList.vue: unused import
- EmployeeCreate.vue: property 'id' missing
- EmployeeList.vue: missing export batchImportEmployees
- SignPage.vue: property 'window' missing

## Self-Check

- [x] 5 个合规 API 函数存在于 attendance.ts
- [x] 13 个合规类型定义存在于 attendance.ts
- [x] 4 个合规路由注册于 router/index.ts
- [x] /compliance/* 加入 isProtectedRoute 守卫
- [x] AppLayout.vue desktop 侧边栏有合规报表菜单
- [x] AppLayout.vue mobile drawer 有合规报表菜单
- [x] Document 图标已导入
- [x] pageTitleMap 包含 4 个合规路由
- [x] ComplianceStatCard.vue 存在（> 40 行）
- [x] ComplianceTable.vue 存在（> 80 行）
- [x] 4 个合规页面文件存在
- [x] vite build 成功

## Self-Check: PASSED
