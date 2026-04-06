package employee

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wencai/easyhr/internal/common/crypto"
	"github.com/wencai/easyhr/internal/common/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupOffboardingTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Discard,
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Organization{}, &model.User{}, &Employee{}, &Offboarding{}))
	return db
}

func createTestOrgForOffboarding(t *testing.T, db *gorm.DB, suffix string) int64 {
	org := &model.Organization{
		Name:       "离职测试企业" + suffix,
		CreditCode: "91110108MA0" + suffix,
		City:       "北京",
		Status:     "active",
	}
	require.NoError(t, db.Create(org).Error)
	return org.ID
}

// createActiveEmployee 辅助函数：创建在职状态的测试员工
func createActiveEmployee(db *gorm.DB, orgID int64, name, phone string) (*Employee, error) {
	emp := &Employee{}
	emp.OrgID = orgID
	emp.CreatedBy = 1
	emp.UpdatedBy = 1
	emp.Name = name
	emp.PhoneEncrypted = "enc_" + phone
	emp.PhoneHash = crypto.HashSHA256(phone)
	emp.IDCardEncrypted = "enc_idcard"
	emp.IDCardHash = crypto.HashSHA256("idcard_" + phone)
	emp.Position = "开发工程师"
	emp.HireDate = time.Now()
	emp.Status = StatusActive
	err := db.Create(emp).Error
	return emp, err
}

func newTestOffboardingService(db *gorm.DB) *OffboardingService {
	obRepo := NewOffboardingRepository(db)
	empRepo := NewRepository(db)
	return NewOffboardingService(obRepo, empRepo)
}

// TestBossResign_Success 老板直接办理离职，Employee.status 更新为 resigned，Offboarding 记录创建
func TestBossResign_Success(t *testing.T) {
	db := setupOffboardingTestDB(t)
	orgID := createTestOrgForOffboarding(t, db, "BR01")
	svc := newTestOffboardingService(db)

	emp, err := createActiveEmployee(db, orgID, "张三", "13800008001")
	require.NoError(t, err)

	req := &BossResignRequest{
		ResignationDate: "2026-04-30",
		Reason:          "个人原因离职",
	}
	err = svc.BossResign(orgID, 1, emp.ID, req)
	require.NoError(t, err)

	// 验证 Employee.status 更新为 resigned
	updatedEmp, err := NewRepository(db).FindByID(orgID, emp.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusResigned, updatedEmp.Status)
	require.NotNil(t, updatedEmp.ResignationDate)
	assert.Equal(t, "2026-04-30", updatedEmp.ResignationDate.Format("2006-01-02"))
	assert.Equal(t, "个人原因离职", updatedEmp.ResignationReason)

	// 验证 Offboarding 记录创建
	obRepo := NewOffboardingRepository(db)
	ob, err := obRepo.FindByEmployeeID(orgID, emp.ID)
	require.NoError(t, err)
	assert.Equal(t, "involuntary", ob.Type)
	assert.Equal(t, "pending", ob.Status)
	assert.Equal(t, emp.ID, ob.EmployeeID)
}

// TestBossResign_AlreadyResigned 已离职员工不可再次办理离职
func TestBossResign_AlreadyResigned(t *testing.T) {
	db := setupOffboardingTestDB(t)
	orgID := createTestOrgForOffboarding(t, db, "BR02")
	svc := newTestOffboardingService(db)

	emp, err := createActiveEmployee(db, orgID, "李四", "13800008002")
	require.NoError(t, err)

	req := &BossResignRequest{
		ResignationDate: "2026-04-30",
		Reason:          "第一次离职",
	}
	err = svc.BossResign(orgID, 1, emp.ID, req)
	require.NoError(t, err)

	// 再次办理离职应失败
	req2 := &BossResignRequest{
		ResignationDate: "2026-05-01",
		Reason:          "重复离职",
	}
	err = svc.BossResign(orgID, 1, emp.ID, req2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "已离职")
}

// TestEmployeeApplyResign 员工申请离职，创建 Offboarding（type=voluntary, status=pending）
func TestEmployeeApplyResign(t *testing.T) {
	db := setupOffboardingTestDB(t)
	orgID := createTestOrgForOffboarding(t, db, "EA01")
	svc := newTestOffboardingService(db)

	emp, err := createActiveEmployee(db, orgID, "王五", "13800008003")
	require.NoError(t, err)

	req := &EmployeeResignRequest{
		ResignationDate: "2026-05-15",
		Reason:          "职业发展",
	}
	err = svc.EmployeeApplyResign(orgID, emp.ID, req)
	require.NoError(t, err)

	// 验证 Offboarding 记录
	obRepo := NewOffboardingRepository(db)
	ob, err := obRepo.FindByEmployeeID(orgID, emp.ID)
	require.NoError(t, err)
	assert.Equal(t, "voluntary", ob.Type)
	assert.Equal(t, "pending", ob.Status)
	assert.Equal(t, emp.ID, ob.EmployeeID)

	// 员工申请离职时，Employee.status 尚未更新为 resigned（等待审批）
	updatedEmp, err := NewRepository(db).FindByID(orgID, emp.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusActive, updatedEmp.Status, "员工申请离职但尚未审批，状态应仍为 active")
}

// TestApproveResign 审批通过后 Offboarding.status=approved，Employee.status=resigned
func TestApproveResign(t *testing.T) {
	db := setupOffboardingTestDB(t)
	orgID := createTestOrgForOffboarding(t, db, "AP01")
	svc := newTestOffboardingService(db)

	emp, err := createActiveEmployee(db, orgID, "赵六", "13800008004")
	require.NoError(t, err)

	req := &EmployeeResignRequest{
		ResignationDate: "2026-05-15",
		Reason:          "家庭原因",
	}
	err = svc.EmployeeApplyResign(orgID, emp.ID, req)
	require.NoError(t, err)

	// 获取 Offboarding ID
	obRepo := NewOffboardingRepository(db)
	ob, err := obRepo.FindByEmployeeID(orgID, emp.ID)
	require.NoError(t, err)

	// 审批通过
	err = svc.ApproveResign(orgID, 1, ob.ID)
	require.NoError(t, err)

	// 验证 Offboarding.status
	ob2, err := obRepo.FindByID(orgID, ob.ID)
	require.NoError(t, err)
	assert.Equal(t, "approved", ob2.Status)
	require.NotNil(t, ob2.ApprovedBy)
	assert.Equal(t, int64(1), *ob2.ApprovedBy)
	require.NotNil(t, ob2.ApprovedAt)

	// 验证 Employee.status
	updatedEmp, err := NewRepository(db).FindByID(orgID, emp.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusResigned, updatedEmp.Status)
	require.NotNil(t, updatedEmp.ResignationDate)
	assert.Equal(t, "2026-05-15", updatedEmp.ResignationDate.Format("2006-01-02"))
}

// TestCompleteOffboarding 完成交接后 Offboarding.status=completed
func TestCompleteOffboarding(t *testing.T) {
	db := setupOffboardingTestDB(t)
	orgID := createTestOrgForOffboarding(t, db, "CO01")
	svc := newTestOffboardingService(db)

	emp, err := createActiveEmployee(db, orgID, "钱七", "13800008005")
	require.NoError(t, err)

	bossReq := &BossResignRequest{
		ResignationDate: "2026-04-30",
		Reason:          "公司裁员",
	}
	err = svc.BossResign(orgID, 1, emp.ID, bossReq)
	require.NoError(t, err)

	obRepo := NewOffboardingRepository(db)
	ob, err := obRepo.FindByEmployeeID(orgID, emp.ID)
	require.NoError(t, err)

	// 完成交接
	err = svc.CompleteOffboarding(orgID, ob.ID)
	require.NoError(t, err)

	ob2, err := obRepo.FindByID(orgID, ob.ID)
	require.NoError(t, err)
	assert.Equal(t, "completed", ob2.Status)
	require.NotNil(t, ob2.CompletedAt)
}

// TestUpdateChecklist 更新交接清单项
func TestUpdateChecklist(t *testing.T) {
	db := setupOffboardingTestDB(t)
	orgID := createTestOrgForOffboarding(t, db, "UC01")
	svc := newTestOffboardingService(db)

	emp, err := createActiveEmployee(db, orgID, "孙八", "13800008006")
	require.NoError(t, err)

	bossReq := &BossResignRequest{
		ResignationDate: "2026-04-30",
		Reason:          "合同到期",
	}
	err = svc.BossResign(orgID, 1, emp.ID, bossReq)
	require.NoError(t, err)

	obRepo := NewOffboardingRepository(db)
	ob, err := obRepo.FindByEmployeeID(orgID, emp.ID)
	require.NoError(t, err)

	// 更新交接清单
	newItems := []ChecklistCategory{
		{
			Category: "资产归还",
			Items: []ChecklistItem{
				{Name: "笔记本电脑", Completed: true},
				{Name: "门禁卡", Completed: false},
			},
		},
	}
	itemsJSON, err := json.Marshal(newItems)
	require.NoError(t, err)

	err = svc.UpdateChecklist(orgID, ob.ID, itemsJSON)
	require.NoError(t, err)

	ob2, err := obRepo.FindByID(orgID, ob.ID)
	require.NoError(t, err)

	var categories []ChecklistCategory
	err = json.Unmarshal(ob2.ChecklistItems, &categories)
	require.NoError(t, err)
	require.Len(t, categories, 1)
	assert.Equal(t, "资产归还", categories[0].Category)
	require.Len(t, categories[0].Items, 2)
	assert.True(t, categories[0].Items[0].Completed)
	assert.False(t, categories[0].Items[1].Completed)
}

// TestDefaultChecklistItems 默认交接清单包含 3 个分类
func TestDefaultChecklistItems(t *testing.T) {
	items := defaultChecklistItems()
	require.NotNil(t, items)

	var categories []ChecklistCategory
	err := json.Unmarshal(items, &categories)
	require.NoError(t, err)

	assert.Len(t, categories, 3, "默认交接清单应包含 3 个分类")

	categoryNames := make([]string, len(categories))
	for i, cat := range categories {
		categoryNames[i] = cat.Category
		assert.NotEmpty(t, cat.Items, "每个分类应至少有 1 个条目")
		for _, item := range cat.Items {
			assert.False(t, item.Completed, "默认条目应为未完成状态")
		}
	}
	assert.Contains(t, categoryNames, "资产归还")
	assert.Contains(t, categoryNames, "工作交接")
	assert.Contains(t, categoryNames, "权限回收")
}

// TestGetOffboarding 返回完整离职详情含 employee_name
func TestGetOffboarding(t *testing.T) {
	db := setupOffboardingTestDB(t)
	orgID := createTestOrgForOffboarding(t, db, "GO01")
	svc := newTestOffboardingService(db)

	emp, err := createActiveEmployee(db, orgID, "周九", "13800008007")
	require.NoError(t, err)

	bossReq := &BossResignRequest{
		ResignationDate: "2026-04-30",
		Reason:          "协商离职",
	}
	err = svc.BossResign(orgID, 1, emp.ID, bossReq)
	require.NoError(t, err)

	obRepo := NewOffboardingRepository(db)
	ob, err := obRepo.FindByEmployeeID(orgID, emp.ID)
	require.NoError(t, err)

	// 获取详情
	detail, err := svc.GetOffboarding(orgID, ob.ID)
	require.NoError(t, err)
	assert.Equal(t, ob.ID, detail.ID)
	assert.Equal(t, emp.ID, detail.EmployeeID)
	assert.Equal(t, "周九", detail.EmployeeName)
	assert.Equal(t, "involuntary", detail.Type)
	assert.Equal(t, "pending", detail.Status)
	assert.Equal(t, "协商离职", detail.Reason)
	require.NotNil(t, detail.ChecklistItems)
}
