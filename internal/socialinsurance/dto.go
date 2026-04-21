package socialinsurance

// CreatePolicyRequest 创建社保政策请求
type CreatePolicyRequest struct {
	CityCode      int64          `json:"city_code" binding:"required"`
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
	CityID        int64          `json:"city_code"`
	CityName      string         `json:"city_name"`
	EffectiveYear int            `json:"effective_year"`
	Config        FiveInsurances `json:"config"`
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
}

// CalculateRequest 社保计算请求
type CalculateRequest struct {
	CityID int64     `json:"city_code" binding:"required"`
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
	CityID   int64 `form:"city_code"`
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// ========== 参保/停缴操作 DTO ==========

// BatchEnrollRequest 批量参保请求
type BatchEnrollRequest struct {
	EmployeeIDs []int64 `json:"employee_ids" binding:"required,min=1"`
	CityID      int64     `json:"city_code" binding:"required"`
	StartMonth  string  `json:"start_month" binding:"required"`
	Salary      float64 `json:"salary"`    // 社保基数（单个增员时传入）
	HFBase      float64 `json:"hf_base"`   // 公积金基数
	HFRatio     float64 `json:"hf_ratio"`  // 公积金比例
}

// EnrollPreviewRequest 参保预览请求
type EnrollPreviewRequest struct {
	EmployeeIDs []int64 `json:"employee_ids" binding:"required,min=1"`
	CityID      int64     `json:"city_code" binding:"required"`
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
	CityID        int64                  `json:"city_code"`
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

// ========== 社保数据看板 DTO（D-SI-05/D-SI-06）==========

// SIDashboardResponse 社保看板响应
type SIDashboardResponse struct {
	Stats        []SIStatItem  `json:"stats"`
	OverdueItems []OverdueItem `json:"overdue_items"` // D-SI-10 欠缴列表
}

// SIStatItem 社保看板单个指标
type SIStatItem struct {
	Label          string   `json:"label"`            // 指标名称
	Value          string   `json:"value"`            // 格式化后的金额
	TrendPercent   *string  `json:"trend_percent"`    // 环比百分比，无上月数据时为 nil
	TrendDirection string   `json:"trend_direction"`  // up/down/neutral
}

// OverdueItem 欠缴列表项（D-SI-10 横幅数据）
type OverdueItem struct {
	ID           int64  `json:"id"`
	EmployeeID   uint   `json:"employee_id"`
	EmployeeName string `json:"employee_name"`
	City         string `json:"city"`
	YearMonth    string `json:"year_month"`
	Amount       string `json:"amount"` // 格式 "¥X,XXX.XX"
}

// ========== 增减员操作 DTO（Phase 8 新增）==========

// EnrollRequest 增员请求（SI-05~SI-08）
type EnrollRequest struct {
	EmployeeID     int64   `json:"employee_id" binding:"required"`
	StartYearMonth string  `json:"start_year_month"`                          // 可选近3个月，默认当月
	CityID         int64   `json:"city_code" binding:"required"`
	SIBase         float64 `json:"si_base" binding:"required,gt=0"`           // 社保基数
	HFBase         float64 `json:"hf_base"`                                    // 公积金基数（可选，默认与社保同步）
	HFRatio        float64 `json:"hf_ratio"`                                   // 公积金比例（可选）
}

// StopRequest 减员请求（SI-09~SI-13）
type StopRequest struct {
	EmployeeID    int64  `json:"employee_id" binding:"required"`
	StopYearMonth string `json:"stop_year_month" binding:"required"` // 不可早于当月
	Reason        string `json:"reason" binding:"required,oneof=跳槽 退休 其他"` // 三选一
	TransferDate  string `json:"transfer_date"`                       // 转出社保日期
	HFFreezeDate  string `json:"hf_freeze_date"`                      // 封存公积金日期
}

// PaymentCallbackRequest 代理缴费 webhook 请求（SI-16）
type PaymentCallbackRequest struct {
	PaymentID int64  `json:"payment_id" binding:"required"`
	Status    string `json:"status" binding:"required,oneof=success failed"`
	Amount    string `json:"amount,omitempty"`
}

// ConfirmPaymentRequest 自主缴费确认请求（SI-15）
type ConfirmPaymentRequest struct {
	PaymentID int64 `json:"payment_id" binding:"required"`
}
