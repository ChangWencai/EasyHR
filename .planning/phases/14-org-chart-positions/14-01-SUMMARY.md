---
phase: "14"
plan: "01"
subsystem: backend
tags: [position, department, employee, org-chart, crud]
dependency_graph:
  requires: [department-module, employee-module, common-model, middleware]
  provides: [position-crud-api, position-select-options, build-tree-v2, transfer-delete, cycle-detection]
  affects: [employee-model, department-service, department-handler, main-go]
tech_stack:
  added: []
  patterns: [Handler-Service-Repository, TenantScope, nullable-FK, BFS-cycle-detection, on-demand-migration]
key_files:
  created:
    - internal/position/model.go
    - internal/position/dto.go
    - internal/position/repository.go
    - internal/position/service.go
    - internal/position/handler.go
  modified:
    - internal/employee/model.go
    - internal/employee/repository.go
    - internal/department/model.go (errors moved)
    - internal/department/dto.go
    - internal/department/repository.go
    - internal/department/service.go
    - internal/department/handler.go
    - cmd/server/main.go
decisions:
  - Position 表独立建模（department_id 可 NULL 表示通用岗位）
  - 迁移逻辑改为 GetTree 按需触发（而非启动时全局迁移，因为启动时 orgID 不可知）
  - Department Service 注入 positionSvc 实现按需迁移
  - BuildTree v2 使用真实 Position 节点替代虚拟文本分组
  - markMatches 配色从 #1677FF 更新为 #4F6EF7（主色调），未匹配使用 opacity
metrics:
  duration: 22m
  completed: "2026-04-21"
---

# Phase 14 Plan 01: 后端 Position CRUD + Department 增强 Summary

Position 独立模块（CRUD + 下拉选项 + 数据迁移）+ Employee.position_id FK + Department BuildTree v2/循环检测/转移删除

## Commits

| # | Commit | Message |
|---|--------|---------|
| 1 | a90e981 | feat(14-01): 新建 Position 模型、DTO 和 Repository |
| 2 | cc6db4e | feat(14-01): 新建 Position Service 和 Handler |
| 3 | 6792a41 | feat(14-01): Employee.position_id + Department 服务增强 |
| 4 | 94cc5d6 | feat(14-01): 接入 Position 模块到 main.go |

## Files Changed

12 files changed, 758 insertions(+), 49 deletions(-)

### New Files (5)

| File | Purpose |
|------|---------|
| `internal/position/model.go` | Position struct (BaseModel + Name/DepartmentID/SortOrder) |
| `internal/position/dto.go` | Create/Update Request, Response, PositionSelectOptions, PositionOption |
| `internal/position/repository.go` | CRUD + ExistsByNameAndDept (NULL-safe) + ListByDepartment + CountByPositionID |
| `internal/position/service.go` | Business logic + GetSelectOptions + MigrateFromEmployeePositions + sentinel errors |
| `internal/position/handler.go` | 5 HTTP endpoints (POST/GET/PUT/DELETE + GET select-options) |

### Modified Files (7)

| File | Change |
|------|--------|
| `internal/employee/model.go` | Added PositionID *int64 field |
| `internal/employee/repository.go` | ListAllByOrg adds position_id + new UpdateDepartmentID method |
| `internal/department/dto.go` | Added TransferDeleteRequest struct |
| `internal/department/repository.go` | Added FindAllByIDs method, removed duplicate error vars |
| `internal/department/service.go` | BuildTree v2 + hasCycle + TransferAndDeleteDepartment + markMatches color update |
| `internal/department/handler.go` | Added TransferDeleteDepartment handler + route |
| `cmd/server/main.go` | Position AutoMigrate + DI + routes |

## API Endpoints

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| POST | /api/v1/positions | owner/admin | 创建岗位 |
| GET | /api/v1/positions | any | 岗位列表 |
| GET | /api/v1/positions/select-options | any | 岗位下拉选项（按部门分组） |
| PUT | /api/v1/positions/:id | owner/admin | 更新岗位 |
| DELETE | /api/v1/positions/:id | owner/admin | 删除岗位 |
| DELETE | /api/v1/departments/:id/transfer | owner/admin | 转移员工并删除部门 |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] 迁移策略从启动时改为按需触发**
- **Found during:** Task 4 (main.go wiring)
- **Issue:** 计划要求在 main.go 启动时调用 `MigrateFromEmployeePositions()`，但该方法需要 orgID 参数，而启动时无法遍历所有 org
- **Fix:** 改为在 Department Service 的 `GetTree` 方法中按需触发迁移（首次访问组织架构树时自动执行），Department Service 注入 positionSvc 支持此功能
- **Files modified:** `cmd/server/main.go`, `internal/department/service.go`
- **Commit:** 94cc5d6

**2. [Rule 3 - Blocking] Department Service 签名变更**
- **Found during:** Task 4 (DI wiring)
- **Issue:** 计划中 `department.NewService` 只接收 3 个参数（repo, empRepo, posRepo），但按需迁移需要 posSvc
- **Fix:** NewService 签名改为接收 4 个参数（+positionSvc），Service struct 增加 positionSvc 字段
- **Files modified:** `internal/department/service.go`, `cmd/server/main.go`
- **Commit:** 94cc5d6

**3. [Rule 1 - Bug] Department 错误变量重复定义**
- **Found during:** Task 3 (moving errors to service.go)
- **Issue:** ErrDepartmentNotFound/ErrHasChildren/ErrHasEmployees 在 repository.go 和 service.go 中重复定义
- **Fix:** 从 repository.go 移除错误定义，统一在 service.go 中定义（扩展了新增的错误变量）
- **Files modified:** `internal/department/repository.go`, `internal/department/service.go`
- **Commit:** 6792a41

## Key Decisions

1. **按需迁移** — Employee.position 文本到 Position 表的迁移在 GetTree 首次调用时按 orgID 触发，而非启动时全局执行
2. **通用岗位** — department_id=NULL 的岗位属于通用岗位，任何部门可见
3. **循环检测** — BFS 向上追溯父链，防止部门 parent_id 形成环
4. **搜索配色** — 匹配节点 #4F6EF7（主色调），未匹配节点 opacity 0.25

## Self-Check

- [x] `go build ./...` succeeds with zero errors
- [x] `internal/position/` package contains 5 files
- [x] Employee model has PositionID field
- [x] BuildTree v2 signature takes positions parameter
- [x] hasCycle method exists and is called before parent_id update
- [x] TransferAndDeleteDepartment method exists with registered route
- [x] MigrateFromEmployeePositions is called on-demand in GetTree
- [x] All 5 position endpoints registered
