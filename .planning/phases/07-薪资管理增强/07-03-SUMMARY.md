---
phase: 07
plan: 03
subsystem: 薪资管理增强
tags: [salary, tax-upload, slip-send, asynq, excel, h5]
dependency_graph:
  requires: [salary-module, salary-slip-existing]
  provides: [tax-upload-api, tax-upload-frontend, batch-slip-send-api, slip-send-frontend, slip-h5-frontend]
  affects: [internal/salary, cmd/server/main.go, frontend/src/api, frontend/src/views, frontend/src/router]
tech_stack:
  added: [asynq]
  patterns: [excel-parsing, name-matching, asynq-queue, mobile-h5]
key_files:
  created:
    - internal/salary/tax_upload_service.go
    - internal/salary/tax_upload_handler.go
    - internal/salary/slip_send_task.go
    - internal/salary/slip_send_service.go
    - internal/salary/slip_send_handler.go
    - frontend/src/views/tool/TaxUpload.vue
    - frontend/src/views/tool/SalarySlipSend.vue
    - frontend/src/views/tool/SalarySlipH5.vue
  modified:
    - internal/salary/repository.go
    - cmd/server/main.go
    - frontend/src/api/salary.ts
    - frontend/src/views/tool/SalaryTool.vue
    - frontend/src/router/index.ts
decisions:
  - SlipSendResult renamed to SlipSendResult (vs existing SendSlipResult in slip.go) to avoid redeclaration
  - generateSlipToken reused from existing service.go instead of duplicating
  - TaxUploadHandler and SlipSendHandler registered separately in main.go
  - H5 slip page uses /salary/slip/:token as independent route (no AppLayout), exempt from auth guard
metrics:
  duration: 22min
  completed: "2026-04-18"
  tasks: 4
  files: 15
---

# Phase 07 Plan 03: 个税上传 + 工资条发送 Summary

后端：个税 Excel 上传解析 + asynq 批量工资条发送队列。前端：TaxUpload.vue + SalarySlipSend.vue + SalarySlipH5.vue。

## Tasks Completed

| Task | Name | Commit | Key Files |
|------|------|--------|-----------|
| 1 | Tax Upload Backend | 22929db | tax_upload_service.go, tax_upload_handler.go |
| 2 | SlipSendLog Repository | 39dbdab | repository.go |
| 3 | asynq Batch Slip Send | 97de198 | slip_send_task.go, slip_send_service.go, slip_send_handler.go, main.go |
| 4 | Frontend Pages | 8030005 | TaxUpload.vue, SalarySlipSend.vue, SalarySlipH5.vue, salary.ts, router/index.ts, SalaryTool.vue |

## Commits

- `22929db` feat(07-03): add tax upload backend with Excel parsing and name matching
- `39dbdab` feat(07-03): add SalarySlipSendLog repository methods
- `97de198` feat(07-03): add asynq batch slip send with worker and handler
- `8030005` feat(07-03): add TaxUpload, SalarySlipSend, SalarySlipH5 pages and routes

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Conflict] SendSlipResult redeclared**
- **Found during:** Task 3 (slip_send_service.go)
- **Issue:** `SendSlipResult` and `generateSlipToken` both already exist in slip.go/service.go
- **Fix:** Renamed local type to `SlipSendResult`; deleted duplicate `generateSlipToken` (reused from service.go)
- **Files modified:** slip_send_service.go
- **Commit:** 97de198

**2. [Rule 3 - Conflict] h.SendSlip undefined**
- **Found during:** Task 3 (slip_send_handler.go)
- **Issue:** `SlipSendHandler` registered `/send` route pointing to non-existent method
- **Fix:** Removed the `/send` route; single-send already exists in main handler.go
- **Files modified:** slip_send_handler.go
- **Commit:** 97de198

## Decisions Made

1. **SlipSendResult vs SendSlipResult** - The existing `SendSlipResult` in slip.go has different fields. Created a new `SlipSendResult` in slip_send_service.go for asynq worker use.

2. **generateSlipToken deduplication** - Removed duplicate from slip_send_service.go; the existing function in service.go is accessible within the salary package.

3. **Separate Handler registration** - `TaxUploadHandler` and `SlipSendHandler` registered independently in main.go rather than extending the existing salary Handler, keeping concerns separated.

4. **H5 auth bypass** - `/salary/slip/:token` added to the unauthenticated route list in router guard since it is a public employee-facing page.

## Backend Verification

- `go build ./cmd/server/` — PASSED (no errors)
- `go get github.com/hibiken/asynq@latest` — added asynq v0.26.0

## Frontend Verification

- `vue-tsc --noEmit` — PASSED (no type errors)

## Self-Check: PASSED

- All 8 created files verified as FOUND
- All 4 commits verified in git log
- No accidental file deletions in any commit
- Backend compiles: `go build ./cmd/server/` passes
- Frontend compiles: `vue-tsc --noEmit` passes

## Known Stubs

| File | Line | Description |
|------|------|-------------|
| TaxUpload.vue | 169 | `downloadSample()` shows message only; actual sample file download not implemented |
| SalarySlipSend.vue | 74 | `loadEmployeeList()` loads payroll records; employee names may be empty if no payroll exists for month |
| SalarySlipH5.vue | 130 | Error messages use generic "工资条链接无效" for non-404/403 errors |
