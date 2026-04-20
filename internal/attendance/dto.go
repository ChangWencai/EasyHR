package attendance

import (
	"strconv"
	"strings"

	"github.com/wencai/easyhr/internal/common/model"
)

// === AttendanceRule DTO ===

type AttendanceRuleResponse struct {
	ID          int64     `json:"id"`
	Mode        string    `json:"mode"`
	WorkDays    []int     `json:"work_days"`
	WorkStart   string    `json:"work_start"`
	WorkEnd     string    `json:"work_end"`
	Location    string    `json:"location"`
	ClockMethod string    `json:"clock_method"`
	Holidays    []Holiday `json:"holidays"`
}

type SaveAttendanceRuleRequest struct {
	Mode        string    `json:"mode" binding:"required,oneof=fixed scheduled free"`
	WorkDays    []int     `json:"work_days"`
	WorkStart   string    `json:"work_start"`
	WorkEnd     string    `json:"work_end"`
	Location    string    `json:"location"`
	ClockMethod string    `json:"clock_method"`
	Holidays    []Holiday `json:"holidays"`
}

// === Shift DTO ===

type ShiftResponse struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	WorkStart      string `json:"work_start"`
	WorkEnd        string `json:"work_end"`
	WorkDateOffset int    `json:"work_date_offset"`
}

type CreateShiftRequest struct {
	Name           string `json:"name" binding:"required"`
	WorkStart      string `json:"work_start" binding:"required"`
	WorkEnd        string `json:"work_end" binding:"required"`
	WorkDateOffset int    `json:"work_date_offset"`
}

type UpdateShiftRequest struct {
	Name           string `json:"name" binding:"required"`
	WorkStart      string `json:"work_start" binding:"required"`
	WorkEnd        string `json:"work_end" binding:"required"`
	WorkDateOffset int    `json:"work_date_offset"`
}

// === Schedule DTO ===

type ScheduleRequest struct {
	EmployeeID int64  `json:"employee_id" binding:"required"`
	Date       string `json:"date" binding:"required"` // YYYY-MM-DD
	ShiftID    *int64 `json:"shift_id"`                // null=休息日
}

type BatchScheduleRequest struct {
	Schedules []ScheduleRequest `json:"schedules" binding:"required"`
}

type ScheduleResponse struct {
	EmployeeID int64  `json:"employee_id"`
	Date       string `json:"date"`
	ShiftID    *int64 `json:"shift_id"`
	ShiftName  string `json:"shift_name,omitempty"`
}

// === ClockRecord DTO ===

type ClockRecordResponse struct {
	EmployeeID     int64  `json:"employee_id"`
	EmployeeName   string `json:"employee_name"`
	DepartmentName string `json:"department_name,omitempty"`
	WorkDate       string `json:"work_date"`
	ClockInTime    string `json:"clock_in_time"`
	ClockOutTime   string `json:"clock_out_time"`
	ClockType      string `json:"clock_type,omitempty"`
	Status         string `json:"status"` // normal/late/absent/no_schedule/not_clocked_in
	ShiftName      string `json:"shift_name,omitempty"`
	PhotoURL       string `json:"photo_url,omitempty"`
}

type ClockLiveResponse struct {
	Date     string                `json:"date"`
	Records  []ClockRecordResponse `json:"records"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}

type CreateClockRecordRequest struct {
	EmployeeID int64  `json:"employee_id" binding:"required"`
	ClockTime  string `json:"clock_time" binding:"required"` // ISO 8601
	ClockType  string `json:"clock_type" binding:"required,oneof=in out"`
	PhotoURL   string `json:"photo_url"`
}

// === LeaveStats DTO ===

type LeaveStatsResponse struct {
	EmployeeID     int64   `json:"employee_id"`
	EmployeeName   string  `json:"employee_name"`
	YearMonth      string  `json:"year_month"`
	LeaveDays      float64 `json:"leave_days"`
	BusinessDays   float64 `json:"business_days"`
	OutsideDays    float64 `json:"outside_days"`
	MakeupCount    int     `json:"makeup_count"`
	ShiftSwapCount int     `json:"shift_swap_count"`
	OvertimeHours  float64 `json:"overtime_hours"`
	PendingDays    float64 `json:"pending_days"`
	ApprovedDays   float64 `json:"approved_days"`
}

type UpdateLeaveStatsRequest struct {
	LeaveDays      *float64 `json:"leave_days"`
	BusinessDays   *float64 `json:"business_days"`
	OutsideDays    *float64 `json:"outside_days"`
	MakeupCount    *int     `json:"makeup_count"`
	ShiftSwapCount *int     `json:"shift_swap_count"`
	OvertimeHours  *float64 `json:"overtime_hours"`
}

// === Approval DTO ===

type ApprovalResponse struct {
	ID           int64    `json:"id"`
	EmployeeID   int64    `json:"employee_id"`
	EmployeeName string   `json:"employee_name"`
	ApprovalType string   `json:"approval_type"`
	TypeName     string   `json:"type_name"`
	StartTime    string   `json:"start_time"`
	EndTime      string   `json:"end_time"`
	Duration     float64  `json:"duration"`
	LeaveType    string   `json:"leave_type,omitempty"`
	Reason       string   `json:"reason"`
	Status       string   `json:"status"`
	ApproverID   *int64   `json:"approver_id,omitempty"`
	ApproverName string   `json:"approver_name,omitempty"`
	ApprovedAt   string   `json:"approved_at,omitempty"`
	RejectedAt   string   `json:"rejected_at,omitempty"`
	RejectedNote string   `json:"rejected_note,omitempty"`
	CancelledAt  string   `json:"cancelled_at,omitempty"`
	Attachments  []string `json:"attachments"`
	CreatedAt    string   `json:"created_at"`
}

type ApprovalListResponse struct {
	List  []ApprovalResponse `json:"list"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
}

type PendingCountResponse struct {
	PendingCount int64 `json:"pending_count"`
}

type CreateApprovalRequest struct {
	ApprovalType string   `json:"approval_type" binding:"required"`
	StartTime    string   `json:"start_time" binding:"required"`
	EndTime      string   `json:"end_time" binding:"required"`
	Reason       string   `json:"reason"`
	LeaveType    string   `json:"leave_type"`
	Attachments  []string `json:"attachments"`
	CCUserIDs    []int64  `json:"cc_user_ids"`
}

type RejectApprovalRequest struct {
	Note string `json:"note"`
}

// === Compliance Report DTO ===

// ComplianceReportRequest 通用报表查询参数
type ComplianceReportRequest struct {
	YearMonth string `form:"year_month" binding:"required"` // YYYY-MM
	DeptIDs   string `form:"dept_ids"`                      // comma-separated, empty=all
}

// parseDeptIDs parses comma-separated dept_ids, returns nil if empty
func (r *ComplianceReportRequest) ParseDeptIDs() []int64 {
	if r.DeptIDs == "" || r.DeptIDs == "__all__" {
		return nil
	}
	ids := strings.Split(r.DeptIDs, ",")
	result := make([]int64, 0, len(ids))
	for _, s := range ids {
		s = strings.TrimSpace(s)
		if s == "" || s == "__all__" {
			continue
		}
		if id, err := strconv.ParseInt(s, 10, 64); err == nil {
			result = append(result, id)
		}
	}
	return result
}

// --- Overtime Report (COMP-05) ---

type OvertimeItem struct {
	EmployeeID     int64   `json:"employee_id"`
	EmployeeName   string  `json:"employee_name"`
	DepartmentName string  `json:"department_name"`
	HolidayHours   float64 `json:"holiday_hours"`     // 法定节假日加班 (0.5h rounding display)
	WeekdayHours   float64 `json:"weekday_hours"`     // 工作日延时加班 (0.5h rounding display)
	WeekendHours   float64 `json:"weekend_hours"`     // 周末加班 (0.5h rounding display)
	TotalHours     float64 `json:"total_hours"`       // 合计
}

type ComplianceOvertimeStats struct {
	TotalHolidayHours float64 `json:"total_holiday_hours"`
	TotalWeekdayHours float64 `json:"total_weekday_hours"`
	TotalWeekendHours float64 `json:"total_weekend_hours"`
}

type ComplianceOvertimeResponse struct {
	YearMonth string                 `json:"year_month"`
	Stats     ComplianceOvertimeStats `json:"stats"`
	List      []OvertimeItem          `json:"list"`
	Total     int64                   `json:"total"`
	Page      int                     `json:"page"`
	PageSize  int                     `json:"page_size"`
}

// --- Leave Compliance Report (COMP-06) ---

type LeaveItem struct {
	EmployeeID     int64   `json:"employee_id"`
	EmployeeName   string  `json:"employee_name"`
	DepartmentName string  `json:"department_name"`
	AnnualQuota    float64 `json:"annual_quota"`   // 管理员配置的年假总额度
	AnnualUsed     float64 `json:"annual_used"`    // 已用年假天数
	AnnualLeft     float64 `json:"annual_left"`    // 剩余年假 (quota - used)
	SickDays       float64 `json:"sick_days"`      // 病假天数
	PersonalDays   float64 `json:"personal_days"`  // 事假天数
}

type ComplianceLeaveStats struct {
	AnnualQuotaEmployeeCount int     `json:"annual_quota_employee_count"`
	TotalAnnualUsed          float64 `json:"total_annual_used"`
	TotalSickDays            float64 `json:"total_sick_days"`
	TotalPersonalDays        float64 `json:"total_personal_days"`
}

type ComplianceLeaveResponse struct {
	YearMonth string              `json:"year_month"`
	Stats     ComplianceLeaveStats `json:"stats"`
	List      []LeaveItem          `json:"list"`
	Total     int64                `json:"total"`
	Page      int                  `json:"page"`
	PageSize  int                  `json:"page_size"`
}

// --- Anomaly Report (COMP-07) ---

type AnomalyItem struct {
	EmployeeID      int64   `json:"employee_id"`
	EmployeeName    string  `json:"employee_name"`
	DepartmentName  string  `json:"department_name"`
	LateCount       int     `json:"late_count"`       // 迟到次数
	EarlyLeaveCount int     `json:"early_leave_count"` // 早退次数
	AbsentDays      float64 `json:"absent_days"`       // 缺勤天数
	AnomalyCount    int     `json:"anomaly_count"`     // 累计异常次数 (late + early + absent)
	IsAnomaly       bool    `json:"is_anomaly"`        // red-highlight: late>3 OR absent>1 per D-12-10
}

type ComplianceAnomalyStats struct {
	AnomalyEmployeeCount int     `json:"anomaly_employee_count"` // 异常员工数
	TotalLateCount      int     `json:"total_late_count"`
	TotalAbsentDays     float64 `json:"total_absent_days"`
}

type ComplianceAnomalyResponse struct {
	YearMonth string               `json:"year_month"`
	Stats     ComplianceAnomalyStats `json:"stats"`
	List      []AnomalyItem         `json:"list"`
	Total     int64                 `json:"total"`
	Page      int                   `json:"page"`
	PageSize  int                   `json:"page_size"`
}

// --- Monthly Compliance Report (COMP-08) ---

type MonthlyComplianceItem struct {
	EmployeeID        int64   `json:"employee_id"`
	EmployeeName      string  `json:"employee_name"`
	DepartmentName    string  `json:"department_name"`
	RequiredDays      float64 `json:"required_days"`      // 应出勤天数
	ActualDays        float64 `json:"actual_days"`        // 实际出勤天数
	LateCount         int     `json:"late_count"`
	EarlyLeaveCount   int     `json:"early_leave_count"`
	AbsentDays        float64 `json:"absent_days"`
	OvertimeHours     float64 `json:"overtime_hours"`   // 合计加班小时数
	AnnualLeaveDays   float64 `json:"annual_leave_days"` // 年假天数
	SickLeaveDays     float64 `json:"sick_leave_days"`   // 病假天数
	PersonalLeaveDays float64 `json:"personal_leave_days"` // 事假天数
	IsAnomaly         bool    `json:"is_anomaly"`       // red-highlight: late>3 OR absent>1
}

type ComplianceMonthlyStats struct {
	TotalRequiredDays   float64 `json:"total_required_days"`
	TotalActualDays     float64 `json:"total_actual_days"`
	TotalOvertimeHours  float64 `json:"total_overtime_hours"`
	TotalAbsentDays     float64 `json:"total_absent_days"`
	TotalAnomalyCount   int     `json:"total_anomaly_count"`
}

type ComplianceMonthlyResponse struct {
	YearMonth string                 `json:"year_month"`
	Stats     ComplianceMonthlyStats `json:"stats"`
	List      []MonthlyComplianceItem `json:"list"`
	Total     int64                  `json:"total"`
	Page      int                    `json:"page"`
	PageSize  int                    `json:"page_size"`
}

// --- Annual Leave Quota Config (COMP-06, D-12-08) ---

// AnnualLeaveQuota 企业员工年假额度配置表
type AnnualLeaveQuota struct {
	model.BaseModel
	EmployeeID int64   `gorm:"column:employee_id;not null;uniqueIndex:idx_annual_quota"`
	Year       int     `gorm:"column:year;type:int;not null;uniqueIndex:idx_annual_quota"`
	Quota      float64 `gorm:"column:quota;default:0;comment:年假总天数"`
}

func (AnnualLeaveQuota) TableName() string { return "attendance_annual_leave_quotas" }
