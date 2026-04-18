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
