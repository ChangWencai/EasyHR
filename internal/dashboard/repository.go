package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/wencai/easyhr/internal/common/middleware"
	"gorm.io/gorm"
)

// DashboardRepository defines the data access interface for the dashboard.
type DashboardRepository interface {
	GetEmployeeStats(ctx context.Context, orgID int64) (active, joined, left int, err error)
	GetPayrollTotal(ctx context.Context, orgID int64) (string, error)
	GetSocialInsuranceTotal(ctx context.Context, orgID int64) (string, error)
	GetPendingVouchers(ctx context.Context, orgID int64) (int, error)
	GetPendingExpenses(ctx context.Context, orgID int64) (int, error)
	GetTaxReminders(ctx context.Context, orgID int64) (int, error)
	GetContractExpirations(ctx context.Context, orgID int64) (int, error)
	GetPendingOffboardings(ctx context.Context, orgID int64) (int, error)
	GetPendingInvitations(ctx context.Context, orgID int64) (int, error)
	GetTodoRingStats(ctx context.Context, orgID int64) (completed, pending int, err error)
	GetTimeLimitedRingStats(ctx context.Context, orgID int64) (completed, pending int, err error)
}

// DashboardRepositoryImpl is the concrete GORM implementation.
type DashboardRepositoryImpl struct {
	db *gorm.DB
}

// NewRepository creates a new DashboardRepositoryImpl.
func NewRepository(db *gorm.DB) *DashboardRepositoryImpl {
	return &DashboardRepositoryImpl{db: db}
}

// GetEmployeeStats returns active employee count, joined this month, and left this month.
func (r *DashboardRepositoryImpl) GetEmployeeStats(ctx context.Context, orgID int64) (active, joined, left int, err error) {
	now := time.Now()
	year, month, _ := now.Date()
	startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, -1)

	// Active employees (active + probation)
	var activeCount int64
	if err := r.db.Model(&EmployeeRecord{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("status IN ?", []string{StatusActive, StatusProbation}).
		Count(&activeCount).Error; err != nil {
		return 0, 0, 0, fmt.Errorf("count active employees: %w", err)
	}

	// Joined this month
	var joinedCount int64
	if err := r.db.Model(&EmployeeRecord{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("status IN ? AND created_at >= ? AND created_at <= ?",
			[]string{StatusActive, StatusProbation}, startOfMonth, endOfMonth).
		Count(&joinedCount).Error; err != nil {
		return 0, 0, 0, fmt.Errorf("count joined employees: %w", err)
	}

	// Left this month
	var leftCount int64
	if err := r.db.Model(&EmployeeRecord{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("status = ? AND updated_at >= ? AND updated_at <= ?",
			StatusResigned, startOfMonth, endOfMonth).
		Count(&leftCount).Error; err != nil {
		return 0, 0, 0, fmt.Errorf("count left employees: %w", err)
	}

	return int(activeCount), int(joinedCount), int(leftCount), nil
}

// GetPayrollTotal returns the sum of net income for paid payroll records this month.
func (r *DashboardRepositoryImpl) GetPayrollTotal(ctx context.Context, orgID int64) (string, error) {
	now := time.Now()
	year, month := now.Year(), int(now.Month())

	var result struct {
		Total string
	}
	err := r.db.Model(&PayrollRecord{}).
		Scopes(middleware.TenantScope(orgID)).
		Select("COALESCE(SUM(net_income), 0) as total").
		Where("year = ? AND month = ? AND status = ?", year, month, PayrollStatusPaid).
		Scan(&result).Error
	if err != nil {
		return "0", fmt.Errorf("get payroll total: %w", err)
	}
	return result.Total, nil
}

// GetSocialInsuranceTotal returns the sum of social insurance (personal + employer) for this month.
func (r *DashboardRepositoryImpl) GetSocialInsuranceTotal(ctx context.Context, orgID int64) (string, error) {
	now := time.Now()
	paymentMonth := fmt.Sprintf("%d-%02d", now.Year(), now.Month())

	// Check if table exists
	if !r.db.Migrator().HasTable(&SIRecord{}) {
		return "0", nil
	}

	var result struct {
		Total string
	}
	err := r.db.Model(&SIRecord{}).
		Scopes(middleware.TenantScope(orgID)).
		Select("COALESCE(SUM(total_company + total_personal), 0) as total").
		Where("start_month = ?", paymentMonth).
		Scan(&result).Error
	if err != nil {
		return "0", fmt.Errorf("get social insurance total: %w", err)
	}
	return result.Total, nil
}

// GetPendingVouchers returns the count of vouchers with status 'submitted'.
func (r *DashboardRepositoryImpl) GetPendingVouchers(ctx context.Context, orgID int64) (int, error) {
	if !r.db.Migrator().HasTable(&VoucherRecord{}) {
		return 0, nil
	}
	var count int64
	err := r.db.Model(&VoucherRecord{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("status = ?", VoucherStatusSubmitted).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("get pending vouchers: %w", err)
	}
	return int(count), nil
}

// GetPendingExpenses returns the count of expense reimbursements with status 'pending'.
func (r *DashboardRepositoryImpl) GetPendingExpenses(ctx context.Context, orgID int64) (int, error) {
	if !r.db.Migrator().HasTable(&ExpenseRecord{}) {
		return 0, nil
	}
	var count int64
	err := r.db.Model(&ExpenseRecord{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("status = ?", ExpenseStatusPending).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("get pending expenses: %w", err)
	}
	return int(count), nil
}

// GetTaxReminders returns the count of unread, undismissed tax reminders due within 3 days.
func (r *DashboardRepositoryImpl) GetTaxReminders(ctx context.Context, orgID int64) (int, error) {
	if !r.db.Migrator().HasTable(&TaxReminderRecord{}) {
		return 0, nil
	}
	now := time.Now()
	deadline := now.AddDate(0, 0, 3)

	var count int64
	err := r.db.Model(&TaxReminderRecord{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("is_read = ? AND is_dismissed = ? AND due_date IS NOT NULL AND due_date <= ?",
			false, false, deadline).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("get tax reminders: %w", err)
	}
	return int(count), nil
}

// GetContractExpirations returns the count of active contracts expiring within 30 days.
func (r *DashboardRepositoryImpl) GetContractExpirations(ctx context.Context, orgID int64) (int, error) {
	if !r.db.Migrator().HasTable(&ContractRecord{}) {
		return 0, nil
	}
	now := time.Now()
	deadline := now.AddDate(0, 0, 30)

	var count int64
	err := r.db.Model(&ContractRecord{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("status = ? AND end_date IS NOT NULL AND end_date <= ? AND end_date >= ?",
			ContractStatusActive, deadline, now).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("get contract expirations: %w", err)
	}
	return int(count), nil
}

// GetPendingOffboardings returns the count of offboardings with status 'pending' or 'approved'.
func (r *DashboardRepositoryImpl) GetPendingOffboardings(ctx context.Context, orgID int64) (int, error) {
	if !r.db.Migrator().HasTable(&OffboardingRecord{}) {
		return 0, nil
	}
	var count int64
	err := r.db.Model(&OffboardingRecord{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("status IN ?", []string{OffboardingStatusPending, OffboardingStatusApproved}).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("get pending offboardings: %w", err)
	}
	return int(count), nil
}

// GetTodoRingStats returns completed/pending counts for all todos.
// Only counts status IN ('pending','completed'). Skips terminated.
func (r *DashboardRepositoryImpl) GetTodoRingStats(ctx context.Context, orgID int64) (completed, pending int, err error) {
	if !r.db.Migrator().HasTable(&TodoItemRecord{}) {
		return 0, 0, nil
	}

	var completedCount, pendingCount int64

	if err := r.db.Model(&TodoItemRecord{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("status = ?", "completed").
		Count(&completedCount).Error; err != nil {
		return 0, 0, fmt.Errorf("count completed todos: %w", err)
	}

	if err := r.db.Model(&TodoItemRecord{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("status = ?", "pending").
		Count(&pendingCount).Error; err != nil {
		return 0, 0, fmt.Errorf("count pending todos: %w", err)
	}

	return int(completedCount), int(pendingCount), nil
}

// GetTimeLimitedRingStats returns completed/pending counts for time-limited todos only.
func (r *DashboardRepositoryImpl) GetTimeLimitedRingStats(ctx context.Context, orgID int64) (completed, pending int, err error) {
	if !r.db.Migrator().HasTable(&TodoItemRecord{}) {
		return 0, 0, nil
	}

	var completedCount, pendingCount int64

	if err := r.db.Model(&TodoItemRecord{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("is_time_limited = ? AND status = ?", true, "completed").
		Count(&completedCount).Error; err != nil {
		return 0, 0, fmt.Errorf("count completed time-limited todos: %w", err)
	}

	if err := r.db.Model(&TodoItemRecord{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("is_time_limited = ? AND status = ?", true, "pending").
		Count(&pendingCount).Error; err != nil {
		return 0, 0, fmt.Errorf("count pending time-limited todos: %w", err)
	}

	return int(completedCount), int(pendingCount), nil
}

// GetPendingInvitations returns the count of invitations with status 'pending'.
func (r *DashboardRepositoryImpl) GetPendingInvitations(ctx context.Context, orgID int64) (int, error) {
	if !r.db.Migrator().HasTable(&InvitationRecord{}) {
		return 0, nil
	}
	var count int64
	err := r.db.Model(&InvitationRecord{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("status = ?", InvitationStatusPending).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("get pending invitations: %w", err)
	}
	return int(count), nil
}

// --- Placeholder types for repository queries ---
// These are minimal structs used only for GORM query scoping.
// The actual models are defined in their respective modules.

// EmployeeRecord mirrors the employee.Employee struct for querying.
type EmployeeRecord struct {
	ID        uint
	OrgID     uint
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (EmployeeRecord) TableName() string { return "employees" }

// PayrollRecord mirrors salary.PayrollRecord for querying.
type PayrollRecord struct {
	OrgID     uint
	Year      int
	Month     int
	Status    string
	NetIncome float64
}

func (PayrollRecord) TableName() string { return "payroll_records" }

// SIRecord mirrors socialinsurance.SocialInsuranceRecord for querying.
type SIRecord struct {
	OrgID         uint
	StartMonth    string
	TotalCompany  float64
	TotalPersonal float64
}

func (SIRecord) TableName() string { return "social_insurance_records" }

// VoucherRecord mirrors finance.Voucher for querying.
type VoucherRecord struct {
	OrgID  uint
	Status string
}

func (VoucherRecord) TableName() string { return "vouchers" }

// ExpenseRecord mirrors finance.ExpenseReimbursement for querying.
type ExpenseRecord struct {
	OrgID   uint
	Status  string
}

func (ExpenseRecord) TableName() string { return "expense_reimbursements" }

// TaxReminderRecord mirrors tax.TaxReminder for querying.
type TaxReminderRecord struct {
	OrgID      uint
	IsRead     bool
	IsDismissed bool
	DueDate    *time.Time
}

func (TaxReminderRecord) TableName() string { return "tax_reminders" }

// ContractRecord mirrors employee.Contract for querying.
type ContractRecord struct {
	OrgID   uint
	Status  string
	EndDate *time.Time
}

func (ContractRecord) TableName() string { return "contracts" }

// OffboardingRecord mirrors employee.Offboarding for querying.
type OffboardingRecord struct {
	OrgID   uint
	Status  string
}

func (OffboardingRecord) TableName() string { return "offboardings" }

// InvitationRecord mirrors employee.Invitation for querying.
type InvitationRecord struct {
	OrgID  uint
	Status string
}

func (InvitationRecord) TableName() string { return "invitations" }

// TodoItemRecord mirrors todo.TodoItem for repository queries.
// Defined here to avoid circular import; actual model in internal/todo/model.go.
type TodoItemRecord struct {
	ID             uint
	OrgID          uint
	Status         string `gorm:"column:status"`
	IsTimeLimited  bool   `gorm:"column:is_time_limited"`
}

func (TodoItemRecord) TableName() string { return "todo_items" }

// Status constants for repository queries.
const (
	StatusActive    = "active"
	StatusProbation = "probation"
	StatusResigned  = "resigned"

	PayrollStatusPaid = "paid"

	VoucherStatusSubmitted = "submitted"

	ExpenseStatusPending = "pending"

	ContractStatusActive = "active"

	OffboardingStatusPending  = "pending"
	OffboardingStatusApproved = "approved"

	InvitationStatusPending = "pending"
)
