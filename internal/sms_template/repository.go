package sms_template

import (
	"github.com/wencai/easyhr/internal/common/middleware"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(tpl *SmsTemplate) error {
	return r.db.Create(tpl).Error
}

func (r *Repository) SeedPresets(orgID int64) error {
	presets := []SmsTemplate{
		{
			Name:         "验证码",
			Scene:        "verification_code",
			TemplateCode: "SMS_VERIFY_CODE",
			Content:      "您的验证码为${code}，有效期${expire}分钟，请勿泄露。",
			IsDefault:    true,
		},
		{
			Name:         "合同签署通知",
			Scene:        "contract_sign",
			TemplateCode: "SMS_CONTRACT_SIGN",
			Content:      "${name}您好，您有一份劳动合同待签署，请点击${link}完成签署，有效期${days}天。",
			IsDefault:    false,
		},
		{
			Name:         "入职邀请",
			Scene:        "registration",
			TemplateCode: "SMS_REGISTRATION",
			Content:      "欢迎加入${company}！请点击${url}完成入职信息填写，链接有效期7天。",
			IsDefault:    false,
		},
		{
			Name:         "工资条通知",
			Scene:        "salary_slip",
			TemplateCode: "SMS_SALARY_SLIP",
			Content:      "${name}您好，您${month}月的工资条已生成，请登录查看。",
			IsDefault:    false,
		},
		{
			Name:         "薪资解锁验证",
			Scene:        "salary_unlock",
			TemplateCode: "SMS_SALARY_UNLOCK",
			Content:      "您正在查看工资条，验证码为${code}，有效期5分钟。",
			IsDefault:    false,
		},
	}

	for i := range presets {
		presets[i].OrgID = orgID
		presets[i].CreatedBy = 0

		var count int64
		r.db.Model(&SmsTemplate{}).
			Where("org_id = ? AND name = ?", orgID, presets[i].Name).
			Count(&count)
		if count == 0 {
			if err := r.db.Create(&presets[i]).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Repository) FindByID(orgID, id int64) (*SmsTemplate, error) {
	var tpl SmsTemplate
	err := r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).First(&tpl).Error
	return &tpl, err
}

func (r *Repository) FindByName(orgID int64, name string) (*SmsTemplate, error) {
	var tpl SmsTemplate
	err := r.db.Scopes(middleware.TenantScope(orgID)).Where("name = ?", name).First(&tpl).Error
	return &tpl, err
}

func (r *Repository) FindByScene(orgID int64, scene string) (*SmsTemplate, error) {
	var tpl SmsTemplate
	err := r.db.Scopes(middleware.TenantScope(orgID)).Where("scene = ?", scene).First(&tpl).Error
	return &tpl, err
}

func (r *Repository) List(orgID int64, page, pageSize int) ([]SmsTemplate, int64, error) {
	var templates []SmsTemplate
	var total int64

	q := r.db.Model(&SmsTemplate{}).Scopes(middleware.TenantScope(orgID))
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("is_default DESC, id DESC").Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

func (r *Repository) Update(orgID, id int64, updates map[string]interface{}) error {
	result := r.db.Model(&SmsTemplate{}).Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).Updates(updates)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

func (r *Repository) ClearDefault(orgID int64) error {
	return r.db.Model(&SmsTemplate{}).Scopes(middleware.TenantScope(orgID)).
		Where("is_default = ?", true).
		Update("is_default", false).Error
}

func (r *Repository) Delete(orgID, id int64) error {
	return r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).Delete(&SmsTemplate{}).Error
}
