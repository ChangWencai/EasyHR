package city

import (
	"time"
)

// Level 常量：省市区层级
const (
	LevelProvince = 1 // 省/直辖市/自治区
	LevelCity     = 2 // 地级市/区
	LevelDistrict = 3 // 区县
)

// AreaCode 行政区划编码（对应 area_code 表）
type AreaCode struct {
	Code      int64     `gorm:"primaryKey;column:code" json:"code"`
	Name      string    `gorm:"column:name;type:varchar(128);not null" json:"name"`
	Level     int       `gorm:"column:level;type:smallint;not null" json:"level"`
	Pcode     int64     `gorm:"column:pcode" json:"pcode"`        // 父级编码，0 表示省级
	Category  int       `gorm:"column:category" json:"category"` // 1=省会/直辖市
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (AreaCode) TableName() string { return "area_code" }