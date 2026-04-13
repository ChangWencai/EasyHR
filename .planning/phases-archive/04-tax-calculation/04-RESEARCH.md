# Phase 4: 个税计算 - Research

**Researched:** 2026-04-07
**Phase:** 04-tax-calculation
**Status:** Research complete

---

## 1. China Individual Income Tax — Tax Rate Table

### 七级超额累进税率表（年度/累计应纳税所得额）

自2019年个税改革后税率表未变，适用于累计预扣预缴法：

| 级数 | 累计应纳税所得额（年度） | 税率 | 速算扣除数 |
|------|--------------------------|------|------------|
| 1 | 不超过 36,000 元 | 3% | 0 |
| 2 | 36,000 - 144,000 元 | 10% | 2,520 |
| 3 | 144,000 - 300,000 元 | 20% | 16,920 |
| 4 | 300,000 - 420,000 元 | 25% | 31,920 |
| 5 | 420,000 - 660,000 元 | 30% | 52,920 |
| 6 | 660,000 - 960,000 元 | 35% | 85,920 |
| 7 | 超过 960,000 元 | 45% | 181,920 |

**Implementation:**
- Store in `TaxBracket` table, OrgID=0, global data
- Fields: `level` (int), `lower_bound` (decimal), `upper_bound` (decimal, nullable for top bracket), `rate` (decimal, e.g. 0.03), `quick_deduction` (decimal)
- Seed data via migration or seeder function
- Admin can update via H5 backend when policy changes

### 基本减除费用（起征点）

- 月度：5,000 元
- 年度：60,000 元
- 自2018年改革后未变，存储于税率表关联配置或单独常量

---

## 2. 累计预扣预缴法（Cumulative Withholding Method）

### Algorithm

```
当月应扣税额 = (累计应纳税所得额 × 适用税率 - 速算扣除数) - 累计已扣税额

其中：
累计应纳税所得额 = 累计收入 - 累计基本减除费用 - 累积专项扣除 - 累计专项附加扣除
```

### Calculation Steps (per month)

1. **累计收入** = SUM(所有月份的 gross_income) for current year
2. **累计基本减除费用** = month_number × 5,000
3. **累积专项扣除** = SUM(所有月份的社保个人缴费) for current year
4. **累计专项附加扣除** = SUM(所有月份的专项附加扣除总额) for current year
5. **累计应纳税所得额** = (1) - (2) - (3) - (4)
6. **适用税率** = 查税率表，找到累计应纳税所得额所在区间
7. **累计应扣税额** = 累计应纳税所得额 × 适用税率 - 速算扣除数
8. **当月应扣税额** = 累计应扣税额 - 累计已扣税额

### Tax Bracket Jump (税率跳档)

Tax bracket jumps occur naturally when cumulative taxable income crosses a threshold. Example:
- Jan-Jun: monthly salary 20,000, cumulative taxable income grows ~87,000 by June (within 3% bracket after deductions)
- Jul: cumulative crosses 36,000 → jumps to 10% bracket
- The algorithm automatically uses the higher rate, no special handling needed

### Edge Cases

- **当月应扣税额 < 0**: Can happen due to late-year deductions exceeding income. Do not refund — carry forward as negative withholding (offset against future months). If year-end is still negative, employee files annual tax return for refund.
- **年中入职**: Cumulative starts from hire month, not January. Month 1 of employment = first month of cumulative calculation.
- **年中离职**: Stop calculating. Employee may need annual reconciliation.
- **月薪变动**: Each month's gross_income may differ. The cumulative method handles this naturally.

---

## 3. 专项附加扣除（Special Additional Deductions）

### 2023年8月更新后的标准（2024-2026年适用）

| 扣除类型 | 标准金额 | 条件/说明 |
|----------|----------|-----------|
| 子女教育 | 2,000 元/月/每个子女 | 学前教育至博士（3岁至全日制学历教育结束） |
| 继续教育 | 400 元/月（学历）或 3,600 元/年（职业资格） | 学历教育最长48个月；职业资格证书取证当年 |
| 大病医疗 | 年度内个人负担超15,000元部分，最高80,000元/年 | 按年计算，非按月。只能由本人或配偶扣除 |
| 住房贷款利息 | 1,000 元/月 | 首套住房贷款利息，最长240个月 |
| 住房租金 | 800/1,100/1,500 元/月 | 直辖市/省会/计划单列市：1,500；市辖区户籍人口>100万：1,100；其他：800 |
| 赡养老人 | 3,000 元/月 | 被赡养人年满60岁。独生子女全扣，非独生子女分摊（每人≤1,500） |
| 3岁以下婴幼儿照护 | 2,000 元/月/每个婴幼儿 | 0-3岁婴幼儿，父母可选择由一方100%扣除或各50% |

### Implementation Model

```go
type SpecialDeduction struct {
    model.BaseModel
    EmployeeID      int64   `gorm:"column:employee_id;not null;index"`
    DeductionType   string  `gorm:"column:deduction_type;type:varchar(30);not null;index"`
    MonthlyAmount   float64 `gorm:"column:monthly_amount;not null"`
    Count           int     `gorm:"column:count;default:1"`           // 子女数/老人数等
    EffectiveStart  string  `gorm:"column:effective_start;type:varchar(7)"` // YYYY-MM
    EffectiveEnd    *string `gorm:"column:effective_end;type:varchar(7)"`   // YYYY-MM, null=ongoing
    Remark          string  `gorm:"column:remark;type:varchar(200)"`
}
```

**Deduction Types (enum):**
- `child_education` — 子女教育
- `continuing_education` — 继续教育
- `serious_illness` — 大病医疗
- `housing_loan` — 住房贷款利息
- `housing_rent` — 住房租金
- `elderly_care` — 赡养老人
- `infant_care` — 3岁以下婴幼儿照护

**Pre-configured standards (system constants or seed data):**
Each deduction type has a standard per-unit amount. The boss selects type + count/conditions, system calculates monthly amount.

---

## 4. Cross-Module Interface Design

### Dependency Map

```
Phase 2 (Employee) ──→ provides employee info, Contract.Salary
Phase 3 (Social Insurance) ──→ provides personal SI deduction amounts
Phase 4 (Tax) ──→ receives gross income as parameter, returns tax amount
Phase 5 (Salary) ──→ calls Tax.CalculateTax(), uses tax as deduction
```

### Interface Definitions

**Tax module defines (consumed by Phase 5):**
```go
// TaxCalculator 个税计算接口（Phase 5 调用）
type TaxCalculator interface {
    CalculateTax(orgID int64, employeeID int64, year int, month int, grossIncome float64) (*TaxResult, error)
}

type TaxResult struct {
    MonthlyTax         float64 `json:"monthly_tax"`
    CumulativeIncome   float64 `json:"cumulative_income"`
    CumulativeDeduction float64 `json:"cumulative_deduction"`
    CumulativeTaxable  float64 `json:"cumulative_taxable"`
    TaxRate            float64 `json:"tax_rate"`
    QuickDeduction     float64 `json:"quick_deduction"`
    CumulativeTax      float64 `json:"cumulative_tax"`
}
```

**Tax module consumes (from Phase 2/3):**
```go
// EmployeeInfoProvider 员工信息接口（由 employee adapter 实现）
type EmployeeInfoProvider interface {
    GetActiveSalary(orgID int64, employeeID int64) (float64, error)
    GetEmployeeHireMonth(orgID int64, employeeID int64) (string, error)
}

// SIDeductionProvider 社保个人扣款接口（由 socialinsurance adapter 实现）
type SIDeductionProvider interface {
    GetPersonalDeduction(orgID int64, employeeID int64, month string) (float64, error)
}
```

### Adapter Pattern (matches existing socialinsurance/employee_adapter.go)

```go
package tax

import (
    "github.com/wencai/easyhr/internal/employee"
    "github.com/wencai/easyhr/internal/socialinsurance"
)

type EmployeeAdapter struct {
    contractRepo *employee.ContractRepository
    empRepo      *employee.Repository
}

type SocialInsuranceAdapter struct {
    siSvc *socialinsurance.Service
}
```

---

## 5. Tax Declaration & Reminder

### 申报周期

- 扣缴义务人（企业）每月申报一次
- 申报截止日：每月15日（遇法定节假日顺延，V1.0使用固定15日）
- 申报渠道：自然人电子税务局（扣缴端），手动操作

### 申报表格式

自然人电子税务局的批量导入模板为 Excel 格式，核心字段：
- 纳税人姓名
- 证件类型（居民身份证）
- 证件号码
- 所得项目（工资薪金）
- 收入额
- 基本减除费用
- 专项扣除（社保个人部分）
- 专项附加扣除
- 其他扣除
- 应纳税所得额
- 税率
- 速算扣除数
- 应扣税额
- 已扣税额
- 本期应扣缴税额

### Reminder Pattern

Reuse `socialinsurance/scheduler.go` pattern:
- gocron v2 daily job at 08:00 CST
- Check: is today >= (15th of current month - 3 days)?
- If yes: generate reminders for all orgs that haven't declared this month
- Reminder model: `TaxReminder` (similar to social insurance Reminder)

---

## 6. Data Model Summary

### TaxBracket (OrgID=0, global)
```
id, org_id(=0), level, lower_bound, upper_bound, rate, quick_deduction, effective_year, created_at, updated_at
```

### SpecialDeduction (tenant-scoped)
```
id, org_id, employee_id, deduction_type, monthly_amount, count, effective_start, effective_end, remark, created_by, updated_by, created_at, updated_at, deleted_at
```

### TaxRecord (tenant-scoped, one per employee per month)
```
id, org_id, employee_id, employee_name, year, month,
gross_income, basic_deduction, si_deduction, special_deduction, total_deduction,
cumulative_income, cumulative_basic_deduction, cumulative_si_deduction, cumulative_special_deduction,
cumulative_taxable_income, tax_rate, quick_deduction, cumulative_tax, monthly_tax,
source (contract/salary), created_by, updated_by, created_at, updated_at, deleted_at
```

### TaxDeclaration (tenant-scoped, one per org per month)
```
id, org_id, year, month, status (pending/declared),
total_employees, total_income, total_tax,
declared_at, declared_by, created_by, updated_by, created_at, updated_at, deleted_at
```

### TaxReminder (tenant-scoped)
```
id, org_id, type (declaration_due), title, year, month, due_date,
is_read, is_dismissed, created_at, updated_at, deleted_at
```

---

## 7. Risks & Pitfalls

1. **Floating point precision**: Tax calculations must be precise to 2 decimal places. Use `math.Round(val*100)/100` for all intermediate results. Consider using integer cents (分) for internal calculations to avoid floating point errors.

2. **Cumulative data consistency**: If a past month's tax record is modified (rare but possible), all subsequent months' cumulative values become invalid. Need a recalculation mechanism.

3. **税率跳档边界值**: The boundary values must use `>=` for lower_bound and `<` for upper_bound consistently. E.g., 36,000 exactly falls in bracket 2 (10%).

4. **年中入职的累计起算**: First month of employment ≠ January. Cumulative starts from hire month. Need to track the hire month from Employee.HireDate.

5. **大病医疗按年非按月**: Serious illness medical deduction is annual, not monthly. It applies to annual reconciliation, not monthly withholding. V1.0 can simplify by not including it in monthly withholding (only include the 6 monthly deductions).

6. **住房贷款 vs 住房租金互斥**: Housing loan interest and housing rent are mutually exclusive per taxpayer. System should enforce this constraint.

---

## 8. Reference Implementations from Codebase

### Pattern: Adapter for Cross-Module Decoupling
- File: `internal/socialinsurance/employee_adapter.go`
- Pattern: Define interface in consumer package, implement adapter in main.go
- Tax module should follow same pattern for EmployeeInfoProvider and SIDeductionProvider

### Pattern: Scheduler for Periodic Reminders
- File: `internal/socialinsurance/scheduler.go`
- Uses gocron v2 with Redis distributed lock
- Tax module should add a second job to same scheduler or create parallel scheduler

### Pattern: Deduction Query Interface
- File: `internal/socialinsurance/dto.go` — `DeductionResponse`
- Social insurance already exposes `GetSocialInsuranceDeduction()` for downstream consumption
- Tax module will consume this via SIDeductionProvider adapter

### Pattern: Contract.Salary as Income Source
- File: `internal/employee/contract_model.go`
- `Contract.Salary` field (decimal(10,2)) stores base salary
- Tax module queries active contract to get base salary for standalone tax viewing

### Pattern: Excel Export
- File: `internal/employee/excel.go` or `internal/socialinsurance/excel.go`
- Uses excelize v2 for Excel generation
- Tax declaration Excel export should follow same pattern

---

## RESEARCH COMPLETE

**Key findings:**
1. Tax rate table unchanged since 2019 reform — stable, can be seeded once
2. Cumulative withholding algorithm is straightforward — SUM historical records, lookup bracket table
3. Special additional deductions updated in Aug 2023 — current standards well-documented
4. Circular dependency with Phase 5 resolved via parameter injection — Tax receives gross income as input
5. Housing loan vs housing rent are mutually exclusive — needs validation constraint
6. Serious illness medical is annual deduction — simplify in V1.0 by excluding from monthly withholding
7. All patterns (adapter, scheduler, Excel export) have working reference implementations in codebase
