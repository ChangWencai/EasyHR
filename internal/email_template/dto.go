package email_template

import "time"

// CreateTemplateRequest 创建模板请求
type CreateTemplateRequest struct {
	Name      string `json:"name" binding:"required,max=100"`
	Subject   string `json:"subject" binding:"required,max=200"`
	Content   string `json:"content" binding:"required"`
	IsDefault bool   `json:"is_default"`
}

// UpdateTemplateRequest 更新模板请求
type UpdateTemplateRequest struct {
	Name      *string `json:"name" binding:"omitempty,max=100"`
	Subject   *string `json:"subject" binding:"omitempty,max=200"`
	Content   *string `json:"content"`
	IsDefault *bool   `json:"is_default"`
}

// ListQuery 列表查询参数
type ListQuery struct {
	Page     int `form:"page,default=1"`
	PageSize int `form:"page_size,default=20"`
}

// TemplateResponse 模板响应
type TemplateResponse struct {
	ID        int64     `json:"id"`
	OrgID     int64     `json:"org_id"`
	Name      string    `json:"name"`
	Subject   string    `json:"subject"`
	Content   string    `json:"content"`
	IsDefault bool      `json:"is_default"`
	CreatedBy int64     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedBy int64     `json:"updated_by,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}
