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
	CityID        int                               `gorm:"column:city_id;not null;index;comment:城市ID" json:"city_id"`
	EffectiveYear int                               `gorm:"column:effective_year;not null;index;comment:生效年份" json:"effective_year"`
	Config        datatypes.JSONType[FiveInsurances] `gorm:"column:config;type:jsonb;comment:五险一金配置（各险种缴费比例和基数上下限）" json:"config"`
}

// TableName 指定表名
func (SocialInsurancePolicy) TableName() string {
	return "social_insurance_policies"
}

// 社保记录状态常量
const (
	SIStatusPending = "pending" // 待参保
	SIStatusActive  = "active"  // 参保中
	SIStatusStopped = "stopped" // 停缴
)

// 变更历史类型常量
const (
	SIChangeEnroll     = "enroll"       // 参保
	SIChangeBaseAdjust = "base_adjust"  // 基数调整
	SIChangeStop       = "stop"         // 停缴
)

// SocialInsuranceRecord 参保记录（一条记录存所有险种明细）
type SocialInsuranceRecord struct {
	model.BaseModel
	EmployeeID    int64          `gorm:"column:employee_id;not null;index;comment:员工ID，外键到employees.id" json:"employee_id"`
	EmployeeName  string         `gorm:"column:employee_name;type:varchar(50);not null;comment:员工姓名" json:"employee_name"`
	CityID        int            `gorm:"column:city_id;not null;comment:参保城市ID" json:"city_id"`
	PolicyID      int64          `gorm:"column:policy_id;not null;comment:社保政策ID，外键到social_insurance_policies.id" json:"policy_id"`
	BaseAmount    float64        `gorm:"column:base_amount;not null;comment:社保缴费基数" json:"base_amount"`
	Status        string         `gorm:"column:status;type:varchar(20);not null;default:pending;comment:参保状态（pending/active/stopped）" json:"status"`
	StartMonth    string         `gorm:"column:start_month;type:varchar(7);not null;comment:参保起始月份（YYYY-MM）" json:"start_month"`
	EndMonth      *string        `gorm:"column:end_month;type:varchar(7);comment:参保结束月份（停缴时填写）" json:"end_month"`
	Details       datatypes.JSON `gorm:"column:details;type:jsonb;comment:各险种明细（JSON格式）" json:"details"`
	TotalCompany  float64        `gorm:"column:total_company;not null;comment:企业总缴费金额" json:"total_company"`
	TotalPersonal float64        `gorm:"column:total_personal;not null;comment:个人总缴费金额" json:"total_personal"`
}

// TableName 指定表名
func (SocialInsuranceRecord) TableName() string {
	return "social_insurance_records"
}

// ChangeHistory 变更历史
type ChangeHistory struct {
	model.BaseModel
	RecordID    int64          `gorm:"column:record_id;not null;index;comment:参保记录ID" json:"record_id"`
	EmployeeID  int64          `gorm:"column:employee_id;not null;index;comment:员工ID" json:"employee_id"`
	ChangeType  string         `gorm:"column:change_type;type:varchar(20);not null;comment:变更类型（enroll/base_adjust/stop）" json:"change_type"`
	BeforeValue datatypes.JSON `gorm:"column:before_value;type:jsonb;comment:变更前数据" json:"before_value"`
	AfterValue  datatypes.JSON `gorm:"column:after_value;type:jsonb;comment:变更后数据" json:"after_value"`
	Remark      string         `gorm:"column:remark;type:varchar(500);comment:备注" json:"remark"`
}

// TableName 指定表名
func (ChangeHistory) TableName() string {
	return "social_insurance_change_histories"
}

// newJSONType 辅助函数：将 FiveInsurances 包装为 datatypes.JSONType
func newJSONType(data FiveInsurances) datatypes.JSONType[FiveInsurances] {
	return datatypes.NewJSONType(data)
}
