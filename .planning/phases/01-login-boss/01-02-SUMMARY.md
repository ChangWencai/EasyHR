---
phase: "01-login-boss"
plan: "02"
subsystem: auth
tags: [jwt, bcrypt, golang, gin, password-login]

# Dependency graph
requires:
  - phase: "01-login-boss"
    plan: "01"
    provides: Login SMS auth, JWT generation, onboarding flow
provides:
  - POST /api/v1/auth/login/password endpoint (phone + bcrypt password login)
  - GET /api/v1/auth/me endpoint (current user + org info)
  - MEMBER role 403 enforcement on all login methods
  - RegisterPassword service method for setting/resetting passwords
affects: [frontend-login, auth-guard, employee-module]

# Tech tracking
tech-stack:
  added: [golang.org/x/crypto/bcrypt]
  patterns: [bcrypt password hashing, JWT token generation, service/repository separation]

key-files:
  created: []
  modified:
    - internal/user/dto.go
    - internal/user/service.go
    - internal/user/handler.go
    - internal/user/repository.go
    - internal/common/model/user.go

key-decisions:
  - "bcrypt.DefaultCost used for password hashing (cost=10, production-ready)"
  - "PasswordHash stored as varchar(200) to accommodate bcrypt hashes"
  - "MEMBER role check added to both SMS Login and password Login methods"
  - "LoginPassword handler placed on non-authenticated route (rg.Group), GetMe on authenticated route (authGroup)"

patterns-established:
  - "Error string matching for MEMBER_ROLE_FORBIDDEN sentinel in handlers"

requirements-completed: [AUTH-01, AUTH-02, AUTH-03, AUTH-04]

# Metrics
duration: 25min
completed: 2026-04-11
---

# Phase 01-login-boss Plan 02 Summary

**Password login endpoint with bcrypt hashing + /auth/me API, MEMBER role returns 403 on all login methods**

## Performance

- **Duration:** ~25 min
- **Started:** 2026-04-11T12:45:00Z
- **Completed:** 2026-04-11T13:10:00Z
- **Tasks:** 5 (4 auto + 1 checkpoint:human-verify auto-approved)
- **Files modified:** 5

## Accomplishments
- POST /api/v1/auth/login/password returns JWT tokens (access_token, refresh_token, onboarding_required)
- GET /api/v1/auth/me returns user info (id, name, phone, role) + org info (id, name, credit_code, city) + onboarding_required
- OWNER/ADMIN login succeeds; MEMBER login returns HTTP 403 with message "您的账号为员工账号，请使用员工端微信小程序登录"
- MEMBER 403 enforcement added to both SMS login (POST /auth/login) and password login (POST /auth/login/password)
- Passwords stored as bcrypt hashes (golang.org/x/crypto/bcrypt, DefaultCost=10)
- PasswordHash field added to User model (GORM auto-migrate on next startup)
- RegisterPassword service method available for setting/resetting passwords

## Task Commits

1. **Task 1: Add password login DTO** - `ebf4cb4` (feat)
2. **Task 2: Add service logic** - `cea9d97` (feat)
3. **Task 3: Add handlers** - `57c5816` (feat)
4. **Task 4: Add repository** - `7042378` (feat)
5. **Fix model syntax** - `a10f7bf` (fix)

## Files Created/Modified

- `internal/user/dto.go` - Added PasswordLoginRequest, PasswordLoginResponse, MeResponse, OrgInfo types
- `internal/user/service.go` - Added bcrypt import, MEMBER check on Login, LoginPassword, GetMe, RegisterPassword methods
- `internal/user/handler.go` - Added strings import, routes for /auth/login/password and /auth/me, MEMBER 403 on Login handler
- `internal/user/repository.go` - Added UpdateUserPassword method
- `internal/common/model/user.go` - Added PasswordHash field to User struct

## Decisions Made

- bcrypt.DefaultCost (10) is used for password hashing — industry-standard, no configuration needed
- PasswordHash column is varchar(200) — bcrypt hashes are ~60 characters, this accommodates future cost increases
- GetMe handler uses authenticated route (authGroup) — requires valid JWT token to access
- LoginPassword handler uses unauthenticated route (rg.Group) — allows login without prior token
- Error string matching ("MEMBER_ROLE_FORBIDDEN") used as sentinel — same pattern as existing codebase error handling

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Missing closing brace in User model**
- **Found during:** Build verification
- **Issue:** Edit to add PasswordHash field accidentally removed struct closing brace, causing syntax error
- **Fix:** Added closing brace back
- **Files modified:** internal/common/model/user.go
- **Verification:** `go build ./cmd/server` passes
- **Committed in:** `a10f7bf`

---

**Total deviations:** 1 auto-fixed (blocking build error)
**Impact on plan:** Auto-fix immediately resolved build failure. No scope change.

## Issues Encountered

None other than the model brace syntax error (fixed immediately).

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Password login and /auth/me endpoints fully implemented and building
- MEMBER 403 enforcement in place for both SMS and password login
- Ready for frontend LoginView.vue to wire up password login UI
- Database migration will auto-add password_hash column on next GORM AutoMigrate

---
*Phase: 01-login-boss*
*Completed: 2026-04-11*
