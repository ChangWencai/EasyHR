package middleware

import "gorm.io/gorm"

func TenantScope(orgID int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("org_id = ?", orgID)
	}
}
