package attendance

import (
	"time"

	"gorm.io/gorm"
)

// AttendanceRepository 数据访问层
type AttendanceRepository struct {
	db *gorm.DB
}

func NewAttendanceRepository(db *gorm.DB) *AttendanceRepository {
	return &AttendanceRepository{db: db}
}

// orgScope 返回当前企业的查询作用域
func (r *AttendanceRepository) orgScope(orgID int64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB { return db.Where("org_id = ?", orgID) }
}

// --- AttendanceRule ---

func (r *AttendanceRepository) GetRule(orgID int64) (*AttendanceRule, error) {
	var rule AttendanceRule
	err := r.db.Scopes(r.orgScope(orgID)).First(&rule).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &rule, err
}

func (r *AttendanceRepository) UpsertRule(orgID int64, rule *AttendanceRule) error {
	existing, err := r.GetRule(orgID)
	if err != nil {
		return err
	}
	if existing != nil {
		rule.ID = existing.ID
		rule.OrgID = orgID
		return r.db.Save(rule).Error
	}
	rule.OrgID = orgID
	return r.db.Create(rule).Error
}

// --- Shift ---

func (r *AttendanceRepository) ListShifts(orgID int64) ([]Shift, error) {
	var shifts []Shift
	err := r.db.Scopes(r.orgScope(orgID)).Order("id ASC").Find(&shifts).Error
	return shifts, err
}

func (r *AttendanceRepository) GetShift(orgID int64, id int64) (*Shift, error) {
	var shift Shift
	err := r.db.Scopes(r.orgScope(orgID)).First(&shift, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &shift, err
}

func (r *AttendanceRepository) CreateShift(shift *Shift) error {
	return r.db.Create(shift).Error
}

func (r *AttendanceRepository) UpdateShift(shift *Shift) error {
	return r.db.Save(shift).Error
}

func (r *AttendanceRepository) DeleteShift(orgID int64, id int64) error {
	return r.db.Scopes(r.orgScope(orgID)).Delete(&Shift{}, id).Error
}

// --- Schedule ---

func (r *AttendanceRepository) ListSchedules(orgID int64, startDate, endDate string, employeeID *int64) ([]Schedule, error) {
	query := r.db.Scopes(r.orgScope(orgID)).Where("date >= ? AND date <= ?", startDate, endDate)
	if employeeID != nil {
		query = query.Where("employee_id = ?", *employeeID)
	}
	var schedules []Schedule
	err := query.Order("date ASC").Find(&schedules).Error
	return schedules, err
}

func (r *AttendanceRepository) BatchUpsertSchedules(orgID int64, schedules []Schedule) error {
	if len(schedules) == 0 {
		return nil
	}
	for i := range schedules {
		schedules[i].OrgID = orgID
	}
	return r.db.Save(&schedules).Error
}

// --- ClockRecord ---

// ListClockRecordsByDate 获取指定日期的打卡记录
func (r *AttendanceRepository) ListClockRecordsByDate(orgID int64, workDate string) ([]ClockRecord, error) {
	var records []ClockRecord
	err := r.db.Scopes(r.orgScope(orgID)).
		Where("work_date = ?", workDate).
		Order("clock_time ASC").
		Find(&records).Error
	return records, err
}

// GetClockRecordsByEmployee 获取员工指定日期范围的打卡记录
func (r *AttendanceRepository) GetClockRecordsByEmployee(orgID int64, employeeID int64, startDate, endDate string) ([]ClockRecord, error) {
	var records []ClockRecord
	err := r.db.Scopes(r.orgScope(orgID)).
		Where("employee_id = ? AND work_date >= ? AND work_date <= ?", employeeID, startDate, endDate).
		Order("work_date ASC, clock_time ASC").
		Find(&records).Error
	return records, err
}

// CreateClockRecord 创建打卡记录
func (r *AttendanceRepository) CreateClockRecord(record *ClockRecord) error {
	return r.db.Create(record).Error
}

// BatchCreateClockRecords 批量创建打卡记录
func (r *AttendanceRepository) BatchCreateClockRecords(records []ClockRecord) error {
	if len(records) == 0 {
		return nil
	}
	return r.db.Create(&records).Error
}

// GetClockRecordByEmployeeDateType 查询员工某日某类型的打卡记录
func (r *AttendanceRepository) GetClockRecordByEmployeeDateType(orgID int64, employeeID int64, workDate, clockType string) (*ClockRecord, error) {
	var record ClockRecord
	err := r.db.Scopes(r.orgScope(orgID)).
		Where("employee_id = ? AND work_date = ? AND clock_type = ?", employeeID, workDate, clockType).
		First(&record).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &record, err
}

// --- Employee JOIN 查询 ---

// EmployeeBrief 员工简要信息（用于 JOIN 查询）
type EmployeeBrief struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	DepartmentID   *int64 `json:"department_id"`
	DepartmentName string `json:"department_name"`
}

// ListEmployeesByIDs 批量查询员工信息（JOIN departments 获取部门名称）
func (r *AttendanceRepository) ListEmployeesByIDs(orgID int64, ids []int64) ([]EmployeeBrief, error) {
	var emps []EmployeeBrief
	if len(ids) == 0 {
		return emps, nil
	}
	err := r.db.Table("employees").
		Select("employees.id, employees.name, employees.department_id, COALESCE(departments.name, '') as department_name").
		Joins("LEFT JOIN departments ON departments.id = employees.department_id AND departments.deleted_at IS NULL").
		Where("employees.org_id = ? AND employees.id IN ? AND employees.deleted_at IS NULL", orgID, ids).
		Find(&emps).Error
	return emps, err
}

// ListAllActiveEmployees 获取所有在职员工（用于打卡实况全员列表）
func (r *AttendanceRepository) ListAllActiveEmployees(orgID int64) ([]EmployeeBrief, error) {
	var emps []EmployeeBrief
	err := r.db.Table("employees").
		Select("employees.id, employees.name, employees.department_id, COALESCE(departments.name, '') as department_name").
		Joins("LEFT JOIN departments ON departments.id = employees.department_id AND departments.deleted_at IS NULL").
		Where("employees.org_id = ? AND employees.status IN ? AND employees.deleted_at IS NULL", orgID, []string{"active", "probation"}).
		Order("employees.name ASC").
		Find(&emps).Error
	return emps, err
}

// --- AttendanceManualStats ---

// GetManualStats 获取员工手动修正的假勤统计
func (r *AttendanceRepository) GetManualStats(orgID int64, employeeID int64, yearMonth string) (*AttendanceManualStats, error) {
	var stats AttendanceManualStats
	err := r.db.Scopes(r.orgScope(orgID)).
		Where("employee_id = ? AND year_month = ?", employeeID, yearMonth).
		First(&stats).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &stats, err
}

// UpsertManualStats 创建或更新手动修正的假勤统计
func (r *AttendanceRepository) UpsertManualStats(stats *AttendanceManualStats) error {
	existing, err := r.GetManualStats(stats.OrgID, stats.EmployeeID, stats.YearMonth)
	if err != nil {
		return err
	}
	if existing != nil {
		stats.ID = existing.ID
		return r.db.Save(stats).Error
	}
	return r.db.Create(stats).Error
}

// --- Approval ---

func (r *AttendanceRepository) CreateApproval(approval *Approval) error {
	return r.db.Create(approval).Error
}

func (r *AttendanceRepository) GetApproval(orgID int64, id int64) (*Approval, error) {
	var approval Approval
	err := r.db.Scopes(r.orgScope(orgID)).First(&approval, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &approval, err
}

func (r *AttendanceRepository) ListApprovals(orgID int64, status, approvalType string, employeeID *int64, page, pageSize int) ([]Approval, int64, error) {
	query := r.db.Scopes(r.orgScope(orgID))
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if approvalType != "" {
		query = query.Where("approval_type = ?", approvalType)
	}
	if employeeID != nil {
		query = query.Where("employee_id = ?", *employeeID)
	}
	var total int64
	query.Model(&Approval{}).Count(&total)
	var approvals []Approval
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&approvals).Error
	return approvals, total, err
}

func (r *AttendanceRepository) ListPendingApprovals(orgID int64, approverID int64) ([]Approval, error) {
	var approvals []Approval
	err := r.db.Scopes(r.orgScope(orgID)).
		Where("status = ?", ApprovalStatusPending).
		Order("created_at ASC").
		Find(&approvals).Error
	return approvals, err
}

func (r *AttendanceRepository) ListApprovalsByEmployeeMonth(orgID int64, employeeID int64, yearMonth string) ([]Approval, error) {
	startDate := yearMonth + "-01"
	endDate := yearMonth + "-31"
	var approvals []Approval
	err := r.db.Scopes(r.orgScope(orgID)).
		Where("employee_id = ? AND status IN (?, ?) AND start_time >= ? AND start_time <= ?",
			employeeID, ApprovalStatusApproved, ApprovalStatusPending,
			startDate, endDate).
		Find(&approvals).Error
	return approvals, err
}

func (r *AttendanceRepository) CountPendingApprovals(orgID int64) (int64, error) {
	var count int64
	err := r.db.Scopes(r.orgScope(orgID)).
		Model(&Approval{}).
		Where("status = ?", ApprovalStatusPending).
		Count(&count).Error
	return count, err
}

func (r *AttendanceRepository) UpdateApproval(approval *Approval) error {
	return r.db.Save(approval).Error
}

// --- AttendanceMonthly ---

func (r *AttendanceRepository) ListAttendanceMonthly(orgID int64, yearMonth string, page, pageSize int) ([]AttendanceMonthly, int64, error) {
	query := r.db.Scopes(r.orgScope(orgID)).Where("year_month = ?", yearMonth)
	var total int64
	query.Model(&AttendanceMonthly{}).Count(&total)
	offset := (page - 1) * pageSize
	var list []AttendanceMonthly
	err := query.Order("employee_id ASC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *AttendanceRepository) UpsertAttendanceMonthly(m *AttendanceMonthly) error {
	return r.db.Save(m).Error
}

func (r *AttendanceRepository) GetDailyClockRecords(orgID int64, employeeID int64, yearMonth string) (map[string]map[string]string, error) {
	startDate := yearMonth + "-01"
	records, err := r.GetClockRecordsByEmployee(orgID, employeeID, startDate, yearMonth+"-31")
	if err != nil {
		return nil, err
	}
	result := make(map[string]map[string]string)
	for _, r := range records {
		dateStr := r.WorkDate.Format("2006-01-02")
		if result[dateStr] == nil {
			result[dateStr] = make(map[string]string)
		}
		result[dateStr][r.ClockType] = r.ClockTime.Format("15:04")
	}
	return result, nil
}

// --- Compliance Report Queries ---

// EmployeeWithDept helper struct for compliance queries
type EmployeeWithDept struct {
	EmployeeID     int64
	EmployeeName   string
	DepartmentName string
}

// ListEmployeesByOrgWithDept lists active employees with their department names, filtered by deptIDs
// Used by: GetComplianceOvertime, GetComplianceLeave, GetComplianceAnomaly
func (r *AttendanceRepository) ListEmployeesByOrgWithDept(orgID int64, deptIDs []int64) ([]EmployeeWithDept, error) {
	type result struct {
		EmployeeID     int64
		EmployeeName   string
		DepartmentName string
	}
	var rows []result
	query := r.db.Table("employees").
		Select("employees.id as employee_id, employees.name as employee_name, departments.name as department_name").
		Joins("LEFT JOIN departments ON departments.id = employees.department_id").
		Where("employees.org_id = ? AND employees.status = ? AND employees.deleted_at IS NULL", orgID, "active")
	if len(deptIDs) > 0 {
		query = query.Where("employees.department_id IN ?", deptIDs)
	}
	err := query.Order("employees.name ASC").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	emps := make([]EmployeeWithDept, len(rows))
	for i, row := range rows {
		emps[i] = EmployeeWithDept{EmployeeID: row.EmployeeID, EmployeeName: row.EmployeeName, DepartmentName: row.DepartmentName}
	}
	return emps, nil
}

// ListClockRecordsByMonth returns all clock records for the given org+employees in a year-month
// Used by: GetComplianceOvertime (for weekday/weekend classification), GetComplianceAnomaly
func (r *AttendanceRepository) ListClockRecordsByMonth(orgID int64, employeeIDs []int64, yearMonth string) ([]ClockRecord, error) {
	if len(employeeIDs) == 0 {
		return nil, nil
	}
	var startTime, endTime time.Time
	parsed, err := time.Parse("2006-01", yearMonth)
	if err != nil {
		return nil, err
	}
	startTime = time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.UTC)
	endTime = startTime.AddDate(0, 1, 0)
	var records []ClockRecord
	err = r.db.Where("org_id = ? AND employee_id IN ? AND clock_time >= ? AND clock_time < ?",
		orgID, employeeIDs, startTime, endTime).Find(&records).Error
	return records, err
}

// ListApprovalsByMonth returns approved approvals for given employees in a year-month
// Used by: GetComplianceOvertime (overtime hours from overtime approvals),
//           GetComplianceLeave (annual/sick/personal leave days from leave approvals)
func (r *AttendanceRepository) ListApprovalsByMonth(orgID int64, employeeIDs []int64, yearMonth string) ([]Approval, error) {
	if len(employeeIDs) == 0 {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01", yearMonth)
	if err != nil {
		return nil, err
	}
	startTime := time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.UTC)
	endTime := startTime.AddDate(0, 1, 0)
	var approvals []Approval
	err = r.db.Where("org_id = ? AND employee_id IN ? AND status = ? AND start_time >= ? AND start_time < ?",
		orgID, employeeIDs, ApprovalStatusApproved, startTime, endTime).Find(&approvals).Error
	return approvals, err
}

// GetAnnualLeaveQuotas returns annual leave quota records for given employees in a given year
func (r *AttendanceRepository) GetAnnualLeaveQuotas(orgID int64, employeeIDs []int64, year int) (map[int64]float64, error) {
	if len(employeeIDs) == 0 {
		return nil, nil
	}
	var quotas []AnnualLeaveQuota
	err := r.db.Scopes(r.orgScope(orgID)).
		Where("employee_id IN ? AND year = ?", employeeIDs, year).
		Find(&quotas).Error
	if err != nil {
		return nil, err
	}
	result := make(map[int64]float64, len(quotas))
	for _, q := range quotas {
		result[q.EmployeeID] = q.Quota
	}
	return result, nil
}

// ListMonthlyAttendanceForCompliance returns monthly attendance summaries for given employees + year_month
// Used by: GetComplianceAnomaly (for late_count, early_leave_count, absent_days)
//          GetComplianceMonthly (full compliance row)
func (r *AttendanceRepository) ListMonthlyAttendanceForCompliance(orgID int64, employeeIDs []int64, yearMonth string) ([]AttendanceMonthly, error) {
	if len(employeeIDs) == 0 {
		return nil, nil
	}
	var records []AttendanceMonthly
	err := r.db.Where("org_id = ? AND employee_id IN ? AND year_month = ?", orgID, employeeIDs, yearMonth).Find(&records).Error
	return records, err
}
