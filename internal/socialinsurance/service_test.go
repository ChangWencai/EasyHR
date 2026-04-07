package socialinsurance

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&SocialInsurancePolicy{})
	require.NoError(t, err)
	return db
}

func createBeijing2025Policy() *SocialInsurancePolicy {
	return &SocialInsurancePolicy{
		CityID:        1,
		EffectiveYear: 2025,
		Config: newJSONType(FiveInsurances{
			Pension: InsuranceItem{
				CompanyRate:  0.16,
				PersonalRate: 0.08,
				BaseLower:    7162,
				BaseUpper:    35811,
			},
			Medical: InsuranceItem{
				CompanyRate:  0.098,
				PersonalRate: 0.02,
				BaseLower:    7162,
				BaseUpper:    35811,
			},
			Unemployment: InsuranceItem{
				CompanyRate:  0.005,
				PersonalRate: 0.005,
				BaseLower:    7162,
				BaseUpper:    35811,
			},
			WorkInjury: InsuranceItem{
				CompanyRate:  0.002,
				PersonalRate: 0.0,
				BaseLower:    7162,
				BaseUpper:    35811,
			},
			Maternity: InsuranceItem{
				CompanyRate:  0.008,
				PersonalRate: 0.0,
				BaseLower:    7162,
				BaseUpper:    35811,
			},
			HousingFund: InsuranceItem{
				CompanyRate:  0.12,
				PersonalRate: 0.12,
				BaseLower:    7162,
				BaseUpper:    35811,
			},
		}),
	}
}

// --- Service Tests ---

func TestCalculateInsuranceAmounts(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	svc := NewService(repo)

	policy := createBeijing2025Policy()
	err := repo.Create(policy)
	require.NoError(t, err)

	resp, err := svc.CalculateInsuranceAmounts(1, 10000, 2025)
	require.NoError(t, err)

	assert.Equal(t, "北京", resp.CityName)
	assert.Equal(t, 10000.0, resp.Salary)

	// 基数应使用薪资（在下限和上限之间）
	assert.Equal(t, 10000.0, resp.BaseAmount)

	// 养老保险：企业 10000 * 0.16 = 1600, 个人 10000 * 0.08 = 800
	pensionItem := findItem(resp.Items, "养老保险")
	require.NotNil(t, pensionItem)
	assert.Equal(t, 10000.0, pensionItem.Base)
	assert.InDelta(t, 1600.0, pensionItem.CompanyAmount, 0.01)
	assert.InDelta(t, 800.0, pensionItem.PersonalAmount, 0.01)

	// 医疗保险：企业 10000 * 0.098 = 980, 个人 10000 * 0.02 = 200
	medicalItem := findItem(resp.Items, "医疗保险")
	require.NotNil(t, medicalItem)
	assert.InDelta(t, 980.0, medicalItem.CompanyAmount, 0.01)
	assert.InDelta(t, 200.0, medicalItem.PersonalAmount, 0.01)

	// 验证总计: 1600+980+50+20+80+1200=3930(企业), 800+200+50+0+0+1200=2250(个人)
	assert.InDelta(t, 3930.0, resp.TotalCompany, 0.01)
	assert.InDelta(t, 2250.0, resp.TotalPersonal, 0.01)
}

func TestCalculateInsuranceAmounts_BelowLower(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	svc := NewService(repo)

	policy := createBeijing2025Policy()
	err := repo.Create(policy)
	require.NoError(t, err)

	resp, err := svc.CalculateInsuranceAmounts(1, 3000, 2025)
	require.NoError(t, err)

	// 薪资3000低于下限7162，基数应使用下限7162
	assert.Equal(t, 7162.0, resp.BaseAmount)

	for _, item := range resp.Items {
		assert.Equal(t, 7162.0, item.Base, "险种 %s 的基数应为下限7162", item.Name)
	}

	// 验证养老: 7162 * 0.16 = 1145.92
	pensionItem := findItem(resp.Items, "养老保险")
	require.NotNil(t, pensionItem)
	assert.InDelta(t, 1145.92, pensionItem.CompanyAmount, 0.01)
	assert.InDelta(t, 572.96, pensionItem.PersonalAmount, 0.01)
}

func TestCalculateInsuranceAmounts_AboveUpper(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	svc := NewService(repo)

	policy := createBeijing2025Policy()
	err := repo.Create(policy)
	require.NoError(t, err)

	resp, err := svc.CalculateInsuranceAmounts(1, 50000, 2025)
	require.NoError(t, err)

	// 薪资50000高于上限35811，基数应使用上限35811
	assert.Equal(t, 35811.0, resp.BaseAmount)

	for _, item := range resp.Items {
		assert.Equal(t, 35811.0, item.Base, "险种 %s 的基数应为上限35811", item.Name)
	}

	// 验证养老: 35811 * 0.16 = 5729.76
	pensionItem := findItem(resp.Items, "养老保险")
	require.NotNil(t, pensionItem)
	assert.InDelta(t, 5729.76, pensionItem.CompanyAmount, 0.01)
	assert.InDelta(t, 2864.88, pensionItem.PersonalAmount, 0.01)
}

func TestWorkInjuryAndMaternityPersonalRateZero(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	svc := NewService(repo)

	policy := createBeijing2025Policy()
	err := repo.Create(policy)
	require.NoError(t, err)

	resp, err := svc.CalculateInsuranceAmounts(1, 10000, 2025)
	require.NoError(t, err)

	// 工伤保险个人缴费金额必须为0
	workInjuryItem := findItem(resp.Items, "工伤保险")
	require.NotNil(t, workInjuryItem)
	assert.Equal(t, 0.0, workInjuryItem.PersonalRate)
	assert.True(t, math.Abs(workInjuryItem.PersonalAmount) < 0.001,
		"工伤保险个人缴费金额必须为0，实际为 %f", workInjuryItem.PersonalAmount)
	assert.Greater(t, workInjuryItem.CompanyAmount, 0.0,
		"工伤保险企业缴费金额应大于0")

	// 生育保险个人缴费金额必须为0
	maternityItem := findItem(resp.Items, "生育保险")
	require.NotNil(t, maternityItem)
	assert.Equal(t, 0.0, maternityItem.PersonalRate)
	assert.True(t, math.Abs(maternityItem.PersonalAmount) < 0.001,
		"生育保险个人缴费金额必须为0，实际为 %f", maternityItem.PersonalAmount)
	assert.Greater(t, maternityItem.CompanyAmount, 0.0,
		"生育保险企业缴费金额应大于0")
}

// --- Repository Tests ---

func TestCreatePolicy(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	policy := createBeijing2025Policy()
	err := repo.Create(policy)
	require.NoError(t, err)
	assert.Greater(t, policy.ID, int64(0))

	// 验证JSONB存取
	found, err := repo.FindByID(policy.ID)
	require.NoError(t, err)
	config := found.Config.Data()
	assert.InDelta(t, 0.16, config.Pension.CompanyRate, 0.001)
	assert.InDelta(t, 0.08, config.Pension.PersonalRate, 0.001)
	assert.Equal(t, 7162.0, config.Pension.BaseLower)
	assert.Equal(t, 35811.0, config.Pension.BaseUpper)
}

func TestFindPolicyByCityAndYear(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	// 创建2024年政策
	policy2024 := createBeijing2025Policy()
	policy2024.EffectiveYear = 2024
	err := repo.Create(policy2024)
	require.NoError(t, err)

	// 创建2025年政策
	policy2025 := createBeijing2025Policy()
	err = repo.Create(policy2025)
	require.NoError(t, err)

	// 查询2025年应返回2025年政策（effective_year <= 2025 的最新记录）
	found, err := repo.FindByCityAndYear(1, 2025)
	require.NoError(t, err)
	assert.Equal(t, 2025, found.EffectiveYear)

	// 查询2025年但只有2024年政策时，应返回2024年政策
	found, err = repo.FindByCityAndYear(1, 2024)
	require.NoError(t, err)
	assert.Equal(t, 2024, found.EffectiveYear)

	// 查询不存在城市的政策应返回错误
	_, err = repo.FindByCityAndYear(999, 2025)
	assert.Error(t, err)
}

func TestListPolicies(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	// 创建多个政策
	policy1 := createBeijing2025Policy()
	require.NoError(t, repo.Create(policy1))

	policy2 := createBeijing2025Policy()
	policy2.CityID = 2
	policy2.EffectiveYear = 2025
	require.NoError(t, repo.Create(policy2))

	// 查询全部
	policies, total, err := repo.List(0, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, policies, 2)

	// 按城市筛选
	policies, total, err = repo.List(1, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, policies, 1)
	assert.Equal(t, 1, policies[0].CityID)

	// 分页测试
	policies, total, err = repo.List(0, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, policies, 1)
}

func TestUpdatePolicy(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	policy := createBeijing2025Policy()
	require.NoError(t, repo.Create(policy))

	// 更新JSONB配置
	newConfig := newJSONType(FiveInsurances{
		Pension: InsuranceItem{
			CompanyRate:  0.20,
			PersonalRate: 0.10,
			BaseLower:    8000,
			BaseUpper:    40000,
		},
		Medical: InsuranceItem{
			CompanyRate:  0.10,
			PersonalRate: 0.02,
			BaseLower:    8000,
			BaseUpper:    40000,
		},
		Unemployment: InsuranceItem{
			CompanyRate:  0.005,
			PersonalRate: 0.005,
			BaseLower:    8000,
			BaseUpper:    40000,
		},
		WorkInjury: InsuranceItem{
			CompanyRate:  0.003,
			PersonalRate: 0.0,
			BaseLower:    8000,
			BaseUpper:    40000,
		},
		Maternity: InsuranceItem{
			CompanyRate:  0.01,
			PersonalRate: 0.0,
			BaseLower:    8000,
			BaseUpper:    40000,
		},
		HousingFund: InsuranceItem{
			CompanyRate:  0.12,
			PersonalRate: 0.12,
			BaseLower:    8000,
			BaseUpper:    40000,
		},
	})

	err := repo.Update(policy.ID, map[string]interface{}{
		"config": newConfig,
	})
	require.NoError(t, err)

	found, err := repo.FindByID(policy.ID)
	require.NoError(t, err)
	config := found.Config.Data()
	assert.InDelta(t, 0.20, config.Pension.CompanyRate, 0.001)
	assert.Equal(t, 8000.0, config.Pension.BaseLower)
}

func TestDeletePolicy(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	policy := createBeijing2025Policy()
	require.NoError(t, repo.Create(policy))

	// 软删除
	err := repo.Delete(policy.ID)
	require.NoError(t, err)

	// 删除后查询不到
	_, err = repo.FindByID(policy.ID)
	assert.Error(t, err)
}

// --- Helper ---

func findItem(items []InsuranceAmountDetail, name string) *InsuranceAmountDetail {
	for i := range items {
		if items[i].Name == name {
			return &items[i]
		}
	}
	return nil
}
