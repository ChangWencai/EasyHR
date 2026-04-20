package salary

// ========== 薪资模板 DTO ==========

// TemplateItemResponse 薪资模板项响应
type TemplateItemResponse struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"` // income/deduction
	SortOrder  int    `json:"sort_order"`
	IsRequired bool   `json:"is_required"`
	IsEnabled  bool   `json:"is_enabled"`
}

// TemplateResponse 薪资模板响应
type TemplateResponse struct {
	Items []TemplateItemResponse `json:"items"`
}

// UpdateTemplateRequest 更新薪资模板请求
type UpdateTemplateRequest struct {
	Items []TemplateItemUpdate `json:"items" binding:"required"`
}

// TemplateItemUpdate 模板项更新
type TemplateItemUpdate struct {
	TemplateItemID int64 `json:"template_item_id" binding:"required"`
	IsEnabled      bool  `json:"is_enabled"`
}

// ========== 员工薪资项 DTO ==========

// EmployeeItemResponse 员工薪资项响应
type EmployeeItemResponse struct {
	TemplateItemID int64   `json:"template_item_id"`
	ItemName       string  `json:"item_name"`
	ItemType       string  `json:"item_type"`
	Amount         float64 `json:"amount"`
}

// EmployeeItemsResponse 员工薪资项列表响应
type EmployeeItemsResponse struct {
	EmployeeID int64                  `json:"employee_id"`
	Month      string                 `json:"month"`
	Items      []EmployeeItemResponse `json:"items"`
}

// SetEmployeeItemsRequest 设置员工薪资项请求
type SetEmployeeItemsRequest struct {
	Month string            `json:"month" binding:"required"` // "YYYY-MM"
	Items []SalaryItemInput `json:"items" binding:"required"`
}

// SalaryItemInput 薪资项输入
type SalaryItemInput struct {
	TemplateItemID int64   `json:"template_item_id" binding:"required"`
	Amount         float64 `json:"amount" binding:"min=0"`
}

// ========== 工资核算 DTO ==========

// CreatePayrollRequest 创建工资表请求
type CreatePayrollRequest struct {
	Year          int     `json:"year" binding:"required,min=2000,max=2100"`
	Month         int     `json:"month" binding:"required,min=1,max=12"`
	CopyFromMonth *string `json:"copy_from_month"` // "YYYY-MM"，可选
}

// PayrollRecordResponse 工资核算记录响应
type PayrollRecordResponse struct {
	ID              int64               `json:"id"`
	EmployeeID      int64               `json:"employee_id"`
	EmployeeName    string              `json:"employee_name"`
	Year            int                 `json:"year"`
	Month           int                 `json:"month"`
	Status          string              `json:"status"`
	GrossIncome     float64             `json:"gross_income"`
	SIDeduction     float64             `json:"si_deduction"`
	Tax             float64             `json:"tax"`
	TotalDeductions float64             `json:"total_deductions"`
	NetIncome       float64             `json:"net_income"`
	PayMethod       string              `json:"pay_method,omitempty"`
	PayDate         *string             `json:"pay_date,omitempty"`
	PayNote         string              `json:"pay_note,omitempty"`
	Items           []PayrollItemResponse `json:"items"`
}

// PayrollItemResponse 工资核算明细响应
type PayrollItemResponse struct {
	ItemName string  `json:"item_name"`
	ItemType string  `json:"item_type"`
	Amount   float64 `json:"amount"`
}

// BatchCalculateResponse 批量核算响应
type BatchCalculateResponse struct {
	TotalEmployees int     `json:"total_employees"`
	TotalNetIncome float64 `json:"total_net_income"`
}

// ConfirmPayrollRequest 确认工资表请求
type ConfirmPayrollRequest struct {
	Year  int `json:"year" binding:"required"`
	Month int `json:"month" binding:"required"`
}

// ConfirmResponse 确认响应
type ConfirmResponse struct {
	ConfirmedCount int             `json:"confirmed_count"`
	AbnormalItems  []AbnormalCheck `json:"abnormal_items,omitempty"`
}

// RecordPaymentRequest 发放记录请求
type RecordPaymentRequest struct {
	PayMethod string `json:"pay_method" binding:"required,oneof=bank_transfer cash other"`
	PayDate   string `json:"pay_date" binding:"required"`
	PayNote   string `json:"pay_note"`
}

// AttendanceImportResult 考勤导入结果
type AttendanceImportResult struct {
	MatchedCount int                  `json:"matched_count"`
	ErrorRows    []AttendanceErrorRow `json:"error_rows,omitempty"`
}

// AttendanceErrorRow 考勤导入错误行
type AttendanceErrorRow struct {
	RowNumber int    `json:"row_number"`
	Name      string `json:"name"`
	Error     string `json:"error"`
}

// ========== 工资单 DTO ==========

// SendSlipRequest 发送工资单请求
type SendSlipRequest struct {
	RecordIDs []int64 `json:"record_ids" binding:"required,min=1"`
}

// ConfirmSlipResponse 确认工资单响应（空响应）
type ConfirmSlipResponse struct{}

// SlipDetailResponse 工资单详情响应
type SlipDetailResponse struct {
	EmployeeName    string           `json:"employee_name"`
	Year            int              `json:"year"`
	Month           int              `json:"month"`
	Items           []SlipItemDetail `json:"items"`
	GrossIncome     float64          `json:"gross_income"`
	SIDeduction     float64          `json:"si_deduction"`
	Tax             float64          `json:"tax"`
	TotalDeductions float64          `json:"total_deductions"`
	NetIncome       float64          `json:"net_income"`
	Status          string           `json:"status"`
	SignedAt        *string         `json:"signed_at,omitempty"`
	ConfirmedAt     *string         `json:"confirmed_at,omitempty"`
}

// SlipItemDetail 工资单明细项
type SlipItemDetail struct {
	ItemName string  `json:"item_name"`
	ItemType string  `json:"item_type"`
	Amount   float64 `json:"amount"`
}

// VerifySlipPhoneRequest 验证工资单手机号请求
type VerifySlipPhoneRequest struct {
	Phone string `json:"phone" binding:"required"`
}

// VerifySlipCodeRequest 验证短信验证码请求
type VerifySlipCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required,len=6"`
}

// ExportPayrollRequest 导出工资表请求
type ExportPayrollRequest struct {
	Year  int `json:"year" binding:"required,min=2000,max=2100"`
	Month int `json:"month" binding:"required,min=1,max=12"`
}
