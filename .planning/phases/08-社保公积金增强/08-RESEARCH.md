# Phase 8: 社保公积金增强 - Research

**Researched:** 2026-04-19
**Domain:** 社保公积金增强 - 数据看板、增减员弹窗优化、缴费渠道管理、状态自动流转、欠缴横幅、五险分项、Excel导出
**Confidence:** HIGH

## Summary

Phase 8 在现有 `internal/socialinsurance` 模块基础上扩展：新增 `SIMonthlyPayment` 月度缴费表解决"谁欠缴/何时欠缴"的问题，asynq 定时任务驱动状态自动流转（正常→待缴→欠缴），前端重构参保操作 Tab（数据看板 + 优化增减员弹窗 + 5状态标签 + 欠缴红色横幅）。核心技术挑战是 D-SI-01/D-SI-03 决策已锁定状态模型和流转规则，实现时需严格区分参保生命周期状态（pending/active/stopped）与月度缴费状态（normal/pending/overdue/transferred/not_transferred）。

**Primary recommendation:** 后端优先实现 `SIMonthlyPayment` 表 + Dashboard 聚合 API，再推进 asynq 定时任务；前端以 SalaryDashboard.vue 为模板复用 4 卡片样式，确保与薪资看板视觉一致。

---

## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-SI-01:** 新建 `SIMonthlyPayment` 表（employee_id + year_month + status + payment_channel + company_amount + personal_amount + total_amount），status 月度独立追踪，不与参保生命周期 conflate
- **D-SI-02:** asynq 定时任务每月生成下月记录 + 删除超过 24 个月的记录
- **D-SI-03:** asynq 定时任务每天凌晨触发状态流转：≥26日未缴→overdue，<26日已确认→normal
- **D-SI-04:** 代理缴费 webhook 更新扣缴状态
- **D-SI-05:** 4 张纯数字卡片（应缴总额/单位部分/个人部分/欠缴金额），与薪资看板风格一致，不加月度筛选器
- **D-SI-06:** 环比上月计算：仅统计 confirmed 状态月份；上月无数据显示"—"
- **D-SI-07:** 增员弹窗（EnrollDialog）：姓名搜索 + 起始月份默认当月可选近3个月 + 社保基数 + 公积金比例和基数
- **D-SI-08:** 减员弹窗（StopDialog）：姓名搜索 + 终止月份默认当月且不可早于当月 + 原因三选一 + 转出/封存日期 + 生效规则提示
- **D-SI-09:** 增减员 INSERT ONLY：ChangeHistory 追加记录
- **D-SI-10:** 红色横幅展示最大欠缴项（员工+城市+欠缴月+金额），下方滚动展示所有未处理欠缴；行内标红背景
- **D-SI-11:** 5 种状态标签（正常-绿/待缴-黄/欠缴-红/已转出-灰/未转出-蓝）+ 缴费渠道列 + 行展开详情
- **D-SI-12:** 五险分项弹窗：养老/医疗/失业/工伤/生育/公积金各自展示单位+个人金额，底部合计 + 其他缴费行
- **D-SI-13:** 复用 excelize 导出模式（与 SalaryList.vue 导出对话框风格一致）

### Claude's Discretion

- 横幅动画效果（静止/滚动/可关闭）
- 五险分项弹窗列宽和数字格式（千分位分隔符）
- 政策通知来源（手动录入 vs 第三方 API）
- 增员公积金单独页签还是同表单内切换
- 欠缴提醒超过多少条时横幅显示策略
- asynq 定时任务具体 cron 表达式（每天凌晨几点）
- SI-15 自主缴费跳转至哪个外部页面

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| 社保数据看板 | API/Backend | Browser | DashboardService 聚合查询，Frontend 纯展示 |
| 增减员弹窗 | Browser/Client | API/Backend | 前端姓名搜索调用 employeeApi，后端校验并写入记录 |
| 状态自动流转 | API/Backend | CDN/Static | asynq 定时任务在服务器端运行，不涉及前端 |
| 欠缴横幅 | Browser/Client | — | 前端实时读取欠缴数据渲染横幅 |
| 五险分项弹窗 | Browser/Client | API/Backend | 前端行展开触发 API 查询明细 |
| Excel 导出 | API/Backend | Browser | 后端 excelize 生成文件流传回前端下载 |
| 代理缴费 webhook | API/Backend | — | 外部回调写入 asynq 队列处理状态更新 |

---

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| asynq | v0.26.0 | 定时任务 + 批量操作队列 | 项目已安装（go.mod）；D-SI-02/D-SI-03 决策锁定；代理缴费 webhook 异步处理 |
| go-redis | v9.18.0 | Redis 客户端 | 项目已安装；asynq 依赖 Redis；scheduler.go 分布式锁已用 gocron+v2.20.0 |
| gocron | v2.20.0 | 定时任务调度 | 项目已安装；scheduler.go 已使用；但 D-SI-03 决策用 asynq 做状态流转（与 scheduler.go 的 gocron 定时任务互补：gocron 驱动 CheckPaymentDueReminders，asynq 驱动月度记录生成+状态流转）|
| excelize | v2.10.1 | Excel 读写 | 项目已安装；D-SI-13 决策锁定复用模式 |
| shopspring/decimal | v1.4.0 | 精确金额计算 | 项目已安装；金额字段必须用 decimal.Decimal，禁止 float64 |
| golang-jwt | v5.3.1 | JWT 认证 | 项目已安装 |
| go-playground/validator | v10.30.2 | 参数校验 | 项目已安装 |
| vue-echarts | v8.0.1 | 图表 | v1.3 已决策引入 |
| @vueuse/core | v14.2.x | 组合式工具 | v1.3 已决策引入 |
| Element Plus | 2.13.6 | UI 组件库 | 项目已安装 |

**Installation (if needed):**
```bash
cd frontend && npm install vue-echarts echarts @vueuse/core
```

**Version verification:**
- asynq: `go list -m github.com/hibiken/asynq` → v0.26.0 [VERIFIED: go.sum]
- gocron: `go list -m github.com/go-co-op/gocron/v2` → v2.20.0 [VERIFIED: go.mod]
- shopspring/decimal: v1.4.0 [VERIFIED: go.mod]

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| asynq | gocron 单机定时任务 | gocron 已用于 CheckPaymentDueReminders，但 asynq 支持任务重试、持久化队列，更适合"状态流转失败后重试"场景 |
| go-redis 直接操作 | eko/gocache | 增加抽象层，直接用 go-redis 操作 SIMonthlyPayment 状态足够清晰 |

---

## Architecture Patterns

### System Architecture Diagram

```
[浏览器 / H5前端]
       │
       │ ① GET /si/dashboard?year=&month=
       ▼
[社保 Handler] ──→ [社保 DashboardService]
       │                  │
       │                  │ ② GORM SELECT SUM(si_monthly_payments)
       │                  │    WHERE year_month = YYYY-MM
       │                  ▼
       │           [PostgreSQL: simonthly_payments]
       │
       │ ③ POST /si/enroll  |  POST /si/stop
       ▼
[社保 Handler] ──→ [社保 Service] ──→ [社保 Repository]
       │                  │                  │
       │                  │                  │ INSERT social_insurance_records
       │                  │                  │ INSERT social_insurance_change_histories
       │                  │                  │ INSERT simonthly_payments (新)
       │                  ▼                  ▼
       │           [PostgreSQL: social_insurance_records]
       │           [PostgreSQL: social_insurance_change_histories]
       │           [PostgreSQL: simonthly_payments]
       │
       │ ④ asynq 定时任务（每天凌晨）
       │    - 生成下月 SIMonthlyPayment 记录
       │    - 状态流转: pending→normal/overdue
       │    - 清理超过24个月的历史记录
       ▼
[asynq Worker] ←── [Redis Queue]
       │
       │ ⑤ 代理缴费 webhook（外部回调）
       │    POST /si/webhook/agent-payment
       ▼
[asynq Worker] → 状态更新: SIMonthlyPayment.status = normal
```

### Recommended Project Structure

```
internal/socialinsurance/
├── model.go              # SocialInsuranceRecord / ChangeHistory（已有）
├── monthly_payment.go     # [NEW] SIMonthlyPayment 模型
├── organization.go       # [NEW] Organization.payment_channel 字段扩展
├── dto.go                # [扩展] EnrollRequest / StopRequest / DashboardResponse / FiveInsDetail
├── repository.go         # [扩展] MonthlyPayment Repository 方法
├── service.go            # [扩展] Enroll/Stop/DashboardService
├── handler.go            # [扩展] Dashboard/Enroll/Stop/Export/AgentWebhook handlers
├── scheduler.go          # [已有] gocron 定时任务（缴费到期提醒）
├── asynq_task.go         # [NEW] asynq task type 定义（MonthlyPaymentGen / StatusTransition）
├── asynq_worker.go       # [NEW] asynq worker 处理器
├── excel.go              # [已有] generatePaymentDetailExcel（扩展五险分项列）
└── reminder_model.go     # [已有] Reminder

frontend/src/views/tool/
├── SITool.vue            # [重构] 参保操作Tab新增数据看板 + 优化增减员弹窗 + 红色横幅
├── SIStatsCard.vue       # [NEW] 4卡片组件（复用 SalaryDashboard 样式）
├── EnrollDialog.vue      # [NEW] 优化增员弹窗（姓名搜索 + 起始月 + 城市 + 基数 + 公积金）
├── StopDialog.vue       # [NEW] 优化减员弹窗（姓名搜索 + 终止月 + 原因 + 转出日期）
├── FiveInsDetailDialog.vue # [NEW] 五险分项弹窗
├── SIDashboard.vue       # [NEW] 独立数据看板页面（4卡片 + 环比）
└── components/
    ├── SIRecordStatusTag.vue  # [NEW] 5状态标签组件
    ├── SIPaymentChannel.vue   # [NEW] 缴费渠道展示组件
    └── SIOverdueBanner.vue    # [NEW] 欠缴红色横幅组件
```

### Pattern 1: Dashboard 聚合查询（复用 SalaryDashboardService 模式）

**What:** 并发查询 4 个指标 + 计算环比百分比，格式与 SalaryDashboardResponse 完全一致。

**When to use:** 社保数据看板 SI-01~SI-04 需要聚合 SIMonthlyPayment 表数据。

**Example:**
```go
// Source: internal/salary/dashboard_service.go（已有模式）
type SIDashboardService struct {
    db *gorm.DB
}

type siDashboardIndicator struct {
    current  float64
    previous float64
}

// GetDashboard 获取社保看板（应缴总额/单位/个人/欠缴 + 环比）
func (s *SIDashboardService) GetDashboard(ctx context.Context, orgID int64, year, month int) (*SIDashboardResponse, error) {
    prevYear, prevMonth := prevYearMonth(year, month)

    var total, company, personal, overdue siDashboardIndicator
    g, _ := errgroup.WithContext(ctx)

    g.Go(func() error {
        curr, _ := s.sumMonthlyPayment(orgID, year, month, "total_amount")
        prev, _ := s.sumMonthlyPayment(orgID, prevYear, prevMonth, "total_amount")
        total = siDashboardIndicator{current: curr, previous: prev}
        return nil
    })
    // ... 并发查询其他 3 个指标
    if err := g.Wait(); err != nil {
        return nil, err
    }
    return s.toResponse([]siDashboardIndicator{total, company, personal, overdue}), nil
}
```

### Pattern 2: asynq Worker 注册 + 任务处理（复用 SlipSendService 模式）

**What:** 复用 `internal/salary/slip_send_task.go` 的 asynq task 注册模式。

**When to use:** 定时生成月度记录、状态流转、代理缴费 webhook。

**Example:**
```go
// Source: internal/salary/slip_send_task.go（已有模式）
// asynq task type
const TypeMonthlyPaymentGen = "si:monthly:gen"
const TypePaymentStatusTransition = "si:payment:status-transition"

type MonthlyPaymentPayload struct {
    OrgID int64 `json:"org_id"`
    Year  int   `json:"year"`
    Month int   `json:"month"`
}

type StatusTransitionPayload struct {
    OrgID int64  `json:"org_id"`
    Day   int    `json:"day"` // 每月第几天触发
}

// NewMonthlyPaymentTask 创建月度记录生成任务
func NewMonthlyPaymentTask(payload *MonthlyPaymentPayload) (*asynq.Task, error) {
    data, _ := json.Marshal(payload)
    return asynq.NewTask(TypeMonthlyPaymentGen, data), nil
}
```

### Pattern 3: 增减员 INSERT ONLY（复用 ChangeHistory 模式）

**What:** 参保记录变更只 INSERT 新记录到 ChangeHistory，原始 SocialInsuranceRecord 只在创建时写入。

**When to use:** 所有增员、减员、基数调整操作。

**Example:**
```go
// Source: internal/socialinsurance/service.go BatchEnroll / BatchStopEnrollment（已有模式）
// 增员：INSERT SocialInsuranceRecord + INSERT ChangeHistory + INSERT SIMonthlyPayment
history := &ChangeHistory{
    RecordID:   record.ID,
    EmployeeID: empID,
    ChangeType: SIChangeEnroll,
    AfterValue: datatypes.JSON(detailsJSON),
    Remark:     "批量参保",
}
// 后续查询时，ChangeHistory 按时间倒序取最新记录作为当前状态
```

### Pattern 4: 5状态标签 UI（复用 el-tag 组件）

**What:** 参保记录列表状态列用 5 种颜色标签（正常-绿/待缴-黄/欠缴-红/已转出-灰/未转出-蓝）。

**When to use:** D-SI-11 参保记录列表状态列。

**Example:**
```vue
<!-- Source: frontend/src/views/tool/SalaryList.vue statusMap 模式（已有）-->
<el-tag :type="siStatusTagType[row.status] || 'info'" size="small">
  {{ siStatusMap[row.status] }}
</el-tag>

<script setup>
// D-SI-11: 5种状态
const siStatusMap: Record<string, string> = {
  normal: '正常',
  pending: '待缴',
  overdue: '欠缴',
  transferred: '已转出',
  not_transferred: '未转出',
}
const siStatusTagType: Record<string, string> = {
  normal: 'success',      // 绿
  pending: 'warning',     // 黄
  overdue: 'danger',       // 红
  transferred: 'info',     // 灰
  not_transferred: '',    // 蓝（default）
}
</script>
```

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| 精确金额计算 | float64 | shopspring/decimal | float64 精度丢失，金额计算必须 decimal.Decimal（项目已安装） |
| 月度缴费状态追踪 | UPDATE SocialInsuranceRecord | SIMonthlyPayment 新表 | conflate 参保生命周期状态与月度缴费状态会导致状态逻辑混乱 |
| 定时任务持久化 | gocron 内存队列 | asynq | asynq 任务持久化到 Redis，worker 重启后任务不丢失；gocron 适合简单定时（CheckPaymentDueReminders），asynq 适合需要重试的任务（状态流转失败重试） |
| Excel 生成 | 手写 xlsx 格式 | excelize | 项目已安装；支持样式/公式/自动列宽，代码量少 |
| 社保五险计算 | 手写计算公式 | calculateDetails 函数（已有） | 已有 `internal/socialinsurance/service.go:calculateDetails`，包含 clamp 逻辑和税率处理 |

**Key insight:** `calculateDetails` 函数已实现五险分项计算（养老/医疗/失业/工伤/生育/公积金），Phase 8 复用此函数，只需扩展公积金基数和比例的独立输入支持。

---

## Runtime State Inventory

> Phase 8 不是 rename/refactor/migration 阶段，跳过 Runtime State Inventory。

---

## Common Pitfalls

### Pitfall 1: conflate 参保生命周期状态与月度缴费状态

**What goes wrong:** 将 `pending/active/stopped`（参保生命周期）与 `normal/pending/overdue`（月度缴费状态）混用，导致逻辑混乱。

**Why it happens:** Phase 8 前只有 3 种参保状态，开发者习惯性地在 SocialInsuranceRecord.status 上扩展。

**How to avoid:** D-SI-01 决策明确：月度缴费状态存 SIMonthlyPayment 表，参保生命周期状态存 SocialInsuranceRecord.status。两套状态完全独立，查询时 JOIN 两个表。

**Warning signs:** 代码中出现 `record.Status == "overdue"` 而未限定表名。

### Pitfall 2: float64 计算金额

**What goes wrong:** 金额 0.1 + 0.2 = 0.30000000000000004，导致账目不平。

**Why it happens:** Go float64 使用 IEEE 754 浮点表示，有精度丢失问题。

**How to avoid:** 所有金额字段用 `shopspring/decimal`，SQL 用 `NUMERIC/DECIMAL` 类型，GORM 自动映射。

**Warning signs:** 代码中出现 `var amount float64` 且该变量用于金额计算。

### Pitfall 3: 状态流转规则写死 15 日（应为 26 日）

**What goes wrong:** D-SI-03 决策明确"≥26日未缴→overdue"，但如果代码中写的是 15 日，会导致状态提前流转。

**Why it happens:** 社保缴费截止日各地不同，上海等城市是每月 15 日，但 D-SI-03 决策规定 26 日为系统统一截止日。

**How to avoid:** 状态流转逻辑以 D-SI-03 决策为准（26 日），注释明确说明"系统统一使用每月 26 日作为缴费截止分界点，与各地实际截止日无关"。

**Warning signs:** 代码中 hardcoded `paymentDay == 15` 或 `day > 15`。

### Pitfall 4: 重复生成月度记录

**What goes wrong:** asynq 定时任务重复执行导致 SIMonthlyPayment 产生重复记录。

**Why it happens:** asynq 任务重复执行（MaxRetry > 0）且没有幂等检查。

**How to avoid:** 在 INSERT 前检查 `WHERE employee_id = ? AND year_month = ?` 是否已存在；使用 `ON CONFLICT DO NOTHING`（PostgreSQL UPSERT）。

**Warning signs:** `repo.CreateMonthlyPayment` 无唯一性检查。

### Pitfall 5: INSERT ONLY 策略被突破

**What goes wrong:** 开发者为"修正历史错误"UPDATE 了 SIMonthlyPayment 记录。

**Why it happens:** 调试阶段方便修改数据。

**How to avoid:** Repository 层不暴露 UPDATE 方法；只提供 `UpdateStatus` 有限更新（且仅限 status 和 payment_channel 字段）；所有其他字段变更走 INSERT 新记录。

**Warning signs:** Repository 出现 `db.Updates(map[string]interface{}{...})` 且包含非 status 字段。

### Pitfall 6: 环比上月计算包含未确认月份

**What goes wrong:** 上月工资表未 confirmed 时，数据被计入环比计算，显示错误的百分比。

**Why it happens:** D-SI-06 要求"仅统计 confirmed 状态的月份"，但代码未加此过滤条件。

**How to avoid:** Dashboard 聚合查询只统计 status='confirmed' 的 SIMonthlyPayment 记录；欠缴金额单独查询 overdue 状态的 SUM。

**Warning signs:** Dashboard SUM 查询无 WHERE status 条件。

---

## Code Examples

### 新增: SIMonthlyPayment 模型

```go
// Source: 扩展 internal/socialinsurance/model.go
// 缴费状态常量（D-SI-01 决策）
const (
    SIPayStatusNormal        = "normal"         // 正常
    SIPayStatusPending       = "pending"         // 待缴
    SIPayStatusOverdue      = "overdue"         // 欠缴
    SIPayStatusTransferred  = "transferred"      // 已转出
    SIPayStatusNotTransferred = "not_transferred" // 未转出
)

// 缴费渠道常量
const (
    SIPayChannelSelf        = "self"             // 自主缴费
    SIPayChannelAgentNew   = "agent_new"         // 代理缴费新客
    SIPayChannelAgentExisting = "agent_existing" // 代理缴费已合作
)

// SIMonthlyPayment 月度缴费记录
type SIMonthlyPayment struct {
    model.BaseModel
    EmployeeID    int64           `gorm:"column:employee_id;not null;index:idx_emp_month,priority:1;comment:员工ID"`
    OrgID         int64           `gorm:"column:org_id;not null;index;comment:组织ID"`
    YearMonth     string          `gorm:"column:year_month;type:varchar(7);not null;index:idx_emp_month,priority:2;comment:年月 YYYY-MM"`
    Status        string          `gorm:"column:status;type:varchar(20);not null;default:pending;comment:缴费状态"`
    PaymentChannel string         `gorm:"column:payment_channel;type:varchar(20);not null;default:self;comment:缴费渠道"`
    CompanyAmount decimal.Decimal `gorm:"column:company_amount;type:decimal(12,2);not null;comment:单位月缴"`
    PersonalAmount decimal.Decimal `gorm:"column:personal_amount;type:decimal(12,2);not null;comment:个人月缴"`
    TotalAmount   decimal.Decimal `gorm:"column:total_amount;type:decimal(12,2);not null;comment:合计"`
    DueDate       *time.Time      `gorm:"column:due_date;comment:应缴日期"`
    PaidAt        *time.Time      `gorm:"column:paid_at;comment:实缴时间"`
}

func (SIMonthlyPayment) TableName() string {
    return "si_monthly_payments"
}
```

### 新增: Dashboard 聚合（复用 salary/dashboard_service.go 模式）

```go
// Source: internal/socialinsurance/dashboard_service.go
type SIDashboardService struct {
    db *gorm.DB
}

type SIDashboardResponse struct {
    Stats []StatItem `json:"stats"`
}

type StatItem struct {
    Label          string  `json:"label"`
    Value          string  `json:"value"`
    TrendPercent   *string `json:"trend_percent"`
    TrendDirection string  `json:"trend_direction"`
}

func (s *SIDashboardService) GetDashboard(ctx context.Context, orgID int64, year, month int) (*SIDashboardResponse, error) {
    prevYear, prevMonth := prevYearMonth(year, month)

    var total, company, personal, overdue siDashboardIndicator
    g, _ := errgroup.WithContext(ctx)

    g.Go(func() error {
        curr, _ := s.sumField(orgID, year, month, "total_amount", "normal", "pending", "overdue")
        prev, _ := s.sumField(orgID, prevYear, prevMonth, "total_amount", "normal", "pending", "overdue")
        total = siDashboardIndicator{current: curr, previous: prev}; return nil
    })
    g.Go(func() error {
        curr, _ := s.sumField(orgID, year, month, "company_amount", "normal", "pending", "overdue")
        prev, _ := s.sumField(orgID, prevYear, prevMonth, "company_amount", "normal", "pending", "overdue")
        company = siDashboardIndicator{current: curr, previous: prev}; return nil
    })
    g.Go(func() error {
        curr, _ := s.sumField(orgID, year, month, "personal_amount", "normal", "pending", "overdue")
        prev, _ := s.sumField(orgID, prevYear, prevMonth, "personal_amount", "normal", "pending", "overdue")
        personal = siDashboardIndicator{current: curr, previous: prev}; return nil
    })
    g.Go(func() error {
        curr, _ := s.sumField(orgID, year, month, "total_amount", "overdue")
        prev, _ := s.sumField(orgID, prevYear, prevMonth, "total_amount", "overdue")
        overdue = siDashboardIndicator{current: curr, previous: prev}; return nil
    })

    if err := g.Wait(); err != nil { return nil, err }

    return &SIDashboardResponse{Stats: []StatItem{
        s.toStatItem("当月应缴总额", total),
        s.toStatItem("单位部分合计", company),
        s.toStatItem("个人部分合计", personal),
        s.toStatItem("欠缴金额", overdue),
    }}, nil
}
```

### 新增: asynq Worker 状态流转（D-SI-03）

```go
// Source: internal/socialinsurance/asynq_worker.go
// StatusTransitionPayload: D-SI-03 状态流转
func (w *Worker) HandleStatusTransition(ctx context.Context, t *asynq.Task) error {
    var payload StatusTransitionPayload
    if err := json.Unmarshal(t.Payload(), &payload); err != nil {
        return fmt.Errorf("unmarshal payload: %w", err)
    }

    cstZone := time.FixedZone("CST", 8*3600)
    today := time.Now().In(cstZone)

    // D-SI-03 决策：每月26日为状态更新分界点
    cutoffDay := 26
    if today.Day() >= cutoffDay {
        // ≥26日：所有 pending 且未缴的 → overdue
        if err := w.repo.UpdateOverduePayments(ctx, payload.OrgID); err != nil {
            return fmt.Errorf("update overdue: %w", err)
        }
    } else {
        // <26日：已确认缴费的 → normal
        if err := w.repo.UpdatePaidToNormal(ctx, payload.OrgID); err != nil {
            return fmt.Errorf("update normal: %w", err)
        }
    }
    return nil
}
```

### 扩展: 五险分项弹窗数据响应

```go
// Source: internal/socialinsurance/dto.go
// D-SI-12: 五险分项响应
type FiveInsDetailResponse struct {
    EmployeeName    string           `json:"employee_name"`
    CityName        string           `json:"city_name"`
    BaseAmount      decimal.Decimal `json:"base_amount"`
    YearMonth       string           `json:"year_month"`
    Items           []FiveInsItem    `json:"items"` // 养老/医疗/失业/工伤/生育/公积金
    OtherPayments   []OtherPayment  `json:"other_payments"` // 滞纳金/残保金/漏缴/补缴
    TotalCompany    decimal.Decimal `json:"total_company"`
    TotalPersonal   decimal.Decimal `json:"total_personal"`
}

type FiveInsItem struct {
    Name           string           `json:"name"`
    CompanyAmount  decimal.Decimal `json:"company_amount"`
    PersonalAmount decimal.Decimal `json:"personal_amount"`
}
```

### 前端: 4 卡片样式（复用 SalaryDashboard.vue）

```vue
<!-- Source: frontend/src/views/tool/SalaryDashboard.vue（已有）-->
<!-- D-SI-05: 社保数据看板复用薪资看板样式 -->
<div class="stats-grid">
  <div v-for="stat in stats" :key="stat.label" class="stat-card">
    <div class="stat-value">{{ stat.value }}</div>
    <div class="stat-label">{{ stat.label }}</div>
    <div v-if="stat.trend_percent" class="stat-trend" :class="stat.trend_direction">
      <span v-if="stat.trend_direction === 'up'">&#8593;</span>
      <span v-else-if="stat.trend_direction === 'down'">&#8595;</span>
      {{ Math.abs(parseFloat(stat.trend_percent)) }}%
    </div>
    <div v-else class="stat-trend neutral">--</div>
  </div>
</div>

<style scoped>
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
}
.stat-card {
  background: #fafafa;
  padding: 16px;
  border-radius: 8px;
  text-align: center;
}
</style>
```

### 前端: 欠缴红色横幅（D-SI-10）

```vue
<!-- Source: D-SI-10 决策 -->
<template>
  <div v-if="overdueItems.length > 0" class="overdue-banner">
    <el-icon class="banner-icon"><WarningFilled /></el-icon>
    <div class="banner-content">
      <span class="banner-title">欠缴提醒</span>
      <span class="banner-main">
        {{ overdueItems[0].employee_name }} - {{ overdueItems[0].city_name }}
        {{ overdueItems[0].year_month }} 欠缴 ¥{{ overdueItems[0].total_amount }}
      </span>
      <span v-if="overdueItems.length > 1" class="banner-more">
        等 {{ overdueItems.length }} 项未处理
      </span>
    </div>
    <el-button class="banner-close" text @click="dismissBanner">关闭</el-button>
  </div>
</template>

<style scoped>
.overdue-banner {
  background: #fff2f0;
  border: 1px solid #ffccc7;
  border-radius: 8px;
  padding: 12px 16px;
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.banner-icon { color: #ff4d4f; }
.banner-title { font-weight: 600; color: #ff4d4f; margin-right: 8px; }
.banner-main { color: #1a1a1a; }
.banner-more { color: #8c8c8c; font-size: 12px; margin-left: 8px; }
</style>
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| 社保缴费状态用 3 种（pending/active/stopped） | SIMonthlyPayment 表独立追踪 5 种缴费状态 | Phase 08 | 月度欠缴管理更精确，状态流转由 asynq 自动化 |
| 增减员用 el-select 下拉选员工 | EnrollDialog/StopDialog 姓名搜索触发 employeeApi.search | Phase 08 | 操作更便捷，符合"3步完成核心操作"Core Value |
| 社保截止日 hardcoded 15 日 | 状态流转以 26 日为分界点（D-SI-03） | Phase 08 | 与实际社保局缴费周期对齐 |
| 参保记录无缴费渠道列 | 新增 payment_channel 列 + 3 种渠道选择 | Phase 08 | 支持自主/代理新客/代理已合作 3 种渠道管理 |
| Excel 导出仅 16 列 | D-SI-13 扩展五险分项 + 其他缴费行 | Phase 08 | 满足 SI-21 Excel 导出要求 |

**Deprecated/outdated:**
- `SocialInsuranceRecord.Status` 作为唯一状态来源：Phase 08 后缴费状态应从 SIMonthlyPayment 读取，不再 conflate
- 增员/减员时直接 UPDATE record.status：Phase 08 后改 INSERT ChangeHistory，遵循 INSERT ONLY

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | SIMonthlyPayment 表在 social_insurance_records 表之外独立存在 | Standard Stack / model.go | CRITICAL - 如果要合并到现有表，实现方式完全不同 |
| A2 | asynq 用于月度记录生成和状态流转（scheduler.go 已用 gocron，两者互补） | Standard Stack | MEDIUM - 如果只用 gocron 做状态流转，worker 注册方式不同 |
| A3 | 代理缴费 webhook 通过 asynq 队列处理 | Code Examples | MEDIUM - 如果用 gocron 处理 webhook，入口路由不同 |
| A4 | D-SI-03 状态流转截止日固定为 26 日（不配置化） | Common Pitfalls | MEDIUM - 如果要配置化，需扩展 Organization 表加字段 |
| A5 | 五险分项弹窗直接读取 SIMonthlyPayment + SocialInsuranceRecord 合并展示 | Code Examples | LOW - 如果需要从其他表读取，调整 SQL JOIN |

---

## Open Questions

1. **asynq cron 表达式（Claude's Discretion）**
   - What we know: D-SI-03 决策每天凌晨触发，需要知道几点运行
   - What's unclear: scheduler.go 现有 CheckPaymentDueReminders 在 08:00 运行，新任务放几点
   - Recommendation: 凌晨 02:00 或 03:00 均可，避开业务高峰期；建议 02:00 作为 Claude's Discretion 推荐

2. **SI-15 自主缴费跳转外部页面**
   - What we know: 需要跳转确认页面
   - What's unclear: 第三方缴费平台 URL，合作渠道未确定
   - Recommendation: 初期跳转前显示"即将跳转到外部缴费页面"确认对话框；URL 作为配置项存 Organization 表

3. **五险分项弹窗数据来源**
   - What we know: D-SI-12 要求展示各险种单位+个人金额
   - What's unclear: 金额来源是 SIMonthlyPayment.total_amount 还是 SocialInsuranceRecord.details JSON
   - Recommendation: SIMonthlyPayment 存储各险种明细（JSONB 字段），避免 JOIN；设计时参考 `InsuranceAmountDetail[]` 结构

4. **欠缴横幅显示策略（Claude's Discretion）**
   - What we know: D-SI-10 要求横幅展示最大欠缴项
   - What's unclear: 超过 N 条时如何显示
   - Recommendation: 显示第 1 条+"等 X 项"，下方展开折叠列表；N 阈值暂定 3 条

---

## Environment Availability

> Step 2.6: SKIPPED（Phase 8 是代码/配置扩展，无新增外部依赖；所有库已在 go.mod 和 package.json 中）

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `go test` + `stretchr/testify` |
| Config file | none — 标准 Go 测试 |
| Quick run command | `go test ./internal/socialinsurance/... -run TestMonthlyPayment -v` |
| Full suite command | `go test ./internal/socialinsurance/... -race -cover` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| SI-01~SI-04 | 社保数据看板 4 指标聚合 | unit | `go test ./internal/socialinsurance/... -run TestSIDashboard -v` | ❌ Wave 0 |
| SI-05~SI-08 | 增员弹窗姓名搜索 + 表单提交 + 校验 | unit+integration | `go test ./internal/socialinsurance/... -run TestEnroll -v` | ❌ Wave 0 |
| SI-09~SI-13 | 减员弹窗终止月校验 + 原因必选 + 生效规则提示 | unit | `go test ./internal/socialinsurance/... -run TestStop -v` | ❌ Wave 0 |
| SI-14~SI-16 | 缴费渠道写入 + 代理缴费 webhook | integration | `go test ./internal/socialinsurance/... -run TestPaymentChannel -v` | ❌ Wave 0 |
| SI-17~SI-18 | 状态自动流转（normal→pending→overdue） | unit | `go test ./internal/socialinsurance/... -run TestStatusTransition -v` | ❌ Wave 0 |
| SI-19~SI-20 | 欠缴横幅 + 五险分项弹窗 | unit | `go test ./internal/socialinsurance/... -run TestFiveInsDetail -v` | ❌ Wave 0 |
| SI-21 | Excel 导出五险分项列 | unit | `go test ./internal/socialinsurance/... -run TestExport -v` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/socialinsurance/... -run <TestName> -v`
- **Per wave merge:** `go test ./internal/socialinsurance/... -race -cover`
- **Phase gate:** Full suite green before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] `internal/socialinsurance/monthly_payment_test.go` — covers SI-01~SI-21
- [ ] `internal/socialinsurance/asynq_worker_test.go` — covers SI-17~SI-18
- [ ] Framework install: `go test ./internal/socialinsurance/...` — verify existing tests pass before adding new

*(Existing test infrastructure: `internal/socialinsurance/repository_test.go`, `internal/socialinsurance/service_test.go` — extend these files)*

---

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | yes | JWT via golang-jwt/v5（已在项目中使用） |
| V3 Session Management | no | 不涉及会话管理 |
| V4 Access Control | yes | org_id 多租户隔离（已在现有代码中强制） |
| V5 Input Validation | yes | go-playground/validator（已在项目中使用）+ 增员/减员表单校验 |
| V6 Cryptography | yes | 身份证号 AES-256-GCM（已有基础设施） |

### Known Threat Patterns for Go/Go + Vue3

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| SQL 注入 | Tampering | GORM 参数化查询（已有） |
| 金额精度丢失 | Tampering | shopspring/decimal（Phase 08 必须使用） |
| 权限绕过（跨租户访问） | Information Disclosure | org_id Scope 强制注入（已有 GORM Hook） |
| 定时任务重复执行 | Denial | asynq 幂等性检查 + UPSERT |
| webhook 重放攻击 | Spoofing | webhook 签名验证 + 时间戳检查 |

---

## Sources

### Primary (HIGH confidence)
- `internal/socialinsurance/service.go` — BatchEnroll/BatchStopEnrollment/calculateDetails 实现
- `internal/socialinsurance/model.go` — SocialInsuranceRecord/ChangeHistory 模型
- `internal/socialinsurance/scheduler.go` — gocron 定时任务框架
- `internal/socialinsurance/excel.go` — excelize 导出模式
- `internal/salary/slip_send_task.go` — asynq task 注册模式
- `internal/salary/dashboard_service.go` — Dashboard 聚合查询模式
- `frontend/src/views/tool/SalaryDashboard.vue` — 4卡片样式参考
- `frontend/src/views/tool/SalaryList.vue` — 导出对话框参考
- `go.mod` — asynq v0.26.0, go-redis v9.18.0, shopspring/decimal v1.4.0, excelize v2.10.1

### Secondary (MEDIUM confidence)
- `.planning/phases/08-社保公积金增强/08-CONTEXT.md` — Phase 8 决策（D-SI-01~D-SI-13）
- `.planning/research/SUMMARY.md` — v1.3 研究综合报告
- `.planning/research/FEATURES.md §2.4` — 社保公积金增强功能分析
- `.planning/REQUIREMENTS.md §4` — SI-01~SI-21 需求定义

### Tertiary (LOW confidence)
- 各地社保缴费截止日规范（需在实现时确认具体城市政策）

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — 全部库已安装并验证版本
- Architecture: HIGH — 基于现有代码审查（scheduler.go、slip_send_service.go、dashboard_service.go）
- Pitfalls: HIGH — 基于代码审查和业务逻辑分析，6 个陷阱均有预防策略

**Research date:** 2026-04-19
**Valid until:** 2026-05-19（30天，稳定阶段）

---

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SI-01 | 数据看板展示当月应缴总额（单位+个人+欠缴），环比上月百分比 | SIDashboardService 并发聚合查询模式 |
| SI-02 | 数据看板展示当月单位部分合计，环比上月百分比 | 同上 |
| SI-03 | 数据看板展示当月个人部分合计，环比上月百分比 | 同上 |
| SI-04 | 数据看板展示欠缴金额（上月+当月截止25日未缴），26日更新数据 | asynq 状态流转（≥26日→overdue） |
| SI-05 | 增员弹窗支持姓名输入或检索，自动填充身份证号码 | EnrollDialog + employeeApi.search |
| SI-06 | 增员弹窗起始月份默认当月，可选近3个月，当月1日前生成欠缴账单 | EnrollDialog date-picker + SIMonthlyPayment 初始化 status=pending |
| SI-07 | 增员弹窗支持输入社保缴费城市和缴费基数 | EnrollDialog 表单字段 + calculateDetails |
| SI-08 | 增员弹窗支持输入公积金缴费比例和基数（默认与社保同步） | EnrollDialog 表单字段扩展 |
| SI-09 | 减员弹窗支持姓名输入或检索，自动填充身份证号码 | StopDialog + employeeApi.search |
| SI-10 | 减员弹窗终止月份默认当月且不可早于当月 | StopDialog date-picker min-value 校验 |
| SI-11 | 减员弹窗必填原因（三选一） | StopDialog el-radio 必选校验 |
| SI-12 | 减员弹窗支持选择转出社保日期和封存公积金日期，自动生成最后缴存月 | StopDialog 日期联动逻辑 |
| SI-13 | 转出日期提示：每月5-26日前转出当月生效，26日后次月生效 | StopDialog 表单内联提示文字 |
| SI-14 | 管理员可选择缴费渠道（自主缴费/代理缴费新客/代理缴费已合作） | Organization.payment_channel 字段 + UI 下拉选择 |
| SI-15 | 自主缴费点击跳转已缴完成确认 | 外部 URL 跳转（Claude's Discretion） |
| SI-16 | 代理缴费（已合作）每月15日自动扣缴，成功/失败发送通知 | asynq webhook 处理器 + 通知服务 |
| SI-17 | 缴费状态自动流转（正常→待缴→欠缴→已转出），26日为分界点 | asynq 定时任务状态流转逻辑 |
| SI-18 | 社保列表新增5种状态（正常/待缴/欠缴/已转出/未转出） | SIPayStatus* 常量 + el-tag UI |
| SI-19 | 社保关键节点红字提醒 | SIOverdueBanner 组件 |
| SI-20 | 社保明细展示五险各项金额（单位+个人）+ 其他缴费 | FiveInsDetailDialog + InsuranceAmountDetail[] |
| SI-21 | 社保列表支持 Excel 格式下载导出 | excelize 扩展五险分项列 |
