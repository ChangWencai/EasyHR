package attendance

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
	"github.com/wencai/easyhr/internal/common/model"
)

// AttendanceService 核心业务逻辑
type AttendanceService struct {
	repo *AttendanceRepository
}

// NewAttendanceService 创建服务
func NewAttendanceService(repo *AttendanceRepository) *AttendanceService {
	return &AttendanceService{repo: repo}
}

// === AttendanceRule ===

func (s *AttendanceService) GetRule(ctx context.Context, orgID int64) (*AttendanceRuleResponse, error) {
	rule, err := s.repo.GetRule(orgID)
	if err != nil {
		return nil, fmt.Errorf("获取打卡规则失败: %w", err)
	}
	if rule == nil {
		return nil, nil
	}
	return toAttendanceRuleResponse(rule), nil
}

func (s *AttendanceService) SaveRule(ctx context.Context, orgID, userID int64, req *SaveAttendanceRuleRequest) (*AttendanceRuleResponse, error) {
	workDaysJSON, err := json.Marshal(req.WorkDays)
	if err != nil {
		return nil, fmt.Errorf("WorkDays 序列化失败: %w", err)
	}
	holidaysJSON, err := json.Marshal(req.Holidays)
	if err != nil {
		return nil, fmt.Errorf("Holidays 序列化失败: %w", err)
	}
	rule := &AttendanceRule{
		BaseModel:   model.BaseModel{OrgID: orgID, CreatedBy: userID, UpdatedBy: userID},
		Mode:        req.Mode,
		WorkDays:    string(workDaysJSON),
		WorkStart:   req.WorkStart,
		WorkEnd:     req.WorkEnd,
		Location:    req.Location,
		ClockMethod: req.ClockMethod,
		Holidays:    holidaysJSON,
	}
	if err := s.repo.UpsertRule(orgID, rule); err != nil {
		return nil, fmt.Errorf("保存打卡规则失败: %w", err)
	}
	return toAttendanceRuleResponse(rule), nil
}

func toAttendanceRuleResponse(rule *AttendanceRule) *AttendanceRuleResponse {
	var workDays []int
	_ = json.Unmarshal([]byte(rule.WorkDays), &workDays)
	var holidays []Holiday
	_ = json.Unmarshal(rule.Holidays, &holidays)
	return &AttendanceRuleResponse{
		ID:          rule.ID,
		Mode:        rule.Mode,
		WorkDays:    workDays,
		WorkStart:   rule.WorkStart,
		WorkEnd:     rule.WorkEnd,
		Location:    rule.Location,
		ClockMethod: rule.ClockMethod,
		Holidays:    holidays,
	}
}

// === Shift ===

func (s *AttendanceService) ListShifts(ctx context.Context, orgID int64) ([]ShiftResponse, error) {
	shifts, err := s.repo.ListShifts(orgID)
	if err != nil {
		return nil, fmt.Errorf("获取班次列表失败: %w", err)
	}
	return toShiftResponses(shifts), nil
}

func (s *AttendanceService) CreateShift(ctx context.Context, orgID, userID int64, req *CreateShiftRequest) (*ShiftResponse, error) {
	shift := &Shift{
		BaseModel:      model.BaseModel{OrgID: orgID, CreatedBy: userID, UpdatedBy: userID},
		Name:           req.Name,
		WorkStart:      req.WorkStart,
		WorkEnd:        req.WorkEnd,
		WorkDateOffset: req.WorkDateOffset,
	}
	if err := s.repo.CreateShift(shift); err != nil {
		return nil, fmt.Errorf("创建班次失败: %w", err)
	}
	return toShiftResponse(shift), nil
}

func (s *AttendanceService) UpdateShift(ctx context.Context, orgID, userID, shiftID int64, req *UpdateShiftRequest) (*ShiftResponse, error) {
	shift, err := s.repo.GetShift(orgID, shiftID)
	if err != nil {
		return nil, fmt.Errorf("获取班次失败: %w", err)
	}
	if shift == nil {
		return nil, fmt.Errorf("班次不存在")
	}
	shift.Name = req.Name
	shift.WorkStart = req.WorkStart
	shift.WorkEnd = req.WorkEnd
	shift.WorkDateOffset = req.WorkDateOffset
	shift.UpdatedBy = userID
	if err := s.repo.UpdateShift(shift); err != nil {
		return nil, fmt.Errorf("更新班次失败: %w", err)
	}
	return toShiftResponse(shift), nil
}

func (s *AttendanceService) DeleteShift(ctx context.Context, orgID int64, shiftID int64) error {
	return s.repo.DeleteShift(orgID, shiftID)
}

func toShiftResponses(shifts []Shift) []ShiftResponse {
	result := make([]ShiftResponse, len(shifts))
	for i, s := range shifts {
		result[i] = *toShiftResponse(&s)
	}
	return result
}

func toShiftResponse(s *Shift) *ShiftResponse {
	return &ShiftResponse{
		ID:             s.ID,
		Name:           s.Name,
		WorkStart:      s.WorkStart,
		WorkEnd:        s.WorkEnd,
		WorkDateOffset: s.WorkDateOffset,
	}
}

// === Schedule ===

func (s *AttendanceService) ListSchedules(ctx context.Context, orgID int64, startDate, endDate string, employeeID *int64) ([]ScheduleResponse, error) {
	schedules, err := s.repo.ListSchedules(orgID, startDate, endDate, employeeID)
	if err != nil {
		return nil, fmt.Errorf("获取排班列表失败: %w", err)
	}
	return toScheduleResponses(schedules), nil
}

func (s *AttendanceService) BatchUpsertSchedules(ctx context.Context, orgID, userID int64, req *BatchScheduleRequest) error {
	now := time.Now()
	schedules := make([]Schedule, len(req.Schedules))
	for i, r := range req.Schedules {
		schedules[i] = Schedule{
			BaseModel: model.BaseModel{
				OrgID:     orgID,
				CreatedBy: userID,
				UpdatedBy: userID,
				CreatedAt: now,
				UpdatedAt: now,
			},
			EmployeeID: r.EmployeeID,
			Date:       parseDate(r.Date),
			ShiftID:    r.ShiftID,
		}
	}
	return s.repo.BatchUpsertSchedules(orgID, schedules)
}

func parseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func toScheduleResponses(schedules []Schedule) []ScheduleResponse {
	result := make([]ScheduleResponse, len(schedules))
	for i, s := range schedules {
		result[i] = ScheduleResponse{
			EmployeeID: s.EmployeeID,
			Date:       s.Date.Format("2006-01-02"),
			ShiftID:    s.ShiftID,
		}
	}
	return result
}

// === ClockRecord & ClockLive ===

// clockPair 存储一个员工当天的上下班打卡记录
type clockPair struct {
	ClockIn  *ClockRecord
	ClockOut *ClockRecord
}

func (s *AttendanceService) GetClockLive(ctx context.Context, orgID int64, date string, departmentID *int64, page, pageSize int) (*ClockLiveResponse, error) {
	// 1. 获取当天打卡记录
	records, err := s.repo.ListClockRecordsByDate(orgID, date)
	if err != nil {
		return nil, fmt.Errorf("获取打卡记录失败: %w", err)
	}

	// 2. 构建员工打卡映射 (employeeID -> {in, out})
	clockMap := make(map[int64]*clockPair)
	for i := range records {
		r := &records[i]
		pair, ok := clockMap[r.EmployeeID]
		if !ok {
			pair = &clockPair{}
			clockMap[r.EmployeeID] = pair
		}
		if r.ClockType == "in" {
			pair.ClockIn = r
		} else if r.ClockType == "out" {
			pair.ClockOut = r
		}
	}

	// 3. 获取考勤规则用于判断迟到
	rule, err := s.repo.GetRule(orgID)
	if err != nil {
		return nil, fmt.Errorf("获取考勤规则失败: %w", err)
	}

	// 4. 获取所有在职员工（全员列表，包含未打卡的）
	allEmps, err := s.repo.ListAllActiveEmployees(orgID)
	if err != nil {
		return nil, fmt.Errorf("获取员工列表失败: %w", err)
	}

	// 5. 构建 ClockRecordResponse 列表
	items := make([]ClockRecordResponse, 0, len(allEmps))
	for _, emp := range allEmps {
		pair := clockMap[emp.ID]
		item := ClockRecordResponse{
			EmployeeID:     emp.ID,
			EmployeeName:   emp.Name,
			DepartmentName: emp.DepartmentName,
			WorkDate:       date,
			Status:         "not_clocked_in",
		}

		if pair != nil && pair.ClockIn != nil {
			item.ClockInTime = pair.ClockIn.ClockTime.Format("15:04")
			item.Status = "normal"
			// 判断迟到
			if rule != nil && rule.WorkStart != "" {
				clockHHMM := pair.ClockIn.ClockTime.Format("15:04")
				if clockHHMM > rule.WorkStart {
					item.Status = "late"
				}
			}
		}

		if pair != nil && pair.ClockOut != nil {
			item.ClockOutTime = pair.ClockOut.ClockTime.Format("15:04")
		}

		items = append(items, item)
	}

	// 6. 分页
	total := int64(len(items))
	start := (page - 1) * pageSize
	if start > int(total) {
		start = int(total)
	}
	end := start + pageSize
	if end > int(total) {
		end = int(total)
	}
	pagedItems := items[start:end]

	return &ClockLiveResponse{
		Date:     date,
		Records:  pagedItems,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// CreateClockRecord 创建打卡记录（管理员代打/邀请点签）
func (s *AttendanceService) CreateClockRecord(ctx context.Context, orgID, userID int64, req *CreateClockRecordRequest) (*ClockRecordResponse, error) {
	clockTime, err := time.Parse(time.RFC3339, req.ClockTime)
	if err != nil {
		return nil, fmt.Errorf("打卡时间格式错误: %w", err)
	}

	// work_date = 打卡日期（按班次起始日，TODO: 后续根据 work_date_offset 调整）
	workDate := clockTime.Truncate(24 * time.Hour)

	record := &ClockRecord{
		BaseModel:  model.BaseModel{OrgID: orgID, CreatedBy: userID, UpdatedBy: userID},
		EmployeeID: req.EmployeeID,
		WorkDate:   workDate,
		ClockTime:  clockTime,
		ClockType:  req.ClockType,
		PhotoURL:   req.PhotoURL,
	}

	if err := s.repo.CreateClockRecord(record); err != nil {
		return nil, fmt.Errorf("创建打卡记录失败: %w", err)
	}

	// 获取考勤规则判断状态
	status := "normal"
	rule, _ := s.repo.GetRule(orgID)
	if rule != nil && rule.WorkStart != "" && req.ClockType == "in" {
		clockHHMM := clockTime.Format("15:04")
		if clockHHMM > rule.WorkStart {
			status = "late"
		}
	}

	// 获取员工姓名
	emps, _ := s.repo.ListEmployeesByIDs(orgID, []int64{req.EmployeeID})
	empName := ""
	if len(emps) > 0 {
		empName = emps[0].Name
	}

	resp := &ClockRecordResponse{
		EmployeeID:   req.EmployeeID,
		EmployeeName: empName,
		WorkDate:     workDate.Format("2006-01-02"),
		Status:       status,
		PhotoURL:     req.PhotoURL,
	}
	if req.ClockType == "in" {
		resp.ClockInTime = clockTime.Format("15:04")
	} else {
		resp.ClockOutTime = clockTime.Format("15:04")
	}
	return resp, nil
}

// GetLeaveStats 获取员工假勤统计数据（从 Approval 表聚合 per ATT-07）
func (s *AttendanceService) GetLeaveStats(ctx context.Context, orgID int64, employeeID int64, yearMonth string) (*LeaveStatsResponse, error) {
	// 获取员工姓名
	emps, _ := s.repo.ListEmployeesByIDs(orgID, []int64{employeeID})
	empName := ""
	if len(emps) > 0 {
		empName = emps[0].Name
	}

	// 从 Approval 表聚合假勤数据
	approvals, err := s.repo.ListApprovalsByEmployeeMonth(orgID, employeeID, yearMonth)
	if err != nil {
		return nil, fmt.Errorf("获取假勤数据失败: %w", err)
	}

	var leaveDays, pendingDays, approvedDays float64
	var businessDays, outsideDays float64
	var makeupCount, shiftSwapCount int
	var overtimeHours float64

	for _, a := range approvals {
		hours := a.Duration
		days := hours / 8.0 // 按每天8小时折算

		switch a.ApprovalType {
		case ApprovalTypePersonalLeave, ApprovalTypeSickLeave, ApprovalTypePTO,
			ApprovalTypeAnnualLeave, ApprovalTypeMarriageLeave,
			ApprovalTypeMaternityLeave, ApprovalTypePaternityLeave:
			leaveDays += days
			if a.Status == ApprovalStatusApproved {
				approvedDays += days
			} else if a.Status == ApprovalStatusPending {
				pendingDays += days
			}
		case ApprovalTypeBusinessTrip:
			businessDays += days
		case ApprovalTypeOutside:
			outsideDays += days
		case ApprovalTypeMakeup:
			makeupCount++
		case ApprovalTypeShiftSwap:
			shiftSwapCount++
		case ApprovalTypeOvertime:
			overtimeHours += hours
		}
	}

	// 获取手动修正数据（覆盖计算值）
	stats, _ := s.repo.GetManualStats(orgID, employeeID, yearMonth)

	resp := &LeaveStatsResponse{
		EmployeeID:     employeeID,
		EmployeeName:   empName,
		YearMonth:      yearMonth,
		LeaveDays:      roundHalf(leaveDays),
		BusinessDays:   roundHalf(businessDays),
		OutsideDays:    roundHalf(outsideDays),
		MakeupCount:    makeupCount,
		ShiftSwapCount: shiftSwapCount,
		OvertimeHours:  roundHalf(overtimeHours), // D-07: 0.5h 取整
		PendingDays:    roundHalf(pendingDays),
		ApprovedDays:   roundHalf(approvedDays),
	}

	// 手动修正覆盖（如果存在）
	if stats != nil {
		if stats.LeaveDays != nil {
			resp.LeaveDays = *stats.LeaveDays
		}
		if stats.BusinessDays != nil {
			resp.BusinessDays = *stats.BusinessDays
		}
		if stats.OutsideDays != nil {
			resp.OutsideDays = *stats.OutsideDays
		}
		if stats.MakeupCount != nil {
			resp.MakeupCount = *stats.MakeupCount
		}
		if stats.ShiftSwapCount != nil {
			resp.ShiftSwapCount = *stats.ShiftSwapCount
		}
		if stats.OvertimeHours != nil {
			resp.OvertimeHours = *stats.OvertimeHours
		}
	}

	return resp, nil
}

func roundHalf(val float64) float64 {
	return float64(int(val/0.5+0.5)) * 0.5
}

// UpdateLeaveStats 手动修正假勤统计数据（ATT-08）
func (s *AttendanceService) UpdateLeaveStats(ctx context.Context, orgID, userID int64, employeeID int64, yearMonth string, req *UpdateLeaveStatsRequest) error {
	stats := &AttendanceManualStats{
		BaseModel:      model.BaseModel{OrgID: orgID, UpdatedBy: userID},
		EmployeeID:     employeeID,
		YearMonth:      yearMonth,
		LeaveDays:      req.LeaveDays,
		BusinessDays:   req.BusinessDays,
		OutsideDays:    req.OutsideDays,
		MakeupCount:    req.MakeupCount,
		ShiftSwapCount: req.ShiftSwapCount,
		OvertimeHours:  req.OvertimeHours,
	}
	return s.repo.UpsertManualStats(stats)
}

// === 出勤月报 ===

type MonthlyReportItem struct {
	EmployeeID     int64   `json:"employee_id"`
	EmployeeName   string  `json:"employee_name"`
	DepartmentName string  `json:"department_name"`
	ActualDays     float64 `json:"actual_days"`
	RequiredDays   float64 `json:"required_days"`
	OvertimeHours  float64 `json:"overtime_hours"`
	AbsentDays     float64 `json:"absent_days"`
	LeaveDays      float64 `json:"leave_days"`
	BusinessDays   float64 `json:"business_days"`
	AttendanceRate float64 `json:"attendance_rate"`
	LateCount      int     `json:"late_count"`
}

type MonthlyStats struct {
	TotalActualDays    float64 `json:"total_actual_days"`
	TotalRequiredDays  float64 `json:"total_required_days"`
	TotalOvertimeHours float64 `json:"total_overtime_hours"`
	TotalAbsentDays    float64 `json:"total_absent_days"`
}

type MonthlyReportResponse struct {
	YearMonth string              `json:"year_month"`
	Stats     MonthlyStats        `json:"stats"`
	List      []MonthlyReportItem `json:"list"`
	Total     int64               `json:"total"`
	Page      int                 `json:"page"`
	PageSize  int                 `json:"page_size"`
}

type DailyRecord struct {
	Date      string `json:"date"`
	ClockIn   string `json:"clock_in"`
	ClockOut  string `json:"clock_out"`
	Status    string `json:"status"`
	IsHoliday bool   `json:"is_holiday"`
	IsWeekend bool   `json:"is_weekend"`
	Symbol    string `json:"symbol"`
}

type DailyRecordsResponse struct {
	EmployeeID int64         `json:"employee_id"`
	YearMonth  string        `json:"year_month"`
	Records    []DailyRecord `json:"records"`
}

func (s *AttendanceService) GetMonthlyReport(ctx context.Context, orgID int64, yearMonth string, page, pageSize int) (*MonthlyReportResponse, error) {
	reports, total, err := s.repo.ListAttendanceMonthly(orgID, yearMonth, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("获取月报数据失败: %w", err)
	}

	var totalActual, totalRequired, totalOvertime, totalAbsent float64
	items := make([]MonthlyReportItem, len(reports))
	for i, r := range reports {
		overtimeDisplay := roundHalf(r.OvertimeHours)
		rate := 0.0
		if r.RequiredDays > 0 {
			rate = float64(int(r.ActualDays/r.RequiredDays*100*10)) / 10
		}
		items[i] = MonthlyReportItem{
			EmployeeID:     r.EmployeeID,
			ActualDays:     r.ActualDays,
			RequiredDays:   r.RequiredDays,
			OvertimeHours:  overtimeDisplay,
			AbsentDays:     r.AbsentDays,
			LeaveDays:      r.LeaveDays,
			BusinessDays:   r.BusinessDays,
			AttendanceRate: rate,
			LateCount:      r.LateCount,
		}
		totalActual += r.ActualDays
		totalRequired += r.RequiredDays
		totalOvertime += r.OvertimeHours
		totalAbsent += r.AbsentDays
	}

	return &MonthlyReportResponse{
		YearMonth: yearMonth,
		Stats: MonthlyStats{
			TotalActualDays:    roundHalf(totalActual),
			TotalRequiredDays:  roundHalf(totalRequired),
			TotalOvertimeHours: roundHalf(totalOvertime),
			TotalAbsentDays:    roundHalf(totalAbsent),
		},
		List:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *AttendanceService) GetDailyRecords(ctx context.Context, orgID int64, employeeID int64, yearMonth string) (*DailyRecordsResponse, error) {
	rule, err := s.repo.GetRule(orgID)
	if err != nil {
		return nil, fmt.Errorf("获取考勤规则失败: %w", err)
	}
	ruleEngine := NewRuleEngine(rule)

	clockMap, err := s.repo.GetDailyClockRecords(orgID, employeeID, yearMonth)
	if err != nil {
		return nil, fmt.Errorf("获取打卡记录失败: %w", err)
	}

	parsed, _ := time.Parse("2006-01", yearMonth)
	year, month, _ := parsed.Date()
	daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()

	records := make([]DailyRecord, daysInMonth)
	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		dateStr := date.Format("2006-01-02")
		weekday := date.Weekday()
		isWeekend := weekday == time.Saturday || weekday == time.Sunday
		isHoliday := ruleEngine.IsHoliday(date)

		dayClocks := clockMap[dateStr]
		clockIn := dayClocks["in"]
		clockOut := dayClocks["out"]

		status := "normal"
		symbol := "√"
		if clockIn == "" && clockOut == "" {
			if !isWeekend && !isHoliday {
				status = "absent"
				symbol = "缺"
			} else {
				status = "no_schedule"
				symbol = "--"
			}
		} else if clockIn != "" && rule != nil && rule.WorkStart != "" {
			if clockIn > rule.WorkStart {
				status = "late"
				symbol = "迟到"
			}
		}

		records[day-1] = DailyRecord{
			Date: dateStr, ClockIn: clockIn, ClockOut: clockOut,
			Status: status, IsHoliday: isHoliday, IsWeekend: isWeekend, Symbol: symbol,
		}
	}

	return &DailyRecordsResponse{EmployeeID: employeeID, YearMonth: yearMonth, Records: records}, nil
}

func (s *AttendanceService) ExportMonthlyExcel(ctx context.Context, orgID int64, yearMonth string) ([]byte, string, error) {
	report, err := s.GetMonthlyReport(ctx, orgID, yearMonth, 1, 1000)
	if err != nil {
		return nil, "", fmt.Errorf("获取月报数据失败: %w", err)
	}

	f := excelize.NewFile()
	sheet := "出勤月报"
	index, _ := f.NewSheet(sheet)
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	headers := []string{"姓名", "部门", "实际出勤(天)", "应出勤(天)", "加班时长(小时)", "缺勤(天)", "请假(天)", "出差(天)", "出勤率(%)"}
	style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}, Fill: excelize.Fill{Type: "pattern", Color: []string{"#E6E6E6"}, Pattern: 1}})
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, style)
	}

	for rowIdx, item := range report.List {
		row := rowIdx + 2
		vals := []interface{}{item.EmployeeName, item.DepartmentName, item.ActualDays, item.RequiredDays, item.OvertimeHours, item.AbsentDays, item.LeaveDays, item.BusinessDays, fmt.Sprintf("%.1f%%", item.AttendanceRate)}
		for colIdx, val := range vals {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, row)
			f.SetCellValue(sheet, cell, val)
		}
	}

	summaryRow := len(report.List) + 2
	f.SetCellValue(sheet, fmt.Sprintf("A%d", summaryRow), "合计")
	f.SetCellValue(sheet, fmt.Sprintf("C%d", summaryRow), report.Stats.TotalActualDays)
	f.SetCellValue(sheet, fmt.Sprintf("D%d", summaryRow), report.Stats.TotalRequiredDays)
	f.SetCellValue(sheet, fmt.Sprintf("E%d", summaryRow), roundHalf(report.Stats.TotalOvertimeHours))
	f.SetCellValue(sheet, fmt.Sprintf("F%d", summaryRow), report.Stats.TotalAbsentDays)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", fmt.Errorf("生成 Excel 失败: %w", err)
	}
	return buf.Bytes(), fmt.Sprintf("出勤月报_%s.xlsx", yearMonth), nil
}

// === Compliance Reports (COMP-05~COMP-08) ===

// GetComplianceOvertime returns overtime statistics by category per employee
// Overtime types: holiday加班(法定节假日), weekday加班(工作日延时), weekend加班
// OvertimeHours from approved overtime-type Approvals + clock records after work_end on weekdays
func (s *AttendanceService) GetComplianceOvertime(ctx context.Context, orgID int64, req *ComplianceReportRequest, page, pageSize int) (*ComplianceOvertimeResponse, error) {
	deptIDs := req.ParseDeptIDs()
	employees, err := s.repo.ListEmployeesByOrgWithDept(orgID, deptIDs)
	if err != nil {
		return nil, fmt.Errorf("获取员工列表失败: %w", err)
	}
	if len(employees) == 0 {
		return &ComplianceOvertimeResponse{YearMonth: req.YearMonth, List: []OvertimeItem{}, Total: 0}, nil
	}

	empIDs := make([]int64, len(employees))
	for i, e := range employees {
		empIDs[i] = e.EmployeeID
	}

	approvals, err := s.repo.ListApprovalsByMonth(orgID, empIDs, req.YearMonth)
	if err != nil {
		return nil, fmt.Errorf("获取加班审批失败: %w", err)
	}

	rule, _ := s.repo.GetRule(orgID)
	ruleEngine := NewRuleEngine(rule)

	// Compute holiday vs weekday vs weekend hours per employee from approved overtime approvals
	type overtimeBreakdown struct {
		Holiday, Weekday, Weekend float64
	}
	breakdown := make(map[int64]*overtimeBreakdown)
	for _, empID := range empIDs {
		breakdown[empID] = &overtimeBreakdown{}
	}
	for _, a := range approvals {
		if a.ApprovalType == ApprovalTypeOvertime && a.Status == ApprovalStatusApproved {
			if breakdown[a.EmployeeID] == nil {
				breakdown[a.EmployeeID] = &overtimeBreakdown{}
			}
			hours := roundHalf(a.Duration) // D-12-04: 0.5h rounding
			category := ruleEngine.ClassifyOvertimeCategory(a.StartTime, a.EndTime) // D-12-03
			switch category {
			case "holiday":
				breakdown[a.EmployeeID].Holiday += hours
			case "weekday":
				breakdown[a.EmployeeID].Weekday += hours
			case "weekend":
				breakdown[a.EmployeeID].Weekend += hours
			}
		}
	}

	// Build item list
	list := make([]OvertimeItem, 0, len(employees))
	var totalHoliday, totalWeekday, totalWeekend float64
	for _, emp := range employees {
		bd := breakdown[emp.EmployeeID]
		holidayH, weekdayH, weekendH := 0.0, 0.0, 0.0
		if bd != nil {
			holidayH = bd.Holiday
			weekdayH = bd.Weekday
			weekendH = bd.Weekend
		}
		totalHours := roundHalf(holidayH + weekdayH + weekendH)
		list = append(list, OvertimeItem{
			EmployeeID:     emp.EmployeeID,
			EmployeeName:   emp.EmployeeName,
			DepartmentName: emp.DepartmentName,
			HolidayHours:   roundHalf(holidayH),
			WeekdayHours:   roundHalf(weekdayH),
			WeekendHours:   roundHalf(weekendH),
			TotalHours:     totalHours,
		})
		totalHoliday += holidayH
		totalWeekday += weekdayH
		totalWeekend += weekendH
	}

	// Paginate
	total := int64(len(list))
	start := (page - 1) * pageSize
	if start > int(total) {
		start = int(total)
	}
	end := start + pageSize
	if end > int(total) {
		end = int(total)
	}
	paged := list[start:end]

	return &ComplianceOvertimeResponse{
		YearMonth: req.YearMonth,
		Stats: ComplianceOvertimeStats{
			TotalHolidayHours: roundHalf(totalHoliday),
			TotalWeekdayHours: roundHalf(totalWeekday),
			TotalWeekendHours: roundHalf(totalWeekend),
		},
		List:     paged,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetComplianceLeave returns leave compliance statistics per employee
// Annual leave quota from AnnualLeaveQuota table (admin configured per D-12-08)
// Sick/personal leave from approved leave-type approvals
func (s *AttendanceService) GetComplianceLeave(ctx context.Context, orgID int64, req *ComplianceReportRequest, page, pageSize int) (*ComplianceLeaveResponse, error) {
	deptIDs := req.ParseDeptIDs()
	employees, err := s.repo.ListEmployeesByOrgWithDept(orgID, deptIDs)
	if err != nil {
		return nil, fmt.Errorf("获取员工列表失败: %w", err)
	}
	if len(employees) == 0 {
		return &ComplianceLeaveResponse{YearMonth: req.YearMonth, List: []LeaveItem{}, Total: 0}, nil
	}

	empIDs := make([]int64, len(employees))
	for i, e := range employees {
		empIDs[i] = e.EmployeeID
	}

	approvals, err := s.repo.ListApprovalsByMonth(orgID, empIDs, req.YearMonth)
	if err != nil {
		return nil, fmt.Errorf("获取请假审批失败: %w", err)
	}

	// Get year from year_month for annual leave quota lookup
	parsed, _ := time.Parse("2006-01", req.YearMonth)
	year := parsed.Year()
	quotas, _ := s.repo.GetAnnualLeaveQuotas(orgID, empIDs, year)

	// Aggregate leave per employee
	type leaveAgg struct {
		annualUsed, sickDays, personalDays float64
	}
	agg := make(map[int64]*leaveAgg)
	for _, empID := range empIDs {
		agg[empID] = &leaveAgg{}
	}
	for _, a := range approvals {
		if a.Status != ApprovalStatusApproved {
			continue
		}
		days := roundHalf(a.Duration / 8.0)
		switch a.LeaveType {
		case "annual_leave", "PTO":
			agg[a.EmployeeID].annualUsed += days
		case "sick_leave":
			agg[a.EmployeeID].sickDays += days
		case "personal_leave":
			agg[a.EmployeeID].personalDays += days
		}
	}

	list := make([]LeaveItem, 0, len(employees))
	var totalAnnualUsed, totalSick, totalPersonal float64
	var quotaCount int
	for _, emp := range employees {
		a := agg[emp.EmployeeID]
		if a == nil {
			a = &leaveAgg{}
		}
		quota := quotas[emp.EmployeeID]
		if quota > 0 {
			quotaCount++
		}
		list = append(list, LeaveItem{
			EmployeeID:     emp.EmployeeID,
			EmployeeName:   emp.EmployeeName,
			DepartmentName: emp.DepartmentName,
			AnnualQuota:    quota,
			AnnualUsed:     roundHalf(a.annualUsed),
			AnnualLeft:     roundHalf(quota - a.annualUsed),
			SickDays:       roundHalf(a.sickDays),
			PersonalDays:   roundHalf(a.personalDays),
		})
		totalAnnualUsed += a.annualUsed
		totalSick += a.sickDays
		totalPersonal += a.personalDays
	}

	total := int64(len(list))
	start := (page - 1) * pageSize
	if start > int(total) {
		start = int(total)
	}
	end := start + pageSize
	if end > int(total) {
		end = int(total)
	}
	paged := list[start:end]

	return &ComplianceLeaveResponse{
		YearMonth: req.YearMonth,
		Stats: ComplianceLeaveStats{
			AnnualQuotaEmployeeCount: quotaCount,
			TotalAnnualUsed:          roundHalf(totalAnnualUsed),
			TotalSickDays:             roundHalf(totalSick),
			TotalPersonalDays:        roundHalf(totalPersonal),
		},
		List:     paged,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetComplianceAnomaly returns attendance anomaly statistics per employee
// Anomaly highlight: late_count > 3 OR absent_days > 1 (D-12-10)
// Data from AttendanceMonthly table (late_count, early_leave_count, absent_days per employee)
func (s *AttendanceService) GetComplianceAnomaly(ctx context.Context, orgID int64, req *ComplianceReportRequest, page, pageSize int) (*ComplianceAnomalyResponse, error) {
	deptIDs := req.ParseDeptIDs()
	employees, err := s.repo.ListEmployeesByOrgWithDept(orgID, deptIDs)
	if err != nil {
		return nil, fmt.Errorf("获取员工列表失败: %w", err)
	}
	if len(employees) == 0 {
		return &ComplianceAnomalyResponse{YearMonth: req.YearMonth, List: []AnomalyItem{}, Total: 0}, nil
	}

	empIDs := make([]int64, len(employees))
	for i, e := range employees {
		empIDs[i] = e.EmployeeID
	}

	monthlyRecords, err := s.repo.ListMonthlyAttendanceForCompliance(orgID, empIDs, req.YearMonth)
	if err != nil {
		return nil, fmt.Errorf("获取月度考勤数据失败: %w", err)
	}

	monthlyMap := make(map[int64]AttendanceMonthly)
	for _, m := range monthlyRecords {
		monthlyMap[m.EmployeeID] = m
	}

	list := make([]AnomalyItem, 0, len(employees))
	var totalLate int
	var totalAbsent float64
	var anomalyCount int
	for _, emp := range employees {
		m := monthlyMap[emp.EmployeeID]
		lateCount := int(m.LateCount)
		earlyCount := int(m.EarlyLeaveCount)
		absentDays := m.AbsentDays
		anomalyTotal := lateCount + earlyCount
		isAnomaly := lateCount > 3 || absentDays > 1
		if isAnomaly {
			anomalyCount++
		}
		totalLate += lateCount
		totalAbsent += absentDays
		list = append(list, AnomalyItem{
			EmployeeID:      emp.EmployeeID,
			EmployeeName:    emp.EmployeeName,
			DepartmentName:  emp.DepartmentName,
			LateCount:       lateCount,
			EarlyLeaveCount: earlyCount,
			AbsentDays:      absentDays,
			AnomalyCount:    anomalyTotal,
			IsAnomaly:       isAnomaly,
		})
	}

	total := int64(len(list))
	start := (page - 1) * pageSize
	if start > int(total) {
		start = int(total)
	}
	end := start + pageSize
	if end > int(total) {
		end = int(total)
	}
	paged := list[start:end]

	return &ComplianceAnomalyResponse{
		YearMonth: req.YearMonth,
		Stats: ComplianceAnomalyStats{
			AnomalyEmployeeCount: anomalyCount,
			TotalLateCount:       totalLate,
			TotalAbsentDays:      totalAbsent,
		},
		List:     paged,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetComplianceMonthly returns monthly compliance summary per employee
// Combines: required/actual/leave/overtime from AttendanceMonthly + leave from approvals
func (s *AttendanceService) GetComplianceMonthly(ctx context.Context, orgID int64, req *ComplianceReportRequest, page, pageSize int) (*ComplianceMonthlyResponse, error) {
	deptIDs := req.ParseDeptIDs()
	employees, err := s.repo.ListEmployeesByOrgWithDept(orgID, deptIDs)
	if err != nil {
		return nil, fmt.Errorf("获取员工列表失败: %w", err)
	}
	if len(employees) == 0 {
		return &ComplianceMonthlyResponse{YearMonth: req.YearMonth, List: []MonthlyComplianceItem{}, Total: 0}, nil
	}

	empIDs := make([]int64, len(employees))
	for i, e := range employees {
		empIDs[i] = e.EmployeeID
	}

	// Get monthly attendance data
	monthlyRecords, err := s.repo.ListMonthlyAttendanceForCompliance(orgID, empIDs, req.YearMonth)
	if err != nil {
		return nil, fmt.Errorf("获取月度考勤数据失败: %w", err)
	}
	monthlyMap := make(map[int64]AttendanceMonthly)
	for _, m := range monthlyRecords {
		monthlyMap[m.EmployeeID] = m
	}

	// Get leave data from approvals for annual/sick/personal leave breakdown
	approvals, _ := s.repo.ListApprovalsByMonth(orgID, empIDs, req.YearMonth)
	type leaveBreakdown struct {
		annual, sick, personal float64
	}
	leaveMap := make(map[int64]*leaveBreakdown)
	for _, empID := range empIDs {
		leaveMap[empID] = &leaveBreakdown{}
	}
	for _, a := range approvals {
		if a.Status != ApprovalStatusApproved {
			continue
		}
		days := roundHalf(a.Duration / 8.0)
		switch a.LeaveType {
		case "annual_leave", "PTO":
			leaveMap[a.EmployeeID].annual += days
		case "sick_leave":
			leaveMap[a.EmployeeID].sick += days
		case "personal_leave":
			leaveMap[a.EmployeeID].personal += days
		}
	}

	list := make([]MonthlyComplianceItem, 0, len(employees))
	var totalReq, totalAct, totalOT float64
	var totalAbsent float64
	var totalAnomaly int
	for _, emp := range employees {
		m := monthlyMap[emp.EmployeeID]
		lb := leaveMap[emp.EmployeeID]
		if lb == nil {
			lb = &leaveBreakdown{}
		}
		lateCount := int(m.LateCount)
		absentDays := m.AbsentDays
		isAnomaly := lateCount > 3 || absentDays > 1
		if isAnomaly {
			totalAnomaly++
		}
		list = append(list, MonthlyComplianceItem{
			EmployeeID:        emp.EmployeeID,
			EmployeeName:      emp.EmployeeName,
			DepartmentName:    emp.DepartmentName,
			RequiredDays:      m.RequiredDays,
			ActualDays:        m.ActualDays,
			LateCount:         lateCount,
			EarlyLeaveCount:   int(m.EarlyLeaveCount),
			AbsentDays:        absentDays,
			OvertimeHours:     roundHalf(m.OvertimeHours),
			AnnualLeaveDays:   roundHalf(lb.annual),
			SickLeaveDays:     roundHalf(lb.sick),
			PersonalLeaveDays: roundHalf(lb.personal),
			IsAnomaly:         isAnomaly,
		})
		totalReq += m.RequiredDays
		totalAct += m.ActualDays
		totalOT += m.OvertimeHours
		totalAbsent += absentDays
	}

	total := int64(len(list))
	start := (page - 1) * pageSize
	if start > int(total) {
		start = int(total)
	}
	end := start + pageSize
	if end > int(total) {
		end = int(total)
	}
	paged := list[start:end]

	return &ComplianceMonthlyResponse{
		YearMonth: req.YearMonth,
		Stats: ComplianceMonthlyStats{
			TotalRequiredDays:  roundHalf(totalReq),
			TotalActualDays:    roundHalf(totalAct),
			TotalOvertimeHours: roundHalf(totalOT),
			TotalAbsentDays:    roundHalf(totalAbsent),
			TotalAnomalyCount:  totalAnomaly,
		},
		List:     paged,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ExportComplianceMonthlyExcel generates a monthly compliance Excel export
// Headers: 姓名/部门/应出勤/实际出勤/迟到/早退/缺勤/加班/年假/病假/事假/异常标记
func (s *AttendanceService) ExportComplianceMonthlyExcel(ctx context.Context, orgID int64, req *ComplianceReportRequest) ([]byte, string, error) {
	report, err := s.GetComplianceMonthly(ctx, orgID, req, 1, 5000)
	if err != nil {
		return nil, "", fmt.Errorf("获取月度汇总数据失败: %w", err)
	}

	f := excelize.NewFile()
	sheet := "月度考勤汇总"
	index, _ := f.NewSheet(sheet)
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	// Headers
	headers := []string{"姓名", "部门", "应出勤(天)", "实际出勤(天)", "迟到(次)", "早退(次)", "缺勤(天)", "加班(小时)", "年假(天)", "病假(天)", "事假(天)", "异常"}
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#EDE9FE"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	anomalyStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#EF4444"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#FEE2E2"}, Pattern: 1},
	})
	normalStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	for rowIdx, item := range report.List {
		row := rowIdx + 2
		vals := []interface{}{
			item.EmployeeName, item.DepartmentName,
			item.RequiredDays, item.ActualDays,
			item.LateCount, item.EarlyLeaveCount, item.AbsentDays,
			item.OvertimeHours,
			item.AnnualLeaveDays, item.SickLeaveDays, item.PersonalLeaveDays,
			"是", // anomaly flag
		}
		if !item.IsAnomaly {
			vals[11] = "否"
		}
		for colIdx, val := range vals {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, row)
			f.SetCellValue(sheet, cell, val)
			if item.IsAnomaly {
				f.SetCellStyle(sheet, cell, cell, anomalyStyle)
			} else if colIdx >= 4 { // center numeric columns
				f.SetCellStyle(sheet, cell, cell, normalStyle)
			}
		}
	}

	// Summary row
	summaryRow := len(report.List) + 2
	f.SetCellValue(sheet, "A"+strconv.Itoa(summaryRow), "合计")
	f.SetCellValue(sheet, "C"+strconv.Itoa(summaryRow), report.Stats.TotalRequiredDays)
	f.SetCellValue(sheet, "D"+strconv.Itoa(summaryRow), report.Stats.TotalActualDays)
	f.SetCellValue(sheet, "H"+strconv.Itoa(summaryRow), roundHalf(report.Stats.TotalOvertimeHours))
	f.SetCellValue(sheet, "G"+strconv.Itoa(summaryRow), report.Stats.TotalAbsentDays)
	f.SetCellValue(sheet, "L"+strconv.Itoa(summaryRow), report.Stats.TotalAnomalyCount)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", fmt.Errorf("生成 Excel 失败: %w", err)
	}
	filename := fmt.Sprintf("月度考勤汇总_%s.xlsx", req.YearMonth)
	return buf.Bytes(), filename, nil
}
