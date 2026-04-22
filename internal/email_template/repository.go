package email_template

import (
	"github.com/wencai/easyhr/internal/common/middleware"
	"gorm.io/gorm"
)

// Repository 邮箱模板数据访问层
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建 Repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建模板
func (r *Repository) Create(tpl *EmailTemplate) error {
	return r.db.Create(tpl).Error
}

// SeedPresets 为指定企业创建预置邮件模板（幂等）
func (r *Repository) SeedPresets(orgID int64) error {
	presets := []EmailTemplate{
		{
			Name:      "入职邀请",
			Subject:   "{{company}} 诚挚邀请您加入我们",
			Content:   "亲爱的 {{name}}：\n\n您好！感谢您应聘 {{company}} 的 {{position}} 岗位，您的申请已通过审核。\n\n请点击以下链接完成入职信息填写：\n{{invite_url}}\n\n该链接有效期为7天，请尽快完成操作。\n\n如有任何疑问，请联系我们。\n\n{{company}} 招聘团队",
			IsDefault: true,
		},
		{
			Name:      "入职确认提醒",
			Subject:   "{{company}} 提醒您：入职邀请即将过期",
			Content:   "亲爱的 {{name}}：\n\n您还未完成 {{company}} 的入职信息填写，您收到的邀请链接将于近期过期。\n\n请尽快点击以下链接完成操作：\n{{invite_url}}\n\n如已自行入职，请忽略此邮件。\n\n{{company}} 招聘团队",
			IsDefault: false,
		},
		{
			Name:      "入职流程说明",
			Subject:   "{{name}}，欢迎加入 {{company}}！入职准备须知",
			Content:   "亲爱的 {{name}}：\n\n恭喜您！您已成功加入 {{company}}，担任 {{position}} 一职。\n\n以下是入职准备须知：\n1. 请准备好身份证、学历证书原件\n2. 熟悉公司规章制度\n3. 按要求完成入职培训\n\n如有疑问，请联系 HR。\n\n{{company}}",
			IsDefault: false,
		},
		{
			Name:      "工资条发放通知",
			Subject:   "{{company}} 工资条发放通知",
			Content:   "亲爱的 {{name}}：\n\n您好！您本月的工资条已生成，请登录 {{company}} 薪资系统查看。\n\n如对工资明细有疑问，请联系 HR。\n\n{{company}}",
			IsDefault: false,
		},
		{
			Name:      "社保操作通知",
			Subject:   "{{company}} 社保操作通知",
			Content:   "亲爱的 {{name}}：\n\n您的社保关系变更已处理完成，请知悉。\n\n\n如需了解详细社保缴纳情况，请登录系统查看。\n\n如有疑问，请联系 HR。\n\n\n{{company}}",
			IsDefault: false,
		},
		{
			Name:      "合同签署提醒",
			Subject:   "{{company}} 劳动合同签署提醒",
			Content:   "亲爱的 {{name}}：\n\n您的劳动合同待签署，请尽快登录系统完成签署操作。\n劳动合同是保障您合法权益的重要文件，请务必及时签署。\n\n\n如有疑问，请联系 HR。\n\n{{company}}",
			IsDefault: false,
		},
		{
			Name:      "考勤异常提醒",
			Subject:   "{{company}} 考勤异常提醒",
			Content:   "亲爱的 {{name}}：\n\n您的本月考勤记录存在异常，请尽快登录系统核实并提交说明。\n\n如已自行处理或为特殊情况，请忽略此邮件。\n\n\n{{company}} HR",
			IsDefault: false,
		},
	}

	for i := range presets {
		presets[i].OrgID = orgID
		presets[i].CreatedBy = 0

		var count int64
		r.db.Model(&EmailTemplate{}).
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

// FindByID 根据 ID 查询
func (r *Repository) FindByID(orgID, id int64) (*EmailTemplate, error) {
	var tpl EmailTemplate
	err := r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).First(&tpl).Error
	return &tpl, err
}

// FindByName 根据名称查询（用于唯一性校验）
func (r *Repository) FindByName(orgID int64, name string) (*EmailTemplate, error) {
	var tpl EmailTemplate
	err := r.db.Scopes(middleware.TenantScope(orgID)).Where("name = ?", name).First(&tpl).Error
	return &tpl, err
}

// List 查询列表
func (r *Repository) List(orgID int64, page, pageSize int) ([]EmailTemplate, int64, error) {
	var templates []EmailTemplate
	var total int64

	q := r.db.Model(&EmailTemplate{}).Scopes(middleware.TenantScope(orgID))
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("is_default DESC, id DESC").Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

// Update 更新模板
func (r *Repository) Update(orgID, id int64, updates map[string]interface{}) error {
	result := r.db.Model(&EmailTemplate{}).Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).Updates(updates)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// ClearDefault 清除该企业的默认标记
func (r *Repository) ClearDefault(orgID int64) error {
	return r.db.Model(&EmailTemplate{}).Scopes(middleware.TenantScope(orgID)).
		Where("is_default = ?", true).
		Update("is_default", false).Error
}

// Delete 删除模板（软删除）
func (r *Repository) Delete(orgID, id int64) error {
	return r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).Delete(&EmailTemplate{}).Error
}
