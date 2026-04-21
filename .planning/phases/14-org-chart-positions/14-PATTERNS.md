# Phase 14: 组织架构图（部门+岗位管理） - Pattern Map

**Mapped:** 2026-04-21
**Files analyzed:** 12
**Analogs found:** 12 / 12

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|-----------------|---------------|
| `internal/position/model.go` | model | CRUD | `internal/department/model.go` | exact |
| `internal/position/repository.go` | repository | CRUD | `internal/department/repository.go` | exact |
| `internal/position/service.go` | service | CRUD | `internal/department/service.go` | exact |
| `internal/position/handler.go` | handler | request-response | `internal/department/handler.go` | exact |
| `internal/position/dto.go` | dto | request-response | `internal/department/dto.go` | exact |
| `internal/employee/model.go` | model | CRUD | existing `internal/employee/model.go` | exact (add field) |
| `internal/department/service.go` | service | CRUD | existing `internal/department/service.go` | exact (add method) |
| `internal/department/repository.go` | repository | CRUD | existing `internal/department/repository.go` | exact (add method) |
| `cmd/server/main.go` | config | batch | existing `cmd/server/main.go` | exact (AutoMigrate) |
| `frontend/src/api/position.ts` | API client | request-response | `frontend/src/api/department.ts` | exact |
| `frontend/src/views/employee/OrgChart.vue` | component | event-driven | existing `OrgChart.vue` | exact (enhance) |
| `frontend/src/views/employee/EmployeeCreate.vue` | component | CRUD | existing `EmployeeCreate.vue` | exact (form field) |

---

## Pattern Assignments

### `internal/position/model.go` (model, CRUD)

**Analog:** `internal/department/model.go`

**BaseModel import + struct pattern** (lines 1-18):
```go
package position

import (
    "github.com/wencai/easyhr/internal/common/model"
)

type Position struct {
    model.BaseModel
    Name         string `gorm:"column:name;type:varchar(100);not null;index;comment:岗位名称" json:"name"`
    DepartmentID *int64 `gorm:"column:department_id;index;comment:所属部门（NULL=通用岗位）" json:"department_id"`
    SortOrder    int    `gorm:"column:sort_order;not null;default:0;comment:排序" json:"sort_order"`
}

func (Position) TableName() string {
    return "positions"
}
```

**Key difference from Department:** `DepartmentID` is `*int64` (nullable for common positions). Add JSON tag `json:"department_id"` matching `DepartmentID` field name.

---

### `internal/position/dto.go` (dto, request-response)

**Analog:** `internal/department/dto.go`

**DTO pattern** (lines 3-45):
```go
// CreatePositionRequest 创建岗位请求
type CreatePositionRequest struct {
    Name         string `json:"name" binding:"required,max=100"`
    DepartmentID *int64 `json:"department_id"`
    SortOrder    int    `json:"sort_order"`
}

// UpdatePositionRequest 更新岗位请求（指针类型支持部分更新）
type UpdatePositionRequest struct {
    Name         *string `json:"name"`
    DepartmentID *int64  `json:"department_id"`
    SortOrder    *int    `json:"sort_order"`
}

// PositionResponse 岗位响应
type PositionResponse struct {
    ID           int64  `json:"id"`
    OrgID        int64  `json:"org_id"`
    Name         string `json:"name"`
    DepartmentID *int64 `json:"department_id"`
    SortOrder    int    `json:"sort_order"`
}

// PositionSelectOptions 下拉选项（分组）
type PositionSelectOptions struct {
    DeptPositions    []PositionOption `json:"dept_positions"`
    CommonPositions  []PositionOption `json:"common_positions"`
    UnassignedOption PositionOption   `json:"unassigned_option"`
}

type PositionOption struct {
    ID   *int64 `json:"id"`
    Name string `json:"name"`
}
```

---

### `internal/position/repository.go` (repository, CRUD)

**Analog:** `internal/department/repository.go`

**Repository struct + tenant scope pattern** (lines 10-38):
```go
package position

import (
    "github.com/wencai/easyhr/internal/common/middleware"
    "gorm.io/gorm"
)

type Repository struct {
    db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
    return &Repository{db: db}
}

// Create 创建岗位
func (r *Repository) Create(pos *Position) error {
    return r.db.Create(pos).Error
}

// FindByID 根据 ID 查找岗位（带租户隔离）
func (r *Repository) FindByID(orgID, id int64) (*Position, error) {
    var pos Position
    err := r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).First(&pos).Error
    if err != nil {
        return nil, err
    }
    return &pos, nil
}

// Update 更新岗位信息（部分更新）
func (r *Repository) Update(orgID, id int64, updates map[string]interface{}) error {
    result := r.db.Model(&Position{}).Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).Updates(updates)
    if result.RowsAffected == 0 {
        return gorm.ErrRecordNotFound
    }
    return result.Error
}

// Delete 软删除岗位
func (r *Repository) Delete(orgID, id int64) error {
    result := r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).Delete(&Position{})
    if result.RowsAffected == 0 {
        return gorm.ErrRecordNotFound
    }
    return result.Error
}

// ListByOrg 获取全部岗位
func (r *Repository) ListByOrg(orgID int64) ([]Position, error) {
    var positions []Position
    err := r.db.Scopes(middleware.TenantScope(orgID)).Order("sort_order ASC, id ASC").Find(&positions).Error
    return positions, err
}

// ExistsByNameAndDept 检查同部门同名岗位（IS NOT DISTINCT FROM for NULL-safe comparison）
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

**Key difference from Department repo:** `ExistsByNameAndDept` handles NULL `department_id` with explicit `IS NULL` check (PostgreSQL NULL semantics — NULL != NULL in raw unique constraints).

---

### `internal/position/service.go` (service, CRUD)

**Analog:** `internal/department/service.go`

**Service struct + dependency injection** (lines 11-23):
```go
type Service struct {
    repo    *Repository
    empRepo *employee.Repository
}

func NewService(repo *Repository, empRepo *employee.Repository) *Service {
    return &Service{repo: repo, empRepo: empRepo}
}

// CreatePosition 创建岗位（含去重校验）
func (s *Service) CreatePosition(orgID, userID int64, req *CreatePositionRequest) (*PositionResponse, error) {
    exists, err := s.repo.ExistsByNameAndDept(orgID, req.DepartmentID, req.Name)
    if err != nil {
        return nil, fmt.Errorf("check position: %w", err)
    }
    if exists {
        return nil, ErrPositionDuplicate
    }
    pos := &Position{
        Name:         req.Name,
        DepartmentID: req.DepartmentID,
        SortOrder:    req.SortOrder,
    }
    pos.OrgID = orgID
    pos.CreatedBy = userID
    pos.UpdatedBy = userID
    if err := s.repo.Create(pos); err != nil {
        return nil, fmt.Errorf("create position: %w", err)
    }
    return toPositionResponse(pos), nil
}
```

**Sentinel errors** (add to repository or service package):
```go
var (
    ErrPositionNotFound   = errors.New("岗位不存在")
    ErrPositionDuplicate  = errors.New("同一部门内该岗位名称已存在")
    ErrPositionInUse      = errors.New("该岗位下有员工，无法删除")
)
```

**GetSelectOptions pattern** — for employee position dropdown:
```go
func (s *Service) GetSelectOptions(orgID int64, deptID *int64) (*PositionSelectOptions, error) {
    positions, err := s.repo.ListByOrg(orgID)
    if err != nil {
        return nil, err
    }
    var deptPositions, commonPositions []PositionOption
    for _, p := range positions {
        opt := PositionOption{ID: &p.ID, Name: p.Name}
        if p.DepartmentID == nil {
            commonPositions = append(commonPositions, opt)
        } else if deptID != nil && *p.DepartmentID == *deptID {
            deptPositions = append(deptPositions, opt)
        }
    }
    return &PositionSelectOptions{
        DeptPositions:    deptPositions,
        CommonPositions:  commonPositions,
        UnassignedOption:  PositionOption{ID: nil, Name: "未分配岗位"},
    }, nil
}
```

**Data migration pattern** — migrate existing `Employee.position` text to Position table:
```go
// MigrateFromEmployeeText 从 Employee.position 文本迁移到 Position 表
func (s *Service) MigrateFromEmployeeText(orgID int64) error {
    // Check if already migrated (any Position records exist)
    existing, _ := s.repo.ListByOrg(orgID)
    if len(existing) > 0 {
        return nil // Already migrated
    }
    employees, err := s.empRepo.ListAllByOrg(orgID)
    if err != nil {
        return err
    }
    // Group unique position names
    posNames := make(map[string]bool)
    for _, emp := range employees {
        if emp.Position != "" {
            posNames[emp.Position] = true
        }
    }
    // Create Position records and update employees
    for name := range posNames {
        pos := &Position{OrgID: orgID, Name: name, SortOrder: 0}
        if err := s.repo.Create(pos); err != nil {
            return err
        }
    }
    // Batch update employee position_id references
    return nil
}
```

---

### `internal/position/handler.go` (handler, request-response)

**Analog:** `internal/department/handler.go`

**Handler struct + RegisterRoutes pattern** (lines 13-35):
```go
type PositionHandler struct {
    svc *Service
}

func NewPositionHandler(svc *Service) *PositionHandler {
    return &PositionHandler{svc: svc}
}

func (h *PositionHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
    authGroup := rg.Group("")
    authGroup.Use(authMiddleware, middleware.RequireOrg)

    authGroup.POST("/positions", middleware.RequireRole("owner", "admin"), h.CreatePosition)
    authGroup.GET("/positions", h.ListPositions)
    authGroup.GET("/positions/select-options", h.GetSelectOptions)
    authGroup.PUT("/positions/:id", middleware.RequireRole("owner", "admin"), h.UpdatePosition)
    authGroup.DELETE("/positions/:id", middleware.RequireRole("owner", "admin"), h.DeletePosition)
}
```

**Handler method pattern** (lines 37-58):
```go
func (h *PositionHandler) CreatePosition(c *gin.Context) {
    var req CreatePositionRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "参数错误: "+err.Error())
        return
    }
    orgID := c.GetInt64("org_id")
    userID := c.GetInt64("user_id")
    pos, err := h.svc.CreatePosition(orgID, userID, &req)
    if err != nil {
        response.Error(c, http.StatusBadRequest, 20300, err.Error())
        return
    }
    response.Success(c, pos)
}
```

**Key routes to add:**
- `GET /positions/select-options?department_id=N` — for employee position dropdown (D-14-13, D-14-14, D-14-15)

---

### `internal/employee/model.go` (model, CRUD — add field)

**Analog:** existing `internal/employee/model.go`

**Add field after `Position string`** (line 22):
```go
// Employee 员工档案模型
type Employee struct {
    model.BaseModel
    // ... existing fields up to Position string ...
    Position                string     `gorm:"column:position;type:varchar(100);not null;index;comment:岗位" json:"position"`
    PositionID              *int64     `gorm:"column:position_id;index;comment:岗位ID（FK，NULL=未分配）" json:"position_id"` // NEW — D-14-01
    DepartmentID            *int64     `gorm:"column:department_id;index;comment:所属部门ID" json:"department_id"`
    // ... rest unchanged ...
}
```

---

### `internal/department/service.go` (service, CRUD — add methods)

**Analog:** existing `internal/department/service.go`

**New method: TransferAndDeleteDepartment** (add after `DeleteDepartment`):
```go
// TransferAndDeleteDepartment 转移员工后删除部门（D-14-09）
func (s *Service) TransferAndDeleteDepartment(orgID, deptID, targetDeptID int64, employeeIDs []int64) error {
    // 1. Validate target dept exists
    _, err := s.repo.FindByID(orgID, targetDeptID)
    if err != nil {
        return ErrDepartmentNotFound
    }
    // 2. Validate employees belong to source dept
    for _, empID := range employeeIDs {
        emp, err := s.empRepo.FindByID(orgID, empID)
        if err != nil {
            return ErrEmployeeNotFound
        }
        if emp.DepartmentID == nil || *emp.DepartmentID != deptID {
            return ErrEmployeeNotInDepartment
        }
    }
    // 3. Transfer employees (position_id stays, D-14-09)
    for _, empID := range employeeIDs {
        if err := s.empRepo.UpdateDepartmentID(orgID, empID, targetDeptID); err != nil {
            return fmt.Errorf("transfer employee %d: %w", empID, err)
        }
    }
    // 4. Delete department (no employees remain)
    return s.repo.Delete(orgID, deptID)
}
```

**BuildTree v2 signature change:** Add `positions []position.Position` parameter, replace position grouping logic from `emp.Position` string to `emp.PositionID` FK join. Keep `markMatches` unchanged.

```go
// BuildTree v2 signature — D-14-01, D-14-02
func (s *Service) BuildTree(departments []Department, employees []employee.Employee, positions []position.Position) []*TreeNode
```

**Cycle detection in UpdateDepartment** — add before applying `parent_id` update:
```go
func (s *Service) hasCycle(orgID int64, deptID, newParentID int64) (bool, error) {
    ancestors := make(map[int64]bool)
    current := newParentID
    for current != 0 {
        if current == deptID {
            return true, nil // Cycle detected
        }
        ancestors[current] = true
        parent, err := s.repo.FindByID(orgID, current)
        if err != nil || parent.ParentID == nil {
            break
        }
        current = *parent.ParentID
    }
    return false, nil
}
```

---

### `internal/department/repository.go` (repository — add method)

**Analog:** existing `internal/department/repository.go`

**Add after existing methods:**
```go
// FindAllByIDs 批量查询部门（用于目标部门下拉列表）
func (r *Repository) FindAllByIDs(orgID int64, ids []int64) ([]Department, error) {
    var departments []Department
    err := r.db.Scopes(middleware.TenantScope(orgID)).Where("id IN ?", ids).Find(&departments).Error
    return departments, err
}
```
Note: Existing `Update` method already supports `parent_id` field via `updates` map — no new repo method needed for UpdateParent.

---

### `cmd/server/main.go` (config, batch — AutoMigrate)

**Analog:** existing `cmd/server/main.go` AutoMigrate block

**AutoMigrate registration pattern** (add after `&department.Department{}`):
```go
// Phase 14: Position model AutoMigrate
import "github.com/wencai/easyhr/internal/position"

// In db.AutoMigrate() call, add:
&position.Position{},
```

**Handler registration pattern** (add after dept handler setup):
```go
// 岗位模块依赖注入
posRepo := position.NewRepository(db)
posSvc := position.NewService(posRepo, empRepo)
posHandler := position.NewPositionHandler(posSvc)
```

**Route registration** (add after `deptHandler.RegisterRoutes`):
```go
posHandler.RegisterRoutes(v1, authMiddleware)
```

**Position migration seed** (add after `deptSvc := department.NewService`):
```go
// Phase 14: Migrate existing employee position text to Position table on startup
if err := posSvc.MigrateFromEmployeePositions(); err != nil {
    logger.Logger.Warn("position migration failed", zap.Error(err))
}
```

---

## Frontend Pattern Assignments

### `frontend/src/api/position.ts` (API client, request-response)

**Analog:** `frontend/src/api/department.ts`

**API client pattern** (lines 30-45):
```typescript
import request from './request'

export interface Position {
  id: number
  name: string
  department_id: number | null
  sort_order: number
}

export interface PositionSelectOptions {
  dept_positions: Array<{ id: number; name: string }>
  common_positions: Array<{ id: number; name: string }>
  unassigned_option: { id: null; name: string }
}

export const positionApi = {
  list: (department_id?: number) =>
    request.get<Position[]>('/positions', { params: { department_id } }),

  getSelectOptions: (department_id?: number) =>
    request.get<PositionSelectOptions>('/positions/select-options', {
      params: { department_id },
    }),

  create: (data: { name: string; department_id?: number | null; sort_order?: number }) =>
    request.post<Position>('/positions', data),

  update: (id: number, data: Partial<Pick<Position, 'name' | 'department_id' | 'sort_order'>>) =>
    request.put<Position>(`/positions/${id}`, data),

  delete: (id: number) => request.delete(`/positions/${id}`),
}
```

---

### `frontend/src/views/employee/OrgChart.vue` (component, event-driven)

**Analog:** existing `OrgChart.vue` (same file — enhance)

**Key enhancement areas:**

**1. ECharts contextmenu for department move** — add after `chartInstance` init:
```typescript
chartInstance?.on('contextmenu', (params: unknown) => {
  const p = params as { data?: TreeNode; offsetX?: number; offsetY?: number }
  if (p.data?.type === 'department') {
    e.preventDefault()
    showMoveMenu(p.data.id, p.data.name, p.offsetX, p.offsetY)
  }
})
```

**2. Inline edit overlay** — coordinate with `params.event.offsetX/Y` from ECharts click event:
```typescript
chartInstance?.on('click', (params: unknown) => {
  const p = params as { data?: TreeNode; event?: { offsetX?: number; offsetY?: number } }
  if (p.data?.type === 'department' && p.data?.id) {
    showInlineEdit(p.data.id, p.data.name, p.event?.offsetX, p.event?.offsetY)
  }
})
```

**3. DeleteTransferDialog** — new component imported and shown when deleting a department with employees. Trigger: click delete button on department node, call `GET /employees?department_id=N` to list employees, show dialog.

**4. Add to imports** (line 86):
```typescript
import { positionApi } from '@/api/position'
```

---

### `frontend/src/views/employee/EmployeeCreate.vue` (component, CRUD — modify position field)

**Analog:** existing `EmployeeCreate.vue` (same file — modify position field)

**Modify position field: `el-input` -> `el-select` with grouped options** (lines 64-68 create mode, lines 176-187 edit mode):

```vue
<!-- Replace el-input with el-select (D-14-12, D-14-13, D-14-14, D-14-15) -->
<el-form-item label="岗位" prop="position_id" class="form-item">
  <el-select
    v-model="form.position_id"
    placeholder="请选择岗位"
    clearable
    size="large"
    class="full-width"
  >
    <el-option :value="null" label="未分配岗位" />
    <el-optgroup label="部门专属岗位" v-if="deptPositions.length">
      <el-option
        v-for="p in deptPositions"
        :key="p.id"
        :value="p.id"
        :label="p.name"
      />
    </el-optgroup>
    <el-optgroup label="通用岗位">
      <el-option
        v-for="p in commonPositions"
        :key="p.id"
        :value="p.id"
        :label="p.name"
      />
    </el-optgroup>
  </el-select>
</el-form-item>
```

**Add reactive data fields** (after `form` reactive object):
```typescript
const deptPositions = ref<Array<{ id: number; name: string }>>([])
const commonPositions = ref<Array<{ id: number; name: string }>>([])
```

**Load position options when department changes** (watch `form.department_id`):
```typescript
async function loadPositionOptions(deptId?: number | null) {
  try {
    const res = await positionApi.getSelectOptions(deptId ?? undefined)
    const data = (res as { data?: PositionSelectOptions }).data ?? (res as unknown as PositionSelectOptions)
    deptPositions.value = data.dept_positions
    commonPositions.value = data.common_positions
  } catch {
    deptPositions.value = []
    commonPositions.value = []
  }
}
```

**Watch department changes** (add near existing watchers):
```typescript
watch(() => form.department_id, (newDeptId) => {
  form.position_id = null // Reset position when dept changes
  loadPositionOptions(newDeptId)
})
```

**Update `form` reactive object** — add `position_id`:
```typescript
const form = reactive({
  // ... existing fields ...
  position: '',       // Keep for API compatibility, map from position.name on load
  position_id: null as number | null, // NEW — D-14-01
  // ...
})
```

**Update `loadEmployee`** — map `position_id` from API:
```typescript
Object.assign(form, {
  // ... existing fields ...
  position: emp.position,
  position_id: (emp as any).position_id ?? null,
})
```

**Update form submission data** — include both `position_id` and `position` name:
```typescript
const data = {
  ...form,
  // position_name resolved from position_id for API compatibility
}
```

---

## Shared Patterns

### Multi-tenancy (org_id scope)
**Source:** `internal/department/repository.go` lines 32-38, `internal/common/middleware/middleware.go` `TenantScope`
**Apply to:** All position CRUD queries
```go
err := r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).First(&pos).Error
```

### Error sentinel variables
**Source:** `internal/department/repository.go` lines 10-14
**Apply to:** All service files (add module-specific errors)
```go
var (
    ErrPositionNotFound  = errors.New("岗位不存在")
    ErrPositionDuplicate = errors.New("同一部门内该岗位名称已存在")
)
```

### Handler logging + response pattern
**Source:** `internal/department/handler.go` lines 37-57
**Apply to:** All handler files
```go
if err != nil {
    logger.SugarLogger.Debugw("CreatePosition: 失败", "error", err.Error(), "org_id", orgID)
    response.Error(c, http.StatusBadRequest, 20300, err.Error())
    return
}
response.Success(c, pos)
```

### Employee repository update pattern
**Source:** `internal/employee/repository.go`
**Apply to:** `UpdateDepartmentID` method needed for transfer operation
```go
func (r *Repository) UpdateDepartmentID(orgID, empID, deptID int64) error {
    return r.db.Model(&Employee{}).Scopes(middleware.TenantScope(orgID)).
        Where("id = ?", empID).Updates(map[string]interface{}{"department_id": deptID}).Error
}
```

### Department handler transfer-delete route
**Source:** `internal/department/handler.go` line 34
**Add after DELETE /departments/:id route:**
```go
authGroup.DELETE("/departments/:id/transfer", middleware.RequireRole("owner", "admin"), h.TransferDeleteDepartment)
```

### Frontend message composable
**Source:** `frontend/src/composables/useMessage.ts`
**Apply to:** All frontend files for user feedback
```typescript
import { useMessage } from '@/composables/useMessage'
const $msg = useMessage()
$msg.success('岗位创建成功')
$msg.error('创建失败')
```

---

## No Analog Found

All 12 files have strong analogs. No files require fallback to RESEARCH.md patterns alone.

---

## Metadata

**Analog search scope:**
- `internal/department/` (model.go, dto.go, repository.go, service.go, handler.go)
- `internal/employee/` (model.go, repository.go)
- `cmd/server/main.go`
- `frontend/src/api/department.ts`
- `frontend/src/views/employee/OrgChart.vue`
- `frontend/src/views/employee/EmployeeCreate.vue`

**Files scanned:** ~3,500 lines total
**Pattern extraction date:** 2026-04-21
