package wxmp

import "context"

// ========== Auth DTOs ==========

// LoginRequest 手机号+验证码登录请求
type LoginRequest struct {
	Phone string `json:"phone" binding:"required,len=11"`
	Code  string `json:"code" binding:"required,len=6"`
}

// LoginResponse 登录成功响应
type LoginResponse struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	ExpiresIn    int        `json:"expires_in"`
	Member       MemberInfo `json:"member"`
}

// MemberInfo 登录会员基本信息
type MemberInfo struct {
	UserID     uint   `json:"user_id"`
	EmployeeID uint   `json:"employee_id"`
	OrgID      uint   `json:"org_id"`
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Role       string `json:"role"`
	HasWechat  bool   `json:"has_wechat"`
}

// ========== Payslip DTOs ==========

// PayslipSummary 工资条摘要（列表用）
type PayslipSummary struct {
	ID       uint   `json:"id"`
	Year     int    `json:"year"`
	Month    int    `json:"month"`
	GrossPay string `json:"gross_pay"`
	NetPay   string `json:"net_pay"`
	Status   string `json:"status"` // pending/confirmed/paid
	SignedAt string `json:"signed_at,omitempty"`
	PaidAt   string `json:"paid_at,omitempty"`
}

// PayslipDetail 工资条明细（详情页用）
type PayslipDetail struct {
	ID           uint             `json:"id"`
	Year         int              `json:"year"`
	Month        int              `json:"month"`
	GrossPay     string           `json:"gross_pay"`
	NetPay       string           `json:"net_pay"`
	Items        []PayrollItemDTO `json:"items"`
	SocialDeduct string           `json:"social_deduct"`
	TaxDeduct    string           `json:"tax_deduct"`
	OtherDeduct  string           `json:"other_deduct"`
	SignedAt     string           `json:"signed_at,omitempty"`
	PaidAt       string           `json:"paid_at,omitempty"`
}

// PayrollItemDTO 工资条明细项
type PayrollItemDTO struct {
	Name   string `json:"name"`
	Type   string `json:"type"` // earning/deduction
	Amount string `json:"amount"`
}

// ========== Contract DTOs ==========

// ContractDTO 合同摘要
type ContractDTO struct {
	ID           uint   `json:"id"`
	ContractType string `json:"contract_type"` // labor/internship
	Status       string `json:"status"`        // pending/signed/expired
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	SignedAt     string `json:"signed_at,omitempty"`
	PDFURL       string `json:"pdf_url,omitempty"`
}

// ContractDetail 合同详情
type ContractDetail struct {
	ID           uint   `json:"id"`
	ContractType string `json:"contract_type"`
	Status       string `json:"status"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	SignedAt     string `json:"signed_at,omitempty"`
	PDFURL       string `json:"pdf_url,omitempty"`
}

// ========== Social Insurance DTOs ==========

// SocialInsuranceDTO 个人社保记录（仅展示个人缴费部分）
type SocialInsuranceDTO struct {
	PaymentMonth  string `json:"payment_month"`
	City          string `json:"city"`
	BaseAmount    string `json:"base_amount"`
	Pension       string `json:"pension"`
	Medical       string `json:"medical"`
	Unemployment  string `json:"unemployment"`
	TotalPersonal string `json:"total_personal"` // 仅个人缴费总额，不含单位
}

// ========== Expense DTOs ==========

// ExpenseRequest 创建报销单请求
type ExpenseRequest struct {
	Type        string   `json:"type" binding:"required,oneof=travel traffic entertainment office other"`
	Amount      string   `json:"amount" binding:"required"`
	Description string   `json:"description"`
	Attachments []string `json:"attachments"` // OSS file keys，最多9张
}

// ExpenseDTO 报销单响应
type ExpenseDTO struct {
	ID           uint     `json:"id"`
	Type         string   `json:"type"`
	Amount       string   `json:"amount"`
	Description  string   `json:"description"`
	Status       string   `json:"status"` // pending/approved/rejected/paid
	Attachments  []string `json:"attachments"`
	CreatedAt    string   `json:"created_at"`
	ApprovedAt   string   `json:"approved_at,omitempty"`
	RejectedAt   string   `json:"rejected_at,omitempty"`
	RejectReason string   `json:"reject_reason,omitempty"`
	PaidAt       string   `json:"paid_at,omitempty"`
	PaidMethod   string   `json:"paid_method,omitempty"`
}

// ========== SMS Verification DTOs ==========

// VerifyPayslipRequest 工资条身份验证请求
type VerifyPayslipRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

// VerifyPayslipResponse 工资条验证响应
type VerifyPayslipResponse struct {
	VerifyToken string `json:"verify_token"`
	ExpiresAt   string `json:"expires_at"`
}

// ========== OSS Upload DTOs ==========

// OssUploadURLRequest OSS预签名上传URL请求
type OssUploadURLRequest struct {
	FileName    string `json:"file_name" binding:"required"`
	FileSize    int64  `json:"file_size" binding:"required,gt=0"`
	ContentType string `json:"content_type" binding:"required"`
}

// OssUploadURLResponse OSS预签名上传URL响应
type OssUploadURLResponse struct {
	UploadURL string `json:"upload_url"`
	FileKey   string `json:"file_key"`
}

// ========== Bind Wechat DTOs ==========

// BindWechatRequest 绑定微信请求
type BindWechatRequest struct {
	Code string `json:"code" binding:"required"` // 微信授权code
}

// ========== Repository Interface ==========

// WXMPRepository 数据访问接口（供 WXMPService 使用）
type WXMPRepository interface {
	// GetMemberByPhone 通过手机号查找会员（user + employee 关联）
	GetMemberByPhone(ctx context.Context, phoneHash string) (*MemberInfo, error)
	// BindWechatOpenID 绑定微信 openid 到用户
	BindWechatOpenID(ctx context.Context, userID uint, openID string) error
	// ListPayslips 查询员工工资条列表
	ListPayslips(ctx context.Context, orgID, employeeID uint) ([]PayslipSummary, error)
	// GetPayslipByID 查询工资条详情（包含明细项）
	GetPayslipByID(ctx context.Context, orgID, employeeID, payslipID uint) (*PayslipDetail, error)
	// ListContracts 查询员工合同列表
	ListContracts(ctx context.Context, orgID, employeeID uint) ([]ContractDTO, error)
	// GetContractByID 查询合同详情
	GetContractByID(ctx context.Context, orgID, employeeID, contractID uint) (*ContractDetail, error)
	// ListSocialInsurance 查询员工社保记录（个人缴费）
	ListSocialInsurance(ctx context.Context, orgID, employeeID uint) ([]SocialInsuranceDTO, error)
	// ListExpenses 查询员工报销单列表
	ListExpenses(ctx context.Context, orgID, employeeID uint) ([]ExpenseDTO, error)
	// GetExpenseByID 查询报销单详情
	GetExpenseByID(ctx context.Context, orgID, employeeID, expenseID uint) (*ExpenseDTO, error)
	// CreateExpense 创建报销单
	CreateExpense(ctx context.Context, orgID, employeeID uint, req *ExpenseRequest) (*ExpenseDTO, error)
	// SignPayslip 更新工资条签收状态
	SignPayslip(ctx context.Context, orgID, employeeID, payslipID uint) error
}
