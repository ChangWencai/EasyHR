---
phase: "09-待办中心"
plan: "03"
subsystem: "todo-scheduler-invite-carousel"
tags: [todo, scheduler, gocron, invite, carousel, upload, invite-fill, carousel-manage]

# Dependency graph
requires:
  - phase: 09-02
    provides: TodoItem/CarouselItem/TodoInvite models, CRUD API, TodoListView
provides:
  - gocron scheduler (6 jobs: urgency scan, carousel activation, monthly, SI change, annual base, contract renewal)
  - Invite/Terminate/SubmitInvite endpoints
  - Carousel CRUD endpoints (admin list, create, update, delete)
  - Image upload endpoint
  - EmployeeService contract_new todo trigger
  - ContractService contract_renew check
  - Frontend InviteFillPage, CarouselManagePage
affects: [employee-service, contract-service, main-go]

# Tech tracking
tech-stack:
  added: [gocron-v2, google-uuid]
  patterns: [interface-injection-to-avoid-circular-import, ContractServiceWrapper]

key-files:
  created:
    - internal/todo/scheduler.go
    - internal/upload/handler.go
    - internal/upload/router.go
    - frontend/src/views/todo/InviteFillPage.vue
    - frontend/src/views/todo/CarouselManagePage.vue
  modified:
    - internal/todo/repository.go
    - internal/todo/service.go
    - internal/todo/handler.go
    - internal/todo/router.go
    - internal/employee/service.go
    - internal/employee/contract_service.go
    - cmd/server/main.go
    - frontend/src/api/todo.ts
    - frontend/src/api/carousel.ts
    - frontend/src/views/todo/TodoListView.vue
    - frontend/src/router/index.ts

key-decisions:
  - "TodoCreator interface in employee package avoids circular import from todo package"
  - "ContractServiceWrapper interface in todo/scheduler.go avoids circular import from employee package"
  - "todo.Service.CreateTodoFromEmployee implements employee.TodoCreator interface"
  - "Upload stores to local filesystem with UUID filename; production should use OSS"
  - "Router split into public routes (invite verify/submit) and auth-protected routes"

requirements-completed: [TODO-04, TODO-05, TODO-11, TODO-12, TODO-13, TODO-14, TODO-15, TODO-16, TODO-17, TODO-18]

# Metrics
duration: 16min
completed: 2026-04-19
---

# Phase 09 Plan 03: 协办邀请 + 终止任务 + gocron 调度器 + 轮播图管理 Summary

**gocron 定时调度器（6个 job） + 协办邀请（Token链接验证） + 终止任务 + 轮播图 CRUD + 图片上传 + 员工入职/合同续签待办触发 + 前端协办填写页和轮播图管理页**

## Performance

- **Duration:** 16 min
- **Started:** 2026-04-18T19:20:20Z
- **Completed:** 2026-04-19
- **Tasks:** 2
- **Files modified:** 16 (8 created + 8 modified)

## Accomplishments
- gocron scheduler with 6 scheduled jobs (urgency scan, carousel activation, monthly todos, SI change, annual base, contract renewal)
- InviteTodo/TerminateTodo/SubmitInvite/VerifyInviteToken endpoints with token-based auth
- Carousel CRUD endpoints for admin management (max 3 images enforced server-side)
- Image upload endpoint with 5MB limit and file type whitelist
- EmployeeService.CreateEmployee triggers contract_new todo with 30-day deadline
- ContractService.CheckContractRenewalReminders scans active contracts expiring in 30 days
- Frontend InviteFillPage with token verification (404/410 error states)
- Frontend CarouselManagePage with upload/edit/delete/toggle functionality
- TodoListView action dropdown with invite/terminate operations

## Task Commits

Each task was committed atomically:

1. **Task 1: Backend invite/terminate/scheduler/carousel CRUD** - `f671099` (feat)
2. **Task 2: Employee triggers + frontend invite/carousel pages** - `a4b0853` (feat)

## Files Created/Modified
- `internal/todo/scheduler.go` - gocron scheduler with 6 jobs, redis distributed lock, ContractServiceWrapper interface
- `internal/todo/repository.go` - ScanUrgencyStatus (D-09-04 rules), UpdateCarouselActivation, GenerateMonthlyTodos, GenerateAnnualBaseTodos, carousel CRUD, invite methods
- `internal/todo/service.go` - InviteTodo, TerminateTodo, VerifyInviteToken, SubmitInvite, carousel CRUD service, CreateTodoFromEmployee (implements TodoCreator)
- `internal/todo/handler.go` - InviteTodo, TerminateTodo, VerifyInviteToken, GetInviteTodo, SubmitInvite, carousel CRUD handlers
- `internal/todo/router.go` - Split public/auth routes, register invite/terminate/carousel management endpoints
- `internal/upload/handler.go` - UploadImage with 5MB limit, file type whitelist, UUID filename
- `internal/upload/router.go` - POST /upload/image registration
- `internal/employee/service.go` - TodoCreator interface, todoSvc field, CreateEmployee triggers contract_new todo
- `internal/employee/contract_service.go` - todoSvc field, CheckContractRenewalReminders, FindContractsExpiringSoon
- `cmd/server/main.go` - todoSvc DI to employee/contract, todo scheduler startup with contractSvc, upload router, AutoMigrate
- `frontend/src/api/todo.ts` - inviteTodo, terminateTodo, verifyInviteToken functions
- `frontend/src/api/carousel.ts` - listAllCarousels, createCarousel, updateCarousel, deleteCarousel, uploadImage
- `frontend/src/views/todo/TodoListView.vue` - Action dropdown with invite/terminate, handleAction function
- `frontend/src/views/todo/InviteFillPage.vue` - Token verification + form submission page (public)
- `frontend/src/views/todo/CarouselManagePage.vue` - Full carousel management with upload/edit/delete/toggle
- `frontend/src/router/index.ts` - /todo/:id/invite (public), /carousel/manage (protected) routes

## Decisions Made
1. **TodoCreator interface in employee package** - Avoids circular import by defining CreateTodoFromEmployee interface in employee package, implemented by todo.Service
2. **ContractServiceWrapper in scheduler.go** - Same pattern for scheduler to call contract renewal check without importing employee package
3. **Router split public/auth** - Invite verify and submit routes are public (no auth), while terminate/invite-creation/carousel CRUD require auth
4. **Upload to local filesystem** - Current implementation stores files locally with UUID names; production should use Aliyun OSS signed URLs

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed OrgID struct literal compilation errors**
- **Found during:** Task 1 (build verification)
- **Issue:** Plan used `OrgID: orgID` in struct literals, but TodoItem/TodoInvite embed model.BaseModel where OrgID is promoted
- **Fix:** Changed to `item.OrgID = orgID` assignment after struct creation
- **Files modified:** internal/todo/repository.go, internal/todo/scheduler.go, internal/todo/service.go
- **Commit:** f671099

**2. [Rule 1 - Bug] Fixed GORM Update return type mismatch**
- **Found during:** Task 1 (build verification)
- **Issue:** Plan used `result, err := r.db.Model().Update()` but GORM Update returns single `*gorm.DB`, not `(Result, error)`
- **Fix:** Changed to `result := r.db.Model().Update()` with `result.Error` check
- **Files modified:** internal/todo/repository.go
- **Commit:** f671099

**3. [Rule 3 - Blocking] Used interface injection to avoid circular imports**
- **Found during:** Task 2 (design phase)
- **Issue:** Plan suggested employee.Service hold `*todo.Repository` directly, which creates circular import
- **Fix:** Created `TodoCreator` interface in employee package with `CreateTodoFromEmployee` method; `todo.Service` implements it
- **Files modified:** internal/employee/service.go, internal/todo/service.go
- **Commit:** a4b0853

## Self-Check: PASSED

- Both commits (f671099, a4b0853) verified in git log
- Backend builds successfully (go build ./...)
- No TypeScript errors in new frontend files (vue-tsc --noEmit)
- All grep verifications pass for required symbols
- InviteFillPage.vue, CarouselManagePage.vue created
- Router correctly registers public and protected routes

## Next Phase Readiness
- All 10 requirements (TODO-04 through TODO-18) completed
- Phase 09 fully implemented (all 3 plans done)
- Todo scheduler runs 6 jobs on configurable schedules
- Employee/contract modules integrated with todo creation triggers

---
*Phase: 09-待办中心*
*Completed: 2026-04-19*
