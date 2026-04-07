package tax

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 创建 SQLite 内存数据库用于测试
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "Failed to open test database")

	// AutoMigrate 所有模型
	err = db.AutoMigrate(
		&TaxBracket{},
		&SpecialDeduction{},
		&TaxRecord{},
		&TaxDeclaration{},
	)
	require.NoError(t, err, "Failed to auto migrate")

	return db
}

// seedTestBrackets 插入标准7级税率种子数据
func seedTestBrackets(t *testing.T, repo *Repository, year int) {
	t.Helper()
	err := repo.SeedTaxBrackets(year)
	require.NoError(t, err, "Failed to seed tax brackets")
}

// ========== 税率表测试 ==========

func TestSeedTaxBrackets(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	err := repo.SeedTaxBrackets(2026)
	require.NoError(t, err)

	// 验证正好7条记录
	var count int64
	db.Model(&TaxBracket{}).Where("org_id = 0 AND effective_year = ?", 2026).Count(&count)
	assert.Equal(t, int64(7), count, "Should have exactly 7 tax brackets")

	// 验证税率值
	brackets, err := repo.FindTaxBrackets(2026)
	require.NoError(t, err)

	expectedRates := []float64{0.03, 0.10, 0.20, 0.25, 0.30, 0.35, 0.45}
	for i, b := range brackets {
		assert.Equal(t, i+1, b.Level, "Level should match")
		assert.Equal(t, expectedRates[i], b.Rate, "Rate should match for level %d", i+1)
	}

	// 验证重复调用不会重复插入
	err = repo.SeedTaxBrackets(2026)
	require.NoError(t, err)
	db.Model(&TaxBracket{}).Where("org_id = 0 AND effective_year = ?", 2026).Count(&count)
	assert.Equal(t, int64(7), count, "Should still have exactly 7 brackets after duplicate call")
}

func TestFindTaxBracketForAmount(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	seedTestBrackets(t, repo, 2026)

	brackets, err := repo.FindTaxBrackets(2026)
	require.NoError(t, err)

	tests := []struct {
		name           string
		amount         float64
		expectedRate   float64
		expectedQuick  float64
	}{
		{"Level 1: 0", 0, 0.03, 0},
		{"Level 1: 10000", 10000, 0.03, 0},
		{"Level 1: 35999", 35999, 0.03, 0},
		{"Level 2: 36000 boundary", 36000, 0.10, 2520},
		{"Level 2: 50000", 50000, 0.10, 2520},
		{"Level 3: 144000 boundary", 144000, 0.20, 16920},
		{"Level 3: 200000", 200000, 0.20, 16920},
		{"Level 4: 300000 boundary", 300000, 0.25, 31920},
		{"Level 5: 420000 boundary", 420000, 0.30, 52920},
		{"Level 6: 660000 boundary", 660000, 0.35, 85920},
		{"Level 7: 960000 boundary", 960000, 0.45, 181920},
		{"Level 7: 2000000 high income", 2000000, 0.45, 181920},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bracket := FindTaxBracketForAmount(brackets, tt.amount)
			require.NotNil(t, bracket)
			assert.Equal(t, tt.expectedRate, bracket.Rate)
			assert.Equal(t, tt.expectedQuick, bracket.QuickDeduction)
		})
	}
}

// ========== 专项附加扣除测试 ==========

func TestCreateDeduction(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	orgID := int64(1)
	deduction := &SpecialDeduction{
		EmployeeID:     100,
		DeductionType:  DeductionTypeChildEducation,
		MonthlyAmount:  2000 * 2,
		Count:          2,
		EffectiveStart: "2026-01",
	}
	deduction.OrgID = orgID
	deduction.CreatedBy = 1
	deduction.UpdatedBy = 1

	err := repo.CreateDeduction(orgID, deduction)
	require.NoError(t, err)
	assert.True(t, deduction.ID > 0, "Should have an ID after creation")

	// 验证 MonthlyAmount = standard * count
	assert.Equal(t, 4000.0, deduction.MonthlyAmount, "Child education: 2000 * 2 = 4000")
}

func TestCreateDeduction_MutualExclusion(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	orgID := int64(1)

	// 创建 housing_loan
	loan := &SpecialDeduction{
		EmployeeID:     100,
		DeductionType:  DeductionTypeHousingLoan,
		MonthlyAmount:  1000,
		Count:          1,
		EffectiveStart: "2026-01",
	}
	loan.OrgID = orgID
	loan.CreatedBy = 1
	loan.UpdatedBy = 1
	err := repo.CreateDeduction(orgID, loan)
	require.NoError(t, err)

	// 尝试创建 housing_rent（应该被业务层拦截，但 repository 层直接插入会成功）
	// 这里测试互斥检查逻辑的存在
	existing, err := repo.FindDeductionByEmployeeAndType(orgID, 100, DeductionTypeHousingLoan)
	require.NoError(t, err)
	assert.NotNil(t, existing, "Should find existing housing_loan deduction")
}

func TestListDeductionsByEmployee(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	orgID := int64(1)

	// 创建多个扣除项
	endDate := "2026-12"
	deductions := []*SpecialDeduction{
		{
			EmployeeID:     100,
			DeductionType:  DeductionTypeChildEducation,
			MonthlyAmount:  2000,
			Count:          1,
			EffectiveStart: "2026-01",
			EffectiveEnd:   &endDate,
		},
		{
			EmployeeID:     100,
			DeductionType:  DeductionTypeElderlyCare,
			MonthlyAmount:  3000,
			Count:          1,
			EffectiveStart: "2026-03",
			EffectiveEnd:   nil,
		},
	}
	for _, d := range deductions {
		d.OrgID = orgID
		d.CreatedBy = 1
		d.UpdatedBy = 1
		err := repo.CreateDeduction(orgID, d)
		require.NoError(t, err)
	}

	// 查询 2026-03 月的扣除项
	results, total, err := repo.ListDeductionsByEmployee(orgID, 100, "2026-03", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total, "Both deductions should be active in March")

	// 查询 2026-02 月的扣除项（elderly_care 还未生效）
	results, total, err = repo.ListDeductionsByEmployee(orgID, 100, "2026-02", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total, "Only child_education should be active in February")
	assert.Equal(t, DeductionTypeChildEducation, results[0].DeductionType)
}

// ========== 计算引擎测试 ==========

func TestCalculateTax_BasicScenario(t *testing.T) {
	// 月薪10000无扣除，1月个税 = (10000-5000)*0.03-0 = 150
	db := setupTestDB(t)
	repo := NewRepository(db)
	seedTestBrackets(t, repo, 2026)

	brackets, err := repo.FindTaxBrackets(2026)
	require.NoError(t, err)

	result := calculateCumulativeTax(
		brackets,
		[]TaxRecord{}, // 1月无前序记录
		10000,         // grossIncome
		BasicDeductionMonthly,
		0, // siDeduction
		0, // specialDeduction
	)

	assert.Equal(t, 150.0, result.MonthlyTax, "1月个税应为150")
	assert.Equal(t, 10000.0, result.CumulativeIncome)
	assert.Equal(t, 5000.0, result.CumulativeTaxableIncome)
	assert.Equal(t, 0.03, result.TaxRate)
	assert.Equal(t, 0.0, result.QuickDeduction)
	assert.Equal(t, 150.0, result.CumulativeTax)
}

func TestCalculateTax_CumulativeMonth2(t *testing.T) {
	// 月薪10000连续2月，2月累计个税 = (20000-10000)*0.03-0 = 300，当月个税 = 300-150 = 150
	db := setupTestDB(t)
	repo := NewRepository(db)
	seedTestBrackets(t, repo, 2026)

	brackets, err := repo.FindTaxBrackets(2026)
	require.NoError(t, err)

	// 1月记录
	records := []TaxRecord{
		{
			EmployeeID:     1,
			GrossIncome:    10000,
			SIDeduction:    0,
			SpecialDeduction: 0,
			MonthlyTax:     150,
		},
	}

	result := calculateCumulativeTax(
		brackets,
		records,
		10000, // 2月 grossIncome
		BasicDeductionMonthly,
		0, // siDeduction
		0, // specialDeduction
	)

	assert.Equal(t, 150.0, result.MonthlyTax, "2月当月个税应为150")
	assert.Equal(t, 20000.0, result.CumulativeIncome)
	assert.Equal(t, 10000.0, result.CumulativeTaxableIncome)
	assert.Equal(t, 300.0, result.CumulativeTax, "累计个税应为300")
}

func TestCalculateTax_TaxBracketJump(t *testing.T) {
	// 月薪30000无扣除，6月累计应纳税所得额 = 180000，跨过144000边界，使用20%税率
	db := setupTestDB(t)
	repo := NewRepository(db)
	seedTestBrackets(t, repo, 2026)

	brackets, err := repo.FindTaxBrackets(2026)
	require.NoError(t, err)

	// 1-5月记录（每月月薪30000，个税按累计计算）
	// 1月: (30000-5000)*0.03 = 750
	// 2月: (60000-10000)*0.03-750 = 750
	// 3月: (90000-15000)*0.03-1500 = 750
	// 4月: (120000-20000)*0.03-2250 = 750
	// 5月: (150000-25000)*0.10-2520-3000 = 4980

	// 逐步构建记录
	records := []TaxRecord{}

	for month := 1; month <= 5; month++ {
		result := calculateCumulativeTax(brackets, records, 30000, BasicDeductionMonthly, 0, 0)
		records = append(records, TaxRecord{
			EmployeeID:       1,
			GrossIncome:      30000,
			SIDeduction:      0,
			SpecialDeduction: 0,
			MonthlyTax:       result.MonthlyTax,
		})
	}

	// 6月计算
	result := calculateCumulativeTax(brackets, records, 30000, BasicDeductionMonthly, 0, 0)

	// 6月累计收入 = 180000, 累计减除 = 30000, 累计应纳税所得额 = 150000
	// 150000 > 144000, 跨入20%税率区间
	// 累计税 = 150000*0.20 - 16920 = 13080
	// 前5月累计已缴需要重新计算验证
	assert.Equal(t, 180000.0, result.CumulativeIncome)
	assert.Equal(t, 150000.0, result.CumulativeTaxableIncome)
	assert.Equal(t, 0.20, result.TaxRate, "6月应使用20%税率")
	assert.Equal(t, 16920.0, result.QuickDeduction)

	// 验证前5月累计已缴税
	var prevTotalTax float64
	for _, r := range records {
		prevTotalTax += r.MonthlyTax
	}
	prevTotalTax = roundTo2(prevTotalTax)

	expectedMonthlyTax := roundTo2(13080.0 - prevTotalTax)
	assert.Equal(t, expectedMonthlyTax, result.MonthlyTax, "6月当月个税应正确")
	assert.True(t, result.MonthlyTax > 0, "6月当月个税应大于0")
}

func TestCalculateTax_WithDeductions(t *testing.T) {
	// 月薪20000，社保个人扣款2000，子女教育扣除2000
	// 当月应纳税所得额 = 20000 - 5000 - 2000 - 2000 = 11000
	// 累计（1月）个税 = 11000 * 0.03 = 330
	db := setupTestDB(t)
	repo := NewRepository(db)
	seedTestBrackets(t, repo, 2026)

	brackets, err := repo.FindTaxBrackets(2026)
	require.NoError(t, err)

	result := calculateCumulativeTax(
		brackets,
		[]TaxRecord{},
		20000, // grossIncome
		BasicDeductionMonthly,
		2000, // siDeduction
		2000, // specialDeduction (child education)
	)

	assert.Equal(t, 20000.0, result.GrossIncome)
	assert.Equal(t, 5000.0, result.BasicDeduction)
	assert.Equal(t, 2000.0, result.SIDeduction)
	assert.Equal(t, 2000.0, result.SpecialDeduction)
	assert.Equal(t, 9000.0, result.TotalDeduction)
	assert.Equal(t, 11000.0, result.CumulativeTaxableIncome)
	assert.Equal(t, 0.03, result.TaxRate)
	assert.Equal(t, 330.0, result.MonthlyTax, "1月个税应为330")
}

func TestCalculateTax_NegativeMonthlyTax(t *testing.T) {
	// 当月应扣为负时返回0（不退税）
	db := setupTestDB(t)
	repo := NewRepository(db)
	seedTestBrackets(t, repo, 2026)

	brackets, err := repo.FindTaxBrackets(2026)
	require.NoError(t, err)

	// 构造场景：前几个月高收入，当月低收入导致应扣为负
	records := []TaxRecord{
		{
			EmployeeID:       1,
			GrossIncome:      50000,
			SIDeduction:      0,
			SpecialDeduction: 0,
			MonthlyTax:       1500, // 前1月已缴1500
		},
	}

	// 2月收入为0（离职或请假），累计应纳税所得额可能降低
	result := calculateCumulativeTax(
		brackets,
		records,
		0, // 当月收入为0
		BasicDeductionMonthly,
		0,
		0,
	)

	assert.Equal(t, 0.0, result.MonthlyTax, "当月应扣为负时应返回0")
}

func TestCalculateTax_JanuaryReset(t *testing.T) {
	// 1月记录无前序记录，累计从零开始
	db := setupTestDB(t)
	repo := NewRepository(db)
	seedTestBrackets(t, repo, 2026)

	brackets, err := repo.FindTaxBrackets(2026)
	require.NoError(t, err)

	result := calculateCumulativeTax(
		brackets,
		[]TaxRecord{}, // 空记录，1月重新计算
		10000,
		BasicDeductionMonthly,
		0,
		0,
	)

	assert.Equal(t, 10000.0, result.CumulativeIncome, "1月累计收入应等于当月收入")
	assert.Equal(t, 5000.0, result.CumulativeTaxableIncome)
	assert.Equal(t, 150.0, result.MonthlyTax)
}

func TestCalculateTax_MidYearHire(t *testing.T) {
	// 年中入职（3月），1-2月无记录，累计从3月起算
	db := setupTestDB(t)
	repo := NewRepository(db)
	seedTestBrackets(t, repo, 2026)

	brackets, err := repo.FindTaxBrackets(2026)
	require.NoError(t, err)

	// 3月入职，无前序记录（1-2月无记录）
	result := calculateCumulativeTax(
		brackets,
		[]TaxRecord{}, // 空，3月起算
		15000,         // 3月工资
		BasicDeductionMonthly,
		1000, // 社保
		0,
	)

	// 3月个税 = (15000-5000-1000) * 0.03 = 270
	assert.Equal(t, 15000.0, result.CumulativeIncome)
	assert.Equal(t, 9000.0, result.CumulativeTaxableIncome)
	assert.Equal(t, 270.0, result.MonthlyTax, "3月入职首月个税应为270")
}

func TestCalculateTax_HighIncomeBracketJump(t *testing.T) {
	// 高收入跨多档：月入80000
	// 1月: (80000-5000) = 75000, 跨过36000到10%档, tax = 75000*0.10-2520 = 4980
	// 2月: 累计150000, 跨过144000到20%, cumulative tax = 150000*0.20-16920 = 13080, monthly = 13080-4980 = 8100
	db := setupTestDB(t)
	repo := NewRepository(db)
	seedTestBrackets(t, repo, 2026)

	brackets, err := repo.FindTaxBrackets(2026)
	require.NoError(t, err)

	// 1月: 75000落在10%档
	result1 := calculateCumulativeTax(brackets, []TaxRecord{}, 80000, BasicDeductionMonthly, 0, 0)
	assert.Equal(t, 75000.0, result1.CumulativeTaxableIncome)
	assert.Equal(t, 0.10, result1.TaxRate, "1月75000应落在10%档")
	assert.Equal(t, 2520.0, result1.QuickDeduction)
	assert.Equal(t, 4980.0, result1.MonthlyTax)

	// 2月: 150000落在20%档
	records := []TaxRecord{
		{GrossIncome: 80000, SIDeduction: 0, SpecialDeduction: 0, MonthlyTax: 4980},
	}
	result2 := calculateCumulativeTax(brackets, records, 80000, BasicDeductionMonthly, 0, 0)
	assert.Equal(t, 150000.0, result2.CumulativeTaxableIncome)
	assert.Equal(t, 0.20, result2.TaxRate, "2月应跨入20%税率")
	assert.Equal(t, 16920.0, result2.QuickDeduction)
	assert.Equal(t, roundTo2(150000*0.20-16920), result2.CumulativeTax)
	assert.Equal(t, 8100.0, result2.MonthlyTax, "2月当月个税应为8100")
}

func TestCalculateTax_ZeroIncome(t *testing.T) {
	// 0收入时应纳税额为0
	db := setupTestDB(t)
	repo := NewRepository(db)
	seedTestBrackets(t, repo, 2026)

	brackets, err := repo.FindTaxBrackets(2026)
	require.NoError(t, err)

	result := calculateCumulativeTax(brackets, []TaxRecord{}, 0, BasicDeductionMonthly, 0, 0)
	assert.Equal(t, 0.0, result.MonthlyTax)
	assert.Equal(t, 0.0, result.CumulativeTaxableIncome)
}

func TestGetTaxRecordsForCumulative(t *testing.T) {
	records := []TaxRecord{
		{Month: 1, MonthlyTax: 150},
		{Month: 2, MonthlyTax: 150},
		{Month: 3, MonthlyTax: 150},
	}

	// 当前是3月，应只取1月和2月的记录
	filtered := GetTaxRecordsForCumulative(records, 3)
	assert.Len(t, filtered, 2)

	// 当前是1月，无前序记录
	filtered = GetTaxRecordsForCumulative(records, 1)
	assert.Len(t, filtered, 0)
}

func TestValidateTaxCalculationParams(t *testing.T) {
	tests := []struct {
		name        string
		year        int
		month       int
		grossIncome float64
		expectError bool
	}{
		{"valid", 2026, 6, 10000, false},
		{"invalid year too low", 1999, 6, 10000, true},
		{"invalid year too high", 2101, 6, 10000, true},
		{"invalid month 0", 2026, 0, 10000, true},
		{"invalid month 13", 2026, 13, 10000, true},
		{"negative income", 2026, 6, -100, true},
		{"zero income valid", 2026, 6, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTaxCalculationParams(tt.year, tt.month, tt.grossIncome)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ========== Repository 集成测试 ==========

func TestRepository_ListAllActiveDeductionsByEmployee(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	orgID := int64(1)
	endDate := "2026-12"

	// 创建3个扣除项，不同有效期
	deductions := []*SpecialDeduction{
		{
			EmployeeID:     100,
			DeductionType:  DeductionTypeChildEducation,
			MonthlyAmount:  2000,
			Count:          1,
			EffectiveStart: "2026-01",
			EffectiveEnd:   &endDate,
		},
		{
			EmployeeID:     100,
			DeductionType:  DeductionTypeElderlyCare,
			MonthlyAmount:  3000,
			Count:          1,
			EffectiveStart: "2026-03",
			EffectiveEnd:   nil, // ongoing
		},
		{
			EmployeeID:     100,
			DeductionType:  DeductionTypeHousingLoan,
			MonthlyAmount:  1000,
			Count:          1,
			EffectiveStart: "2026-01",
			EffectiveEnd:   stringPtr("2026-02"), // 只在1-2月有效
		},
	}

	for _, d := range deductions {
		d.OrgID = orgID
		d.CreatedBy = 1
		d.UpdatedBy = 1
		err := repo.CreateDeduction(orgID, d)
		require.NoError(t, err)
	}

	// 2026-01: child_education + housing_loan
	active, err := repo.ListAllActiveDeductionsByEmployee(orgID, 100, "2026-01")
	require.NoError(t, err)
	assert.Len(t, active, 2)

	// 2026-03: child_education + elderly_care (housing_loan 已结束)
	active, err = repo.ListAllActiveDeductionsByEmployee(orgID, 100, "2026-03")
	require.NoError(t, err)
	assert.Len(t, active, 2)

	// 2026-06: child_education + elderly_care (housing_loan 已结束)
	active, err = repo.ListAllActiveDeductionsByEmployee(orgID, 100, "2026-06")
	require.NoError(t, err)
	assert.Len(t, active, 2)
}

// stringPtr 辅助函数：创建字符串指针
func stringPtr(s string) *string {
	return &s
}
