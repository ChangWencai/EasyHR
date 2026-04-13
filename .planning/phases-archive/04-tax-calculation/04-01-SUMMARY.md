---
phase: 04-tax-calculation
plan: 01
subsystem: tax
tags: [tax, cumulative-withholding, progressive-tax, special-deductions, gorm, sqlite-testing]

# Dependency graph
requires:
  - phase: 03-social-insurance
    provides: "SocialInsurance service pattern, EmployeeAdapter pattern, TenantScope"
provides:
  - "TaxCalculator interface for Phase 5 salary module to call tax calculation"
  - "calculateCumulativeTax pure function implementing cumulative withholding method"
  - "EmployeeInfoProvider/SIDeductionProvider cross-module interfaces"
  - "7-level progressive tax bracket seed data (OrgID=0 global)"
  - "Special deduction CRUD with mutual exclusion (housing_loan/housing_rent)"
  - "TaxRecord per-employee monthly snapshot for auditability"
affects: [05-salary-calculation, tax-api, tax-declaration]

# Tech tracking
tech-stack:
  added: []
  patterns: [cumulative-withholding-algorithm, cross-module-adapter-interfaces, global-vs-tenant-data-separation]

key-files:
  created:
    - internal/tax/model.go
    - internal/tax/dto.go
    - internal/tax/repository.go
    - internal/tax/calculator.go
    - internal/tax/calculator_test.go
    - internal/tax/service.go
    - internal/tax/adapter.go
    - internal/tax/errors.go
  modified: []

key-decisions:
  - "TaxBracket uses OrgID=0 global data pattern (same as SocialInsurancePolicy) for nationally standardized tax rates"
  - "Special deductions include 6 monthly types only (excluding serious illness per D-07 research)"
  - "Housing loan and housing rent are mutually exclusive per tax regulation"
  - "Monthly tax floor at 0 - no negative tax refund via withholding"
  - "roundTo2 applied to all intermediate calculation results for precision"
  - "TaxCalculator interface accepts grossIncome parameter for Phase 5 decoupling"

patterns-established:
  - "Cross-module interface pattern: EmployeeInfoProvider/SIDeductionProvider defined in consumer module"
  - "Pure calculation functions with dependency injection for testability"
  - "TaxRecord stores complete calculation snapshot for historical auditability"

requirements-completed: [TAX-01, TAX-02]

# Metrics
duration: 12min
completed: 2026-04-07
---

# Phase 4 Plan 1: Tax Calculation Engine Summary

**Cumulative withholding tax engine with 7-level progressive rates, special deduction CRUD with mutual exclusion, and cross-module adapter interfaces for Phase 5 integration**

## Performance

- **Duration:** 12 min
- **Started:** 2026-04-07T16:18:23Z
- **Completed:** 2026-04-07T16:30:33Z
- **Tasks:** 2
- **Files modified:** 8

## Accomplishments
- Complete cumulative withholding tax calculation engine per China tax regulation
- 7-level progressive tax bracket seed data with correct boundary handling (36000/144000/300000/etc.)
- 6 special deduction types with automatic monthly amount calculation and housing mutual exclusion
- 17 test cases covering basic scenarios, cumulative months, bracket jumps, boundary values, mid-year hire, and zero income
- Cross-module interfaces (EmployeeInfoProvider, SIDeductionProvider, TaxCalculator) for Phase 5 decoupling
- TaxRecord stores complete calculation snapshots for historical traceability

## Task Commits

Each task was committed atomically:

1. **Task 1: Tax data models + special deduction CRUD + tax bracket seed data** - `309d746` (feat)
2. **Task 2: Cumulative withholding calculator engine + Service layer + cross-module interfaces** - `687c803` (feat)

## Files Created/Modified
- `internal/tax/model.go` - TaxBracket, SpecialDeduction, TaxRecord, TaxDeclaration models with DeductionStandard mapping
- `internal/tax/dto.go` - All request/response DTOs for deductions, brackets, tax records, declarations
- `internal/tax/repository.go` - Full CRUD with TenantScope for tenant data and OrgID=0 for global data
- `internal/tax/errors.go` - Error definitions with 40xxx code range
- `internal/tax/calculator.go` - Cumulative withholding pure function + TaxCalculator interface
- `internal/tax/calculator_test.go` - 17 test cases with SQLite in-memory database
- `internal/tax/service.go` - Business logic layer: deduction management, tax calculation, declaration management
- `internal/tax/adapter.go` - EmployeeInfoProvider and SIDeductionProvider interface definitions

## Decisions Made
- TaxBracket uses OrgID=0 global data pattern (same as SocialInsurancePolicy) -- nationally standardized rates shared across all tenants
- Special deductions include only 6 monthly deduction types (excluding serious illness per D-07 research -- serious illness is settled annually, not monthly)
- Housing loan interest and housing rent are mutually exclusive per tax regulation pitfall #6
- Monthly tax minimum at 0 -- negative results from cumulative calculation are clamped to 0 (no refund through withholding, per regulation)
- roundTo2 precision applied to all intermediate calculation results to prevent floating point drift
- TaxCalculator interface accepts grossIncome as parameter -- Phase 5 passes salary in without creating circular dependency

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- TestCalculateTax_HighIncomeBracketJump had incorrect expected values (assumed 75000 falls in 3% bracket, but it falls in 10% bracket since 75000 > 36000). Fixed test expectations to match correct tax math.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Tax calculation engine is ready for API layer (04-02 plan)
- TaxCalculator interface ready for Phase 5 salary module integration
- EmployeeInfoProvider/SIDeductionProvider need concrete adapters wired in main.go during Phase 5
- SeedTaxBrackets should be called during application startup for the current year

---
*Phase: 04-tax-calculation*
*Completed: 2026-04-07*
