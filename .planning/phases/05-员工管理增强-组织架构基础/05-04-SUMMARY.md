---
phase: 05-员工管理增强-组织架构基础
plan: 04
subsystem: [employee, api, ui]
tags: [offboarding, reject, approval, social-insurance, vue3, element-plus, gin]

# Dependency graph
requires:
  - phase: 05-02
    provides: Offboarding 基础模型（pending/approved/completed 状态 + ApproveResign 方法）
  - phase: 05-03
    provides: 前端 OffboardingList 基础列表 + SocialInsuranceEventHandler 接口
provides:
  - OffboardingStatusRejected = "rejected" 状态常量
  - RejectResign API（PUT /offboardings/:id/reject）
  - CompleteOffboardingFromSI 方法（社保减员完成后自动更新离职状态）
  - 前端行内审批按钮（同意/驳回）、驳回弹窗、去减员跳转
  - 状态筛选下拉包含"已驳回"选项
affects: [06-*, socialinsurance]

# Tech tracking
tech-stack:
  added: []
  patterns: [行内审批按钮模式, el-popconfirm 确认, 跨模块回调 CompleteOffboardingFromSI]

key-files:
  created: []
  modified:
    - internal/employee/offboarding_model.go
    - internal/employee/offboarding_service.go
    - internal/employee/offboarding_handler.go
    - internal/employee/offboarding_dto.go
    - frontend/src/views/employee/OffboardingList.vue
    - frontend/src/api/employee.ts
    - frontend/src/views/employee/statusMap.ts
    - frontend/src/views/employee/statusMap.vue

key-decisions:
  - "前端状态值从 pending_review 修正为 pending，与后端模型 source of truth 保持一致"
  - "CompleteOffboardingFromSI 方法接收 employeeID 而非 offboardingID，因为社保模块只知道员工 ID"
  - "驳回 API 使用 PUT 方法，与 approve/complete 保持 RESTful 一致性"
  - "前端 API 调用从 POST 改为 PUT，与后端路由对齐"

patterns-established:
  - "跨模块回调模式：社保模块停缴完成后调用 employee.CompleteOffboardingFromSI 自动更新离职状态（EMP-12）"
  - "行内审批模式：列表中按状态显示不同操作按钮，使用 el-popconfirm 确认"

requirements-completed: [EMP-09, EMP-10, EMP-11, EMP-12]

# Metrics
duration: 8min
completed: 2026-04-18
---

# Phase 05 Plan 04: 离职审批流程优化 Summary

**离职驳回 API + 行内审批按钮 + 审批通过后去减员跳转 + 社保减员完成后自动更新离职状态**

## Performance

- **Duration:** 8 min
- **Started:** 2026-04-18T03:24:54Z
- **Completed:** 2026-04-18T03:33:00Z
- **Tasks:** 2
- **Files modified:** 8

## Accomplishments
- 后端新增 RejectResign API（仅允许 pending -> rejected 转换，服务端严格校验）
- 后端新增 CompleteOffboardingFromSI 方法，供社保模块停缴完成后自动将离职状态更新为 completed
- 前端离职列表行内审批：pending 状态显示同意+驳回按钮，approved 状态显示去减员+完成离职按钮
- 驳回弹窗包含原因输入框（选填），同意和完成操作使用 el-popconfirm 确认
- 去减员按钮跳转社保减员页面，携带 employee_id + employee_name 查询参数
- 修正前后端状态值不一致：pending_review -> pending

## Task Commits

Each task was committed atomically:

1. **Task 1: 后端 RejectResign API + rejected 状态 + 社保减员自动完成回调** - `c6fd5f0` (feat)
2. **Task 2: 前端 OffboardingList 行内审批扩展** - `311e08c` (feat)

## Files Created/Modified
- `internal/employee/offboarding_model.go` - 新增 OffboardingStatusRejected = "rejected" 状态常量
- `internal/employee/offboarding_service.go` - 新增 RejectResign 方法 + CompleteOffboardingFromSI 方法
- `internal/employee/offboarding_handler.go` - 新增 PUT /offboardings/:id/reject 路由和 handler
- `internal/employee/offboarding_dto.go` - 新增 RejectResignRequest DTO
- `frontend/src/views/employee/OffboardingList.vue` - 重写：行内审批按钮、驳回弹窗、去减员跳转、状态筛选
- `frontend/src/api/employee.ts` - 新增 rejectOffboarding API，修正 Offboarding 接口定义
- `frontend/src/views/employee/statusMap.ts` - 新增 rejected 映射，pending_review -> pending
- `frontend/src/views/employee/statusMap.vue` - 同步更新

## Decisions Made
- 前端状态值从 pending_review 修正为 pending，与后端模型保持一致（source of truth 是后端）
- CompleteOffboardingFromSI 接收 employeeID 而非 offboardingID，因为社保模块只持有员工 ID
- 前端 API 调用从 POST 改为 PUT，与后端路由对齐（approve/reject/complete 统一使用 PUT）

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] 修正前后端状态值不一致**
- **Found during:** Task 2（前端 OffboardingList 扩展）
- **Issue:** 前端 statusMap 使用 `pending_review`，但后端模型定义状态为 `pending`，导致列表无法正确显示状态
- **Fix:** 将前端 statusMap.ts 和 statusMap.vue 中的 `pending_review` 统一改为 `pending`，新增 `rejected` 映射
- **Files modified:** frontend/src/views/employee/statusMap.ts, frontend/src/views/employee/statusMap.vue
- **Verification:** `npx vue-tsc --noEmit` 通过，前后端状态值一致
- **Committed in:** 311e08c（Task 2 commit）

**2. [Rule 1 - Bug] 前端 API HTTP 方法与后端路由不匹配**
- **Found during:** Task 2
- **Issue:** 前端 approveOffboarding 和 completeOffboarding 使用 `request.post`，但后端路由注册为 `PUT`
- **Fix:** 将前端 API 调用从 `request.post` 改为 `request.put`
- **Files modified:** frontend/src/api/employee.ts
- **Verification:** `npx vue-tsc --noEmit` 通过
- **Committed in:** 311e08c（Task 2 commit）

**3. [Rule 2 - Missing Critical] 前端 Offboarding 接口定义与后端 DTO 不同步**
- **Found during:** Task 2
- **Issue:** 前端 Offboarding 接口使用 `resign_reason`/`last_workday`/`checklist` 字段名，但后端 OffboardingDetailResponse 返回 `reason`/`resignation_date`/`checklist_items`
- **Fix:** 更新 Offboarding 接口定义匹配后端 DTO 字段名
- **Files modified:** frontend/src/api/employee.ts
- **Verification:** `npx vue-tsc --noEmit` 通过
- **Committed in:** 311e08c（Task 2 commit）

---

**Total deviations:** 3 auto-fixed（2 bug, 1 missing critical）
**Impact on plan:** 所有修正均为必要的前后端一致性修复，无 scope creep。

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- 离职审批流程完整（pending -> approved/rejected, approved -> completed）
- EMP-12 回调机制就绪：社保模块需在 BatchStopEnrollment 成功后调用 employee.CompleteOffboardingFromSI
- 前端状态映射已统一，后续离职相关页面可直接复用 statusMap

---
*Phase: 05-员工管理增强-组织架构基础*
*Completed: 2026-04-18*

## Self-Check: PASSED

All 8 modified files verified as existing on disk. Both task commits (c6fd5f0, 311e08c) confirmed in git log.
