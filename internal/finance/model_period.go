package finance

import (
	"time"

	"github.com/wencai/easyhr/internal/common/model"
)

// Period represents an accounting period (会计期间).
type Period struct {
	model.BaseModel
	Year             int          `gorm:"not null;uniqueIndex:idx_period_org_ym,priority:2;comment:年份" json:"year"`
	Month            int          `gorm:"not null;uniqueIndex:idx_period_org_ym,priority:3;comment:月份" json:"month"`
	Status           PeriodStatus `gorm:"type:varchar(10);default:'OPEN';comment:期间状态（OPEN/CLOSED）" json:"status"`
	VoucherNoCounter int          `gorm:"default:0;comment:凭证号计数器" json:"voucher_no_counter"`
	ClosedBy        *int64       `gorm:"column:closed_by;comment:结账人ID" json:"closed_by,omitempty"`
	ClosedAt        *time.Time   `gorm:"column:closed_at;comment:结账时间" json:"closed_at,omitempty"`
}

// TableName returns the table name for Period.
func (Period) TableName() string {
	return "periods"
}
