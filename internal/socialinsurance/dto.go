package socialinsurance

// CreatePolicyRequest 创建社保政策请求
type CreatePolicyRequest struct {
	CityID        int            `json:"city_id" binding:"required"`
	EffectiveYear int            `json:"effective_year" binding:"required"`
	Config        FiveInsurances `json:"config" binding:"required"`
}

// UpdatePolicyRequest 更新社保政策请求
type UpdatePolicyRequest struct {
	Config FiveInsurances `json:"config" binding:"required"`
}

// PolicyResponse 社保政策响应
type PolicyResponse struct {
	ID            int64           `json:"id"`
	CityID        int             `json:"city_id"`
	CityName      string          `json:"city_name"`
	EffectiveYear int             `json:"effective_year"`
	Config        FiveInsurances  `json:"config"`
	CreatedAt     string          `json:"created_at"`
	UpdatedAt     string          `json:"updated_at"`
}

// CalculateRequest 社保计算请求
type CalculateRequest struct {
	CityID int     `json:"city_id" binding:"required"`
	Salary float64 `json:"salary" binding:"required,gt=0"`
	Year   int     `json:"year" binding:"required"`
}

// InsuranceAmountDetail 险种金额明细
type InsuranceAmountDetail struct {
	Name           string  `json:"name"`
	Base           float64 `json:"base"`
	CompanyRate    float64 `json:"company_rate"`
	CompanyAmount  float64 `json:"company_amount"`
	PersonalRate   float64 `json:"personal_rate"`
	PersonalAmount float64 `json:"personal_amount"`
}

// CalculateResponse 社保计算响应
type CalculateResponse struct {
	CityName      string                  `json:"city_name"`
	Salary        float64                 `json:"salary"`
	BaseAmount    float64                 `json:"base_amount"`
	TotalCompany  float64                 `json:"total_company"`
	TotalPersonal float64                 `json:"total_personal"`
	Items         []InsuranceAmountDetail `json:"items"`
}

// PolicyListQuery 政策列表查询参数
type PolicyListQuery struct {
	CityID   int `form:"city_id"`
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}
