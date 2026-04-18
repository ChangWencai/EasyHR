package attendance

import (
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
