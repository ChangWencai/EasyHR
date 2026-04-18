package salary

// ========== 调薪 DTO ==========

// AdjustmentRequest 单人调薪请求
type AdjustmentRequest struct {
	EmployeeID     int64   `json:"employee_id" binding:"required"`
	EffectiveMonth string  `json:"effective_month" binding:"required,len=7"`
	AdjustmentType string  `json:"adjustment_type" binding:"required,oneof=base_salary allowance bonus year_end_bonus other"`
	AdjustBy       string  `json:"adjust_by" binding:"required,oneof=amount ratio"`
	OldValue       float64 `json:"old_value" binding:"min=0"`
	NewValue       float64 `json:"new_value" binding:"min=0"`
}

// MassAdjustmentRequest 普调请求（按部门）
type MassAdjustmentRequest struct {
	DepartmentIDs  []int64 `json:"department_ids" binding:"required,min=1"`
	EffectiveMonth string  `json:"effective_month" binding:"required,len=7"`
	AdjustmentType string  `json:"adjustment_type" binding:"required,oneof=base_salary allowance bonus year_end_bonus other"`
	AdjustBy       string  `json:"adjust_by" binding:"required,oneof=amount ratio"`
	OldValue       float64 `json:"old_value" binding:"min=0"`
	NewValue       float64 `json:"new_value" binding:"min=0"`
}

// AdjustmentPreviewRequest 调薪预览请求
type AdjustmentPreviewRequest struct {
	EmployeeIDs    []int64 `json:"employee_ids"`
	DepartmentIDs  []int64 `json:"department_ids"`
	EffectiveMonth string  `json:"effective_month" binding:"required,len=7"`
	AdjustmentType string  `json:"adjustment_type" binding:"required,oneof=base_salary allowance bonus year_end_bonus other"`
	AdjustBy       string  `json:"adjust_by" binding:"required,oneof=amount ratio"`
	OldValue       float64 `json:"old_value" binding:"min=0"`
	NewValue       float64 `json:"new_value" binding:"min=0"`
}

// AdjustmentPreviewResponse 调薪预览响应
type AdjustmentPreviewResponse struct {
	EmployeeCount int     `json:"employee_count"`
	MonthlyImpact float64 `json:"monthly_impact"`
	AnnualImpact  float64 `json:"annual_impact"`
}

// AdjustmentListResponse 调薪记录列表项
type AdjustmentListResponse struct {
	ID             int64    `json:"id"`
	EmployeeID     *int64   `json:"employee_id"`
	DepartmentID   *int64   `json:"department_id"`
	Type           string   `json:"type"`
	EffectiveMonth string   `json:"effective_month"`
	AdjustmentType string   `json:"adjustment_type"`
	AdjustBy       string   `json:"adjust_by"`
	OldValue       float64  `json:"old_value"`
	NewValue       float64  `json:"new_value"`
	Status         string   `json:"status"`
	CreatedBy      int64    `json:"created_by"`
	CreatedAt      string   `json:"created_at"`
}
