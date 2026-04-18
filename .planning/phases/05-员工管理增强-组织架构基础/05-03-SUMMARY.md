---
phase: 05-员工管理增强-组织架构基础
plan: 03
subsystem: api, ui
tags: [registration, token, gorm, vue3, element-plus, qrcode, sms]

requires:
  - phase: 05-02
    provides: "Employee model with encrypted fields, Repository with FindByPhoneHash/FindByIDCardHash"
provides:
  - "Registration backend module (model/dto/repository/service/handler)"
  - "Public API for employees to submit info via token link (no auth required)"
  - "Admin API for CRUD registration records (auth + role check)"
  - "SMS SendTemplateMessage method for custom template sending"
  - "4 frontend components: RegistrationList, RegistrationCreate, RegistrationForwardDialog, RegisterPage"
affects: [05-04, 05-05]

tech-stack:
  added: [qrcode (npm)]
  patterns: ["Token-based public form submission with transactional employee upsert", "Mixed public/auth route registration pattern"]

key-files:
  created:
    - internal/employee/registration_model.go
    - internal/employee/registration_dto.go
    - internal/employee/registration_repository.go
    - internal/employee/registration_service.go
    - internal/employee/registration_handler.go
    - frontend/src/views/employee/RegistrationList.vue
    - frontend/src/views/employee/RegistrationCreate.vue
    - frontend/src/views/employee/RegistrationForwardDialog.vue
    - frontend/src/views/employee/RegisterPage.vue
  modified:
    - pkg/sms/client.go
    - cmd/server/main.go
    - frontend/src/api/employee.ts
    - frontend/package.json

key-decisions:
  - "Registration uses same generateToken() pattern as Invitation (crypto/rand 32-byte hex)"
  - "SubmitRegistration does transactional upsert: find existing employee by phone_hash/id_card_hash, overwrite or create"
  - "Public routes (GET/POST /registrations/:token) placed before auth middleware, admin routes under auth group"
  - "SMS SendTemplateMessage method added for custom template support, reusing existing signing logic"

patterns-established:
  - "Mixed public/auth route pattern: rg.GET/POST for public, authGroup.POST/GET/DELETE for admin"
  - "Transactional employee upsert: find by hash -> create or overwrite with latest data"
  - "Token-based public form: no auth required, token validation + expiry check + status check"

requirements-completed: [EMP-05, EMP-06, EMP-07, EMP-08]

duration: 10min
completed: 2026-04-18
---

# Phase 05 Plan 03: Employee Registration Module Summary

**Token-based employee info registration with public H5 form, admin CRUD, and transactional employee upsert via crypto/rand tokens**

## Performance

- **Duration:** 10 min
- **Started:** 2026-04-18T03:09:47Z
- **Completed:** 2026-04-18T03:20:23Z
- **Tasks:** 2
- **Files modified:** 12

## Accomplishments
- Complete Registration backend module: model/dto/repository/service/handler with token-based public submission
- Public API for employees to submit info without login (GET/POST /registrations/:token)
- Admin API for managing registration records (POST/GET/DELETE /registrations)
- SubmitRegistration transaction: upserts employee data with encrypted sensitive fields (AES-256-GCM + SHA-256)
- SMS SendTemplateMessage extension for custom template sending
- 4 frontend components: management list, create dialog, forward dialog (QR+copy+SMS), standalone H5 register page

## Task Commits

1. **Task 1: Backend Registration module + SMS extension** - `ac191be` (feat)
2. **Task 2: Frontend registration components** - `d8a7dff` (feat)

## Files Created/Modified
- `internal/employee/registration_model.go` - Registration model with Token/Status/ExpiresAt, 7-day expiry
- `internal/employee/registration_dto.go` - Request/Response DTOs for registration CRUD and public submission
- `internal/employee/registration_repository.go` - Data access with public token lookup and tenant-scoped admin queries
- `internal/employee/registration_service.go` - Business logic: CreateRegistration, SubmitRegistration (transactional upsert), ListRegistrations, DeleteRegistration
- `internal/employee/registration_handler.go` - HTTP handlers with mixed public/auth routes
- `pkg/sms/client.go` - Added SendTemplateMessage method for custom SMS templates
- `cmd/server/main.go` - DI registration, route registration, AutoMigrate for Registration model
- `frontend/src/api/employee.ts` - Added registrationApi with full CRUD + submit
- `frontend/src/views/employee/RegistrationList.vue` - Management list with status tags, forward, delete
- `frontend/src/views/employee/RegistrationCreate.vue` - Dialog form with department select
- `frontend/src/views/employee/RegistrationForwardDialog.vue` - QR code + copy link + SMS forwarding
- `frontend/src/views/employee/RegisterPage.vue` - Standalone mobile H5 page for employees

## Decisions Made
- Reused Invitation module's generateToken() function (same package employee, already exported)
- Registration token uses crypto/rand 32-byte hex (2^256 space) matching Invitation pattern
- SubmitRegistration checks phone_hash then id_card_hash to find existing employees, overwrites with latest data per D-08 requirement
- Public GET/POST routes registered on unprotected router group, admin routes on auth group with RequireRole("owner", "admin")
- SMS forwarding in frontend is placeholder (TODO: backend SMS API integration) since SMS template needs to be configured in Aliyun console first
- QR code uses qrcode npm package with canvas rendering

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- git add failed for cmd/server/main.go due to `server` pattern in .gitignore; resolved with `git add -f`
- Required installing qrcode + @types/qrcode npm packages for RegistrationForwardDialog QR code generation

## User Setup Required

**External services require manual configuration:**
- Aliyun SMS: Create SMS signature and template in Aliyun SMS console for registration link sending
- Environment variables: SMS_ACCESS_KEY_ID, SMS_ACCESS_KEY_SECRET
- SMS template example: "易人事提醒您，请填写员工信息登记表，链接: ${link}，有效期7天。"

## Next Phase Readiness
- Registration module complete, ready for integration with Plan 05-04/05-05
- Employee data flow: Registration -> Employee record with encrypted fields
- QR code forwarding requires Aliyun SMS template configuration for full functionality

## Self-Check: PASSED

- All 13 files verified to exist on disk
- Both task commits (ac191be, d8a7dff) verified in git log
- `go build ./...` passes with 0 errors
- `npx vue-tsc --noEmit` passes with 0 errors
- All 14 acceptance criteria from PLAN.md verified

---
*Phase: 05-员工管理增强-组织架构基础*
*Completed: 2026-04-18*
