# Phase 9: 待办中心 - Research

**Researched:** 2026-04-19
**Domain:** 待办中心全栈实现（Go后端 + Vue3前端 + asynq定时任务）
**Confidence:** HIGH（基于代码库现场验证）

## Summary

Phase 9 为管理员构建统一的待办聚合入口，核心挑战在于：
- **数据模型**：扩展现有 `TodoItem` 为带时限的完整任务实体（deadline/is_time_limited/urgency_status），新增 `CarouselItem` 表
- **7种限时任务**：由各模块（employee/attendance/salary/socialinsurance/tax）触发创建，asynq 每日凌晨扫描状态
- **环形图**：复用 vue-echarts 已有栈，配置 PieChart 组件
- **协办邀请**：复用 Phase 5 `Invitation` Token 机制（crypto/rand 32-byte hex），无登录填写页
- **Excel 导出**：复用 excelize 已有栈

本阶段仅 H5 管理后台，后端 API 配合新增。

---

## User Constraints (from CONTEXT.md)

### Locked Decisions
- D-09-03: 不新建表，扩展 TodoItem 模型：deadline/is_time_limited/urgency_status
- D-09-04: 1-7天超期→超时（红色）；15天+超期→失效（灰色）；中间态→超期
- D-09-05: asynq 每日凌晨扫描限时任务更新 urgency_status；7种任务由各模块触发
- D-09-06: 新建 CarouselItem 表（id/org_id/image_url/link_url/sort_order/active/start_at/end_at）
- D-09-07: asynq 定时任务在 start_at/end_at 时间段内自动激活/停用
- D-09-08: 保留现有6个快捷入口，追加3个（新入职/调薪/考勤）
- D-09-09: 协办 Token 独立验证，无需登录，仅填写数据
- D-09-10: 复用 Phase 5 Token 机制，链接格式 `/todo/:id/invite?token=xxx`
- D-09-11: 终止任务保留数据，状态改为 terminated，不软删除

### Claude's Discretion
- 环形图具体配色（蓝色系 vs 品牌主色）
- 轮播图切换动画
- 快捷入口图标选择
- 限时任务触发时机（asynq cron vs 各模块直接创建）
- 协办填写页字段和布局
- TodoItem 列表分页大小

### Deferred Ideas (OUT OF SCOPE)
None

---

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| TODO-01 | 待办汇总列表，默认按时间由近到远排序 | TodoItem repository + HomeView 重构 |
| TODO-02 | 按关键字搜索待办事项 | repository WHERE LIKE 查询 |
| TODO-03 | 按时间段筛选（不超过60天） | repository WHERE date_range 查询 |
| TODO-04 | 邀请协办（转发给内部或外部人） | Invitation Token 复用 + 新 TodoInvite 表 |
| TODO-05 | 终止选中待办（状态变为"终止"） | TodoItem.status = terminated |
| TODO-06 | 待办置顶/取消置顶排序 | is_pinned + sort_order 字段 |
| TODO-07 | 显示事项名称/发起人/时间/状态 | TodoItem 字段完整化 |
| TODO-08 | Excel 格式下载导出 | excelize 复用 + ExportTodoList |
| TODO-09 | 首页展示1-3张轮播图，点击跳转详情 | CarouselItem + HomeView 扩展 |
| TODO-10 | 首页快捷入口（新入职/调薪/考勤/个税/社保） | gridItems 追加3项 |
| TODO-11 | 自动生成合同续签限时任务（到期前1个月） | contract_module 触发创建 |
| TODO-12 | 自动生成合同新签限时任务（入职30日内） | employee_module 触发创建 |
| TODO-13 | 自动生成个税申报限时任务（每月1-15日） | tax_module 触发创建 |
| TODO-14 | 自动生成社保公积金缴费限时任务（每月1-15日） | socialinsurance_module 触发创建 |
| TODO-15 | 自动生成社保公积金增减员限时任务（每月5-20日） | socialinsurance_module 触发创建 |
| TODO-16 | 自动生成年度社保基数调整限时任务（6月15日-7月15日） | asynq annual job 触发创建 |
| TODO-17 | 自动生成年度公积金基数调整限时任务（6月15日-7月15日） | asynq annual job 触发创建 |
| TODO-18 | 限时任务显示剩余时间，超时1-7日→超时，15日+→失效 | urgency_status 状态机 |
| TODO-19 | 全部事项完成率环形图 | ECharts PieChart + GetTodoStats |
| TODO-20 | 限时任务完成率环形图 | ECharts PieChart + GetTodoStats (is_time_limited=true) |

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| 待办列表搜索/筛选/分页 | API/Backend | Frontend Server | Go repository WHERE 查询，精确控制 |
| urgency_status 状态机 | API/Backend (asynq) | — | 每日凌晨 cron，Gorm 更新，无前端交互 |
| 轮播图激活/停用 | API/Backend (asynq) | — | 定时器控制，Redis 分布式锁 |
| 7种限时任务创建 | 各业务模块 | API/Backend | employee/salary/socialinsurance/tax 在关键操作时创建 |
| 环形图统计聚合 | API/Backend | Frontend Server | errgroup 并发聚合，Gin 返回 JSON |
| 协办邀请 Token 验证 | Frontend Server (SSR) | API/Backend | 链接打开后前端请求验证，后端查 Token |
| Excel 导出 | API/Backend | — | Go excelize 生成，Streaming 下载 |
| CarouselItem 图片上传 | CDN/Static (OSS) | API/Backend | 签名 URL 直传，CarouselItem 存 URL |
| 快捷入口展示 | Frontend Server | — | HomeView.vue 静态配置 |
| 环形图渲染 | Frontend Server | Browser | vue-echarts PieChart，Browser 渲染 |

---

## Standard Stack

### Core Dependencies

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `echarts` | 6.0.0 | 环形图数据可视化 | 项目已有，前端已配置 |
| `vue-echarts` | 8.0.1 | Vue 3 ECharts 封装 | 项目已有，`<v-chart>` 组件模式 |
| `excelize` | (已安装) | Excel 导出 | 项目已有，Phase 8 已验证 |
| `asynq` | (已安装) | 定时任务队列 | 项目已有，Phase 8 scheduler.go 模式 |
| `go-redis` | (已安装) | Redis 客户端 | 项目已有，asynq 依赖 |
| `gocron` | (已安装) | 定时任务调度 | 项目已有，Phase 8 socialinsurance/scheduler.go |
| `crypto/rand` | 标准库 | Token 生成 | 项目已有，Phase 5 invitation_model.go 验证 |

### Frontend Pattern: vue-echarts PieChart

**验证来源:** `frontend/src/views/employee/OrgChart.vue` (项目已有使用)

```typescript
// Source: frontend/src/views/employee/OrgChart.vue (verified existing pattern)
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { PieChart } from 'echarts/charts'
import { TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

use([PieChart, TooltipComponent, LegendComponent, CanvasRenderer])

// 使用方式
<v-chart :option="option" autoresize style="height: 280px" />

// 环形图配置（D-09-01: radius=['40%','70%']）
const option = {
  tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
  legend: { bottom: 0, type: 'plain' },
  series: [{
    type: 'pie',
    radius: ['40%', '70%'],
    label: { show: true, formatter: '{d}%' },
    data: [
      { value: completed, name: '已完成', itemStyle: { color: '#4F6EF7' } },
      { value: pending, name: '待办', itemStyle: { color: '#E8EEFF' } },
    ]
  }]
}
```

**安装确认:** `npm list echarts vue-echarts` → echarts@6.0.0, vue-echarts@8.0.1（项目已有）

---

## Architecture Patterns

### System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Browser / H5 Admin                              │
│  ┌──────────────────┐   ┌────────────────────┐   ┌──────────────────────┐   │
│  │  HomeView.vue    │   │  TodoListView.vue   │   │  InvitePage.vue      │   │
│  │  (环形图+轮播图+ │   │  (搜索/筛选/导出)   │   │  (Token验证+填写)    │   │
│  │   快捷入口+待办) │   │                     │   │                      │   │
│  └────────┬─────────┘   └─────────┬───────────┘   └──────────┬───────────┘   │
└───────────┼───────────────────────┼─────────────────────────┼───────────────┘
            │ GET /dashboard         │ GET /todos              │ POST /todos/:id/submit
            │ GET /carousels         │ GET /todos/export       │
            │                        │ POST /todos/:id/invite  │
            │                        │ PUT  /todos/:id/terminate│
┌───────────┴───────────────────────┴─────────────────────────┴───────────────┐
│                              Go API Server (Gin)                            │
│  ┌──────────────────┐   ┌────────────────────┐   ┌──────────────────────┐   │
│  │ DashboardHandler │   │  TodoHandler        │   │  CarouselHandler     │   │
│  │  GetDashboard()  │   │  ListTodos()        │   │  ListCarousels()     │   │
│  │  GetTodoStats()  │   │  InviteTodo()       │   │  CreateCarousel()     │   │
│  └────────┬─────────┘   │  TerminateTodo()    │   │  UpdateCarousel()     │   │
│           │              │  ExportTodos()       │   │  DeleteCarousel()     │   │
│           │              └─────────┬───────────┘   └──────────────────────┘   │
│           │                        │                                        │
│  ┌────────┴────────────────────────┴────────────────────────────────────┐   │
│  │                         DashboardService                               │   │
│  │  GetDashboard()   — errgroup 并发聚合各模块统计                        │   │
│  │  GetTodoStats()   — 完成率环形图数据（全部 / 限时任务）                 │   │
│  └────────┬─────────────────────────────────────────────────────────────┘   │
│           │                                                                  │
│  ┌────────┴─────────────────────────────────────────────────────────────┐   │
│  │                    TodoRepository / CarouselRepository (GORM)        │   │
│  │  WHERE org_id=? AND status IN (?)  — 列表查询                         │   │
│  │  UPDATE urgency_status — 每日凌晨批量更新                              │   │
│  └───────────────────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────────────────────┘
            │
            │ asynq tasks
┌───────────┴──────────────────────────────────────────────────────────────┐
│                              Redis + asynq                                  │
│  ┌────────────────────────────┐  ┌──────────────────────────────────────┐  │
│  │ gocron scheduler           │  │ asynq worker                         │  │
│  │  每天凌晨02:00             │  │  ProcessUrgencyStatus() — 更新超期   │  │
│  │  → ScanUrgencyStatus       │  │  ProcessCarouselActivation() — 激活  │  │
│  │                             │  │  Process7TypeTaskGeneration() — 生成  │  │
│  │  每天08:00 CST             │  │  (可选：7种任务由各模块触发创建)     │  │
│  │  → CheckCarouselActivation │  │                                      │  │
│  └────────────────────────────┘  └──────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────────────┘
```

### Recommended Project Structure

```
frontend/src/
├── views/home/
│   ├── HomeView.vue              # 环形图+轮播图+快捷入口+待办卡片
│   └── components/
│       ├── TodoRingChart.vue     # 完成率环形图（复用 vue-echarts）
│       └── HomeCarousel.vue       # 轮播图组件
├── views/todo/
│   ├── TodoListView.vue          # 待办完整列表（搜索/筛选/分页/导出）
│   ├── TodoDetailDrawer.vue      # 待办详情抽屉
│   └── InviteFillPage.vue        # 协办填写页（/todo/:id/invite?token=xxx）
├── api/todo.ts                   # 待办相关 API（list/export/invite/terminate）
├── api/carousel.ts              # 轮播图 API（CRUD）
└── stores/todo.ts                # TodoStore（待办列表状态）

internal/
├── todo/
│   ├── model.go                  # TodoItem 模型（扩展字段）
│   ├── carousel_model.go         # CarouselItem 模型
│   ├── todo_invite_model.go      # TodoInvite（协办邀请，Token表）
│   ├── repository.go             # TodoRepository（列表/搜索/筛选/导出）
│   ├── carousel_repository.go     # CarouselRepository
│   ├── service.go                # TodoService（业务逻辑+状态机）
│   ├── handler.go                # Gin Handler（API 端点）
│   ├── router.go                 # 路由注册
│   ├── excel.go                  # Excel 导出
│   └── scheduler.go              # asynq/gocron 定时任务
└── dashboard/
    └── service.go                # 扩展 GetTodoStats() 方法
```

### Pattern 1: urgency_status 状态机

**What:** 限时任务 urgency_status 状态自动流转：normal → overdue → expired
**When to use:** 每日凌晨 asynq cron 扫描所有 is_time_limited=true 且未完成的 TodoItem
**Example:**
```go
// Source: [ASSUMED - 基于 D-09-04 规则推导]
// 状态判定伪代码（Go实现）
func ComputeUrgencyStatus(deadline time.Time, currentStatus string) string {
    if currentStatus == "completed" || currentStatus == "terminated" {
        return currentStatus // 已完成/已终止不改变
    }

    now := time.Now()
    daysUntil := int(deadline.Sub(now).Hours() / 24)

    if daysUntil < -15 {
        return "expired" // 超期15天以上 → 失效
    } else if daysUntil < 0 || (daysUntil >= 0 && daysUntil <= 7) {
        return "overdue" // 已超期 或 剩余<=7天 → 超时警告
    } else if daysUntil <= 30 {
        return "normal" // 正常（有剩余时间）
    }
    return "normal"
}
```

**状态流转规则（D-09-04）:**

| 条件 | 状态 |
|------|------|
| deadline - now > 7 天 | `normal`（正常） |
| deadline - now ∈ [0, 7] 天 | `overdue`（超时警告，红色） |
| deadline - now ∈ [-15, 0) 天 | `overdue`（超时警告，红色） |
| deadline - now < -15 天 | `expired`（已失效，灰色） |
| status = completed | `completed`（已完成） |
| status = terminated | `terminated`（已终止） |

### Pattern 2: 7种限时任务的创建时机

**What:** 各模块在关键业务操作时触发创建 TodoItem，而非 asynq 轮询
**When to use:** 在员工入职、合同签署、薪资发放等业务操作时直接创建
**Example:**
```go
// 合同模块 - 合同签署时（D-09-05: 合同新签，入职30日内截止）
func (s *ContractService) SignContract(ctx context.Context, employeeID int64, contract *Contract) error {
    // 原有逻辑...
    if err := s.repo.Create(contract); err != nil {
        return err
    }
    // 新增：创建合同新签限时任务
    deadline := contract.StartDate.AddDate(0, 0, 30)
    todo := &todo.TodoItem{
        Type:          todo.TodoTypeContractNew,
        Title:         "合同新签：请在入职30日内完成劳动合同签署",
        EmployeeID:    employeeID,
        Deadline:      deadline,
        IsTimeLimited: true,
        Status:        "pending",
        SourceType:    "contract",
        SourceID:      contract.ID,
    }
    return s.todoRepo.Create(ctx, todo)
}

// 个税模块 - 每月1日生成申报任务（D-09-05: 每月1-15日，发薪日截止）
func (s *Service) CheckDeclarationReminders() {
    // 原有逻辑...

    // 新增：创建个税申报限时任务（15日截止）
    cst := time.FixedZone("CST", 8*3600)
    today := time.Now().In(cst)
    if today.Day() == 1 {
        for _, orgID := range activeOrgIDs {
            deadline := time.Date(today.Year(), today.Month(), 15, 23, 59, 59, 0, cst)
            todo := &todo.TodoItem{
                Type:          todo.TodoTypeTaxDeclaration,
                Title:         fmt.Sprintf("%d月个税申报，请于15日前完成", today.Month()),
                OrgID:         orgID,
                Deadline:      deadline,
                IsTimeLimited: true,
                Status:        "pending",
                SourceType:    "tax",
                SourceID:      0,
            }
            s.todoRepo.Create(ctx, todo)
        }
    }
}
```

**7种限时任务汇总:**

| 任务类型 | 触发时机 | Deadline | SourceModule |
|----------|----------|----------|--------------|
| 合同新签 (contract_new) | 员工入职/合同签署时 | 入职日期+30天 | employee |
| 合同续签 (contract_renew) | 合同到期前1个月 | 合同到期日 | employee |
| 个税申报 (tax_declaration) | 每月1日 | 每月15日 | tax |
| 社保缴费 (si_payment) | 每月1日 | 每月15日 | socialinsurance |
| 社保增减员 (si_change) | 每月5日 | 每月20日 | socialinsurance |
| 年度社保基数调整 (si_base_annual) | 每年6月15日 | 每年7月15日 | socialinsurance |
| 年度公积金基数调整 (fund_base_annual) | 每年6月15日 | 每年7月15日 | socialinsurance |

### Pattern 3: 协办邀请 Token 机制（复用 Phase 5）

**What:** 复用 `employee/invitation_model.go` 的 Token 生成和验证逻辑
**When to use:** 协作者无需登录即可填写待办关联数据
**Example:**
```go
// Source: internal/employee/invitation_model.go (verified existing pattern)
// 复用 generateToken 逻辑（crypto/rand 32-byte hex）
func generateInviteToken() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return hex.EncodeToString(b), nil
}

// TodoInvite 表（独立于 TodoItem）
type TodoInvite struct {
    ID        int64     `gorm:"primaryKey"`
    OrgID     int64     `gorm:"index"`
    TodoID    int64     `gorm:"index;not null"`
    Token     string    `gorm:"uniqueIndex;type:varchar(64)"`
    Status    string    `gorm:"default:pending"` // pending/used/expired
    ExpiresAt time.Time `gorm:"not null"`
    UsedAt    *time.Time
    CreatedBy int64     `gorm:"not null"`
}

const InviteExpiryDuration = 7 * 24 * time.Hour // 7天有效期
```

**前端协办填写页（参考 RegisterPage.vue 模式）:**
```vue
<!-- Source: frontend/src/views/employee/RegisterPage.vue (verified existing pattern) -->
<!-- 复用：token 提取 → API 验证 → 表单填写 → 提交 → 成功/失败状态 -->
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
// 协办填写页与 RegisterPage.vue 完全相同模式
// 仅字段不同（由各待办类型决定具体字段）
</script>
```

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Token 生成 | 手写随机数 | `crypto/rand.Read` | Phase 5 已验证，密码学安全 |
| Excel 导出 | 手写 CSV | `excelize` | Phase 8 socialinsurance/excel.go 已验证 |
| ECharts 图表 | 原生 canvas | `vue-echarts` | 项目已有栈，OrgChart.vue 已验证 |
| 定时任务 | 轮询 + time.Sleep | `asynq` + `gocron` | Phase 8 scheduler.go 已验证 |
| 协办邀请状态 | 自建状态管理 | GORM model + 常量 | 与现有 Approval/Invitation 模式一致 |

---

## Common Pitfalls

### Pitfall 1: urgency_status 重复计算（多次扫描）
**What goes wrong:** asynq cron 重复更新已完成的任务状态
**Why it happens:** 没有在 UPDATE WHERE 中排除已完成/已终止状态
**How to avoid:**
```go
// 正确：WHERE 条件必须排除已完成态
db.Model(&TodoItem{}).
    Where("is_time_limited = ? AND status NOT IN ?",
        true, []string{"completed", "terminated", "expired"}).
    Updates(map[string]interface{}{"urgency_status": newStatus})
```

### Pitfall 2: 轮播图时区问题（跨天激活）
**What goes wrong:** 轮播图 start_at/end_at 使用 UTC，管理员看到的时间与实际不符
**Why it happens:** 数据库存 UTC，前端显示 CST
**How to avoid:** 所有时间字段统一使用 CST（time.FixedZone("CST", 8*3600)），与 Phase 8 一致

### Pitfall 3: 环形图空数据除零
**What goes wrong:** 全部完成或全部待办时，completed=0 导致除零
**Why it happens:** 百分比计算 (completed/(completed+pending))*100
**How to avoid:** 使用整数计算，0/0 显示 0% 而非 NaN：
```typescript
// 正确
const total = completed + pending
const percent = total === 0 ? 0 : Math.round((completed / total) * 100)
```

### Pitfall 4: 轮播图 OSS 上传与 URL 混淆
**What goes wrong:** 上传图片后 URL 没更新或重复上传
**Why it happens:** CarouselItem.image_url 存的是最终 URL，但上传流程与现有 upload API 不一致
**How to avoid:** 参考现有 upload-to-oss 模式（如有）；新实现时使用签名 URL 直传

### Pitfall 5: asynq cron 分布式重复执行
**What goes wrong:** 多实例部署时，cron 同时触发同一任务
**Why it happens:** 未使用 Redis 分布式锁
**How to avoid:** 复用 Phase 8 socialinsurance/scheduler.go 的 `redisLocker` 模式：
```go
// Source: internal/socialinsurance/scheduler.go (verified existing pattern)
if rdb != nil {
    locker := newRedisLocker(rdb, "easyhr:todo:")
    opts = append(opts, gocron.WithDistributedLocker(locker))
}
```

### Pitfall 6: 限时任务幂等创建（重复创建同一任务）
**What goes wrong:** 每月1日多次触发个税申报任务，创建多条记录
**Why it happens:** 各模块在 CheckDeclarationReminders 中未去重
**How to avoid:** 在 Create 之前查询是否已存在（year_month + type + org_id 唯一索引）：
```go
func (r *TodoRepository) ExistsBySource(orgID int64, sourceType, sourceID string) bool {
    var count int64
    r.db.Model(&TodoItem{}).
        Where("org_id = ? AND source_type = ? AND source_id = ?", orgID, sourceType, sourceID).
        Count(&count)
    return count > 0
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Dashboard TodoItem 仅 count 聚合 | 完整 TodoItem 表，支持搜索/筛选/导出 | Phase 9 新建 | 从"快捷预览"升级为"管理功能" |
| 限时任务独立表 | 扩展 TodoItem 字段（D-09-03） | Phase 9 | 统一数据模型，避免两张表关联查询 |
| 无轮播图 | CarouselItem 表 + asynq 激活 | Phase 9 新建 | 管理员自主配置首页内容 |
| 协办需要登录 | Token 链接无登录（D-09-09） | Phase 9 | 外部人员无需注册即可填写 |

---

## Assumptions Log

> 以下假设基于代码库分析和训练知识，**需要用户确认**

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | vue-echarts 8.0.1 + echarts 6.0.0 的组合正常工作（实测已安装） | Standard Stack | 低 — 已有 OrgChart.vue 使用此组合 |
| A2 | 协办填写页字段由各待办类型决定，具体字段待定 | 7种限时任务 | 中 — 前端页面布局和 API 字段数量受影响 |
| A3 | TodoItem 分页大小使用 20 条/页（标准后台列表） | Claude's Discretion | 中 — 影响用户体验，20/50/100 均可 |
| A4 | CarouselItem 表不与现有图片上传 API 集成，新实现上传逻辑 | CarouselItem | 中 — 可能需要复用现有 OSS 上传模式 |
| A5 | asynq 每日凌晨 02:00 扫描 urgency_status（与 Phase 8 一致） | asynq scheduler | 低 — 仅影响 cron 时间，逻辑正确即可 |
| A6 | 快捷入口追加3个路径为 /employee/create、/tool/salary、/attendance/clock-live | 快捷入口 | 低 — 路径与 Phase 5/7/6 对应 |
| A7 | Excel 导出字段：事项名称/发起人/创建时间/截止时间/状态/是否限时 | Excel 导出 | 中 — 具体字段需与用户确认 |

---

## Open Questions

1. **协办填写页字段范围**
   - What we know: D-09-09 描述"补充员工信息/提交假勤申请"
   - What's unclear: 具体填写哪些字段？是否需要根据 TodoItem.type 动态渲染？
   - Recommendation: 协办页为通用填写页，根据 todo.source_type 渲染不同表单；字段范围在 PLAN.md 中定义

2. **环形图配色方案**
   - What we know: D-09-01 使用蓝色系，brand color #4F6EF7
   - What's unclear: 超时/失效状态用什么颜色？灰色 #8c8c8c？
   - Recommendation: 使用 Element Plus 语义色（正常#4F6EF7蓝色，超时#FF5630红色，失效#8c8c8c灰色）

3. **轮播图上传方式**
   - What we know: D-09-06 图片存阿里云 OSS
   - What's unclear: 现有项目是否有统一的 OSS 上传 API？
   - Recommendation: 检查 frontend/src/api/ 下是否有 upload 相关 API；如有则复用；如无则新建 `/upload/image` handler

4. **7种限时任务的完成触发**
   - What we know: 各模块触发创建，但何时标记为"已完成"？
   - What's unclear: 管理员完成对应业务操作时自动完成，还是手动点击完成？
   - Recommendation: 与业务操作联动（签署合同→自动完成，申报个税→管理员手动标记）

5. **前端 TodoList 是否为独立页面还是 HomeView 扩展**
   - What we know: TODO-01 "待办汇总列表"，HomeView 有精简待办卡片
   - What's unclear: 是否需要 `/todo` 独立路由页面，还是在 HomeView 内展开更多行？
   - Recommendation: 独立 `/todo` 路由页面，HomeView 保留精简待办卡片（≤3条，点击展开到完整列表）

---

## Environment Availability

Step 2.6: SKIPPED (no external dependencies beyond project codebase — asynq/Redis already configured, vue-echarts already installed, excelize already installed)

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Backend | `go test` + `testify` (已配置，Phase 8 已验证) |
| Frontend | `vitest` (已配置，package.json scripts.test:unit: vitest run) |
| Quick run | `go test ./internal/todo/... -v` / `npm run test:unit` |
| Full suite | `go test ./... -race -cover` / `npm run test:unit` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| TODO-01 | 待办汇总列表默认按时间排序 | unit | `go test ./internal/todo/... -run TestListTodos` | NEW |
| TODO-02 | 关键字搜索返回匹配项 | unit | `go test ./internal/todo/... -run TestSearchTodos` | NEW |
| TODO-03 | 时间段筛选（60天限制） | unit | `go test ./internal/todo/... -run TestFilterByDateRange` | NEW |
| TODO-04 | 协办邀请生成 Token 并可访问 | unit | `go test ./internal/todo/... -run TestInviteTodo` | NEW |
| TODO-05 | 终止任务状态变为 terminated | unit | `go test ./internal/todo/... -run TestTerminateTodo` | NEW |
| TODO-06 | 置顶/取消置顶排序 | unit | `go test ./internal/todo/... -run TestPinTodo` | NEW |
| TODO-07 | 列表显示名称/发起人/时间/状态 | integration | `go test ./internal/todo/... -run TestTodoListFields` | NEW |
| TODO-08 | Excel 导出返回 xlsx 文件 | integration | `go test ./internal/todo/... -run TestExportTodos` | NEW |
| TODO-09 | 轮播图列表 API（start_at/end_at 过滤） | unit | `go test ./internal/todo/... -run TestListCarousels` | NEW |
| TODO-10 | 快捷入口配置在 HomeView.vue | smoke | 查看 gridItems 数组 | 需修改 HomeView.vue |
| TODO-11 | 合同新签限时任务（30日截止） | unit | `go test ./internal/employee/... -run TestContractTriggersTodo` | NEW |
| TODO-12 | 合同续签限时任务（到期前1个月） | unit | `go test ./internal/employee/... -run TestContractRenewTriggersTodo` | NEW |
| TODO-13 | 个税申报限时任务（每月1-15日） | unit | `go test ./internal/tax/... -run TestTaxDeclarationTodo` | NEW |
| TODO-14 | 社保缴费限时任务（每月1-15日） | unit | `go test ./internal/socialinsurance/... -run TestSIPaymentTodo` | NEW |
| TODO-15 | 社保增减员限时任务（每月5-20日） | unit | `go test ./internal/socialinsurance/... -run TestSIChangeTodo` | NEW |
| TODO-16 | 年度社保基数调整任务（6-7月） | unit | `go test ./internal/socialinsurance/... -run TestSIAnnualBaseTodo` | NEW |
| TODO-17 | 年度公积金基数调整任务（6-7月） | unit | `go test ./internal/socialinsurance/... -run TestFundAnnualBaseTodo` | NEW |
| TODO-18 | urgency_status 超时/失效状态计算 | unit | `go test ./internal/todo/... -run TestUrgencyStatus` | NEW |
| TODO-19 | 全部事项完成率环形图 API | unit | `go test ./internal/dashboard/... -run TestGetTodoStats` | 需扩展 dashboard/service.go |
| TODO-20 | 限时任务完成率环形图 API | unit | `go test ./internal/dashboard/... -run TestGetTimeLimitedStats` | 需扩展 dashboard/service.go |

### Wave 0 Gaps
- [ ] `internal/todo/` 目录不存在 — 需要创建 todo model/repository/service/handler/excel/scheduler
- [ ] `internal/todo/carousel_model.go` — CarouselItem 模型
- [ ] `internal/todo/todo_invite_model.go` — TodoInvite（协办邀请）模型
- [ ] `frontend/src/views/todo/` — 待办完整列表页（TodoListView.vue）
- [ ] `frontend/src/api/todo.ts` — 待办 API（list/export/invite/terminate）
- [ ] `frontend/src/api/carousel.ts` — 轮播图 API
- [ ] `frontend/src/stores/todo.ts` — TodoStore（可选，Pinia）
- [ ] `internal/dashboard/model.go` — 扩展 TodoItem 字段
- [ ] `internal/dashboard/repository.go` — 扩展 GetTodoStats 方法

---

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | NO | 协办邀请无需登录（Token 验证替代） |
| V3 Session Management | NO | Token 一次性，不建立会话 |
| V4 Access Control | YES | org_id 租户隔离，Token 验证确保只能访问对应 Todo |
| V5 Input Validation | YES | go-playground/validator（项目已有）+ Zod（前端已有） |
| V6 Cryptography | YES | Token 用 crypto/rand（Phase 5 已验证） |

### Known Threat Patterns

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Token 枚举/暴力破解 | Information Disclosure | crypto/rand 32字节熵，7天过期 |
| 跨租户访问（Token 伪造） | Information Disclosure | Token 验证时同时校验 org_id + todo_id |
| 恶意图片上传（轮播图） | Tampering | OSS 签名 URL 限制 MIME type + 文件大小 |
| Excel 注入（导出） | Injection | 不使用 CSV，避免公式注入；excelize 输出 xlsx |
| 搜索 XSS（待办标题） | XSS | Gin HTML 转义，Element Plus 自动防护 |

---

## Sources

### Primary (HIGH confidence)
- `internal/dashboard/service.go` — 现有 DashboardService 聚合模式，errgroup 并发
- `internal/dashboard/model.go` — 现有 TodoItem 字段，扩展依据
- `internal/dashboard/repository.go` — 现有 repository 查询模式
- `frontend/src/views/home/HomeView.vue` — 首页现有结构，扩展依据
- `frontend/src/views/employee/OrgChart.vue` — vue-echarts 使用验证（TreeChart）
- `frontend/src/views/employee/RegisterPage.vue` — Token 验证+填写页模式
- `internal/employee/invitation_model.go` — Token 生成逻辑（crypto/rand）
- `internal/socialinsurance/scheduler.go` — asynq/gocron 定时任务模式（含 Redis 分布式锁）
- `internal/socialinsurance/excel.go` — excelize 导出模式
- `internal/employee/contract_model.go` — Contract.EndDate 字段（限时任务依据）
- `internal/attendance/model.go` — AttendanceRule/Approval 现有结构
- `frontend/package.json` — echarts@6.0.0 + vue-echarts@8.0.1（npm verified）

### Secondary (MEDIUM confidence)
- ECharts 环形图配置：[vue-echarts 官方文档](https://github.com/ecomfe/vue-echarts) — 验证 `<v-chart>` 用法
- excelize API: [xuri/excelify GitHub](https://github.com/xuri/excelify) — Phase 8 实际使用

### Tertiary (LOW confidence)
- 7种限时任务触发时机：D-09-05 描述为"各模块触发"，具体触发点基于业务逻辑推导，需要实现时验证
- CarouselItem OSS 上传模式：项目中未找到现有 upload API，需新建或复用
