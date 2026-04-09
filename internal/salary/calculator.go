package salary

import (
	"math"

	"github.com/wencai/easyhr/internal/tax"
)

// PayrollItemInput 核算输入项
type PayrollItemInput struct {
	ItemName string
	ItemType string // income/deduction
	Amount   float64
}

// PayrollResult 核算结果
type PayrollResult struct {
	GrossIncome     float64
	SIDeduction     float64
	Tax             float64
	TotalDeductions float64 // 扣款项合计（事假/病假/其他扣款）
	NetIncome       float64
	TaxResult       *tax.TaxResult // 完整个税结果（用于审计）
}

// AbnormalCheck 异常发放检查结果
type AbnormalCheck struct {
	EmployeeID    int64
	EmployeeName  string
	CurrentNet    float64
	PreviousNet   float64
	DeviationRate float64 // 偏差百分比
}

// calculateDailyWage 日薪计算（per D-14: 月基本工资 / 21.75）
func calculateDailyWage(monthlyBase float64) float64 {
	return roundTo2Salary(monthlyBase / 21.75)
}

// calculateLeaveDeduction 请假扣款计算
func calculateLeaveDeduction(dailyWage float64, days float64) float64 {
	return roundTo2Salary(dailyWage * days)
}

// calculatePayroll 工资核算纯函数（per D-06）
// 输入：薪资项列表、社保个人扣款、个税结果
// 输出：完整核算结果
func calculatePayroll(items []PayrollItemInput, siDeduction float64, taxResult *tax.TaxResult) *PayrollResult {
	var grossIncome float64
	var totalDeductions float64

	for _, item := range items {
		if item.ItemType == "income" {
			grossIncome = roundTo2Salary(grossIncome + item.Amount)
		} else {
			totalDeductions = roundTo2Salary(totalDeductions + item.Amount)
		}
	}

	monthlyTax := float64(0)
	if taxResult != nil {
		monthlyTax = taxResult.MonthlyTax
	}

	netIncome := roundTo2Salary(grossIncome - siDeduction - monthlyTax - totalDeductions)

	return &PayrollResult{
		GrossIncome:     grossIncome,
		SIDeduction:     siDeduction,
		Tax:             monthlyTax,
		TotalDeductions: totalDeductions,
		NetIncome:       netIncome,
		TaxResult:       taxResult,
	}
}

// AbnormalCheckInput 异常检查输入
type AbnormalCheckInput struct {
	EmployeeID   int64
	EmployeeName string
	NetIncome    float64
}

// checkAbnormalPayments 异常发放检查（per D-17: 偏差>30% 标记为异常）
func checkAbnormalPayments(currentRecords []AbnormalCheckInput, previousRecords map[int64]float64) []AbnormalCheck {
	var abnormal []AbnormalCheck
	for _, r := range currentRecords {
		prevNet, exists := previousRecords[r.EmployeeID]
		if !exists || prevNet == 0 {
			continue // 无上月数据或上月为0，不检查
		}
		deviation := math.Abs(r.NetIncome-prevNet) / prevNet
		if deviation > 0.30 {
			abnormal = append(abnormal, AbnormalCheck{
				EmployeeID:    r.EmployeeID,
				EmployeeName:  r.EmployeeName,
				CurrentNet:    r.NetIncome,
				PreviousNet:   prevNet,
				DeviationRate: roundTo2Salary(deviation * 100),
			})
		}
	}
	return abnormal
}

// roundTo2Salary 保留两位小数
func roundTo2Salary(val float64) float64 {
	return math.Round(val*100) / 100
}
