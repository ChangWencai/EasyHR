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
	ID            int64          `json:"id"`
	CityID        int            `json:"city_id"`
	CityName      string         `json:"city_name"`
	EffectiveYear int            `json:"effective_year"`
	Config        FiveInsurances `json:"config"`
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
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

// ========== 参保/停缴操作 DTO ==========

// BatchEnrollRequest 批量参保请求
type BatchEnrollRequest struct {
	EmployeeIDs []int64 `json:"employee_ids" binding:"required,min=1"`
	CityID      int     `json:"city_id" binding:"required"`
	StartMonth  string  `json:"start_month" binding:"required"`
}

// EnrollPreviewRequest 参保预览请求
type EnrollPreviewRequest struct {
	EmployeeIDs []int64 `json:"employee_ids" binding:"required,min=1"`
	CityID      int     `json:"city_id" binding:"required"`
}

// EnrollPreviewItem 参保预览单项
type EnrollPreviewItem struct {
	EmployeeID    int64                  `json:"employee_id"`
	EmployeeName  string                 `json:"employee_name"`
	BaseAmount    float64                `json:"base_amount"`
	TotalCompany  float64                `json:"total_company"`
	TotalPersonal float64                `json:"total_personal"`
	Items         []InsuranceAmountDetail `json:"items"`
}

// EnrollPreviewResponse 参保预览响应
type EnrollPreviewResponse struct {
	Items []EnrollPreviewItem `json:"items"`
}

// BatchEnrollResult 批量参保结果
type BatchEnrollResult struct {
	SuccessCount int            `json:"success_count"`
	FailCount    int            `json:"fail_count"`
	Failures     []EnrollFailure `json:"failures,omitempty"`
}

// EnrollFailure 参保失败项
type EnrollFailure struct {
	EmployeeID   int64  `json:"employee_id"`
	EmployeeName string `json:"employee_name"`
	Reason       string `json:"reason"`
}

// BatchStopRequest 批量停缴请求
type BatchStopRequest struct {
	RecordIDs []int64 `json:"record_ids" binding:"required,min=1"`
	EndMonth  string  `json:"end_month" binding:"required"`
}

// BatchStopResult 批量停缴结果
type BatchStopResult struct {
	SuccessCount int            `json:"success_count"`
	FailCount    int            `json:"fail_count"`
	Failures     []StopFailure  `json:"failures,omitempty"`
}

// StopFailure 停缴失败项
type StopFailure struct {
	RecordID int64  `json:"record_id"`
	Reason   string `json:"reason"`
}

// RecordListQueryParams 参保记录列表查询参数
type RecordListQueryParams struct {
	Status       string `form:"status"`
	EmployeeName string `form:"employee_name"`
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
}

// RecordResponse 参保记录响应
type RecordResponse struct {
	ID            int64                  `json:"id"`
	EmployeeID    int64                  `json:"employee_id"`
	EmployeeName  string                 `json:"employee_name"`
	CityID        int                    `json:"city_id"`
	CityName      string                 `json:"city_name"`
	BaseAmount    float64                `json:"base_amount"`
	Status        string                 `json:"status"`
	StartMonth    string                 `json:"start_month"`
	EndMonth      *string                `json:"end_month"`
	Details       []InsuranceAmountDetail `json:"details"`
	TotalCompany  float64                `json:"total_company"`
	TotalPersonal float64                `json:"total_personal"`
	CreatedAt     string                 `json:"created_at"`
}

// ChangeHistoryResponse 变更历史响应
type ChangeHistoryResponse struct {
	ID           int64       `json:"id"`
	RecordID     int64       `json:"record_id"`
	EmployeeID   int64       `json:"employee_id"`
	EmployeeName string      `json:"employee_name"`
	ChangeType   string      `json:"change_type"`
	BeforeValue  interface{} `json:"before_value"`
	AfterValue   interface{} `json:"after_value"`
	Remark       string      `json:"remark"`
	CreatedAt    string      `json:"created_at"`
}

// DeductionItem 扣款明细项
type DeductionItem struct {
	Name           string  `json:"name"`
	PersonalRate   float64 `json:"personal_rate"`
	PersonalAmount float64 `json:"personal_amount"`
}

// DeductionResponse 社保扣款响应（供 Phase 5 调用）
type DeductionResponse struct {
	Items         []DeductionItem `json:"items"`
	TotalPersonal float64         `json:"total_personal"`
}
