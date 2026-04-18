---
phase: 07
plan: 01
subsystem: 薪资管理增强
tags: [salary, dashboard, adjustment, performance, sick-leave]
dependency_graph:
  requires: [salary-module-existing]
  provides: [salary-dashboard-api, salary-adjustment-api, salary-performance-api, sick-leave-policy-api, salary-dashboard-frontend]
  affects: [internal/salary, cmd/server/main.go, frontend/src/api, frontend/src/views, frontend/src/router]
tech_stack:
  added: [errgroup, shopspring/decimal]
  patterns: [concurrent-aggregation, insert-only-model, upsert-pattern, seed-data]
key_files:
  created:
    - internal/salary/dashboard_service.go
    - internal/salary/dashboard_handler.go
    - internal/salary/dashboard_dto.go
    - internal/salary/adjustment_model.go
    - internal/salary/adjustment_service.go
    - internal/salary/adjustment_handler.go
    - internal/salary/adjustment_dto.go
    - internal/salary/adjustment_repository.go
    - internal/salary/performance_model.go
    - internal/salary/performance_service.go
    - internal/salary/performance_handler.go
    - internal/salary/performance_repository.go
    - internal/salary/sick_leave_policy_model.go
    - internal/salary/sick_leave_policy_service.go
    - frontend/src/views/tool/SalaryDashboard.vue
  modified:
    - internal/salary/handler.go
    - cmd/server/main.go
    - frontend/src/api/salary.ts
    - frontend/src/router/index.ts
decisions:
  - DashboardHandler methods merged onto main salary Handler to simplify route registration
  - SalaryAdjustment uses plain struct (no BaseModel) since INSERT ONLY has no soft-delete or updated_by
  - SickLeavePolicy and SalarySlipSendLog models co-located in sick_leave_policy_model.go
metrics:
  duration: 13min
  completed: "2026-04-18"
  tasks: 5
  files: 19
---

# Phase 07 Plan 01: 薪资数据看板 + 后端基础设施 Summary

薪资看板使用 errgroup 并发聚合 4 指标（应发/实发/社保/个税）+ 环比趋势；调薪 INSERT ONLY 模型；绩效系数 0.0-1.0 范围校验；病假系数按城市工龄档位配置；前端 4 卡片看板跟随 EmployeeDashboard 样式。

## Tasks Completed

| Task | Name | Commit | Key Files |
|------|------|--------|-----------|
| 1 | SalaryDashboard backend | 75436d7 | dashboard_service.go, dashboard_handler.go, dashboard_dto.go |
| 2 | SalaryAdjustment backend | a4c0b70 | adjustment_model.go, adjustment_service.go, adjustment_handler.go, adjustment_dto.go, adjustment_repository.go |
| 3 | PerformanceCoefficient backend | 85c55ff | performance_model.go, performance_service.go, performance_handler.go, performance_repository.go |
| 4 | SickLeavePolicy + AutoMigrate | 49bd3c4 | sick_leave_policy_model.go, sick_leave_policy_service.go, main.go |
| 5 | SalaryDashboard.vue frontend | ba33940 | SalaryDashboard.vue, salary.ts, index.ts |

## Commits

- `75436d7` feat(07-01): add salary dashboard backend with errgroup concurrent aggregation
- `a4c0b70` feat(07-01): add salary adjustment backend with INSERT ONLY model
- `85c55ff` feat(07-01): add performance coefficient backend with upsert and 0.0-1.0 validation
- `49bd3c4` feat(07-01): add sick leave policy and register 4 new tables in AutoMigrate
- `ba33940` feat(07-01): add SalaryDashboard.vue frontend with 4 stat cards and API extensions

## Deviations from Plan

None - plan executed exactly as written.

## Decisions Made

1. **DashboardHandler merged into Handler** - Instead of a separate DashboardHandler struct, GetDashboard was added as a method on the main salary Handler. This simplifies route registration since the salary group already exists in RegisterRoutes.

2. **SalaryAdjustment without BaseModel** - The INSERT ONLY model uses a plain struct instead of BaseModel because it has no updated_by, updated_at, or deleted_at fields by design. Only org_id is retained for multi-tenant isolation.

3. **SickLeavePolicy + SalarySlipSendLog co-located** - Both models placed in sick_leave_policy_model.go since SalarySlipSendLog is a forward reference for Plan 07-03 and is small enough to share the file.

## Self-Check: PASSED

- All 15 created files verified as FOUND
- All 5 commits verified in git log
- No accidental file deletions in any commit
- Backend compiles: `go build ./cmd/server/` passes
- Frontend compiles: `vue-tsc --noEmit` passes
