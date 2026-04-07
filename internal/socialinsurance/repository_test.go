package socialinsurance

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicyCRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	// Create
	policy := createBeijing2025Policy()
	err := repo.Create(policy)
	require.NoError(t, err)
	assert.Greater(t, policy.ID, int64(0))
	originalID := policy.ID

	// Read
	found, err := repo.FindByID(originalID)
	require.NoError(t, err)
	assert.Equal(t, 1, found.CityID)
	assert.Equal(t, 2025, found.EffectiveYear)
	config := found.Config.Data()
	assert.InDelta(t, 0.16, config.Pension.CompanyRate, 0.001)

	// Update
	err = repo.Update(originalID, map[string]interface{}{
		"effective_year": 2026,
	})
	require.NoError(t, err)

	found, err = repo.FindByID(originalID)
	require.NoError(t, err)
	assert.Equal(t, 2026, found.EffectiveYear)

	// Delete
	err = repo.Delete(originalID)
	require.NoError(t, err)

	_, err = repo.FindByID(originalID)
	assert.Error(t, err)
}

func TestFindByCityAndYear_MultipleYears(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	// 创建多个年度政策
	for _, year := range []int{2022, 2023, 2024, 2025} {
		policy := createBeijing2025Policy()
		policy.EffectiveYear = year
		require.NoError(t, repo.Create(policy))
	}

	// 查询2025年 -> 应返回2025年政策
	found, err := repo.FindByCityAndYear(1, 2025)
	require.NoError(t, err)
	assert.Equal(t, 2025, found.EffectiveYear)

	// 查询2023年 -> 应返回2023年政策
	found, err = repo.FindByCityAndYear(1, 2023)
	require.NoError(t, err)
	assert.Equal(t, 2023, found.EffectiveYear)

	// 查询2026年（未来年份，存在2025年政策）-> 应返回2025年政策
	found, err = repo.FindByCityAndYear(1, 2026)
	require.NoError(t, err)
	assert.Equal(t, 2025, found.EffectiveYear)

	// 查询2021年（早于所有政策）-> 应报错
	_, err = repo.FindByCityAndYear(1, 2021)
	assert.Error(t, err)
}

func TestListPolicies_Pagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	// 创建5条政策（不同城市）
	for i := 1; i <= 5; i++ {
		policy := createBeijing2025Policy()
		policy.CityID = i
		require.NoError(t, repo.Create(policy))
	}

	// 第1页，每页2条
	policies, total, err := repo.List(0, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, policies, 2)

	// 第2页
	policies, total, err = repo.List(0, 2, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, policies, 2)

	// 第3页
	policies, total, err = repo.List(0, 3, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, policies, 1)

	// 按城市筛选
	policies, total, err = repo.List(3, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, policies, 1)
	assert.Equal(t, 3, policies[0].CityID)
}

func TestSoftDelete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	policy := createBeijing2025Policy()
	require.NoError(t, repo.Create(policy))

	// 软删除
	err := repo.Delete(policy.ID)
	require.NoError(t, err)

	// FindByID 查不到（已软删除）
	_, err = repo.FindByID(policy.ID)
	assert.Error(t, err)

	// FindByCityAndYear 也查不到
	_, err = repo.FindByCityAndYear(1, 2025)
	assert.Error(t, err)

	// List 也查不到
	_, total, err := repo.List(0, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)

	// 确认记录仍在数据库中（有 deleted_at 值）
	var count int64
	db.Unscoped().Model(&SocialInsurancePolicy{}).Count(&count)
	assert.Equal(t, int64(1), count)
}
