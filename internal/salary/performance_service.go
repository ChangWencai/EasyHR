package salary

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// PerformanceService 绩效系数业务逻辑层
type PerformanceService struct {
	repo *PerformanceRepository
}

// NewPerformanceService 创建绩效系数 Service
func NewPerformanceService(repo *PerformanceRepository) *PerformanceService {
	return &PerformanceService{repo: repo}
}

// SetCoefficient 设置绩效系数（upsert）
func (s *PerformanceService) SetCoefficient(orgID, userID int64, year, month int, coefficients []CoefficientInput) error {
	for _, c := range coefficients {
		d := decimal.NewFromFloat(c.Coefficient)
		minVal := decimal.NewFromFloat(0.0)
		maxVal := decimal.NewFromFloat(1.0)

		if d.LessThan(minVal) || d.GreaterThan(maxVal) {
			return fmt.Errorf("员工 %d 绩效系数 %.4f 超出范围 [0.0, 1.0]", c.EmployeeID, c.Coefficient)
		}

		if err := s.repo.Upsert(orgID, c.EmployeeID, userID, year, month, c.Coefficient); err != nil {
			return fmt.Errorf("设置员工 %d 绩效系数失败: %w", c.EmployeeID, err)
		}
	}
	return nil
}

// GetCoefficients 获取某月所有员工绩效系数
func (s *PerformanceService) GetCoefficients(orgID int64, year, month int) ([]PerformanceCoefficientResponse, error) {
	records, err := s.repo.ListByMonth(orgID, year, month)
	if err != nil {
		return nil, fmt.Errorf("查询绩效系数失败: %w", err)
	}

	result := make([]PerformanceCoefficientResponse, 0, len(records))
	for _, r := range records {
		result = append(result, PerformanceCoefficientResponse{
			EmployeeID:  r.EmployeeID,
			Coefficient: r.Coefficient,
		})
	}

	return result, nil
}

// GetCoefficientForEmployee 获取员工某月绩效系数，无记录则返回 1.0
func (s *PerformanceService) GetCoefficientForEmployee(orgID, employeeID int64, year, month int) decimal.Decimal {
	record, err := s.repo.FindByEmployee(orgID, employeeID, year, month)
	if err != nil {
		return decimal.NewFromInt(1) // 默认系数 1.0
	}
	return decimal.NewFromFloat(record.Coefficient)
}

// CoefficientInput 绩效系数输入
type CoefficientInput struct {
	EmployeeID  int64   `json:"employee_id" binding:"required"`
	Coefficient float64 `json:"coefficient" binding:"required"`
}

// PerformanceCoefficientResponse 绩效系数响应
type PerformanceCoefficientResponse struct {
	EmployeeID  int64   `json:"employee_id"`
	Coefficient float64 `json:"coefficient"`
}
