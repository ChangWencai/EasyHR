package employee

import (
	"fmt"
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

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Discard,
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Organization{}, &model.User{}, &Employee{}))
	return db
}

func createTestOrg(t *testing.T, db *gorm.DB, suffix string) int64 {
	org := &model.Organization{
		Name:       "测试企业" + suffix,
		CreditCode: "91110108MA0" + suffix,
		City:       "北京",
		Status:     "active",
	}
	require.NoError(t, db.Create(org).Error)
	return org.ID
}

// createEmployee 辅助函数：创建测试用员工
func createEmployee(orgID int64, name, phone, position, status string) *Employee {
	emp := &Employee{}
	emp.OrgID = orgID
	emp.Name = name
	emp.PhoneEncrypted = "enc_" + phone
	emp.PhoneHash = crypto.HashSHA256(phone)
	emp.IDCardEncrypted = "enc_idcard_" + phone
	emp.IDCardHash = crypto.HashSHA256("idcard_" + phone)
	emp.Position = position
	emp.Status = status
	emp.HireDate = time.Now()
	return emp
}

// TestCreateEmployee 创建员工后验证基本字段
func TestCreateEmployee(t *testing.T) {
	db := setupTestDB(t)
	orgID := createTestOrg(t, db, "TEST1")
	repo := NewRepository(db)

	hireDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.Local)
	emp := createEmployee(orgID, "张三", "13800008000", "前端开发", StatusActive)
	emp.HireDate = hireDate

	err := repo.Create(emp)
	require.NoError(t, err)
	assert.Greater(t, emp.ID, int64(0))
	assert.Equal(t, orgID, emp.OrgID)
}

// TestDuplicatePhoneHash 同 org_id 下重复 phone_hash 返回错误
func TestDuplicatePhoneHash(t *testing.T) {
	db := setupTestDB(t)
	orgID := createTestOrg(t, db, "TEST2")
	repo := NewRepository(db)

	phone := "13800008000"
	emp1 := createEmployee(orgID, "张三", phone, "前端开发", StatusActive)
	require.NoError(t, repo.Create(emp1))

	emp2 := createEmployee(orgID, "李四", phone, "后端开发", StatusActive)
	err := repo.Create(emp2)
	assert.Error(t, err, "同 org_id 下重复 phone_hash 应返回错误")
}

// TestCrossOrgSamePhone 不同 org_id 下相同 phone_hash 可正常创建
func TestCrossOrgSamePhone(t *testing.T) {
	db := setupTestDB(t)
	orgID1 := createTestOrg(t, db, "ORG1")
	orgID2 := createTestOrg(t, db, "ORG2")
	repo := NewRepository(db)

	phone := "13800008000"
	emp1 := createEmployee(orgID1, "张三", phone, "前端开发", StatusActive)
	require.NoError(t, repo.Create(emp1))

	emp2 := createEmployee(orgID2, "李四", phone, "后端开发", StatusActive)
	err := repo.Create(emp2)
	assert.NoError(t, err, "不同 org_id 下相同 phone_hash 应可创建")
}

// TestSearchByName 搜索姓名 ILIKE 模糊匹配
func TestSearchByName(t *testing.T) {
	db := setupTestDB(t)
	orgID := createTestOrg(t, db, "TEST3")
	repo := NewRepository(db)

	names := []string{"张三", "张四", "李五"}
	for i, name := range names {
		phone := fmt.Sprintf("1380000800%d", i)
		emp := createEmployee(orgID, name, phone, "开发", StatusActive)
		require.NoError(t, repo.Create(emp))
	}

	// 搜索 "张" 应返回 2 条
	results, total, err := repo.List(orgID, SearchParams{Name: "张"}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, results, 2)

	// 搜索 "李" 应返回 1 条
	results, total, err = repo.List(orgID, SearchParams{Name: "李"}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, results, 1)

	// 搜索不存在的名字
	results, total, err = repo.List(orgID, SearchParams{Name: "王"}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, results, 0)
}

// TestSearchByPhoneHash 手机号通过 hash 精确匹配
func TestSearchByPhoneHash(t *testing.T) {
	db := setupTestDB(t)
	orgID := createTestOrg(t, db, "TEST4")
	repo := NewRepository(db)

	phone := "13800008000"
	emp := createEmployee(orgID, "张三", phone, "开发", StatusActive)
	require.NoError(t, repo.Create(emp))

	// 通过明文手机号搜索
	results, total, err := repo.List(orgID, SearchParams{Phone: phone}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, results, 1)
	assert.Equal(t, "张三", results[0].Name)

	// 搜索不存在的手机号
	results, total, err = repo.List(orgID, SearchParams{Phone: "13900009000"}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
}

// TestSearchByStatus 按状态筛选
func TestSearchByStatus(t *testing.T) {
	db := setupTestDB(t)
	orgID := createTestOrg(t, db, "TEST5")
	repo := NewRepository(db)

	activeEmp := createEmployee(orgID, "张三", "13800008001", "开发", StatusActive)
	require.NoError(t, repo.Create(activeEmp))

	pendingEmp := createEmployee(orgID, "李四", "13800008002", "测试", StatusPending)
	require.NoError(t, repo.Create(pendingEmp))

	// 筛选 active
	results, total, err := repo.List(orgID, SearchParams{Status: StatusActive}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, results, 1)
	assert.Equal(t, "张三", results[0].Name)

	// 筛选 pending
	results, total, err = repo.List(orgID, SearchParams{Status: StatusPending}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, results, 1)
	assert.Equal(t, "李四", results[0].Name)
}

// TestPagination 分页查询返回正确的 total 和数据量
func TestPagination(t *testing.T) {
	db := setupTestDB(t)
	orgID := createTestOrg(t, db, "TEST6")
	repo := NewRepository(db)

	for i := 0; i < 15; i++ {
		phone := fmt.Sprintf("13800008%03d", i)
		emp := createEmployee(orgID, "员工", phone, "开发", StatusActive)
		require.NoError(t, repo.Create(emp))
	}

	// 第1页，每页10条
	results, total, err := repo.List(orgID, SearchParams{}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(15), total)
	assert.Len(t, results, 10)

	// 第2页，每页10条
	results, total, err = repo.List(orgID, SearchParams{}, 2, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(15), total)
	assert.Len(t, results, 5)
}

// TestSoftDelete 软删除后查询不到
func TestSoftDelete(t *testing.T) {
	db := setupTestDB(t)
	orgID := createTestOrg(t, db, "TEST7")
	repo := NewRepository(db)

	emp := createEmployee(orgID, "张三", "13800008000", "开发", StatusActive)
	require.NoError(t, repo.Create(emp))

	// 删除
	err := repo.Delete(orgID, emp.ID)
	require.NoError(t, err)

	// 查询不到
	_, err = repo.FindByID(orgID, emp.ID)
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)

	// 列表也不出现
	results, total, err := repo.List(orgID, SearchParams{}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, results, 0)
}

// TestExtractFromIDCard 验证身份证提取 gender/birthDate 正确
func TestExtractFromIDCard(t *testing.T) {
	// 男性身份证号（第17位奇数）
	gender, birthDate, err := extractFromIDCard("110108199001011234")
	require.NoError(t, err)
	assert.Equal(t, "男", gender)
	expectedDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedDate.Year(), birthDate.Year())
	assert.Equal(t, expectedDate.Month(), birthDate.Month())
	assert.Equal(t, expectedDate.Day(), birthDate.Day())

	// 女性身份证号（第17位偶数）
	gender, birthDate, err = extractFromIDCard("110108199503152345")
	require.NoError(t, err)
	assert.Equal(t, "女", gender)
	expectedDate = time.Date(1995, 3, 15, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedDate.Year(), birthDate.Year())
	assert.Equal(t, expectedDate.Month(), birthDate.Month())
	assert.Equal(t, expectedDate.Day(), birthDate.Day())

	// 无效身份证号
	_, _, err = extractFromIDCard("12345")
	assert.Error(t, err)
}
