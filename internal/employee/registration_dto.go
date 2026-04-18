package employee

import "time"

// CreateRegistrationRequest 管理员创建登记表请求
type CreateRegistrationRequest struct {
	EmployeeID   *int64 `json:"employee_id"`                                      // 关联员工（可选）
	Name         string `json:"name" binding:"required,min=2,max=50"`             // 员工姓名
	DepartmentID *int64 `json:"department_id"`                                    // 部门ID（可选）
	Position     string `json:"position" binding:"required,min=1,max=100"`        // 岗位
	HireDate     string `json:"hire_date" binding:"required"`                     // 入职日期 YYYY-MM-DD
}

// SubmitRegistrationRequest 员工提交登记信息（公开接口）
type SubmitRegistrationRequest struct {
	Phone              string `json:"phone"`                         // 手机号码
	Address            string `json:"address"`                       // 居住地址
	IDCard             string `json:"id_card"`                       // 身份证号
	IDCardFrontURL     string `json:"id_card_front_url"`             // 身份证正面照URL
	IDCardBackURL      string `json:"id_card_back_url"`              // 身份证反面照URL
	BankAccount        string `json:"bank_account"`                  // 银行卡号
	BankName           string `json:"bank_name"`                     // 开户行
	BankCardFrontURL   string `json:"bank_card_front_url"`           // 银行卡正面照URL
	BankCardBackURL    string `json:"bank_card_back_url"`            // 银行卡反面照URL
	EducationCertURL   string `json:"education_cert_url"`            // 学历证书URL
	EmergencyContact   string `json:"emergency_contact"`             // 紧急联系人姓名
	EmergencyPhone     string `json:"emergency_phone"`               // 紧急联系人电话
	EmergencyRelation  string `json:"emergency_relation"`            // 与本人关系
}

// RegistrationResponse 登记表响应
type RegistrationResponse struct {
	ID             int64      `json:"id"`
	EmployeeID     *int64     `json:"employee_id"`
	Token          string     `json:"token"`
	Status         string     `json:"status"`
	ExpiresAt      time.Time  `json:"expires_at"`
	UsedAt         *time.Time `json:"used_at"`
	CreatedAt      time.Time  `json:"created_at"`
	EmployeeName   string     `json:"employee_name,omitempty"`
	DepartmentName string     `json:"department_name,omitempty"`
}

// RegistrationListQueryParams 登记列表查询参数
type RegistrationListQueryParams struct {
	Status   string `form:"status"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
}

// RegistrationDetailResponse 登记详情响应（公开接口，仅返回基础信息）
type RegistrationDetailResponse struct {
	Name         string `json:"name"`
	DepartmentID *int64 `json:"department_id"`
	Position     string `json:"position"`
	HireDate     string `json:"hire_date"`
	Status       string `json:"status"`
}
