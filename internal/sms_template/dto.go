package sms_template

import "time"

type CreateSmsTemplateRequest struct {
	Name         string `json:"name" binding:"required,max=100"`
	Scene        string `json:"scene" binding:"required,max=50"`
	TemplateCode string `json:"template_code" binding:"required,max=50"`
	Content      string `json:"content" binding:"required"`
	IsDefault    bool   `json:"is_default"`
}

type UpdateSmsTemplateRequest struct {
	Name         *string `json:"name" binding:"omitempty,max=100"`
	Scene        *string `json:"scene" binding:"omitempty,max=50"`
	TemplateCode *string `json:"template_code" binding:"omitempty,max=50"`
	Content      *string `json:"content"`
	IsDefault    *bool   `json:"is_default"`
}

type ListQuery struct {
	Page     int `form:"page,default=1"`
	PageSize int `form:"page_size,default=20"`
}

type SmsTemplateResponse struct {
	ID           int64     `json:"id"`
	OrgID        int64     `json:"org_id"`
	Name         string    `json:"name"`
	Scene        string    `json:"scene"`
	TemplateCode string    `json:"template_code"`
	Content      string    `json:"content"`
	IsDefault    bool      `json:"is_default"`
	CreatedBy    int64     `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedBy    int64     `json:"updated_by,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}
