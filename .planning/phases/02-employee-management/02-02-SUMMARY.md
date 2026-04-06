---
phase: 02-employee-management
plan: 02
subsystem: employee
tags: [invitation, onboarding, token, crypto]
dependency_graph:
  requires:
    - "Phase 2 Plan 01: Employee model, repository, service"
  provides:
    - "Invitation model (internal/employee)"
    - "Invitation CRUD + token-based public access"
    - "Employee self-service onboarding via invitation link"
  affects:
    - "cmd/server/main.go (AutoMigrate + DI + routes)"
tech_stack:
  added: []
  patterns:
    - "crypto/rand 32-byte hex token for invitation links"
    - "7-day expiration with one-time use enforcement"
    - "Public endpoint (no auth) for employee self-service"
key_files:
  created:
    - "internal/employee/invitation_model.go"
    - "internal/employee/invitation_dto.go"
    - "internal/employee/invitation_repository.go"
    - "internal/employee/invitation_service.go"
    - "internal/employee/invitation_handler.go"
    - "internal/employee/invitation_service_test.go"
  modified:
    - "cmd/server/main.go"
decisions:
  - "Token generated via crypto/rand hex encoding (32 bytes) for security"
  - "Invitation status: pending → submitted → confirmed/expired"
  - "Public endpoint for employee to view invitation and submit info"
metrics:
  tasks_completed: 2
  files_created: 6
  files_modified: 1
  tests_passed: 6
  completed_date: "2026-04-07"
---

# Phase 2 Plan 02: Employee Invitation / Onboarding Summary

Invitation-based onboarding system: boss creates invitation (crypto/rand token + 7-day expiry), employee views and submits personal info via public link, system auto-creates Employee record (status=pending), boss confirms to activate.

## Tasks Completed

### Task 1: Invitation Model + Repository + Service

- Created Invitation model with token, expiry, status, employee info fields
- Repository with CRUD + token lookup + expiry check
- Service layer: CreateInvitation (generate token), GetByToken, SubmitEmployeeInfo, ConfirmInvitation
- Token security: crypto/rand 32-byte hex, one-time use, 7-day expiry

### Task 2: Invitation Handler + Route Registration

- Handler with endpoints: create invitation, get invitation detail (public), submit info (public), confirm invitation
- Route registration in main.go with public + authenticated routes
- 6 service unit tests all passing

## Files Created

| File | Purpose |
|------|---------|
| `internal/employee/invitation_model.go` | Invitation struct with token, expiry, status |
| `internal/employee/invitation_dto.go` | Request/Response DTOs |
| `internal/employee/invitation_repository.go` | GORM repository for invitations |
| `internal/employee/invitation_service.go` | Business logic: token gen, submit, confirm |
| `internal/employee/invitation_handler.go` | HTTP handlers with public/auth routes |
| `internal/employee/invitation_service_test.go` | 6 service unit tests |

## Self-Check: PASSED
