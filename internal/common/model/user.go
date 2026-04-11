package model

type User struct {
	BaseModel
	Phone     string `gorm:"column:phone;type:varchar(200)" json:"phone"`
	PhoneHash string `gorm:"column:phone_hash;type:varchar(64);uniqueIndex:idx_user_phone_hash,where:deleted_at IS NULL" json:"-"`
	Name      string `gorm:"column:name;type:varchar(50)" json:"name"`
	Role      string `gorm:"column:role;type:varchar(20);not null" json:"role"`
	Status       string `gorm:"column:status;type:varchar(20);default:active" json:"status"`
	PasswordHash string `gorm:"column:password_hash;type:varchar(200)" json:"-"`
}

func (User) TableName() string {
	return "users"
}
