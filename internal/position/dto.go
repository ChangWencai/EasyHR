package position

// CreatePositionRequest 创建岗位请求
type CreatePositionRequest struct {
	Name         string `json:"name" binding:"required,max=100"`
	DepartmentID *int64 `json:"department_id"`
	SortOrder    int    `json:"sort_order"`
}

// UpdatePositionRequest 更新岗位请求（指针类型支持部分更新）
type UpdatePositionRequest struct {
	Name         *string `json:"name"`
	DepartmentID *int64  `json:"department_id"`
	SortOrder    *int    `json:"sort_order"`
}

// PositionResponse 岗位响应
type PositionResponse struct {
	ID           int64  `json:"id"`
	OrgID        int64  `json:"org_id"`
	Name         string `json:"name"`
	DepartmentID *int64 `json:"department_id"`
	SortOrder    int    `json:"sort_order"`
}

// PositionSelectOptions 岗位下拉选项（分组：部门专属岗位 + 通用岗位）
type PositionSelectOptions struct {
	DeptPositions   []PositionOption `json:"dept_positions"`
	CommonPositions []PositionOption `json:"common_positions"`
	Unassigned      PositionOption   `json:"unassigned_option"`
}

// PositionOption 岗位选项
type PositionOption struct {
	ID   *int64 `json:"id"`
	Name string `json:"name"`
}
