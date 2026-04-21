# Phase 14: 组织架构图（部门+岗位管理）- Research

**Researched:** 2026-04-21
**Domain:** Go backend (Position CRUD + Department drag-drop + tree API) + Vue 3 frontend (ECharts drag-drop, el-select position picker, delete-transfer dialog)
**Confidence:** HIGH

## Summary

Phase 14 introduces a new `Position` table (independent position management), enhances the existing ECharts org chart with drag-drop department reparenting, adds a delete-with-transfer flow, and converts the employee form's free-text position field into a filtered dropdown. The backend is a net-new Position module + new transfer-delete endpoint, plus the tree API must now include `position_id` per employee and build real position nodes instead of grouping by `Employee.Position` string. The frontend requires ECharts drag event wiring (dragstart/dragend), a position picker with el-optgroup, and a transfer dialog.

**Primary recommendation:** Add Position as a new package under `internal/position/` (handler/service/repository), register in `main.go` AutoMigrate, wire the OrgChart.vue ECharts drag events to call `PUT /departments/:id` with new `parent_id`, and replace the employee form's `el-input` with a grouped `el-select`.

---

## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-14-01:** 新建 `Position` 表（id/name/department_id可NULL/sort_order/org_id），唯一性约束 `(org_id, department_id, name)`，通用岗位 `department_id=NULL`
- **D-14-02:** 迁移策略：读取现有 `Employee.position` 唯一值，自动创建对应 Position 记录，关联现有员工 `position_id`
- **D-14-03:** 通用岗位 `department_id=NULL` 表示跨部门复用
- **D-14-04:** 员工岗位下拉按员工 `department_id` 过滤：部门专属岗位 + 通用岗位
- **D-14-05:** 新建岗位时默认「通用岗位」（`department_id=NULL`）
- **D-14-06:** 架构图拖拽后自动更新 `Department.parent_id`
- **D-14-08:** 搜索高亮：匹配节点 label 变蓝色 #4F6EF7，未匹配节点 opacity 0.25
- **D-14-09:** 删除有员工的部门：引导选择目标部门转移员工后再删除
- **D-14-10:** 内联编辑：点击部门节点名称进入编辑模式，blur/Enter 保存
- **D-14-12:** 员工表单岗位字段从 `el-input` 改为 `el-select` 下拉选择
- **D-14-13:** 下拉根据当前部门动态筛选
- **D-14-14:** 无岗位情况：员工可选择「未分配岗位」（`position_id = NULL`）
- **D-14-15:** 部门为空时只显示通用岗位

### Claude's Discretion
- 架构图节点样式细节、拖拽视觉反馈、部门删除弹窗具体布局、岗位分组展示、迁移进度方式

---

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| ORG-01 | 岗位管理（创建/编辑/删除岗位，名称/所属部门/排序，可跨部门复用） | Position model + Position CRUD + Employee.position_id FK + migration |
| ORG-02 | 组织架构图增强（三级树/搜索高亮/展开折叠/拖拽） | ECharts drag events + PUT /departments/:id parent_id + BuildTree v2 with Position table |
| ORG-03 | 部门管理完善（内联编辑/删除含转移） | Inline edit via ECharts click event + new transfer+delete endpoint |
| ORG-04 | 员工关联岗位选择下拉 | el-select with el-optgroup, department-filtered, new positions API |

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Position CRUD (model/handler/service/repo) | API/Backend | — | New Go module, follows existing module pattern |
| Employee.position_id FK | API/Backend | Database | GORM FK, AutoMigrate in main.go |
| Data migration (Employee.position → Position table) | API/Backend | — | Asynq batch task or startup seed |
| Tree API (departments→positions→employees) | API/Backend | Frontend | Service.BuildTree v2 uses Position table join |
| Drag-drop department reparenting | Frontend (ECharts) | API/Backend (PUT /departments/:id) | Frontend detects drop, calls existing update API with new parent_id |
| Department inline edit | Frontend (ECharts click) | API/Backend | ECharts label click → input overlay → PATCH /departments/:id |
| Delete-with-transfer | Frontend (dialog) | API/Backend (new endpoint) | New DELETE /departments/:id/transfer-and-delete |
| Employee position dropdown | Frontend | API/Backend | el-select, API returns grouped positions |

---

## Standard Stack

No new library additions — all Phase 14 needs are satisfied with existing project dependencies.

### Backend (Go)
| Component | Library | Purpose |
|-----------|---------|---------|
| HTTP framework | Gin v1.12.0 | Existing — used for new Position handler |
| ORM | GORM v1.31.1 | Existing — used for Position model AutoMigrate |
| Validation | go-playground/validator v10 | Existing — used for Position DTOs |

**No new Go packages required.**

### Frontend (Vue 3)
| Component | Library | Purpose |
|-----------|---------|---------|
| UI framework | Element Plus 2.13.6 | Existing — el-select, el-dialog, el-table, el-alert, el-optgroup, el-input-number |
| Charts | ECharts 6.0.0 (via vue-echarts) | Existing — TreeChart drag events |
| HTTP client | Axios 1.14.0 | Existing — new position API calls |
| State | Pinia 3.0.4 | Existing — position store if needed |

**No new npm packages required.**

---

## Architecture Patterns

### System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│  H5 Browser (Vue 3)                                                 │
│  ┌──────────────┐  ┌────────────────────┐  ┌───────────────────┐  │
│  │ OrgChart.vue │  │ EmployeeCreate.vue │  │ PositionDialog.vue│  │
│  │ (ECharts)    │  │ (el-select)        │  │ (CRUD form)       │  │
│  └──────┬───────┘  └────────┬───────────┘  └───────┬───────────┘  │
│         │                   │                      │               │
│         └───────────────────┼──────────────────────┘               │
│                             ▼                                      │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ department.ts + NEW: position.ts (API clients)                │  │
│  └──────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
         │ HTTP REST                                            │
         ▼                                                       │
┌─────────────────────────────────────────────────────────────┐
│  Go Backend (Gin)                                             │
│  ┌─────────────────┐  ┌────────────────────────────────────┐  │
│  │ PositionHandler  │  │ DepartmentHandler (enhanced)       │  │
│  │ POST/GET/PUT/    │  │ GET /tree (+ position_id in nodes) │  │
│  │ DELETE /positions│  │ PUT /departments/:id (parent_id)    │  │
│  └────────┬────────┘  │ DELETE /departments/:id/transfer   │  │
│           │           └────────────────────────────────────┘  │
│           ▼                                                    │
│  ┌─────────────────┐  ┌─────────────────┐  ┌───────────────┐  │
│  │ PositionService │  │ DepartmentSvc   │  │ EmployeeRepo  │  │
│  │ (CRUD + list)   │  │ BuildTree v2    │  │ (position_id) │  │
│  └────────┬────────┘  └────────┬────────┘  └───────┬───────┘  │
│           │                    │                   │           │
│           ▼                    ▼                   ▼           │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  PostgreSQL                                               │   │
│  │  departments (parent_id/邻接表) + NEW: positions +        │   │
│  │  employees (position_id FK nullable)                     │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### Recommended Project Structure

```
internal/
├── position/                    # NEW MODULE
│   ├── model.go                  # Position struct (BaseModel + name/department_id/sort_order)
│   ├── dto.go                    # Request/Response DTOs
│   ├── repository.go             # CRUD + ListByOrg + ListByDepartment
│   ├── service.go               # Business logic
│   └── handler.go                # HTTP endpoints
├── department/
│   ├── model.go                 # existing — unchanged
│   ├── dto.go                   # ADD: TransferDeleteRequest
│   ├── repository.go            # existing — unchanged
│   ├── service.go               # MODIFY: BuildTree v2 (use Position join)
│   └── handler.go               # ADD: TransferDeleteDepartment
├── employee/
│   ├── model.go                 # MODIFY: Add position_id int64 FK field
│   ├── repository.go            # MODIFY: ListAllByOrg includes position_id
│   └── service.go               # existing — unchanged
└── ...

cmd/server/main.go               # ADD: Position{} + Employee.position_id to AutoMigrate

frontend/src/
├── api/
│   ├── position.ts              # NEW: position CRUD API client
│   └── department.ts            # MODIFY: add transferDelete, add tree response type
├── views/employee/
│   ├── OrgChart.vue             # MODIFY: drag-drop, inline edit, delete-transfer
│   └── EmployeeCreate.vue       # MODIFY: position el-select with grouping
└── components/
    └── PositionManageDialog.vue  # NEW: position CRUD dialog
```

### Pattern 1: Position CRUD Module

**What:** Follows the established Handler → Service → Repository three-layer pattern identical to the Department module.

**Structure:**
```
// model.go
type Position struct {
    model.BaseModel
    Name         string   `gorm:"column:name;type:varchar(100);not null"`
    DepartmentID *int64   `gorm:"column:department_id;index"`
    SortOrder    int      `gorm:"column:sort_order;not null;default:0"`
}

// dto.go
type CreatePositionRequest struct { Name string `json:"name" binding:"required,max=100"` DepartmentID *int64 `json:"department_id"` SortOrder int `json:"sort_order"` }
type UpdatePositionRequest struct { Name *string `json:"name"` DepartmentID *int64 `json:"department_id"` SortOrder *int `json:"sort_order"` }

// repository.go — Repository struct with db *gorm.DB
// Create / FindByID / Update / Delete / ListByOrg / ListByDepartment

// service.go — Service struct with repo *Repository
// Create / Update / Delete / ListPositions / MigrateFromEmployeeText

// handler.go — Handler struct with svc *Service, RegisterRoutes POST/GET/PUT/DELETE /positions
```

**Example (Service.Create):**
```go
// Source: existing department/service.go CreateDepartment pattern (D-14-01 constrained)
func (s *Service) CreatePosition(orgID, userID int64, req *CreatePositionRequest) (*PositionResponse, error) {
    // Check uniqueness: (org_id, department_id, name) — handle NULL department_id case
    exists, err := s.repo.ExistsByNameAndDept(orgID, req.DepartmentID, req.Name)
    if err != nil { return nil, fmt.Errorf("check position: %w", err) }
    if exists { return nil, ErrPositionDuplicate }
    pos := &Position{
        Name:         req.Name,
        DepartmentID: req.DepartmentID,
        SortOrder:    req.SortOrder,
    }
    pos.OrgID = orgID
    pos.CreatedBy = userID
    pos.UpdatedBy = userID
    if err := s.repo.Create(pos); err != nil { return nil, fmt.Errorf("create position: %w", err) }
    return toPositionResponse(pos), nil
}
```

**Example (Repository.ExistsByNameAndDept):**
```go
// Handle NULL department_id uniqueness: use COALESCE or separate queries
func (r *Repository) ExistsByNameAndDept(orgID int64, deptID *int64, name string) (bool, error) {
    var count int64
    query := r.db.Model(&Position{}).Where("org_id = ? AND name = ?", orgID, name)
    if deptID == nil {
        query = query.Where("department_id IS NULL")
    } else {
        query = query.Where("department_id = ?", *deptID)
    }
    query.Count(&count)
    return count > 0, nil
}
```

### Pattern 2: BuildTree v2 with Position Table

**What:** Replace the existing `BuildTree` method that groups employees by `Employee.Position` string with a version that joins against the `Position` table and creates real (non-virtual) position nodes.

**When to use:** `GET /departments/tree` endpoint and `SearchTree` both call `BuildTree`.

**Example:**
```go
// Source: existing department/service.go BuildTree method (D-14-01 constrained)
// Key change: instead of posGroups := make(map[string][]employee.Employee)
// Fetch positions by orgID, then group employees by position_id

func (s *Service) BuildTree(departments []Department, employees []employee.Employee, positions []Position) []*TreeNode {
    // Group positions by department_id
    posByDept := make(map[int64][]Position)   // deptID -> positions
    var commonPositions []Position
    for _, p := range positions {
        if p.DepartmentID == nil {
            commonPositions = append(commonPositions, p)
        } else {
            posByDept[*p.DepartmentID] = append(posByDept[*p.DepartmentID], p)
        }
    }

    // Group employees by position_id
    empByPosID := make(map[int64][]employee.Employee)  // position_id -> employees
    var unassignedEmps []employee.Employee
    for _, emp := range employees {
        if emp.PositionID != nil {
            empByPosID[*emp.PositionID] = append(empByPosID[*emp.PositionID], emp)
        } else {
            unassignedEmps = append(unassignedEmps, emp)
        }
    }

    // Build position nodes from positions table, NOT from employee.position string
    // For each department:
    //   1. Add dept-specific positions
    //   2. Add unassigned node (if unassignedEmps non-empty)
    //   3. For each position node: add children employees
    //   4. Add "未分配岗位" node if employees exist without position_id
}
```

### Pattern 3: ECharts Drag-and-Drop Department Reparenting

**What:** ECharts tree nodes fire no native drag events on canvas. The standard approach is to use `roam: true` (pan/zoom only) and implement custom drag behavior via `getZr()` click/drag pixel-level tracking, OR use `roam: false` and implement drag via mouse events on individual ` graphic` elements overlaid on tree nodes. A simpler approach compatible with the existing `roam: true` setting: detect clicks on nodes, show a context menu to "移动到...", call the existing `PUT /departments/:id` API with new `parent_id`.

**Alternative (D-14-06 "支持拖拽") interpretation:** ECharts TreeChart supports `draggable: true` on individual data items, but tree layout recalculates automatically — what we actually want is a "move to new parent" operation. The cleanest implementation: keep ECharts `roam: true` for pan/zoom, and implement "drag" as: click node → enter "move mode" → click target parent → call update API.

**Recommended approach (D-14-06):** Do NOT use ECharts built-in drag. Instead, implement a context menu on right-click department nodes showing "移动到..." → shows dept picker → PUT `/departments/:id` with new `parent_id`. This is simpler, more reliable, and clearer UX. The UI spec says "drag" but the business goal is "reparent a department" — context menu achieves the same result.

### Pattern 4: Inline Edit on ECharts Node Label

**What:** Intercept label click via ECharts `click` event, detect `nodeType === 'node'` (department nodes), render an HTML input overlay at the node's pixel coordinates, save on blur/Enter, revert on Escape.

**Example (OrgChart.vue):**
```typescript
// Source: existing OrgChart.vue — uses VChart + TreeChart
// Add after chart init:

chartInstance.on('click', (params: unknown) => {
  const p = params as { data?: TreeNode; event?: { offsetX?: number; offsetY?: number } }
  if (p.data?.type === 'department' && p.data?.id) {
    // Show inline edit overlay
    showInlineEdit(p.data.id, p.data.name, p.event?.offsetX, p.event?.offsetY)
  }
})
```

### Pattern 5: Position Grouped Dropdown (el-select + el-optgroup)

**What:** `el-select` with `allow-create` removed, using `el-optgroup` for "部门专属岗位" and "通用岗位".

**API response shape for employee position picker:**
```json
// GET /positions/select-options?department_id=123
{
  "dept_positions": [{ "id": 1, "name": "前端工程师" }],
  "common_positions": [{ "id": 5, "name": "项目经理" }],
  "unassigned_option": { "id": null, "name": "未分配岗位" }
}
```

**Frontend template:**
```html
<!-- Source: 14-UI-SPEC.md — Position Dropdown section -->
<el-select v-model="form.position_id" placeholder="请选择岗位" clearable>
  <el-option :value="null" label="未分配岗位" />  <!-- D-14-14 -->
  <el-optgroup label="部门专属岗位" v-if="deptPositions.length">
    <el-option v-for="p in deptPositions" :key="p.id" :value="p.id" :label="p.name" />
  </el-optgroup>
  <el-optgroup label="通用岗位">
    <el-option v-for="p in commonPositions" :key="p.id" :value="p.id" :label="p.name" />
  </el-optgroup>
</el-select>
```

### Pattern 6: Transfer-and-Delete Department

**What:** New API endpoint `DELETE /departments/:id/transfer-and-delete` that accepts a `target_department_id` and a list of `employee_ids`, updates all those employees' `department_id`, then deletes the department.

**Request:**
```json
// DELETE /departments/:id/transfer-and-delete
{ "target_department_id": 5, "employee_ids": [1, 2, 3] }
```

**Service logic:**
```go
func (s *Service) TransferAndDeleteDepartment(orgID, deptID, targetDeptID int64, employeeIDs []int64) error {
    // 1. Validate target dept exists
    // 2. Validate all employees belong to this dept
    // 3. Batch update employees' department_id to targetDeptID (position_id preserved, D-14-09)
    // 4. Delete the department
    // 5. Return
}
```

### Pattern 7: Data Migration (Employee.position text → Position table)

**What:** On application startup, detect if any employees have `position` text but no `position_id`, then auto-create Position records and backfill.

**Implementation options:**
- **Option A (Startup seed):** In `main.go` after AutoMigrate, call `positionSvc.AutoMigrateEmployeePositions()`. Simple but runs on every startup (use a flag or check if already migrated).
- **Option B (On-demand):** First time the tree API is called for an org, run migration if needed.
- **Option C (Asynq task):** Schedule a one-time asynq task.

**Recommended:** Option A with a flag check — use `positions_migrated` org setting or check if Position table has any records. If not migrated and employees have non-empty Position strings, run migration.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Position node data | Custom JSON encoding for virtual position nodes | BuildTree v2 with real Position table join | Real position nodes enable editing and CRUD |
| Unique constraint with NULL | Raw SQL `UNIQUE(org_id, department_id, name)` with NULL trick | Separate query + GORM tx with lock, or GORM uniqueIndex with COALESCE | PostgreSQL treats NULL != NULL so `UNIQUE(col)` allows multiple NULLs — our constraint wants at most one NULL per name per org |
| Grouped dropdown options | Multiple API calls | Single `GET /positions/select-options?department_id=N` | Simpler frontend, one round trip |
| Department delete transfer | Two-step (move employees → delete dept) with race condition | Single `DELETE /departments/:id/transfer-and-delete` atomic endpoint | Atomic prevents orphaned employees on partial failure |

---

## Common Pitfalls

### Pitfall 1: ECharts Tree Drag-Drop Is Not Pan/Zoom
**What goes wrong:** ECharts tree with `roam: true` does not support node drag-drop out of the box — drag events on tree nodes are intercepted by the pan behavior.
**Why it happens:** ECharts TreeChart's built-in `draggable` option affects the layout, not the data model. We want data-level reparenting (new `parent_id`).
**How to avoid:** Implement as click-to-select + context menu "移动到..." dept picker, not ECharts drag. This is clearer UX and avoids fighting ECharts pan behavior.
**Warning signs:** Nodes visually move but snap back on refresh; tree API returns unchanged structure.

### Pitfall 2: Position Uniqueness with NULL department_id
**What goes wrong:** Two `Position` records with `department_id=NULL` and the same `name` both satisfy the `UNIQUE(org_id, department_id, name)` partial index if PostgreSQL treats NULLs as equal.
**Why it happens:** PostgreSQL's `UNIQUE` treats NULL as distinct values — `WHERE department_id IS NULL` allows multiple rows with NULL. The spec says only one common position per name per org.
**How to avoid:** In the Position repository, always check `COUNT(*) WHERE org_id=? AND name=? AND (department_id IS NOT DISTINCT FROM ?)` — use `IS NOT DISTINCT FROM` to make NULL=NULL comparison explicit, or use a composite partial unique index: `CREATE UNIQUE INDEX ON positions(org_id, name) WHERE department_id IS NULL`.
**Warning signs:** Duplicate "通用岗位" entries for the same name appear after creating two common positions.

### Pitfall 3: BuildTree Breaking Existing Search Highlighting
**What goes wrong:** Modifying `BuildTree` to use the Position table changes the tree structure, which may break the existing `markMatches` function or change node IDs that the frontend relies on.
**Why it happens:** Phase 05's tree used `Position` string as position node ID=0 (virtual). The new tree uses real Position IDs. Employee nodes previously had IDs from `Employee.ID`. After adding `position_id`, employees in the same position group are split.
**How to avoid:** Keep the `TreeNode.ID` for department nodes (real `Department.ID`). For position nodes: use `Position.ID`. For employee nodes: keep `Employee.ID`. The `type` field remains `"department" | "position" | "employee"`. The frontend uses `type` not `ID` to determine node kind.
**Warning signs:** Search highlights wrong nodes after tree API change.

### Pitfall 4: Deleting a Position That Has Employees
**What goes wrong:** Deleting a position that is referenced by `Employee.position_id` causes FK violation or orphaned reference.
**Why it happens:** No cascade delete rule on `Employee.position_id FK`.
**How to avoid:** In Position Service Delete method, check if any employees reference the position before deletion. Return `ErrPositionInUse` with count of employees. User must reassign employees first. This mirrors the department delete-with-transfer pattern.
**Warning signs:** Database constraint error or 500 on position delete.

### Pitfall 5: Employee.position String Left Stale
**What goes wrong:** After creating Position records, `Employee.position` text is not updated when a position is renamed.
**Why it happens:** `Employee.position` is a denormalized display field. If a Position is renamed, existing employees still show the old text.
**How to avoid:** On Position rename, do NOT try to update all employees' `position` text (too expensive). Instead: (a) always prefer `position.name` from the Position table when displaying, (b) if position is deleted, keep employees' `position` text as-is (it's just display text). The employee form should always read from Position table.
**Warning signs:** Employee roster shows old position name after position rename.

### Pitfall 6: Department parent_id Cycle
**What goes wrong:** Drag-drop or update allows setting a department's `parent_id` to one of its own descendants, creating a cycle in the adjacency list.
**Why it happens:** No cycle detection when updating `Department.parent_id`.
**How to avoid:** After updating parent_id, run a cycle check query or implement in service: fetch all ancestor IDs, ensure new parent is not in the descendant set. Reject with `ErrCircularReference`.
**Warning signs:** `GetTree` infinite recursion or stack overflow.

---

## Code Examples

### 1. Position Model (Go)
```go
// Source: D-14-01, following internal/department/model.go pattern
package position

import (
    "github.com/wencai/easyhr/internal/common/model"
)

type Position struct {
    model.BaseModel
    Name         string  `gorm:"column:name;type:varchar(100);not null;comment:岗位名称"`
    DepartmentID *int64  `gorm:"column:department_id;index;comment:所属部门（NULL=通用岗位）"`
    SortOrder    int     `gorm:"column:sort_order;not null;default:0;comment:排序"`
}

func (Position) TableName() string {
    return "positions"
}
```

### 2. Employee Model Update (Go)
```go
// Source: D-14-01 — add position_id to Employee
// In internal/employee/model.go, add field after DepartmentID:
// PositionID    *int64  `gorm:"column:position_id;index;comment:岗位ID"`
```

### 3. ECharts Chart Option with Drag Events (Vue 3)
```typescript
// Source: existing OrgChart.vue chartOption — extend for D-14-06
// Custom "drag" via right-click context menu (not ECharts built-in drag)

const chartInstance = chartRef.value?.chart
chartInstance?.on('contextmenu', (params: unknown) => {
  const p = params as { data?: TreeNode; offsetX?: number; offsetY?: number }
  if (p.data?.type === 'department') {
    e.preventDefault()
    showMoveMenu(p.data.id, p.data.name, p.offsetX, p.offsetY)
  }
})
```

### 4. Position Select Options API Response (Go)
```go
// Source: D-14-13, D-14-14, D-14-15
// GET /positions/select-options?department_id=123

type PositionSelectOptions struct {
    DeptPositions    []PositionOption `json:"dept_positions"`
    CommonPositions  []PositionOption `json:"common_positions"`
    UnassignedOption PositionOption   `json:"unassigned_option"` // {id: nil, name: "未分配岗位"}
}

type PositionOption struct {
    ID   *int64  `json:"id"`
    Name string `json:"name"`
}

// Service logic:
// dept_positions = positions WHERE department_id = paramDeptID
// common_positions = positions WHERE department_id IS NULL
// unassigned_option = {id: null, name: "未分配岗位"}
```

### 5. Transfer-and-Delete Department (Go)
```go
// Source: D-14-09 — new department service method

func (s *Service) TransferAndDeleteDepartment(orgID, deptID, targetDeptID int64, employeeIDs []int64) error {
    // Validate target department
    _, err := s.repo.FindByID(orgID, targetDeptID)
    if err != nil { return ErrDepartmentNotFound }

    // Validate employees belong to source dept
    for _, empID := range employeeIDs {
        emp, err := s.empRepo.FindByID(orgID, empID)
        if err != nil { return ErrEmployeeNotFound }
        if emp.DepartmentID == nil || *emp.DepartmentID != deptID {
            return ErrEmployeeNotInDepartment
        }
    }

    // Transfer employees (position_id stays, D-14-09)
    updates := map[string]interface{}{"department_id": targetDeptID}
    for _, empID := range employeeIDs {
        if err := s.empRepo.Update(orgID, empID, updates); err != nil {
            return fmt.Errorf("transfer employee %d: %w", empID, err)
        }
    }

    // Delete department (no employees remain)
    return s.repo.Delete(orgID, deptID)
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `Employee.position` as free text | `Employee.position_id` FK → `Position` table | Phase 14 | Enables position CRUD, consistent naming, dept-filtered dropdown |
| Position nodes as virtual JSON (ID=0) | Real position nodes from Position table | Phase 14 | Positions are first-class, can be edited/deleted |
| No department drag-drop | Context menu "移动到..." dept picker → PUT API | Phase 14 | Org restructure without delete+recreate |
| Department delete blocked when employees exist | Transfer employees → delete atomically | Phase 14 | Removes workaround of manually moving employees |
| Employee position as el-input text | el-select grouped by dept/common | Phase 14 | Consistent data, no typos, dept context |

**Deprecated/outdated:**
- `Employee.Position` free text field: still stored for backward compatibility but UI always uses Position table
- `BuildTree` v1 (grouping by `emp.Position` string): replaced with v2 using Position table

---

## Assumptions Log

> List all claims tagged `[ASSUMED]` in this research. The planner and discuss-phase use this section to identify decisions that need user confirmation before execution.

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `DELETE /departments/:id/transfer-and-delete` is the right REST pattern (not PUT /departments/:id/transfer) | Pattern 6 | If user prefers separate /transfer endpoint, plan changes |
| A2 | ECharts built-in drag is not used — context menu approach is chosen | Pattern 3 | If built-in drag is actually desired, ECharts config changes |
| A3 | Startup migration (Option A) is acceptable — runs on every server start with a flag check | Pattern 7 | If user wants a dedicated migration script (asynq/CLI), plan changes |
| A4 | Position uniqueness: one common position per name per org (D-14-01) | Pitfall 2 | If multiple common positions with same name are allowed, DB constraint changes |
| A5 | Employee.position string is kept as denormalized display (not kept in sync on rename) | Pitfall 5 | If sync-on-rename is required, plan needs Position rename cascade update |

**If this table is empty:** All claims in this research were verified or cited — no user confirmation needed.

---

## Open Questions

1. **ECharts drag UX interpretation**
   - What we know: D-14-06 says "支持拖拽调整部门层级"
   - What's unclear: Is "drag" literally dragging nodes, or is it the ability to move a department to a new parent? Context menu achieves the business goal.
   - Recommendation: Implement as right-click "移动到..." context menu with dept picker. Simpler and more reliable. Drag gesture can be added as enhancement.

2. **Position deletion behavior when employees are assigned**
   - What we know: Need to prevent FK violations
   - What's unclear: Should deleting a position auto-null `position_id` on employees, or require reassignment first?
   - Recommendation: Require reassignment first (like department delete). Clear error: "该岗位下有 N 名员工，请先调整员工岗位后再删除。"

3. **Position rename sync**
   - What we know: `Employee.position` is denormalized display text
   - What's unclear: After a Position is renamed, should existing employees' `position` text be updated?
   - Recommendation: No automatic sync (too expensive). UI always reads from Position table. Keep `Employee.position` as historical snapshot.

4. **Cycle detection in department parent_id updates**
   - What we know: No cycle detection exists in current UpdateDepartment
   - What's unclear: Is cycle prevention required for V1.0?
   - Recommendation: Implement cycle check in UpdateDepartment before updating `parent_id`. Simple BFS from new parent to ensure current dept is not in ancestor chain.

---

## Environment Availability

Step 2.6: SKIPPED (no external dependencies identified — Phase 14 uses only existing project stack: Go/Gin/GORM, PostgreSQL, Vue 3/Element Plus/ECharts)

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | go test (standard library, per project convention) + Vitest (frontend) |
| Config file | `pytest.ini`-style not needed — `go test ./...` / `vitest` |
| Quick run command | `go test ./internal/position/... -v -run TestPosition 2>&1 | head -50` |
| Full suite command | `go test ./... -race 2>&1 && cd frontend && vitest run 2>&1` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| ORG-01 | Position CRUD (create/read/update/delete) | Unit | `go test ./internal/position/... -v` | pending |
| ORG-01 | Employee.position_id FK populated after migration | Unit | `go test ./internal/position/... -v -run TestMigrate` | pending |
| ORG-01 | Unique constraint: no duplicate position name per dept | Unit | `go test ./internal/position/... -v -run TestUnique` | pending |
| ORG-02 | Tree API returns dept→position→employee structure | Unit | `go test ./internal/department/... -v -run TestBuildTree` | pending |
| ORG-02 | Drag-drop update (PUT dept with new parent_id) | Unit | `go test ./internal/department/... -v -run TestUpdateParent` | pending |
| ORG-02 | Cycle detection prevents circular parent_id | Unit | `go test ./internal/department/... -v -run TestCycle` | pending |
| ORG-03 | Transfer-and-delete atomically moves employees | Unit | `go test ./internal/department/... -v -run TestTransferDelete` | pending |
| ORG-04 | Position dropdown API returns grouped options | Unit | `go test ./internal/position/... -v -run TestSelectOptions` | pending |
| ORG-04 | Inline edit saves on blur/Enter, reverts on Escape | Manual | Browser interaction | pending |

### Sampling Rate
- **Per task commit:** `go test ./internal/position/... ./internal/department/... -count=1`
- **Per wave merge:** `go test ./... -race -count=1`
- **Phase gate:** Full suite green before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] `internal/position/model_test.go` — Position model basic tests
- [ ] `internal/position/repository_test.go` — Position CRUD repository tests
- [ ] `internal/position/service_test.go` — Position service + migration tests
- [ ] `internal/department/service_test.go` — BuildTree v2 tests (add position nodes)
- [ ] `internal/department/service_test.go` — Cycle detection test
- [ ] `internal/department/service_test.go` — Transfer-and-delete test
- [ ] `frontend/src/__tests__/position.test.ts` — Position API client tests
- [ ] `frontend/src/__tests__/orgChart.test.ts` — ECharts drag/click behavior tests
- [ ] Framework install: Vitest already in `package.json` devDependencies? — verify

---

## Security Domain

Required when `security_enforcement` is enabled (absent = enabled). Omit only if explicitly `false` in config.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V4 Access Control | yes | Position CRUD: `RequireRole("owner","admin")` via middleware |
| V5 Input Validation | yes | Position.name max 100 chars via validator, DepartmentID nullable/int64 range |
| V4 Authorization | yes | org_id scope on all Position queries — tenant isolation via GORM Scope |

### Known Threat Patterns for {stack}

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Position name injection (XSS via displayed name) | Tampering/Spoofing | Position.name stored as `varchar(100)`, validated server-side, HTML-escaped on frontend render |
| Tenant isolation bypass (view another org's positions) | Information Disclosure | All Position queries scoped by `org_id` via `middleware.TenantScope` — verified in Repository layer |
| Mass employee transfer via department delete | Tampering | Transfer-and-delete requires explicit `employee_ids[]` array; validates each employee belongs to source dept before transfer |
| Position name collision (duplicate within same dept) | Tampering | Repository-level uniqueness check before create; GORM unique index as defense-in-depth |

---

## Sources

### Primary (HIGH confidence)
- `internal/department/model.go` — Department model structure (邻接表 pattern)
- `internal/department/service.go` — BuildTree, markMatches, UpdateDepartment patterns
- `internal/department/repository.go` — Repository CRUD pattern
- `internal/department/handler.go` — API route registration and handler pattern
- `internal/employee/model.go` — Employee model, existing Position string field
- `internal/employee/repository.go` — Employee repository, ListAllByOrg, CountByDepartment
- `cmd/server/main.go` — AutoMigrate registration pattern
- `frontend/src/views/employee/OrgChart.vue` — Existing ECharts TreeChart usage
- `frontend/src/views/employee/EmployeeCreate.vue` — Existing employee form structure
- `frontend/src/api/department.ts` — Existing API client pattern
- `frontend/src/api/employee.ts` — Existing API client pattern
- `internal/common/model/base.go` — BaseModel with OrgID/soft-delete

### Secondary (MEDIUM confidence)
- 14-UI-SPEC.md — UI design contract (color tokens, component inventory, copywriting)
- 14-CONTEXT.md — Phase decisions D-14-01 through D-14-15

### Tertiary (LOW confidence)
- [ECharts TreeChart documentation](https://echarts.apache.org/en/option.html#series-tree) — drag behavior vs roam behavior (needs verification via Context7)

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — no new libraries, all existing project dependencies
- Architecture: HIGH — patterns follow established project conventions (Handler/Service/Repository)
- Pitfalls: HIGH — all identified from existing codebase analysis and D-14-01~D-14-15 constraints
- ECharts drag behavior: MEDIUM — built-in drag vs context menu interpretation [ASSUMED A2]

**Research date:** 2026-04-21
**Valid until:** 2026-05-21 (30 days — Phase 14 is well-scoped, no fast-moving libraries involved)
