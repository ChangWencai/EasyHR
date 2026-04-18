package employee

import (
	"time"

	"gorm.io/datatypes"
)

// BossResignRequest 老板办理离职请求（直接办理，立即生效）
type BossResignRequest struct {
	ResignationDate string `json:"resignation_date" binding:"required"` // YYYY-MM-DD
	Reason          string `json:"reason" binding:"required,min=1,max=500"`
}

// EmployeeResignRequest 员工申请离职请求（需老板审批）
type EmployeeResignRequest struct {
	ResignationDate string `json:"resignation_date" binding:"required"` // YYYY-MM-DD
	Reason          string `json:"reason" binding:"required,min=1,max=500"`
}

// RejectResignRequest 驳回离职申请请求
type RejectResignRequest struct {
	Reason string `json:"reason"` // 驳回原因（选填）
}

// UpdateChecklistRequest 更新交接清单请求
type UpdateChecklistRequest struct {
	ChecklistItems datatypes.JSON `json:"checklist_items" binding:"required"`
}

// OffboardingDetailResponse 离职详情响应
type OffboardingDetailResponse struct {
	ID              int64          `json:"id"`
	EmployeeID      int64          `json:"employee_id"`
	EmployeeName    string         `json:"employee_name"`
	Type            string         `json:"type"`
	ResignationDate time.Time      `json:"resignation_date"`
	Reason          string         `json:"reason"`
	Status          string         `json:"status"`
	ChecklistItems  datatypes.JSON `json:"checklist_items"`
	CompletedAt     *time.Time     `json:"completed_at"`
	ApprovedBy      *int64         `json:"approved_by"`
	ApprovedAt      *time.Time     `json:"approved_at"`
	CreatedAt       time.Time      `json:"created_at"`
}

// OffboardingListQueryParams 离职列表查询参数
type OffboardingListQueryParams struct {
	Status   string `form:"status"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
}
