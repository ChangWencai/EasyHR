package salary

import (
	"fmt"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// SickLeavePolicyService 病假系数服务
type SickLeavePolicyService struct {
	db *gorm.DB
}

// NewSickLeavePolicyService 创建病假系数服务
func NewSickLeavePolicyService(db *gorm.DB) *SickLeavePolicyService {
	return &SickLeavePolicyService{db: db}
}

// GetSickLeaveCoefficient 根据城市和工龄月数获取病假系数
func (s *SickLeavePolicyService) GetSickLeaveCoefficient(city string, tenureMonths int) decimal.Decimal {
	bucket := TenureBucketOver6Months
	if tenureMonths < 6 {
		bucket = TenureBucketWithin6Months
	}

	var policy SickLeavePolicy
	err := s.db.Where("city = ? AND tenure_bucket = ?", city, bucket).First(&policy).Error
	if err != nil {
		// 未找到策略，返回默认系数 1.0
		return decimal.NewFromInt(1)
	}

	return decimal.NewFromFloat(policy.Coefficient)
}

// SeedInitialPolicies 初始化一线城市病假系数策略
func (s *SickLeavePolicyService) SeedInitialPolicies() error {
	policies := []SickLeavePolicy{
		// 北京
		{City: "北京", TenureBucket: TenureBucketWithin6Months, Coefficient: 0.60},
		{City: "北京", TenureBucket: TenureBucketOver6Months, Coefficient: 0.40},
		// 上海
		{City: "上海", TenureBucket: TenureBucketWithin6Months, Coefficient: 0.60},
		{City: "上海", TenureBucket: TenureBucketOver6Months, Coefficient: 0.40},
		// 广州
		{City: "广州", TenureBucket: TenureBucketWithin6Months, Coefficient: 0.60},
		{City: "广州", TenureBucket: TenureBucketOver6Months, Coefficient: 0.40},
		// 深圳
		{City: "深圳", TenureBucket: TenureBucketWithin6Months, Coefficient: 0.60},
		{City: "深圳", TenureBucket: TenureBucketOver6Months, Coefficient: 0.40},
	}

	for _, p := range policies {
		var count int64
		s.db.Model(&SickLeavePolicy{}).
			Where("city = ? AND tenure_bucket = ?", p.City, p.TenureBucket).
			Count(&count)
		if count == 0 {
			if err := s.db.Create(&p).Error; err != nil {
				return fmt.Errorf("seed sick leave policy %s/%s: %w", p.City, p.TenureBucket, err)
			}
		}
	}

	return nil
}
