package email_template

import (
	"github.com/wencai/easyhr/internal/common/model"
)

// EmailTemplate 邮箱模板模型
type EmailTemplate struct {
	model.BaseModel
	Name      string `gorm:"column:name;type:varchar(100);not null;comment:模板名称" json:"name"`
	Subject   string `gorm:"column:subject;type:varchar(200);not null;comment:邮件主题" json:"subject"`
	Content   string `gorm:"column:content;type:text;not null;comment:邮件正文，支持变量占位符" json:"content"`
	IsDefault bool   `gorm:"column:is_default;not null;default:false;comment:是否为默认模板" json:"is_default"`
}

// TableName 指定表名
func (EmailTemplate) TableName() string {
	return "email_templates"
}
