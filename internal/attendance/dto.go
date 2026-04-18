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
