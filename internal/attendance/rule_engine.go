package attendance

import (
	"encoding/json"
	"time"
)

// RuleEngine 打卡规则引擎
type RuleEngine struct {
	rule *AttendanceRule
}

// NewRuleEngine 创建规则引擎
func NewRuleEngine(rule *AttendanceRule) *RuleEngine {
	return &RuleEngine{rule: rule}
}

// IsWorkDay 判断指定日期是否为工作日（周一=1, 周日=7）
func (e *RuleEngine) IsWorkDay(date time.Time) bool {
	if e.rule == nil {
		return false
	}
	if e.rule.Mode != "fixed" && e.rule.Mode != "free" {
		return true // 按排班模式不适用固定工作日判断
	}
	if e.IsHoliday(date) {
		return false
	}
	var workDays []int
	if err := json.Unmarshal([]byte(e.rule.WorkDays), &workDays); err != nil || len(workDays) == 0 {
		return false
	}
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7 // 周日转为7
	}
	for _, d := range workDays {
		if d == weekday {
			return true
		}
	}
	return false
}

// IsHoliday 判断指定日期是否为节假日
func (e *RuleEngine) IsHoliday(date time.Time) bool {
	if e.rule == nil || len(e.rule.Holidays) == 0 {
		return false
	}
	var holidays []Holiday
	if err := json.Unmarshal(e.rule.Holidays, &holidays); err != nil {
		return false
	}
	dateStr := date.Format("2006-01-02")
	for _, h := range holidays {
		if h.Date == dateStr {
			return true
		}
	}
	return false
}

// GetExpectedClockTimes 返回指定日期的期望打卡时间
// 适用于 fixed 和 free 模式
func (e *RuleEngine) GetExpectedClockTimes(date time.Time) (workStart, workEnd string, isRestDay bool) {
	if e.rule == nil {
		return "", "", true
	}
	isRestDay = !e.IsWorkDay(date)
	return e.rule.WorkStart, e.rule.WorkEnd, isRestDay
}

// ClassifyOvertimeCategory classifies the overtime approval time window into one of three categories per D-12-03:
// - "holiday": StartTime falls on a holiday (IsHoliday)
// - "weekend": StartTime falls on Saturday/Sunday (Weekday == 6 or 0)
// - "weekday": all other work days
func (e *RuleEngine) ClassifyOvertimeCategory(start, end time.Time) string {
	// Use the start date of the overtime window for classification
	startDate := start
	// For multi-day approvals, use the date of the first working day covered
	if start.Hour() == 0 && start.Minute() == 0 {
		// All-day approvals: classify by start date
		startDate = start
	}

	if e.IsHoliday(startDate) {
		return "holiday"
	}
	weekday := int(startDate.Weekday()) // 0=Sun, 1=Mon, ..., 6=Sat
	if weekday == 0 || weekday == 6 {
		return "weekend"
	}
	return "weekday"
}
