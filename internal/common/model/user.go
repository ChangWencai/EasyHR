package model

type User struct {
	BaseModel
	Phone      string `gorm:"column:phone;type:varchar(200);comment:加密手机号" json:"phone"`
	PhoneHash  string `gorm:"column:phone_hash;type:varchar(64);uniqueIndex:idx_user_phone_hash,where:deleted_at IS NULL;comment:手机号哈希索引（用于唯一性校验）" json:"-"`
	Name       string `gorm:"column:name;type:varchar(50);comment:用户姓名" json:"name"`
	Role       string `gorm:"column:role;type:varchar(20);not null;comment:角色（owner/admin/member）" json:"role"`
	Status     string `gorm:"column:status;type:varchar(20);default:active;comment:状态（active/inactive）" json:"status"`
	PasswordHash string `gorm:"column:password_hash;type:varchar(200);comment:密码哈希值" json:"-"`
	Avatar     string `gorm:"column:avatar;type:text;comment:头像URL" json:"avatar"`
}

func (User) TableName() string {
	return "users"
}
