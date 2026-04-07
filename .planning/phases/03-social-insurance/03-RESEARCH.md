# Phase 3: 社保管理 - Research

**Researched:** 2026-04-07
**Domain:** 社保五险一金管理（Go/Gin/GORM/PostgreSQL/JSONB/gocron）
**Confidence:** HIGH

## Summary

Phase 3 实现社保五险一金管理模块，核心功能包括：自建30+城市政策库（单表+JSONB）、员工参保/停缴批量操作、缴费到期自动提醒（gocron定时任务）、缴费明细查询与导出。模块遵循已建立的三层架构（handler/service/repository），复用Phase 1-2的多租户隔离、RBAC、审计日志等基础设施。

关键技术决策已由CONTEXT.md锁定：JSONB存储政策数据、gocron v2.19.1实现定时提醒、离职触发社保停缴提醒（复用onEmployeeResigned回调）、参保3步操作流程。本项目已有`gorm.io/datatypes v1.2.7`依赖，可直接使用`datatypes.JSON`或`datatypes.JSONType[T]`处理JSONB字段。

**Primary recommendation:** 使用`datatypes.JSONType[T]`强类型封装JSONB政策数据，结合gocron+Redis分布式锁实现每日缴费到期扫描，通过接口解耦社保模块与员工模块（社保定义接口，员工模块回调）。

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** 单表+JSONB存储社保政策库。一个城市一条记录，字段包括：城市ID、生效年份、五险一金配置（JSONB）。JSONB结构包含每个险种的企业缴费比例、个人缴费比例、基数下限、基数上限。
- **D-02:** 管理员在H5后台手动录入/编辑政策数据。提供政策编辑页面。初始30+城市数据通过管理后台逐个录入。
- **D-03:** 五险一金全覆盖：养老保险、医疗保险、失业保险、工伤保险、生育保险、住房公积金。
- **D-04:** 政策按年度生效（effective_year字段）。查询时取 effective_year <= 当前年份 的最新记录。
- **D-05:** 参保流程3步：选择员工 -> 匹配基数+预览金额 -> 确认批量生效。
- **D-06:** 支持批量参保/停缴。前端传 employee_ids 数组。
- **D-07:** 三种停缴触发：手动停缴、离职自动触发提醒、转正自动检查。
- **D-08:** 社保记录状态：pending -> active -> stopped。
- **D-09:** 使用 gocron v2.19.1 实现定时检查。每日扫描缴费到期情况。
- **D-10:** 提醒方式：APP内消息 + 首页待办卡片。不用短信/微信模板消息。
- **D-11:** 缴费截止日为每月固定日期（全局配置，非企业自定义）。
- **D-12:** 社保模块提供查询接口 `GetSocialInsuranceDeduction(orgID, employeeID, month)` 供Phase 5调用。单向依赖。
- **D-13:** 薪资变动时只提醒不自动调整社保基数。具体实现留Phase 5联动。
- **D-14:** RBAC权限：OWNER全部操作、ADMIN同OWNER、MEMBER仅查看自己社保记录。

### Claude's Discretion
- 社保政策库 JSONB 内部具体字段命名
- 参保记录的具体数据模型（是否按险种拆分行 vs 一条记录存所有险种）
- 缴费明细的记录粒度（按月汇总 vs 按次记录）
- 导出凭证的 Excel/PDF 格式细节
- 社保模块内部目录结构

### Deferred Ideas (OUT OF SCOPE)
None
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SOCL-01 | 根据员工城市+岗位自动匹配社保参保基数（自建30+城市政策库） | JSONB政策库 + 城市ID关联（D-01/D-03/D-04），本项目已有37城市数据 |
| SOCL-02 | 老板勾选员工+确认参保，一键生成参保材料PDF | 批量参保流程（D-05/D-06），复用pdf.go模式（已有go-pdf/fpdf依赖） |
| SOCL-03 | 社保缴费到期前3天自动提醒老板（APP内消息+首页卡片） | gocron v2.19.1 + Redis分布式锁（D-09/D-10），每日定时扫描 |
| SOCL-04 | 记录缴费明细，支持查询社保缴纳状态 | 参保记录状态机 pending->active->stopped（D-08），缴费明细表 |
| SOCL-05 | 支持打印/导出社保缴费凭证 | 复用excelize v2.10.1（已在go.mod），Excel导出模式与员工模块一致 |
| SOCL-06 | 员工岗位/薪资变动时自动触发社保基数调整提醒 | 薪资变动提醒（D-13），Phase 5联动时实现，本阶段预留接口 |
| SOCL-07 | 记录社保变更历史，可追溯（参保/基数调整/停缴） | 变更历史表，审计日志Module="social_insurance"自动记录 |
</phase_requirements>

## Standard Stack

### Core (已在go.mod中)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| gorm.io/datatypes | v1.2.7 | JSONB数据类型 | GORM官方JSON类型库，支持PostgreSQL JSONB查询，项目已引入 |
| gocron/v2 | v2.19.1 (需新增) | 定时任务 | CONTEXT.md D-09锁定，支持分布式锁 |
| gocron-redis-lock/v2 | v2.2.1 (需新增) | gocron分布式锁 | Redis分布式锁实现，多实例安全 |
| go-pdf/fpdf | v0.9.0 | 参保材料PDF | 已在go.mod，复用employee/pdf.go模式 |
| excelize/v2 | v2.10.1 | 凭证Excel导出 | 已在go.mod，复用员工导出模式 |

### Supporting (项目已有)
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Gin | v1.12.0 | HTTP框架 | 路由注册，与Phase 1-2一致 |
| GORM | v1.31.1 | ORM | 数据模型、查询、迁移 |
| testify | v1.11.1 | 测试断言 | 单元测试 |
| go-redis/v9 | v9.18.0 | Redis客户端 | gocron分布式锁 + 缓存 |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| datatypes.JSONType[T] | datatypes.JSON (raw bytes) | JSONType提供强类型但暂不支持JSONQuery；本场景政策库按整行读取不需要JSON内部查询，JSONType[T]更适合 |
| gocron | asynq | CONTEXT.md D-09锁定gocron；asynq更适合异步任务队列场景（如工资核算），定时扫描用gocron更轻量 |
| 单条记录存所有险种 | 按险种拆分行 | 单条记录更简单、查询更快，符合"3步操作"原则；拆分行在汇总报表场景有优势但增加复杂度 |

**Installation:**
```bash
go get github.com/go-co-op/gocron/v2@v2.19.1
go get github.com/go-co-op/gocron-redis-lock/v2@v2.2.1
```

**Version verification:**
- gocron v2.19.1 — confirmed available via `go list -m -versions` (latest: v2.20.0, v2.19.1 is stable)
- gocron-redis-lock v2.2.1 — confirmed available (latest stable)
- gorm.io/datatypes v1.2.7 — already in go.mod

## Architecture Patterns

### Recommended Project Structure
```
internal/socialinsurance/
├── model.go              # 数据模型：SocialInsurancePolicy, SocialInsuranceRecord, PaymentDetail, ChangeHistory
├── dto.go                # 请求/响应 DTO
├── handler.go            # HTTP 端点 + 路由注册
├── service.go            # 业务逻辑（参保、停缴、匹配基数、计算金额）
├── repository.go         # 数据访问（GORM查询）
├── scheduler.go          # gocron 定时任务（缴费到期扫描）
├── pdf.go                # 参保材料PDF生成
├── service_test.go       # Service 单元测试
├── repository_test.go    # Repository 单元测试
└── scheduler_test.go     # 定时任务逻辑测试
```

### Pattern 1: JSONB强类型政策数据
**What:** 使用`datatypes.JSONType[T]`封装JSONB字段，Go代码直接操作结构体
**When to use:** 社保政策库的五险一金配置
**Example:**
```go
// Source: go-gorm/datatypes GitHub README + 项目需求
import "gorm.io/datatypes"

// InsuranceItem 单个险种配置
type InsuranceItem struct {
    CompanyRate  float64 `json:"company_rate"`   // 企业缴费比例（如0.16=16%）
    PersonalRate float64 `json:"personal_rate"`  // 个人缴费比例（如0.08=8%）
    BaseLower    float64 `json:"base_lower"`     // 基数下限
    BaseUpper    float64 `json:"base_upper"`     // 基数上限
}

// FiveInsurances 五险一金配置
type FiveInsurances struct {
    Pension       InsuranceItem `json:"pension"`       // 养老保险
    Medical       InsuranceItem `json:"medical"`       // 医疗保险
    Unemployment  InsuranceItem `json:"unemployment"`  // 失业保险
    WorkInjury    InsuranceItem `json:"work_injury"`   // 工伤保险
    Maternity     InsuranceItem `json:"maternity"`     // 生育保险
    HousingFund   InsuranceItem `json:"housing_fund"`  // 住房公积金
}

// SocialInsurancePolicy 社保政策
type SocialInsurancePolicy struct {
    model.BaseModel
    CityID         int                                `gorm:"column:city_id;not null;index" json:"city_id"`
    EffectiveYear  int                                `gorm:"column:effective_year;not null;index" json:"effective_year"`
    Config         datatypes.JSONType[FiveInsurances] `gorm:"column:config;type:jsonb" json:"config"`
}
```

### Pattern 2: gocron定时任务 + Redis分布式锁
**What:** 每日定时扫描缴费到期，使用Redis锁确保多实例安全
**When to use:** 缴费到期提醒（D-09）
**Example:**
```go
// Source: gocron-redis-lock GitHub README + gocron v2 docs
import (
    "github.com/go-co-op/gocron/v2"
    "github.com/go-co-op/gocron-redis-lock/v2"
    redislock "github.com/go-co-op/gocron-redis-lock/v2"
)

func StartScheduler(rdb *redis.Client, svc *Service) (gocron.Scheduler, error) {
    locker, err := redislock.NewLocker(rdb, redislock.WithLockerPrefix("easyhr:social:"))
    if err != nil {
        return nil, fmt.Errorf("create locker: %w", err)
    }

    s, err := gocron.NewScheduler(
        gocron.WithDistributedLocker(locker),
    )
    if err != nil {
        return nil, fmt.Errorf("create scheduler: %w", err)
    }

    // 每天早上8点扫描缴费到期
    _, err = s.NewJob(
        gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(8, 0, 0))),
        gocron.NewTask(func() {
            svc.CheckPaymentDueReminders()
        }),
    )

    s.Start()
    return s, nil
}
```

### Pattern 3: 参保记录状态机
**What:** 社保参保记录生命周期管理
**When to use:** 参保/停缴操作
**Example:**
```go
// 状态常量
const (
    SIStatusPending = "pending"  // 待参保
    SIStatusActive  = "active"   // 参保中
    SIStatusStopped = "stopped"  // 已停缴
)

// 参保记录模型（一条记录存所有险种明细）
type SocialInsuranceRecord struct {
    model.BaseModel
    EmployeeID    int64          `gorm:"column:employee_id;not null;index" json:"employee_id"`
    CityID        int            `gorm:"column:city_id;not null" json:"city_id"`
    BaseAmount    float64        `gorm:"column:base_amount;not null" json:"base_amount"`
    Status        string         `gorm:"column:status;type:varchar(20);not null;default:pending" json:"status"`
    StartMonth    string         `gorm:"column:start_month;type:varchar(7);not null" json:"start_month"` // YYYY-MM
    EndMonth      *string        `gorm:"column:end_month;type:varchar(7)" json:"end_month"`
    PolicyID      int64          `gorm:"column:policy_id;not null" json:"policy_id"`
    Details       datatypes.JSON `gorm:"column:details;type:jsonb" json:"details"` // 各险种明细
}
```

### Pattern 4: 员工模块回调集成
**What:** 通过接口注入实现员工离职事件触发社保提醒
**When to use:** onEmployeeResigned回调（D-07）
**Example:**
```go
// socialinsurance/service.go 定义接口
type EmployeeEventHandler interface {
    OnEmployeeResigned(orgID, employeeID int64)
}

// employee/offboarding_service.go 注入
// 修改 OffboardingService 结构体，增加 socialInsHandler EmployeeEventHandler
// 在 onEmployeeResigned 中调用 socialInsHandler.OnEmployeeResigned(orgID, employeeID)
```

### Anti-Patterns to Avoid
- **JSONB内部查询频繁：** 政策库设计为整行读取（按city_id+effective_year），不要在JSONB内部做复杂查询。`datatypes.JSONType[T]`不支持JSONQuery，但本场景不需要。
- **社保记录按险种拆分行：** 增加6倍数据量和查询复杂度，一条记录存所有险种明细更适合小微企业。
- **硬编码社保比例：** 政策必须从数据库读取，不能硬编码任何城市的比例或基数。
- **自动执行停缴：** D-07明确"仅提醒不自动执行"，停缴必须老板确认。

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| JSONB类型映射 | 手动json.Marshal/Unmarshal + []byte | gorm.io/datatypes JSONType[T] | 自动序列化、类型安全、GORM原生支持 |
| 定时任务 | time.Ticker + goroutine | gocron v2 | 支持cron表达式、分布式锁、错误处理、单实例保证 |
| 分布式锁 | 手写Redis SETNX | gocron-redis-lock | 官方维护、与gocron无缝集成、自动续期 |
| PDF生成 | 复杂排版引擎 | go-pdf/fpdf | 已在项目，纯Go无CGO，参保材料格式简单 |
| Excel导出 | CSV或手动xlsx | excelize v2 | 已在项目，样式/图表/格式化完善 |

**Key insight:** 项目已有`gorm.io/datatypes v1.2.7`（用于离职交接清单JSONB），社保政策库JSONB直接复用同模式。

## Common Pitfalls

### Pitfall 1: 社保基数匹配逻辑错误
**What goes wrong:** 员工薪资低于下限时用薪资而非下限，或高于上限时未封顶
**Why it happens:** 对"基数=薪资但有上下限约束"规则理解不精确
**How to avoid:** 基数计算函数 `clamp(salary, baseLower, baseUpper)` 必须单独测试
**Warning signs:** 缴费金额出现负数、或极小值

### Pitfall 2: 工伤/生育保险个人比例为零
**What goes wrong:** 前端显示工伤/生育的个人扣款为0时用户困惑，或误将企业比例算入个人扣款
**Why it happens:** 中国社保规则中工伤和生育由企业全额承担
**How to avoid:** JSONB中每个险种都明确存储company_rate和personal_rate，工伤/生育的personal_rate=0
**Warning signs:** 员工个人扣款总额不等于各险种个人部分之和

### Pitfall 3: 政策年度查询取错记录
**What goes wrong:** 取了effective_year=当前年份的记录，但7月前应该用上一年度政策
**Why it happens:** 各城市政策调整时间不同（通常7月），新政策未发布时旧政策仍有效
**How to avoid:** 查询条件 `effective_year <= currentYear ORDER BY effective_year DESC LIMIT 1`，确保取最新有效政策
**Warning signs:** 7月后政策未更新，或新年度政策缺失时报错

### Pitfall 4: gocron时区问题
**What goes wrong:** 定时任务在错误时区执行（Go默认UTC，中国是UTC+8）
**Why it happens:** gocron默认使用系统时区，Docker容器可能设为UTC
**How to avoid:** 明确设置 `gocron.WithLocation(time.FixedZone("CST", 8*3600))`
**Warning signs:** 提醒在凌晨而非早上8点触发

### Pitfall 5: 批量参保部分失败
**What goes wrong:** 批量参保10个员工，第5个失败导致前4个也回滚，或前4个成功但5-10未处理
**Why it happens:** 事务粒度选择不当
**How to avoid:** 逐条处理 + 收集成功/失败结果，返回部分成功报告。不使用整体事务包裹批量操作
**Warning signs:** 批量操作API返回非原子性结果但未告知用户哪些成功哪些失败

### Pitfall 6: JSONB在SQLite测试中的兼容性
**What goes wrong:** 单元测试使用SQLite，JSONB类型在SQLite中为TEXT
**Why it happens:** Phase 2已遇到此问题，使用LIKE替代ILIKE
**How to avoid:** datatypes.JSONType[T]已处理SQLite兼容性（存为TEXT）；但如果需要JSON内部查询，需加`json1` build tag
**Warning signs:** 测试中JSONB字段查询失败

## Code Examples

### 参保基数匹配与金额计算
```go
// Source: 业务逻辑推导 + 中国社保规则
func (s *Service) CalculateInsuranceAmounts(cityID int, salary float64, year int) (*InsuranceCalcResult, error) {
    // 1. 查询适用政策
    policy, err := s.repo.FindPolicyByCityAndYear(cityID, year)
    if err != nil {
        return nil, fmt.Errorf("未找到该城市社保政策: %w", err)
    }

    config := policy.Config.Data()
    result := &InsuranceCalcResult{}

    // 2. 对每个险种计算
    calculateItem := func(item InsuranceItem, name string) InsuranceAmount {
        base := salary
        if base < item.BaseLower {
            base = item.BaseLower
        }
        if base > item.BaseUpper {
            base = item.BaseUpper
        }
        return InsuranceAmount{
            Name:          name,
            Base:          base,
            CompanyAmount:  base * item.CompanyRate,
            PersonalAmount: base * item.PersonalRate,
            CompanyRate:   item.CompanyRate,
            PersonalRate:  item.PersonalRate,
        }
    }

    result.Pension = calculateItem(config.Pension, "养老保险")
    result.Medical = calculateItem(config.Medical, "医疗保险")
    result.Unemployment = calculateItem(config.Unemployment, "失业保险")
    result.WorkInjury = calculateItem(config.WorkInjury, "工伤保险")
    result.Maternity = calculateItem(config.Maternity, "生育保险")
    result.HousingFund = calculateItem(config.HousingFund, "住房公积金")

    return result, nil
}
```

### 政策查询（取最新有效年度）
```go
// Source: GORM查询模式 + D-04需求
func (r *Repository) FindPolicyByCityAndYear(cityID int, year int) (*SocialInsurancePolicy, error) {
    var policy SocialInsurancePolicy
    err := r.db.Where("city_id = ? AND effective_year <= ? AND org_id = 0", cityID, year).
        Order("effective_year DESC").
        First(&policy).Error
    if err != nil {
        return nil, err
    }
    return &policy, nil
}
```

### 参保材料PDF生成（复用pdf.go模式）
```go
// Source: internal/employee/pdf.go 参考实现
func GenerateEnrollmentPDF(data *EnrollmentPDFData) ([]byte, error) {
    pdf := fpdf.New("P", "mm", "A4", "")
    pdf.AddPage()
    pdf.SetFont("Helvetica", "B", 16)
    pdf.CellFormat(0, 12, "Social Insurance Enrollment", "", 1, "C", false, 0, "")

    pdf.SetFont("Helvetica", "", 10)
    // 员工信息
    pdf.CellFormat(0, 8, fmt.Sprintf("Employee: %s", data.EmployeeName), "", 1, "L", false, 0, "")
    pdf.CellFormat(0, 8, fmt.Sprintf("City: %s", data.CityName), "", 1, "L", false, 0, "")
    pdf.CellFormat(0, 8, fmt.Sprintf("Base Amount: %.2f CNY", data.BaseAmount), "", 1, "L", false, 0, "")

    // 各险种明细表
    pdf.Ln(5)
    pdf.SetFont("Helvetica", "B", 10)
    pdf.CellFormat(60, 7, "Insurance", "", 0, "L", false, 0, "")
    pdf.CellFormat(35, 7, "Company", "", 0, "C", false, 0, "")
    pdf.CellFormat(35, 7, "Personal", "", 1, "C", false, 0, "")

    pdf.SetFont("Helvetica", "", 10)
    for _, item := range data.Items {
        pdf.CellFormat(60, 7, item.Name, "", 0, "L", false, 0, "")
        pdf.CellFormat(35, 7, fmt.Sprintf("%.2f", item.CompanyAmount), "", 0, "C", false, 0, "")
        pdf.CellFormat(35, 7, fmt.Sprintf("%.2f", item.PersonalAmount), "", 1, "C", false, 0, "")
    }

    var buf bytes.Buffer
    if err := pdf.Output(&buf); err != nil {
        return nil, fmt.Errorf("generate enrollment PDF: %w", err)
    }
    return buf.Bytes(), nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| datatypes.JSON (raw []byte) | datatypes.JSONType[T] (generics) | gorm.io/datatypes v1.2+ | 强类型JSONB，减少手动json.Unmarshal |
| robfig/cron | gocron v2 | 2024-2025 | 支持分布式锁、更友好的API、活跃维护 |
| 单实例定时任务 | gocron + Redis分布式锁 | gocron v2.x | 多实例部署安全，确保单实例执行 |

**Deprecated/outdated:**
- `robfig/cron`: 自2021年起停更，STATE.md已记录此决策

## 中国社保政策数据参考 (2025年)

基于搜索结果的关键城市五险一金数据（用于初始政策库数据）：

### 典型城市缴费比例

| 城市 | 养老(企/个) | 医疗(企/个) | 失业(企/个) | 基数上限 | 基数下限 |
|------|-----------|-----------|-----------|---------|---------|
| 北京 | 16%/8% | 9.8%/2% | 0.5%/0.5% | 35,811 | 7,162 |
| 上海 | 16%/8% | 9%/2% | 0.5%/0.5% | 37,302 | 7,460 |
| 深圳 | 16%/8% | 5%/2%(一档) | 0.8%/0.2% | 27,549 | 5,510 |
| 广州 | 16%/8% | 5.5%/2% | 0.32%/0.2% | ~27,000 | ~5,500 |
| 杭州 | 14%/8% | 9.5%/2% | 0.5%/0.5% | ~22,000 | ~4,400 |

**注意事项：**
- 工伤保险比例按行业风险浮动（0.2%-1.9%），企业全额承担（个人比例=0）
- 生育保险在部分城市已并入医疗保险（北京、上海），深圳仍单独缴纳（企业0.5%）
- 住房公积金比例5%-12%可选，企业和个人同等比例
- 每年7月左右各地调整基数上下限
- 深圳养老保险深户另加1%地方补充

## Open Questions

1. **政策库是否区分org_id？**
   - What we know: D-01说"单表+JSONB"，BaseModel有org_id字段
   - What's unclear: 政策库是全局共享（org_id=0或特殊值）还是企业独有
   - Recommendation: 政策库为全局共享数据（org_id=0），管理员维护全国政策。参保记录按org_id隔离。政策表不嵌入BaseModel或使用org_id=0作为全局标记。

2. **参保记录中的details JSONB存储格式？**
   - What we know: D-08定义状态机，一条记录存所有险种
   - What's unclear: details是存计算后的金额快照还是存当时的比例+基数
   - Recommendation: 存金额快照（base_amount, company_amount, personal_amount per险种），因为政策可能后续变化，历史记录必须保留计算时的金额。

3. **员工城市信息来源？**
   - What we know: Employee模型没有city_id字段，city模块只有37城市列表
   - What's unclear: 如何确定员工参保城市
   - Recommendation: 参保时由老板选择参保城市（默认企业所在城市），参保记录中存city_id。不需要员工模型增加城市字段（员工可能在异地参保）。

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| PostgreSQL 16 | JSONB存储 | via Docker | 16.x | -- |
| Redis 7 | gocron分布式锁 + 缓存 | via Docker | 7.x | 单实例模式无需锁 |
| Go 1.25.0 | 编译运行 | local | 1.25.0 | -- |
| gocron/v2 | 定时任务 | 需安装 | v2.19.1 | -- |
| gocron-redis-lock/v2 | 分布式锁 | 需安装 | v2.2.1 | 单实例可不用锁 |

**Missing dependencies with no fallback:**
- 无。所有核心依赖要么已安装，要么为新增go get包。

**Missing dependencies with fallback:**
- gocron-redis-lock: 单实例部署时可不使用分布式锁（开发阶段），但生产环境必须使用。

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + testify v1.11.1 |
| Config file | none (SQLite内存数据库) |
| Quick run command | `go test ./internal/socialinsurance/... -count=1 -v` |
| Full suite command | `go test ./... -count=1 -race` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| SOCL-01 | 根据城市+薪资匹配社保基数 | unit | `go test ./internal/socialinsurance/... -run TestCalculateInsurance -v` | Wave 0 |
| SOCL-02 | 生成参保材料PDF | unit | `go test ./internal/socialinsurance/... -run TestGenerateEnrollmentPDF -v` | Wave 0 |
| SOCL-03 | 缴费到期扫描生成提醒 | unit | `go test ./internal/socialinsurance/... -run TestCheckPaymentDue -v` | Wave 0 |
| SOCL-04 | 缴费明细查询 | unit | `go test ./internal/socialinsurance/... -run TestPaymentDetail -v` | Wave 0 |
| SOCL-05 | 导出社保凭证Excel | unit | `go test ./internal/socialinsurance/... -run TestExportExcel -v` | Wave 0 |
| SOCL-06 | 薪资变动提醒预留接口 | unit | `go test ./internal/socialinsurance/... -run TestSalaryChangeReminder -v` | Wave 0 |
| SOCL-07 | 变更历史记录 | unit | `go test ./internal/socialinsurance/... -run TestChangeHistory -v` | Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/socialinsurance/... -count=1`
- **Per wave merge:** `go test ./... -count=1 -race`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/socialinsurance/service_test.go` — covers SOCL-01 through SOCL-07
- [ ] `internal/socialinsurance/repository_test.go` — covers policy CRUD, record CRUD
- [ ] `internal/socialinsurance/scheduler_test.go` — covers payment due scan logic
- [ ] Framework install: `go get github.com/go-co-op/gocron/v2@v2.19.1 && go get github.com/go-co-op/gocron-redis-lock/v2@v2.2.1`

## Sources

### Primary (HIGH confidence)
- go-gorm/datatypes GitHub README — JSONType[T] API, JSONSet, JSONQuery
- go-co-op/gocron GitHub — v2 API, WithDistributedLocker
- go-co-op/gocron-redis-lock GitHub — Redis locker implementation
- 项目已有代码: internal/employee/ (三层架构模式), internal/common/model/base.go (BaseModel), internal/city/model.go (37城市)

### Secondary (MEDIUM confidence)
- [51社保 — 2025年全国各地五险一金基数和缴费比例一览表](https://www.51shebao.com/article/detail/8093) — 城市社保比例数据
- [北京本地宝 — 2025年北京社保缴费标准](http://bj.bendibao.com/zffw/2023726/350882.shtm) — 北京数据
- [上海本地宝 — 上海五险一金缴费标准2025](http://sh.bendibao.com/zffw/2025213/294609.shtm) — 上海数据
- [深圳本地宝 — 深圳社保缴费基数2025](http://bsy.sz.bendibao.com/bsyDetail/636939.html) — 深圳数据

### Tertiary (LOW confidence)
- [Alex Edwards — Using PostgreSQL JSONB with Go](https://www.alexedwards.net/blog/using-postgresql-jsonb) — JSONB最佳实践参考

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — 所有库已在项目go.mod或经版本验证
- Architecture: HIGH — 遵循已建立的三层架构模式，JSONB模式已在Phase 2验证
- Pitfalls: HIGH — 基于PITFALLS.md已有研究 + 中国社保规则确认
- 社保政策数据: MEDIUM — 2025年数据来自权威渠道，但各城市每年调整需管理员手动更新

**Research date:** 2026-04-07
**Valid until:** 2026-05-07 (社保政策可能每年7月调整，数据需及时更新)
