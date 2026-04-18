---
phase: "06"
plan: "01"
subsystem: attendance
tags: [backend, frontend, attendance, rule-engine, shifts, schedules]
dependency_graph:
  requires: [auth, employee-model]
  provides: [attendance-rules, shifts, schedules, rule-engine]
  affects: [main.go, router, app-layout]
tech_stack:
  added: [gorm-datatypes-jsonb]
  patterns: [Handler-Service-Repository, orgScope-multi-tenant]
key_files:
  created:
    - internal/attendance/model.go
    - internal/attendance/repository.go
    - internal/attendance/rule_engine.go
    - internal/attendance/dto.go
    - internal/attendance/service.go
    - internal/attendance/handler.go
    - frontend/src/api/attendance.ts
    - frontend/src/views/attendance/AttendanceRule.vue
    - frontend/src/components/attendance/ClockStatusTag.vue
  modified:
    - cmd/server/main.go
    - frontend/src/router/index.ts
    - frontend/src/views/layout/AppLayout.vue
decisions:
  - Handler 使用 c.GetInt64("org_id") 模式获取用户信息（匹配项目现有模式）
  - 使用 response.Error() 替代不存在的 response.InternalError()
  - Schedule 唯一索引 idx_schedule_emp_date 防止重复排班
  - AttendanceRule Holidays 使用 JSONB 存储节假日列表
metrics:
  duration: ~12min
  tasks: 3
  files: 12
---

# Phase 06 Plan 01: 打卡规则设置 Summary

管理员可通过顶部 3-Tab 页面配置固定时间/按排班/自由工时三种打卡模式，后端新建 attendance 模块含完整模型、规则引擎和 CRUD API。

## Completed Tasks

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | 创建 Attendance 模型 + Repository 层 | bc06997 | model.go, repository.go |
| 2 | 创建规则引擎 + Service + Handler 层 | 47b0861 | rule_engine.go, dto.go, service.go, handler.go, main.go |
| 3 | 创建前端打卡规则设置页面 | 7f93448 | attendance.ts, AttendanceRule.vue, ClockStatusTag.vue, index.ts, AppLayout.vue |

## Key Changes

### Backend (Go)
- **model.go**: AttendanceRule（打卡规则）、Shift（班次）、Schedule（排班）三个 GORM 模型，Shift 含 `work_date_offset` 跨天班次字段
- **repository.go**: orgScope 多租户隔离，UpsertRule/CRUD Shift/List+BatchUpsert Schedule
- **rule_engine.go**: IsWorkDay/IsHoliday/GetExpectedClockTimes 规则引擎，支持 fixed/scheduled/free 三种模式
- **dto.go**: 请求/响应 DTO，含 binding 标签验证
- **service.go**: 业务逻辑层，JSON 序列化 WorkDays/Holidays，DTO 转换
- **handler.go**: 8 个 API 端点注册到 /api/v1/attendance 路由组
- **main.go**: 注册 attendance 依赖注入、AutoMigrateTables、路由

### Frontend (Vue 3)
- **attendance.ts**: API 客户端含 getRule/saveRule/listShifts/createShift/updateShift/deleteShift/listSchedules/batchUpsertSchedules
- **AttendanceRule.vue**: 3-Tab 页面（固定时间/按排班/自由工时），含班次管理弹窗
- **ClockStatusTag.vue**: 打卡状态标签组件
- **router**: /attendance/rule 路由注册 + auth guard
- **AppLayout**: 桌面端+移动端侧边栏添加考勤管理菜单（Clock 图标）

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Handler 签名适配项目模式**
- **Found during:** Task 2
- **Issue:** Plan 使用 `middleware.RequireAuth` 和 `middleware.GetUserClaims`，但项目中不存在这些函数
- **Fix:** 改为 `authMiddleware gin.HandlerFunc` 参数模式 + `c.GetInt64("org_id")` 获取用户信息，与现有模块（employee/finance）一致
- **Files modified:** handler.go
- **Commit:** 47b0861

**2. [Rule 3 - Blocking] response.InternalError 不存在**
- **Found during:** Task 2
- **Issue:** 项目 response 包只有 BadRequest/Error/Success/Unauthorized/Forbidden，没有 InternalError
- **Fix:** 使用 `response.Error(c, http.StatusInternalServerError, 500, msg)` 替代
- **Files modified:** handler.go
- **Commit:** 47b0861

## Verification Results

- `go build ./...` — PASSED
- `npx tsc --noEmit` — PASSED (frontend)
- No accidental file deletions in any commit

## Self-Check: PASSED
