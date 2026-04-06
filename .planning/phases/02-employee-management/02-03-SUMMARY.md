---
phase: 02-employee-management
plan: 03
subsystem: employee
tags: [offboarding, resignation, checklist, event]
dependency_graph:
  requires:
    - "Phase 2 Plan 01: Employee model, repository, service"
  provides:
    - "Offboarding model with checklist (internal/employee)"
    - "Boss-initiated and employee-initiated resignation flows"
    - "Template-based handover checklist (asset/work/permission)"
    - "OnEmployeeResigned event interface for downstream phases"
  affects:
    - "cmd/server/main.go (AutoMigrate + DI + routes)"
tech_stack:
  added: []
  patterns:
    - "Default checklist template (asset return, work handover, permission revoke)"
    - "Event interface for cross-phase triggers (Phase 3 social insurance)"
key_files:
  created:
    - "internal/employee/offboarding_model.go"
    - "internal/employee/offboarding_dto.go"
    - "internal/employee/offboarding_repository.go"
    - "internal/employee/offboarding_service.go"
    - "internal/employee/offboarding_handler.go"
    - "internal/employee/offboarding_service_test.go"
  modified:
    - "cmd/server/main.go"
decisions:
  - "Default checklist items: asset return, work handover, permission revoke"
  - "Boss can edit/supplement checklist items before completion"
  - "OnEmployeeResigned event interface for Phase 3 social insurance integration"
  - "Employee status transitions to resigned on offboarding completion"
metrics:
  tasks_completed: 2
  files_created: 6
  files_modified: 1
  tests_passed: 6
  completed_date: "2026-04-07"
---

# Phase 2 Plan 03: Offboarding / Resignation Management Summary

Full offboarding workflow: boss-initiated or employee-initiated resignation, template-based handover checklist (asset/work/permission), boss can edit checklist, employee status updates to resigned on completion, event interface for downstream phase integration.

## Tasks Completed

### Task 1: Offboarding Model + Repository + Service

- Created Offboarding model with checklist items, resignation type, reason, dates
- Repository with CRUD + org-scoped queries
- Service: BossResign, EmployeeResign, UpdateChecklist, CompleteOffboarding
- Default checklist generation: asset return, work handover, permission revoke
- OnEmployeeResigned event interface

### Task 2: Offboarding Handler + Route Registration

- Handler with endpoints: boss resign, employee resign, update checklist, complete offboarding, get detail
- Route registration in main.go with RBAC (OWNER/ADMIN for resign)
- 6 service unit tests all passing

## Files Created

| File | Purpose |
|------|---------|
| `internal/employee/offboarding_model.go` | Offboarding + ChecklistItem structs |
| `internal/employee/offboarding_dto.go` | Request/Response DTOs |
| `internal/employee/offboarding_repository.go` | GORM repository |
| `internal/employee/offboarding_service.go` | Business logic with event interface |
| `internal/employee/offboarding_handler.go` | HTTP handlers |
| `internal/employee/offboarding_service_test.go` | 6 service unit tests |

## Self-Check: PASSED
