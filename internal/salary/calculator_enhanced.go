package salary

import (
	"github.com/shopspring/decimal"
)

// CalculateBillingDays 计薪天数 = 实际出勤 + 法定节假日 + 带薪假天数
// per D-SAL-ATT-01
func CalculateBillingDays(actual, legalHoliday, paidLeave float64) float64 {
	return roundTo2Decimal(actual + legalHoliday + paidLeave)
}

// CalculateSalaryByBillingDays 基本工资按计薪天数计算
// salary_for_month = base_salary / should_attend * billing_days
func CalculateSalaryByBillingDays(baseSalary, shouldAttend, billingDays float64) float64 {
	if shouldAttend <= 0 {
		return roundTo2Decimal(baseSalary)
	}
	base := decimal.NewFromFloat(baseSalary)
	attend := decimal.NewFromFloat(shouldAttend)
	days := decimal.NewFromFloat(billingDays)
	result := base.Div(attend).Mul(days)
	rounded, _ := result.Round(2).Float64()
	return rounded
}

// CalculateSickLeaveWage 病假工资 = 基本工资 * 病假系数
// per SAL-14 + D-SAL-ATT-02
// Returns the sick leave wage amount (positive), caller should apply as deduction or income
func CalculateSickLeaveWage(baseSalary, sickCoefficient float64) float64 {
	base := decimal.NewFromFloat(baseSalary)
	coeff := decimal.NewFromFloat(sickCoefficient)
	result := base.Mul(coeff)
	rounded, _ := result.Round(2).Float64()
	return rounded
}

// CalculateSickLeaveDeduction 病假扣款 = 日工资 * 病假天数 * (1 - 病假系数)
// 即少发的部分：正常日工资 - 病假日工资
func CalculateSickLeaveDeduction(baseSalary, shouldAttend, sickLeaveDays, sickCoefficient float64) float64 {
	if shouldAttend <= 0 {
		return 0
	}
	base := decimal.NewFromFloat(baseSalary)
	attend := decimal.NewFromFloat(shouldAttend)
	days := decimal.NewFromFloat(sickLeaveDays)
	coeff := decimal.NewFromFloat(sickCoefficient)

	dailyWage := base.Div(attend)
	// 正常工资 = dailyWage * days
	// 病假工资 = dailyWage * days * coefficient
	// 扣款 = dailyWage * days * (1 - coefficient)
	deduction := dailyWage.Mul(days).Mul(decimal.NewFromInt(1).Sub(coeff))
	result := deduction.Round(2)
	rounded, _ := result.Float64()
	return rounded
}

// CalculateOvertimePay 加班费 = base_salary/should_attend/8h * overtime_hours * rate
// per SAL-15 + D-SAL-ATT-03
// weekday rate=1.5, weekend rate=2.0, holiday rate=3.0
func CalculateOvertimePay(baseSalary, shouldAttend float64, weekdayHours, weekendHours, holidayHours float64) float64 {
	if shouldAttend <= 0 {
		return 0
	}
	base := decimal.NewFromFloat(baseSalary)
	attend := decimal.NewFromFloat(shouldAttend)
	hoursPerDay := decimal.NewFromInt(8)

	hourlyRate := base.Div(attend).Div(hoursPerDay)

	weekdayPay := hourlyRate.Mul(decimal.NewFromFloat(weekdayHours)).Mul(decimal.NewFromFloat(1.5))
	weekendPay := hourlyRate.Mul(decimal.NewFromFloat(weekendHours)).Mul(decimal.NewFromInt(2))
	holidayPay := hourlyRate.Mul(decimal.NewFromFloat(holidayHours)).Mul(decimal.NewFromInt(3))

	total := weekdayPay.Add(weekendPay).Add(holidayPay)
	result := total.Round(2)
	rounded, _ := result.Float64()
	return rounded
}

// roundTo2Decimal rounds to 2 decimal places using shopspring/decimal
func roundTo2Decimal(val float64) float64 {
	d := decimal.NewFromFloat(val)
	rounded, _ := d.Round(2).Float64()
	return rounded
}
