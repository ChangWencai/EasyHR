---
phase: 02-employee-management
plan: 04
subsystem: employee
tags: [contract, pdf, oss, lifecycle]
dependency_graph:
  requires:
    - "Phase 2 Plan 01: Employee model, repository, service"
  provides:
    - "Contract model with full lifecycle (draft → signed → expired → terminated)"
    - "PDF template generation (fpdf + Chinese font support)"
    - "Contract CRUD + status transitions"
    - "OSS upload integration point for signed scans"
  affects:
    - "cmd/server/main.go (AutoMigrate + DI + routes)"
    - "go.mod (fpdf dependency)"
tech_stack:
  added:
    - "go-pdf/fpdf v0.9.0 (PDF generation)"
  patterns:
    - "One employee can have multiple contracts (renewal scenario)"
    - "PDF template auto-fills company and employee info"
key_files:
  created:
    - "internal/employee/contract_model.go"
    - "internal/employee/contract_dto.go"
    - "internal/employee/contract_repository.go"
    - "internal/employee/contract_service.go"
    - "internal/employee/contract_handler.go"
    - "internal/employee/contract_service_test.go"
    - "internal/employee/pdf.go"
  modified:
    - "cmd/server/main.go"
    - "go.mod"
decisions:
  - "Contract lifecycle: draft → signed → active → expired/terminated"
  - "One employee can have multiple contracts (renewal)"
  - "PDF generation via fpdf with Chinese font support"
  - "OSS upload integration point prepared for signed document scans"
metrics:
  tasks_completed: 2
  files_created: 7
  files_modified: 2
  tests_passed: 6
  completed_date: "2026-04-07"
---

# Phase 2 Plan 04: Contract Lifecycle Management Summary

Contract full lifecycle management: create contract (draft) → generate PDF template (fpdf + Chinese font, auto-fill company/employee info) → download → sign → upload to OSS → status transitions. Supports multiple contracts per employee for renewal scenarios.

## Tasks Completed

### Task 1: Contract Model + Repository + Service

- Created Contract model with type, status, dates, file paths
- Repository with CRUD + employee-scoped queries
- Service: CreateContract, GeneratePDF, UpdateStatus, ListByEmployee
- PDF generation with fpdf: auto-fill company name, employee name, dates, contract terms

### Task 2: Contract Handler + Route Registration

- Handler with endpoints: create, generate PDF, download, update status, list by employee
- Route registration in main.go with RBAC (OWNER/ADMIN for create/edit/terminate)
- 6 service unit tests all passing

## Files Created

| File | Purpose |
|------|---------|
| `internal/employee/contract_model.go` | Contract struct with lifecycle statuses |
| `internal/employee/contract_dto.go` | Request/Response DTOs |
| `internal/employee/contract_repository.go` | GORM repository |
| `internal/employee/contract_service.go` | Business logic + PDF generation |
| `internal/employee/contract_handler.go` | HTTP handlers |
| `internal/employee/contract_service_test.go` | 6 service unit tests |
| `internal/employee/pdf.go` | PDF template generation with fpdf |

## Self-Check: PASSED
