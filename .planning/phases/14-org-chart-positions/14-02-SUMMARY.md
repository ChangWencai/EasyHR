---
phase: "14"
plan: "02"
subsystem: frontend
tags: [org-chart, echarts-tree, contextmenu, inline-edit, el-select, el-optgroup, position]
dependency_graph:
  requires:
    - phase: "14-01"
      provides: "Position CRUD API, select-options endpoint, BuildTree v2, transfer-delete endpoint"
  provides:
    - position-api-client
    - org-chart-contextmenu
    - org-chart-inline-edit
    - delete-transfer-dialog
    - move-dept-dialog
    - search-highlight-color-flowthrough
    - position-el-select-dropdown
  affects: [employee-form, org-chart]
tech_stack:
  added: []
  patterns: [el-optgroup-grouping, echarts-event-binding, position-resolve-name]
key_files:
  created:
    - frontend/src/api/position.ts
  modified:
    - frontend/src/api/department.ts
    - frontend/src/views/employee/OrgChart.vue
    - frontend/src/views/employee/EmployeeCreate.vue
key_decisions:
  - "el-option value 使用空字符串替代 null 避免 TS2322 类型错误（EpPropMergeType 不接受 null）"
  - "chartOption itemStyle 改为函数模式：后端搜索高亮色透传，无后端色时按节点类型着色"
  - "ECharts 事件绑定在 loadTree 后执行（确保 chartRef 已挂载）"
  - "EmployeeCreate 表单 position_id 和 position 双字段共存（position 保留用于 API 兼容和显示）"
patterns-established:
  - "Position API Client: positionApi 封装 5 个端点 + getSelectOptions 分组返回"
  - "OrgChart 右键菜单: contextmenu 事件 + 浮动定位 div + document click 关闭"
  - "岗位下拉分组: el-optgroup 按部门专属/通用分组 + 未分配岗位选项"
requirements-completed:
  - ORG-02
  - ORG-03
  - ORG-04
metrics:
  duration: 13min
  completed: "2026-04-21"
---

# Phase 14 Plan 02: 前端 OrgChart 增强 + 岗位下拉 + 员工表单适配 Summary

组织架构图增强（右键菜单/内联编辑/删除转移对话框/搜索高亮色透传）+ 岗位 API 客户端 + 员工表单岗位 el-select 分组下拉

## Performance

- **Duration:** 13 min
- **Started:** 2026-04-21T09:15:33Z
- **Completed:** 2026-04-21T09:28:33Z
- **Tasks:** 3
- **Files modified:** 4

## Accomplishments
- 新建 position.ts API 客户端，封装 Position CRUD + getSelectOptions 分组端点
- department.ts 新增 transferDelete 方法，支持删除部门时员工转移
- OrgChart.vue 增强右键菜单（移动到/删除部门）、内联编辑部门名称、删除转移对话框、搜索高亮色 #4F6EF7 透传
- EmployeeCreate.vue 岗位字段从 el-input 改为 el-select + el-optgroup 分组下拉（部门专属/通用/未分配）

## Task Commits

Each task was committed atomically:

1. **Task 1: Create position.ts API client + update department.ts** - `cd17d70` (feat)
2. **Task 2: Enhance OrgChart.vue** - `6706824` (feat)
3. **Task 3: EmployeeCreate.vue position el-select** - `b3d8470` (feat)

## Files Created/Modified
- `frontend/src/api/position.ts` - Position/PositionSelectOptions 类型 + positionApi 5 个端点
- `frontend/src/api/department.ts` - 新增 transferDelete 方法
- `frontend/src/views/employee/OrgChart.vue` - 右键菜单 + 内联编辑 + 删除转移 + 移动部门 + 节点着色 + 搜索高亮透传
- `frontend/src/views/employee/EmployeeCreate.vue` - 岗位 el-select 分组下拉 + position_id + loadPositionOptions

## Decisions Made
- el-option value 使用空字符串替代 null 避免 TS2322 类型错误
- chartOption itemStyle 改为函数模式实现后端搜索高亮色透传
- ECharts 事件绑定在 loadTree 数据加载后执行
- EmployeeCreate 表单 position_id 和 position 双字段共存

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] el-option :value="null" TS2322 类型错误**
- **Found during:** Task 3 (EmployeeCreate.vue 修改后构建验证)
- **Issue:** Element Plus el-option 的 value prop 类型不接受 null，TS 编译报错
- **Fix:** 将 `<el-option :value="null">` 改为 `<el-option value="">`，空字符串表示未分配岗位
- **Files modified:** frontend/src/views/employee/EmployeeCreate.vue
- **Verification:** vue-tsc --noEmit 通过（0 errors）
- **Committed in:** b3d8470 (Task 3 commit)

**2. [Rule 1 - Bug] OrgChart.vue 未使用的 positionApi/Position import**
- **Found during:** Task 3 (构建验证)
- **Issue:** positionApi 和 Position 类型在 OrgChart.vue 中 import 但未使用，vue-tsc 报 TS6133
- **Fix:** 移除 OrgChart.vue 中未使用的 positionApi 和 Position import
- **Files modified:** frontend/src/views/employee/OrgChart.vue
- **Verification:** vue-tsc --noEmit 通过
- **Committed in:** b3d8470 (Task 3 commit)

**3. [Rule 1 - Bug] EmployeeCreate.vue 未使用的 watch/Briefcase import**
- **Found during:** Task 3 (构建验证)
- **Issue:** watch 和 Briefcase 在 EmployeeCreate.vue 中 import 但未使用，TS6133 错误
- **Fix:** 移除未使用的 watch 和 Briefcase import
- **Files modified:** frontend/src/views/employee/EmployeeCreate.vue
- **Verification:** vue-tsc --noEmit 通过
- **Committed in:** b3d8470 (Task 3 commit)

---

**Total deviations:** 3 auto-fixed (3 bugs — TS 类型错误和未使用 import)
**Impact on plan:** 所有修复都是构建正确性必需的，无范围蔓延。

## Issues Encountered
- 预存在的 TS 错误（createFormRef TS6133、employeeApi.create .id TS2339、Record conversion TS2352）不是本计划引入的，已在 stash 对比中确认

## User Setup Required
None - 无需外部服务配置。

## Self-Check: PASSED

- [x] frontend/src/api/position.ts exists
- [x] frontend/src/api/department.ts exists (with transferDelete)
- [x] frontend/src/views/employee/OrgChart.vue exists (with contextmenu/inline-edit/delete-transfer)
- [x] frontend/src/views/employee/EmployeeCreate.vue exists (with el-select/el-optgroup)
- [x] 14-02-SUMMARY.md exists
- [x] Commit cd17d70 (Task 1) found
- [x] Commit 6706824 (Task 2) found
- [x] Commit b3d8470 (Task 3) found

## Next Phase Readiness
- Phase 14 前端全部完成
- OrgChart.vue 支持右键菜单/内联编辑/删除转移/移动部门/搜索高亮
- EmployeeCreate.vue 岗位下拉选择器已就绪
- position.ts API 客户端可供其他组件复用

---
*Phase: 14-org-chart-positions*
*Completed: 2026-04-21*
