package attendance

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
