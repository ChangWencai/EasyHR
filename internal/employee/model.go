package employee

import (
	"time"

	"github.com/wencai/easyhr/internal/common/model"
	"gorm.io/gorm"
)

// Employee 员工档案模型
// 敏感字段采用双列模式：encrypted 存储加密值，hash 存储哈希索引
// 唯一性通过 repository 层事务校验实现（兼容 SQLite 测试和 PostgreSQL 生产）
type Employee struct {
	model.BaseModel
	Name                    string     `gorm:"column:name;type:varchar(50);not null;index;comment:员工姓名" json:"name"`
	PhoneEncrypted          string     `gorm:"column:phone_encrypted;type:varchar(200);not null;comment:加密手机号" json:"-"`
	PhoneHash               string     `gorm:"column:phone_hash;type:varchar(64);not null;index;comment:手机号哈希索引（用于唯一性校验）" json:"-"`
	IDCardEncrypted         string     `gorm:"column:id_card_encrypted;type:varchar(200);comment:加密身份证号" json:"-"`
	IDCardHash              string     `gorm:"column:id_card_hash;type:varchar(64);index;comment:身份证号哈希索引" json:"-"`
	Gender                  string     `gorm:"column:gender;type:varchar(10);comment:性别" json:"gender"`
	BirthDate               *time.Time `gorm:"column:birth_date;type:date;comment:出生日期" json:"birth_date"`
	Position                string     `gorm:"column:position;type:varchar(100);not null;index;comment:岗位" json:"position"`
	HireDate                time.Time  `gorm:"column:hire_date;type:date;not null;comment:入职日期" json:"hire_date"`
	Status                  string     `gorm:"column:status;type:varchar(20);not null;default:pending;index;comment:状态（pending/probation/active/resigned）" json:"status"`
	UserID                  *int64     `gorm:"column:user_id;comment:关联用户ID（绑定账号后填写）" json:"user_id"`
	BankName                string     `gorm:"column:bank_name;type:varchar(100);comment:开户银行名称" json:"bank_name"`
	BankAccountEncrypted    string     `gorm:"column:bank_account_encrypted;type:varchar(200);comment:加密银行卡号" json:"-"`
	BankAccountHash         string     `gorm:"column:bank_account_hash;type:varchar(64);index;comment:银行卡号哈希索引" json:"-"`
	EmergencyContact        string     `gorm:"column:emergency_contact;type:varchar(50);comment:紧急联系人姓名" json:"emergency_contact"`
	EmergencyPhoneEncrypted string     `gorm:"column:emergency_phone_encrypted;type:varchar(200);comment:加密紧急联系人电话" json:"-"`
	EmergencyPhoneHash      string     `gorm:"column:emergency_phone_hash;type:varchar(64);comment:紧急联系人电话哈希索引" json:"-"`
	Address                 string     `gorm:"column:address;type:varchar(500);comment:居住地址" json:"address"`
	Remark                  string     `gorm:"column:remark;type:text;comment:备注" json:"remark"`
	ResignationDate         *time.Time `gorm:"column:resignation_date;type:date;comment:离职日期" json:"resignation_date"`
	ResignationReason       string     `gorm:"column:resignation_reason;type:varchar(500);comment:离职原因" json:"resignation_reason"`
}

// TableName 指定表名
func (Employee) TableName() string {
	return "employees"
}

// EmployeeStatus 员工状态常量
const (
	StatusPending   = "pending"   // 待入职
	StatusProbation = "probation" // 试用期
	StatusActive    = "active"    // 在职
	StatusResigned  = "resigned"  // 已离职
)

// extractFromIDCard 从18位身份证号提取性别和出生日期
// 第17位奇数=男，偶数=女
// 第7-14位=YYYYMMDD
func extractFromIDCard(idCard string) (gender string, birthDate time.Time, err error) {
	if len(idCard) != 18 {
		return "", time.Time{}, gorm.ErrRecordNotFound
	}

	// 提取性别：第17位（索引16）
	genderDigit := idCard[16]
	if genderDigit%2 == 1 {
		gender = "男"
	} else {
		gender = "女"
	}

	// 提取出生日期：第7-14位（索引6:14）
	birthDate, err = time.Parse("20060102", idCard[6:14])
	if err != nil {
		return "", time.Time{}, err
	}

	return gender, birthDate, nil
}
