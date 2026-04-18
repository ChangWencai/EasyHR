---
phase: 09-待办中心
plan: 02
subsystem: todo-list
tags: [todo, excel, excelize, el-table, search, filter, pin]

# Dependency graph
requires:
  - phase: 09-01
    provides: dashboard ring stats API, carousel API, HomeView carousel
provides:
  - TodoItem/CarouselItem/TodoInvite backend models
  - Todo CRUD API (list/search/filter/pin/export/carousels)
  - Frontend /todo page with full list management
affects: [todo-scheduler, invite-fill-page, HomeView-shortcuts]

# Tech tracking
tech-stack:
  added: []
  patterns: [todo-crud-pattern, excel-export-pattern, status-filter-pattern]

key-files:
  created:
    - internal/todo/model.go
    - internal/todo/repository.go
    - internal/todo/service.go
    - internal/todo/handler.go
    - internal/todo/router.go
    - internal/todo/excel.go
    - frontend/src/api/todo.ts
    - frontend/src/views/todo/TodoListView.vue
  modified:
    - cmd/server/main.go
    - frontend/src/router/index.ts

key-decisions:
  - "TodoItem uses model.BaseModel for soft-delete and tenant isolation"
  - "Export function named ExportTodosExcel to avoid collision with handler method"
  - "Router registered at v1.Group('') so routes are /api/v1/todos and /api/v1/carousels"

patterns-established:
  - "Todo CRUD: Repository -> Service -> Handler -> Router, same as dashboard pattern"
  - "Excel export: excelize with styled headers, status colors, CST timezone conversion"

requirements-completed: [TODO-01, TODO-02, TODO-03, TODO-06, TODO-07, TODO-08]

# Metrics
duration: 5min
completed: 2026-04-19
---

# Phase 09 Plan 02: 待办事项完整列表页 Summary

**TodoItem/CarouselItem/TodoInvite 后端模型 + CRUD API（搜索/筛选/置顶/导出）+ 前端 /todo 页面（el-table + 状态筛选 + 分页 + Excel 导出）**

## Performance

- **Duration:** 5 min
- **Started:** 2026-04-18T19:12:31Z
- **Completed:** 2026-04-19T04:18:00Z
- **Tasks:** 2
- **Files modified:** 9

## Accomplishments
- Backend TodoItem/CarouselItem/TodoInvite models with full field definitions
- CRUD API: list with pagination, keyword search, date range filter (60-day limit), status filter, pin toggle, Excel export, carousels
- Frontend /todo page with el-table, status/keyword/date search, pin toggle, Excel download, pagination (20/50/100)

## Task Commits

Each task was committed atomically:

1. **Task 1: Backend TodoItem/CarouselItem/TodoInvite models + Repository + Service + Handler + Router + Excel** - `baf1828` (feat)
2. **Task 2: Frontend todo list page TodoListView.vue + api/todo.ts** - `e87ba4b` (feat)

## Files Created/Modified
- `internal/todo/model.go` - TodoItem, CarouselItem, TodoInvite models with status/urgency constants
- `internal/todo/repository.go` - ListTodos, SearchTodos, FilterByDateRange, PinTodo, ListAllForExport, ListCarousels, CreateTodo, FindTodoByID, UpdateTodoStatus, ExistsBySource
- `internal/todo/service.go` - ListTodos with keyword/date/status routing, PinTodo, ExportTodos, CreateTodo with idempotency, ListCarousels, ComputeUrgencyStatus, GenerateInviteToken
- `internal/todo/handler.go` - ListTodos, PinTodo, ExportTodos, ListCarousels endpoints
- `internal/todo/router.go` - RegisterRouter with auth middleware, registers /todos, /todos/:id/pin, /todos/export, /carousels
- `internal/todo/excel.go` - ExportTodosExcel with styled headers, status colors, CST timezone, xlsx output
- `frontend/src/api/todo.ts` - listTodos, pinTodo, listCarousels, exportTodos API functions
- `frontend/src/views/todo/TodoListView.vue` - Full todo list page with search, status filter, date range, pin toggle, export, pagination
- `cmd/server/main.go` - Added todo import and RegisterRouter call
- `frontend/src/router/index.ts` - Added /todo route and auth guard

## Decisions Made
1. **Export function named ExportTodosExcel** - Avoids name collision between handler method ExportTodos and excel helper function
2. **Router at v1.Group("")** - Routes become /api/v1/todos and /api/v1/carousels (not nested under /todo)
3. **disabledDate blocks future dates only** - Plan's 60-day limit is enforced server-side; frontend simply disables future dates

## Deviations from Plan

None - plan executed exactly as written.

## Self-Check: PASSED

- All 8 created files verified present
- Both commits (baf1828, e87ba4b) verified in git log
- Backend build passes (go build ./internal/todo/...)

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- TodoItem model ready for Plan 03 (scheduler, urgency scan, invite flow)
- CarouselItem model ready for admin management UI
- TodoInvite model ready for token-based invite submission

---
*Phase: 09-待办中心*
*Completed: 2026-04-19*
