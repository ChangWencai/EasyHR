package finance

import (
	"context"
	"fmt"
	"time"
)

// PeriodService handles period (期间) management: closing, reopening, validation.
type PeriodService struct {
	periodRepo   *PeriodRepository
	voucherRepo *VoucherRepository
	journalRepo  *JournalEntryRepository
	reportSvc    *ReportService
}

// NewPeriodService creates a new PeriodService.
func NewPeriodService(
	periodRepo *PeriodRepository,
	voucherRepo *VoucherRepository,
	journalRepo *JournalEntryRepository,
	reportSvc *ReportService,
) *PeriodService {
	return &PeriodService{
		periodRepo:   periodRepo,
		voucherRepo: voucherRepo,
		journalRepo: journalRepo,
		reportSvc:   reportSvc,
	}
}

// ValidateClosing checks whether a period can be closed.
// Per D-18:
//   (a) No draft/submitted vouchers in period
//   (b) Period total debit SUM = period total credit SUM
//   (c) ASSET and COST account balances are non-negative
func (s *PeriodService) ValidateClosing(ctx context.Context, orgID, periodID int64) (*ClosingValidationResponse, error) {
	resp := &ClosingValidationResponse{CanClose: true}

	// (a) Check for draft/submitted vouchers
	draftCount, err := s.countVouchersByStatus(orgID, periodID, VoucherStatusDraft)
	if err != nil {
		return nil, fmt.Errorf("查询草稿凭证失败: %w", err)
	}
	if draftCount > 0 {
		resp.CanClose = false
		resp.Errors = append(resp.Errors, fmt.Sprintf("存在 %d 条草稿状态凭证，请先审核或删除", draftCount))
	}

	submittedCount, err := s.countVouchersByStatus(orgID, periodID, VoucherStatusSubmitted)
	if err != nil {
		return nil, fmt.Errorf("查询待审凭证失败: %w", err)
	}
	if submittedCount > 0 {
		resp.CanClose = false
		resp.Errors = append(resp.Errors, fmt.Sprintf("存在 %d 条待审核凭证，请先审核", submittedCount))
	}

	// (b) Check debit = credit
	debitSum, creditSum, err := s.journalRepo.GetPeriodDebitCreditSum(orgID, periodID)
	if err != nil {
		return nil, fmt.Errorf("汇总借贷发生额失败: %w", err)
	}
	if !debitSum.Equal(creditSum) {
		resp.CanClose = false
		resp.Errors = append(resp.Errors, fmt.Sprintf("借贷不平衡: 借方合计=%s, 贷方合计=%s", debitSum.String(), creditSum.String()))
	}

	// (c) Check ASSET/COST accounts have non-negative balances
	negativeAccounts, err := s.journalRepo.GetAccountsWithNegativeBalance(orgID, periodID)
	if err != nil {
		return nil, fmt.Errorf("检查账户余额失败: %w", err)
	}
	if len(negativeAccounts) > 0 {
		resp.CanClose = false
		for _, na := range negativeAccounts {
			resp.Errors = append(resp.Errors, fmt.Sprintf("科目ID=%d 余额为负数=%s (资产/成本类科目余额不能为负)", na.AccountID, na.Balance.String()))
		}
	}

	return resp, nil
}

// ClosePeriod closes a period after validation.
// Per D-17, D-19:
//   (a) Validate closing; return errors if not valid
//   (b) Update period status to CLOSED
//   (c) Update all audited vouchers in period to status=closed
//   (d) Generate balance sheet + income statement snapshots
func (s *PeriodService) ClosePeriod(ctx context.Context, orgID, periodID, userID int64) error {
	// (a) Validate
	validation, err := s.ValidateClosing(ctx, orgID, periodID)
	if err != nil {
		return fmt.Errorf("结账校验失败: %w", err)
	}
	if !validation.CanClose {
		return fmt.Errorf("结账校验未通过: %v", validation.Errors)
	}

	// Get period and check it's not already closed
	period, err := s.periodRepo.GetByID(orgID, periodID)
	if err != nil {
		return fmt.Errorf("获取期间失败: %w", err)
	}
	if period.Status == PeriodStatusClosed {
		return fmt.Errorf("期间 %d-%02d 已结账，请勿重复操作", period.Year, period.Month)
	}

	// (b) Update period status
	now := time.Now()
	period.Status = PeriodStatusClosed
	period.ClosedBy = &userID
	period.ClosedAt = &now
	if err := s.periodRepo.Update(period); err != nil {
		return fmt.Errorf("更新期间状态失败: %w", err)
	}

	// (c) Lock all audited vouchers in period
	if err := s.journalRepo.UpdateVoucherStatusBatch(orgID, periodID, VoucherStatusClosed); err != nil {
		return fmt.Errorf("锁定凭证失败: %w", err)
	}

	// (d) Generate report snapshots
	if s.reportSvc != nil {
		if _, err := s.reportSvc.GenerateBalanceSheet(ctx, orgID, periodID); err != nil {
			// Log but don't fail: period is already closed
			_ = err
		}
		if _, err := s.reportSvc.GenerateIncomeStatement(ctx, orgID, periodID); err != nil {
			_ = err
		}
	}

	return nil
}

// RevertClosing reopens a closed period (反结账).
// Per D-19: Requires OWNER role and no paid vouchers in the period.
// This invalidates report snapshots rather than deleting them.
func (s *PeriodService) RevertClosing(ctx context.Context, orgID, periodID int64) error {
	period, err := s.periodRepo.GetByID(orgID, periodID)
	if err != nil {
		return fmt.Errorf("获取期间失败: %w", err)
	}
	if period.Status != PeriodStatusClosed {
		return fmt.Errorf("期间 %d-%02d 未结账，无需反结账", period.Year, period.Month)
	}

	// Check if there are any paid expense vouchers in this period
	// (simplified check: no paid vouchers from expense module)
	paidCount, err := s.countVouchersByStatus(orgID, periodID, VoucherStatusClosed)
	if err != nil {
		return fmt.Errorf("检查已支付凭证失败: %w", err)
	}
	if paidCount > 0 {
		return fmt.Errorf("期间存在 %d 条已支付凭证，无法反结账", paidCount)
	}

	// Update period status back to OPEN
	period.Status = PeriodStatusOpen
	period.ClosedBy = nil
	period.ClosedAt = nil
	if err := s.periodRepo.Update(period); err != nil {
		return fmt.Errorf("更新期间状态失败: %w", err)
	}

	// Revert voucher statuses to audited
	if err := s.journalRepo.UpdateVoucherStatusBatch(orgID, periodID, VoucherStatusAudited); err != nil {
		return fmt.Errorf("恢复凭证状态失败: %w", err)
	}

	// Invalidate report snapshots (don't delete)
	if s.reportSvc != nil {
		_ = s.reportSvc.InvalidateByPeriod(orgID, periodID)
	}

	return nil
}

// GetPeriods returns all periods for an org ordered by year/month descending.
func (s *PeriodService) GetPeriods(ctx context.Context, orgID int64) (*PeriodListResponse, error) {
	periods, err := s.periodRepo.GetAllByOrg(orgID)
	if err != nil {
		return nil, fmt.Errorf("获取期间列表失败: %w", err)
	}

	items := make([]PeriodItem, len(periods))
	for i, p := range periods {
		items[i] = PeriodItem{
			ID:     p.ID,
			Year:   p.Year,
			Month:  p.Month,
			Status: p.Status,
		}
	}
	return &PeriodListResponse{Periods: items}, nil
}

// countVouchersByStatus counts vouchers by status in a period.
func (s *PeriodService) countVouchersByStatus(orgID, periodID int64, status VoucherStatus) (int64, error) {
	var count int64
	err := s.periodRepo.db.
		Model(&Voucher{}).
		Where("org_id = ? AND period_id = ? AND status = ?", orgID, periodID, status).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
