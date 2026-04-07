package tax

import (
	"fmt"
	"math"
)

// TaxCalculator 个税计算接口（Phase 5 调用, per D-08）
type TaxCalculator interface {
	CalculateTax(orgID, employeeID int64, year, month int, grossIncome float64) (*TaxResult, error)
}

// roundTo2 保留两位小数
func roundTo2(val float64) float64 {
	return math.Round(val*100) / 100
}

// calculateCumulativeTax 累计预扣预缴计算引擎（纯函数）
// 核心算法 per research section 2, per D-13:
// 1. 累计收入 = SUM(records的gross_income) + 当月grossIncome
// 2. 累计基本减除 = (len(records)+1) * 5000.0 (per D-03)
// 3. 累计社保扣款 = SUM(records的si_deduction) + 当月siDeduction
// 4. 累计专项附加扣除 = SUM(records的special_deduction) + 当月specialDeduction
// 5. 累计应纳税所得额 = (1) - (2) - (3) - (4)
// 6. 如果 累计应纳税所得额 <= 0, 当月个税 = 0
// 7. 查税率表找到对应区间 (per D-15)
// 8. 累计应扣税额 = 累计应纳税所得额 * 税率 - 速算扣除数
// 9. 累计已扣税额 = SUM(records的monthly_tax)
// 10. 当月应扣税额 = 累计应扣税额 - 累计已扣税额
// 11. 如果 当月应扣税额 < 0, 设为 0 (不退税, per research pitfall #1)
func calculateCumulativeTax(
	brackets []TaxBracket,
	records []TaxRecord,
	grossIncome float64,
	basicDeduction float64,
	siDeduction float64,
	specialDeduction float64,
) *TaxResult {
	// 1. 累计收入
	var cumulativeIncome float64
	for _, r := range records {
		cumulativeIncome += r.GrossIncome
	}
	cumulativeIncome = roundTo2(cumulativeIncome + grossIncome)

	// 2. 累计基本减除
	cumulativeBasicDeduction := roundTo2(float64(len(records)+1) * basicDeduction)

	// 3. 累计社保扣款
	var cumulativeSIDeduction float64
	for _, r := range records {
		cumulativeSIDeduction += r.SIDeduction
	}
	cumulativeSIDeduction = roundTo2(cumulativeSIDeduction + siDeduction)

	// 4. 累计专项附加扣除
	var cumulativeSpecialDeduction float64
	for _, r := range records {
		cumulativeSpecialDeduction += r.SpecialDeduction
	}
	cumulativeSpecialDeduction = roundTo2(cumulativeSpecialDeduction + specialDeduction)

	// 5. 累计应纳税所得额
	totalDeduction := roundTo2(basicDeduction + siDeduction + specialDeduction)
	cumulativeTaxableIncome := roundTo2(cumulativeIncome - cumulativeBasicDeduction - cumulativeSIDeduction - cumulativeSpecialDeduction)

	// 6. 如果 累计应纳税所得额 <= 0, 当月个税 = 0
	if cumulativeTaxableIncome <= 0 {
		cumulativeTaxableIncome = 0
		return &TaxResult{
			MonthlyTax:                0,
			CumulativeIncome:          cumulativeIncome,
			CumulativeDeduction:       roundTo2(cumulativeBasicDeduction + cumulativeSIDeduction + cumulativeSpecialDeduction),
			CumulativeTaxableIncome:   0,
			TaxRate:                   0,
			QuickDeduction:            0,
			CumulativeTax:             0,
			GrossIncome:               roundTo2(grossIncome),
			BasicDeduction:            roundTo2(basicDeduction),
			SIDeduction:               roundTo2(siDeduction),
			SpecialDeduction:          roundTo2(specialDeduction),
			TotalDeduction:            totalDeduction,
		}
	}

	// 7. 查税率表找到对应区间
	bracket := FindTaxBracketForAmount(brackets, cumulativeTaxableIncome)
	if bracket == nil {
		return &TaxResult{
			MonthlyTax:                0,
			CumulativeIncome:          cumulativeIncome,
			CumulativeDeduction:       roundTo2(cumulativeBasicDeduction + cumulativeSIDeduction + cumulativeSpecialDeduction),
			CumulativeTaxableIncome:   cumulativeTaxableIncome,
			TaxRate:                   0,
			QuickDeduction:            0,
			CumulativeTax:             0,
			GrossIncome:               roundTo2(grossIncome),
			BasicDeduction:            roundTo2(basicDeduction),
			SIDeduction:               roundTo2(siDeduction),
			SpecialDeduction:          roundTo2(specialDeduction),
			TotalDeduction:            totalDeduction,
		}
	}

	// 8. 累计应扣税额 = 累计应纳税所得额 * 税率 - 速算扣除数
	cumulativeTax := roundTo2(cumulativeTaxableIncome*bracket.Rate - bracket.QuickDeduction)
	if cumulativeTax < 0 {
		cumulativeTax = 0
	}

	// 9. 累计已扣税额
	var previousCumulativeTax float64
	for _, r := range records {
		previousCumulativeTax += r.MonthlyTax
	}
	previousCumulativeTax = roundTo2(previousCumulativeTax)

	// 10. 当月应扣税额 = 累计应扣税额 - 累计已扣税额
	monthlyTax := roundTo2(cumulativeTax - previousCumulativeTax)

	// 11. 如果 当月应扣税额 < 0, 设为 0 (不退税, per research pitfall #1)
	if monthlyTax < 0 {
		monthlyTax = 0
	}

	return &TaxResult{
		MonthlyTax:                monthlyTax,
		CumulativeIncome:          cumulativeIncome,
		CumulativeDeduction:       roundTo2(cumulativeBasicDeduction + cumulativeSIDeduction + cumulativeSpecialDeduction),
		CumulativeTaxableIncome:   cumulativeTaxableIncome,
		TaxRate:                   bracket.Rate,
		QuickDeduction:            bracket.QuickDeduction,
		CumulativeTax:             cumulativeTax,
		GrossIncome:               roundTo2(grossIncome),
		BasicDeduction:            roundTo2(basicDeduction),
		SIDeduction:               roundTo2(siDeduction),
		SpecialDeduction:          roundTo2(specialDeduction),
		TotalDeduction:            totalDeduction,
	}
}

// FilterRecordsBeforeMonth 过滤出指定月份之前的记录（用于年中入职场景）
// 例如 3月入职，只有 3月之前的记录（即无前序记录），从入职月起计算
func FilterRecordsBeforeMonth(records []TaxRecord, month int) []TaxRecord {
	var filtered []TaxRecord
	for _, r := range records {
		if r.Month < month {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// GetTaxRecordsForCumulative 获取用于累计计算的记录列表
// 过滤掉指定月份之后的记录（防止重复计算）
func GetTaxRecordsForCumulative(records []TaxRecord, currentMonth int) []TaxRecord {
	var filtered []TaxRecord
	for _, r := range records {
		if r.Month < currentMonth {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// ValidateTaxCalculationParams 校验计算参数
func ValidateTaxCalculationParams(year, month int, grossIncome float64) error {
	if year < 2000 || year > 2100 {
		return fmt.Errorf("无效的年份: %d", year)
	}
	if month < 1 || month > 12 {
		return fmt.Errorf("无效的月份: %d", month)
	}
	if grossIncome < 0 {
		return fmt.Errorf("收入不能为负数: %.2f", grossIncome)
	}
	return nil
}
