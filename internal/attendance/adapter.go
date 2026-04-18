package attendance

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// MonthlyAttendance 出勤月报数据（供 salary 模块使用）
type MonthlyAttendance struct {
	ActualDays           float64 // 实际出勤天数
	ShouldAttend         float64 // 应出勤天数
	OvertimeHours        float64 // 总加班时长（小时）
	OvertimeWeekdayHours float64 // 工作日加班时长
	OvertimeWeekendHours float64 // 双休日加班时长
	OvertimeHolidayHours float64 // 节假日加班时长
	PaidLeaveDays        float64 // 带薪假天数（年假/婚假/产假/陪产假/调休）
	LegalHolidayDays     float64 // 法定节假日天数（从 AttendanceRule.Holidays 读取）
	SickLeaveDays        float64 // 病假天数
}

// AttendanceProvider 考勤数据提供者接口（供 salary 模块调用）
type AttendanceProvider interface {
	GetMonthlyAttendance(orgID, employeeID int64, yearMonth string) (*MonthlyAttendance, error)
}

// attendanceProvider AttendanceProvider 的实现
type attendanceProvider struct {
	db *gorm.DB
}

// NewAttendanceProvider 创建考勤数据提供者
func NewAttendanceProvider(db *gorm.DB) AttendanceProvider {
	return &attendanceProvider{db: db}
}

// GetMonthlyAttendance 获取员工月度考勤数据
func (p *attendanceProvider) GetMonthlyAttendance(orgID, employeeID int64, yearMonth string) (*MonthlyAttendance, error) {
	repo := NewAttendanceRepository(p.db)

	// 获取考勤规则（用于判断节假日和工作日）
	rule, err := repo.GetRule(orgID)
	if err != nil {
		return nil, fmt.Errorf("获取考勤规则失败: %w", err)
	}
	ruleEngine := NewRuleEngine(rule)

	// 解析年月获取月份起止日期
	parsed, err := time.Parse("2006-01", yearMonth)
	if err != nil {
		return nil, fmt.Errorf("年月格式错误 %s: %w", yearMonth, err)
	}
	year, month, _ := parsed.Date()
	monthStart := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	monthEnd := time.Date(year, month+1, 0, 23, 59, 59, 0, time.UTC)

	// 查找 AttendanceMonthly 记录
	var monthly AttendanceMonthly
	err = p.db.Where("org_id = ? AND employee_id = ? AND year_month = ?",
		orgID, employeeID, yearMonth).First(&monthly).Error

	if err == gorm.ErrRecordNotFound {
		// 无记录时返回"全勤"默认值
		result := &MonthlyAttendance{
			ShouldAttend:         countWorkDays(ruleEngine, year, month),
			OvertimeWeekdayHours: 0,
			OvertimeWeekendHours: 0,
			OvertimeHolidayHours: 0,
		}
		result.ActualDays = result.ShouldAttend - countLegalHolidays(rule, yearMonth)
		result.LegalHolidayDays = countLegalHolidays(rule, yearMonth)
		return result, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询月度考勤失败: %w", err)
	}

	result := &MonthlyAttendance{
		ActualDays:   monthly.ActualDays,
		ShouldAttend: monthly.RequiredDays,
		OvertimeHours: monthly.OvertimeHours,
		LegalHolidayDays: countLegalHolidays(rule, yearMonth),
	}

	// 从 attendance_approvals 推导加班分档
	approvals, err := repo.ListApprovalsByEmployeeMonth(orgID, employeeID, yearMonth)
	if err != nil {
		return nil, fmt.Errorf("获取审批记录失败: %w", err)
	}

	for _, a := range approvals {
		if a.Status != ApprovalStatusApproved {
			continue
		}

		switch a.ApprovalType {
		case ApprovalTypeOvertime:
			classifyOvertime(ruleEngine, a, result, monthStart, monthEnd)
		case ApprovalTypeSickLeave:
			result.SickLeaveDays += a.Duration / 8.0
		case ApprovalTypePTO, ApprovalTypeAnnualLeave,
			ApprovalTypeMarriageLeave, ApprovalTypeMaternityLeave,
			ApprovalTypePaternityLeave:
			result.PaidLeaveDays += a.Duration / 8.0
		}
	}

	return result, nil
}

// classifyOvertime 根据 StartTime 所在日期将加班时长分类到工作日/双休日/节假日
func classifyOvertime(ruleEngine *RuleEngine, a Approval, result *MonthlyAttendance, monthStart, monthEnd time.Time) {
	// 加班开始时间必须在月份范围内
	startTime := a.StartTime
	if startTime.Before(monthStart) {
		startTime = monthStart
	}
	endTime := a.EndTime
	if endTime.After(monthEnd) {
		endTime = monthEnd
	}

	// 按天拆分加班时长并分类
	// 简化实现：以开始时间所在日判断类型
	hours := a.Duration
	date := startTime

	// 判断加班日的类型
	if ruleEngine.IsHoliday(date) {
		result.OvertimeHolidayHours += hours
	} else if isWeekend(date) {
		result.OvertimeWeekendHours += hours
	} else {
		result.OvertimeWeekdayHours += hours
	}
}

// countWorkDays 统计当月工作日天数（排除周末和法定节假日中的工作日）
func countWorkDays(ruleEngine *RuleEngine, year int, month time.Month) float64 {
	daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	count := 0
	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		if ruleEngine.IsWorkDay(date) {
			count++
		}
	}
	return float64(count)
}

// countLegalHolidays 统计当月法定节假日天数
func countLegalHolidays(rule *AttendanceRule, yearMonth string) float64 {
	if rule == nil || len(rule.Holidays) == 0 {
		return 0
	}
	var holidays []Holiday
	if err := json.Unmarshal(rule.Holidays, &holidays); err != nil {
		return 0
	}

	count := 0
	for _, h := range holidays {
		if len(h.Date) >= 7 && h.Date[:7] == yearMonth {
			count++
		}
	}
	return float64(count)
}

// isWeekend 判断是否为周末
func isWeekend(date time.Time) bool {
	weekday := date.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}
