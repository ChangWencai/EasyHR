package finance

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
)

// FinanceService is the top-level service that coordinates all finance sub-services.
// It also exposes integration methods called by other modules (e.g., salary).
type FinanceService struct {
	voucherService *VoucherService
	accountService *AccountService
}

// NewFinanceService creates a new FinanceService.
func NewFinanceService(voucherService *VoucherService, accountService *AccountService) *FinanceService {
	return &FinanceService{
		voucherService: voucherService,
		accountService: accountService,
	}
}

// GeneratePayrollVoucher generates a wage voucher after PayrollRecord is confirmed.
// Per D-27/D-28/D-29:
//   - source_type = "payroll", source_id = payrollRecordID
//   - DEBIT:  管理费用-工资 (account code 660201 or 6602),  amount = wageTotal
//   - CREDIT: 应付职工薪酬-工资 (account code 280101 or 2801), amount = wageTotal
//
// This is called by salary/service.go after PayrollRecord status transitions to confirmed.
func (s *FinanceService) GeneratePayrollVoucher(ctx context.Context, payrollRecordID, orgID, userID int64, wageTotal decimal.Decimal) (*Voucher, error) {
	// Find the 管理费用-工资 account (code 660201 for sub-account, or 6602 for parent)
	debitAcct, err := s.findAccountByCode(orgID, "660201")
	if err != nil {
		// Fallback to parent account 6602
		debitAcct, err = s.findAccountByCode(orgID, "6602")
		if err != nil {
			return nil, &FinanceError{Code: CodeAccountNotFound, Err: fmt.Errorf("找不到科目 660201 (管理费用-工资)，请确保科目已初始化")}
		}
	}

	// Find the 应付职工薪酬-工资 account (code 280101 for sub-account, or 2801 for parent)
	creditAcct, err := s.findAccountByCode(orgID, "280101")
	if err != nil {
		// Fallback to parent account 2801
		creditAcct, err = s.findAccountByCode(orgID, "2801")
		if err != nil {
			return nil, &FinanceError{Code: CodeAccountNotFound, Err: fmt.Errorf("找不到科目 280101 (应付职工薪酬-工资)，请确保科目已初始化")}
		}
	}

	// Get or create current period
	period, err := s.accountService.GetOrCreateCurrentPeriod(orgID)
	if err != nil {
		return nil, fmt.Errorf("获取会计期间失败: %w", err)
	}

	if period.Status != PeriodStatusOpen {
		return nil, &FinanceError{Code: CodePeriodClosed, Err: fmt.Errorf("当前期间 %d-%02d 已结账，无法生成工资凭证", period.Year, period.Month)}
	}

	// Build voucher date from period
	voucherDate := fmt.Sprintf("%d-%02d-01", period.Year, period.Month)
	summary := fmt.Sprintf("计提%d年%d月工资", period.Year, period.Month)

	req := &CreateVoucherRequest{
		PeriodID:    period.ID,
		VoucherDate: voucherDate,
		Summary:     summary,
		SourceType: SourceTypePayroll,
		SourceID:    &payrollRecordID,
		Entries: []JournalEntryInput{
			{
				AccountID: debitAcct.ID,
				DC:        "debit",
				Amount:    wageTotal.String(),
				Summary:   fmt.Sprintf("计提%d年%d月工资", period.Year, period.Month),
			},
			{
				AccountID: creditAcct.ID,
				DC:        "credit",
				Amount:    wageTotal.String(),
				Summary:   fmt.Sprintf("应付%d年%d月工资", period.Year, period.Month),
			},
		},
	}

	voucher, err := s.voucherService.CreateVoucher(orgID, userID, req)
	if err != nil {
		return nil, fmt.Errorf("生成工资凭证失败: %w", err)
	}

	return voucher, nil
}

// findAccountByCode looks up an account by its code within an org.
func (s *FinanceService) findAccountByCode(orgID int64, code string) (*Account, error) {
	// Access via the account service's repo
	repo := s.accountService.repo
	return repo.GetByCode(orgID, code)
}
