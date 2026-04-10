package finance

import (
	"time"

	"github.com/wencai/easyhr/internal/common/model"
)

// Period represents an accounting period (会计期间).
type Period struct {
	model.BaseModel
	Year             int          `gorm:"not null;uniqueIndex:idx_period_org_ym,priority:2" json:"year"`
	Month            int          `gorm:"not null;uniqueIndex:idx_period_org_ym,priority:3" json:"month"`
	Status           PeriodStatus `gorm:"type:varchar(10);default:'OPEN'" json:"status"`
	VoucherNoCounter int          `gorm:"default:0" json:"voucher_no_counter"`
	ClosedBy        *int64       `gorm:"column:closed_by" json:"closed_by,omitempty"`
	ClosedAt        *time.Time  `gorm:"column:closed_at" json:"closed_at,omitempty"`
}

// TableName returns the table name for Period.
func (Period) TableName() string {
	return "periods"
}
