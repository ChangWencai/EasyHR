package tax

import (
	"time"

	"github.com/wencai/easyhr/internal/common/model"
)

// BasicDeductionMonthly 基本减除费用（个税起征点）per D-03
const BasicDeductionMonthly = 5000.0

// TaxBracket 税率表模型（全局共享数据，OrgID=0）
// 存储七级超额累进税率
type TaxBracket struct {
	model.BaseModel
	Level          int     `gorm:"column:level;not null;index;comment:税级" json:"level"`
	LowerBound     float64 `gorm:"column:lower_bound;not null;comment:级距下限" json:"lower_bound"`
	UpperBound     float64 `gorm:"column:upper_bound;comment:级距上限（顶级无上限）" json:"upper_bound"`
	Rate           float64 `gorm:"column:rate;not null;comment:税率（如0.03=3%）" json:"rate"`
	QuickDeduction float64 `gorm:"column:quick_deduction;not null;default:0;comment:速算扣除数" json:"quick_deduction"`
	EffectiveYear  int     `gorm:"column:effective_year;not null;index;comment:生效年份" json:"effective_year"`
}

// TableName 指定表名
func (TaxBracket) TableName() string {
	return "tax_brackets"
}

// 扣除类型常量 (6种按月扣除, 不含大病医疗按 D-07)
const (
	DeductionTypeChildEducation     = "child_education"
	DeductionTypeContinuingEducation = "continuing_education"
	DeductionTypeHousingLoan        = "housing_loan"
	DeductionTypeHousingRent        = "housing_rent"
	DeductionTypeElderlyCare        = "elderly_care"
	DeductionTypeInfantCare         = "infant_care"
)

// DeductionStandard 预置扣除标准映射 (per D-06)
// 每种类型的单人/单份月度扣除金额
var DeductionStandard = map[string]float64{
	DeductionTypeChildEducation:     2000,
	DeductionTypeContinuingEducation: 400,
	DeductionTypeHousingLoan:        1000,
	DeductionTypeHousingRent:        1500,
	DeductionTypeElderlyCare:        3000,
	DeductionTypeInfantCare:         2000,
}

// ValidDeductionTypes 合法的扣除类型列表
var ValidDeductionTypes = []string{
	DeductionTypeChildEducation,
	DeductionTypeContinuingEducation,
	DeductionTypeHousingLoan,
	DeductionTypeHousingRent,
	DeductionTypeElderlyCare,
	DeductionTypeInfantCare,
}

// MutualExclusionGroup 互斥组: housing_loan 和 housing_rent 互斥 (per research pitfall #6)
var MutualExclusionGroup = map[string]string{
	DeductionTypeHousingLoan: DeductionTypeHousingRent,
	DeductionTypeHousingRent: DeductionTypeHousingLoan,
}

// isValidDeductionType 检查扣除类型是否合法
func isValidDeductionType(dtype string) bool {
	_, ok := DeductionStandard[dtype]
	return ok
}

// SpecialDeduction 专项附加扣除模型（租户隔离）
type SpecialDeduction struct {
	model.BaseModel
	EmployeeID     int64    `gorm:"column:employee_id;not null;index;comment:员工ID" json:"employee_id"`
	DeductionType  string   `gorm:"column:deduction_type;type:varchar(30);not null;index;comment:扣除类型" json:"deduction_type"`
	MonthlyAmount  float64  `gorm:"column:monthly_amount;not null;comment:每月扣除金额" json:"monthly_amount"`
	Count          int      `gorm:"column:count;default:1;comment:子女数量/份额数" json:"count"`
	EffectiveStart string   `gorm:"column:effective_start;type:varchar(7);not null;comment:生效起始月份（YYYY-MM）" json:"effective_start"`
	EffectiveEnd   *string  `gorm:"column:effective_end;type:varchar(7);comment:生效结束月份（持续为空）" json:"effective_end"`
	Remark         string   `gorm:"column:remark;type:varchar(200);comment:备注" json:"remark"`
}

// TableName 指定表名
func (SpecialDeduction) TableName() string {
	return "special_deductions"
}

// TaxRecord 个税计算记录（租户隔离，每员工每月一条）
type TaxRecord struct {
	model.BaseModel
	EmployeeID                int64   `gorm:"column:employee_id;not null;index;comment:员工ID" json:"employee_id"`
	EmployeeName              string  `gorm:"column:employee_name;type:varchar(50);not null;comment:员工姓名" json:"employee_name"`
	Year                      int     `gorm:"column:year;not null;index;comment:年份" json:"year"`
	Month                     int     `gorm:"column:month;not null;comment:月份" json:"month"`
	GrossIncome               float64 `gorm:"column:gross_income;not null;comment:税前工资" json:"gross_income"`
	BasicDeduction            float64 `gorm:"column:basic_deduction;not null;comment:基本减除费用（5000元）" json:"basic_deduction"`
	SIDeduction               float64 `gorm:"column:si_deduction;not null;default:0;comment:社保扣除" json:"si_deduction"`
	SpecialDeduction          float64 `gorm:"column:special_deduction;not null;default:0;comment:专项附加扣除" json:"special_deduction"`
	TotalDeduction            float64 `gorm:"column:total_deduction;not null;comment:扣除合计" json:"total_deduction"`
	CumulativeIncome          float64 `gorm:"column:cumulative_income;not null;comment:累计收入" json:"cumulative_income"`
	CumulativeBasicDeduction  float64 `gorm:"column:cumulative_basic_deduction;not null;comment:累计基本减除" json:"cumulative_basic_deduction"`
	CumulativeSIDeduction     float64 `gorm:"column:cumulative_si_deduction;not null;default:0;comment:累计社保扣除" json:"cumulative_si_deduction"`
	CumulativeSpecialDeduction float64 `gorm:"column:cumulative_special_deduction;not null;default:0;comment:累计专项附加扣除" json:"cumulative_special_deduction"`
	CumulativeTaxableIncome   float64 `gorm:"column:cumulative_taxable_income;not null;comment:累计应纳税所得额" json:"cumulative_taxable_income"`
	TaxRate                   float64 `gorm:"column:tax_rate;not null;comment:适用税率" json:"tax_rate"`
	QuickDeduction            float64 `gorm:"column:quick_deduction;not null;default:0;comment:速算扣除数" json:"quick_deduction"`
	CumulativeTax             float64 `gorm:"column:cumulative_tax;not null;comment:累计应缴税额" json:"cumulative_tax"`
	MonthlyTax                float64 `gorm:"column:monthly_tax;not null;comment:当月应缴税额" json:"monthly_tax"`
	Source                    string  `gorm:"column:source;type:varchar(20);not null;default:contract;comment:数据来源（contract/salary）" json:"source"`
}

// TableName 指定表名
func (TaxRecord) TableName() string {
	return "tax_records"
}

// TaxDeclaration 个税申报记录（租户隔离，每企业每月一条）
type TaxDeclaration struct {
	model.BaseModel
	Year           int        `gorm:"column:year;not null;index;comment:申报年份" json:"year"`
	Month          int        `gorm:"column:month;not null;comment:申报月份" json:"month"`
	Status         string     `gorm:"column:status;type:varchar(20);not null;default:pending;comment:申报状态（pending/declared）" json:"status"`
	TotalEmployees int        `gorm:"column:total_employees;not null;default:0;comment:申报人数" json:"total_employees"`
	TotalIncome   float64    `gorm:"column:total_income;not null;default:0;comment:收入合计" json:"total_income"`
	TotalTax      float64    `gorm:"column:total_tax;not null;default:0;comment:税额合计" json:"total_tax"`
	DeclaredAt    *time.Time `gorm:"column:declared_at;comment:申报时间" json:"declared_at"`
	DeclaredBy    int64      `gorm:"column:declared_by;comment:申报人ID" json:"declared_by"`
}

// Declaration status constants
const (
	DeclarationStatusPending  = "pending"
	DeclarationStatusDeclared = "declared"
)

// Reminder type constants
const (
	ReminderTypeDeclarationDue = "declaration_due"
)

// TaxReminder 个税申报提醒模型（租户隔离）
type TaxReminder struct {
	model.BaseModel
	Type        string     `gorm:"column:type;type:varchar(30);not null;index;comment:提醒类型" json:"type"`
	Title       string     `gorm:"column:title;type:varchar(200);not null;comment:提醒标题" json:"title"`
	Year        int        `gorm:"column:year;not null;comment:年份" json:"year"`
	Month       int        `gorm:"column:month;not null;comment:月份" json:"month"`
	DueDate     *time.Time `gorm:"column:due_date;type:date;comment:截止日期" json:"due_date"`
	IsRead      bool       `gorm:"column:is_read;default:false;comment:是否已读" json:"is_read"`
	IsDismissed bool       `gorm:"column:is_dismissed;default:false;comment:是否已忽略" json:"is_dismissed"`
}

// TableName 指定表名
func (TaxReminder) TableName() string {
	return "tax_reminders"
}
