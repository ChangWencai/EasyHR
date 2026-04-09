package salary

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wencai/easyhr/internal/tax"
)

// TestCalculateDailyWage 验证日薪计算
func TestCalculateDailyWage(t *testing.T) {
	result := calculateDailyWage(10000)
	assert.Equal(t, 459.77, result)
}

// TestCalculateDailyWage_Zero 验证零月薪
func TestCalculateDailyWage_Zero(t *testing.T) {
	result := calculateDailyWage(0)
	assert.Equal(t, 0.0, result)
}

// TestCalculateLeaveDeduction 验证请假扣款
func TestCalculateLeaveDeduction(t *testing.T) {
	dailyWage := calculateDailyWage(10000) // 459.77
	result := calculateLeaveDeduction(dailyWage, 2)
	assert.Equal(t, 919.54, result)
}

// TestCalculatePayroll_Basic 基本场景
func TestCalculatePayroll_Basic(t *testing.T) {
	items := []PayrollItemInput{
		{ItemName: "基本工资", ItemType: "income", Amount: 8000},
	}
	result := calculatePayroll(items, 500, &tax.TaxResult{MonthlyTax: 45})
	assert.Equal(t, 8000.0, result.GrossIncome)
	assert.Equal(t, 500.0, result.SIDeduction)
	assert.Equal(t, 45.0, result.Tax)
	assert.Equal(t, 0.0, result.TotalDeductions)
	assert.Equal(t, 7455.0, result.NetIncome)
}

// TestCalculatePayroll_WithMultipleIncome 含绩效/补贴
func TestCalculatePayroll_WithMultipleIncome(t *testing.T) {
	items := []PayrollItemInput{
		{ItemName: "基本工资", ItemType: "income", Amount: 8000},
		{ItemName: "绩效工资", ItemType: "income", Amount: 2000},
		{ItemName: "岗位补贴", ItemType: "income", Amount: 500},
		{ItemName: "其他扣款", ItemType: "deduction", Amount: 100},
	}
	result := calculatePayroll(items, 500, &tax.TaxResult{MonthlyTax: 150})
	assert.Equal(t, 10500.0, result.GrossIncome)
	assert.Equal(t, 100.0, result.TotalDeductions)
	assert.Equal(t, 9750.0, result.NetIncome) // 10500 - 500 - 150 - 100
}

// TestCalculatePayroll_WithLeaveDeduction 含请假扣款
func TestCalculatePayroll_WithLeaveDeduction(t *testing.T) {
	items := []PayrollItemInput{
		{ItemName: "基本工资", ItemType: "income", Amount: 10000},
		{ItemName: "事假扣款", ItemType: "deduction", Amount: 919.54},
	}
	result := calculatePayroll(items, 500, &tax.TaxResult{MonthlyTax: 0})
	assert.Equal(t, 10000.0, result.GrossIncome)
	assert.Equal(t, 919.54, result.TotalDeductions)
	// 10000 - 500 - 0 - 919.54 = 8580.46
	assert.Equal(t, 8580.46, result.NetIncome)
}

// TestCalculatePayroll_ZeroIncome 零收入边界
func TestCalculatePayroll_ZeroIncome(t *testing.T) {
	items := []PayrollItemInput{}
	result := calculatePayroll(items, 0, nil)
	assert.Equal(t, 0.0, result.GrossIncome)
	assert.Equal(t, 0.0, result.SIDeduction)
	assert.Equal(t, 0.0, result.Tax)
	assert.Equal(t, 0.0, result.TotalDeductions)
	assert.Equal(t, 0.0, result.NetIncome)
}

// TestCalculatePayroll_NilTaxResult 个税结果为 nil
func TestCalculatePayroll_NilTaxResult(t *testing.T) {
	items := []PayrollItemInput{
		{ItemName: "基本工资", ItemType: "income", Amount: 5000},
	}
	result := calculatePayroll(items, 0, nil)
	assert.Equal(t, 5000.0, result.NetIncome)
}

// TestRoundTo2 精度验证
func TestRoundTo2(t *testing.T) {
	assert.Equal(t, 459.77, roundTo2Salary(10000.0/21.75))
	assert.Equal(t, 0.0, roundTo2Salary(0.0))
	assert.Equal(t, 100.12, roundTo2Salary(100.115))
	assert.Equal(t, 100.13, roundTo2Salary(100.125))
}

// TestCheckAbnormalPayments_Flagged 偏差>30% 标记异常
func TestCheckAbnormalPayments_Flagged(t *testing.T) {
	current := []AbnormalCheckInput{
		{EmployeeID: 1, EmployeeName: "张三", NetIncome: 5000},
		{EmployeeID: 2, EmployeeName: "李四", NetIncome: 8000},
	}
	previous := map[int64]float64{
		1: 8000, // |5000-8000|/8000 = 37.5% > 30%
		2: 8000, // 0% < 30%
	}

	abnormal := checkAbnormalPayments(current, previous)
	assert.Len(t, abnormal, 1)
	assert.Equal(t, int64(1), abnormal[0].EmployeeID)
	assert.Equal(t, "张三", abnormal[0].EmployeeName)
	assert.Equal(t, 5000.0, abnormal[0].CurrentNet)
	assert.Equal(t, 8000.0, abnormal[0].PreviousNet)
	assert.Equal(t, 37.5, abnormal[0].DeviationRate)
}

// TestCheckAbnormalPayments_WithinThreshold 偏差<30% 不标记
func TestCheckAbnormalPayments_WithinThreshold(t *testing.T) {
	current := []AbnormalCheckInput{
		{EmployeeID: 1, EmployeeName: "张三", NetIncome: 6000},
	}
	previous := map[int64]float64{
		1: 8000, // 25% < 30%
	}

	abnormal := checkAbnormalPayments(current, previous)
	assert.Len(t, abnormal, 0)
}

// TestCheckAbnormalPayments_NoPrevious 无上月数据不检查
func TestCheckAbnormalPayments_NoPrevious(t *testing.T) {
	current := []AbnormalCheckInput{
		{EmployeeID: 1, EmployeeName: "新员工", NetIncome: 6000},
	}
	previous := map[int64]float64{}

	abnormal := checkAbnormalPayments(current, previous)
	assert.Len(t, abnormal, 0)
}

// TestCheckAbnormalPayments_ZeroPrevious 上月为零不检查
func TestCheckAbnormalPayments_ZeroPrevious(t *testing.T) {
	current := []AbnormalCheckInput{
		{EmployeeID: 1, EmployeeName: "张三", NetIncome: 6000},
	}
	previous := map[int64]float64{
		1: 0,
	}

	abnormal := checkAbnormalPayments(current, previous)
	assert.Len(t, abnormal, 0)
}

// TestCalculateDailyWage_Precision 验证各种月薪的日薪精度
func TestCalculateDailyWage_Precision(t *testing.T) {
	tests := []struct {
		salary   float64
		expected float64
	}{
		{5000, 229.89},
		{8000, 367.82},
		{10000, 459.77},
		{15000, 689.66},
		{20000, 919.54},
		{30000, 1379.31},
	}
	for _, tt := range tests {
		result := calculateDailyWage(tt.salary)
		assert.Equal(t, tt.expected, result, "月薪 %.0f 的日薪计算", tt.salary)
	}
}

// TestCalculatePayroll_WithSIAndTax 社保+个税联动
func TestCalculatePayroll_WithSIAndTax(t *testing.T) {
	items := []PayrollItemInput{
		{ItemName: "基本工资", ItemType: "income", Amount: 10000},
		{ItemName: "绩效工资", ItemType: "income", Amount: 2000},
	}
	result := calculatePayroll(items, 1050.8, &tax.TaxResult{MonthlyTax: 390})
	assert.Equal(t, 12000.0, result.GrossIncome)
	assert.Equal(t, 1050.8, result.SIDeduction)
	assert.Equal(t, 390.0, result.Tax)
	// 12000 - 1050.8 - 390 - 0 = 10559.2
	assert.Equal(t, 10559.2, result.NetIncome)
}
