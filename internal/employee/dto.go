package employee

import "time"

// CreateEmployeeRequest 创建员工请求（手动录入）
type CreateEmployeeRequest struct {
	Name             string   `json:"name" binding:"required,min=2,max=50"`
	Phone            string   `json:"phone" binding:"required,len=11"`
	IDCard           string   `json:"id_card" binding:"required,len=18"`
	Position         string   `json:"position" binding:"omitempty,max=100"` // 岗位名称（可选，position_id 关联时可不填）
	PositionID       *int64   `json:"position_id"`                         // 岗位ID（可选）
	DepartmentID     *int64   `json:"department_id"`                       // 部门ID（可选）
	HireDate         string   `json:"hire_date" binding:"required"`
	Salary           *float64 `json:"salary"`           // 正式薪资（月）
	ProbationSalary  *float64 `json:"probation_salary"`  // 试用期薪资（月）
	BankName         string   `json:"bank_name" binding:"omitempty,max=100"`
	BankAccount      string   `json:"bank_account" binding:"omitempty"`
	EmergencyContact string   `json:"emergency_contact" binding:"omitempty,max=50"`
	EmergencyPhone   string   `json:"emergency_phone" binding:"omitempty,len=11"`
	Address          string   `json:"address" binding:"omitempty,max=500"`
	Remark           string   `json:"remark" binding:"omitempty"`
}

// UpdateEmployeeRequest 更新员工请求（部分更新，仅非 nil 字段更新）
type UpdateEmployeeRequest struct {
	Name             *string `json:"name" binding:"omitempty,min=2,max=50"`
	Phone            *string `json:"phone" binding:"omitempty,len=11"`
	IDCard           *string `json:"id_card" binding:"omitempty,len=18"`
	Position         *string `json:"position" binding:"omitempty,min=1,max=100"`
	PositionID       *int64  `json:"position_id" binding:"omitempty"`
	DepartmentID     *int64  `json:"department_id" binding:"omitempty"`
	HireDate         *string `json:"hire_date" binding:"omitempty"`
	BankName         *string `json:"bank_name" binding:"omitempty,max=100"`
	BankAccount      *string `json:"bank_account" binding:"omitempty"`
	EmergencyContact *string `json:"emergency_contact" binding:"omitempty,max=50"`
	EmergencyPhone   *string `json:"emergency_phone" binding:"omitempty,len=11"`
	Address          *string `json:"address" binding:"omitempty,max=500"`
	Remark           *string `json:"remark" binding:"omitempty"`
}

// EmployeeResponse 员工列表响应（脱敏）
type EmployeeResponse struct {
	ID               int64      `json:"id"`
	Name             string     `json:"name"`
	Phone            string     `json:"phone"`
	IDCard           string     `json:"id_card"`
	Gender           string     `json:"gender"`
	BirthDate        *time.Time `json:"birth_date"`
	Position         string     `json:"position"`
	PositionID       *int64     `json:"position_id"`
	DepartmentID     *int64     `json:"department_id"`
	HireDate         time.Time  `json:"hire_date"`
	Status           string     `json:"status"`
	BankName         string     `json:"bank_name,omitempty"`
	BankAccount      string     `json:"bank_account,omitempty"`
	EmergencyContact string     `json:"emergency_contact,omitempty"`
	Address          string     `json:"address,omitempty"`
	Remark           string     `json:"remark,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

// SensitiveInfoResponse 敏感信息响应（完整解密，仅 OWNER/ADMIN 可访问）
type SensitiveInfoResponse struct {
	Phone          string `json:"phone"`
	IDCard         string `json:"id_card"`
	BankAccount    string `json:"bank_account,omitempty"`
	EmergencyPhone string `json:"emergency_phone,omitempty"`
}

// ListQueryParams 员工列表查询参数
type ListQueryParams struct {
	Name         string `form:"name"`
	Position     string `form:"position"`
	Phone        string `form:"phone"`
	Status       string `form:"status"`
	Search       string `form:"search"`        // 花名册综合搜索（姓名/手机号/岗位）
	DepartmentID *int64 `form:"department_id"` // 部门筛选
	Page         int    `form:"page,default=1"`
	PageSize     int    `form:"page_size,default=20"`
}

// EmployeeRosterItem 花名册列表项（聚合多表数据）
type EmployeeRosterItem struct {
	ID                 int64   `json:"id"`
	Name               string  `json:"name"`
	Status             string  `json:"status"`
	Position           string  `json:"position"`
	DepartmentID       *int64  `json:"department_id"`
	DepartmentName     string  `json:"department_name"`
	Phone              string  `json:"phone"`                // 脱敏后的手机号
	SalaryAmount       float64 `json:"salary_amount"`        // 岗位薪资（月）
	YearsOfService     string  `json:"years_of_service"`     // 在职年限（如 "2年3月"）
	ContractExpiryDays *int    `json:"contract_expiry_days"` // 合同到期天数（nil=无合同/无固定期限）
}
