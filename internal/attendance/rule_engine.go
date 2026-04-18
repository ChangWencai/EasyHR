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
