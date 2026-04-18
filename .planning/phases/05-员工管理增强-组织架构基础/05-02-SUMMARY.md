---
phase: 05
plan: 02
subsystem: department
tags: [backend, frontend, org-chart, echarts, department, tree]
dependency_graph:
  requires: [employee module, BaseModel, TenantScope, response helpers]
  provides: [Department CRUD API, Organization tree API, OrgChart.vue component]
  affects: [cmd/server/main.go, internal/employee/model.go, internal/employee/repository.go]
tech_stack:
  added: [echarts, vue-echarts]
  patterns: [adjacency-list department model, ECharts tree visualization, 3-layer tree (dept->position->employee)]
key_files:
  created:
    - internal/department/model.go
    - internal/department/dto.go
    - internal/department/repository.go
    - internal/department/service.go
    - internal/department/handler.go
    - frontend/src/api/department.ts
    - frontend/src/views/employee/OrgChart.vue
  modified:
    - internal/employee/model.go (added DepartmentID field)
    - internal/employee/repository.go (added ListAllByOrg, CountByDepartment)
    - cmd/server/main.go (registered Department module)
decisions:
  - Adjacency list model for department hierarchy (parent_id self-reference)
  - 3-layer tree structure: department -> position -> employee (position as virtual middle layer)
  - SearchTree uses backend-side matching with visual highlighting (ItemStyle/Label on TreeNode)
  - Frontend uses 300ms debounce for search input
  - Department deletion guarded by child and employee count checks
metrics:
  duration: 7m
  completed: "2026-04-18"
  tasks: 2
  files_created: 7
  files_modified: 3
---

# Phase 05 Plan 02: Department Module + Organization Chart Summary

新建 Department 模块（邻接表模型），Employee 新增 department_id 字段，后端提供部门 CRUD + 组织架构树 API，前端使用 ECharts tree 渲染可视化图表

## Tasks Completed

| Task | Name | Commit | Key Files |
|------|------|--------|-----------|
| 1 | Department backend module + Employee department_id | 8aa9873 | internal/department/{model,dto,repository,service,handler}.go, internal/employee/model.go, internal/employee/repository.go, cmd/server/main.go |
| 2 | Frontend OrgChart.vue visualization | d8b8d24 | frontend/src/api/department.ts, frontend/src/views/employee/OrgChart.vue, frontend/package.json |

## What Was Built

### Task 1: Backend Department Module
- **Department model**: Adjacency list with `parent_id` self-reference, `sort_order` for ordering
- **DTO layer**: CreateDepartmentRequest, UpdateDepartmentRequest, DepartmentResponse, TreeNode, DepartmentListQueryParams, SearchTreeRequest
- **Repository**: Full CRUD + ListAll + CountChildren + List (paginated), all with TenantScope
- **Service**: CRUD methods + GetTree (builds 3-layer dept->position->employee tree) + SearchTree (keyword matching with visual highlighting)
- **Handler**: REST endpoints with RequireRole for write operations, tree and search endpoints for all authenticated users
- **Employee changes**: Added `DepartmentID *int64` field, added `ListAllByOrg` and `CountByDepartment` repository methods
- **Threat mitigations**: T-05-03 (RequireRole owner/admin on writes), T-05-04 (delete guarded by child+employee checks)

### Task 2: Frontend OrgChart Component
- **department.ts API**: Full CRUD + getTree + searchTree
- **OrgChart.vue**: ECharts tree visualization with orthogonal LR layout, search with 300ms debounce, create department dialog
- **Visual features**: Search highlighting (blue for matches, gray for non-matches), empty/error states, responsive layout
- **ECharts config**: roam enabled, animation disabled, initialTreeDepth -1 (fully expanded)

## Deviations from Plan

None - plan executed exactly as written.

## Verification Results

- `go build ./...` exits 0
- `npx vue-tsc --noEmit` passes (no new type errors)
- All 13 acceptance criteria PASS

## Self-Check: PASSED

All files exist, all commits verified.
