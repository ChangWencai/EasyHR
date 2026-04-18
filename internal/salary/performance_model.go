package salary

import (
	"time"
)

// PerformanceCoefficient 绩效系数模型
type PerformanceCoefficient struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrgID       int64     `gorm:"column:org_id;index;not null" json:"org_id"`
	EmployeeID  int64     `gorm:"column:employee_id;not null" json:"employee_id"`
	Year        int       `gorm:"column:year;not null" json:"year"`
	Month       int       `gorm:"column:month;not null" json:"month"`
	Coefficient float64   `gorm:"column:coefficient;type:decimal(5,4);not null;default:1.0000" json:"coefficient"`
	CreatedBy   int64     `gorm:"column:created_by" json:"created_by"`
	UpdatedBy   int64     `gorm:"column:updated_by" json:"updated_by"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (PerformanceCoefficient) TableName() string { return "performance_coefficients" }
