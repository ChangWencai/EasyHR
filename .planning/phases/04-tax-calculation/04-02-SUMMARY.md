---
phase: 04-tax-calculation
plan: 02
subsystem: tax-api
tags: [tax, gin, gocron, excelize, fpdf, handler, scheduler, excel-export, pdf-export, rbac]

# Dependency graph
requires:
  - phase: 04-tax-calculation
    provides: "Tax Service layer with CalculateTax, deduction CRUD, declaration management"
provides:
  - "Complete tax HTTP API with RBAC (OWNER/ADMIN mutations, MEMBER read-only)"
  - "Daily gocron scheduler for tax declaration reminders (12th-15th of month)"
  - "Excel export aligned with Natural Person Tax Bureau 15-column format"
  - "PDF tax certificate with employee/org details and tax breakdown"
  - "TaxReminder model with deduplication for declaration_due type"
  - "Cross-module adapters: EmployeeAdapter and SocialInsuranceAdapter"
  - "main.go integration: AutoMigrate, DI, routes, scheduler, bracket seeding"
affects: [05-salary-calculation, frontend-tax-pages]

# Tech tracking
tech-stack:
  added: [gocron/v2, excelize/v2, go-pdf/fpdf]
  patterns: [cross-module-adapter-in-main, redis-distributed-scheduler, excel-export-template, pdf-certificate-template]

key-files:
  created:
    - internal/tax/handler.go
    - internal/tax/scheduler.go
    - internal/tax/excel.go
    - internal/tax/pdf.go
    - internal/tax/employee_adapter.go
    - internal/tax/si_adapter.go
  modified:
    - internal/tax/model.go
    - internal/tax/repository.go
    - internal/tax/service.go
    - cmd/server/main.go

key-decisions:
  - "ContractRepo moved before tax module in DI chain to resolve dependency ordering"
  - "TaxReminder deduplication by org_id + year + month (one reminder per org per month)"
  - "Excel format uses cumulative taxable income column matching tax bureau reporting"
  - "PDF uses Helvetica font (V1.0), consistent with other modules (socialinsurance, employee)"
  - "GetMyTaxRecords returns error stub for Phase 5 user-employee mapping completion"

patterns-established:
  - "Cross-module adapter pattern: tax.EmployeeAdapter wraps employee.ContractRepository + employee.Repository"
  - "Handler export pattern: Content-Type + Content-Disposition headers + c.Data() for binary response"
  - "Scheduler pattern: redis-distributed lock + daily 08:00 CST + service method callback"

requirements-completed: [TAX-03, TAX-04, TAX-05, TAX-06]

# Metrics
duration: 12min
completed: 2026-04-07
---

# Phase 4 Plan 2: Tax API + Scheduler + Export Summary

**Tax module HTTP API with 20 endpoints (RBAC-enforced), gocron daily declaration reminders, Excel export aligned with Natural Person Tax Bureau format, and PDF tax certificate generation**

## Performance

- **Duration:** 12 min
- **Started:** 2026-04-07T16:35:28Z
- **Completed:** 2026-04-07T16:47:47Z
- **Tasks:** 2
- **Files modified:** 10

## Accomplishments
- Complete HTTP API with 20 endpoints covering brackets, deductions, calculation, records, declarations, reminders, and export
- RBAC enforcement: OWNER/ADMIN for mutations, OWNER for bracket seeding, MEMBER read-only for own records
- Daily gocron scheduler scanning tax declaration reminders on 12th-15th of each month with Redis distributed lock
- Excel export with 15-column format aligned with Natural Person Tax Bureau batch import template
- PDF tax certificate with employee info, org name, tax detail table, and print date
- Cross-module adapters: EmployeeAdapter (contract salary + hire date) and SocialInsuranceAdapter (personal deduction)
- main.go integration: AutoMigrate for 5 tax models, DI wiring, route registration, scheduler startup, bracket seeding

## Task Commits

Each task was committed atomically:

1. **Task 1: HTTP Handler + scheduler + adapters + main.go integration** - `557202f` (feat)
2. **Task 2: Excel export + PDF export implementation** - `9c8c339` (feat)

## Files Created/Modified
- `internal/tax/handler.go` - HTTP Handler with 20 endpoints and RegisterRoutes
- `internal/tax/scheduler.go` - gocron daily scheduler with Redis distributed lock
- `internal/tax/excel.go` - Excel export with 15-column tax declaration format
- `internal/tax/pdf.go` - PDF tax certificate with TaxCertificateData struct
- `internal/tax/employee_adapter.go` - EmployeeAdapter implementing EmployeeInfoProvider
- `internal/tax/si_adapter.go` - SocialInsuranceAdapter implementing SIDeductionProvider
- `internal/tax/model.go` - Added TaxReminder model with declaration_due type
- `internal/tax/repository.go` - Added reminder, aggregation, and export query methods
- `internal/tax/service.go` - Added CheckDeclarationReminders, export methods, reminder CRUD
- `cmd/server/main.go` - Integrated tax module: AutoMigrate, DI, routes, scheduler, bracket seeding

## Decisions Made
- ContractRepo DI ordering moved before tax module to resolve compile dependency (contractRepo needed by tax.EmployeeAdapter)
- TaxReminder deduplication uses org_id + year + month combination -- one reminder per org per month prevents duplicate alerts
- Excel format uses cumulative taxable income (not monthly) to match how tax bureau reporting works
- PDF uses Helvetica font (V1.0 limitation) consistent with socialinsurance and employee module patterns
- GetMyTaxRecords returns a descriptive error stub since user-to-employee mapping needs Phase 5 employee service integration

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Reordered DI chain in main.go**
- **Found during:** Task 1 (main.go integration)
- **Issue:** tax.NewEmployeeAdapter requires contractRepo, but contractRepo was defined after the tax DI block
- **Fix:** Moved contractRepo definition before tax module DI, reordering to: socialinsurance -> contracts -> tax -> offboarding
- **Files modified:** cmd/server/main.go
- **Verification:** go build ./cmd/server/ passes
- **Committed in:** 557202f (Task 1 commit)

---
**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Minor DI reordering, no scope creep.

## Issues Encountered
None - plan executed smoothly.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Tax module fully integrated and operational
- TaxCalculator interface ready for Phase 5 salary module to call directly
- GetMyTaxRecords needs Phase 5 user-employee mapping for MEMBER self-service queries
- All 20 API endpoints ready for frontend H5 management dashboard integration

## Self-Check: PASSED

- All 8 created/modified files verified present
- Commit 557202f (Task 1) verified in git log
- Commit 9c8c339 (Task 2) verified in git log
- go build ./cmd/server/ passes
- go test -race -count=1 ./internal/tax/... passes

---
*Phase: 04-tax-calculation*
*Completed: 2026-04-07*
