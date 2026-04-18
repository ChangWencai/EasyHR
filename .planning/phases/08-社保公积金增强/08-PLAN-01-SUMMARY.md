---
phase: 08
plan: 01
subsystem: socialinsurance
tags: [backend, model, service, handler, scheduler, asynq]
dependency_graph:
  requires: [Phase 05]
  provides: [SIMonthlyPayment model, SIDashboardService, SIMonthlyPaymentRepository, API routes, asynq tasks]
  affects: [internal/socialinsurance/*, internal/common/model/org.go, cmd/server/main.go]
tech_stack:
  added: [shopspring/decimal for amount fields, asynq worker pattern]
  patterns: [errgroup concurrent queries, idempotent batch upsert, INSERT ONLY status updates]
key_files:
  created:
    - internal/socialinsurance/dashboard_service.go
  modified:
    - internal/socialinsurance/model.go
    - internal/socialinsurance/repository.go
    - internal/socialinsurance/service.go (no changes, existing service preserved)
    - internal/socialinsurance/handler.go
    - internal/socialinsurance/scheduler.go
    - internal/socialinsurance/dto.go
    - internal/common/model/org.go
    - cmd/server/main.go
decisions:
  - SIMonthlyPayment uses model.BaseModel with uint EmployeeID to match plan spec
  - YearMonth stored as varchar(6) in YYYYMM format per plan spec
  - All amount fields use shopspring/decimal to prevent float64 precision loss
  - Dashboard DTOs reuse salary StatItem pattern for frontend consistency
  - asynq worker complements existing gocron scheduler (not replacing)
  - NewHandler signature extended with dashboardSvc and paymentRepo params
  - HMAC webhook signature validation marked as TODO (T-08-04)
metrics:
  duration: 10min
  completed: 2026-04-19
  tasks_completed: 3
  files_changed: 9
  commits: 3
---

# Phase 08 Plan 01: 社保公积金增强后端基础设施 Summary

SIMonthlyPayment 月度缴费模型 + DashboardService 4指标并发聚合 + asynq 定时任务生成/状态流转 + 7个新API路由

## Tasks Completed

| Task | Name | Commit | Key Files |
|------|------|--------|-----------|
| 1 | 新增 SIMonthlyPayment 模型 | f009ab7 | model.go, org.go |
| 2 | 实现 DashboardService 和 Repository | 71f8198 | repository.go, dashboard_service.go, dto.go |
| 3 | 新增 API Handler 和 asynq 定时任务 | 4276b75 | handler.go, scheduler.go, main.go |

## Implementation Details

### Task 1: SIMonthlyPayment 模型
- 新增 `SIMonthlyPayment` 结构体，包含 `PaymentStatus` 五种状态枚举（normal/pending/overdue/transferred/not_transferred）
- 新增缴费渠道常量（self/agent_new/agent_existing）
- 所有金额字段使用 `shopspring/decimal`（CompanyAmount/PersonalAmount/TotalAmount）
- 添加 `Organization.SIPaymentChannel` 字段作为企业默认缴费渠道
- YearMonth 格式 YYYYMM（varchar(6)），复合索引 idx_org_employee_month

### Task 2: DashboardService 和 Repository
- `SIDashboardService.GetDashboard` 使用 errgroup 并发查询 4 个指标（应缴总额/单位/个人/欠缴）+ 环比计算
- `SIMonthlyPaymentRepository` 提供 Create/GetByOrgAndEmployee/GetByOrgAndYearMonth/GetOverdueByOrg/UpdateStatus/BatchUpsert/SumFieldByOrgAndYearMonth/UpdateOverduePayments/DeleteOlderThan 方法
- `BatchUpsert` 使用 PostgreSQL ON CONFLICT DO NOTHING 实现幂等
- Dashboard DTOs（SIDashboardResponse/SIStatItem/OverdueItem）复用 salary StatItem 模式

### Task 3: API Handler 和 asynq 定时任务
- 7 个新路由：dashboard/enroll/single/stop/single/payment-callback/monthly-records/monthly-records/:id/confirm-payment
- `MonthlyPaymentWorker`：HandleGenerateMonthlyPayments（批量生成+幂等UPSERT）、HandleCheckPaymentStatus（D-SI-03: >=26日 pending->overdue）
- asynq 任务类型：TypeGenerateMonthlyPayments、TypeCheckPaymentStatus
- main.go 更新 DI 注入 siPaymentRepo、siDashboardSvc，AutoMigrate 添加 SIMonthlyPayment

## Deviations from Plan

None - plan executed exactly as written.

## Threat Model Compliance

| Threat ID | Component | Mitigation | Status |
|-----------|-----------|------------|--------|
| T-08-01 | dashboard API | org_id 隔离 via TenantScope | Implemented |
| T-08-02 | enroll/stop API | org_id 隔离 + go-playground/validator | Implemented |
| T-08-04 | payment-callback | HMAC 签名验证 | TODO (marked in code) |
| T-08-05 | scheduler | asynq 幂等 UPSERT | Implemented |

## Self-Check: PASSED

All 9 files verified present. All 3 commits (f009ab7, 71f8198, 4276b75) verified in git log.
