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
	err = db.AutoMigrate(
		&SocialInsurancePolicy{},
		&SocialInsuranceRecord{},
		&ChangeHistory{},
	)
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

// --- Mock EmployeeQuerier ---

type mockEmployeeQuerier struct {
	employees     map[int64]EmployeeInfo
	userToEmployee map[int64]EmployeeInfo
}

func newMockEmployeeQuerier() *mockEmployeeQuerier {
	return &mockEmployeeQuerier{
		employees:     make(map[int64]EmployeeInfo),
		userToEmployee: make(map[int64]EmployeeInfo),
	}
}

func (m *mockEmployeeQuerier) addEmployee(id int64, name string, orgID int64, userID *int64) {
	info := EmployeeInfo{ID: id, Name: name, OrgID: orgID, UserID: userID}
	m.employees[id] = info
	if userID != nil {
		m.userToEmployee[*userID] = info
	}
}

func (m *mockEmployeeQuerier) FindEmployeeByIDs(orgID int64, ids []int64) ([]EmployeeInfo, error) {
	var result []EmployeeInfo
	for _, id := range ids {
		if emp, ok := m.employees[id]; ok && emp.OrgID == orgID {
			result = append(result, emp)
		}
	}
	return result, nil
}

func (m *mockEmployeeQuerier) FindEmployeeByUserID(orgID int64, userID int64) (*EmployeeInfo, error) {
	if emp, ok := m.userToEmployee[userID]; ok && emp.OrgID == orgID {
		return &emp, nil
	}
	return nil, gorm.ErrRecordNotFound
}

// --- Service Tests: Policy & Calculate (keep existing) ---

func TestCalculateInsuranceAmounts(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	svc := NewService(repo, nil)

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

	// 验证总计: 1600+980+50+20+80+1200=3930(企业), 800+200+50+0+0+1200=2250(个人)
	assert.InDelta(t, 3930.0, resp.TotalCompany, 0.01)
	assert.InDelta(t, 2250.0, resp.TotalPersonal, 0.01)
}

func TestCalculateInsuranceAmounts_BelowLower(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	svc := NewService(repo, nil)

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
	svc := NewService(repo, nil)

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
	svc := NewService(repo, nil)

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

// --- Service Tests: Batch Enroll ---

func TestEnrollEmployees(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	mockEmp := newMockEmployeeQuerier()
	svc := NewService(repo, mockEmp)

	// 创建政策
	policy := createBeijing2025Policy()
	require.NoError(t, repo.Create(policy))

	// mock 3个员工
	mockEmp.addEmployee(1, "张三", 100, nil)
	mockEmp.addEmployee(2, "李四", 100, nil)
	mockEmp.addEmployee(3, "王五", 100, nil)

	// 批量参保
	result, err := svc.BatchEnroll(100, 1, &BatchEnrollRequest{
		EmployeeIDs: []int64{1, 2, 3},
		CityID:      1,
		StartMonth:  "2025-07",
	})
	require.NoError(t, err)

	// 验证结果
	assert.Equal(t, 3, result.SuccessCount)
	assert.Equal(t, 0, result.FailCount)

	// 验证参保记录
	record1, err := repo.FindActiveRecordByEmployee(100, 1)
	require.NoError(t, err)
	assert.Equal(t, SIStatusActive, record1.Status)
	assert.Equal(t, "张三", record1.EmployeeName)
	assert.Equal(t, "2025-07", record1.StartMonth)
	assert.Equal(t, 1, record1.CityID)
	assert.Equal(t, policy.ID, record1.PolicyID)
	assert.InDelta(t, 7162.0, record1.BaseAmount, 0.01)
	// 企业合计：基数7162时 1145.92+701.88+35.81+14.32+57.30+859.44=2814.67
	assert.InDelta(t, 2814.67, record1.TotalCompany, 0.01)

	// 验证变更历史
	histories, total, err := repo.ListChangeHistories(100, 1, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, SIChangeEnroll, histories[0].ChangeType)
}

func TestEnrollEmployees_DuplicateEnrollment(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	mockEmp := newMockEmployeeQuerier()
	svc := NewService(repo, mockEmp)

	// 创建政策
	require.NoError(t, repo.Create(createBeijing2025Policy()))

	// mock 1个员工
	mockEmp.addEmployee(1, "张三", 100, nil)

	// 第一次参保
	result, err := svc.BatchEnroll(100, 1, &BatchEnrollRequest{
		EmployeeIDs: []int64{1},
		CityID:      1,
		StartMonth:  "2025-07",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, result.SuccessCount)

	// 重复参保应失败
	result, err = svc.BatchEnroll(100, 1, &BatchEnrollRequest{
		EmployeeIDs: []int64{1},
		CityID:      1,
		StartMonth:  "2025-08",
	})
	require.NoError(t, err)
	assert.Equal(t, 0, result.SuccessCount)
	assert.Equal(t, 1, result.FailCount)
	assert.Equal(t, "该员工已有参保中记录", result.Failures[0].Reason)
}

func TestEnrollEmployees_NoPolicy(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	mockEmp := newMockEmployeeQuerier()
	svc := NewService(repo, mockEmp)

	// 不创建政策
	mockEmp.addEmployee(1, "张三", 100, nil)

	result, err := svc.BatchEnroll(100, 1, &BatchEnrollRequest{
		EmployeeIDs: []int64{1},
		CityID:      999, // 不存在的城市
		StartMonth:  "2025-07",
	})
	require.NoError(t, err)
	assert.Equal(t, 0, result.SuccessCount)
	assert.Equal(t, 1, result.FailCount)
	assert.Equal(t, "该城市无社保政策", result.Failures[0].Reason)
}

func TestEnrollEmployees_PartialFailure(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	mockEmp := newMockEmployeeQuerier()
	svc := NewService(repo, mockEmp)

	// 只为城市1创建政策
	require.NoError(t, repo.Create(createBeijing2025Policy()))

	// mock 3个员工
	mockEmp.addEmployee(1, "张三", 100, nil)
	mockEmp.addEmployee(2, "李四", 100, nil)
	mockEmp.addEmployee(3, "王五", 100, nil)

	// 1号员工先参保（制造重复）
	_, err := svc.BatchEnroll(100, 1, &BatchEnrollRequest{
		EmployeeIDs: []int64{1},
		CityID:      1,
		StartMonth:  "2025-07",
	})
	require.NoError(t, err)

	// 批量参保包含重复 + 不存在城市的场景
	// 注意：这里3个员工使用城市1，但1号已参保，所以2成功1失败
	result, err := svc.BatchEnroll(100, 1, &BatchEnrollRequest{
		EmployeeIDs: []int64{1, 2, 3},
		CityID:      1,
		StartMonth:  "2025-07",
	})
	require.NoError(t, err)
	assert.Equal(t, 2, result.SuccessCount)
	assert.Equal(t, 1, result.FailCount)
	assert.Len(t, result.Failures, 1)
	assert.Equal(t, int64(1), result.Failures[0].EmployeeID)
}

// --- Service Tests: Batch Stop ---

func TestStopEnrollment(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	mockEmp := newMockEmployeeQuerier()
	svc := NewService(repo, mockEmp)

	require.NoError(t, repo.Create(createBeijing2025Policy()))
	mockEmp.addEmployee(1, "张三", 100, nil)

	// 先参保
	enrollResult, err := svc.BatchEnroll(100, 1, &BatchEnrollRequest{
		EmployeeIDs: []int64{1},
		CityID:      1,
		StartMonth:  "2025-07",
	})
	require.NoError(t, err)
	require.Equal(t, 1, enrollResult.SuccessCount)

	// 获取记录ID
	record, err := repo.FindActiveRecordByEmployee(100, 1)
	require.NoError(t, err)

	// 停缴
	stopResult, err := svc.BatchStopEnrollment(100, 1, &BatchStopRequest{
		RecordIDs: []int64{record.ID},
		EndMonth:  "2025-09",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, stopResult.SuccessCount)
	assert.Equal(t, 0, stopResult.FailCount)

	// 验证状态已变为 stopped
	updated, err := repo.FindRecordByID(100, record.ID)
	require.NoError(t, err)
	assert.Equal(t, SIStatusStopped, updated.Status)
	assert.NotNil(t, updated.EndMonth)
	assert.Equal(t, "2025-09", *updated.EndMonth)

	// 验证变更历史
	histories, _, err := repo.ListChangeHistories(100, 1, 1, 10)
	require.NoError(t, err)
	// 应该有 enroll + stop 两条历史
	assert.Len(t, histories, 2)

	// 找到 stop 类型的历史
	var stopHistory *ChangeHistory
	for i := range histories {
		if histories[i].ChangeType == SIChangeStop {
			stopHistory = &histories[i]
			break
		}
	}
	require.NotNil(t, stopHistory)
	assert.Equal(t, SIChangeStop, stopHistory.ChangeType)
	assert.NotNil(t, stopHistory.BeforeValue)
	assert.NotNil(t, stopHistory.AfterValue)
}

func TestStopEnrollment_NotActive(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	mockEmp := newMockEmployeeQuerier()
	svc := NewService(repo, mockEmp)

	require.NoError(t, repo.Create(createBeijing2025Policy()))
	mockEmp.addEmployee(1, "张三", 100, nil)

	// 参保
	enrollResult, err := svc.BatchEnroll(100, 1, &BatchEnrollRequest{
		EmployeeIDs: []int64{1},
		CityID:      1,
		StartMonth:  "2025-07",
	})
	require.NoError(t, err)
	require.Equal(t, 1, enrollResult.SuccessCount)

	record, err := repo.FindActiveRecordByEmployee(100, 1)
	require.NoError(t, err)

	// 第一次停缴
	stopResult, err := svc.BatchStopEnrollment(100, 1, &BatchStopRequest{
		RecordIDs: []int64{record.ID},
		EndMonth:  "2025-09",
	})
	require.NoError(t, err)
	require.Equal(t, 1, stopResult.SuccessCount)

	// 再次停缴应失败（已 stopped）
	stopResult, err = svc.BatchStopEnrollment(100, 1, &BatchStopRequest{
		RecordIDs: []int64{record.ID},
		EndMonth:  "2025-10",
	})
	require.NoError(t, err)
	assert.Equal(t, 0, stopResult.SuccessCount)
	assert.Equal(t, 1, stopResult.FailCount)
	assert.Equal(t, "参保记录非参保中状态，无法停缴", stopResult.Failures[0].Reason)
}

func TestBatchStopEnrollment(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	mockEmp := newMockEmployeeQuerier()
	svc := NewService(repo, mockEmp)

	require.NoError(t, repo.Create(createBeijing2025Policy()))
	mockEmp.addEmployee(1, "张三", 100, nil)
	mockEmp.addEmployee(2, "李四", 100, nil)
	mockEmp.addEmployee(3, "王五", 100, nil)

	// 批量参保
	enrollResult, err := svc.BatchEnroll(100, 1, &BatchEnrollRequest{
		EmployeeIDs: []int64{1, 2, 3},
		CityID:      1,
		StartMonth:  "2025-07",
	})
	require.NoError(t, err)
	require.Equal(t, 3, enrollResult.SuccessCount)

	// 获取所有参保记录
	var records []SocialInsuranceRecord
	err = db.Where("org_id = ? AND status = ?", 100, SIStatusActive).Find(&records).Error
	require.NoError(t, err)
	require.Len(t, records, 3)

	// 批量停缴
	var recordIDs []int64
	for _, r := range records {
		recordIDs = append(recordIDs, r.ID)
	}

	stopResult, err := svc.BatchStopEnrollment(100, 1, &BatchStopRequest{
		RecordIDs: recordIDs,
		EndMonth:  "2025-09",
	})
	require.NoError(t, err)
	assert.Equal(t, 3, stopResult.SuccessCount)
	assert.Equal(t, 0, stopResult.FailCount)

	// 验证所有记录状态
	var activeCount int64
	db.Model(&SocialInsuranceRecord{}).Where("org_id = ? AND status = ?", 100, SIStatusActive).Count(&activeCount)
	assert.Equal(t, int64(0), activeCount)

	var stoppedCount int64
	db.Model(&SocialInsuranceRecord{}).Where("org_id = ? AND status = ?", 100, SIStatusStopped).Count(&stoppedCount)
	assert.Equal(t, int64(3), stoppedCount)
}

// --- Service Tests: List Records ---

func TestListRecords_FilterByStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	mockEmp := newMockEmployeeQuerier()
	svc := NewService(repo, mockEmp)

	require.NoError(t, repo.Create(createBeijing2025Policy()))
	mockEmp.addEmployee(1, "张三", 100, nil)
	mockEmp.addEmployee(2, "李四", 100, nil)

	// 参保2个员工
	_, err := svc.BatchEnroll(100, 1, &BatchEnrollRequest{
		EmployeeIDs: []int64{1, 2},
		CityID:      1,
		StartMonth:  "2025-07",
	})
	require.NoError(t, err)

	// 停缴1个
	record, err := repo.FindActiveRecordByEmployee(100, 1)
	require.NoError(t, err)
	_, err = svc.BatchStopEnrollment(100, 1, &BatchStopRequest{
		RecordIDs: []int64{record.ID},
		EndMonth:  "2025-09",
	})
	require.NoError(t, err)

	// 查询 active 记录
	records, total, _, _, err := svc.ListRecords(100, RecordListQueryParams{Status: SIStatusActive})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, records, 1)
	assert.Equal(t, SIStatusActive, records[0].Status)

	// 查询 stopped 记录
	records, total, _, _, err = svc.ListRecords(100, RecordListQueryParams{Status: SIStatusStopped})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, records, 1)
	assert.Equal(t, SIStatusStopped, records[0].Status)

	// 查询全部
	records, total, _, _, err = svc.ListRecords(100, RecordListQueryParams{})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, records, 2)
}

// --- Service Tests: Get My Records (MEMBER role) ---

func TestGetMyRecords_MemberRole(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	mockEmp := newMockEmployeeQuerier()
	svc := NewService(repo, mockEmp)

	require.NoError(t, repo.Create(createBeijing2025Policy()))

	// mock 员工（关联 user_id=10）
	userID := int64(10)
	mockEmp.addEmployee(1, "张三", 100, &userID)

	// 参保
	_, err := svc.BatchEnroll(100, 1, &BatchEnrollRequest{
		EmployeeIDs: []int64{1},
		CityID:      1,
		StartMonth:  "2025-07",
	})
	require.NoError(t, err)

	// 通过 user_id 查询自己的记录
	records, err := svc.GetMyRecords(100, 10)
	require.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, int64(1), records[0].EmployeeID)
	assert.Equal(t, "张三", records[0].EmployeeName)
	assert.Equal(t, SIStatusActive, records[0].Status)
}

// --- Service Tests: Change History ---

func TestGetChangeHistory(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	mockEmp := newMockEmployeeQuerier()
	svc := NewService(repo, mockEmp)

	require.NoError(t, repo.Create(createBeijing2025Policy()))
	mockEmp.addEmployee(1, "张三", 100, nil)

	// 参保（生成 enroll 历史）
	_, err := svc.BatchEnroll(100, 1, &BatchEnrollRequest{
		EmployeeIDs: []int64{1},
		CityID:      1,
		StartMonth:  "2025-07",
	})
	require.NoError(t, err)

	// 停缴（生成 stop 历史）
	record, err := repo.FindActiveRecordByEmployee(100, 1)
	require.NoError(t, err)
	_, err = svc.BatchStopEnrollment(100, 1, &BatchStopRequest{
		RecordIDs: []int64{record.ID},
		EndMonth:  "2025-09",
	})
	require.NoError(t, err)

	// 查询变更历史
	histories, total, err := svc.GetChangeHistory(100, 1, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, histories, 2)

	// 按时间倒序，最新在前
	// 第一条是 stop
	assert.Equal(t, SIChangeStop, histories[0].ChangeType)
	// 第二条是 enroll
	assert.Equal(t, SIChangeEnroll, histories[1].ChangeType)
}

// --- Service Tests: Get Deduction (D-12) ---

func TestGetSocialInsuranceDeduction(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	mockEmp := newMockEmployeeQuerier()
	svc := NewService(repo, mockEmp)

	require.NoError(t, repo.Create(createBeijing2025Policy()))
	mockEmp.addEmployee(1, "张三", 100, nil)

	// 参保
	_, err := svc.BatchEnroll(100, 1, &BatchEnrollRequest{
		EmployeeIDs: []int64{1},
		CityID:      1,
		StartMonth:  "2025-07",
	})
	require.NoError(t, err)

	// 查询扣款
	deduction, err := svc.GetSocialInsuranceDeduction(100, 1, "2025-07")
	require.NoError(t, err)
	assert.NotEmpty(t, deduction.Items)

	// 验证个人扣款总计
	// 养老个人: 7162*0.08=572.96, 医疗个人: 7162*0.02=143.24, 失业个人: 7162*0.005=35.81
	// 工伤个人: 0, 生育个人: 0, 公积金个人: 7162*0.12=859.44
	// 总计: 572.96+143.24+35.81+0+0+859.44 = 1611.45
	expectedTotal := 572.96 + 143.24 + 35.81 + 859.44
	assert.InDelta(t, expectedTotal, deduction.TotalPersonal, 0.02)

	// 验证每项都有个人金额
	for _, item := range deduction.Items {
		if item.Name == "工伤保险" || item.Name == "生育保险" {
			assert.InDelta(t, 0.0, item.PersonalAmount, 0.001, "%s 个人金额应为0", item.Name)
		} else {
			assert.Greater(t, item.PersonalAmount, 0.0, "%s 个人金额应大于0", item.Name)
		}
	}
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
