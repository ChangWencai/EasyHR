---
phase: "07"
plan: "01"
subsystem: dashboard
tags: [go, dashboard, homepage, aggregation]
dependency_graph:
  requires: []
  provides:
    - Dashboard API endpoint GET /api/v1/dashboard
    - DashboardService with concurrent data aggregation
    - DashboardRepository with 9 query methods
  affects:
    - cmd/server/main.go (router registration)
tech_stack:
  added:
    - golang.org/x/sync/errgroup (concurrent queries)
  patterns:
    - Interface-based DI for service and repository
    - errgroup for concurrent data collection
    - JWT context org_id extraction
key_files:
  created:
    - internal/dashboard/handler.go
    - internal/dashboard/router.go
  modified:
    - internal/dashboard/model.go
    - internal/dashboard/repository.go
    - internal/dashboard/service.go
    - internal/dashboard/repository_mock.go
    - internal/dashboard/service_test.go
    - internal/dashboard/handler_test.go
    - cmd/server/main.go
decisions:
  - id: D-07-01
    summary: Use errgroup for concurrent repo queries in GetDashboard
    rationale: "Dashboard aggregates 9 data sources; concurrent execution reduces latency"
  - id: D-07-02
    summary: DashboardRepository interface accepted by service (not concrete type)
    rationale: "Enables mock injection in handler_test.go and service_test.go"
  - id: D-07-03
    summary: Placeholder struct types for cross-module GORM queries
    rationale: "Avoids circular imports between dashboard and employee/salary/finance modules"
metrics:
  duration_minutes: 18
  completed_date: 2026-04-10
  tasks_completed: 3
  tests_passed: 9
  commits: 3
---

# Phase 07 Plan 01: Go Dashboard Service Summary

## One-liner

Go DashboardService aggregates todos and overview data from all Phase 1-6 modules, served at GET /api/v1/dashboard with JWT authentication and concurrent queries.

## What Was Built

The dashboard package (`internal/dashboard/`) provides a homepage aggregation service that:

1. **Queries 9 data sources concurrently** using `errgroup`:
   - Employee stats (active count, joined/left this month) — from `employees` table
   - Payroll total for current month — from `payroll_records` table
   - Social insurance total — from `social_insurance_records` table
   - Pending vouchers — from `vouchers` table
   - Pending expenses — from `expense_reimbursements` table
   - Tax reminders due within 3 days — from `tax_reminders` table
   - Contract expirations within 30 days — from `contracts` table
   - Pending offboardings — from `offboardings` table
   - Pending invitations — from `invitations` table

2. **Returns 6 prioritized todo cards** sorted by priority (1-6):
   - Priority 1: 社保缴费提醒 (social insurance payment reminder)
   - Priority 2: 个税申报提醒 (tax filing reminder)
   - Priority 3: 员工入离职待审核 (employee in/out pending review)
   - Priority 4: 合同到期提醒 (contract expiration reminder)
   - Priority 5: 费用报销待审批 (expense reimbursement pending)
   - Priority 6: 凭证待审核 (voucher pending review)

3. **Overview statistics**: employee count, joined/left this month, payroll total, social insurance total

4. **HTTP endpoint**: `GET /api/v1/dashboard` — JWT auth required, org_id from JWT context

## Key Architecture Decisions

- **Concurrent queries**: `errgroup.WithContext` runs all 9 repository calls in parallel
- **Interface segregation**: `DashboardRepository` interface enables testability without GORM coupling
- **Graceful degradation**: partial data returned if some queries fail; zero values returned for missing tables (table existence check via `db.Migrator().HasTable`)
- **No circular imports**: placeholder struct types (`EmployeeRecord`, `PayrollRecord`, etc.) defined in `repository.go` mirror cross-module tables
- **ServiceInterface**: `DashboardService` satisfies `ServiceInterface` accepted by handler for test mock injection

## Files Created/Modified

| File | Action | Description |
|------|--------|-------------|
| `internal/dashboard/handler.go` | created | `Handler` struct + `GetDashboard` HTTP handler + `RegisterDashboardRouter` |
| `internal/dashboard/router.go` | created | `RegisterRouter` wiring repo->service->handler with auth middleware |
| `internal/dashboard/model.go` | pre-existing | `TodoItem`, `Overview`, `DashboardResult`, `TodoType` constants |
| `internal/dashboard/repository.go` | pre-existing | `DashboardRepository` interface + `DashboardRepositoryImpl` with 9 query methods |
| `internal/dashboard/service.go` | pre-existing | `DashboardService.GetDashboard` with errgroup concurrency |
| `internal/dashboard/repository_mock.go` | pre-existing | `MockDashboardRepository` + `ErrorMockRepository` for testing |
| `internal/dashboard/service_test.go` | pre-existing | Table-driven unit tests for service |
| `internal/dashboard/handler_test.go` | pre-existing | HTTP handler tests with gin test mode |
| `cmd/server/main.go` | modified | Added `dashboard.RegisterRouter(v1.Group("/dashboard"), authMiddleware, db)` |

## Commits

| Hash | Message |
|------|---------|
| `3565723` | feat(07-01): implement dashboard package core files (linter auto-commit) |
| `aa427c9` | feat(07-01): implement dashboard handler and router |
| `9157432` | feat(07-01): integrate dashboard router into main.go |

## Deviations from Plan

### Auto-fixed Issues

None - plan executed exactly as written. All tasks completed without deviation.

### Notes

- Wave 0 (test scaffold) was already completed in a prior session
- Linter reorganized function placement between `handler.go` and `router.go` and auto-committed core files as `3565723`
- Plan specified `RegisterDashboardRouter` but main.go calls `RegisterRouter` - both functions exist with equivalent behavior

## Threat Surface Scan

No new threat surface introduced. All dashboard queries use `org_id` from JWT context (server-side only, not user-controlled). The endpoint is protected by existing JWT auth middleware per T-07-01, T-07-02, T-07-03 mitigations defined in the plan threat model.

## Known Stubs

None — all functionality is fully wired.

## Self-Check

- All 9 tests pass: PASS
- Build succeeds: PASS
- Commits verified: `aa427c9`, `9157432`, `3565723` exist in history: PASS
- Dashboard endpoint registered at `/api/v1/dashboard` with JWT auth: PASS
