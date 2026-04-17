# Architecture Patterns -- v1.3 New Features Integration

**Domain:** EasyHR v1.3 功能增强 -- 待办中心、考勤管理、薪资增强、社保增强、员工管理增强
**Researched:** 2026-04-17
**Confidence:** HIGH

## Executive Summary

v1.3 在已有模块化单体架构上新增 5 大功能模块/增强。核心挑战不在于新建独立模块，而在于**跨模块数据聚合**（待办中心）和**跨模块联动**（考勤->薪资、离职->社保停缴）的深度集成。现有架构以 adapter 接口模式解决跨模块依赖（salary -> tax/si/employee 通过 adapter），v1.3 需延续此模式并扩展。

关键架构决策：考勤管理作为全新 `internal/attendance` 模块，待办中心作为全新 `internal/todo` 模块，薪资/社保/员工增强在现有模块内扩展。

## Recommended Architecture for v1.3

```
                          +-----------------------------------+
                          |         Clients (H5 only)         |
                          +----------------+------------------+
                                           |
                                  HTTPS / REST API
                                           |
                          +----------------v------------------+
                          |         API Gateway Layer          |
                          | Auth / RateLimit / CORS / Tenant   |
                          +----------------+------------------+
                                           |
        +------------------+----------------+------------------+------------------+
        |                  |                |                  |                  |
  +-----v------+   +------v------+  +------v------+   +------v------+   +------v------+
  |  todo mod  |   | attendance  |  | salary mod  |   |  social mod |   | employee mod |
  | (待办中心)  |   | (考勤管理)  |  | (薪资增强)  |   | (社保增强)  |   | (员工增强)   |
  | Handler    |   | Handler     |  | Handler     |   | Handler     |   | Handler      |
  | Service    |   | Service     |  | Service     |   | Service     |   | Service      |
  | Repository |   | Repository  |  | Repository  |   | Repository  |   | Repository   |
  +-----+------+   +------+------+  +------+------+   +------+------+   +------+------+
        |                  |                |                  |                  |
        +------------------+----------------+------------------+------------------+
                                           |
                                  +--------v--------+
                                  |   PostgreSQL    |
                                  |   (shared DB)   |
                                  +--------+--------+
                                           |
                                  +--------v--------+
                                  |     Redis       |
                                  | (cache/queue)   |
                                  +-----------------+
```

## New vs Modified Components

### 1. 待办中心 -- 全新 `internal/todo` 模块

**职责：** 聚合多个模块的待办事项，提供统一查询、搜索、限时任务追踪、完成率统计。

**需要创建的文件：**

| 文件 | 职责 |
|------|------|
| `internal/todo/model.go` | TodoItem 数据模型（通用待办事项表）|
| `internal/todo/dto.go` | 请求/响应 DTO |
| `internal/todo/repository.go` | 数据访问层 |
| `internal/todo/service.go` | 业务逻辑（聚合、搜索、限时任务计算）|
| `internal/todo/handler.go` | HTTP 端点 |
| `internal/todo/scheduler.go` | 限时任务生成定时任务 |

**新增数据模型：**

```go
// TodoItem 通用待办事项（聚合多个模块来源）
type TodoItem struct {
    model.BaseModel
    Source      string     `gorm:"column:source;type:varchar(30);not null;index;comment:来源模块"` // contract/tax/si/employee/attendance/approval
    SourceID    int64      `gorm:"column:source_id;index;comment:关联来源记录ID"`
    Title       string     `gorm:"column:title;type:varchar(200);not null;comment:待办标题"`
    Initiator   string     `gorm:"column:initiator;type:varchar(50);comment:发起人"`
    Status      string     `gorm:"column:status;type:varchar(20);not null;default:pending;comment:pending/completed/terminated"`
    Priority    int        `gorm:"column:priority;default:0;comment:优先级（0普通/1置顶）"`
    IsTimed     bool       `gorm:"column:is_timed;default:false;comment:是否限时任务"`
    DueDate     *time.Time `gorm:"column:due_date;type:date;comment:截止日期"`
    CompletedAt *time.Time `gorm:"column:completed_at;comment:完成时间"`
    ModuleRoute string     `gorm:"column:module_route;type:varchar(100);comment:跳转路由路径"`
}

// TimedTaskRule 限时任务规则（系统预置）
type TimedTaskRule struct {
    model.BaseModel
    TaskType    string `gorm:"column:task_type;type:varchar(50);not null;uniqueIndex;comment:任务类型"`
    Name        string `gorm:"column:name;type:varchar(100);not null;comment:任务名称"`
    CronExpr    string `gorm:"column:cron_expr;type:varchar(50);comment:Cron表达式（生成周期）"`
    StartOffset int    `gorm:"column:start_offset;comment:开始偏移天数"`
    DueOffset   int    `gorm:"column:due_offset;comment:截止偏移天数"`
    IsEnabled   bool   `gorm:"column:is_enabled;default:true;comment:是否启用"`
}
```

**与现有模块的集成点：**

待办中心作为聚合层，不直接写入数据，而是通过以下方式收集待办：

1. **各模块主动推送** -- 现有模块在产生待办事项时调用 `todo.Service.CreateItem()`
2. **定时扫描** -- `todo.Scheduler` 定期扫描各模块数据生成限时任务

**需要修改的现有文件：**

| 文件 | 修改内容 |
|------|---------|
| `cmd/server/main.go` | 注册 todo 模块路由、定时任务、AutoMigrate |
| `internal/socialinsurance/service.go` | 在 `CreateStopReminder` / `CheckPaymentDueReminders` 中同步创建 TodoItem |
| `internal/tax/service.go` | 在 `CheckDeclarationReminders` 中同步创建 TodoItem |
| `internal/employee/contract_service.go` | 合同到期前创建 TodoItem |
| `internal/employee/offboarding_service.go` | 离职审批创建 TodoItem |
| `internal/dashboard/service.go` | 可选：待办卡片改为从 todo 模块读取 |

**关键架构决策：** 待办中心不替代各模块的 reminder 系统（社保 `Reminder` 表、个税 `TaxReminder` 表保留），而是作为**聚合展示层**统一读取。限时任务（如"3月社保缴费"）由 `todo` 模块的定时任务生成。

**数据流：**

```
各业务模块业务操作
    |
    v
todo.Service.CreateItem()  -- 创建待办记录
    |
    v
todo_items 表
    |
    v
GET /api/v1/todo  -- 前端查询（搜索/筛选/分页）
```

### 2. 考勤管理 -- 全新 `internal/attendance` 模块

**职责：** 打卡规则引擎（3种排班模式）、审批流引擎（多类型审批）、出勤月报统计。

**需要创建的文件：**

| 文件 | 职责 |
|------|------|
| `internal/attendance/model.go` | AttendanceRule, Shift, Schedule, ClockRecord, Approval, AttendanceMonthly 数据模型 |
| `internal/attendance/dto.go` | 请求/响应 DTO |
| `internal/attendance/repository.go` | 数据访问层 |
| `internal/attendance/rule_engine.go` | 打卡规则引擎（3种模式计算）|
| `internal/attendance/approval_service.go` | 审批流服务 |
| `internal/attendance/service.go` | 核心业务逻辑 |
| `internal/attendance/handler.go` | HTTP 端点 |
| `internal/attendance/adapter.go` | 跨模块接口定义 |

**新增数据模型：**

```go
// AttendanceRule 打卡规则（每企业一条）
type AttendanceRule struct {
    model.BaseModel
    Mode         string  `gorm:"column:mode;type:varchar(20);not null;comment:模式（fixed/scheduled/free）"`
    WorkDays     string  `gorm:"column:work_days;type:varchar(50);comment:上班日（JSON数组，如[1,2,3,4,5]）"`
    WorkStart    string  `gorm:"column:work_start;type:varchar(5);comment:上班时间（HH:mm）"`
    WorkEnd      string  `gorm:"column:work_end;type:varchar(5);comment:下班时间（HH:mm）"`
    Location     string  `gorm:"column:location;type:varchar(200);comment:打卡位置"`
    ClockMethod  string  `gorm:"column:clock_method;type:varchar(20);default:click;comment:打卡方式（click/photo）"`
    Holidays     datatypes.JSON `gorm:"column:holidays;type:jsonb;comment:特殊日期（不用打卡的日期）"`
}

// Shift 班次（排班模式使用）
type Shift struct {
    model.BaseModel
    Name      string `gorm:"column:name;type:varchar(50);not null;comment:班次名称"`
    WorkStart string `gorm:"column:work_start;type:varchar(5);not null;comment:上班时间"`
    WorkEnd   string `gorm:"column:work_end;type:varchar(5);not null;comment:下班时间"`
}

// Schedule 排班记录（排班模式下每个员工每天的班次）
type Schedule struct {
    model.BaseModel
    EmployeeID int64      `gorm:"column:employee_id;not null;index;comment:员工ID"`
    ShiftID    *int64     `gorm:"column:shift_id;comment:班次ID（null=休息）"`
    Date       time.Time  `gorm:"column:date;type:date;not null;comment:日期"`
}

// ClockRecord 打卡记录
type ClockRecord struct {
    model.BaseModel
    EmployeeID int64      `gorm:"column:employee_id;not null;index;comment:员工ID"`
    ClockTime  time.Time  `gorm:"column:clock_time;not null;comment:打卡时间"`
    ClockType  string     `gorm:"column:clock_type;type:varchar(10);not null;comment:in/out"`
    PhotoURL   string     `gorm:"column:photo_url;type:varchar(500);comment:打卡照片"`
}

// Approval 审批记录（请假/加班/补卡/调班/出差/外出）
type Approval struct {
    model.BaseModel
    EmployeeID     int64          `gorm:"column:employee_id;not null;index;comment:申请员工ID"`
    ApprovalType   string         `gorm:"column:approval_type;type:varchar(20);not null;index;comment:类型"`
    StartTime      time.Time      `gorm:"column:start_time;not null;comment:开始时间"`
    EndTime        time.Time      `gorm:"column:end_time;not null;comment:结束时间"`
    Duration       float64        `gorm:"column:duration;comment:时长（小时）"`
    Reason         string         `gorm:"column:reason;type:varchar(500);comment:事由"`
    LeaveType      string         `gorm:"column:leave_type;type:varchar(20);comment:请假类型（事假/病假/调休/年假/婚假/产假/陪产假）"`
    Status         string         `gorm:"column:status;type:varchar(20);not null;default:pending;comment:pending/approved/rejected"`
    ApproverID     *int64         `gorm:"column:approver_id;comment:审批人ID"`
    ApprovedAt     *time.Time     `gorm:"column:approved_at;comment:审批时间"`
    AttachmentURLs datatypes.JSON `gorm:"column:attachment_urls;type:jsonb;comment:附件URL列表"`
}

// AttendanceMonthly 出勤月报（每员工每月汇总）
type AttendanceMonthly struct {
    model.BaseModel
    EmployeeID      int64   `gorm:"column:employee_id;not null;index;comment:员工ID"`
    Year            int     `gorm:"column:year;not null;comment:年份"`
    Month           int     `gorm:"column:month;not null;comment:月份"`
    RequiredDays    float64 `gorm:"column:required_days;comment:应出勤天数"`
    ActualDays      float64 `gorm:"column:actual_days;comment:实际出勤天数"`
    OvertimeWeekday float64 `gorm:"column:overtime_weekday;comment:工作日加班（小时）"`
    OvertimeWeekend float64 `gorm:"column:overtime_weekend;comment:双休日加班（小时）"`
    OvertimeHoliday float64 `gorm:"column:overtime_holiday;comment:节假日加班（小时）"`
    AbsentDays      float64 `gorm:"column:absent_days;comment:缺勤天数"`
}
```

**与现有模块的集成点：**

| 集成点 | 方向 | 说明 |
|--------|------|------|
| attendance -> salary | 审批通过的请假数据流入薪资计算 | `SIDeductionProvider` 模式，attendance 提供 `AttendanceProvider` 接口 |
| attendance -> salary | 出勤天数影响基本工资计算 | `RequiredDays` / `ActualDays` 影响日薪计算 |
| attendance -> salary | 加班时长计算加班费 | `OvertimeWeekday/Weekend/Holiday` 三档费率 |
| employee -> attendance | 员工列表用于排班和打卡 | `EmployeeProvider` 复用现有接口 |
| attendance -> todo | 审批待办推送到待办中心 | 审批创建时推送 TodoItem |

**关键架构决策 -- 考勤与薪资的关联：**

PRD 定义的薪资算法依赖考勤数据：
- `计薪天数 = 实际出勤天数 + 法定节假日天数 + 带薪假天数`
- `病假工资 = 基本工资 * 病假系数`
- `加班费 = (基本工资 / 应出勤 / 8小时) * 加班时长 * 费率`

建议通过 adapter 接口解耦：

```go
// attendance/adapter.go -- 考勤模块对外接口
type AttendanceProvider interface {
    GetMonthlyAttendance(orgID, employeeID int64, year, month int) (*MonthlyAttendance, error)
    GetApprovedLeaves(orgID, employeeID int64, year, month int) ([]LeaveRecord, error)
    GetApprovedOvertime(orgID, employeeID int64, year, month int) ([]OvertimeRecord, error)
}

type LeaveRecord struct {
    LeaveType string
    Duration  float64 // 小时
    StartTime time.Time
    EndTime   time.Time
}

type OvertimeRecord struct {
    OvertimeType string // weekday/weekend/holiday
    Duration     float64 // 小时
}
```

salary 模块在 `CalculatePayroll` 时通过此接口获取考勤数据，替代现有的 Excel 导入方式。

### 3. 薪资管理增强 -- 扩展 `internal/salary` 模块

**需要新增的文件：**

| 文件 | 职责 |
|------|------|
| `internal/salary/salary_adjustment.go` | 调薪/普调模型和逻辑 |
| `internal/salary/salary_adjustment_dto.go` | 调薪请求/响应 DTO |
| `internal/salary/salary_adjustment_handler.go` | 调薪 API 端点 |
| `internal/salary/performance.go` | 绩效系数模型和逻辑 |
| `internal/salary/dashboard_service.go` | 薪资数据看板 |
| `internal/salary/dashboard_handler.go` | 看板 API 端点 |
| `internal/salary/tax_upload.go` | 个税上传解析逻辑 |

**新增数据模型：**

```go
// SalaryAdjustment 调薪记录
type SalaryAdjustment struct {
    model.BaseModel
    EmployeeID    int64   `gorm:"column:employee_id;not null;index;comment:员工ID"`
    DepartmentID  *int64  `gorm:"column:department_id;index;comment:部门ID（普调时使用）"`
    Type          string  `gorm:"column:type;type:varchar(20);not null;comment:individual/mass"`
    EffectiveDate string  `gorm:"column:effective_date;type:varchar(7);not null;comment:生效月份（YYYY-MM）"`
    // 调整项（JSON 存储各薪资项调整值）
    Adjustments   datatypes.JSON `gorm:"column:adjustments;type:jsonb;comment:调整项明细"`
    Status        string  `gorm:"column:status;type:varchar(20);not null;default:pending;comment:pending/active/expired"`
}

// PerformanceCoefficient 绩效系数
type PerformanceCoefficient struct {
    model.BaseModel
    EmployeeID int64    `gorm:"column:employee_id;not null;index;comment:员工ID"`
    Year       int      `gorm:"column:year;not null;comment:年份"`
    Month      int      `gorm:"column:month;not null;comment:月份"`
    Coefficient float64 `gorm:"column:coefficient;type:decimal(5,4);not null;default:1.0;comment:绩效系数（0.0-1.0）"`
}
```

**需要修改的现有文件：**

| 文件 | 修改内容 |
|------|---------|
| `internal/salary/calculator.go` | 新增病假工资计算、绩效系数计算、加班费计算函数 |
| `internal/salary/service.go` | `CalculatePayroll` 方法增强：集成考勤数据、绩效系数、调薪生效期 |
| `internal/salary/handler.go` | 新增调薪/普调/绩效系数/看板/个税上传/工资条推送增强端点 |
| `internal/salary/model.go` | `PayrollRecord` 增加 `gp_fund_deduction`（公积金扣除）字段 |
| `internal/salary/adapter.go` | 新增 `AttendanceProvider` 接口 |
| `internal/salary/dto.go` | 新增调薪/看板相关 DTO |

**薪资计算增强数据流：**

```
salary.Service.CalculatePayroll(orgID, year, month)
    |
    |-- 1. 获取在职员工列表（现有 EmployeeProvider）
    |-- 2. 获取员工薪资项（现有 repo）
    |-- 3. 获取调薪记录（新增，按生效期取最新）
    |-- 4. 获取绩效系数（新增）
    |-- 5. 获取考勤月报（新增 AttendanceProvider）
    |       |-- 出勤天数 -> 计薪天数计算
    |       |-- 病假时长 -> 病假工资计算
    |       |-- 加班时长 -> 加班费计算
    |-- 6. 计算基本工资 = 基本工资/应出勤*计薪天数 + 病假工资
    |-- 7. 计算绩效工资 = 绩效工资*绩效系数
    |-- 8. 获取社保扣款（现有 SIProvider）
    |-- 9. 计算个税（现有 TaxProvider）
    |-- 10. 汇总 = 税前 - 社保 - 公积金 - 个税 - 其他扣款 = 实发
    |-- 11. 写入 PayrollRecord + PayrollItem
```

### 4. 社保公积金增强 -- 扩展 `internal/socialinsurance` 模块

**需要修改/新增的文件：**

| 文件 | 职责 |
|------|------|
| `internal/socialinsurance/model.go` | `SocialInsuranceRecord` 增加缴费状态字段 |
| `internal/socialinsurance/payment_service.go` | 缴费渠道管理、欠缴状态更新逻辑 |
| `internal/socialinsurance/dashboard_service.go` | 社保数据看板 |
| `internal/socialinsurance/dashboard_handler.go` | 看板 API 端点 |
| `internal/socialinsurance/dto.go` | 新增缴费/看板 DTO |

**现有模型变更：**

`SocialInsuranceRecord` 状态从 3 种扩展为 5 种：

```go
// 现有
SIStatusPending = "pending"  // 待参保
SIStatusActive  = "active"   // 参保中
SIStatusStopped = "stopped"  // 停缴

// 新增
SIStatusArrears    = "arrears"    // 欠缴 -- 上月有应缴未缴，当月25日后
SIStatusTransferred = "transferred" // 已转出 -- 完成减员转出
```

增加缴费相关字段：

```go
// SocialInsuranceRecord 新增字段
PaymentStatus  string  `gorm:"column:payment_status;type:varchar(20);default:unpaid;comment:缴费状态（unpaid/paid）"`
PaymentMonth   string  `gorm:"column:payment_month;type:varchar(7);comment:最近缴费月份"`
LastPaidAt      *time.Time `gorm:"column:last_paid_at;comment:最近缴费时间"`
```

**需要修改的现有文件：**

| 文件 | 修改内容 |
|------|---------|
| `internal/socialinsurance/model.go` | 增加 `PaymentStatus`, `PaymentMonth`, `LastPaidAt` 字段，新增状态常量 |
| `internal/socialinsurance/service.go` | 增加缴费渠道方法、欠缴检测逻辑、增减员优化 |
| `internal/socialinsurance/handler.go` | 新增缴费/看板端点 |
| `internal/socialinsurance/scheduler.go` | 增加欠缴状态自动更新定时任务 |
| `internal/socialinsurance/employee_adapter.go` | 增加按部门查询员工的方法 |

**社保状态机变更：**

```
现有：
pending -> active -> stopped

增强：
pending -> active -> stopped (停缴)
                 -> arrears (欠缴) -> active (补缴后)
                 -> transferred (已转出)

active 正常流转:
active + 当月15日前未缴 -> active (待缴，提示)
active + 上月欠缴 + 当月25日后 -> arrears
active + 离职减员完成 -> transferred
```

### 5. 员工管理增强 -- 扩展 `internal/employee` 模块

**需要新增的文件：**

| 文件 | 职责 |
|------|------|
| `internal/employee/department_model.go` | 部门/组织架构模型 |
| `internal/employee/department_repository.go` | 部门数据访问 |
| `internal/employee/department_service.go` | 组织架构业务逻辑 |
| `internal/employee/department_handler.go` | 组织架构 API |
| `internal/employee/registration_model.go` | 员工信息登记模型 |
| `internal/employee/registration_service.go` | 员工信息登记逻辑 |
| `internal/employee/registration_handler.go` | 信息登记 API |
| `internal/employee/dashboard_service.go` | 员工数据看板 |
| `internal/employee/dashboard_handler.go` | 看板 API |

**新增数据模型：**

```go
// Department 部门
type Department struct {
    model.BaseModel
    Name     string  `gorm:"column:name;type:varchar(100);not null;comment:部门名称"`
    ParentID *int64  `gorm:"column:parent_id;index;comment:上级部门ID（null=顶级）"`
    SortOrder int    `gorm:"column:sort_order;default:0;comment:排序"`
}

// EmployeeRegistration 员工信息登记
type EmployeeRegistration struct {
    model.BaseModel
    EmployeeID        int64          `gorm:"column:employee_id;not null;index;comment:员工ID"`
    Status            string         `gorm:"column:status;type:varchar(20);not null;default:draft;comment:draft/submitted"`
    PhoneEncrypted    string         `gorm:"column:phone_encrypted;type:varchar(200);comment:加密手机号"`
    PhoneHash         string         `gorm:"column:phone_hash;type:varchar(64);index;comment:手机号哈希"`
    Address           string         `gorm:"column:address;type:varchar(500);comment:居住地址"`
    IDCardFrontURL    string         `gorm:"column:id_card_front_url;type:varchar(500);comment:身份证正面"`
    IDCardBackURL     string         `gorm:"column:id_card_back_url;type:varchar(500);comment:身份证背面"`
    BankCardFrontURL  string         `gorm:"column:bank_card_front_url;type:varchar(500);comment:银行卡正面"`
    BankCardBackURL   string         `gorm:"column:bank_card_back_url;type:varchar(500);comment:银行卡背面"`
    EducationCertURLs datatypes.JSON `gorm:"column:education_cert_urls;type:jsonb;comment:学历证书URL列表"`
    EmergencyContact  string         `gorm:"column:emergency_contact;type:varchar(50);comment:紧急联系人"`
    EmergencyRelation string         `gorm:"column:emergency_relation;type:varchar(20);comment:紧急联系人关系"`
    EmergencyPhoneEnc string         `gorm:"column:emergency_phone_enc;type:varchar(200);comment:加密紧急联系人电话"`
    EmergencyPhoneHsh string         `gorm:"column:emergency_phone_hsh;type:varchar(64);comment:紧急联系人电话哈希"`
}
```

**需要修改的现有文件：**

| 文件 | 修改内容 |
|------|---------|
| `internal/employee/model.go` | `Employee` 增加 `DepartmentID` 字段 |
| `internal/employee/dto.go` | 增加 `DepartmentID` 相关参数 |
| `internal/employee/repository.go` | 增加按部门筛选查询 |
| `internal/employee/offboarding_service.go` | 离职联动增强：完成离职后触发社保减员提醒 |
| `cmd/server/main.go` | 注册新路由、AutoMigrate 新模型 |

**员工模型变更：**

```go
// Employee 新增字段
DepartmentID *int64 `gorm:"column:department_id;index;comment:所属部门ID" json:"department_id"`
```

## Component Boundaries Summary

### v1.3 完整模块依赖图

```
                    +-----------+
                    |   user    |
                    +-----+-----+
                          |
               +----------+----------+
               |                     |
        +------v------+       +------v------+
        |  employee   |       |    todo     | <-- NEW
        |  (增强)      |       |  (待办中心)  |
        +------+------+       +------+------+
               |                     |
       +-------+-------+            |
       |       |       |            |
  +----v--+ +--v---+ +-v--------+  |
  |social | |salary| |attendance|  |
  |(增强)  | |(增强) | | (NEW)    |  |
  +----+--+ +--+---+ +----+-----+  |
       |       |           |        |
       +-------+-----------+--------+
               |
        +------v------+
        |   finance   |
        +-------------+
```

### 跨模块依赖接口清单

| 接口 | 定义位置 | 实现位置 | 调用方 |
|------|---------|---------|--------|
| `EmployeeProvider` | `salary/adapter.go` | `salary/employee_adapter.go` | salary |
| `TaxProvider` | `salary/adapter.go` | `salary/tax_adapter.go` | salary |
| `SIDeductionProvider` | `salary/adapter.go` | `salary/si_adapter.go` | salary |
| `BaseAdjustmentProvider` | `salary/adapter.go` | `socialinsurance/service.go` | salary |
| `SocialInsuranceEventHandler` | `employee/offboarding_service.go` | `socialinsurance/service.go` | employee |
| `EmployeeQuerier` | `socialinsurance/employee_adapter.go` | `socialinsurance/employee_adapter.go` | socialinsurance |
| **`AttendanceProvider`** (NEW) | `salary/adapter.go` | `attendance/adapter_impl.go` | salary |
| **`TodoCreator`** (NEW) | `todo/adapter.go` | `todo/service.go` | 各业务模块 |
| **`DepartmentProvider`** (NEW) | `salary/adapter.go` | `employee/department_service.go` | salary (普调按部门) |

## Data Flow for Key v1.3 Scenarios

### 场景 1：待办中心 -- 限时任务生成

```
gocron 定时任务 (每日 08:00)
    |
    v
todo.Scheduler.CheckTimedTasks()
    |
    |-- 扫描 TimedTaskRule 配置
    |-- 检查是否满足生成条件（如每月1日生成社保缴费任务）
    |-- 查询各模块数据确认是否需要生成
    |       |-- socialinsurance: 查询 active 记录 -> 社保缴费任务
    |       |-- tax: 查询当月申报状态 -> 个税申报任务
    |       |-- employee/contract: 查询到期合同 -> 合同续签任务
    |-- todo.Repository.BatchCreate(todoItems)
```

### 场景 2：考勤 -> 薪资联动

```
attendance.ApprovalService.Approve(approvalID)
    |
    |-- 更新 Approval 状态 = approved
    |-- 如果是请假审批:
    |       |-- 记录请假时长到 AttendanceMonthly
    |-- 如果是加班审批:
    |       |-- 记录加班时长到 AttendanceMonthly
    |-- 触发异步任务: 更新月报统计

salary.Service.CalculatePayroll(orgID, year, month)
    |
    |-- attendanceProvider.GetMonthlyAttendance()
    |       |-- 获取 actualDays, requiredDays, overtime
    |-- attendanceProvider.GetApprovedLeaves()
    |       |-- 获取病假/事假/带薪假明细
    |-- 计算基本工资 (含病假系数)
    |-- 计算加班费 (三档费率)
    |-- 计算绩效工资 * 绩效系数
    |-- ... (其余现有逻辑)
```

### 场景 3：离职联动（增强）

```
employee.OffboardingService.CompleteOffboarding()
    |
    |-- 更新 Offboarding 状态 = completed
    |-- 更新 Employee 状态 = resigned
    |
    |-- siHandler.OnEmployeeResigned()  (现有)
    |       |-- socialinsurance: 创建停缴提醒
    |       |-- todo: 创建"社保减员"待办
    |
    |-- todoCreator.CreateItem("离职交接", ...)  (新增)
    |
    |-- attendance: 标记员工排班为无效 (新增)
```

### 场景 4：调薪生效期到薪资计算

```
salary.SalaryAdjustmentService.CreateAdjustment()
    |
    |-- 创建 SalaryAdjustment 记录
    |-- 设置 effective_date
    |-- 可选: todoCreator.CreateItem("调薪生效提醒")

salary.Service.CalculatePayroll(orgID, year, month)
    |
    |-- 查询当月生效的调薪记录
    |-- 合并调薪数据到薪资项:
    |       |-- 基本工资: 取调薪后的值
    |       |-- 补贴/奖金/年终奖: 取调薪中的值
    |       |-- 扣款: 取调薪中的值
    |-- 正常核算流程
```

## Build Order for v1.3

基于模块依赖关系和功能耦合度，建议以下构建顺序：

### Phase A: 员工管理增强（部门 + 信息登记 + 看板）

**先行的原因：**
- 部门是薪资普调和社保查询的基础维度
- 信息登记是员工数据的补充，不影响其他模块
- 看板为独立查询，无写依赖

**依赖关系：** 仅依赖现有 employee 模块，不依赖其他新模块。

### Phase B: 考勤管理（打卡规则 + 审批流 + 月报）

**第二步的原因：**
- 考勤是薪资计算增强的上游数据源
- 审批流产生的请假/加班数据直接影响薪资
- 打卡规则引擎独立性强，可独立开发

**依赖关系：** 依赖 employee（员工列表）、Phase A（部门）。

### Phase C: 薪资增强（调薪/普调 + 绩效系数 + 个税上传 + 看板 + 加班费/病假）

**第三步的原因：**
- 需要考勤模块提供出勤数据
- 需要部门模型支持普调
- 是 v1.3 最复杂的跨模块聚合点

**依赖关系：** 依赖 Phase A（部门）、Phase B（考勤数据）。

### Phase D: 社保增强（缴费状态 + 增减员优化 + 看板）

**可与 Phase C 并行的原因：**
- 社保状态管理相对独立
- 与薪资的集成点（社保扣款）已存在
- 增减员优化是现有功能的增强

**依赖关系：** 依赖 employee（增减员联动）。

### Phase E: 待办中心（事项聚合 + 限时任务 + 完成率）

**最后构建的原因：**
- 待办中心是聚合展示层，需要各模块先实现待办生成逻辑
- 限时任务规则依赖社保/个税/合同模块的数据
- 各模块先完成核心功能，再接入待办推送

**依赖关系：** 依赖所有其他模块（读取/聚合）。

### Build Order Visualization

```
Week:  1   2   3   4   5   6   7   8
       +---+---+---+---+---+---+---+---+
A:     |█████████████|                   员工增强（部门+登记+看板）
B:         |████████████████|            考勤管理（规则+审批+月报）
C:                 |█████████████████|   薪资增强（调薪+绩效+看板）
D:                 |████████████|        社保增强（状态+增减员+看板）
E:                         |█████████████| 待办中心
       +---+---+---+---+---+---+---+---+
```

## API Routes Addition

### 待办中心

```
GET    /api/v1/todo                    -- 查询待办列表（分页/搜索/筛选）
GET    /api/v1/todo/timed              -- 查询限时任务
GET    /api/v1/todo/completion-rate    -- 完成率统计（环形图数据）
PUT    /api/v1/todo/:id/complete       -- 标记完成
PUT    /api/v1/todo/:id/terminate      -- 终止任务
PUT    /api/v1/todo/:id/pin            -- 置顶/取消置顶
GET    /api/v1/todo/export             -- 导出待办 Excel
```

### 考勤管理

```
GET    /api/v1/attendance/rule         -- 查询打卡规则
PUT    /api/v1/attendance/rule         -- 设置打卡规则
POST   /api/v1/attendance/shifts       -- 创建班次（排班模式）
GET    /api/v1/attendance/shifts       -- 班次列表
PUT    /api/v1/attendance/schedule     -- 设置排班
GET    /api/v1/attendance/today        -- 今日打卡实况
POST   /api/v1/attendance/clock        -- 打卡（员工端）
GET    /api/v1/attendance/approvals    -- 审批列表
POST   /api/v1/attendance/approvals    -- 创建审批申请
PUT    /api/v1/attendance/approvals/:id/approve  -- 审批通过
PUT    /api/v1/attendance/approvals/:id/reject   -- 审批驳回
GET    /api/v1/attendance/monthly     -- 出勤月报
PUT    /api/v1/attendance/monthly/:id -- 手动修改月报数据
GET    /api/v1/attendance/export       -- 导出考勤 Excel
```

### 薪资增强

```
POST   /api/v1/salary/adjustment       -- 创建调薪（单人）
POST   /api/v1/salary/mass-adjustment  -- 普调（按部门）
GET    /api/v1/salary/adjustments      -- 调薪记录列表
PUT    /api/v1/salary/performance      -- 设置绩效系数
GET    /api/v1/salary/dashboard        -- 薪资数据看板
POST   /api/v1/salary/tax-upload       -- 个税上传解析
POST   /api/v1/salary/slip/send-all    -- 全员发工资条（当月）
POST   /api/v1/salary/slip/send        -- 选定员工发工资条（已有，增强月份选择）
```

### 社保增强

```
GET    /api/v1/social-insurance/dashboard           -- 社保数据看板
PUT    /api/v1/social-insurance/records/:id/pay      -- 标记缴费
POST   /api/v1/social-insurance/batch-pay            -- 批量标记缴费
GET    /api/v1/social-insurance/arrears               -- 欠缴列表
PUT    /api/v1/social-insurance/enroll/:id            -- 增员增强（支持起始月份限制）
PUT    /api/v1/social-insurance/stop/:id              -- 减员增强（转出日期联动）
```

### 员工管理增强

```
GET    /api/v1/departments                        -- 部门列表（树形）
POST   /api/v1/departments                        -- 创建部门
PUT    /api/v1/departments/:id                     -- 编辑部门
DELETE /api/v1/departments/:id                     -- 删除部门
GET    /api/v1/employees/org-chart                 -- 组织架构可视化
POST   /api/v1/employees/:id/registration          -- 创建信息登记
PUT    /api/v1/employees/:id/registration           -- 更新信息登记
POST   /api/v1/employees/:id/registration/invite    -- 邀请员工填写
GET    /api/v1/employees/dashboard                  -- 员工数据看板
GET    /api/v1/employees/roster                     -- 花名册（增强版）
```

## Scalability Considerations

| Concern | 当前影响 | 优化方案 |
|---------|---------|---------|
| 待办聚合查询 | 跨 6+ 模块扫描 | 各模块写入 todo_items 统一表，定时预聚合 |
| 考勤打卡并发 | 每日早晚高峰 50 人同时打卡 | Redis 缓存 + 批量写入 |
| 月报统计 | 全员月度聚合 | 定时任务预计算 attendance_monthly 表 |
| 薪资核算批量 | 50 人一次性计算含考勤/社保/个税 | 并行计算 (errgroup)，复用现有模式 |
| 审批流并发 | 多人同时提交审批 | PostgreSQL 行锁，乐观并发控制 |

## Database Migration Impact

### 新增表

| 表名 | 所属模块 | 预估行数 |
|------|---------|---------|
| `todo_items` | todo | 中（每月 10-50 条/企业）|
| `timed_task_rules` | todo | 小（10-20 条全局预置）|
| `attendance_rules` | attendance | 小（每企业 1 条）|
| `shifts` | attendance | 小（每企业 3-5 条）|
| `schedules` | attendance | 中（50人*30天=1500条/月）|
| `clock_records` | attendance | 中（50人*2次*22天=2200条/月）|
| `approvals` | attendance | 中（每月 20-100 条）|
| `attendance_monthly` | attendance | 中（50 条/月）|
| `salary_adjustments` | salary | 小（每月 5-20 条）|
| `performance_coefficients` | salary | 中（50 条/月）|
| `departments` | employee | 小（每企业 5-20 条）|
| `employee_registrations` | employee | 小（每员工 1 条）|

### 变更表

| 表名 | 变更内容 |
|------|---------|
| `employees` | 新增 `department_id` 字段 |
| `social_insurance_records` | 新增 `payment_status`, `payment_month`, `last_paid_at` 字段；状态枚举扩展 |

## Sources

- `/Users/wencai/github/EasyHR/cmd/server/main.go` -- 依赖注入和路由注册 (HIGH)
- `/Users/wencai/github/EasyHR/internal/employee/model.go` -- 员工数据模型 (HIGH)
- `/Users/wencai/github/EasyHR/internal/salary/model.go` -- 薪资数据模型 (HIGH)
- `/Users/wencai/github/EasyHR/internal/salary/calculator.go` -- 薪资计算逻辑 (HIGH)
- `/Users/wencai/github/EasyHR/internal/salary/adapter.go` -- 跨模块接口定义 (HIGH)
- `/Users/wencai/github/EasyHR/internal/socialinsurance/model.go` -- 社保数据模型 (HIGH)
- `/Users/wencai/github/EasyHR/internal/socialinsurance/service.go` -- 社保业务逻辑 (HIGH)
- `/Users/wencai/github/EasyHR/internal/dashboard/service.go` -- 现有待办聚合逻辑 (HIGH)
- `/Users/wencai/github/EasyHR/internal/dashboard/repository.go` -- 跨模块查询模式 (HIGH)
- `/Users/wencai/github/EasyHR/internal/employee/offboarding_service.go` -- 离职联动模式 (HIGH)
- `/Users/wencai/github/EasyHR/internal/common/middleware/tenant.go` -- 多租户模式 (HIGH)
- `/Users/wencai/github/EasyHR/internal/common/model/base.go` -- 基础模型模式 (HIGH)
- `/Users/wencai/github/EasyHR/prd1.1.md` -- PRD 需求定义 (HIGH)
- `/Users/wencai/github/EasyHR/.planning/research/ARCHITECTURE.md` -- 已有架构文档 (HIGH)
