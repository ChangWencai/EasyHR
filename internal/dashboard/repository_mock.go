package dashboard

import (
	"context"
	"errors"
)

// MockDashboardRepository implements DashboardRepository for testing.
type MockDashboardRepository struct {
	EmployeeCount       int
	JoinedThisMonth     int
	LeftThisMonth       int
	PayrollTotal       string
	SocialInsuranceAmt string
	PendingVouchers     int
	PendingExpenses     int
	TaxReminders        int
	ContractExpirations int
	PendingOffboardings int
	PendingInvitations  int
	Err                 error
}

var _ DashboardRepository = (*MockDashboardRepository)(nil)

func (m *MockDashboardRepository) GetEmployeeStats(ctx context.Context, orgID int64) (active, joined, left int, err error) {
	return m.EmployeeCount, m.JoinedThisMonth, m.LeftThisMonth, m.Err
}

func (m *MockDashboardRepository) GetPayrollTotal(ctx context.Context, orgID int64) (string, error) {
	return m.PayrollTotal, m.Err
}

func (m *MockDashboardRepository) GetSocialInsuranceTotal(ctx context.Context, orgID int64) (string, error) {
	return m.SocialInsuranceAmt, m.Err
}

func (m *MockDashboardRepository) GetPendingVouchers(ctx context.Context, orgID int64) (int, error) {
	return m.PendingVouchers, m.Err
}

func (m *MockDashboardRepository) GetPendingExpenses(ctx context.Context, orgID int64) (int, error) {
	return m.PendingExpenses, m.Err
}

func (m *MockDashboardRepository) GetTaxReminders(ctx context.Context, orgID int64) (int, error) {
	return m.TaxReminders, m.Err
}

func (m *MockDashboardRepository) GetContractExpirations(ctx context.Context, orgID int64) (int, error) {
	return m.ContractExpirations, m.Err
}

func (m *MockDashboardRepository) GetPendingOffboardings(ctx context.Context, orgID int64) (int, error) {
	return m.PendingOffboardings, m.Err
}

func (m *MockDashboardRepository) GetPendingInvitations(ctx context.Context, orgID int64) (int, error) {
	return m.PendingInvitations, m.Err
}

func (m *MockDashboardRepository) GetTodoRingStats(ctx context.Context, orgID int64) (int, int, error) {
	return 0, 0, m.Err
}

func (m *MockDashboardRepository) GetTimeLimitedRingStats(ctx context.Context, orgID int64) (int, int, error) {
	return 0, 0, m.Err
}

// ErrorMockRepository returns an error on every call — useful for error propagation tests.
type ErrorMockRepository struct{}

var _ DashboardRepository = (*ErrorMockRepository)(nil)

func (e *ErrorMockRepository) GetEmployeeStats(ctx context.Context, orgID int64) (int, int, int, error) {
	return 0, 0, 0, errors.New("employee stats error")
}

func (e *ErrorMockRepository) GetPayrollTotal(ctx context.Context, orgID int64) (string, error) {
	return "0", errors.New("payroll total error")
}

func (e *ErrorMockRepository) GetSocialInsuranceTotal(ctx context.Context, orgID int64) (string, error) {
	return "0", errors.New("social insurance error")
}

func (e *ErrorMockRepository) GetPendingVouchers(ctx context.Context, orgID int64) (int, error) {
	return 0, errors.New("pending vouchers error")
}

func (e *ErrorMockRepository) GetPendingExpenses(ctx context.Context, orgID int64) (int, error) {
	return 0, errors.New("pending expenses error")
}

func (e *ErrorMockRepository) GetTaxReminders(ctx context.Context, orgID int64) (int, error) {
	return 0, errors.New("tax reminders error")
}

func (e *ErrorMockRepository) GetContractExpirations(ctx context.Context, orgID int64) (int, error) {
	return 0, errors.New("contract expirations error")
}

func (e *ErrorMockRepository) GetPendingOffboardings(ctx context.Context, orgID int64) (int, error) {
	return 0, errors.New("pending offboardings error")
}

func (e *ErrorMockRepository) GetPendingInvitations(ctx context.Context, orgID int64) (int, error) {
	return 0, errors.New("pending invitations error")
}

func (e *ErrorMockRepository) GetTodoRingStats(ctx context.Context, orgID int64) (int, int, error) {
	return 0, 0, errors.New("todo ring stats error")
}

func (e *ErrorMockRepository) GetTimeLimitedRingStats(ctx context.Context, orgID int64) (int, int, error) {
	return 0, 0, errors.New("time-limited ring stats error")
}
