package salary

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// AdjustmentService 调薪业务逻辑层
type AdjustmentService struct {
	repo *AdjustmentRepository
}

// NewAdjustmentService 创建调薪 Service
func NewAdjustmentService(repo *AdjustmentRepository) *AdjustmentService {
	return &AdjustmentService{repo: repo}
}

// CreateAdjustment 单人调薪
func (s *AdjustmentService) CreateAdjustment(orgID, userID int64, req *AdjustmentRequest) error {
	if err := validateEffectiveMonth(req.EffectiveMonth); err != nil {
		return err
	}

	oldValue := decimal.NewFromFloat(req.OldValue)
	newValue := decimal.NewFromFloat(req.NewValue)

	if req.AdjustBy == AdjustmentByRatio {
		// ratio 模式：NewValue 是百分比，从 OldValue 换算绝对值
		ratio := decimal.NewFromFloat(req.NewValue).Div(decimal.NewFromInt(100))
		newValue = oldValue.Mul(ratio)
	}

	adj := &SalaryAdjustment{
		OrgID:          orgID,
		EmployeeID:     &req.EmployeeID,
		Type:           AdjustmentTypeIndividual,
		EffectiveMonth: req.EffectiveMonth,
		AdjustmentType: req.AdjustmentType,
		AdjustBy:       req.AdjustBy,
		OldValue:       oldValue.Round(2).InexactFloat64(),
		NewValue:       newValue.Round(2).InexactFloat64(),
		Status:         "active",
		CreatedBy:      userID,
	}

	return s.repo.Create(adj)
}

// CreateMassAdjustment 部门普调
func (s *AdjustmentService) CreateMassAdjustment(orgID, userID int64, req *MassAdjustmentRequest) error {
	if err := validateEffectiveMonth(req.EffectiveMonth); err != nil {
		return err
	}

	// 获取部门下所有员工
	employeeIDs, err := s.repo.GetEmployeeIDsByDepartments(orgID, req.DepartmentIDs)
	if err != nil {
		return fmt.Errorf("获取部门员工失败: %w", err)
	}

	if len(employeeIDs) == 0 {
		return fmt.Errorf("所选部门下没有在职员工")
	}

	oldValue := decimal.NewFromFloat(req.OldValue)
	newValue := decimal.NewFromFloat(req.NewValue)

	if req.AdjustBy == AdjustmentByRatio {
		ratio := decimal.NewFromFloat(req.NewValue).Div(decimal.NewFromInt(100))
		newValue = oldValue.Mul(ratio)
	}

	// 为每个部门创建一条普调记录
	for _, deptID := range req.DepartmentIDs {
		adj := &SalaryAdjustment{
			OrgID:          orgID,
			DepartmentID:   &deptID,
			Type:           AdjustmentTypeDepartment,
			EffectiveMonth: req.EffectiveMonth,
			AdjustmentType: req.AdjustmentType,
			AdjustBy:       req.AdjustBy,
			OldValue:       oldValue.Round(2).InexactFloat64(),
			NewValue:       newValue.Round(2).InexactFloat64(),
			Status:         "active",
			CreatedBy:      userID,
		}
		if err := s.repo.Create(adj); err != nil {
			return fmt.Errorf("创建部门 %d 调薪记录失败: %w", deptID, err)
		}
	}

	return nil
}

// Preview 调薪预览
func (s *AdjustmentService) Preview(orgID int64, req *AdjustmentPreviewRequest) (*AdjustmentPreviewResponse, error) {
	var employeeCount int64
	var err error

	if len(req.DepartmentIDs) > 0 {
		employeeCount, err = s.repo.CountEmployeesByDepartments(orgID, req.DepartmentIDs)
		if err != nil {
			return nil, fmt.Errorf("统计部门员工失败: %w", err)
		}
	} else {
		employeeCount = int64(len(req.EmployeeIDs))
	}

	if employeeCount == 0 {
		return &AdjustmentPreviewResponse{
			EmployeeCount: 0,
			MonthlyImpact: 0,
			AnnualImpact:  0,
		}, nil
	}

	oldValue := decimal.NewFromFloat(req.OldValue)
	newValue := decimal.NewFromFloat(req.NewValue)

	if req.AdjustBy == AdjustmentByRatio {
		ratio := decimal.NewFromFloat(req.NewValue).Div(decimal.NewFromInt(100))
		newValue = oldValue.Mul(ratio)
	}

	delta := newValue.Sub(oldValue)
	monthlyImpact := delta.Mul(decimal.NewFromInt(employeeCount))
	annualImpact := monthlyImpact.Mul(decimal.NewFromInt(12))

	return &AdjustmentPreviewResponse{
		EmployeeCount: int(employeeCount),
		MonthlyImpact: monthlyImpact.Round(2).InexactFloat64(),
		AnnualImpact:  annualImpact.Round(2).InexactFloat64(),
	}, nil
}

// GetAdjustmentList 调薪记录列表
func (s *AdjustmentService) GetAdjustmentList(orgID int64, effectiveMonth string, page, pageSize int) ([]AdjustmentListResponse, int64, error) {
	records, total, err := s.repo.List(orgID, effectiveMonth, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	result := make([]AdjustmentListResponse, 0, len(records))
	for _, r := range records {
		result = append(result, AdjustmentListResponse{
			ID:             r.ID,
			EmployeeID:     r.EmployeeID,
			DepartmentID:   r.DepartmentID,
			Type:           r.Type,
			EffectiveMonth: r.EffectiveMonth,
			AdjustmentType: r.AdjustmentType,
			AdjustBy:       r.AdjustBy,
			OldValue:       r.OldValue,
			NewValue:       r.NewValue,
			Status:         r.Status,
			CreatedBy:      r.CreatedBy,
			CreatedAt:      r.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return result, total, nil
}

// validateEffectiveMonth 校验生效月份 >= 当前月
func validateEffectiveMonth(effectiveMonth string) error {
	now := time.Now()
	currentMonth := fmt.Sprintf("%d-%02d", now.Year(), now.Month())
	if effectiveMonth < currentMonth {
		return fmt.Errorf("生效月份不能早于当前月份")
	}
	return nil
}
