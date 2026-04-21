package department

// CreateDepartmentRequest 创建部门请求
type CreateDepartmentRequest struct {
	Name      string `json:"name" binding:"required"`
	ParentID  *int64 `json:"parent_id"`
	SortOrder int    `json:"sort_order"`
}

// UpdateDepartmentRequest 更新部门请求（指针类型支持部分更新）
type UpdateDepartmentRequest struct {
	Name      *string `json:"name"`
	ParentID  *int64  `json:"parent_id"`
	SortOrder *int    `json:"sort_order"`
}

// DepartmentResponse 部门响应
type DepartmentResponse struct {
	ID        int64  `json:"id"`
	OrgID     int64  `json:"org_id"`
	Name      string `json:"name"`
	ParentID  *int64 `json:"parent_id"`
	SortOrder int    `json:"sort_order"`
}

// TreeNode 组织架构树节点
type TreeNode struct {
	ID        int64                  `json:"id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"` // "department" | "position" | "employee"
	Children  []*TreeNode            `json:"children,omitempty"`
	ItemStyle map[string]interface{} `json:"itemStyle,omitempty"`
	Label     map[string]interface{} `json:"label,omitempty"`
}

// DepartmentListQueryParams 部门列表查询参数
type DepartmentListQueryParams struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// SearchTreeRequest 搜索树请求
type SearchTreeRequest struct {
	Keyword string `form:"keyword" binding:"required,min=1"`
}

// TransferDeleteRequest 转移员工并删除部门请求（D-14-09）
type TransferDeleteRequest struct {
	TargetDepartmentID int64   `json:"target_department_id" binding:"required"`
	EmployeeIDs        []int64 `json:"employee_ids" binding:"required,min=1"`
}
