package finance

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wencai/easyhr/internal/common/model"
)

// VoucherServiceInterface is the interface for voucher operations used by ExpenseService.
type VoucherServiceInterface interface {
	CreateVoucher(orgID, userID int64, req *CreateVoucherRequest) (*Voucher, error)
}

// ExpenseService handles business logic for expense reimbursements.
type ExpenseService struct {
	expenseRepo *ExpenseRepository
	accountRepo *AccountRepository
	voucherSvc  VoucherServiceInterface
}

// NewExpenseService creates a new ExpenseService.
func NewExpenseService(expenseRepo *ExpenseRepository, accountRepo *AccountRepository, voucherSvc VoucherServiceInterface) *ExpenseService {
	return &ExpenseService{
		expenseRepo: expenseRepo,
		accountRepo: accountRepo,
		voucherSvc:  voucherSvc,
	}
}

// CreateExpense creates a new expense reimbursement record (submitted by MEMBER).
func (s *ExpenseService) CreateExpense(orgID, userID int64, req *CreateExpenseRequest) (*ExpenseReimbursement, error) {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, &FinanceError{Code: 60240, Err: fmt.Errorf("金额格式错误: %s", req.Amount)}
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, &FinanceError{Code: 60240, Err: fmt.Errorf("金额必须大于零")}
	}

	attachments := ""
	if len(req.Attachments) > 0 {
		if len(req.Attachments) > 9 {
			return nil, &FinanceError{Code: 60241, Err: fmt.Errorf("附件最多9张")}
		}
		urlsJSON, err := json.Marshal(req.Attachments)
		if err != nil {
			return nil, &FinanceError{Code: 60241, Err: fmt.Errorf("附件格式错误")}
		}
		attachments = string(urlsJSON)
	}

	expense := &ExpenseReimbursement{
		BaseModel:   model.BaseModel{OrgID: orgID, CreatedBy: userID, UpdatedBy: userID},
		OrgID:       orgID,
		EmployeeID:  req.EmployeeID,
		Amount:      amount,
		ExpenseType: req.ExpenseType,
		Description: req.Description,
		Attachments: attachments,
		Status:      ExpenseStatusPending,
	}

	if err := s.expenseRepo.Create(expense); err != nil {
		return nil, fmt.Errorf("创建报销单失败: %w", err)
	}
	return expense, nil
}

// ApproveExpense approves a pending expense and auto-generates an expense voucher.
// Per D-24: pending -> approved, then auto-create voucher:
//   DEBIT: 管理费用-XXX (based on expense_type per D-25)
//   CREDIT: 其他应付款-员工借款 (2241)
func (s *ExpenseService) ApproveExpense(orgID, approverID, expenseID int64, note string) (*ExpenseReimbursement, *Voucher, error) {
	expense, err := s.expenseRepo.GetByID(orgID, expenseID)
	if err != nil {
		return nil, nil, &FinanceError{Code: CodeVoucherNotFound, Err: fmt.Errorf("报销单不存在或无权访问")}
	}

	if expense.Status != ExpenseStatusPending {
		return nil, nil, &FinanceError{Code: CodeInvalidStatus, Err: fmt.Errorf("只能审批待审批状态的报销单")}
	}

	// Update status to approved
	now := time.Now()
	expense.Status = ExpenseStatusApproved
	expense.ApproverID = &approverID
	expense.ApprovedAt = &now
	expense.ApprovedNote = note
	expense.UpdatedBy = approverID

	// Find management expense account by type
	expenseAccount, err := s.findExpenseAccount(orgID, expense.ExpenseType)
	if err != nil {
		return nil, nil, fmt.Errorf("查找费用科目失败: %w", err)
	}

	// Find other payables - employee loan account (2241)
	otherPayablesAccount, err := s.accountRepo.GetByCode(orgID, "2241")
	if err != nil {
		return nil, nil, &FinanceError{Code: CodeAccountNotFound, Err: fmt.Errorf("其他应付款-员工借款科目不存在，请先初始化会计科目")}
	}

	// Create voucher: DEBIT expense account, CREDIT other payables
	voucherReq := &CreateVoucherRequest{
		VoucherDate: now.Format("2006-01-02"),
		Summary:     fmt.Sprintf("费用报销-%s", expense.Description),
		SourceType:  SourceTypeExpense,
		SourceID:    &expenseID,
		Entries: []JournalEntryInput{
			{
				AccountID: expenseAccount.ID,
				DC:        string(DCDebit),
				Amount:    expense.Amount.String(),
				Summary:   fmt.Sprintf("费用报销-%s", expense.Description),
			},
			{
				AccountID: otherPayablesAccount.ID,
				DC:        string(DCCredit),
				Amount:    expense.Amount.String(),
				Summary:   "待支付",
			},
		},
	}

	voucher, err := s.voucherSvc.CreateVoucher(orgID, approverID, voucherReq)
	if err != nil {
		return nil, nil, fmt.Errorf("生成费用凭证失败: %w", err)
	}

	expense.VoucherID = &voucher.ID
	if err := s.expenseRepo.Update(expense); err != nil {
		return nil, nil, fmt.Errorf("更新报销单失败: %w", err)
	}

	return expense, voucher, nil
}

// RejectExpense rejects a pending expense reimbursement.
func (s *ExpenseService) RejectExpense(orgID, approverID, expenseID int64, note string) (*ExpenseReimbursement, error) {
	if note == "" {
		return nil, &FinanceError{Code: 60242, Err: fmt.Errorf("驳回原因不能为空")}
	}

	expense, err := s.expenseRepo.GetByID(orgID, expenseID)
	if err != nil {
		return nil, &FinanceError{Code: CodeVoucherNotFound, Err: fmt.Errorf("报销单不存在或无权访问")}
	}

	if expense.Status != ExpenseStatusPending {
		return nil, &FinanceError{Code: CodeInvalidStatus, Err: fmt.Errorf("只能驳回待审批状态的报销单")}
	}

	now := time.Now()
	expense.Status = ExpenseStatusRejected
	expense.RejectedAt = &now
	expense.RejectedNote = note
	expense.UpdatedBy = approverID

	if err := s.expenseRepo.Update(expense); err != nil {
		return nil, fmt.Errorf("更新报销单失败: %w", err)
	}
	return expense, nil
}

// MarkExpensePaid marks an approved expense as paid and generates a payment voucher.
// Per D-25: DEBIT 其他应付款-员工借款, CREDIT 银行存款
func (s *ExpenseService) MarkExpensePaid(orgID, approverID, expenseID int64, note string) (*ExpenseReimbursement, *Voucher, error) {
	expense, err := s.expenseRepo.GetByID(orgID, expenseID)
	if err != nil {
		return nil, nil, &FinanceError{Code: CodeVoucherNotFound, Err: fmt.Errorf("报销单不存在或无权访问")}
	}

	if expense.Status != ExpenseStatusApproved {
		return nil, nil, &FinanceError{Code: CodeInvalidStatus, Err: fmt.Errorf("只能对已批准的报销单进行支付操作")}
	}

	// Update status to paid
	now := time.Now()
	expense.Status = ExpenseStatusPaid
	expense.PaidAt = &now
	expense.PaidNote = note
	expense.UpdatedBy = approverID

	// Find other payables - employee loan account (2241)
	otherPayablesAccount, err := s.accountRepo.GetByCode(orgID, "2241")
	if err != nil {
		return nil, nil, &FinanceError{Code: CodeAccountNotFound, Err: fmt.Errorf("其他应付款-员工借款科目不存在")}
	}

	// Find bank account (1002)
	bankAccount, err := s.accountRepo.GetByCode(orgID, "1002")
	if err != nil {
		return nil, nil, &FinanceError{Code: CodeAccountNotFound, Err: fmt.Errorf("银行存款科目(1002)不存在")}
	}

	// Create payment voucher: DEBIT other payables, CREDIT bank account
	voucherReq := &CreateVoucherRequest{
		VoucherDate: now.Format("2006-01-02"),
		Summary:     fmt.Sprintf("报销支付-%s", expense.Description),
		SourceType:  SourceTypeExpense,
		SourceID:    &expenseID,
		Entries: []JournalEntryInput{
			{
				AccountID: otherPayablesAccount.ID,
				DC:        string(DCDebit),
				Amount:    expense.Amount.String(),
				Summary:   "报销支付",
			},
			{
				AccountID: bankAccount.ID,
				DC:        string(DCCredit),
				Amount:    expense.Amount.String(),
				Summary:   "报销支付",
			},
		},
	}

	voucher, err := s.voucherSvc.CreateVoucher(orgID, approverID, voucherReq)
	if err != nil {
		return nil, nil, fmt.Errorf("生成支付凭证失败: %w", err)
	}

	if err := s.expenseRepo.Update(expense); err != nil {
		return nil, nil, fmt.Errorf("更新报销单失败: %w", err)
	}

	return expense, voucher, nil
}

// GetExpense returns an expense by ID.
func (s *ExpenseService) GetExpense(orgID, expenseID int64) (*ExpenseReimbursement, error) {
	return s.expenseRepo.GetByID(orgID, expenseID)
}

// ListExpenses returns paginated expenses with optional filters.
func (s *ExpenseService) ListExpenses(orgID int64, req *ListExpenseRequest) ([]ExpenseReimbursement, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}
	return s.expenseRepo.List(orgID, req.Status, req.EmployeeID, req.Page, req.Limit)
}

// findExpenseAccount returns the management expense account matching the expense type.
// Per D-25: travel->管理费用-差旅费, transport->管理费用-交通费,
// entertainment->管理费用-业务招待费, office->管理费用-办公费, other->管理费用-其他
func (s *ExpenseService) findExpenseAccount(orgID int64, expenseType ExpenseType) (*Account, error) {
	namePattern := map[ExpenseType]string{
		ExpenseTypeTravel:       "管理费用-差旅费",
		ExpenseTypeTransport:    "管理费用-交通费",
		ExpenseTypeEntertainment: "管理费用-业务招待费",
		ExpenseTypeOffice:       "管理费用-办公费",
		ExpenseTypeOther:        "管理费用-其他",
	}

	name, ok := namePattern[expenseType]
	if !ok {
		name = "管理费用-其他"
	}

	accounts, err := s.accountRepo.GetActiveByOrg(orgID)
	if err != nil {
		return nil, err
	}
	for _, acct := range accounts {
		if acct.Name == name {
			return &acct, nil
		}
	}
	return nil, &FinanceError{Code: CodeAccountNotFound, Err: fmt.Errorf("未找到费用科目: %s", name)}
}
