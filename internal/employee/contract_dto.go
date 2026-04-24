package employee

// CreateContractRequest 创建合同请求
type CreateContractRequest struct {
	ContractType    string   `json:"contract_type" binding:"required,oneof=fixed_term indefinite intern"`
	StartDate       string   `json:"start_date" binding:"required"`       // YYYY-MM-DD
	EndDate         string   `json:"end_date" binding:"omitempty"`        // 无固定期限不传
	Salary          *float64 `json:"salary" binding:"omitempty,gt=0"`
	ProbationMonths *int     `json:"probation_months" binding:"omitempty,min=0,max=12"`
	ProbationSalary *float64 `json:"probation_salary" binding:"omitempty,gt=0"`
}

// UpdateContractRequest 更新合同请求（部分更新）
type UpdateContractRequest struct {
	ContractType    *string   `json:"contract_type" binding:"omitempty,oneof=fixed_term indefinite intern"`
	StartDate       *string   `json:"start_date" binding:"omitempty"`
	EndDate         *string   `json:"end_date" binding:"omitempty"`
	Salary          *float64  `json:"salary" binding:"omitempty"`
	ProbationMonths *int      `json:"probation_months" binding:"omitempty,min=0,max=12"`
	ProbationSalary *float64  `json:"probation_salary" binding:"omitempty,gt=0"`
}

// UploadSignedRequest 上传签署扫描件请求
type UploadSignedRequest struct {
	SignedPDFURL string `json:"signed_pdf_url" binding:"required,url"`
	SignDate     string `json:"sign_date" binding:"required"` // YYYY-MM-DD
}

// TerminateContractRequest 终止合同请求
type TerminateContractRequest struct {
	TerminateDate   string `json:"terminate_date" binding:"required"`
	TerminateReason string `json:"terminate_reason" binding:"required,min=1,max=500"`
}

// ContractResponse 合同响应
type ContractResponse struct {
	ID               int64   `json:"id"`
	EmployeeID       int64   `json:"employee_id"`
	EmployeeName     string  `json:"employee_name"`
	ContractType     string  `json:"contract_type"`
	StartDate        string  `json:"start_date"`    // YYYY-MM-DD
	EndDate          *string `json:"end_date"`       // YYYY-MM-DD，无固定期限为 nil
	Salary           float64 `json:"salary"`
	ProbationMonths  int     `json:"probation_months"`
	ProbationSalary  float64 `json:"probation_salary"`
	Status           string  `json:"status"`
	PDFURL           string  `json:"pdf_url,omitempty"`
	SignedPDFURL     string  `json:"signed_pdf_url,omitempty"`
	SignDate         *string `json:"sign_date,omitempty"`       // YYYY-MM-DD
	TerminateDate    *string `json:"terminate_date,omitempty"` // YYYY-MM-DD
	TerminateReason  string  `json:"terminate_reason,omitempty"`
	CreatedAt        string  `json:"created_at"` // RFC3339
}

// ContractListQueryParams 合同列表查询参数
type ContractListQueryParams struct {
	Status   string `form:"status"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
}

// SendSignCodeRequest 发送签署验证码请求
type SendSignCodeRequest struct {
	ContractID int64  `json:"contract_id" binding:"required"`
	Phone     string `json:"phone" binding:"required,len=11"`
}

// SendSignCodeResponse 发送签署验证码响应
type SendSignCodeResponse struct {
	Message   string `json:"message"`
	ExpiresIn int    `json:"expires_in"` // 秒
}

// VerifySignCodeRequest 校验签署验证码请求
type VerifySignCodeRequest struct {
	ContractID int64  `json:"contract_id" binding:"required"`
	Phone     string `json:"phone" binding:"required,len=11"`
	Code      string `json:"code" binding:"required,len=6"`
}

// VerifySignCodeResponse 校验成功响应（含 sign_token）
type VerifySignCodeResponse struct {
	SignToken    string `json:"sign_token"` // 用于 ConfirmSign
	ExpiresIn   int    `json:"expires_in"` // 秒
	EmployeeName string `json:"employee_name"`
	ContractType string `json:"contract_type"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date,omitempty"`
	OrgName     string `json:"org_name"`
}

// ConfirmSignRequest 确认签署请求
type ConfirmSignRequest struct {
	ContractID int64  `json:"contract_id" binding:"required"`
	SignToken string `json:"sign_token" binding:"required"`
}

// ConfirmSignResponse 确认签署响应
type ConfirmSignResponse struct {
	SignedPDFURL string `json:"signed_pdf_url"`
	Message     string `json:"message"`
}

// GetSignedPdfResponse 获取已签PDF响应
type GetSignedPdfResponse struct {
	URL string `json:"url"`
}
