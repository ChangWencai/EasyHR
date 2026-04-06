package employee

import "time"

// CreateInvitationRequest 老板创建邀请请求
type CreateInvitationRequest struct {
	Position string `json:"position" binding:"omitempty,max=100"` // 预设岗位（可选）
}

// ListInvitationsQuery 邀请列表查询参数
type ListInvitationsQuery struct {
	Status   string `form:"status"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
}

// SubmitInvitationRequest 员工提交信息（公开接口）
type SubmitInvitationRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=50"`
	Phone    string `json:"phone" binding:"required,len=11"`
	IDCard   string `json:"id_card" binding:"required,len=18"`
	Position string `json:"position" binding:"required,min=1,max=100"`
	HireDate string `json:"hire_date" binding:"required"` // YYYY-MM-DD
}

// InvitationDetailResponse 邀请详情响应（公开，不含敏感信息）
type InvitationDetailResponse struct {
	OrgName   string `json:"org_name"`
	Position  string `json:"position"`
	Status    string `json:"status"`
	ExpiresAt string `json:"expires_at"`
}

// InvitationListItem 邀请列表响应项
type InvitationListItem struct {
	ID           int64      `json:"id"`
	Token        string     `json:"token"`
	Position     string     `json:"position"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    time.Time  `json:"expires_at"`
	UsedAt       *time.Time `json:"used_at"`
	EmployeeID   *int64     `json:"employee_id"`
	EmployeeName string     `json:"employee_name,omitempty"`
}

// CreateInvitationResponse 创建邀请响应
type CreateInvitationResponse struct {
	Token     string `json:"token"`
	InviteURL string `json:"invite_url"` // /invite/{token}
	ExpiresAt string `json:"expires_at"`
}
