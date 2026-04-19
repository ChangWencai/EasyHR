package employee

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wencai/easyhr/internal/common/config"
	"github.com/wencai/easyhr/internal/common/crypto"
	"github.com/wencai/easyhr/internal/common/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupContractTestDB 创建合同测试数据库
func setupContractTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Discard,
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Organization{}, &Employee{}, &Contract{}))
	return db
}

// mockTodoCreator is a no-op TodoCreator for testing
type mockContractTodoCreator struct{}

func (m *mockContractTodoCreator) CreateTodoFromEmployee(orgID int64, title string, todoType string, employeeID *int64, employeeName string, deadline *time.Time, sourceType string, sourceID *int64) error {
	return nil
}

func (m *mockContractTodoCreator) ExistsBySource(ctx context.Context, orgID int64, sourceType string, sourceID *int64) (bool, error) {
	return false, nil
}

// setupContractTestService 创建合同测试 Service
func setupContractTestService(t *testing.T, db *gorm.DB) *ContractService {
	contractRepo := NewContractRepository(db)
	empRepo := NewRepository(db)
	cfg := config.CryptoConfig{AESKey: testAESKey}
	return NewContractService(contractRepo, empRepo, db, cfg, &mockContractTodoCreator{})
}

// createContractTestOrg 创建测试企业
func createContractTestOrg(t *testing.T, db *gorm.DB) int64 {
	org := &model.Organization{
		Name:       "测试合同企业",
		CreditCode: "91110108MA0CONTRACT",
		City:       "北京",
		Status:     "active",
	}
	require.NoError(t, db.Create(org).Error)
	return org.ID
}

// createContractTestEmployee 创建测试用员工（用于合同关联）
func createContractTestEmployee(t *testing.T, db *gorm.DB, orgID int64, name, phone string) *Employee {
	aesKey := []byte(testAESKey)
	phoneEncrypted, _ := crypto.Encrypt(phone, aesKey)
	phoneHash := crypto.HashSHA256(phone)
	idCard := "110108199001011234"
	idCardEncrypted, _ := crypto.Encrypt(idCard, aesKey)
	idCardHash := crypto.HashSHA256(idCard)

	emp := &Employee{}
	emp.OrgID = orgID
	emp.CreatedBy = 1
	emp.UpdatedBy = 1
	emp.Name = name
	emp.PhoneEncrypted = phoneEncrypted
	emp.PhoneHash = phoneHash
	emp.IDCardEncrypted = idCardEncrypted
	emp.IDCardHash = idCardHash
	emp.Position = "开发工程师"
	emp.HireDate = time.Now()
	emp.Status = StatusActive
	require.NoError(t, db.Create(emp).Error)
	return emp
}

// TestContractCreate_Draft 创建合同 status=draft
func TestContractCreate_Draft(t *testing.T) {
	db := setupContractTestDB(t)
	orgID := createContractTestOrg(t, db)
	emp := createContractTestEmployee(t, db, orgID, "张三", "13800008000")
	svc := setupContractTestService(t, db)

	req := &CreateContractRequest{
		ContractType: ContractTypeFixedTerm,
		StartDate:    "2026-01-01",
		EndDate:      "2027-12-31",
		Salary:       15000.00,
	}

	resp, err := svc.CreateContract(nil, orgID, 1, emp.ID, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Greater(t, resp.ID, int64(0))
	assert.Equal(t, emp.ID, resp.EmployeeID)
	assert.Equal(t, ContractStatusDraft, resp.Status)
	assert.Equal(t, ContractTypeFixedTerm, resp.ContractType)
	assert.Equal(t, 15000.00, resp.Salary)
}

// TestContractGeneratePDF_ValidPDF 返回 []byte 以 %PDF- 开头
func TestContractGeneratePDF_ValidPDF(t *testing.T) {
	db := setupContractTestDB(t)
	orgID := createContractTestOrg(t, db)
	emp := createContractTestEmployee(t, db, orgID, "张三", "13800008000")
	svc := setupContractTestService(t, db)

	// 先创建合同
	req := &CreateContractRequest{
		ContractType: ContractTypeFixedTerm,
		StartDate:    "2026-01-01",
		EndDate:      "2027-12-31",
		Salary:       15000.00,
	}
	contractResp, err := svc.CreateContract(nil, orgID, 1, emp.ID, req)
	require.NoError(t, err)

	// 生成 PDF
	pdfBytes, err := svc.GeneratePDF(nil, orgID, contractResp.ID)
	require.NoError(t, err)
	require.NotEmpty(t, pdfBytes)

	// 验证 PDF 头部
	assert.True(t, len(pdfBytes) > 5, "PDF 应有内容")
	assert.Equal(t, "%PDF-", string(pdfBytes[:5]), "PDF 应以 %PDF- 开头")

	// 验证合同状态变为 pending_sign
	updated, _ := svc.GetContract(nil, orgID, contractResp.ID)
	assert.Equal(t, ContractStatusPendingSign, updated.Status)
}

// TestContractUploadSigned_StatusToSigned signed_pdf_url 更新，status=signed 或 active
func TestContractUploadSigned_StatusToSigned(t *testing.T) {
	db := setupContractTestDB(t)
	orgID := createContractTestOrg(t, db)
	emp := createContractTestEmployee(t, db, orgID, "张三", "13800008000")
	svc := setupContractTestService(t, db)

	// 创建合同（结束日期在未来）
	req := &CreateContractRequest{
		ContractType: ContractTypeFixedTerm,
		StartDate:    "2026-01-01",
		EndDate:      "2028-12-31",
		Salary:       15000.00,
	}
	contractResp, err := svc.CreateContract(nil, orgID, 1, emp.ID, req)
	require.NoError(t, err)

	// 先生成 PDF（状态变为 pending_sign）
	_, err = svc.GeneratePDF(nil, orgID, contractResp.ID)
	require.NoError(t, err)

	// 上传签署扫描件
	uploadReq := &UploadSignedRequest{
		SignedPDFURL: "https://oss.example.com/contracts/signed_001.pdf",
		SignDate:     "2026-01-15",
	}
	resp, err := svc.UploadSigned(nil, orgID, contractResp.ID, uploadReq)
	require.NoError(t, err)
	assert.Equal(t, "https://oss.example.com/contracts/signed_001.pdf", resp.SignedPDFURL)
	require.NotNil(t, resp.SignDate)
	assert.Equal(t, "2026-01-15", resp.SignDate.Format("2006-01-02"))
	// 签署日期在合同期限内，状态应为 active
	assert.Equal(t, ContractStatusActive, resp.Status)
}

// TestContractUploadSigned_SignedPastEndDate 签署日期超过结束日期状态为 signed
func TestContractUploadSigned_SignedPastEndDate(t *testing.T) {
	db := setupContractTestDB(t)
	orgID := createContractTestOrg(t, db)
	emp := createContractTestEmployee(t, db, orgID, "李四", "13800008001")
	svc := setupContractTestService(t, db)

	// 创建合同（结束日期在过去）
	req := &CreateContractRequest{
		ContractType: ContractTypeFixedTerm,
		StartDate:    "2024-01-01",
		EndDate:      "2025-12-31",
		Salary:       12000.00,
	}
	contractResp, err := svc.CreateContract(nil, orgID, 1, emp.ID, req)
	require.NoError(t, err)

	// 上传签署扫描件（签署日期超过结束日期）
	uploadReq := &UploadSignedRequest{
		SignedPDFURL: "https://oss.example.com/contracts/signed_002.pdf",
		SignDate:     "2026-01-15", // 超过 2025-12-31 结束日期
	}
	resp, err := svc.UploadSigned(nil, orgID, contractResp.ID, uploadReq)
	require.NoError(t, err)
	// 签署日期在结束日期之后，状态应为 signed（非 active）
	assert.Equal(t, ContractStatusSigned, resp.Status)
}

// TestContractTerminate status=terminated，记录终止日期和原因
func TestContractTerminate(t *testing.T) {
	db := setupContractTestDB(t)
	orgID := createContractTestOrg(t, db)
	emp := createContractTestEmployee(t, db, orgID, "张三", "13800008000")
	svc := setupContractTestService(t, db)

	// 创建合同
	req := &CreateContractRequest{
		ContractType: ContractTypeFixedTerm,
		StartDate:    "2026-01-01",
		EndDate:      "2027-12-31",
		Salary:       15000.00,
	}
	contractResp, err := svc.CreateContract(nil, orgID, 1, emp.ID, req)
	require.NoError(t, err)

	// 终止合同
	termReq := &TerminateContractRequest{
		TerminateDate:   "2026-06-30",
		TerminateReason: "员工主动离职",
	}
	resp, err := svc.TerminateContract(nil, orgID, contractResp.ID, termReq)
	require.NoError(t, err)
	assert.Equal(t, ContractStatusTerminated, resp.Status)
	require.NotNil(t, resp.TerminateDate)
	assert.Equal(t, "2026-06-30", resp.TerminateDate.Format("2006-01-02"))
	assert.Equal(t, "员工主动离职", resp.TerminateReason)
}

// TestContractListByEmployee 返回指定员工的合同列表
func TestContractListByEmployee(t *testing.T) {
	db := setupContractTestDB(t)
	orgID := createContractTestOrg(t, db)
	emp1 := createContractTestEmployee(t, db, orgID, "张三", "13800008000")
	emp2 := createContractTestEmployee(t, db, orgID, "李四", "13800008001")
	svc := setupContractTestService(t, db)

	// emp1 创建 2 份合同
	_, err := svc.CreateContract(nil, orgID, 1, emp1.ID, &CreateContractRequest{
		ContractType: ContractTypeFixedTerm,
		StartDate:    "2024-01-01",
		EndDate:      "2025-12-31",
		Salary:       12000.00,
	})
	require.NoError(t, err)

	_, err = svc.CreateContract(nil, orgID, 1, emp1.ID, &CreateContractRequest{
		ContractType: ContractTypeIndefinite,
		StartDate:    "2026-01-01",
		Salary:       15000.00,
	})
	require.NoError(t, err)

	// emp2 创建 1 份合同
	_, err = svc.CreateContract(nil, orgID, 1, emp2.ID, &CreateContractRequest{
		ContractType: ContractTypeFixedTerm,
		StartDate:    "2026-01-01",
		EndDate:      "2027-12-31",
		Salary:       13000.00,
	})
	require.NoError(t, err)

	// 查询 emp1 的合同
	contracts, total, err := svc.ListByEmployee(nil, orgID, emp1.ID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, contracts, 2)

	// 查询 emp2 的合同
	contracts, total, err = svc.ListByEmployee(nil, orgID, emp2.ID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, contracts, 1)
}

// TestContractMultiplePerEmployee 同一员工可创建多份合同
func TestContractMultiplePerEmployee(t *testing.T) {
	db := setupContractTestDB(t)
	orgID := createContractTestOrg(t, db)
	emp := createContractTestEmployee(t, db, orgID, "张三", "13800008000")
	svc := setupContractTestService(t, db)

	// 创建第一份合同（固定期限）
	resp1, err := svc.CreateContract(nil, orgID, 1, emp.ID, &CreateContractRequest{
		ContractType: ContractTypeFixedTerm,
		StartDate:    "2024-01-01",
		EndDate:      "2025-12-31",
		Salary:       12000.00,
	})
	require.NoError(t, err)
	assert.Greater(t, resp1.ID, int64(0))

	// 创建第二份合同（续签，无固定期限）
	resp2, err := svc.CreateContract(nil, orgID, 1, emp.ID, &CreateContractRequest{
		ContractType: ContractTypeIndefinite,
		StartDate:    "2026-01-01",
		Salary:       15000.00,
	})
	require.NoError(t, err)
	assert.Greater(t, resp2.ID, int64(0))
	assert.NotEqual(t, resp1.ID, resp2.ID, "两份合同应有不同 ID")

	// 创建第三份合同（实习）
	resp3, err := svc.CreateContract(nil, orgID, 1, emp.ID, &CreateContractRequest{
		ContractType: ContractTypeIntern,
		StartDate:    "2026-03-01",
		EndDate:      "2026-08-31",
		Salary:       3000.00,
	})
	require.NoError(t, err)
	assert.Greater(t, resp3.ID, int64(0))

	// 验证同一员工有 3 份合同
	contracts, total, err := svc.ListByEmployee(nil, orgID, emp.ID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, contracts, 3)
}

// TestContractIndefinite 无固定期限合同 EndDate 为 nil
func TestContractIndefinite(t *testing.T) {
	db := setupContractTestDB(t)
	orgID := createContractTestOrg(t, db)
	emp := createContractTestEmployee(t, db, orgID, "王五", "13800008002")
	svc := setupContractTestService(t, db)

	req := &CreateContractRequest{
		ContractType: ContractTypeIndefinite,
		StartDate:    "2026-01-01",
		Salary:       20000.00,
	}
	resp, err := svc.CreateContract(nil, orgID, 1, emp.ID, req)
	require.NoError(t, err)
	assert.Nil(t, resp.EndDate, "无固定期限合同 EndDate 应为 nil")
	assert.Equal(t, ContractTypeIndefinite, resp.ContractType)
}

// TestContractStatusFlow 完整合同状态流转
func TestContractStatusFlow(t *testing.T) {
	db := setupContractTestDB(t)
	orgID := createContractTestOrg(t, db)
	emp := createContractTestEmployee(t, db, orgID, "赵六", "13800008003")
	svc := setupContractTestService(t, db)

	// 1. 创建合同 -> draft
	resp, err := svc.CreateContract(nil, orgID, 1, emp.ID, &CreateContractRequest{
		ContractType: ContractTypeFixedTerm,
		StartDate:    "2026-01-01",
		EndDate:      "2028-12-31",
		Salary:       15000.00,
	})
	require.NoError(t, err)
	assert.Equal(t, ContractStatusDraft, resp.Status)

	// 2. 生成 PDF -> pending_sign
	_, err = svc.GeneratePDF(nil, orgID, resp.ID)
	require.NoError(t, err)
	resp, _ = svc.GetContract(nil, orgID, resp.ID)
	assert.Equal(t, ContractStatusPendingSign, resp.Status)

	// 3. 上传签署件 -> active
	_, err = svc.UploadSigned(nil, orgID, resp.ID, &UploadSignedRequest{
		SignedPDFURL: "https://oss.example.com/signed.pdf",
		SignDate:     "2026-01-15",
	})
	require.NoError(t, err)
	resp, _ = svc.GetContract(nil, orgID, resp.ID)
	assert.Equal(t, ContractStatusActive, resp.Status)

	// 4. 终止合同 -> terminated
	_, err = svc.TerminateContract(nil, orgID, resp.ID, &TerminateContractRequest{
		TerminateDate:   "2026-06-30",
		TerminateReason: "协商解除",
	})
	require.NoError(t, err)
	resp, _ = svc.GetContract(nil, orgID, resp.ID)
	assert.Equal(t, ContractStatusTerminated, resp.Status)
}
