---
phase: 02-employee-management
plan: 01
subsystem: employee
tags: [model, crud, search, encryption, rbac, excel-export]
dependency_graph:
  requires:
    - "Phase 1: common/crypto, common/middleware, common/response, common/model"
  provides:
    - "Employee model (internal/employee)"
    - "Employee CRUD API (POST/GET/PUT/DELETE /api/v1/employees)"
    - "Employee search (name/position LIKE + phone hash + status filter)"
    - "Excel export (excelize, masked sensitive fields)"
    - "RBAC routes (OWNER/ADMIN full, MEMBER read-only masked)"
  affects:
    - "cmd/server/main.go (AutoMigrate + DI + routes)"
    - "test/testutil/db.go (Employee AutoMigrate + CreateTestEmployee)"
tech_stack:
  added:
    - "excelize v2.10.1 (Excel export)"
  patterns:
    - "Dual-column encryption (encrypted + hash index)"
    - "Repository-layer transaction uniqueness check (SQLite compatible)"
    - "Service-layer mask/decrypt for API responses"
key_files:
  created:
    - "internal/employee/model.go"
    - "internal/employee/dto.go"
    - "internal/employee/repository.go"
    - "internal/employee/service.go"
    - "internal/employee/handler.go"
    - "internal/employee/repository_test.go"
    - "internal/employee/service_test.go"
  modified:
    - "cmd/server/main.go"
    - "test/testutil/db.go"
decisions:
  - "Repository layer transaction-based uniqueness check for phone_hash and id_card_hash (SQLite test compatibility, PostgreSQL partial unique index in production)"
  - "LIKE (not ILIKE) used for name/position search in repository for SQLite test compatibility"
  - "StatusProbation constant added for full employee lifecycle (pending/probation/active/resigned)"
metrics:
  duration: "10 min"
  tasks_completed: 2
  files_created: 7
  files_modified: 2
  tests_passed: 19
  completed_date: "2026-04-07"
---

# Phase 2 Plan 01: Employee Model + CRUD + Search + Excel Export Summary

Employee data model with AES-256-GCM dual-column encryption, full CRUD with multi-tenant isolation, 4-dimension search (name/position LIKE + phone hash exact + status filter), paginated list, Excel export via excelize with masked sensitive fields, and RBAC route protection (OWNER/ADMIN full access, MEMBER read-only masked data).

## Tasks Completed

### Task 1: Employee Model + Repository + Unit Tests

- Created Employee model with all specified fields (name, phone/idcard/bank dual-column encrypted, gender, birth_date, position, hire_date, status, user_id, bank/emergency/address/remark, resignation fields)
- Added `extractFromIDCard()` function for automatic gender and birth date extraction from 18-digit ID card
- Implemented Repository layer with transaction-based uniqueness validation (SQLite compatible)
- CRUD operations: Create, FindByID, Update, Delete (soft delete)
- Search with dynamic filters: name LIKE, position LIKE, phone hash exact, status exact
- Pagination with total count
- FindAllForExport for Excel export (no pagination)
- 9 repository unit tests all passing

### Task 2: Employee Service + Handler + Excel Export + Route Registration

- Created DTO layer: CreateEmployeeRequest, UpdateEmployeeRequest, EmployeeResponse (masked), SensitiveInfoResponse, ListQueryParams
- Service layer: CreateEmployee (encrypt + extract ID card), ListEmployees (masked), GetEmployee, UpdateEmployee (partial update with re-encryption), DeleteEmployee, GetSensitiveInfo (full decrypt), ExportExcel (excelize with freeze panes)
- Handler layer with 7 endpoints and RBAC via RequireRole middleware
- Registered employee routes in cmd/server/main.go with dependency injection
- Updated test/testutil/db.go with Employee AutoMigrate and CreateTestEmployee helper
- 10 service unit tests all passing

## Files Created

| File | Purpose |
|------|---------|
| `internal/employee/model.go` | Employee struct with encrypted dual-column fields, status constants, extractFromIDCard |
| `internal/employee/dto.go` | Request/Response DTOs for employee CRUD |
| `internal/employee/repository.go` | GORM repository with CRUD, search, pagination, multi-tenant isolation |
| `internal/employee/service.go` | Business logic: encrypt, decrypt, mask, Excel export |
| `internal/employee/handler.go` | HTTP handlers with RBAC route registration |
| `internal/employee/repository_test.go` | 9 repository unit tests |
| `internal/employee/service_test.go` | 10 service unit tests |

## Files Modified

| File | Change |
|------|--------|
| `cmd/server/main.go` | Added employee import, AutoMigrate, DI (empRepo/empSvc/empHandler), route registration |
| `test/testutil/db.go` | Added employee import, Employee AutoMigrate, CreateTestEmployee helper |

## Verification Results

- `go test ./internal/employee/... -v` : 19/19 PASS
- `go build ./cmd/server` : SUCCESS
- Employee model covers all CONTEXT.md D-06~D-10 fields
- Search supports 4 dimensions: name/position LIKE + phone hash + status
- Excel export generates valid xlsx (PK header verified)
- RBAC: MEMBER read-only masked, OWNER/ADMIN full access

## API Endpoints

| Method | Path | Roles | Description |
|--------|------|-------|-------------|
| POST | /api/v1/employees | owner, admin | Create employee |
| GET | /api/v1/employees | all (masked) | List/search employees |
| GET | /api/v1/employees/export | owner, admin | Export Excel |
| GET | /api/v1/employees/:id | all (masked) | Get employee detail |
| PUT | /api/v1/employees/:id | owner, admin | Update employee |
| DELETE | /api/v1/employees/:id | owner, admin | Delete employee |
| POST | /api/v1/employees/:id/sensitive | owner, admin | View sensitive info |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Fixed handler.go CreateEmployee call signature mismatch**
- **Found during:** Task 2 compilation
- **Issue:** handler.go passed `context.Context` as first arg but service method signature had no ctx parameter
- **Fix:** Removed `c.Request.Context()` from handler call to match service method signature
- **Files modified:** internal/employee/handler.go

**2. [Rule 3 - Blocking] Fixed excelize.Panes field name TopLeft -> TopLeftCell**
- **Found during:** Task 2 compilation
- **Issue:** excelize v2.10.1 uses `TopLeftCell` not `TopLeft`
- **Fix:** Changed field name in service.go ExportExcel
- **Files modified:** internal/employee/service.go

**3. [Rule 2 - Missing] Added StatusProbation constant**
- **Found during:** Task 1 review
- **Issue:** CONTEXT.md D-07 defines 4 lifecycle states but original code only had 3
- **Fix:** Added StatusProbation = "probation" constant
- **Files modified:** internal/employee/model.go

## Known Stubs

None. All data flows are fully wired -- no hardcoded empty values, no placeholder text, no disconnected components.

## Self-Check: PASSED

- All 9 created/modified files verified to exist on disk
- `go test ./internal/employee/... -v`: 19/19 PASS
- `go build ./cmd/server`: SUCCESS
