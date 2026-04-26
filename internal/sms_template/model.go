package sms_template

import (
	"github.com/wencai/easyhr/internal/common/model"
)

type SmsTemplate struct {
	model.BaseModel
	Name         string `gorm:"column:name;type:varchar(100);not null;comment:模板名称" json:"name"`
	Scene        string `gorm:"column:scene;type:varchar(50);not null;comment:使用场景(verification_code/contract_sign/registration/salary_slip/salary_unlock)" json:"scene"`
	TemplateCode string `gorm:"column:template_code;type:varchar(50);not null;comment:阿里云短信模板代码" json:"template_code"`
	Content      string `gorm:"column:content;type:text;not null;comment:模板内容，支持变量占位符" json:"content"`
	IsDefault    bool   `gorm:"column:is_default;not null;default:false;comment:是否为默认模板" json:"is_default"`
}

func (SmsTemplate) TableName() string {
	return "sms_templates"
}
