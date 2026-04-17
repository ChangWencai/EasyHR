# Phase 5: 员工管理增强 + 组织架构基础 - Pattern Map

**Mapped:** 2026-04-17
**Files analyzed:** 25 (15 new + 10 modified)
**Analogs found:** 23 / 25

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/department/model.go` | model | CRUD | `internal/employee/model.go` | exact |
| `internal/department/dto.go` | model | CRUD | `internal/employee/dto.go` | exact |
| `internal/department/repository.go` | repository | CRUD | `internal/employee/repository.go` | exact |
| `internal/department/service.go` | service | CRUD | `internal/employee/service.go` | exact |
| `internal/department/handler.go` | controller | request-response | `internal/employee/contract_handler.go` | exact |
| `internal/employee/registration_model.go` | model | request-response | `internal/employee/invitation_model.go` | exact |
| `internal/employee/registration_dto.go` | model | request-response | `internal/employee/invitation_dto.go` | exact |
| `internal/employee/registration_repository.go` | repository | CRUD | `internal/employee/invitation_repository.go` | exact |
| `internal/employee/registration_service.go` | service | CRUD | `internal/employee/invitation_service.go` | exact |
| `internal/employee/registration_handler.go` | controller | request-response | `internal/employee/invitation_handler.go` | exact |
| `internal/sms/service.go` | service | event-driven | `internal/common/middleware/ratelimit.go` | partial |
| `internal/sms/templates.go` | config | -- | (no analog) | none |
| `internal/employee/model.go` (修改) | model | -- | (self) | -- |
| `internal/employee/service.go` (修改) | service | CRUD | (self) | -- |
| `internal/employee/offboarding_service.go` (修改) | service | CRUD | (self) | -- |
| `internal/employee/offboarding_model.go` (修改) | model | -- | (self) | -- |
| `internal/dashboard/service.go` (修改) | service | request-response | (self) | -- |
| `frontend/src/views/employee/EmployeeDashboard.vue` | component | request-response | `frontend/src/views/home/HomeView.vue` | role-match |
| `frontend/src/views/employee/OrgChart.vue` | component | request-response | (no analog - ECharts tree 新组件) | none |
| `frontend/src/views/employee/EmployeeDrawer.vue` | component | request-response | `frontend/src/views/employee/EmployeeList.vue` | role-match |
| `frontend/src/views/employee/RegistrationList.vue` | component | CRUD | `frontend/src/views/employee/InvitationList.vue` | exact |
| `frontend/src/views/employee/RegisterPage.vue` | component | request-response | `frontend/src/views/employee/EmployeeCreate.vue` | role-match |
| `frontend/src/views/employee/OffboardingList.vue` (修改) | component | request-response | (self) | -- |
| `frontend/src/api/department.ts` | utility | request-response | `frontend/src/api/employee.ts` | exact |
| `frontend/src/router/index.ts` (修改) | config | -- | (self) | -- |

## Pattern Assignments

### `internal/department/model.go` (model, CRUD)

**Analog:** `internal/employee/model.go`

**Model 定义模式** (lines 13-36):
```go
package department

import (
	"time"

	"github.com/wencai/easyhr/internal/common/model"
)

// Department 部门模型
type Department struct {
	model.BaseModel
	Name      string `gorm:"column:name;type:varchar(100);not null;index;comment:部门名称" json:"name"`
	ParentID  *int64 `gorm:"column:parent_id;index;comment:父部门ID（顶级为空）" json:"parent_id"`
	SortOrder int    `gorm:"column:sort_order;not null;default:0;comment:排序" json:"sort_order"`
}

func (Department) TableName() string { return "departments" }
```

**要点:**
- 嵌入 `model.BaseModel`（包含 ID/OrgID/CreatedBy/UpdatedAt/DeletedAt）
- ParentID 使用 `*int64` 指针类型表示可选（顶级部门为 nil）
- 表名使用复数形式

**BaseModel 结构** (`internal/common/model/base.go` lines 9-17):
```go
type BaseModel struct {
	ID        int64          `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	OrgID     int64          `gorm:"column:org_id;index;not null;comment:所属企业ID" json:"org_id"`
	CreatedBy int64          `gorm:"column:created_by;comment:创建人ID" json:"created_by"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedBy int64          `gorm:"column:updated_by;comment:更新人ID" json:"updated_by"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index;comment:软删除时间戳" json:"-"`
}
```

---

### `internal/department/dto.go` (model, CRUD)

**Analog:** `internal/employee/dto.go`

**DTO 模式** (lines 1-70):
```go
package department

// CreateDepartmentRequest 创建部门请求
type CreateDepartmentRequest struct {
	Name      string `json:"name" binding:"required,min=1,max=100"`
	ParentID  *int64 `json:"parent_id"`
	SortOrder int    `json:"sort_order"`
}

// UpdateDepartmentRequest 更新部门请求（部分更新）
type UpdateDepartmentRequest struct {
	Name      *string `json:"name" binding:"omitempty,min=1,max=100"`
	ParentID  **int64 `json:"parent_id"`
	SortOrder *int    `json:"sort_order"`
}

// DepartmentResponse 部门响应
type DepartmentResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	ParentID  *int64 `json:"parent_id"`
	SortOrder int    `json:"sort_order"`
}

// DepartmentListQueryParams 部门列表查询参数
type DepartmentListQueryParams struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
}

// TreeNode 组织架构树节点（用于 ECharts tree 数据）
type TreeNode struct {
	ID       int64       `json:"id"`
	Name     string      `json:"name"`
	Type     string      `json:"type"` // "department" | "position" | "employee"
	Children []*TreeNode `json:"children,omitempty"`
}
```

**要点:**
- Create 使用 `binding:"required"` 验证必填字段
- Update 使用指针类型 `*string` / `**int64` 支持部分更新（nil=不更新）
- TreeNode 是给前端 ECharts tree 使用的数据结构

---

### `internal/department/repository.go` (repository, CRUD)

**Analog:** `internal/employee/repository.go`

**Repository 模式** (lines 27-93):
```go
package department

import (
	"fmt"

	"github.com/wencai/easyhr/internal/common/middleware"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建部门
func (r *Repository) Create(dept *Department) error {
	return r.db.Create(dept).Error
}

// FindByID 根据 ID 查找部门（带租户隔离）
func (r *Repository) FindByID(orgID, id int64) (*Department, error) {
	var dept Department
	err := r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).First(&dept).Error
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

// Update 更新部门信息（部分更新）
func (r *Repository) Update(orgID, id int64, updates map[string]interface{}) error {
	result := r.db.Model(&Department{}).Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete 软删除部门
func (r *Repository) Delete(orgID, id int64) error {
	result := r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).Delete(&Department{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListAll 查询企业全部部门（用于构建树）
func (r *Repository) ListAll(orgID int64) ([]Department, error) {
	var depts []Department
	err := r.db.Scopes(middleware.TenantScope(orgID)).Order("sort_order ASC, id ASC").Find(&depts).Error
	if err != nil {
		return nil, fmt.Errorf("list departments: %w", err)
	}
	return depts, nil
}
```

**要点:**
- 所有查询都使用 `middleware.TenantScope(orgID)` 注入 org_id 条件
- Update/Delete 检查 `RowsAffected == 0` 返回 `gorm.ErrRecordNotFound`
- ListAll 不分页（全量查询用于内存构建树）

---

### `internal/department/service.go` (service, CRUD)

**Analog:** `internal/employee/service.go`

**Service 模式** (lines 14-25):
```go
package department

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}
```

**树构建核心逻辑** (参考 RESEARCH.md Pattern 1):
```go
// BuildTree 从扁平记录构建树（单次查询 + 内存递归，3层深度足够）
func (s *Service) BuildTree(depts []Department, employees []employee.Employee) []*TreeNode {
	// 1. 按 parent_id 分组部门
	// 2. 递归构建部门树（最多3层）
	// 3. 员工按 department_id 挂载到对应部门节点下
	// 4. 返回根节点列表（parent_id IS NULL 的部门）
}
```

**要点:**
- 新建 Department 模块需要独立 package，但读取 Employee 时需避免循环依赖
- 方案：Service 接受 employee.Repository 作为依赖注入参数（参考 OffboardingService 注入 empRepo 的模式）

---

### `internal/department/handler.go` (controller, request-response)

**Analog:** `internal/employee/contract_handler.go`

**Handler 注册模式** (lines 24-39):
```go
package department

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
)

type DepartmentHandler struct {
	svc *Service
}

func NewDepartmentHandler(svc *Service) *DepartmentHandler {
	return &DepartmentHandler{svc: svc}
}

func (h *DepartmentHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	authGroup := rg.Group("")
	authGroup.Use(authMiddleware)

	// 部门 CRUD — OWNER/ADMIN
	authGroup.POST("/departments", middleware.RequireRole("owner", "admin"), h.CreateDepartment)
	authGroup.GET("/departments", h.ListDepartments) // 所有角色
	authGroup.GET("/departments/:id", h.GetDepartment)
	authGroup.PUT("/departments/:id", middleware.RequireRole("owner", "admin"), h.UpdateDepartment)
	authGroup.DELETE("/departments/:id", middleware.RequireRole("owner", "admin"), h.DeleteDepartment)
	// 组织架构树
	authGroup.GET("/departments/tree", h.GetTree)
}
```

**Handler 方法模式** (参考 contract_handler.go lines 42-65):
```go
func (h *DepartmentHandler) CreateDepartment(c *gin.Context) {
	var req CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	result, err := h.svc.CreateDepartment(orgID, userID, &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 20200, err.Error())
		return
	}

	response.Success(c, result)
}
```

**要点:**
- 使用 `response.Success/Error/BadRequest/PageSuccess` 统一响应格式
- `c.GetInt64("org_id")` 和 `c.GetInt64("user_id")` 从 auth middleware 注入的 context 获取
- `middleware.RequireRole("owner", "admin")` 控制写操作权限
- 错误码使用 20200+ 范围（参考现有 handler 的 code 分配模式）

---

### `internal/employee/registration_model.go` (model, request-response)

**Analog:** `internal/employee/invitation_model.go`

**Token 模型模式** (lines 9-26):
```go
package employee

import (
	"time"
)

// Registration 员工信息登记模型
type Registration struct {
	ID         int64      `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	OrgID      int64      `gorm:"column:org_id;index;not null;comment:所属企业ID" json:"org_id"`
	EmployeeID *int64     `gorm:"column:employee_id;index;comment:关联员工ID" json:"employee_id"`
	Token      string     `gorm:"column:token;type:varchar(64);uniqueIndex;not null;comment:登记Token" json:"token"`
	Status     string     `gorm:"column:status;type:varchar(20);not null;default:pending;comment:状态" json:"status"`
	ExpiresAt  time.Time  `gorm:"column:expires_at;not null;comment:过期时间" json:"expires_at"`
	UsedAt     *time.Time `gorm:"column:used_at;comment:使用时间" json:"used_at"`
	CreatedBy  int64      `gorm:"column:created_by;not null;comment:创建人ID" json:"created_by"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
}

func (Registration) TableName() string { return "registrations" }

const (
	RegistrationStatusPending = "pending"
	RegistrationStatusUsed    = "used"
	RegistrationStatusExpired = "expired"
)

const RegistrationExpiryDuration = 7 * 24 * time.Hour
```

**要点:**
- 复用 Invitation 的 Token 模式：crypto/rand 32-byte hex (64 chars)
- 独立模型，不嵌入 BaseModel（简化字段，仅需 ID/OrgID/Token/Status/ExpiresAt）
- EmployeeID 指针类型表示可选（创建时可能不关联已有员工）

---

### `internal/employee/registration_service.go` (service, CRUD)

**Analog:** `internal/employee/invitation_service.go`

**Token 生成复用** (invitation_service.go lines 46-52):
```go
// generateToken 使用 crypto/rand 生成 32 字节随机数，返回 64 字符 hex 字符串
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("生成 token 失败: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
```

**事务提交模式** (invitation_service.go lines 193-228):
```go
return s.invRepo.DB().Transaction(func(tx *gorm.DB) error {
	// 在事务中校验唯一性并创建 Employee
	if emp.PhoneHash != "" {
		var count int64
		tx.Model(&Employee{}).Scopes(middleware.TenantScope(emp.OrgID)).
			Where("phone_hash = ?", emp.PhoneHash).Count(&count)
		if count > 0 {
			return fmt.Errorf("该手机号已存在")
		}
	}
	// ...创建 Employee + 更新 Invitation 状态
	return nil
})
```

**Registration 提交差异点** (D-08 提交即更新):
- 已有员工匹配（phone_hash / id_card_hash）-> 覆盖更新
- 无匹配员工 -> 创建新 Employee 记录
- 更新 Registration 状态为 used

**Service 构造函数模式** (invitation_service.go lines 25-29):
```go
type RegistrationService struct {
	regRepo   *RegistrationRepository
	empRepo   *Repository
	cryptoCfg config.CryptoConfig
}

func NewRegistrationService(regRepo *RegistrationRepository, empRepo *Repository, cryptoCfg config.CryptoConfig) *RegistrationService {
	return &RegistrationService{regRepo: regRepo, empRepo: empRepo, cryptoCfg: cryptoCfg}
}
```

---

### `internal/employee/registration_handler.go` (controller, request-response)

**Analog:** `internal/employee/invitation_handler.go`

**公开路由 + 认证路由混合模式** (lines 23-36):
```go
func (h *RegistrationHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	authGroup := rg.Group("")
	authGroup.Use(authMiddleware)

	// 公开接口（无需认证）— 员工填写信息
	rg.GET("/registrations/:token", h.GetRegistrationDetail)
	rg.POST("/registrations/:token/submit", h.SubmitRegistration)

	// 需要认证的接口 — 管理员操作
	authGroup.POST("/registrations", middleware.RequireRole("owner", "admin"), h.CreateRegistration)
	authGroup.GET("/registrations", h.ListRegistrations)
}
```

**公开接口错误处理模式** (invitation_handler.go lines 102-129):
```go
func (h *InvitationHandler) SubmitInvitation(c *gin.Context) {
	token := c.Param("token")

	var req SubmitInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.SubmitInvitation(token, &req); err != nil {
		if err == ErrInvitationNotFound {
			response.Error(c, http.StatusNotFound, 20206, err.Error())
			return
		}
		if err == ErrInvitationExpired {
			response.Error(c, http.StatusGone, 20207, err.Error())
			return
		}
		response.Error(c, http.StatusBadRequest, 20209, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "提交成功"})
}
```

**要点:**
- 公开接口路由注册在 `rg`（非 `authGroup`）上，不使用 authMiddleware
- Token 过期返回 410 Gone（`http.StatusGone`）
- Token 已使用/不存在返回 400/404

---

### `internal/sms/service.go` (service, event-driven)

**Analog:** 无完全匹配。参考 `internal/common/middleware/ratelimit.go` 的 Redis 使用模式和 `internal/employee/service.go` 的 config 注入模式。

**Service 构造函数模式** (参考 invitation_service.go):
```go
package sms

import (
	"github.com/wencai/easyhr/internal/common/config"
	"github.com/wencai/easyhr/internal/common/logger"
	"go.uber.org/zap"
)

type Service struct {
	cfg config.SMSConfig
}

func NewService(cfg config.SMSConfig) *Service {
	return &Service{cfg: cfg}
}

func (s *Service) SendRegistrationLink(phone string, token string, orgName string) error {
	// 使用 resty 调用阿里云 SMS API
	// 参考 ratelimit.go 的 Redis key 前缀模式
}
```

**要点:**
- 新建独立 package `internal/sms/`
- 使用 resty 发送 HTTP 请求（已在 go.mod 中）
- 配置通过 config.SMSConfig 注入（AccessKey/Secret/SignName/TemplateCode）

---

### `internal/employee/model.go` (修改 - 新增 department_id)

**变更:** 在 Employee struct 中新增一行字段定义

```go
// 新增字段（在 Position 字段之后）
DepartmentID *int64 `gorm:"column:department_id;index;comment:所属部门ID" json:"department_id"`
```

**要点:**
- 使用指针 `*int64` 表示可选（员工可能未分配部门）
- 添加 index 加速按部门查询

---

### `internal/employee/offboarding_service.go` (修改 - 新增 RejectResign)

**Analog:** 同文件中的 `ApproveResign` 方法 (lines 135-167)

**审批方法模式:**
```go
// ApproveResign 审批通过员工离职申请 (lines 135-167)
func (s *OffboardingService) ApproveResign(orgID, approverID, offboardingID int64) error {
	ob, err := s.obRepo.FindByID(orgID, offboardingID)
	if err != nil {
		return fmt.Errorf("离职记录不存在")
	}
	if ob.Status != OffboardingStatusPending {
		return fmt.Errorf("当前状态不可审批（状态: %s）", ob.Status)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":      OffboardingStatusApproved,
		"approved_by": approverID,
		"approved_at": now,
		"updated_by":  approverID,
	}
	if err := s.obRepo.Update(orgID, offboardingID, updates); err != nil {
		return fmt.Errorf("审批离职失败: %w", err)
	}
	// ... 更新 Employee 状态
	return nil
}
```

**RejectResign 需遵循的模式:**
1. 查找 offboarding 记录 + 租户隔离
2. 验证状态为 pending（只有 pending 可以驳回）
3. map[string]interface{} updates 模式更新
4. 不更新 Employee 状态（驳回 = 撤销离职流程）

---

### `internal/employee/offboarding_model.go` (修改 - 新增 rejected 状态)

**变更:** 在状态常量块中新增一行

```go
const (
	OffboardingStatusPending   = "pending"
	OffboardingStatusApproved  = "approved"
	OffboardingStatusCompleted = "completed"
	OffboardingStatusRejected  = "rejected" // 新增
)
```

---

### `internal/dashboard/service.go` (修改 - 新增离职率计算)

**errgroup 并发聚合模式** (lines 40-127):
```go
g, ctx := errgroup.WithContext(ctx)

g.Go(func() error {
	active, joined, left, err := s.repo.GetEmployeeStats(ctx, orgID)
	if err != nil {
		return err
	}
	empStats.active = active
	empStats.joined = joined
	empStats.left = left
	return nil
})
// ... 其他并发查询

if err := g.Wait(); err != nil {
	return nil, err
}
```

**新增离职率计算（在 g.Wait() 之后）:**
```go
// 离职率 = 离职人数 / (离职人数 + 期末人数) x 100%  (D-02)
turnoverRate := 0.0
denominator := float64(empStats.left + empStats.active)
if denominator > 0 {
	turnoverRate = float64(empStats.left) / denominator * 100
}
```

---

### `frontend/src/views/employee/EmployeeDashboard.vue` (component, request-response)

**Analog:** `frontend/src/views/home/HomeView.vue` -- 数据概览卡片区域

**卡片布局模式** (HomeView.vue lines 83-104):
```vue
<div v-if="store.overviewExpanded && store.overview" class="overview-grid">
  <div class="overview-item">
    <div class="overview-value">{{ store.overview.employee_count }}</div>
    <div class="overview-label">在职员工</div>
  </div>
  <div class="overview-item">
    <div class="overview-value small">
      <span class="green">+{{ store.overview.joined_this_month }}</span>
      <span class="sep">/</span>
      <span class="red">-{{ store.overview.left_this_month }}</span>
    </div>
    <div class="overview-label">本月入/离职</div>
  </div>
</div>
```

**卡片样式模式** (HomeView.vue lines 367-403):
```scss
.overview-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.overview-item {
  text-align: center;
  padding: 14px 12px;
  background: #fafafa;
  border-radius: 8px;
}

.overview-value {
  font-size: 24px;
  font-weight: 700;
  color: #1677ff;
  line-height: 1.2;
}

.overview-label {
  font-size: 12px;
  color: #8c8c8c;
  margin-top: 4px;
}
```

**要点:**
- 4 张纯数字卡片（在职人数/新入职/离职/离职率），不引入图表
- 使用 CSS Grid 2x2 布局
- 值使用 700 font-weight + #1677ff 蓝色主题（与首页一致）
- 3 步内完成查看 = 直接展示，无需操作

---

### `frontend/src/views/employee/OrgChart.vue` (component, request-response)

**Analog:** 无直接匹配 -- ECharts tree 组件是新引入的模式

**ECharts tree 基本模式** (参考 RESEARCH.md Pattern 3):
```vue
<template>
  <div class="org-chart">
    <el-input v-model="searchKeyword" placeholder="搜索部门/岗位/员工" clearable class="search-input" />
    <v-chart :option="chartOption" autoresize style="height: 600px" />
  </div>
</template>

<script setup lang="ts">
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { TreeChart } from 'echarts/charts'
import { TooltipComponent, TitleComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

use([TreeChart, TooltipComponent, TitleComponent, CanvasRenderer])
</script>
```

**要点:**
- 按需引入 ECharts 模块（不要全量 `import * as echarts`）
- 搜索高亮通过遍历 chartOption.data 设置 itemStyle.color 实现
- 关闭初始动画（`animation: false`）提升50人+节点性能

---

### `frontend/src/views/employee/EmployeeDrawer.vue` (component, request-response)

**Analog:** `frontend/src/views/employee/EmployeeList.vue` -- 同模块内的组件模式

**Element Plus Drawer 模式:**
```vue
<template>
  <el-drawer v-model="visible" title="员工详情" size="480px" direction="rtl">
    <el-descriptions :column="1" border>
      <el-descriptions-item label="姓名">{{ employee.name }}</el-descriptions-item>
      <el-descriptions-item label="状态">
        <el-tag :type="statusTagType[employee.status]" size="small">{{ statusMap[employee.status] }}</el-tag>
      </el-descriptions-item>
      <!-- 更多字段... -->
    </el-descriptions>
  </el-drawer>
</template>
```

**要点:**
- 使用 `el-drawer` 组件，`direction="rtl"` 从右侧滑出（D-10）
- 使用 `el-descriptions` 展示键值对信息
- 调用 `GetSensitiveInfo` API 获取完整信息（需 ADMIN 权限）
- 父组件通过 `v-model` 控制显示/隐藏，通过 prop 传入 employeeId

---

### `frontend/src/views/employee/RegistrationList.vue` (component, CRUD)

**Analog:** `frontend/src/views/employee/InvitationList.vue` (虽然未读取，但结构与 EmployeeList.vue 相同)

**列表页模式** (EmployeeList.vue lines 57-104):
```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { registrationApi } from '@/api/employee'
import { ElMessage } from 'element-plus'

const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

async function load(p = 1) {
  page.value = p
  loading.value = true
  try {
    const res = await registrationApi.list({ page: p, page_size: pageSize.value })
    list.value = res.list
    total.value = res.total
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => load())
</script>
```

**要点:**
- 标准列表页模式：ref 声明状态 + async load + onMounted 触发
- 使用 `ElMessage.error` 统一错误提示
- 分页使用 `el-pagination` 组件

---

### `frontend/src/views/employee/RegisterPage.vue` (component, request-response)

**Analog:** `frontend/src/views/employee/EmployeeCreate.vue` -- 表单页面模式

**表单页模式** (参考 EmployeeCreate 的 Element Plus 表单):
```vue
<template>
  <div class="register-page">
    <h2>{{ orgName }} - 员工信息登记</h2>
    <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
      <el-form-item label="姓名" prop="name">
        <el-input v-model="form.name" />
      </el-form-item>
      <!-- 更多字段... -->
      <el-form-item>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">提交</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>
```

**要点:**
- 独立路由 `/register/:token`，无需登录认证
- 从 URL 参数获取 token -> 调用公开 API 获取详情
- 提交后调用公开 API `/registrations/:token/submit`
- 敏感字段（手机号/身份证/银行卡号）在提交前不加密（HTTPS 传输安全）

---

### `frontend/src/api/department.ts` (utility, request-response)

**Analog:** `frontend/src/api/employee.ts`

**API 层模式** (employee.ts lines 1-91):
```typescript
import request from './request'

export interface Department {
  id: number
  name: string
  parent_id: number | null
  sort_order: number
}

export interface TreeNode {
  id: number
  name: string
  type: 'department' | 'position' | 'employee'
  children?: TreeNode[]
}

export const departmentApi = {
  list: () => request.get<Department[]>('/departments'),

  getTree: () => request.get<TreeNode[]>('/departments/tree'),

  create: (data: Partial<Department>) => request.post<Department>('/departments', data),

  update: (id: number, data: Partial<Department>) =>
    request.put<Department>(`/departments/${id}`, data),

  delete: (id: number) => request.delete(`/departments/${id}`),
}
```

**要点:**
- 使用 `request` 实例（已配置 baseURL、token 拦截器）
- TypeScript 接口定义请求/响应类型
- RESTful 路径约定：`/departments`、`/departments/:id`、`/departments/tree`

---

### `frontend/src/router/index.ts` (修改 - 新增路由)

**路由注册模式** (index.ts lines 12-38):
```typescript
// 员工管理（扩展）
{
  path: '/employee/org-chart',
  name: 'employee-org-chart',
  component: () => import('@/views/employee/OrgChart.vue'),
},
{
  path: '/employee/registrations',
  name: 'employee-registrations',
  component: () => import('@/views/employee/RegistrationList.vue'),
},

// 独立页面（不走 AppLayout，无需登录）
{
  path: '/register/:token',
  name: 'register',
  component: () => import('@/views/employee/RegisterPage.vue'),
},
```

**要点:**
- 组织架构和登记管理页面放在 `/employee` 子路由下
- RegisterPage 放在独立路由区域（不走 AppLayout，类似 `/login` 的模式）
- Auth Guard 中需排除 `/register` 路径（参考 lines 134-148 的 `isProtectedRoute` 判断）

---

### `frontend/src/views/employee/OffboardingList.vue` (修改 - 行内审批按钮)

**行内操作按钮模式** (OffboardingList.vue lines 42-59):
```vue
<el-table-column label="操作" width="180" fixed="right">
  <template #default="{ row }">
    <el-button
      v-if="row.status === 'pending'"
      size="small"
      type="primary"
      @click="handleApprove(row.id)"
    >
      同意
    </el-button>
    <!-- 新增：驳回按钮 -->
    <el-button
      v-if="row.status === 'pending'"
      size="small"
      type="danger"
      @click="handleReject(row.id)"
    >
      驳回
    </el-button>
    <!-- 新增：去减员按钮（approved 状态显示） -->
    <el-button
      v-if="row.status === 'approved'"
      size="small"
      type="warning"
      @click="goToSIRegister(row.employee_id, row.employee_name)"
    >
      去减员
    </el-button>
    <el-button
      v-if="row.status === 'approved'"
      size="small"
      type="success"
      @click="handleComplete(row.id)"
    >
      完成离职
    </el-button>
  </template>
</el-table-column>
```

**状态筛选新增 rejected** (OffboardingList.vue lines 13-19 扩展):
```vue
<el-option label="已驳回" value="rejected" />
```

---

## Shared Patterns

### 多租户隔离 (TenantScope)
**Source:** `internal/common/middleware/tenant.go`
**Apply to:** 所有 department 和 registration 的 Repository 方法
```go
func TenantScope(orgID int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("org_id = ?", orgID)
	}
}
```

### 敏感字段加密/脱敏 (AES-256-GCM + SHA-256 + Mask)
**Source:** `internal/common/crypto/aes.go`, `hash.go`, `mask.go`
**Apply to:** `registration_service.go` 提交时加密员工敏感字段
```go
// 加密: crypto.Encrypt(plaintext, aesKey) -> ciphertext
// 哈希: crypto.HashSHA256(plaintext) -> hash (用于唯一性校验和搜索)
// 脱敏: crypto.MaskPhone("13800008000") -> "138****8000"
//       crypto.MaskIDCard("110108199001011234") -> "110108****1234"
```

### 统一响应格式
**Source:** `internal/common/response/response.go`
**Apply to:** 所有 Handler 方法
```go
response.Success(c, data)              // 200 {"code":0, "message":"success", "data":...}
response.Error(c, status, code, msg)   // 4xx/5xx {"code":code, "message":msg, "data":null}
response.BadRequest(c, msg)            // 400 {"code":40000, "message":msg, "data":null}
response.PageSuccess(c, list, total, page, pageSize) // 200 带分页 meta
response.Forbidden(c, msg)             // 403 {"code":40300, "message":msg, "data":null}
```

### RBAC 权限控制
**Source:** `internal/common/middleware/rbac.go`
**Apply to:** 所有写操作路由
```go
// 路由注册时添加中间件
authGroup.POST("/departments", middleware.RequireRole("owner", "admin"), h.CreateDepartment)
```

### 前端 API 拦截器 (Token 注入 + 错误处理)
**Source:** `frontend/src/api/request.ts`
**Apply to:** 所有前端 API 调用
```typescript
// 自动注入 Bearer token（公开接口排除）
// 401 自动跳转登录页
// 响应自动解包 response.data
```

### Pinia Store 模式
**Source:** `frontend/src/stores/dashboard.ts`
**Apply to:** 员工数据看板和组织架构树如需全局状态
```typescript
export const useDashboardStore = defineStore('dashboard', () => {
  const todos = ref<TodoItem[]>([])
  const loading = ref(false)

  async function load() {
    loading.value = true
    try {
      const data = await fetchDashboard()
      todos.value = data.todos
    } finally {
      loading.value = false
    }
  }

  return { todos, loading, load }
})
```

### 测试模式 (Go)
**Source:** `internal/employee/service_test.go`
**Apply to:** 所有新增的 service 测试
```go
// SQLite 内存数据库 + AutoMigrate
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Discard,
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Employee{}))
	return db
}

// testify/assert + require 断言
// table-driven tests（参见 dashboard/service_test.go）
```

---

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `internal/sms/service.go` | service | event-driven | 项目中无 HTTP 外部服务调用封装的先例。Planner 应参考 resty 库文档和阿里云 SMS API 文档实现。降级方案：二维码+链接复制可替代短信发送（D-07） |
| `internal/sms/templates.go` | config | -- | 无匹配。仅为短信模板常量定义，trivial |
| `frontend/src/views/employee/OrgChart.vue` | component | request-response | ECharts tree 是新引入的依赖和组件模式。Planner 应参考 RESEARCH.md Pattern 3 的 vue-echarts 配置 |

## Wave 0 Dependencies

| Action | Command | Blocking? |
|--------|---------|-----------|
| 安装 ECharts + vue-echarts | `cd frontend && npm install echarts@6.0.0 vue-echarts@8.0.1` | Yes -- OrgChart.vue 依赖 |
| 创建阿里云 SMS 签名和模板 | 阿里云控制台操作 | No -- 二维码+链接可降级 |

## Metadata

**Analog search scope:**
- `internal/employee/` -- Employee/Invitation/Offboarding/Contract 全部文件
- `internal/dashboard/` -- Service/Model/Router/Handler/Test
- `internal/salary/` -- Model/DTO（用于花名册薪资关联参考）
- `internal/common/` -- Model/BaseModel, Middleware/TenantScope/RBAC/RateLimit, Response, Crypto, Config
- `frontend/src/` -- views/employee/, views/home/, api/, stores/, router/

**Files scanned:** 42
**Pattern extraction date:** 2026-04-17
