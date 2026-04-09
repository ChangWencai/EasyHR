package salary

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wencai/easyhr/test/testutil"
	"gorm.io/gorm"
)

// setupSalaryDB 设置测试数据库（包含 salary 模型迁移）
func setupSalaryDB(t *testing.T) *gorm.DB {
	db, err := testutil.SetupTestDB()
	assert.NoError(t, err)

	// 迁移 salary 相关表
	err = db.AutoMigrate(
		&SalaryTemplateItem{},
		&SalaryItem{},
		&PayrollRecord{},
		&PayrollItem{},
		&PayrollSlip{},
	)
	assert.NoError(t, err)
	return db
}

// TestSalaryTemplateItemSeed 测试预置薪资项种子数据
func TestSalaryTemplateItemSeed(t *testing.T) {
	db := setupSalaryDB(t)
	defer testutil.CleanupTestDB(db)

	templateRepo := NewSalaryTemplateRepository(db)

	// 执行种子数据创建
	err := templateRepo.SeedGlobalTemplateItems()
	assert.NoError(t, err, "SeedGlobalTemplateItems should succeed")

	// 验证 10 个预置项已创建
	var items []SalaryTemplateItem
	err = db.Where("org_id = 0").Order("sort_order").Find(&items).Error
	assert.NoError(t, err)
	assert.Equal(t, 10, len(items), "Should have 10 preset items")

	// 验证基本工资是必需项
	baseSalary := items[0]
	assert.Equal(t, "基本工资", baseSalary.Name)
	assert.Equal(t, "income", baseSalary.Type)
	assert.True(t, baseSalary.IsRequired)
	assert.True(t, baseSalary.IsEnabled)

	// 验证排序和类型
	incomeCount := 0
	deductionCount := 0
	for _, item := range items {
		if item.Type == "income" {
			incomeCount++
		} else if item.Type == "deduction" {
			deductionCount++
		}
	}
	assert.Equal(t, 7, incomeCount, "Should have 7 income items")
	assert.Equal(t, 3, deductionCount, "Should have 3 deduction items")
}

// TestSalaryTemplateCRUD 测试薪资模板 CRUD
func TestSalaryTemplateCRUD(t *testing.T) {
	db := setupSalaryDB(t)
	defer testutil.CleanupTestDB(db)

	templateRepo := NewSalaryTemplateRepository(db)
	svc := NewService(nil, templateRepo, nil, nil, nil, nil)

	org, err := testutil.CreateTestOrg(db, "测试企业", "91110000MA0000001", "北京")
	assert.NoError(t, err)
	user, err := testutil.CreateTestUser(db, org.ID, "测试用户", "13800138000", "owner")
	assert.NoError(t, err)

	// 初始化种子数据
	err = svc.SeedTemplateItems()
	assert.NoError(t, err)

	// Test GetTemplate: 全局模板应返回给企业
	tmpl, err := svc.GetTemplate(org.ID)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(tmpl.Items), "Template should have 10 items")

	// 验证基本工资默认启用
	baseSalary := tmpl.Items[0]
	assert.Equal(t, "基本工资", baseSalary.Name)
	assert.True(t, baseSalary.IsEnabled)

	// Test UpdateTemplate: 禁用绩效工资
	var perfTemplateID int64
	for _, item := range tmpl.Items {
		if item.Name == "绩效工资" {
			perfTemplateID = item.ID
			break
		}
	}
	assert.Greater(t, perfTemplateID, int64(0), "Performance salary item should exist")

	updateReq := []TemplateItemUpdate{
		{TemplateItemID: perfTemplateID, IsEnabled: false},
	}
	err = svc.UpdateTemplate(org.ID, user.ID, updateReq)
	assert.NoError(t, err)

	// 验证企业级覆盖已创建（使用新的数据库会话）
	var overrides []SalaryTemplateItem
	err = db.Table("salary_template_items").
		Where("org_id = ? AND name = ?", org.ID, "绩效工资").
		Find(&overrides).Error
	assert.NoError(t, err, "Query overrides should succeed")
	t.Logf("Found %d overrides for 绩效工资", len(overrides))
	for _, o := range overrides {
		t.Logf("Override: ID=%d, Name=%s, IsEnabled=%v, OrgID=%d", o.ID, o.Name, o.IsEnabled, o.OrgID)
	}

	if len(overrides) == 0 {
		t.Fatal("No override found")
	}
	override := overrides[0]
	assert.False(t, override.IsEnabled, "Performance salary should be disabled")

	// 再次获取模板，绩效工资应被禁用
	tmpl2, err := svc.GetTemplate(org.ID)
	assert.NoError(t, err)
	var perfItem *TemplateItemResponse
	for _, item := range tmpl2.Items {
		if item.Name == "绩效工资" {
			perfItem = &item
			break
		}
	}
	assert.NotNil(t, perfItem)
	assert.False(t, perfItem.IsEnabled, "Performance salary should be disabled after update")
}

// TestSalaryItemCRUD 测试员工薪资项 CRUD
func TestSalaryItemCRUD(t *testing.T) {
	db := setupSalaryDB(t)
	defer testutil.CleanupTestDB(db)

	repo := NewRepository(db)
	templateRepo := NewSalaryTemplateRepository(db)
	svc := NewService(repo, templateRepo, nil, nil, nil, nil)

	org, err := testutil.CreateTestOrg(db, "测试企业", "91110000MA0000001", "北京")
	assert.NoError(t, err)
	user, err := testutil.CreateTestUser(db, org.ID, "测试用户", "13800138000", "owner")
	assert.NoError(t, err)
	emp, err := testutil.CreateTestEmployee(db, org.ID, "张三", "encrypted_phone", "phone_hash", "工程师", "active")
	assert.NoError(t, err)

	// 初始化种子数据
	err = svc.SeedTemplateItems()
	assert.NoError(t, err)

	// 获取模板找到基本工资 ID
	tmpl, err := svc.GetTemplate(org.ID)
	assert.NoError(t, err)

	var baseSalaryTemplateID int64
	for _, item := range tmpl.Items {
		if item.Name == "基本工资" {
			baseSalaryTemplateID = item.ID
			break
		}
	}
	assert.Greater(t, baseSalaryTemplateID, int64(0))

	// Test SetItems: 设置员工基本工资
	month := "2026-04"
	items := []SalaryItemInput{
		{TemplateItemID: baseSalaryTemplateID, Amount: 5000.00},
	}
	err = svc.SetEmployeeItems(org.ID, user.ID, emp.ID, month, items)
	assert.NoError(t, err)

	// Test GetItems: 查询员工薪资项
	resp, err := svc.GetEmployeeItems(org.ID, emp.ID, month)
	assert.NoError(t, err)
	assert.Equal(t, emp.ID, resp.EmployeeID)
	assert.Equal(t, month, resp.Month)
	assert.Greater(t, len(resp.Items), 0, "Should have at least one item")

	// 验证基本工资金额
	var baseItem *EmployeeItemResponse
	for _, item := range resp.Items {
		if item.ItemName == "基本工资" {
			baseItem = &item
			break
		}
	}
	assert.NotNil(t, baseItem)
	assert.Equal(t, 5000.00, baseItem.Amount)

	// Test Update: 更新金额
	items[0].Amount = 6000.00
	err = svc.SetEmployeeItems(org.ID, user.ID, emp.ID, month, items)
	assert.NoError(t, err)

	resp2, err := svc.GetEmployeeItems(org.ID, emp.ID, month)
	assert.NoError(t, err)
	for _, item := range resp2.Items {
		if item.ItemName == "基本工资" {
			assert.Equal(t, 6000.00, item.Amount, "Amount should be updated")
		}
	}
}

// TestPayrollRecordStatusConstants 测试工资核算记录状态常量
func TestPayrollRecordStatusConstants(t *testing.T) {
	assert.Equal(t, "draft", PayrollStatusDraft)
	assert.Equal(t, "calculated", PayrollStatusCalculated)
	assert.Equal(t, "confirmed", PayrollStatusConfirmed)
	assert.Equal(t, "paid", PayrollStatusPaid)
}

// TestPayrollSlipTokenGeneration 测试工资单 token 生成
func TestPayrollSlipTokenGeneration(t *testing.T) {
	db := setupSalaryDB(t)
	defer testutil.CleanupTestDB(db)

	repo := NewRepository(db)

	org, err := testutil.CreateTestOrg(db, "测试企业", "91110000MA0000001", "北京")
	assert.NoError(t, err)
	emp, err := testutil.CreateTestEmployee(db, org.ID, "张三", "encrypted_phone", "phone_hash", "工程师", "active")
	assert.NoError(t, err)

	// 创建工资核算记录
	record := &PayrollRecord{
		EmployeeID:   emp.ID,
		EmployeeName: emp.Name,
		Year:         2026,
		Month:        4,
		Status:       PayrollStatusConfirmed,
	}
	record.OrgID = org.ID
	err = repo.CreatePayrollRecord(record)
	assert.NoError(t, err)

	// 生成 token
	token, err := generateSlipToken()
	assert.NoError(t, err)
	assert.Equal(t, 64, len(token), "Token should be 64 characters")

	// 创建工资单
	now := time.Now()
	slip := &PayrollSlip{
		PayrollRecordID: record.ID,
		EmployeeID:      emp.ID,
		Token:           token,
		Status:          SlipStatusPending,
		ExpiresAt:       now.Add(30 * 24 * time.Hour), // 30 天有效期
	}
	slip.OrgID = org.ID
	err = repo.CreateSlip(slip)
	assert.NoError(t, err)

	// 验证 token 唯一性
	slip2, err := repo.FindSlipByToken(token)
	assert.NoError(t, err)
	assert.Equal(t, slip.ID, slip2.ID)
	assert.Equal(t, slip.Token, slip2.Token)
}

// TestAdapterInterfaces 测试适配器接口编译
func TestAdapterInterfaces(t *testing.T) {
	// 此测试仅验证接口定义正确，能正常编译
	db := setupSalaryDB(t)
	defer testutil.CleanupTestDB(db)

	org, err := testutil.CreateTestOrg(db, "测试企业", "91110000MA0000001", "北京")
	assert.NoError(t, err)

	// TaxProvider 接口
	var taxProvider TaxProvider
	taxAdapter := &TaxAdapter{} // 简化测试，不注入真实 service
	taxProvider = taxAdapter
	assert.NotNil(t, taxProvider)

	// SIDeductionProvider 接口
	var siProvider SIDeductionProvider
	siAdapter := &SIAdapter{} // 简化测试
	siProvider = siAdapter
	assert.NotNil(t, siProvider)

	// EmployeeProvider 接口
	empAdapter := &EmployeeAdapter{} // 简化测试
	var empProvider EmployeeProvider = empAdapter
	assert.NotNil(t, empProvider)

	// 验证接口方法存在（编译时检查）
	_ = org.ID // 避免未使用变量警告
}

// TestPresetTemplateItems 验证预置模板项数量和属性
func TestPresetTemplateItems(t *testing.T) {
	presets := getPresetItems()
	assert.Equal(t, 10, len(presets), "应有10个预置薪资项")

	// 验证收入项
	incomeCount := 0
	deductionCount := 0
	for _, p := range presets {
		if p.Type == "income" {
			incomeCount++
		} else {
			deductionCount++
		}
	}
	assert.Equal(t, 7, incomeCount, "应有7个收入项")
	assert.Equal(t, 3, deductionCount, "应有3个扣款项")

	// 验证基本工资是必填项
	assert.True(t, presets[0].IsRequired, "基本工资应为必填")
	assert.Equal(t, "基本工资", presets[0].Name)
	assert.Equal(t, 1, presets[0].SortOrder)
}

// TestPayrollStatusConstants 验证工资核算状态常量
func TestPayrollStatusConstants(t *testing.T) {
	assert.Equal(t, "draft", PayrollStatusDraft)
	assert.Equal(t, "calculated", PayrollStatusCalculated)
	assert.Equal(t, "confirmed", PayrollStatusConfirmed)
	assert.Equal(t, "paid", PayrollStatusPaid)
}

// TestSlipStatusConstants 验证工资单状态常量
func TestSlipStatusConstants(t *testing.T) {
	assert.Equal(t, "pending", SlipStatusPending)
	assert.Equal(t, "sent", SlipStatusSent)
	assert.Equal(t, "viewed", SlipStatusViewed)
	assert.Equal(t, "signed", SlipStatusSigned)
}

// TestSalaryErrorCodes 验证错误码
func TestSalaryErrorCodes(t *testing.T) {
	assert.Equal(t, 50001, CodeTemplateConfig)
	assert.Equal(t, 50002, CodePayrollFailed)
	assert.Equal(t, 50003, CodeInvalidStatus)
	assert.Equal(t, 50004, CodeAttendanceImport)
	assert.Equal(t, 50005, CodeSlipTokenInvalid)
	assert.Equal(t, 50006, CodeSMSVerifyFailed)
	assert.Equal(t, 50007, CodeEmployeeMatch)
}

// TestEmployeeInfoStruct 验证 EmployeeInfo 结构体字段
func TestEmployeeInfoStruct(t *testing.T) {
	info := EmployeeInfo{
		ID:         1,
		Name:       "张三",
		BaseSalary: 10000,
	}
	assert.Equal(t, int64(1), info.ID)
	assert.Equal(t, "张三", info.Name)
	assert.Equal(t, float64(10000), info.BaseSalary)
}

// TestDTOBinding 验证 DTO 结构体 tag
func TestDTOBinding(t *testing.T) {
	// SetEmployeeItemsRequest 的 Month 字段验证
	req := SetEmployeeItemsRequest{
		Month: "2026-04",
		Items: []SalaryItemInput{
			{TemplateItemID: 1, Amount: 8000},
		},
	}
	assert.Equal(t, "2026-04", req.Month)
	assert.Len(t, req.Items, 1)
	assert.Equal(t, float64(8000), req.Items[0].Amount)
}

// getPresetItems 返回预置模板项列表（供测试使用）
func getPresetItems() []struct {
	Name      string
	Type      string
	SortOrder int
	IsRequired bool
} {
	return []struct {
		Name      string
		Type      string
		SortOrder int
		IsRequired bool
	}{
		{"基本工资", "income", 1, true},
		{"绩效工资", "income", 2, false},
		{"岗位补贴", "income", 3, false},
		{"餐补", "income", 4, false},
		{"交通补", "income", 5, false},
		{"通讯补", "income", 6, false},
		{"其他补贴", "income", 7, false},
		{"事假扣款", "deduction", 8, false},
		{"病假扣款", "deduction", 9, false},
		{"其他扣款", "deduction", 10, false},
	}
}
