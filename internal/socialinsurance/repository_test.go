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
	assert.Equal(t, int64(110100000000), found.CityCode)
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
	found, err := repo.FindByCityAndYear(110100000000, 2025)
	require.NoError(t, err)
	assert.Equal(t, 2025, found.EffectiveYear)

	// 查询2023年 -> 应返回2023年政策
	found, err = repo.FindByCityAndYear(110100000000, 2023)
	require.NoError(t, err)
	assert.Equal(t, 2023, found.EffectiveYear)

	// 查询2026年（未来年份，存在2025年政策）-> 应返回2025年政策
	found, err = repo.FindByCityAndYear(110100000000, 2026)
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
		policy.CityCode = int64(i)
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
	assert.Equal(t, int64(3), policies[0].CityCode)
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

// --- Repository Tests: Record CRUD ---

func TestRecordCRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	// 创建政策
	require.NoError(t, repo.Create(createBeijing2025Policy()))
	policy, _ := repo.FindByCityAndYear(110100000000, 2025)

	// Create Record
	record := &SocialInsuranceRecord{
		EmployeeID:    1,
		EmployeeName:  "张三",
		CityCode:        110100000000,
		PolicyID:      policy.ID,
		BaseAmount:    7162.0,
		Status:        SIStatusActive,
		StartMonth:    "2025-07",
		TotalCompany:  3930.0,
		TotalPersonal: 2250.0,
	}
	record.OrgID = 100

	err := repo.CreateRecord(record)
	require.NoError(t, err)
	assert.Greater(t, record.ID, int64(0))

	// FindRecordByID
	found, err := repo.FindRecordByID(100, record.ID)
	require.NoError(t, err)
	assert.Equal(t, "张三", found.EmployeeName)
	assert.Equal(t, SIStatusActive, found.Status)

	// FindActiveRecordByEmployee
	activeRecord, err := repo.FindActiveRecordByEmployee(100, 1)
	require.NoError(t, err)
	assert.Equal(t, record.ID, activeRecord.ID)

	// UpdateRecord
	err = repo.UpdateRecord(100, record.ID, map[string]interface{}{
		"status":    SIStatusStopped,
		"end_month": "2025-09",
	})
	require.NoError(t, err)

	updated, err := repo.FindRecordByID(100, record.ID)
	require.NoError(t, err)
	assert.Equal(t, SIStatusStopped, updated.Status)
	assert.NotNil(t, updated.EndMonth)
	assert.Equal(t, "2025-09", *updated.EndMonth)

	// FindActiveRecordByEmployee should fail now
	_, err = repo.FindActiveRecordByEmployee(100, 1)
	assert.Error(t, err)
}

func TestListRecords_WithFilters(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	// 创建多条记录
	for i := 1; i <= 3; i++ {
		record := &SocialInsuranceRecord{
			EmployeeID:    int64(i),
			EmployeeName:  string(rune('张' + i - 1)),
			CityCode:        110100000000,
			PolicyID:      1,
			BaseAmount:    7162.0,
			Status:        SIStatusActive,
			StartMonth:    "2025-07",
			TotalCompany:  3930.0,
			TotalPersonal: 2250.0,
		}
		record.OrgID = 100
		require.NoError(t, repo.CreateRecord(record))
	}

	// 停缴一个
	repo.UpdateRecord(100, 1, map[string]interface{}{"status": SIStatusStopped})

	// 查询全部
	records, total, err := repo.ListRecords(100, "", "", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, records, 3)

	// 按状态筛选 active
	records, total, err = repo.ListRecords(100, SIStatusActive, "", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, records, 2)

	// 按状态筛选 stopped
	records, total, err = repo.ListRecords(100, SIStatusStopped, "", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, records, 1)
}

// --- Repository Tests: Change History CRUD ---

func TestChangeHistoryCRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	// 创建变更历史
	history := &ChangeHistory{
		RecordID:   1,
		EmployeeID: 1,
		ChangeType: SIChangeEnroll,
		AfterValue: nil,
		Remark:     "批量参保",
	}
	history.OrgID = 100

	err := repo.CreateChangeHistory(history)
	require.NoError(t, err)
	assert.Greater(t, history.ID, int64(0))

	// 查询
	histories, total, err := repo.ListChangeHistories(100, 1, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, histories, 1)
	assert.Equal(t, SIChangeEnroll, histories[0].ChangeType)
	assert.Equal(t, "批量参保", histories[0].Remark)
}
