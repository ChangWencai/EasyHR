package attendance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
