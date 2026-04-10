package wxmp

import (
	"context"
	"errors"
)

// MockWXMPRepository implements WXMPRepository for testing.
type MockWXMPRepository struct {
	Member          *MemberInfo
	Payslips       []PayslipSummary
	PayslipDetail  *PayslipDetail
	Contracts      []ContractDTO
	SocialRecords  []SocialInsuranceDTO
	Expenses       []ExpenseDTO
	Err            error
	SignPayslipErr error
}

func (m *MockWXMPRepository) GetMemberByPhone(ctx context.Context, phoneHash string) (*MemberInfo, error) {
	return m.Member, m.Err
}

func (m *MockWXMPRepository) BindWechatOpenID(ctx context.Context, userID uint, openID string) error {
	return m.Err
}

func (m *MockWXMPRepository) ListPayslips(ctx context.Context, orgID, employeeID uint) ([]PayslipSummary, error) {
	return m.Payslips, m.Err
}

func (m *MockWXMPRepository) GetPayslipByID(ctx context.Context, orgID, employeeID, payslipID uint) (*PayslipDetail, error) {
	return m.PayslipDetail, m.Err
}

func (m *MockWXMPRepository) ListContracts(ctx context.Context, orgID, employeeID uint) ([]ContractDTO, error) {
	return m.Contracts, m.Err
}

func (m *MockWXMPRepository) GetContractByID(ctx context.Context, orgID, employeeID, contractID uint) (*ContractDetail, error) {
	return nil, m.Err
}

func (m *MockWXMPRepository) ListSocialInsurance(ctx context.Context, orgID, employeeID uint) ([]SocialInsuranceDTO, error) {
	return m.SocialRecords, m.Err
}

func (m *MockWXMPRepository) ListExpenses(ctx context.Context, orgID, employeeID uint) ([]ExpenseDTO, error) {
	return m.Expenses, m.Err
}

func (m *MockWXMPRepository) GetExpenseByID(ctx context.Context, orgID, employeeID, expenseID uint) (*ExpenseDTO, error) {
	if len(m.Expenses) == 0 {
		return nil, errors.New("expense not found")
	}
	return &m.Expenses[0], m.Err
}

func (m *MockWXMPRepository) CreateExpense(ctx context.Context, orgID, employeeID uint, req *ExpenseRequest) (*ExpenseDTO, error) {
	return &ExpenseDTO{
		ID:       1,
		Type:     req.Type,
		Amount:   req.Amount,
		Status:   "pending",
	}, m.Err
}

func (m *MockWXMPRepository) SignPayslip(ctx context.Context, orgID, employeeID, payslipID uint) error {
	return m.SignPayslipErr
}
