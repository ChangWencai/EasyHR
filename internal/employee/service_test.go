package employee

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wencai/easyhr/internal/common/config"
	"github.com/wencai/easyhr/internal/common/crypto"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 测试用 AES 密钥（32字节）
const testAESKey = "01234567890123456789012345678901"

func setupServiceTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Discard,
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Employee{}))
	return db
}

func newTestService(db *gorm.DB) *Service {
	repo := NewRepository(db)
	cfg := config.CryptoConfig{AESKey: testAESKey}
	return NewService(repo, cfg)
}

// createTestEmployeeViaService 通过 Service 创建测试员工
func createTestEmployeeViaService(s *Service, orgID, userID int64, name, phone, idCard, position, hireDate string) (*EmployeeResponse, error) {
	req := &CreateEmployeeRequest{
		Name:     name,
		Phone:    phone,
		IDCard:   idCard,
		Position: position,
		HireDate: hireDate,
	}
	return s.CreateEmployee(orgID, userID, req)
}

// TestCreateEmployee_Success 验证创建后加密存储正确
func TestCreateEmployee_Success(t *testing.T) {
	db := setupServiceTestDB(t)
	svc := newTestService(db)

	resp, err := createTestEmployeeViaService(svc, 1, 1, "张三", "13800008000", "110108199001011234", "前端开发", "2026-01-15")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Greater(t, resp.ID, int64(0))
	assert.Equal(t, "张三", resp.Name)
	assert.Equal(t, "138****8000", resp.Phone) // 脱敏手机号
	assert.Equal(t, "110108****1234", resp.IDCard) // 脱敏身份证
	assert.Equal(t, "男", resp.Gender)
	assert.Equal(t, "pending", resp.Status)

	// 验证数据库中存储了加密值
	repo := NewRepository(db)
	emp, err := repo.FindByID(1, resp.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, emp.PhoneEncrypted)
	assert.Equal(t, crypto.HashSHA256("13800008000"), emp.PhoneHash)
	assert.NotEmpty(t, emp.IDCardEncrypted)
	assert.Equal(t, crypto.HashSHA256("110108199001011234"), emp.IDCardHash)
}

// TestCreateEmployee_DuplicatePhone 重复手机号返回错误
func TestCreateEmployee_DuplicatePhone(t *testing.T) {
	db := setupServiceTestDB(t)
	svc := newTestService(db)

	_, err := createTestEmployeeViaService(svc, 1, 1, "张三", "13800008000", "110108199001011234", "前端开发", "2026-01-15")
	require.NoError(t, err)

	_, err = createTestEmployeeViaService(svc, 1, 1, "李四", "13800008000", "110108199503152345", "后端开发", "2026-02-01")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "该手机号已存在")
}

// TestCreateEmployee_ExtractIDCard 验证从身份证号自动提取性别和出生日期
func TestCreateEmployee_ExtractIDCard(t *testing.T) {
	db := setupServiceTestDB(t)
	svc := newTestService(db)

	// 男性身份证号（第17位奇数）
	resp, err := createTestEmployeeViaService(svc, 1, 1, "张三", "13800008000", "110108199001011234", "前端开发", "2026-01-15")
	require.NoError(t, err)
	assert.Equal(t, "男", resp.Gender)
	require.NotNil(t, resp.BirthDate)
	assert.Equal(t, 1990, resp.BirthDate.Year())
	assert.Equal(t, time.January, resp.BirthDate.Month())
	assert.Equal(t, 1, resp.BirthDate.Day())

	// 女性身份证号（第17位偶数）
	resp2, err := createTestEmployeeViaService(svc, 2, 1, "李四", "13900009000", "110108199503152345", "后端开发", "2026-02-01")
	require.NoError(t, err)
	assert.Equal(t, "女", resp2.Gender)
	require.NotNil(t, resp2.BirthDate)
	assert.Equal(t, 1995, resp2.BirthDate.Year())
	assert.Equal(t, time.March, resp2.BirthDate.Month())
	assert.Equal(t, 15, resp2.BirthDate.Day())
}

// TestListEmployees_MaskedData 验证列表返回脱敏数据
func TestListEmployees_MaskedData(t *testing.T) {
	db := setupServiceTestDB(t)
	svc := newTestService(db)

	// 创建测试企业和员工
	_, err := createTestEmployeeViaService(svc, 1, 1, "张三", "13800008000", "110108199001011234", "前端开发", "2026-01-15")
	require.NoError(t, err)

	employees, total, err := svc.ListEmployees(1, ListQueryParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, employees, 1)

	// 验证手机号脱敏格式
	assert.Equal(t, "138****8000", employees[0].Phone)
	// 验证身份证号脱敏格式
	assert.Equal(t, "110108****1234", employees[0].IDCard)
}

// TestExportExcel_ValidXLSX 验证导出返回有效的 xlsx 二进制数据
func TestExportExcel_ValidXLSX(t *testing.T) {
	db := setupServiceTestDB(t)
	svc := newTestService(db)

	// 创建多个员工用于导出
	for i := 0; i < 3; i++ {
		phone := fmt.Sprintf("13800008%03d", i)
		idCard := fmt.Sprintf("11010819900101%04d", i*111+1234)
		_, err := createTestEmployeeViaService(svc, 1, 1, fmt.Sprintf("员工%d", i), phone, idCard, "开发", "2026-01-15")
		require.NoError(t, err)
	}

	data, err := svc.ExportExcel(1, ListQueryParams{})
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// 验证 xlsx 文件头：PK\x03\x04（ZIP 格式）
	assert.Equal(t, byte(0x50), data[0], "xlsx 文件应以 PK 开头")
	assert.Equal(t, byte(0x4B), data[1])
}

// TestGetSensitiveInfo 验证返回完整解密的敏感信息
func TestGetSensitiveInfo(t *testing.T) {
	db := setupServiceTestDB(t)
	svc := newTestService(db)

	resp, err := createTestEmployeeViaService(svc, 1, 1, "张三", "13800008000", "110108199001011234", "前端开发", "2026-01-15")
	require.NoError(t, err)

	info, err := svc.GetSensitiveInfo(1, resp.ID)
	require.NoError(t, err)
	assert.Equal(t, "13800008000", info.Phone) // 完整明文手机号
	assert.Equal(t, "110108199001011234", info.IDCard) // 完整明文身份证号
}

// TestUpdateEmployee 验证更新员工信息
func TestUpdateEmployee(t *testing.T) {
	db := setupServiceTestDB(t)
	svc := newTestService(db)

	resp, err := createTestEmployeeViaService(svc, 1, 1, "张三", "13800008000", "110108199001011234", "前端开发", "2026-01-15")
	require.NoError(t, err)

	newName := "张三四"
	newPosition := "高级前端开发"
	updated, err := svc.UpdateEmployee(1, 1, resp.ID, &UpdateEmployeeRequest{
		Name:     &newName,
		Position: &newPosition,
	})
	require.NoError(t, err)
	assert.Equal(t, "张三四", updated.Name)
	assert.Equal(t, "高级前端开发", updated.Position)
}

// TestDeleteEmployee 验证软删除
func TestDeleteEmployee(t *testing.T) {
	db := setupServiceTestDB(t)
	svc := newTestService(db)

	resp, err := createTestEmployeeViaService(svc, 1, 1, "张三", "13800008000", "110108199001011234", "前端开发", "2026-01-15")
	require.NoError(t, err)

	err = svc.DeleteEmployee(1, resp.ID)
	require.NoError(t, err)

	// 删除后查询应返回错误
	_, err = svc.GetEmployee(1, resp.ID)
	assert.Error(t, err)
}

// TestListEmployees_SearchByName 验证通过姓名搜索
func TestListEmployees_SearchByName(t *testing.T) {
	db := setupServiceTestDB(t)
	svc := newTestService(db)

	_, err := createTestEmployeeViaService(svc, 1, 1, "张三", "13800008001", "110108199001011234", "前端开发", "2026-01-15")
	require.NoError(t, err)
	_, err = createTestEmployeeViaService(svc, 1, 1, "李四", "13800008002", "110108199503152345", "后端开发", "2026-02-01")
	require.NoError(t, err)

	employees, total, err := svc.ListEmployees(1, ListQueryParams{Name: "张", Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, employees, 1)
	assert.Equal(t, "张三", employees[0].Name)
}

// TestMaskBankAccount 验证银行卡号脱敏
func TestMaskBankAccount(t *testing.T) {
	assert.Equal(t, "****5678", maskBankAccount("12345678"))
	assert.Equal(t, "****9012", maskBankAccount("6222000012349012"))
	assert.Equal(t, "1234", maskBankAccount("1234")) // 短号码不脱敏
}
