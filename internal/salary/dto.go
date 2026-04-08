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
