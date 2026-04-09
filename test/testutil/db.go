package testutil

import (
	"time"

	"github.com/wencai/easyhr/internal/audit"
	"github.com/wencai/easyhr/internal/common/config"
	"github.com/wencai/easyhr/internal/common/model"
	"github.com/wencai/easyhr/internal/employee"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SetupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(
		&model.Organization{},
		&model.User{},
		&audit.AuditLog{},
		&employee.Employee{},
	); err != nil {
		return nil, err
	}
	return db, nil
}

func CreateTestOrg(db *gorm.DB, name, creditCode, city string) (*model.Organization, error) {
	org := &model.Organization{
		Name:       name,
		CreditCode: creditCode,
		City:       city,
		Status:     "active",
	}
	if err := db.Create(org).Error; err != nil {
		return nil, err
	}
	return org, nil
}

func CreateTestUser(db *gorm.DB, orgID int64, name, phoneHash, role string) (*model.User, error) {
	user := &model.User{
		Phone:     phoneHash,
		PhoneHash: phoneHash,
		Name:      name,
		Role:      role,
		Status:    "active",
	}
	user.OrgID = orgID
	if err := db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func CleanupTestDB(db *gorm.DB) error {
	return db.Exec("DELETE FROM audit_logs; DELETE FROM users; DELETE FROM organizations;").Error
}

// CreateTestEmployee 创建测试用员工记录
func CreateTestEmployee(db *gorm.DB, orgID int64, name, phoneEncrypted, phoneHash, position, status string) (*employee.Employee, error) {
	emp := &employee.Employee{}
	emp.OrgID = orgID
	emp.Name = name
	emp.PhoneEncrypted = phoneEncrypted
	emp.PhoneHash = phoneHash
	emp.Position = position
	emp.Status = status
	emp.HireDate = time.Now()
	if err := db.Create(emp).Error; err != nil {
		return nil, err
	}
	return emp, nil
}

func WaitForDB(db *gorm.DB) error {
	for i := 0; i < 10; i++ {
		sqlDB, err := db.DB()
		if err == nil && sqlDB != nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

// TestCryptoConfig 返回测试用加密配置
func TestCryptoConfig() config.CryptoConfig {
	return config.CryptoConfig{
		AESKey: "32-byte-long-key-for-testing-12345678",
	}
}
