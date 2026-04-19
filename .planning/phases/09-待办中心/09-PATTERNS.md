# Phase 9: 待办中心 - Pattern Map

**Mapped:** 2026-04-19
**Files analyzed:** 18 new/modified files
**Analogs found:** 18/18 with matches

---

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/todo/model.go` | model | CRUD | `internal/dashboard/model.go` | exact-role |
| `internal/todo/repository.go` | repository | CRUD | `internal/dashboard/repository.go` | exact-role |
| `internal/todo/service.go` | service | CRUD | `internal/dashboard/service.go` | exact-role |
| `internal/todo/handler.go` | handler | request-response | `internal/employee/invitation_handler.go` | exact-role |
| `internal/todo/router.go` | router | request-response | `internal/dashboard/router.go` | exact |
| `internal/todo/excel.go` | utility | file-I/O | `internal/socialinsurance/excel.go` | exact-role |
| `internal/todo/scheduler.go` | service | event-driven | `internal/socialinsurance/scheduler.go` | exact |
| `internal/dashboard/service.go` | service (mod) | CRUD | `internal/dashboard/service.go` | self |
| `internal/dashboard/model.go` | model (mod) | CRUD | `internal/dashboard/model.go` | self |
| `frontend/src/views/home/components/TodoRingChart.vue` | component | transform | `frontend/src/views/employee/OrgChart.vue` | exact-role |
| `frontend/src/views/home/components/HomeCarousel.vue` | component | request-response | `frontend/src/views/home/HomeView.vue` | role-match |
| `frontend/src/views/home/HomeView.vue` | view (mod) | request-response | `frontend/src/views/home/HomeView.vue` | self |
| `frontend/src/views/todo/TodoListView.vue` | view | request-response | `frontend/src/views/home/HomeView.vue` | role-match |
| `frontend/src/views/todo/InviteFillPage.vue` | view | request-response | `frontend/src/views/employee/RegisterPage.vue` | exact |
| `frontend/src/api/todo.ts` | API client | request-response | `frontend/src/api/dashboard.ts` | exact |
| `frontend/src/api/carousel.ts` | API client | request-response | `frontend/src/api/dashboard.ts` | exact |
| `frontend/src/stores/dashboard.ts` | store (mod) | request-response | `frontend/src/stores/dashboard.ts` | self |
| `frontend/src/api/request.ts` | utility | request-response | (unchanged, referenced) | - |

---

## Pattern Assignments

### `internal/todo/model.go` (model, CRUD)

**Analog:** `internal/dashboard/model.go` (lines 1-46)

**Model structure** — replicate the same GORM + JSON struct pattern:

```go
// Copy from internal/dashboard/model.go lines 1-46
// TodoItem extends dashboard.TodoItem with deadline/urgency fields

type TodoType string

const (
    TodoStatusPending    TodoStatus = "pending"
    TodoStatusCompleted  TodoStatus = "completed"
    TodoStatusTerminated TodoStatus = "terminated"
)

type UrgencyStatus string

const (
    UrgencyNormal   UrgencyStatus = "normal"
    UrgencyOverdue  UrgencyStatus = "overdue"
    UrgencyExpired  UrgencyStatus = "expired"
)

// TodoItem — extended with D-09-03 fields
type TodoItem struct {
    ID             int64        `gorm:"primaryKey;autoIncrement" json:"id"`
    OrgID          int64        `gorm:"index;not null" json:"org_id"`
    Type           TodoType     `gorm:"type:varchar(50);not null;index" json:"type"`
    Title          string       `gorm:"type:varchar(200);not null" json:"title"`
    Description    string       `gorm:"type:text" json:"description,omitempty"`
    Status         TodoStatus   `gorm:"type:varchar(20);not null;default:pending;index" json:"status"`
    Priority       int          `gorm:"default:0" json:"priority"`
    IsTimeLimited  bool         `gorm:"default:false" json:"is_time_limited"`
    Deadline       *time.Time   `gorm:"index" json:"deadline,omitempty"`
    UrgencyStatus  UrgencyStatus `gorm:"type:varchar(20);default:normal" json:"urgency_status"`
    IsPinned       bool         `gorm:"default:false" json:"is_pinned"`
    SourceType     string       `gorm:"type:varchar(50)" json:"source_type,omitempty"` // contract/tax/si/...
    SourceID       int64        `json:"source_id,omitempty"`
    EmployeeID     *int64       `json:"employee_id,omitempty"`
    CreatedBy      int64        `gorm:"not null" json:"created_by"`
    CreatedAt      time.Time    `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt      time.Time    `gorm:"autoUpdateTime" json:"updated_at"`
}

func (TodoItem) TableName() string { return "todo_items" }

// CarouselItem model (D-09-06)
type CarouselItem struct {
    ID        int64      `gorm:"primaryKey;autoIncrement" json:"id"`
    OrgID     int64      `gorm:"index;not null" json:"org_id"`
    ImageURL  string     `gorm:"type:varchar(500);not null" json:"image_url"`
    LinkURL   string     `gorm:"type:varchar(500)" json:"link_url,omitempty"`
    SortOrder int        `gorm:"default:0" json:"sort_order"`
    Active    bool       `gorm:"default:false" json:"active"`
    StartAt   *time.Time `gorm:"index" json:"start_at,omitempty"`
    EndAt     *time.Time `gorm:"index" json:"end_at,omitempty"`
    CreatedBy int64      `gorm:"not null" json:"created_by"`
    CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (CarouselItem) TableName() string { return "carousel_items" }

// TodoInvite model (D-09-09/D-09-10)
type TodoInvite struct {
    ID        int64      `gorm:"primaryKey;autoIncrement" json:"id"`
    OrgID     int64      `gorm:"index;not null" json:"org_id"`
    TodoID    int64      `gorm:"index;not null" json:"todo_id"`
    Token     string     `gorm:"uniqueIndex;type:varchar(64);not null" json:"token"`
    Status    string     `gorm:"type:varchar(20);default:pending" json:"status"` // pending/used/expired
    ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
    UsedAt    *time.Time `json:"used_at,omitempty"`
    CreatedBy int64      `gorm:"not null" json:"created_by"`
    CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (TodoInvite) TableName() string { return "todo_invites" }

const TodoInviteExpiryDuration = 7 * 24 * time.Hour
```

---

### `internal/todo/repository.go` (repository, CRUD)

**Analog:** `internal/dashboard/repository.go` (lines 1-100)

**Multi-table GORM pattern** — same org_id scoping + middleware.TenantScope:

```go
// Copy from internal/dashboard/repository.go lines 1-33
import (
    "context"
    "fmt"
    "time"

    "github.com/wencai/easyhr/internal/common/middleware"
    "gorm.io/gorm"
)

// TodoRepositoryInterface
type TodoRepositoryInterface interface {
    List(ctx context.Context, orgID int64, query ListQuery) ([]TodoItem, int64, error)
    Create(ctx context.Context, item *TodoItem) error
    Update(ctx context.Context, orgID, id int64, updates map[string]interface{}) error
    Terminate(ctx context.Context, orgID, id int64) error
    UpdateUrgencyStatus(ctx context.Context) error
    ExistsBySource(ctx context.Context, orgID int64, sourceType string, sourceID int64) (bool, error)
}

type TodoRepositoryImpl struct {
    db *gorm.DB
}

func NewTodoRepository(db *gorm.DB) *TodoRepositoryImpl {
    return &TodoRepositoryImpl{db: db}
}

// List — copy WHERE pattern from dashboard/repository.go lines 36-71
// Scope: middleware.TenantScope(orgID)
// WHERE status IN (?) — filter by pending/completed/terminated
// WHERE deadline BETWEEN ? AND ? — date range filter (max 60 days)
// WHERE title LIKE ? — keyword search
// ORDER BY is_pinned DESC, deadline ASC, created_at DESC
func (r *TodoRepositoryImpl) List(ctx context.Context, orgID int64, query ListQuery) ([]TodoItem, int64, int, error) {
    var items []TodoItem
    var total int64

    db := r.db.Model(&TodoItem{}).Scopes(middleware.TenantScope(orgID))

    if query.Status != "" {
        db = db.Where("status = ?", query.Status)
    }
    if query.Keyword != "" {
        db = db.Where("title LIKE ?", "%"+query.Keyword+"%")
    }
    if !query.StartDate.IsZero() && !query.EndDate.IsZero() {
        db = db.Where("deadline BETWEEN ? AND ?", query.StartDate, query.EndDate)
    }
    if query.IsTimeLimited != nil {
        db = db.Where("is_time_limited = ?", *query.IsTimeLimited)
    }

    db.Count(&total)

    db = db.Order("is_pinned DESC, deadline ASC, created_at DESC")
    offset := (query.Page - 1) * query.PageSize
    db = db.Offset(offset).Limit(query.PageSize)

    if err := db.Find(&items).Error; err != nil {
        return nil, 0, 0, fmt.Errorf("list todos: %w", err)
    }
    return items, total, query.Page, nil
}

// UpdateUrgencyStatus — daily asynq scan (D-09-04)
// Copy anti-pattern fix from RESEARCH.md Pitfall 1:
// WHERE is_time_limited = true AND status NOT IN ('completed','terminated','expired')
func (r *TodoRepositoryImpl) UpdateUrgencyStatus(ctx context.Context) error {
    now := time.Now()
    // overdue: deadline in past OR deadline within 7 days
    // expired: deadline more than 15 days past
    return r.db.WithContext(ctx).Model(&TodoItem{}).
        Where("is_time_limited = ? AND status NOT IN ?", true,
            []string{"completed", "terminated", "expired"}).
        Where("deadline IS NOT NULL").
        Updates(map[string]interface{}{
            "urgency_status": computeUrgencyStatus(now),
        }).Error
}
```

---

### `internal/todo/service.go` (service, CRUD)

**Analog:** `internal/dashboard/service.go` (lines 1-50)

**Service interface + struct pattern:**

```go
// Copy from internal/dashboard/service.go lines 13-26
import (
    "context"
    "errors"
    "time"

    "github.com/wencai/easyhr/internal/common/config"
)

// ServiceInterface
type TodoServiceInterface interface {
    ListTodos(ctx context.Context, orgID int64, query ListQuery) (*TodoListResponse, error)
    InviteTodo(ctx context.Context, orgID, userID, todoID int64) (*InviteResponse, error)
    TerminateTodo(ctx context.Context, orgID, todoID int64) error
    PinTodo(ctx context.Context, orgID, todoID int64, pinned bool) error
    ValidateInviteToken(ctx context.Context, token string) (*TodoInvite, error)
    SubmitInvite(ctx context.Context, token string, data map[string]interface{}) error
}

type TodoService struct {
    repo       TodoRepository
    carouselRepo CarouselRepository
    cryptoCfg  config.CryptoConfig
}

func NewTodoService(repo TodoRepository, carouselRepo CarouselRepository, cryptoCfg config.CryptoConfig) *TodoService {
    return &TodoService{repo: repo, carouselRepo: carouselRepo, cryptoCfg: cryptoCfg}
}

// generateToken — copy exact pattern from internal/employee/invitation_service.go lines 46-52
import "crypto/rand"

func generateToken() (string, error) {
    bytes := make([]byte, 32)
    if _, err := rand.Read(bytes); err != nil {
        return "", fmt.Errorf("generate token: %w", err)
    }
    return hex.EncodeToString(bytes), nil
}

// urgency status computation — from RESEARCH.md Pattern 1
func computeUrgencyStatus(now time.Time) string {
    // Implementation per D-09-04 rules:
    // daysUntil < -15 → expired
    // daysUntil < 0 OR daysUntil <= 7 → overdue
    // else → normal
    // Note: called in SQL context, pass now as parameter
}
```

---

### `internal/todo/handler.go` (handler, request-response)

**Analog:** `internal/employee/invitation_handler.go` (lines 1-180)

**Handler pattern with mixed public/auth routes:**

```go
// Copy from internal/employee/invitation_handler.go lines 1-19
package todo

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/wencai/easyhr/internal/common/response"
)

// Handler — copy struct pattern
type Handler struct {
    svc *TodoService
}

// NewHandler — copy constructor
func NewHandler(svc *TodoService) *Handler {
    return &Handler{svc: svc}
}

// RegisterRoutes — copy from invitation_handler.go lines 23-36
// Public routes (no auth): GET /todos/:id/invite/verify, POST /todos/:invite/:token/submit
// Auth routes: GET /todos, POST /todos/:id/invite, PUT /todos/:id/terminate, PUT /todos/:id/pin, GET /todos/export
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
    authGroup := rg.Group("")
    authGroup.Use(authMiddleware)

    // Public — invite token verification and submission (D-09-09/D-09-10)
    rg.GET("/todos/invite/:token/verify", h.VerifyInviteToken)
    rg.POST("/todos/invite/:token/submit", h.SubmitInvite)

    // Auth required
    authGroup.GET("/todos", h.ListTodos)
    authGroup.POST("/todos/:id/invite", h.InviteTodo)
    authGroup.PUT("/todos/:id/terminate", h.TerminateTodo)
    authGroup.PUT("/todos/:id/pin", h.PinTodo)
    authGroup.GET("/todos/export", h.ExportTodos)
    authGroup.GET("/carousels", h.ListCarousels)
    authGroup.POST("/carousels", h.CreateCarousel)
    authGroup.PUT("/carousels/:id", h.UpdateCarousel)
    authGroup.DELETE("/carousels/:id", h.DeleteCarousel)
}

// ListTodos — copy org_id extraction pattern from internal/dashboard/handler.go lines 22-41
func (h *Handler) ListTodos(c *gin.Context) {
    orgID := c.GetInt64("org_id")
    // ... bind query, call service, response.Success or response.PageSuccess
}

// VerifyInviteToken — public endpoint, no org_id
// Copy error handling from invitation_handler.go lines 76-99
func (h *Handler) VerifyInviteToken(c *gin.Context) {
    token := c.Param("token")
    invite, err := h.svc.ValidateInviteToken(c.Request.Context(), token)
    if err != nil {
        if err == ErrInviteNotFound {
            response.Error(c, http.StatusNotFound, 90101, "邀请不存在")
            return
        }
        if err == ErrInviteExpired {
            response.Error(c, http.StatusGone, 90102, "邀请已过期")
            return
        }
        response.Error(c, http.StatusBadRequest, 90103, err.Error())
        return
    }
    response.Success(c, invite)
}
```

---

### `internal/todo/router.go` (router, request-response)

**Analog:** `internal/dashboard/router.go` (lines 1-18)

```go
// Copy exact pattern from internal/dashboard/router.go
package todo

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

func RegisterRouter(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, db *gorm.DB) {
    repo := NewTodoRepository(db)
    carouselRepo := NewCarouselRepository(db)
    svc := NewTodoService(repo, carouselRepo, cfg)
    handler := NewHandler(svc)

    // handler.RegisterRoutes handles both auth and public routes internally
    handler.RegisterRoutes(rg, authMiddleware)
}
```

---

### `internal/todo/excel.go` (utility, file-I/O)

**Analog:** `internal/socialinsurance/excel.go` (lines 1-260)

**excelize streaming pattern:**

```go
// Copy from internal/socialinsurance/excel.go lines 1-12, 47-90
import (
    "fmt"
    "net/url"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/xuri/excelize/v2"
)

// ExportTodos — follow same pattern as ExportSIRecordsWithDetails
// Headers: 事项名称/发起人/创建时间/截止时间/状态/是否限时/紧急程度
// Header style: blue (#4472C4) background, white bold font
// Number format: keep 2 decimal places for amounts
// Filename: 待办事项_YYYYMMDD_HHMMSS.xlsx
// Copy HTTP response pattern from socialinsurance/excel.go lines 247-261:
// c.Header("Content-Type", ...) + c.Header("Content-Disposition", ...) + c.Data(200, ...)
```

---

### `internal/todo/scheduler.go` (service, event-driven)

**Analog:** `internal/socialinsurance/scheduler.go` (lines 1-200)

**asynq + gocron + Redis locker pattern:**

```go
// Copy from internal/socialinsurance/scheduler.go lines 1-50, 131-199
import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/go-co-op/gocron/v2"
    "github.com/redis/go-redis/v9"
)

// Task type constants
const (
    TypeScanUrgencyStatus    = "todo:scan_urgency"
    TypeCheckCarouselActive  = "todo:check_carousel"
    TypeGenerateMonthlyTodo = "todo:generate_monthly" // 7种限时任务
)

// redisLocker + redisLock — copy exactly from scheduler.go lines 131-165
// newRedisLocker(rdb, "easyhr:todo:") — use "easyhr:todo:" prefix

// StartScheduler — copy pattern from scheduler.go lines 169-199
// gocron.WithLocation(time.FixedZone("CST", 8*3600))
// Daily 02:00: ScanUrgencyStatus
// Daily 08:00: CheckCarouselActivation
// June 15 annually: GenerateAnnualBaseAdjustments
func StartScheduler(rdb *redis.Client, svc *TodoService) (gocron.Scheduler, error) {
    opts := []gocron.SchedulerOption{
        gocron.WithLocation(time.FixedZone("CST", 8*3600)),
    }
    if rdb != nil {
        locker := newRedisLocker(rdb, "easyhr:todo:")
        opts = append(opts, gocron.WithDistributedLocker(locker))
    }
    // ...
}
```

---

### `internal/dashboard/service.go` (service, extension)

**Analog:** `internal/dashboard/service.go` (self-extension)

**Add GetTodoStats method** to existing service using errgroup pattern:

```go
// Add to existing DashboardService (copy from lines 28-130)
// New method: GetTodoStats(ctx, orgID, isTimeLimited bool) (*TodoStatsResult, error)
// Use errgroup for concurrent count queries (same as GetDashboard lines 42-129):
// g, ctx := errgroup.WithContext(ctx)
// g.Go(func() error { completed, err := s.repo.CountTodos(ctx, orgID, "completed", isTimeLimited); ... })
// g.Go(func() error { pending, err := s.repo.CountTodos(ctx, orgID, "pending", isTimeLimited); ... })
// Return: { Completed: completed, Pending: pending, Rate: computed as integer }
```

---

### `internal/dashboard/model.go` (model, extension)

**Analog:** `internal/dashboard/model.go` (self)

**Add fields to existing TodoItem + new result types:**

```go
// Add to existing TodoItem struct (copy from model.go lines 15-22):
// Deadline     string   `json:"deadline,omitempty"`  // Already exists — verify
// Add new fields:
type TodoItem struct {
    Type           TodoType     `json:"type"`
    Title          string       `json:"title"`
    Count          int          `json:"count"`
    Deadline       string       `json:"deadline,omitempty"`
    IsTimeLimited  bool         `json:"is_time_limited"`
    UrgencyStatus  string       `json:"urgency_status,omitempty"` // normal/overdue/expired
    Priority       int          `json:"priority"`
}

// New result types for GetTodoStats:
type TodoStatsResult struct {
    Completed    int `json:"completed"`
    Pending      int `json:"pending"`
    Total        int `json:"total"`
    CompletedRate int `json:"completed_rate"` // integer percentage
}
```

---

### `frontend/src/views/home/components/TodoRingChart.vue` (component, transform)

**Analog:** `frontend/src/views/employee/OrgChart.vue` (vue-echarts usage)

**vue-echarts PieChart pattern:**

```vue
<!-- Copy from RESEARCH.md Standard Stack / OrgChart.vue pattern -->
<script setup lang="ts">
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { PieChart } from 'echarts/charts'
import { TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

use([PieChart, TooltipComponent, LegendComponent, CanvasRenderer])

// Props: completed (number), pending (number), title (string)
const option = computed(() => {
  const total = props.completed + props.pending
  const rate = total === 0 ? 0 : Math.round((props.completed / total) * 100) // avoid /0
  return {
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    legend: { bottom: 0, type: 'plain' },
    series: [{
      type: 'pie',
      radius: ['40%', '70%'], // D-09-01
      label: { show: true, formatter: '{d}%' },
      data: [
        { value: props.completed, name: '已完成', itemStyle: { color: '#4F6EF7' } },
        { value: props.pending, name: '待办', itemStyle: { color: '#E8EEFF' } },
      ]
    }]
  }
})
</script>

<template>
  <v-chart :option="option" autoresize style="height: 200px" />
</template>
```

---

### `frontend/src/views/home/components/HomeCarousel.vue` (component, request-response)

**Analog:** `frontend/src/views/home/HomeView.vue` (self-extension)

**Element Plus Carousel + API call pattern:**

```vue
<!-- Follow HomeView.vue section style (lines 166-211) -->
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { fetchCarousels } from '@/api/carousel'
import type { CarouselItem } from '@/api/carousel'

const carousels = ref<CarouselItem[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    carousels.value = await fetchCarousels()
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <el-carousel :interval="4000" type="card" height="160px" indicator-position="outside">
    <el-carousel-item v-for="item in carousels" :key="item.id">
      <router-link :to="item.link_url || '/home'">
        <img :src="item.image_url" :alt="item.title" style="width:100%;height:100%;object-fit:cover;border-radius:8px" />
      </router-link>
    </el-carousel-item>
  </el-carousel>
</template>
```

---

### `frontend/src/views/home/HomeView.vue` (view, extension)

**Analog:** `frontend/src/views/home/HomeView.vue` (self)

**Extend with ring charts + carousel + shortcuts** — add new sections below page-header:

```vue
<!-- Add below page-header (line 15), before todo-section (line 18) -->
<!-- Ring charts section — two VCharts side by side -->
<div class="section ring-section">
  <TodoRingChart :completed="allCompleted" :pending="allPending" title="全部事项" />
  <TodoRingChart :completed="timeLimitedCompleted" :pending="timeLimitedPending" title="限时任务" />
</div>

<!-- Add HomeCarousel above shortcuts-section -->
<div class="section carousel-section">
  <HomeCarousel />
</div>

<!-- gridItems — add 3 entries (D-09-08) -->
<!-- Copy from HomeView.vue lines 139-146, append: -->
{ path: '/employee/create', label: '新入职', icon: UserAdd, color: '#fa8c16', bg: '#fff7e6' },
{ path: '/tool/salary', label: '调薪', icon: Money, color: '#52c41a', bg: '#f6ffed' }, // already exists
{ path: '/attendance/clock-live', label: '考勤', icon: Clock, color: '#1677ff', bg: '#e6f4ff' },
```

---

### `frontend/src/views/todo/TodoListView.vue` (view, request-response)

**Analog:** `frontend/src/views/home/HomeView.vue` (list patterns)

**Element Plus Table + Search + Pagination pattern:**

```vue
<!-- Follow HomeView.vue structure (lines 1-108) -->
<!-- Use el-table + el-pagination like other list views -->
<!-- Copy el-form search pattern from other views -->
<!-- Excel export: use raw <a> tag pointing to GET /todos/export endpoint -->
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { fetchTodos, terminateTodo, pinTodo } from '@/api/todo'
import { ElMessage, ElMessageBox } from 'element-plus'

const todos = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

// Search/filter state
const keyword = ref('')
const dateRange = ref<Date[]>([])
const statusFilter = ref('')

async function load() {
  const data = await fetchTodos({ page: page.value, pageSize: pageSize.value, keyword: keyword.value, start_date: dateRange.value[0], end_date: dateRange.value[1], status: statusFilter.value })
  todos.value = data.items
  total.value = data.total
}
</script>
```

---

### `frontend/src/views/todo/InviteFillPage.vue` (view, request-response)

**Analog:** `frontend/src/views/employee/RegisterPage.vue` (lines 1-185)

**Token verification + form submission pattern:**

```vue
<!-- Copy exact structure from RegisterPage.vue lines 1-185 -->
<!-- Three states: errorState / submitted / form -->
<!-- Script: token from route.params.token, error handling for 404/410 -->

<!-- Differences from RegisterPage: -->
<!-- 1. Form fields vary by todo.source_type (render dynamic fields) -->
<!-- 2. Submit calls todoApi.submitInvite(token, form.value) -->
<!-- 3. Success message: "信息已提交，感谢配合" -->
<!-- 4. Token verification error messages: -->
<!--    - 404: "邀请链接无效" -->
<!--    - 410: "该链接已过期" -->
<!--    - 400 (used/cancelled): "该邀请已被使用/已取消" -->
```

---

### `frontend/src/api/todo.ts` (API client, request-response)

**Analog:** `frontend/src/api/dashboard.ts` (lines 1-26)

**API module pattern with request interceptors:**

```typescript
// Copy from frontend/src/api/dashboard.ts lines 1-26
import request from '@/api/request'

// TodoItem interface — extend from dashboard.ts TodoItem
export interface TodoItem {
  type: string
  title: string
  count: number
  deadline?: string
  is_time_limited?: boolean
  urgency_status?: string // normal/overdue/expired
  priority: number
  is_pinned?: boolean
}

export interface TodoListResponse {
  items: TodoItem[]
  total: number
  page: number
  page_size: number
}

export interface InviteResponse {
  token: string
  invite_url: string
  expires_at: string
}

// List todos
export function fetchTodos(params: {
  page?: number
  page_size?: number
  keyword?: string
  start_date?: string
  end_date?: string
  status?: string
  is_time_limited?: boolean
}): Promise<TodoListResponse> {
  return request.get('/todos', { params }).then((res) => res.data)
}

// Invite todo
export function inviteTodo(todoId: number): Promise<InviteResponse> {
  return request.post(`/todos/${todoId}/invite`).then((res) => res.data)
}

// Terminate todo
export function terminateTodo(todoId: number): Promise<void> {
  return request.put(`/todos/${todoId}/terminate`).then((res) => res.data)
}

// Pin/unpin todo
export function pinTodo(todoId: number, pinned: boolean): Promise<void> {
  return request.put(`/todos/${todoId}/pin`, { pinned }).then((res) => res.data)
}

// Verify invite token (public, no auth)
export function verifyInviteToken(token: string): Promise<{ todo_id: number; title: string }> {
  return request.get(`/todos/invite/${token}/verify`, {
    headers: { Authorization: undefined }, // bypass auth interceptor
  }).then((res) => res.data)
}

// Submit invite (public)
export function submitInvite(token: string, data: Record<string, unknown>): Promise<void> {
  return request.post(`/todos/invite/${token}/submit`, data, {
    headers: { Authorization: undefined },
  }).then((res) => res.data)
}

// Export todos (returns blob)
export function exportTodos(params: Record<string, unknown>): Promise<void> {
  return request.get('/todos/export', { params, responseType: 'blob' }).then((res) => res.data)
}
```

---

### `frontend/src/api/carousel.ts` (API client, request-response)

**Analog:** `frontend/src/api/dashboard.ts` (lines 1-26)

```typescript
import request from '@/api/request'

export interface CarouselItem {
  id: number
  image_url: string
  link_url?: string
  sort_order: number
  active: boolean
  start_at?: string
  end_at?: string
}

export function fetchCarousels(): Promise<CarouselItem[]> {
  return request.get('/carousels').then((res) => res.data)
}

export function createCarousel(data: Partial<CarouselItem>): Promise<CarouselItem> {
  return request.post('/carousels', data).then((res) => res.data)
}

export function updateCarousel(id: number, data: Partial<CarouselItem>): Promise<void> {
  return request.put(`/carousels/${id}`, data).then((res) => res.data)
}

export function deleteCarousel(id: number): Promise<void> {
  return request.delete(`/carousels/${id}`).then((res) => res.data)
}
```

---

### `frontend/src/stores/dashboard.ts` (store, extension)

**Analog:** `frontend/src/stores/dashboard.ts` (self)

**Extend TodoItem interface + add ring chart data:**

```typescript
// Copy from frontend/src/stores/dashboard.ts lines 1-31
// Add new state for ring charts
interface TodoStats {
  allCompleted: number
  allPending: number
  timeLimitedCompleted: number
  timeLimitedPending: number
}

const todos = ref<TodoItem[]>([])
const stats = ref<TodoStats>({
  allCompleted: 0,
  allPending: 0,
  timeLimitedCompleted: 0,
  timeLimitedPending: 0,
})

// Extend load() to also fetch /dashboard/todo-stats
async function load() {
  loading.value = true
  try {
    const data = await fetchDashboard()
    todos.value = data.todos
    overview.value = data.overview
    // Fetch ring chart stats
    const statsData = await fetchTodoStats()
    stats.value = {
      allCompleted: statsData.all_completed,
      allPending: statsData.all_pending,
      timeLimitedCompleted: statsData.time_limited_completed,
      timeLimitedPending: statsData.time_limited_pending,
    }
  } finally {
    loading.value = false
  }
}

return { todos, overview, loading, overviewExpanded, stats, load, toggleOverview, removeTodo }
```

---

## Shared Patterns

### Authentication / Token Verification
**Source:** `internal/employee/invitation_handler.go` (lines 76-99)
**Apply to:** `internal/todo/handler.go` — VerifyInviteToken / SubmitInvite endpoints

```go
// Public endpoint (no auth middleware), but validate token carefully
token := c.Param("token")
invite, err := h.svc.ValidateInviteToken(c.Request.Context(), token)
if err != nil {
    if err == ErrInviteNotFound {
        response.Error(c, http.StatusNotFound, code, "message")
        return
    }
    if err == ErrInviteExpired {
        response.Error(c, http.StatusGone, code, "message")
        return
    }
    // ... other cases
}
response.Success(c, invite)
```

### org_id Extraction
**Source:** `internal/dashboard/handler.go` (lines 22-41)
**Apply to:** All authenticated todo handler methods

```go
orgIDVal, exists := c.Get("org_id")
if !exists {
    response.Error(c, http.StatusUnauthorized, 40100, "missing org_id in context")
    return
}
orgID, ok := orgIDVal.(int64)
if !ok {
    response.Error(c, http.StatusUnauthorized, 40100, "invalid org_id type")
    return
}
```

### Redis Distributed Lock
**Source:** `internal/socialinsurance/scheduler.go` (lines 131-165, 169-178)
**Apply to:** `internal/todo/scheduler.go` — urgency status scan + carousel activation

```go
// Use same redisLocker with "easyhr:todo:" prefix
if rdb != nil {
    locker := newRedisLocker(rdb, "easyhr:todo:")
    opts = append(opts, gocron.WithDistributedLocker(locker))
}
```

### CST Timezone
**Source:** `internal/socialinsurance/scheduler.go` (line 114) + RESEARCH.md Pitfall 2
**Apply to:** `internal/todo/scheduler.go` — all time comparisons

```go
cstZone := time.FixedZone("CST", 8*3600)
today := time.Now().In(cstZone)
```

### vue-echarts ECharts Registration
**Source:** `frontend/src/views/employee/OrgChart.vue` (verified existing)
**Apply to:** `frontend/src/views/home/components/TodoRingChart.vue`

```typescript
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { PieChart } from 'echarts/charts'
import { TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

use([PieChart, TooltipComponent, LegendComponent, CanvasRenderer])
```

### Token Generation
**Source:** `internal/employee/invitation_service.go` (lines 46-52)
**Apply to:** `internal/todo/service.go` — CreateInvite

```go
import "crypto/rand"

func generateToken() (string, error) {
    bytes := make([]byte, 32)
    if _, err := rand.Read(bytes); err != nil {
        return "", fmt.Errorf("generate token: %w", err)
    }
    return hex.EncodeToString(bytes), nil
}
```

### Excel Streaming Response
**Source:** `internal/socialinsurance/excel.go` (lines 247-261)
**Apply to:** `internal/todo/excel.go`

```go
buf, err := f.WriteToBuffer()
if err != nil {
    return fmt.Errorf("write excel buffer: %w", err)
}
c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(filename)))
c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
```

### RegisterPage Token Flow
**Source:** `frontend/src/views/employee/RegisterPage.vue` (lines 112-181)
**Apply to:** `frontend/src/views/todo/InviteFillPage.vue`

```typescript
const token = route.params.token as string  // line 120
const errorState = ref(false)                // line 123
const submitted = ref(false)                 // line 125
const form = ref({...})                      // line 128

async function loadDetail() {
  try {
    detail.value = await todoApi.verifyInviteToken(token)
  } catch (err) {
    errorState.value = true
    // Status 404 → "链接无效" / Status 410 → "链接已过期"
  }
}
```

### Dashboard API Request Interceptor
**Source:** `frontend/src/api/request.ts` (lines 22-31)
**Apply to:** All API modules — they all import and use this

### Multi-tenant GORM Scope
**Source:** `internal/dashboard/repository.go` (line 45)
**Apply to:** All todo repository queries

```go
r.db.Model(&TodoItem{}).Scopes(middleware.TenantScope(orgID))
```

---

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `frontend/src/views/home/components/TodoRingChart.vue` | component | transform | Uses vue-echarts PieChart, closest analog OrgChart.vue uses TreeChart (different chart type) but same echarts registration pattern confirmed |
| `frontend/src/api/todo.ts` | API client | request-response | Full module is new but follows dashboard.ts pattern exactly |
| `frontend/src/api/carousel.ts` | API client | request-response | Full module is new but follows dashboard.ts pattern exactly |

---

## Metadata

**Analog search scope:**
- Backend: `internal/dashboard/`, `internal/employee/`, `internal/socialinsurance/`
- Frontend: `frontend/src/views/home/`, `frontend/src/api/`, `frontend/src/stores/`, `frontend/src/views/employee/`

**Files scanned:** ~20 files
**Pattern extraction date:** 2026-04-19
