package finance

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wencai/easyhr/internal/common/model"
	"gorm.io/gorm"
)

// VoucherService handles business logic for vouchers and journal entries.
type VoucherService struct {
	voucherRepo *VoucherRepository
	periodRepo  *PeriodRepository
	accountRepo *AccountRepository
}

// NewVoucherService creates a new VoucherService.
func NewVoucherService(voucherRepo *VoucherRepository, periodRepo *PeriodRepository, accountRepo *AccountRepository) *VoucherService {
	return &VoucherService{
		voucherRepo: voucherRepo,
		periodRepo:  periodRepo,
		accountRepo: accountRepo,
	}
}

// CreateVoucher creates a new voucher with journal entries.
// It validates: (a) period is OPEN, (b) SUM(debit) == SUM(credit) using decimal.Decimal.
func (s *VoucherService) CreateVoucher(orgID, userID int64, req *CreateVoucherRequest) (*Voucher, error) {
	// (a) Validate period is OPEN (D-04) — use req.PeriodID directly
	period, err := s.periodRepo.GetByID(orgID, req.PeriodID)
	if err == gorm.ErrRecordNotFound {
		return nil, &FinanceError{Code: 60223, Err: fmt.Errorf("期间不存在: period_id=%d", req.PeriodID)}
	}
	if err != nil {
		return nil, fmt.Errorf("查询期间失败: %w", err)
	}
	if period.Status == PeriodStatusClosed {
		return nil, &FinanceError{Code: CodePeriodClosed, Err: ErrPeriodClosed.Err}
	}

	// (b) Validate and compute debit/credit sums using decimal.Decimal (D-02, D-03)
	debitSum := decimal.Zero
	creditSum := decimal.Zero
	for _, entry := range req.Entries {
		dc := DCType(entry.DC)
		if dc != DCDebit && dc != DCCredit {
			return nil, &FinanceError{Code: CodeInvalidDC, Err: fmt.Errorf("无效的借贷方向: %s", entry.DC)}
		}
		amount, err := decimal.NewFromString(entry.Amount)
		if err != nil {
			return nil, &FinanceError{Code: 60220, Err: fmt.Errorf("金额格式错误: %s", entry.Amount)}
		}
		if amount.LessThanOrEqual(decimal.Zero) {
			return nil, &FinanceError{Code: 60221, Err: fmt.Errorf("金额必须大于零")}
		}
		if dc == DCDebit {
			debitSum = debitSum.Add(amount)
		} else {
			creditSum = creditSum.Add(amount)
		}
	}

	// (c) Check balance (D-03): SUM(debit) == SUM(credit) — ErrVoucherUnbalanced returned if not equal
	if !debitSum.Equal(creditSum) {
		return nil, WrapError(CodeVoucherUnbalanced, fmt.Errorf("借贷不平衡：借方 %s，贷方 %s", debitSum.String(), creditSum.String()))
	}

	// Parse voucher date
	voucherDate, err := time.Parse("2006-01-02", req.VoucherDate)
	if err != nil {
		return nil, &FinanceError{Code: 60222, Err: fmt.Errorf("日期格式错误，请使用 YYYY-MM-DD")}
	}

	// Determine period from voucher date if not provided
	var periodID int64
	if req.PeriodID > 0 {
		periodID = req.PeriodID
	} else {
		p, err := s.periodRepo.GetOrCreate(orgID, voucherDate.Year(), int(voucherDate.Month()))
		if err != nil {
			return nil, fmt.Errorf("获取期间失败: %w", err)
		}
		if p.Status != PeriodStatusOpen {
			return nil, &FinanceError{Code: CodePeriodClosed, Err: fmt.Errorf("期间 %d-%02d 已结账", p.Year, p.Month)}
		}
		periodID = p.ID
	}

	// Generate voucher number (D-06: "YYYYMM-XXXX")
	voucherNo, err := s.voucherRepo.GetNextVoucherNo(orgID, periodID)
	if err != nil {
		return nil, fmt.Errorf("生成凭证号失败: %w", err)
	}

	sourceType := req.SourceType
	if sourceType == "" {
		sourceType = SourceTypeManual
	}

	// Build voucher and entries
	voucher := &Voucher{
		BaseModel:   model.BaseModel{OrgID: orgID, CreatedBy: userID, UpdatedBy: userID},
		PeriodID:    periodID,
		VoucherNo:   voucherNo,
		Date:        voucherDate,
		Status:      VoucherStatusDraft,
		SourceType:  sourceType,
		SourceID:    req.SourceID,
		Summary:     req.Summary,
	}

	journalEntries := make([]JournalEntry, 0, len(req.Entries))
	for _, entry := range req.Entries {
		amount, _ := decimal.NewFromString(entry.Amount)
		journalEntries = append(journalEntries, JournalEntry{
			BaseModel: model.BaseModel{OrgID: orgID, CreatedBy: userID, UpdatedBy: userID},
			AccountID: entry.AccountID,
			DC:        DCType(entry.DC),
			Amount:    amount,
			Summary:   entry.Summary,
		})
	}

	if err := s.voucherRepo.Create(voucher, journalEntries); err != nil {
		return nil, fmt.Errorf("保存凭证失败: %w", err)
	}

	return voucher, nil
}

// SubmitVoucher transitions a voucher from draft to submitted.
func (s *VoucherService) SubmitVoucher(orgID int64, voucherID int64) error {
	voucher, err := s.voucherRepo.GetByID(orgID, voucherID)
	if err != nil {
		return &FinanceError{Code: CodeVoucherNotFound, Err: fmt.Errorf("凭证不存在或无权访问")}
	}

	if voucher.Status != VoucherStatusDraft {
		return &FinanceError{Code: CodeInvalidStatus, Err: fmt.Errorf("只有草稿状态凭证可以提交")}
	}

	// Check period is not closed — use voucher.PeriodID
	period, err := s.periodRepo.GetByID(orgID, voucher.PeriodID)
	if err == gorm.ErrRecordNotFound {
		return &FinanceError{Code: 60223, Err: fmt.Errorf("期间不存在: period_id=%d", voucher.PeriodID)}
	}
	if err != nil {
		return fmt.Errorf("查询期间失败: %w", err)
	}
	if period.Status == PeriodStatusClosed {
		return &FinanceError{Code: CodePeriodClosed, Err: ErrPeriodClosed.Err}
	}

	return s.voucherRepo.UpdateStatus(orgID, voucherID, VoucherStatusSubmitted)
}

// AuditVoucher transitions a voucher from submitted to audited.
// Only OWNER can audit per D-30.
func (s *VoucherService) AuditVoucher(orgID int64, voucherID int64) error {
	voucher, err := s.voucherRepo.GetByID(orgID, voucherID)
	if err != nil {
		return &FinanceError{Code: CodeVoucherNotFound, Err: fmt.Errorf("凭证不存在或无权访问")}
	}

	if voucher.Status != VoucherStatusSubmitted {
		return &FinanceError{Code: CodeInvalidStatus, Err: fmt.Errorf("只有已提交状态凭证可以审核")}
	}

	return s.voucherRepo.UpdateStatus(orgID, voucherID, VoucherStatusAudited)
}

// ReverseVoucher creates a reversal voucher for an audited voucher.
// Per D-05: DC direction flipped, amount unchanged, reversal_of set, description
// prefixed with "红冲凭证".
func (s *VoucherService) ReverseVoucher(orgID int64, voucherID int64) (*Voucher, error) {
	original, err := s.voucherRepo.GetByID(orgID, voucherID)
	if err != nil {
		return nil, &FinanceError{Code: CodeVoucherNotFound, Err: fmt.Errorf("凭证不存在或无权访问")}
	}

	if original.Status != VoucherStatusAudited {
		return nil, &FinanceError{Code: CodeInvalidStatus, Err: fmt.Errorf("只有已审核凭证可以红冲")}
	}

	// Check period is not closed — use original.PeriodID
	period, err := s.periodRepo.GetByID(orgID, original.PeriodID)
	if err == gorm.ErrRecordNotFound {
		return nil, &FinanceError{Code: 60223, Err: fmt.Errorf("期间不存在: period_id=%d", original.PeriodID)}
	}
	if err != nil {
		return nil, fmt.Errorf("查询期间失败: %w", err)
	}
	if period.Status == PeriodStatusClosed {
		return nil, &FinanceError{Code: CodePeriodClosed, Err: ErrPeriodClosed.Err}
	}

	// Reverse each entry: flip DC direction, amount unchanged
	entries := make([]JournalEntry, len(original.Entries))
	for i, entry := range original.Entries {
		var flippedDC DCType
		if entry.DC == DCDebit {
			flippedDC = DCCredit
		} else {
			flippedDC = DCDebit
		}
		entries[i] = JournalEntry{
			BaseModel: model.BaseModel{OrgID: orgID},
			AccountID: entry.AccountID,
			DC:        flippedDC,
			Amount:    entry.Amount,
			Summary:   entry.Summary,
		}
	}

	return s.voucherRepo.CreateReversal(original, entries)
}

// ListVouchers returns paginated vouchers for a period.
func (s *VoucherService) ListVouchers(orgID int64, req *ListVoucherRequest) ([]Voucher, int64, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}

	if req.PeriodID != nil && *req.PeriodID > 0 {
		return s.voucherRepo.ListByPeriod(orgID, *req.PeriodID, page, limit)
	}
	return s.voucherRepo.Search(orgID, nil, nil, req.Keyword, page, limit)
}

// GetVoucher returns a voucher with its entries.
func (s *VoucherService) GetVoucher(orgID, voucherID int64) (*Voucher, error) {
	return s.voucherRepo.GetByID(orgID, voucherID)
}
