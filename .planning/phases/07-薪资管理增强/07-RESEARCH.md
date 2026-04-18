# Phase 7: 薪资管理增强 - Research

**Researched:** 2026-04-18
**Domain:** Go backend + Vue3 frontend 薪资增强模块，集成考勤数据驱动薪资核算
**Confidence:** HIGH

## Summary

Phase 7 是 v1.3 的薪资增强模块，核心是三件事：(1) 将考勤数据（出勤天数、病假、加班时长）集成到薪资算法中；(2) 新增调薪/普调、绩效系数、个税上传三个管理功能；(3) 新增薪资看板和工资条发送。Phase 5 已建立 Department 模型，Phase 6 已建立 AttendanceMonthly 出勤数据，这两个是本阶段的前置依赖。

现有代码库已有 `PayrollRecord`/`PayrollItem`/`SalaryItem` 模型、`CalculatePayroll` 骨架、`PayrollSlip` Token 模式、`excelize` Excel 导入模式、`errgroup` 并发聚合模式。Phase 7 需要新增 4 张表（salary_adjustments、performance_coefficients、salary_slip_send_logs、sick_leave_policies）、扩展 `AttendanceProvider` 接口、增强 `CalculatePayroll` 算法。asynq 尚未在项目中使用，需要新增。

---

## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-SAL-DASH-01:** 4张纯数字卡片，与员工数据看板风格一致，环比仅展示已确认月份
- **D-SAL-ADJ-01:** INSERT ONLY + 月份只读保护，调薪记录 INSERT 新行禁止 UPDATE，confirmed/paid 月份的 PayrollRecord 保持不变
- **D-SAL-ADJ-02:** 绩效系数使用独立月度表：performance_coefficients（employee_id + year_month + coefficient），默认值 1.0
- **D-SAL-TAX-01:** 员工姓名精确匹配为主（精确匹配 → 模糊匹配 → 跳过），无法匹配行记录在错误日志
- **D-SAL-TAX-02:** 部分成功提示策略：部分匹配成功则成功，列出未匹配行；全部失败则整体失败
- **D-SAL-TAX-03:** 上传成功后自动更新当月工资表个税字段，并标记工资表状态回 draft
- **D-SAL-ATT-01:** 计薪天数 = 实际出勤 + 法定节假日 + 带薪假天数；基本工资 = 基本工资项/应出勤 × 计薪天数
- **D-SAL-ATT-02:** 病假工资 = 基本工资 × 病假系数，系数存于 sick_leave_policies 表，初期仅支持北上广深
- **D-SAL-ATT-03:** 加班费按法定系数（工作日 150%/双休日 200%/节假日 300%），Phase 6 已决定加班时长按 0.5h 取整
- **D-SAL-ATT-04:** AttendanceProvider.GetMonthlyAttendance() 返回：actual_days、should_attend、overtime_hours、paid_leave_days、legal_holiday_days、sick_leave_days
- **D-SAL-SLIP-01:** 发送通道优先级：微信小程序 → 短信 → H5 链接
- **D-SAL-SLIP-02:** 全员发送使用 asynq 队列后台处理
- **D-SAL-DATA-01:** draft/calculated 可重新编辑；confirmed/paid 禁止修改，需输入解锁码（企业主手机验证码）才能临时解锁
- **D-SAL-DATA-02:** 解锁操作必须记录审计日志

### Claude's Discretion
- 薪资数据看板卡片的排列顺序和具体样式细节
- 普调按部门选择的具体 UI（多选部门？全选按钮？）
- 个税 Excel 列名识别算法（支持哪些别名映射）
- 计薪天数中"法定节假日"的来源（AttendanceRule.Holidays JSON）
- 工资条 H5 页面具体样式和内容结构
- sick_leave_policies 表的初始数据（北上广深各档位系数）

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope.

---

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SAL-01 | 薪资数据看板展示当月应发总额，环比上月百分比 | 扩展 Dashboard 聚合，errgroup 并发查询 |
| SAL-02 | 薪资数据看板展示当月实发工资总额，环比上月百分比 | 同上 |
| SAL-03 | 薪资数据看板展示当月社保公积金总额，环比上月百分比 | 同上 |
| SAL-04 | 薪资数据看板展示当月个税总额，环比上月百分比 | 同上 |
| SAL-05 | 管理员可对选定员工进行调薪（岗位工资/补贴/奖金/年终奖/其他扣除/生效期限） | SalaryAdjustment INSERT ONLY 表设计 |
| SAL-06 | 管理员可对选定部门进行普调（调整规则同调薪） | Department 模型 + 批量调整逻辑 |
| SAL-07 | 调薪数据按生效期限自动应用于当月工资核算 | SalaryItem.effective_month 自然生效 |
| SAL-08 | 管理员可上传个税 Excel 文件，系统自动抓取关键字段 | excelize 解析，姓名匹配逻辑 |
| SAL-09 | 个税上传成功后自动更新当月工资表中个税数据 | PayrollRecord.tax 字段更新 + 状态回 draft |
| SAL-10 | 个税上传失败时提示上传失败原因或重新上传 | 错误行日志 + 前端提示 |
| SAL-11 | 管理员可为选定员工设置绩效系数（0%-100%） | PerformanceCoefficient 表 + 滑块 UI |
| SAL-12 | 绩效系数自动挂钩绩效工资计算（绩效工资 = 绩效工资标准 × 绩效系数） | 扩展 CalculatePayroll 读取绩效系数 |
| SAL-13 | 基本工资按计薪天数计算，计薪天数=实际出勤+法定节假日+带薪假天数 | AttendanceProvider 集成 |
| SAL-14 | 病假工资按工龄系数计算（6个月内/超6个月） | sick_leave_policies 表 + 工龄查询 |
| SAL-15 | 加班费按法定系数计算（工作日150%/双休日200%/节假日300%） | AttendanceMonthly.overtime_hours + 三档费率 |
| SAL-16 | 管理员可查看税前工资明细（分项展示） | PayrollItem 已有字段，UI 展示扩展 |
| SAL-17 | 管理员可向全员发送当月工资条通知 | asynq 队列批量发送 |
| SAL-18 | 管理员可向选定员工发送指定月份工资条 | SendSlip 扩展月份参数 |
| SAL-19 | 薪资列表支持 Excel 格式下载（当前页/含税前明细） | excelize 导出扩展，含 PayrollItem 分项 |

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| 薪资数据看板聚合 | Backend (API) | Frontend (卡片展示) | 后端并发聚合 4 个指标，前端只做展示 |
| 调薪/普调 INSERT | Backend (API) | Frontend (表单) | 后端 INSERT 历史不可更新，前端提交表单 |
| 个税 Excel 上传 | Backend (API) | Frontend (上传区) | excelize 在服务端解析，后端更新数据库 |
| 绩效系数设置 | Backend (API) | Frontend (滑块) | 后端读写 performance_coefficients 表 |
| 薪资算法增强（考勤联动） | Backend (API) | Attendance Module | salary 调用 AttendanceProvider 读取考勤数据 |
| 工资条发送 | Backend (Queue) | WeChat/SMS Service | asynq 异步发送，避免 HTTP 超时 |
| 工资条 H5 查看 | CDN/Static | Backend (Token 验证) | 无状态 H5 页面，Token 验证后返回数据 |
| 薪资列表导出 | Backend (API) | Frontend (下载) | 服务端生成 Excel stream 返回 |

---

## Standard Stack

### Core (Backend - Go)

| Library | Version | Purpose | Why Standard | Status |
|---------|---------|---------|--------------|--------|
| shopspring/decimal | v1.4.0 | 薪资精确计算 | 已有依赖，避免 float64 精度丢失 | [VERIFIED: go.mod] |
| excelize/v2 | v2.10.1 | Excel 读写 | 已有依赖，个税上传和薪资导出复用 | [VERIFIED: go.mod] |
| qmuntal/stateless | v1.8.0 | 状态机 | Phase 6 已引入 | [VERIFIED: go.mod] |
| asynq | **NEW** (v0.26.0) | 异步任务队列 | 工资条批量发送，解压 HTTP 请求 | [VERIFIED: GitHub releases] |
| golang-jwt | v5.3.1 | JWT 认证 | 已有依赖 | [VERIFIED: go.mod] |

**asynq 安装命令:** `go get github.com/hibiken/asynq@latest`

### Supporting (Backend)

| Library | Purpose | When to Use |
|---------|---------|-------------|
| excelize | Excel 解析和生成 | 个税上传、薪资导出、工资条导出 |
| resty | HTTP 客户端 | 微信小程序通知、阿里云 SMS |
| crypto/rand | Token 生成 | 工资条 Token（已有模式复用） |

### Core (Frontend - Vue 3)

| Library | Version | Purpose | Status |
|---------|---------|---------|--------|
| vue | 3.5.32 | 框架 | [VERIFIED: package.json] |
| element-plus | 2.13.6 | UI 组件 | [VERIFIED: package.json] |
| @vueuse/core | 14.2.1 | 组合式工具 | [VERIFIED: package.json] |
| xlsx (SheetJS) | 0.18.5 | 浏览器端 Excel | [VERIFIED: package.json] |
| dayjs | 1.11.20 | 日期处理 | [VERIFIED: package.json] |

**无新增前端依赖。** xlsx 和 dayjs 已在项目中。

---

## Architecture Patterns

### System Architecture Diagram

```
[H5 Browser]
    |
    | HTTPS / REST
    v
[API Gateway Layer] -- Auth / CORS / Tenant
    |
    +---> salary/handler.go (薪资 API)
    |       |
    |       +---> salary/dashboard_service.go  (4指标聚合)
    |       +---> salary/adjustment_service.go (调薪/普调 INSERT ONLY)
    |       +---> salary/performance_service.go (绩效系数 CRUD)
    |       +---> salary/tax_upload_service.go  (Excel 解析 + 匹配)
    |       +---> salary/slip_service.go        (发送 + asynq)
    |       +---> salary/calculator.go          (增强考勤联动)
    |
    +---> attendance/adapter.go (AttendanceProvider 实现)
    |       +---> attendance_monthly 读取
    |       +---> attendance_approvals 读取（病假明细）
    |       +---> attendance_rules.Holidays JSON 读取
    |
    +---> asynq (Redis 队列)
            |
            +---> SlipSendTask (工资条批量发送)
                    +---> 微信小程序通知 / 阿里云 SMS / H5 链接
```

### Recommended Project Structure

```
internal/salary/
├── model.go                    # 现有：PayrollRecord, PayrollItem, SalaryItem, PayrollSlip
├── calculator.go               # 现有：calculatePayroll，增强考勤联动
├── service.go                  # 现有：CalculatePayroll, SendSlip 等
├── slip.go                     # 现有：SendSlip, GetSlipByToken
├── excel.go                    # 现有：parseAttendanceExcel, ExportPayrollExcel
├── adapter.go                  # 现有：TaxProvider, SIDeductionProvider, EmployeeProvider
├── dto.go                      # 现有：SalaryItemInput, PayrollResult, PayrollRecordResponse
├── errors.go                   # 现有：错误定义
├── repository.go               # 现有：Repository
├── salary_template*.go         # 现有：薪资模板 CRUD
│
# --- Phase 7 新增文件 ---
├── dashboard_service.go        # 薪资看板：4指标 + 环比聚合
├── dashboard_handler.go        # 薪资看板 API 端点
├── adjustment_model.go         # SalaryAdjustment 模型（INSERT ONLY）
├── adjustment_service.go       # 调薪/普调逻辑
├── adjustment_handler.go       # 调薪 API 端点
├── adjustment_dto.go          # 调薪请求/响应 DTO
├── adjustment_repository.go    # 调薪记录查询
├── performance_model.go        # PerformanceCoefficient 模型
├── performance_service.go      # 绩效系数 CRUD
├── performance_handler.go     # 绩效系数 API 端点
├── performance_repository.go  # 绩效系数查询
├── tax_upload_service.go       # 个税 Excel 上传解析
├── tax_upload_handler.go       # 个税上传 API 端点
├── slip_send_service.go       # 工资条发送（asynq 批量）
├── slip_send_handler.go       # 工资条发送 API 端点
├── slip_send_task.go          # asynq Task 定义
├── slip_send_log_model.go     # SalarySlipSendLog 模型（发送记录）
├── calculator_enhanced.go      # 薪资算法增强（考勤联动、病假、加班）
├── sick_leave_policy_model.go # SickLeavePolicy 模型
├── sick_leave_policy_service.go # 病假系数查询服务
└── salary_unlock_service.go   # confirmed/paid 解锁服务

internal/attendance/
├── model.go                    # 现有：AttendanceMonthly, Approval 等
├── adapter.go                 # 新增：AttendanceProvider 接口定义
├── adapter_impl.go            # 新增：AttendanceProvider 接口实现
└── ...                        # Phase 6 其他文件

frontend/src/views/tool/
├── SalaryTool.vue             # 重写：扩展为 8 个 Tab
├── SalaryDashboard.vue        # 新增：薪资数据看板
├── SalaryAdjustment.vue       # 新增：调薪/普调 Tab
├── TaxUpload.vue              # 新增：个税上传 Tab
├── PerformanceCoefficient.vue # 新增：绩效系数设置 Tab
├── SalarySlipSend.vue         # 新增：发工资条 Tab
├── SalaryList.vue             # 新增：薪资列表/导出 Tab
└── SalarySlipH5.vue          # 新增：独立 H5 工资条页面（无侧边栏）
```

### Pattern 1: AttendanceProvider Integration

**What:** salary 模块通过接口读取考勤数据，保持模块解耦。
**When to use:** CalculatePayroll 调用考勤月报数据时。
**Example:**
```go
// internal/attendance/adapter.go -- 新增接口定义
package attendance

// MonthlyAttendance 出勤月报数据（供 salary 模块使用）
type MonthlyAttendance struct {
    ActualDays       float64 // 实际出勤天数
    ShouldAttend     float64 // 应出勤天数
    OvertimeHours    float64 // 总加班时长（小时），已按 0.5h 取整
    PaidLeaveDays    float64 // 带薪假天数（年假/婚假/产假/陪产假/调休）
    LegalHolidayDays float64 // 法定节假日天数（从 AttendanceRule.Holidays 读取）
    SickLeaveDays    float64 // 病假天数
    // 加班明细（用于三档费率）
    OvertimeWeekdayHours  float64 // 工作日加班
    OvertimeWeekendHours  float64 // 双休日加班
    OvertimeHolidayHours  float64 // 节假日加班
}

// AttendanceProvider 考勤数据提供者接口（供 salary 模块调用）
type AttendanceProvider interface {
    // GetMonthlyAttendance 获取员工月度出勤数据
    GetMonthlyAttendance(orgID, employeeID int64, yearMonth string) (*MonthlyAttendance, error)
}
```

**Source:** [CODEBASE: internal/salary/adapter.go] -- adapter 接口模式已建立，直接扩展；[CODEBASE: internal/attendance/model.go] -- AttendanceMonthly 已有的字段：ActualDays, RequiredDays(should_attend), OvertimeHours, LeaveDays, AbsentDays；[CODEBASE: 06-CONTEXT.md D-SAL-ATT-04] -- 确认接口返回字段定义。

### Pattern 2: errgroup Dashboard Aggregation

**What:** 薪资看板 4 个指标并发查询，借鉴 `internal/dashboard/service.go` 的 `errgroup.WithContext` 模式。
**When to use:** SalaryDashboardService 并发查询应发/实发/社保/个税 4 个指标时。
**Example:**
```go
// internal/salary/dashboard_service.go
func (s *SalaryDashboardService) GetDashboard(ctx context.Context, orgID int64, year, month int) (*SalaryDashboardResult, error) {
    g, ctx := errgroup.WithContext(ctx)
    var gross, net, si, tax float64

    g.Go(func() error {
        total, err := s.repo.GetGrossIncomeTotal(orgID, year, month)
        if err == nil { gross = total }
        return err
    })
    g.Go(func() error {
        total, err := s.repo.GetNetIncomeTotal(orgID, year, month)
        if err == nil { net = total }
        return err
    })
    g.Go(func() error {
        total, err := s.repo.GetSITotal(orgID, year, month)
        if err == nil { si = total }
        return err
    })
    g.Go(func() error {
        total, err := s.repo.GetTaxTotal(orgID, year, month)
        if err == nil { tax = total }
        return err
    })
    if err := g.Wait(); err != nil {
        return nil, err
    }
    // 环比上月
    prevMonth := getPreviousMonth(year, month)
    // ... 环比计算逻辑
}
```

**Source:** [CODEBASE: internal/dashboard/service.go] -- errgroup 并发聚合模式已验证。

### Pattern 3: Salary Adjustment INSERT ONLY

**What:** 调薪记录 INSERT 新行，通过 effective_month 自然生效，禁止 UPDATE 历史。
**When to use:** 员工调薪和部门普调时。
**Example:**
```go
// internal/salary/adjustment_model.go
type SalaryAdjustment struct {
    model.BaseModel
    EmployeeID     int64          `gorm:"column:employee_id;index"`
    DepartmentID   *int64         `gorm:"column:department_id;index"` // 普调时使用
    Type           string         `gorm:"column:type;type:varchar(20)"` // individual/department
    EffectiveMonth string        `gorm:"column:effective_month;type:varchar(7);index"`
    AdjustmentType string         `gorm:"column:adjustment_type;type:varchar(20)"` // base_salary/allowance/bonus/year_end_bonus/other
    AdjustBy       string         `gorm:"column:adjust_by;type:varchar(10)"` // amount/ratio
    OldValue       decimal.Decimal `gorm:"column:old_value;type:decimal(12,2)"`
    NewValue       decimal.Decimal `gorm:"column:new_value;type:decimal(12,2)"`
    Status         string         `gorm:"column:status;type:varchar(20)"` // active/expired
}

func (SalaryAdjustment) TableName() string { return "salary_adjustments" }
```

**关键：** CreatePayroll 时，对于每个员工的每项薪资项，查询 `effective_month <= 当前月份` 的最新调薪记录并应用。

### Pattern 4: asynq Batch Slip Send

**What:** 工资条批量发送通过 asynq 队列异步处理，避免 HTTP 超时。
**When to use:** SAL-17 向全员发送当月工资条时。
**Example:**
```go
// internal/salary/slip_send_task.go
package salary

import (
    "github.com/hibiken/asynq"
)

const TypeSlipSend = "salary:slip:send"

// SlipSendPayload asynq 任务 payload
type SlipSendPayload struct {
    OrgID       int64
    UserID      int64
    Year        int
    Month       int
    EmployeeIDs []int64 // 空=全员
    Channel     string  // miniapp/sms/h5
}

func EnqueueSlipSend(payload *SlipSendPayload) error {
    task, _ := NewTask(TypeSlipSend, payload)
    return AsynqClient.Enqueue(task, asynq.MaxRetry(3))
}
```

**Source:** [CODEBASE: 07-CONTEXT.md D-SAL-SLIP-02] -- 确认使用 asynq 队列；asynq 在 go.mod 中不存在（NOT FOUND），需新增安装。

### Anti-Patterns to Avoid

- **UPDATE 历史 PayrollRecord（CRITICAL）：** confirmed/paid 月份禁止任何字段修改，违反则破坏工资数据完整性。预防：Service 层加状态校验。
- **浮点计算薪资（CRITICAL）：** float64 导致 0.1+0.2!=0.3 的精度问题。预防：所有薪资计算使用 `shopspring/decimal`，最终结果 `Round(2)` 后再转为 float64 存储到 DB。
- **调薪 UPDATE 历史（D-SAL-ADJ-01）：** 必须 INSERT 新行，effective_month 自然生效，禁止 UPDATE。
- **全员工资条同步发送：** 50 人同步发送导致 HTTP 超时。预防：使用 asynq 队列，HTTP 请求立即返回"发送中"。
- **AttendanceProvider 返回空数据未处理：** 员工当月无考勤记录时（可能未配置打卡规则），按全勤处理（actual_days = required_days）。
- **个税上传后未标记工资表回 draft：** 上传个税改变了税额，draft 前必须重新核算。

---

## Implementation Approach

### Plan 07-01: 薪资数据看板 + 后端基础设施（调薪模型 + 绩效模型 + 病假系数表）

**Backend 新增文件：**
1. `internal/salary/dashboard_service.go` -- errgroup 并发聚合 4 指标，环比上月计算
2. `internal/salary/dashboard_handler.go` -- `GET /api/v1/salary/dashboard?year=&month=`
3. `internal/salary/adjustment_model.go` -- `SalaryAdjustment` 表（INSERT ONLY）
4. `internal/salary/adjustment_repository.go` -- 调薪记录 CRUD
5. `internal/salary/adjustment_service.go` -- 调薪/普调逻辑
6. `internal/salary/adjustment_handler.go` -- `POST /api/v1/salary/adjustment`, `POST /api/v1/salary/mass-adjustment`
7. `internal/salary/adjustment_dto.go` -- 调薪请求/响应 DTO
8. `internal/salary/performance_model.go` -- `PerformanceCoefficient` 表
9. `internal/salary/performance_repository.go` -- 绩效系数 CRUD
10. `internal/salary/performance_service.go` -- 绩效系数读写
11. `internal/salary/performance_handler.go` -- `PUT /api/v1/salary/performance`, `GET /api/v1/salary/performance`
12. `internal/salary/sick_leave_policy_model.go` -- `SickLeavePolicy` 表（城市×工龄档位×系数）
13. `internal/salary/sick_leave_policy_service.go` -- 病假系数查询
14. `internal/salary/sick_leave_policy_data.go` -- 初始数据：北上广深各档位系数

**Frontend 新增/修改文件：**
15. `frontend/src/views/tool/SalaryTool.vue` -- 重写，新增 Dashboard Tab（第一个 Tab）
16. `frontend/src/views/tool/SalaryDashboard.vue` -- 薪资看板 4 张卡片（复用 AttendanceStatsCard 样式）
17. `frontend/src/api/salary.ts` -- 新增 dashboard/adjustment/performance API 方法
18. `frontend/src/router/index.ts` -- 注册薪资管理子菜单路由

**数据库迁移：**
```sql
CREATE TABLE salary_adjustments (
    id BIGSERIAL PRIMARY KEY,
    org_id BIGINT NOT NULL,
    employee_id BIGINT,  -- 单人调薪
    department_id BIGINT,  -- 普调时使用
    type VARCHAR(20) NOT NULL,  -- individual/department
    effective_month VARCHAR(7) NOT NULL,
    adjustment_type VARCHAR(20) NOT NULL,  -- base_salary/allowance/bonus/year_end_bonus/other
    adjust_by VARCHAR(10) NOT NULL,  -- amount/ratio
    old_value DECIMAL(12,2),
    new_value DECIMAL(12,2),
    status VARCHAR(20) DEFAULT 'active',
    created_by BIGINT,
    created_at TIMESTAMP DEFAULT NOW(),
    INDEX idx_salary_adj_emp (org_id, employee_id, effective_month),
    INDEX idx_salary_adj_dept (org_id, department_id, effective_month)
);

CREATE TABLE performance_coefficients (
    id BIGSERIAL PRIMARY KEY,
    org_id BIGINT NOT NULL,
    employee_id BIGINT NOT NULL,
    year INT NOT NULL,
    month INT NOT NULL,
    coefficient DECIMAL(5,4) NOT NULL DEFAULT 1.0,  -- 0.0000 ~ 2.0000
    created_by BIGINT,
    updated_by BIGINT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE INDEX idx_perf_emp_ym (org_id, employee_id, year, month)
);

CREATE TABLE sick_leave_policies (
    id BIGSERIAL PRIMARY KEY,
    city VARCHAR(50) NOT NULL,  -- 北上广深
    tenure_bucket VARCHAR(20) NOT NULL,  -- within_6months / over_6months
    coefficient DECIMAL(4,2) NOT NULL,  -- 0.40 ~ 1.00
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE salary_slip_send_logs (
    id BIGSERIAL PRIMARY KEY,
    org_id BIGINT NOT NULL,
    payroll_record_id BIGINT NOT NULL,
    employee_id BIGINT NOT NULL,
    channel VARCHAR(20) NOT NULL,  -- miniapp/sms/h5
    status VARCHAR(20) NOT NULL,  -- pending/sending/sent/failed
    error_message TEXT,
    sent_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### Plan 07-02: 薪资算法增强（考勤联动 + 病假 + 加班费）

**核心修改：**
1. `internal/attendance/adapter.go` -- 新增 `AttendanceProvider` 接口定义
2. `internal/attendance/adapter_impl.go` -- 实现 AttendanceProvider：
   - 读取 `attendance_monthly`（actual_days, required_days, overtime_hours, leave_days）
   - 读取 `attendance_approvals`（筛选 approved + 特定请假类型，统计 sick_leave_days 和 paid_leave_days）
   - 读取 `attendance_rules.holidays` JSON（统计当月法定节假日天数）
3. `internal/salary/calculator_enhanced.go` -- 新增函数：
   - `calculateBillingDays()` -- 计薪天数 = actual_days + legal_holiday_days + paid_leave_days
   - `calculateSalaryByBillingDays()` -- 基本工资 = base_salary / should_attend × billing_days
   - `calculateSickLeaveWage()` -- 病假工资 = base_salary × sick_leave_coefficient
   - `calculateOvertimePay()` -- 加班费 = base_salary / should_attend / 8h × overtime_hours × rate（150%/200%/300%）
4. `internal/salary/service.go` -- 增强 `CalculatePayroll`：
   - 调用 AttendanceProvider 获取考勤数据
   - 调用 SickLeavePolicyService 获取病假系数
   - 调用 PerformanceCoefficientRepo 获取绩效系数（默认 1.0）
   - 调用 SalaryAdjustmentRepo 获取生效期内调薪记录
   - 按增强算法重新计算各项薪资

**关键算法细节：**
```
// 计薪天数（per D-SAL-ATT-01）
billing_days = actual_days + legal_holiday_days + paid_leave_days

// 基本工资按计薪天数计算（per SAL-13）
salary_for_month = base_salary / should_attend * billing_days

// 病假工资（per SAL-14 + D-SAL-ATT-02）
// 1. 从 attendance_approvals 读取已批准 sick_leave 的总天数
// 2. 查询员工入职日期，计算工龄档位
// 3. 查询 sick_leave_policies 获取系数
// 4. sick_leave_wage = salary_for_month * coefficient
// 5. 不得低于当地最低工资 80%（由 SickLeavePolicyService 校验）

// 加班费（per SAL-15 + D-SAL-ATT-03）
// Phase 6 加班时长已按 0.5h 取整，存储精确到 0.01h
hourly_rate = base_salary / should_attend / 8
overtime_weekday_pay = hourly_rate * overtime_weekday_hours * 1.5
overtime_weekend_pay = hourly_rate * overtime_weekend_hours * 2.0
overtime_holiday_pay = hourly_rate * overtime_holiday_hours * 3.0
```

**AttendanceMonthly 补充字段：** Phase 6 已有 `OvertimeHours`（总加班时长），但需要分工作日/双休/节假日三档。Phase 6 的 AttendanceMonthly 结构中没有这三个分字段——需要在 Phase 6 中补充，或在 AttendanceProvider 实现中从 `attendance_approvals` 统计 overtime 类型。具体字段取决于 Phase 6 最终实现的 AttendanceMonthly 是否包含加班分档。

### Plan 07-03: 个税上传 + 工资条发送

**Backend 新增/修改文件：**
1. `internal/salary/tax_upload_service.go` -- excelize 解析 + 姓名匹配 + PayrollRecord.tax 更新
2. `internal/salary/tax_upload_handler.go` -- `POST /api/v1/salary/tax-upload`
3. `internal/salary/slip_send_log_model.go` -- SalarySlipSendLog 表
4. `internal/salary/slip_send_task.go` -- asynq Task 定义
5. `internal/salary/slip_send_service.go` -- 工资条发送（单条 + 批量 asynq）
6. `internal/salary/slip_send_handler.go` -- `POST /api/v1/salary/slip/send-all`, `POST /api/v1/salary/slip/send`
7. `internal/salary/slip.go` -- 扩展 SendSlip 支持月份参数
8. `cmd/server/main.go` -- 注册 asynq Worker + Redis 连接配置

**Frontend 新增文件：**
9. `frontend/src/views/tool/TaxUpload.vue` -- 个税上传 Tab
10. `frontend/src/views/tool/SalarySlipSend.vue` -- 发工资条 Tab
11. `frontend/src/views/tool/SalarySlipH5.vue` -- 独立 H5 工资条页面

**个税 Excel 匹配逻辑（per D-SAL-TAX-01/02）：**
```
1. excelize.OpenReader 解析上传文件
2. 遍历行，跳过表头，找到姓名列和应补退列
   - 姓名列别名：["姓名", "员工姓名", "纳税人姓名"]
   - 应补退别名：["应补/应退额", "应补退额", "应补（退）额", "应补", "应退"]
   - 个税金额列别名：["个税", "税额", "个人所得税", "本期应扣缴税额"]
3. 精确匹配（姓名完全一致）→ 模糊匹配（LIKE %name%）→ 跳过
4. 部分成功：返回成功行 + 失败行日志；全部失败：整体失败 + 错误原因
5. 确认后批量 UPDATE payroll_records SET tax = matched_value WHERE employee_id IN (...)
6. UPDATE payroll_records SET status = 'draft' WHERE year = ? AND month = ? AND status IN ('calculated', 'confirmed')
```

**工资条 H5 路由：**
```
GET /salary/slip/:token  --> SalarySlipH5.vue（独立页面，无 Layout）
```

### Plan 07-04: 调薪管理 UI + 薪资列表导出增强

**Backend 新增/修改文件：**
1. `internal/salary/adjustment_handler.go` -- 扩展调薪 API（增加预览接口）
2. `internal/salary/adjustment_service.go` -- 扩展普调预览逻辑（影响人数/预估金额）
3. `internal/salary/excel.go` -- 扩展导出逻辑（含税前明细选项）
4. `internal/salary/salary_list_handler.go` -- 薪资列表 API（部门筛选、姓名搜索）
5. `internal/salary/salary_unlock_service.go` -- confirmed/paid 解锁服务（含验证码）
6. `internal/salary/salary_unlock_handler.go` -- 解锁 API

**Frontend 新增/修改文件：**
7. `frontend/src/views/tool/SalaryTool.vue` -- 新增调薪管理 Tab（第二个 Tab）
8. `frontend/src/views/tool/SalaryAdjustment.vue` -- 调薪/普调表单 + 预览面板
9. `frontend/src/views/tool/PerformanceCoefficient.vue` -- 绩效系数设置 Tab
10. `frontend/src/views/tool/SalaryList.vue` -- 薪资列表 Tab（含导出弹窗）
11. `frontend/src/views/tool/SalarySlipH5.vue` -- H5 工资条（完善）
12. `frontend/src/router/index.ts` -- 注册 `/salary/slip/:token` 路由（无 Layout）

**调薪预览逻辑：**
```
POST /api/v1/salary/adjustment/preview
- input: { employee_ids / department_ids, adjustment_type, adjust_by, value, effective_month }
- logic:
  1. 获取影响员工列表（按 employee_ids 或 department_ids）
  2. 获取员工当前薪资项（SalaryItem）
  3. 应用调薪（按金额或比例）
  4. 计算月度影响 = sum(调整后 - 调整前)
  5. 年度影响 = 月度影响 × 12
- output: { employee_count, monthly_impact, annual_impact, effective_month }
```

---

## Common Pitfalls

### Pitfall 1: 浮点精度丢失（CRITICAL）
**What goes wrong:** `float64` 计算 `0.1 + 0.2` 得到 `0.30000000000000004`，导致薪资金额误差。
**Why it happens:** IEEE 754 浮点表示法无法精确表示十进制小数。
**How to avoid:** 核心计算使用 `shopspring/decimal`，最终 `Round(2)` 后再转为 `float64` 存入 DB。`calculator_enhanced.go` 所有中间变量用 `decimal.Decimal` 类型。
**Warning signs:** `PayrollRecord.NetIncome` 出现 `xxx.0099999999` 或 `xxx.0000000001`。

### Pitfall 2: 考勤数据不存在时的兜底逻辑（HIGH）
**What goes wrong:** 员工当月无 AttendanceMonthly 记录（未配置打卡规则），导致薪资算法 panic 或返回零。
**Why it happens:** Phase 6 的打卡规则是可选配置的，未配置时无出勤数据。
**How to avoid:** AttendanceProvider 实现中，未找到 AttendanceMonthly 时返回默认值（全勤）：`actual_days = required_days = 当月应出勤`, `overtime_hours = 0`, `paid_leave_days = 0`, `legal_holiday_days = 0`。
**Warning signs:** 新员工首月工资为 0 或异常低。

### Pitfall 3: 调薪生效期导致重复生效（MEDIUM）
**What goes wrong:** 调薪 effective_month 早于当前月，但当月工资表已经 draft/calculated，调薪记录被应用后导致重复调薪。
**Why it happens:** CreatePayroll 复制上月薪资项时，未考虑 effective_month <= 当前月份 的调薪记录。
**How to avoid:** 在 `CalculatePayroll` 中，每次核算时重新从 SalaryAdjustment 表读取 `effective_month <= 当前月份` 的最新调薪记录并应用，而不是仅在 CreatePayroll 时一次性应用。

### Pitfall 4: 加班时长加班分档（MEDIUM）
**What goes wrong:** AttendanceMonthly 只存储了 `OvertimeHours`（总时长），无法区分工作日/双休/节假日三档费率。
**Why it happens:** Phase 6 AttendanceMonthly 结构确认中没有加班分字段，取决于 Phase 6 实现。
**How to avoid:** AttendanceProvider 实现中，从 `attendance_approvals` 读取 `approval_type = 'overtime'` 且 `status = 'approved'` 的记录，按 overtime_type 分类统计。需要在 Phase 6 实现 AttendanceMonthly 补充加班分档字段，或在 AttendanceProvider 中实时聚合。
**Recommendation:** 在 Phase 6 的 AttendanceMonthly 模型中增加 `OvertimeWeekdayHours`、`OvertimeWeekendHours`、`OvertimeHolidayHours` 三个字段，由 Phase 6 出勤月报预计算时填入（这是 Phase 7 SAL-15 正确实现的前提）。

### Pitfall 5: asynq Worker 未注册（MEDIUM）
**What goes wrong:** asynq 任务入队成功，但 Worker 未注册导致任务永不执行。
**Why it happens:** `cmd/server/main.go` 未调用 `asynq.NewServer()` + 注册 Handler。
**How to avoid:** 在 main.go 中注册 asynq Worker，参考 asynq 官方模式：
```go
// cmd/server/main.go
import "github.com/hibiken/asynq"
redisConn, _ := redis.Dial("tcp", viper.GetString("redis.addr"))
srv := asynq.NewServer(redisConn, asynq.Config{
    Concurrency: 10,
})
mux := asynq.NewServeMux()
mux.HandleFunc(salary.TypeSlipSend, salary.HandleSlipSendTask)
go srv.Run(mux)
```

### Pitfall 6: 个税上传状态回滚（HIGH）
**What goes wrong:** 个税上传后标记工资表回 draft，但部分 UPDATE 失败导致数据不一致。
**Why it happens:** 未使用事务，部分员工 UPDATE 成功、部分失败。
**How to avoid:** 使用 GORM 事务包裹：
```go
return s.repo.db.Transaction(func(tx *gorm.DB) error {
    for _, update := range updates {
        if err := tx.Model(&PayrollRecord{}).Where(...).Update("tax", update.Tax).Error; err != nil {
            return err
        }
    }
    return tx.Model(&PayrollRecord{}).Where("...").Updates(map[string]interface{}{
        "status": PayrollStatusDraft,
    }).Error
})
```

### Pitfall 7: 普调部门为空时全量员工调薪（MEDIUM）
**What goes wrong:** 普调时选择"全选"部门时，department_ids 为空数组，被解读为"所有部门"，导致影响范围超出预期。
**How to avoid:** 前端传 `department_ids: []` 时，后端识别为"全企业调薪"，需二次确认。后端 API 明确区分：空数组 = 不指定部门（警告）vs null = 全企业。

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| 薪资精确计算 | float64 四则运算 | shopspring/decimal | 避免 0.1+0.2 精度丢失，v1.4.0 已在 go.mod |
| Excel 解析/导出 | 手写 xlsx 生成代码 | excelize v2.10.1 | 已有依赖，支持样式/公式/流式写入 |
| 批量工资条发送 | 同步循环 HTTP 请求 | asynq 任务队列 | 50 人同步发送会超时；asynq 提供重试/超时/持久化 |
| 考勤数据获取 | salary 直接 JOIN attendance 表 | AttendanceProvider 接口 | 模块解耦，attendance 独立演进 |
| 姓名匹配 | 简单字符串相等 | 精确匹配 + 模糊匹配 + 日志 | 用户输入可能有空格/全角字/别名 |
| JSONB 节假日解析 | 手写 JSON 解析 | GORM datatypes.JSON + json.RawMessage | 已有 AttendanceRule.Holidays 使用此模式 |

---

## Open Questions

1. **AttendanceMonthly 加班分档字段**
   - What we know: Phase 6 AttendanceMonthly 有 `OvertimeHours float64`（总时长），D-SAL-ATT-03 需要工作日/双休/节假日三档。
   - What's unclear: Phase 6 是否在 AttendanceMonthly 中预计算了加班分档字段？
   - Recommendation: 在 Phase 7 RESEARCH 中标注此依赖，要求 Phase 6 实现补充 `OvertimeWeekdayHours`/`OvertimeWeekendHours`/`OvertimeHolidayHours` 字段，或在 AttendanceProvider 中实时从 approvals 聚合。

2. **sick_leave_policies 初始数据**
   - What we know: 初期仅支持北上广深，按工龄档位（6个月内/超6个月）存储系数。
   - What's unclear: 各城市具体系数值（需法规依据）。
   - Recommendation: 使用 ASSUMED 值（上海：6个月内 60%，超6个月 40%），在实施前需产品确认或查阅各城市最新标准。

3. **工资条 H5 页面域名**
   - What we know: 工资条 H5 需要独立路由 `/salary/slip/:token`，不走 AppLayout。
   - What's unclear: 前端是否有独立 H5 入口页面（不需要登录的公开页面）？
   - Recommendation: Vue Router 配置 `SalarySlipH5.vue` 为独立页面（无 Layout wrapper），token 验证在后端。

4. **微信小程序通知是否需要真实对接**
   - What we know: 工资条发送通道优先级为微信小程序 → 短信 → H5 链接。
   - What's unclear: WeChat Mini Program 通知是否已对接（项目已有 silenceper/wechat SDK）？
   - Recommendation: 初期仅实现 H5 链接和手机号记录，微信小程序/SMS 通知接口预留待后续集成。

---

## Environment Availability

> Step 2.6: SKIPPED (no external dependencies identified beyond existing project stack).

所有依赖均已在项目中（Go 标准库 + go.mod 已有包）。新增依赖 asynq 需 `go get`。

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go 标准 `testing` + `testify/assert` |
| Config file | none — 测试直接调用 Service |
| Quick run command | `go test ./internal/salary/... -run TestCalculatePayroll -v` |
| Full suite command | `go test ./internal/salary/... -cover` |

### Phase Requirements to Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| SAL-01~04 | 薪资看板 4 指标 + 环比 | unit | `go test ./internal/salary/... -run TestDashboard -v` | N/A (new file) |
| SAL-05~07 | 调薪 INSERT ONLY + 生效期 | unit | `go test ./internal/salary/... -run TestAdjustment -v` | N/A (new file) |
| SAL-13 | 计薪天数算法 | unit | `go test ./internal/salary/... -run TestBillingDays -v` | N/A (new file) |
| SAL-14 | 病假工资系数 | unit | `go test ./internal/salary/... -run TestSickLeave -v` | N/A (new file) |
| SAL-15 | 加班费三档费率 | unit | `go test ./internal/salary/... -run TestOvertime -v` | N/A (new file) |
| SAL-08~10 | 个税 Excel 解析 + 匹配 | unit | `go test ./internal/salary/... -run TestTaxUpload -v` | N/A (new file) |
| SAL-11~12 | 绩效系数 × 绩效工资 | unit | `go test ./internal/salary/... -run TestPerformance -v` | N/A (new file) |

### Sampling Rate
- **Per task commit:** `go test ./internal/salary/... -run <test_name> -v`
- **Per wave merge:** `go test ./internal/salary/... -cover -race`
- **Phase gate:** Full suite green before `/gsd-verify-work`

### Wave 0 Gaps

- [ ] `internal/salary/calculator_enhanced_test.go` -- 计薪天数、病假、加班费测试用例（需先实现 calculator_enhanced.go）
- [ ] `internal/salary/adjustment_service_test.go` -- 调薪 INSERT ONLY 测试
- [ ] `internal/salary/tax_upload_service_test.go` -- 个税 Excel 解析 + 姓名匹配测试（mock EmployeeProvider）
- [ ] `internal/salary/dashboard_service_test.go` -- 看板聚合测试（mock Repository）
- [ ] `internal/attendance/adapter_impl_test.go` -- AttendanceProvider 测试

*(Existing test infrastructure: `internal/salary/calculator_test.go` — 扩展覆盖考勤联动场景)*

---

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | YES | 工资条 Token 验证（无密码，Token 一次性有效）|
| V3 Session Management | NO | 无会话管理需求 |
| V4 Access Control | YES | 工资条 Token 仅限对应员工查看（slip token + employee_id 绑定校验）|
| V5 Input Validation | YES | excelize 解析列名白名单校验；调薪金额/比例范围校验 |
| V6 Cryptography | YES | 手机号加密存储（已有 crypto.Encrypt 模式复用）|

### Known Threat Patterns for Salary Module

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| 调薪绕过生效期提前生效 | Tampering | Service 层强制 effective_month >= 当前月份，DB 触发器禁止 UPDATE history |
| 个税 Excel 注入字段名 | Injection | 列名白名单校验：仅允许预定义列名，拒绝任意列名 |
| 工资条 Token 枚举 | Information Disclosure | 64 字符随机 hex Token（256 位熵），链接设置 7 天过期 |
| 历史工资数据篡改 | Tampering | confirmed/paid 月份禁止 UPDATE，需企业主手机验证码解锁并记录审计日志 |
| 全员工资条隐私泄露 | Information Disclosure | Token 不包含 employee_id，仅通过 Token 查询关联员工 |

---

## Sources

### Primary (HIGH confidence)
- [CODEBASE: internal/salary/model.go] -- PayrollRecord/PayrollItem/SalaryItem 字段定义
- [CODEBASE: internal/salary/calculator.go] -- calculatePayroll 骨架，现有 calculateDailyWage 逻辑
- [CODEBASE: internal/salary/service.go] -- CalculatePayroll 全流程，Service 依赖注入模式
- [CODEBASE: internal/salary/slip.go] -- PayrollSlip.Token 模式，SendSlip 逻辑
- [CODEBASE: internal/salary/excel.go] -- excelize 解析模式，getSalaryItemAmount 工具函数
- [CODEBASE: internal/attendance/model.go] -- AttendanceMonthly 结构（ActualDays/RequiredDays/OvertimeHours/LeaveDays）
- [CODEBASE: internal/dashboard/service.go] -- errgroup 并发聚合模式
- [VERIFIED: go.mod] -- shopspring/decimal v1.4.0, excelize/v2 v2.10.1, qmuntal/stateless v1.8.0
- [VERIFIED: GitHub releases] -- asynq v0.26.0 最新版本

### Secondary (MEDIUM confidence)
- [CODEBASE: 06-CONTEXT.md D-SAL-ATT-04] -- AttendanceProvider 接口返回字段定义
- [CODEBASE: 07-CONTEXT.md D-SAL-ADJ-01/02, D-SAL-TAX-01/02/03, D-SAL-ATT-01/02/03/04, D-SAL-SLIP-01/02, D-SAL-DATA-01/02] -- 17 项锁定决策
- [CODEBASE: .planning/research/ARCHITECTURE.md] -- v1.3 架构设计，adapter 接口扩展建议
- [CODEBASE: .planning/research/FEATURES.md §2.3] -- 请假类型与薪资影响（病假/年假/调休计算规则）

### Tertiary (LOW confidence)
- [ASSUMED] -- sick_leave_policies 初始数据（北上广深各档位系数），需法规依据确认
- [ASSUMED] -- AttendanceMonthly 是否包含加班分档字段，取决于 Phase 6 最终实现

---

## Metadata

**Confidence breakdown:**
- Standard Stack: HIGH — 全部基于已验证的 go.mod/npm registry 依赖
- Architecture: HIGH — 基于现有 adapter/errgroup/INSERT ONLY 模式验证
- Pitfalls: HIGH — 基于代码审查和已有 v1.3 research（SUMMARY.md/FEATURES.md）

**Research date:** 2026-04-18
**Valid until:** 2026-05-18 (30 days — salary module APIs are stable)
