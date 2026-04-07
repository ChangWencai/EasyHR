package tax

import "time"

// ========== 专项附加扣除 DTO ==========

// CreateDeductionRequest 创建专项附加扣除请求
type CreateDeductionRequest struct {
	EmployeeID     int64   `json:"employee_id" binding:"required"`
	DeductionType  string  `json:"deduction_type" binding:"required"`
	Count          int     `json:"count" binding:"required,min=1"`
	EffectiveStart string  `json:"effective_start" binding:"required"`
	EffectiveEnd   *string `json:"effective_end"`
	Remark         string  `json:"remark"`
}

// UpdateDeductionRequest 更新专项附加扣除请求
type UpdateDeductionRequest struct {
	Count        int     `json:"count" binding:"required,min=1"`
	EffectiveEnd *string `json:"effective_end"`
	Remark       string  `json:"remark"`
}

// DeductionResponse 专项附加扣除响应
type DeductionResponse struct {
	ID             int64    `json:"id"`
	EmployeeID     int64    `json:"employee_id"`
	EmployeeName   string   `json:"employee_name"`
	DeductionType  string   `json:"deduction_type"`
	MonthlyAmount  float64  `json:"monthly_amount"`
	Count          int      `json:"count"`
	EffectiveStart string   `json:"effective_start"`
	EffectiveEnd   *string  `json:"effective_end"`
	Remark         string   `json:"remark"`
	CreatedAt      string   `json:"created_at"`
}

// DeductionListQuery 专项附加扣除列表查询参数
type DeductionListQuery struct {
	EmployeeID int64 `form:"employee_id"`
	Page       int   `form:"page"`
	PageSize   int   `form:"page_size"`
}

// ========== 税率表 DTO ==========

// TaxBracketResponse 税率表响应
type TaxBracketResponse struct {
	ID             int64   `json:"id"`
	Level          int     `json:"level"`
	LowerBound     float64 `json:"lower_bound"`
	UpperBound     float64 `json:"upper_bound"`
	Rate           float64 `json:"rate"`
	QuickDeduction float64 `json:"quick_deduction"`
	EffectiveYear  int     `json:"effective_year"`
}

// TaxBracketListQuery 税率表列表查询参数
type TaxBracketListQuery struct {
	EffectiveYear int `form:"effective_year"`
	Page          int `form:"page"`
	PageSize      int `form:"page_size"`
}

// ========== 个税计算结果 DTO ==========

// TaxResult 个税计算结果
type TaxResult struct {
	MonthlyTax                float64 `json:"monthly_tax"`
	CumulativeIncome          float64 `json:"cumulative_income"`
	CumulativeDeduction       float64 `json:"cumulative_deduction"`
	CumulativeTaxableIncome   float64 `json:"cumulative_taxable_income"`
	TaxRate                   float64 `json:"tax_rate"`
	QuickDeduction            float64 `json:"quick_deduction"`
	CumulativeTax             float64 `json:"cumulative_tax"`
	GrossIncome               float64 `json:"gross_income"`
	BasicDeduction            float64 `json:"basic_deduction"`
	SIDeduction               float64 `json:"si_deduction"`
	SpecialDeduction          float64 `json:"special_deduction"`
	TotalDeduction            float64 `json:"total_deduction"`
}

// TaxRecordResponse 个税计算记录响应
type TaxRecordResponse struct {
	ID                         int64   `json:"id"`
	EmployeeID                 int64   `json:"employee_id"`
	EmployeeName               string  `json:"employee_name"`
	Year                       int     `json:"year"`
	Month                      int     `json:"month"`
	GrossIncome                float64 `json:"gross_income"`
	BasicDeduction             float64 `json:"basic_deduction"`
	SIDeduction                float64 `json:"si_deduction"`
	SpecialDeduction           float64 `json:"special_deduction"`
	TotalDeduction             float64 `json:"total_deduction"`
	CumulativeIncome           float64 `json:"cumulative_income"`
	CumulativeBasicDeduction   float64 `json:"cumulative_basic_deduction"`
	CumulativeSIDeduction      float64 `json:"cumulative_si_deduction"`
	CumulativeSpecialDeduction float64 `json:"cumulative_special_deduction"`
	CumulativeTaxableIncome    float64 `json:"cumulative_taxable_income"`
	TaxRate                    float64 `json:"tax_rate"`
	QuickDeduction             float64 `json:"quick_deduction"`
	CumulativeTax              float64 `json:"cumulative_tax"`
	MonthlyTax                 float64 `json:"monthly_tax"`
	Source                     string  `json:"source"`
	CreatedAt                  string  `json:"created_at"`
}

// TaxRecordListQuery 个税记录列表查询参数
type TaxRecordListQuery struct {
	EmployeeID int64 `form:"employee_id"`
	Year       int   `form:"year"`
	Month      int   `form:"month"`
	Page       int   `form:"page"`
	PageSize   int   `form:"page_size"`
}

// ========== 申报管理 DTO ==========

// DeclarationResponse 申报记录响应
type DeclarationResponse struct {
	ID            int64      `json:"id"`
	Year          int        `json:"year"`
	Month         int        `json:"month"`
	Status        string     `json:"status"`
	TotalEmployees int       `json:"total_employees"`
	TotalIncome   float64    `json:"total_income"`
	TotalTax      float64    `json:"total_tax"`
	DeclaredAt    *time.Time `json:"declared_at"`
	DeclaredBy    int64      `json:"declared_by"`
	CreatedAt     string     `json:"created_at"`
}

// DeclarationListQuery 申报列表查询参数
type DeclarationListQuery struct {
	Year     int `form:"year"`
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}
