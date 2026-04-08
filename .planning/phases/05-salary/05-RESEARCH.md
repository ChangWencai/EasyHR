# Phase 5: 工资核算 — Research

**Phase:** 05-salary
**Researched:** 2026-04-08
**Status:** RESEARCH COMPLETE

## Domain Knowledge

### 中国工资核算规则

1. **月计薪天数**: 21.75天（根据劳动法规定：(365-104)/12 = 21.75）。用于计算日薪和小时工资。
2. **日薪公式**: 日薪 = 月基本工资 / 21.75
3. **事假扣款**: 事假扣款 = 日薪 × 事假天数（按100%扣）
4. **病假扣款**: V1.0简化处理，病假扣款 = 日薪 × 病假天数（实际病假工资标准因地而异，V1.0按100%扣简化）
5. **工资核算流程**: 税前收入（各项收入之和）→ 减社保个人部分 → 减个税 → 减其他扣款 → 实发工资

### 累计预扣预缴法与工资的关系

- Phase 4 已实现 TaxCalculator 接口，接受 `grossIncome` 参数
- 工资模块将 税前收入(全部收入项之和) 作为 grossIncome 传入
- TaxCalculator 内部自动处理累计预扣、税率跳档
- 返回 TaxResult.MonthlyTax 即为当月应扣个税

### 薪资结构设计

根据 CONTEXT.md D-01~D-03 决策：
- 预置固定薪资项模板（基本工资、绩效、各类补贴、各类扣款）
- 企业启用/禁用薪资项
- 每员工独立填写各项金额
- SalaryItem 表独立存储，与 Contract.Salary 解耦

## Technical Analysis

### 1. 数据模型设计

根据 CONTEXT.md D-19 决策和现有代码模式（BaseModel + 多租户），Phase 5 需要以下核心模型：

**SalaryTemplateItem**（薪资项模板定义）
- 系统预置的薪资项列表（全局，OrgID=0）+ 企业启用状态（OrgID隔离）
- 字段：name, type(income/deduction), sort_order, is_required(基本工资必填), is_enabled(企业级别)
- 预置项：基本工资(income,required)、绩效工资(income)、岗位补贴(income)、餐补(income)、交通补(income)、通讯补(income)、其他补贴(income)、事假扣款(deduction)、病假扣款(deduction)、其他扣款(deduction)

**SalaryItem**（员工薪资项金额）
- 每员工每启用的薪资项一条记录
- 字段：employee_id, template_item_id, amount(float64), effective_month
- 支持按月变化（绩效、补贴等每月可不同）

**PayrollRecord**（月度工资核算主表）
- 每员工每月一条记录
- 字段：employee_id, year, month, status(draft/calculated/confirmed/paid), gross_income, si_deduction, tax, total_deductions, net_income, pay_method, pay_date, pay_note
- 确认后不可修改（快照性质）

**PayrollItem**（工资核算明细）
- 每员工多条，记录各项名称/类型/金额
- 字段：payroll_record_id, item_name, item_type(income/deduction), amount
- 核算时快照当时薪资项名称和金额

**PayrollSlip**（工资单，用于员工查看）
- 字段：payroll_record_id, token(unique, 64char hex), phone_encrypted, phone_hash, status(sent/viewed/signed), sent_at, viewed_at, signed_at, expires_at
- 复用邀请链接的 token 生成模式（crypto/rand 32字节 hex编码）
- 短信验证后查看

**AttendanceImport**（考勤导入记录，辅助表）
- 字段：org_id, year, month, file_url, status, imported_count
- 存储导入文件和结果

### 2. 跨模块集成设计

Phase 5 需要定义以下接口（adapter 模式，与 tax/adapter.go 一致）：

```go
// tax_provider.go - 个税计算接口
type TaxProvider interface {
    CalculateTax(orgID, employeeID int64, year, month int, grossIncome float64) (*tax.TaxResult, error)
}

// si_provider.go - 社保扣款接口
type SIDeductionProvider interface {
    GetPersonalDeduction(orgID, employeeID int64, month string) (float64, error)
}

// employee_provider.go - 员工信息接口
type EmployeeProvider interface {
    GetActiveEmployees(orgID int64) ([]EmployeeInfo, error)
    GetEmployeeByID(orgID, employeeID int64) (*EmployeeInfo, error)
}
```

Adapter 实现在各自的包中，在 main.go 中注入。

**关键集成点：**
- salary.Service 内部调用 taxSvc.CalculateTax() 获取个税
- salary.Service 内部调用 siSvc.GetSocialInsuranceDeduction() 获取社保扣款
- 工资确认后调用 siSvc.SuggestBaseAdjustment() 检查社保基数
- TaxRecord.Source 字段需设置为 "salary"（目前 tax 模块硬编码为 "contract"，需修改为可配置或增加参数）

**TaxRecord.Source 问题：**
当前 `tax.Service.CalculateTax()` 方法在 service.go:257 硬编码 `Source: "contract"`。Phase 5 需要调用同一个方法但设置 Source 为 "salary"。解决方案：
- 方案A：在 CalculateTax 接口增加 source 参数
- 方案B：增加 CalculateTaxWithSource 方法
- 方案C：在 TaxCalculator 接口不变，salary adapter 包装调用后修改 TaxRecord.Source
- **推荐方案A**：最小改动，向后兼容（默认值 "contract"）

### 3. 工资核算计算引擎

核算逻辑（纯函数式设计，便于测试）：

```
输入: []SalaryItem（员工当月各项金额）, siDeduction（社保个人扣款）, taxResult（个税结果）

计算步骤:
1. grossIncome = sum(SalaryItem where type=income)
2. totalDeductions = sum(SalaryItem where type=deduction)
3. 调用 TaxCalculator.CalculateTax(orgID, empID, year, month, grossIncome) 获取 taxResult
4. 调用 SIDeductionProvider.GetPersonalDeduction(orgID, empID, month) 获取 siDeduction
5. netIncome = grossIncome - siDeduction - taxResult.MonthlyTax - totalDeductions
6. 返回完整核算结果
```

注意：个税计算时，社保个人扣款已在 TaxCalculator 内部处理（通过 SIDeductionProvider），所以 netIncome 公式中社保扣款和个税是独立计算的。

**日薪计算辅助函数：**
```go
func calculateDailyWage(monthlyBase float64) float64 {
    return roundTo2(monthlyBase / 21.75)
}
```

### 4. H5 工资单查看流程

复用 Phase 2 邀请链接的 token 模式：

1. 老板确认工资表后，选择发送工资单给指定员工
2. 系统为每个员工的 PayrollRecord 生成 PayrollSlip（token + 过期时间 7天）
3. 老板获取分享链接（`/salary/slip/{token}`），通过微信/短信发送
4. 员工打开链接 → 输入手机号 → 系统发送验证码（复用 pkg/sms/）→ 验证通过后查看工资单
5. 员工点击"确认签收" → 更新 PayrollSlip.status 为 signed

**安全设计：**
- Token 为一次性使用（查看后不失效，但签收后不可更改）
- 短信验证码复用现有 SMS 基础设施
- 工资单展示脱敏数据（与 H5 管理后台一致）

### 5. 考勤 Excel 导入

使用 excelize 库（已在 Phase 2 引入）：

**导入模板：**
| 员工姓名 | 事假(天) | 病假(天) | 备注 |

**导入流程：**
1. 老板上传 Excel 文件
2. 系统解析 Excel，按姓名匹配员工
3. 计算扣款：事假扣款 = (基本工资/21.75) × 事假天数，病假扣款同理
4. 自动填入当月工资表的事假扣款、病假扣款项
5. 如果工资表中已有值，导入时覆盖

**匹配策略：**
- 按姓名精确匹配（小企业同名概率低）
- 未匹配的行标记为错误，提示老板手动处理
- 支持多sheet页（只读第一个）

### 6. 工资条 Excel 导出

使用 excelize 库，参考 internal/employee/excel.go 模式：

**导出格式：**
| 员工姓名 | 基本工资 | 绩效 | 补贴合计 | 事假扣款 | 病假扣款 | 其他扣款 | 税前收入 | 社保个人 | 个税 | 实发工资 |

**功能点：**
- 按月份导出全员工资表
- 蓝底白字表头样式
- SUM 合计行
- 数值格式保留两位小数

### 7. 异常发放提醒

根据 CONTEXT.md D-17 决策，不做定时任务，在确认操作时实时检查：

**检查逻辑：**
- 查询每个员工上月 PayrollRecord
- 对比本月实发与上月实发
- 偏差 > 30% 的员工标记为异常
- 在确认 API 响应中返回异常列表，前端高亮显示
- 不阻断确认操作，仅提醒

### 8. 复制上月工资表

**流程：**
1. 老板选择"新建当月工资表"
2. 系统查询上月 PayrollItem 数据
3. 复制所有薪资项金额到当月 SalaryItem（或直接创建 draft PayrollRecord + PayrollItem）
4. 老板修改变动项后一键核算

**注意：**
- 新入职员工无上月数据，各项默认为 0（基本工资从合同同步）
- 离职员工不纳入当月工资表

## Codebase Patterns to Follow

### 三层架构
```
internal/salary/
├── model.go          # SalaryTemplateItem, SalaryItem, PayrollRecord, PayrollItem, PayrollSlip
├── repository.go     # 数据访问层，TenantScope 自动注入
├── service.go        # 业务逻辑，注入 TaxProvider + SIDeductionProvider
├── handler.go        # HTTP Handler + RegisterRoutes
├── dto.go            # 请求/响应 DTO
├── adapter.go        # 接口定义（TaxProvider, SIDeductionProvider, EmployeeProvider）
├── employee_adapter.go  # EmployeeProvider 适配器实现
├── tax_adapter.go       # TaxProvider 适配器实现（包装 tax.Service）
├── si_adapter.go        # SIDeductionProvider 适配器实现（包装 socialinsurance.Service）
├── calculator.go     # 工资核算纯函数
├── calculator_test.go # 核算逻辑单元测试
├── excel.go          # 考勤导入 + 工资表导出
├── errors.go         # 错误定义 + 错误码 50xxx
└── slip.go           # 工资单 token 生成、验证、签收逻辑
```

### 错误码范围
- 50001: 薪资模板配置错误
- 50002: 工资核算失败
- 50003: 工资表状态不允许操作
- 50004: 考勤导入失败
- 50005: 工资单token无效或过期
- 50006: 短信验证失败
- 50007: 员工匹配失败

### RBAC 权限
- OWNER/ADMIN: 全部操作（配置薪资结构、核算、确认、发放、导出、推送工资单）
- MEMBER: 仅通过 H5 链接查看签收（不经过管理后台 RBAC）

### API 端点设计

```
# 薪资结构配置
GET    /salary/template          获取企业薪资模板（含启用状态）
PUT    /salary/template          更新薪资项启用/禁用

# 员工薪资项
GET    /salary/items             获取员工薪资项列表
PUT    /salary/items/:employee_id 更新员工薪资项金额

# 工资核算
POST   /salary/payroll           新建月度工资表（支持 copy_from_month）
GET    /salary/payroll           获取月度工资表列表
GET    /salary/payroll/:id       获取工资表明细
POST   /salary/payroll/calculate 一键核算（批量）
PUT    /salary/payroll/:id/confirm 确认工资表
PUT    /salary/payroll/:id/pay   记录发放

# 考勤导入
POST   /salary/attendance/import 导入考勤Excel

# 导出
GET    /salary/payroll/export    导出工资条Excel

# 工资单推送
POST   /salary/slip/send         发送工资单给员工

# H5 工资单查看（公开端点，不需要JWT）
GET    /salary/slip/:token       获取工资单页面信息
POST   /salary/slip/:token/verify 短信验证
POST   /salary/slip/:token/sign  签收确认
```

## Validation Architecture

### Dimension 1: Model Integrity
- SalaryTemplateItem 预置种子数据在系统启动时自动创建（OrgID=0）
- SalaryItem 的 amount 字段不能为负（收入项 >= 0，扣款项 >= 0）
- PayrollRecord 的 year/month 组合唯一（同一员工同一月份只能有一条记录）

### Dimension 2: Calculation Accuracy
- calculator_test.go 覆盖：
  - 基本场景（固定薪资，无变动项）
  - 含绩效/补贴场景
  - 含事假/病假扣款场景
  - 社保+个税联动计算
  - 日薪计算精度（21.75 除法保留2位小数）
  - 零收入/零扣款边界

### Dimension 3: State Machine
- PayrollRecord 状态流转：draft → calculated → confirmed → paid
- 只能在 draft/calculated 状态下编辑薪资项
- calculated 状态下可以修改后重新核算（回到 draft 再算）
- confirmed 后不可修改，只能查看

### Dimension 4: Cross-Module Integration
- tax.CalculateTax 正确接受 grossIncome 并返回 MonthlyTax
- socialinsurance.GetSocialInsuranceDeduction 正确返回 TotalPersonal
- 确认后 SuggestBaseAdjustment 正确触发

### Dimension 5: Token Security
- Token 使用 crypto/rand 生成（32字节，hex编码64字符）
- Token 有效期7天
- 短信验证码复用现有 SMS 基础设施（6位数字，5分钟有效期）

### Dimension 6: Excel Import Safety
- 文件大小限制（最大5MB）
- 行数限制（最大200行，覆盖50人企业）
- 只读取第一个sheet
- 未匹配姓名标记为错误，不静默跳过

---
*Research completed: 2026-04-08*
