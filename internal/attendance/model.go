package attendance

import (
	"time"

	"github.com/wencai/easyhr/internal/common/model"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AttendanceRule 打卡规则（每企业最多一条）
type AttendanceRule struct {
	model.BaseModel
	Mode        string         `gorm:"column:mode;type:varchar(20);not null;comment:模式 fixed=固定时间 scheduled=按排班 free=自由工时"`
	WorkDays    string         `gorm:"column:work_days;type:varchar(50);comment:上班日 JSON数组 [1,2,3,4,5]"`
	WorkStart   string         `gorm:"column:work_start;type:varchar(5);comment:上班时间 HH:mm"`
	WorkEnd     string         `gorm:"column:work_end;type:varchar(5);comment:下班时间 HH:mm"`
	Location    string         `gorm:"column:location;type:varchar(200);comment:打卡位置"`
	ClockMethod string         `gorm:"column:clock_method;type:varchar(20);default:click;comment:打卡方式 click=手动 photo=拍照"`
	Holidays    datatypes.JSON `gorm:"column:holidays;type:jsonb;comment:节假日列表 JSON"`
}

// Holiday 节假日结构（存储在 AttendanceRule.Holidays JSONB）
type Holiday struct {
	Date string `json:"date"` // YYYY-MM-DD
	Name string `json:"name"` // 如：元旦
}

// Shift 班次定义（排班模式使用）
type Shift struct {
	model.BaseModel
	Name           string `gorm:"column:name;type:varchar(50);not null;comment:班次名称 如：早班/晚班"`
	WorkStart      string `gorm:"column:work_start;type:varchar(5);not null;comment:上班时间 HH:mm"`
	WorkEnd        string `gorm:"column:work_end;type:varchar(5);not null;comment:下班时间 HH:mm"`
	WorkDateOffset int    `gorm:"column:work_date_offset;default:0;comment:跨天班次偏移 0=当日 1=次日 D-13"`
}

// Schedule 员工排班记录（员工-日期-班次关联）
type Schedule struct {
	model.BaseModel
	EmployeeID int64     `gorm:"column:employee_id;not null;index:idx_schedule_emp_date,unique"`
	ShiftID    *int64    `gorm:"column:shift_id;index:idx_schedule_emp_date,unique;comment:null=休息日"`
	Date       time.Time `gorm:"column:date;type:date;not null;index:idx_schedule_emp_date,unique;comment:归属工作日 D-12"`
}

// ClockRecord 打卡记录（work_date 与 clock_time 分离 per D-14）
type ClockRecord struct {
	model.BaseModel
	EmployeeID int64     `gorm:"column:employee_id;not null;index;comment:员工ID"`
	WorkDate   time.Time `gorm:"column:work_date;type:date;not null;index;comment:归属工作日 D-12 D-14"`
	ClockTime  time.Time `gorm:"column:clock_time;not null;comment:实际打卡时间 D-14"`
	ClockType  string    `gorm:"column:clock_type;type:varchar(10);not null;comment:in=上班 out=下班"`
	PhotoURL   string    `gorm:"column:photo_url;type:varchar(500);comment:打卡照片"`
}

// AttendanceManualStats 手动修正假勤统计数据（ATT-08: 管理员可手动修改 per must_haves）
type AttendanceManualStats struct {
	model.BaseModel
	EmployeeID     int64     `gorm:"column:employee_id;not null;uniqueIndex:idx_manual_emp_month"`
	YearMonth      string    `gorm:"column:year_month;type:varchar(7);not null;uniqueIndex:idx_manual_emp_month"`
	LeaveDays      *float64  `gorm:"column:leave_days;default:0"`
	BusinessDays   *float64  `gorm:"column:business_days;default:0"`
	OutsideDays    *float64  `gorm:"column:outside_days;default:0"`
	MakeupCount    *int      `gorm:"column:makeup_count;default:0"`
	ShiftSwapCount *int      `gorm:"column:shift_swap_count;default:0"`
	OvertimeHours  *float64  `gorm:"column:overtime_hours;default:0"`
}

// TableName 设置表名
func (AttendanceRule) TableName() string          { return "attendance_rules" }
func (Shift) TableName() string                    { return "attendance_shifts" }
func (Schedule) TableName() string                 { return "attendance_schedules" }
func (ClockRecord) TableName() string              { return "attendance_clock_records" }
func (AttendanceManualStats) TableName() string    { return "attendance_manual_stats" }

// AutoMigrateTables 注册 GORM AutoMigrate（由 main.go 调用）
func AutoMigrateTables(db *gorm.DB) error {
	return db.AutoMigrate(&AttendanceRule{}, &Shift{}, &Schedule{}, &ClockRecord{}, &AttendanceManualStats{})
}
