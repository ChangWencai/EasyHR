package salary

import (
	"fmt"
)

// SalaryListFilter 薪资列表筛选条件
type SalaryListFilter struct {
	OrgID        int64
	Year         int
	Month        int
	DepartmentID *int64
	Keyword      string
	Page         int
	PageSize     int
}

// SalaryListRecord 薪资列表返回记录（含部门名称）
type SalaryListRecord struct {
	ID                int64   `json:"id"`
	EmployeeID        int64   `json:"employee_id"`
	EmployeeName      string  `json:"employee_name"`
	DepartmentName    string  `json:"department_name"`
	GrossIncome       float64 `json:"gross_income"`
	TotalDeductions   float64 `json:"total_deductions"`
	Tax               float64 `json:"tax"`
	SIDeduction       float64 `json:"si_deduction"`
	NetIncome         float64 `json:"net_income"`
	Status            string  `json:"status"`
}

// ListSalaryRecords 查询薪资列表（支持部门/姓名筛选）
func (s *Service) ListSalaryRecords(filter SalaryListFilter) ([]SalaryListRecord, int64, error) {
	var records []SalaryListRecord
	var total int64

	// 先通过员工表 Join 获取部门名称
	type joinedRecord struct {
		ID              int64   `gorm:"column:id"`
		EmployeeID      int64   `gorm:"column:employee_id"`
		EmployeeName    string  `gorm:"column:employee_name"`
		DepartmentName  string  `gorm:"column:department_name"`
		GrossIncome     float64 `gorm:"column:gross_income"`
		TotalDeductions float64 `gorm:"column:total_deductions"`
		Tax             float64 `gorm:"column:tax"`
		SIDeduction     float64 `gorm:"column:si_deduction"`
		NetIncome       float64 `gorm:"column:net_income"`
		Status          string  `gorm:"column:status"`
	}

	q := s.repo.db.Table("payroll_records AS pr").
		Select(`pr.id, pr.employee_id, pr.employee_name,
			COALESCE(d.name, '-') AS department_name,
			pr.gross_income, pr.total_deductions, pr.tax, pr.si_deduction, pr.net_income, pr.status`).
		Joins("LEFT JOIN employees AS e ON e.id = pr.employee_id").
		Joins("LEFT JOIN departments AS d ON d.id = e.department_id").
		Where("pr.org_id = ? AND pr.year = ? AND pr.month = ?", filter.OrgID, filter.Year, filter.Month)

	if filter.DepartmentID != nil {
		q = q.Where("e.department_id = ?", *filter.DepartmentID)
	}

	if filter.Keyword != "" {
		q = q.Where("pr.employee_name LIKE ?", "%"+filter.Keyword+"%")
	}

	// Count
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count salary records: %w", err)
	}

	// Paginate
	offset := (filter.Page - 1) * filter.PageSize
	var rows []joinedRecord
	if err := q.Offset(offset).Limit(filter.PageSize).Order("pr.employee_name").Find(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("list salary records: %w", err)
	}

	for _, r := range rows {
		records = append(records, SalaryListRecord{
			ID:              r.ID,
			EmployeeID:      r.EmployeeID,
			EmployeeName:    r.EmployeeName,
			DepartmentName:  r.DepartmentName,
			GrossIncome:     r.GrossIncome,
			TotalDeductions: r.TotalDeductions,
			Tax:             r.Tax,
			SIDeduction:     r.SIDeduction,
			NetIncome:       r.NetIncome,
			Status:          r.Status,
		})
	}

	return records, total, nil
}

// GetSalaryListExportRecords 查询薪资导出数据（含明细列）
func (s *Service) GetSalaryListExportRecords(orgID int64, year, month int) ([]PayrollRecordWithItems, error) {
	records, err := s.repo.FindPayrollRecordsByMonth(orgID, year, month)
	if err != nil {
		return nil, fmt.Errorf("query payroll records: %w", err)
	}

	var result []PayrollRecordWithItems
	for _, record := range records {
		items, err := s.repo.FindPayrollItemsByRecord(orgID, record.ID)
		if err != nil {
			continue
		}
		result = append(result, PayrollRecordWithItems{
			Record: record,
			Items:  items,
		})
	}

	return result, nil
}

// Ensure Repository has ListPayrollRecordsByFilter
func (r *Repository) ListPayrollRecordsByFilter(filter SalaryListFilter) ([]SalaryListRecord, int64, error) {
	var records []SalaryListRecord
	var total int64

	type joinedRecord struct {
		ID              int64   `gorm:"column:id"`
		EmployeeID      int64   `gorm:"column:employee_id"`
		EmployeeName    string  `gorm:"column:employee_name"`
		DepartmentName  string  `gorm:"column:department_name"`
		GrossIncome     float64 `gorm:"column:gross_income"`
		TotalDeductions float64 `gorm:"column:total_deductions"`
		Tax             float64 `gorm:"column:tax"`
		SIDeduction     float64 `gorm:"column:si_deduction"`
		NetIncome       float64 `gorm:"column:net_income"`
		Status          string  `gorm:"column:status"`
	}

	q := r.db.Table("payroll_records AS pr").
		Select(`pr.id, pr.employee_id, pr.employee_name,
			COALESCE(d.name, '-') AS department_name,
			pr.gross_income, pr.total_deductions, pr.tax, pr.si_deduction, pr.net_income, pr.status`).
		Joins("LEFT JOIN employees AS e ON e.id = pr.employee_id").
		Joins("LEFT JOIN departments AS d ON d.id = e.department_id").
		Where("pr.org_id = ? AND pr.year = ? AND pr.month = ?", filter.OrgID, filter.Year, filter.Month)

	if filter.DepartmentID != nil {
		q = q.Where("e.department_id = ?", *filter.DepartmentID)
	}
	if filter.Keyword != "" {
		q = q.Where("pr.employee_name LIKE ?", "%"+filter.Keyword+"%")
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	var rows []joinedRecord
	if err := q.Offset(offset).Limit(filter.PageSize).Order("pr.employee_name").Find(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("find: %w", err)
	}

	for _, r := range rows {
		records = append(records, SalaryListRecord{
			ID:              r.ID,
			EmployeeID:      r.EmployeeID,
			EmployeeName:    r.EmployeeName,
			DepartmentName:  r.DepartmentName,
			GrossIncome:     r.GrossIncome,
			TotalDeductions: r.TotalDeductions,
			Tax:             r.Tax,
			SIDeduction:     r.SIDeduction,
			NetIncome:       r.NetIncome,
			Status:          r.Status,
		})
	}

	return records, total, nil
}
