package model

type Organization struct {
	BaseModel
	Name         string `gorm:"column:name;type:varchar(200);not null" json:"name"`
	CreditCode   string `gorm:"column:credit_code;type:varchar(18);uniqueIndex:idx_org_credit_code,where:deleted_at IS NULL;not null" json:"credit_code"`
	City         string `gorm:"column:city;type:varchar(50);not null" json:"city"`
	ContactName  string `gorm:"column:contact_name;type:varchar(50)" json:"contact_name"`
	ContactPhone string `gorm:"column:contact_phone;type:varchar(200)" json:"contact_phone"`
	Status       string `gorm:"column:status;type:varchar(20);default:active" json:"status"`
}

func (Organization) TableName() string {
	return "organizations"
}
