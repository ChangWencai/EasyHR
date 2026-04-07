package socialinsurance

import (
	"github.com/wencai/easyhr/internal/common/model"
	"gorm.io/datatypes"
)

// InsuranceItem 单个险种配置
type InsuranceItem struct {
	CompanyRate  float64 `json:"company_rate"`  // 企业缴费比例（如0.16=16%）
	PersonalRate float64 `json:"personal_rate"` // 个人缴费比例（如0.08=8%）
	BaseLower    float64 `json:"base_lower"`    // 基数下限
	BaseUpper    float64 `json:"base_upper"`    // 基数上限
}

// FiveInsurances 五险一金配置
type FiveInsurances struct {
	Pension      InsuranceItem `json:"pension"`       // 养老保险
	Medical      InsuranceItem `json:"medical"`       // 医疗保险
	Unemployment InsuranceItem `json:"unemployment"`  // 失业保险
	WorkInjury   InsuranceItem `json:"work_injury"`   // 工伤保险
	Maternity    InsuranceItem `json:"maternity"`     // 生育保险
	HousingFund  InsuranceItem `json:"housing_fund"`  // 住房公积金
}

// SocialInsurancePolicy 社保政策模型
// 政策库为全局共享数据，OrgID 设为 0（不使用 TenantScope）
type SocialInsurancePolicy struct {
	model.BaseModel
	CityID        int                               `gorm:"column:city_id;not null;index" json:"city_id"`
	EffectiveYear int                               `gorm:"column:effective_year;not null;index" json:"effective_year"`
	Config        datatypes.JSONType[FiveInsurances] `gorm:"column:config;type:jsonb" json:"config"`
}

// TableName 指定表名
func (SocialInsurancePolicy) TableName() string {
	return "social_insurance_policies"
}

// newJSONType 辅助函数：将 FiveInsurances 包装为 datatypes.JSONType
func newJSONType(data FiveInsurances) datatypes.JSONType[FiveInsurances] {
	return datatypes.NewJSONType(data)
}
