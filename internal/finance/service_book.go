package finance

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// BookService provides real-time book (账簿) queries: trial balance, ledger, account balance.
// Per D-10: V1.0 computes balances on-the-fly from journal_entries without intermediate tables.
type BookService struct {
	db          *gorm.DB
	accountRepo *AccountRepository
	journalRepo *JournalEntryRepository
	periodRepo  *PeriodRepository
}

// NewBookService creates a new BookService.
func NewBookService(db *gorm.DB, accountRepo *AccountRepository, journalRepo *JournalEntryRepository, periodRepo *PeriodRepository) *BookService {
	return &BookService{
		db:          db,
		accountRepo: accountRepo,
		journalRepo: journalRepo,
		periodRepo:  periodRepo,
	}
}

// GetTrialBalance returns the real-time trial balance for a period.
// It JOINs accounts with journal_entries filtered by period, GROUP BY account,
// and computes balance from the account's normal balance direction (D-10).
func (s *BookService) GetTrialBalance(ctx context.Context, orgID, periodID int64) (*TrialBalanceResponse, error) {
	period, err := s.periodRepo.GetByID(orgID, periodID)
	if err != nil {
		return nil, fmt.Errorf("获取期间失败: %w", err)
	}

	accounts, err := s.accountRepo.GetActiveByOrg(orgID)
	if err != nil {
		return nil, fmt.Errorf("获取科目列表失败: %w", err)
	}

	balanceMap, err := s.journalRepo.SumByAccount(orgID, periodID)
	if err != nil {
		return nil, fmt.Errorf("汇总账户发生额失败: %w", err)
	}

	items := make([]AccountBalanceResponse, 0, len(accounts))
	var totalDebit, totalCredit decimal.Decimal

	for _, acct := range accounts {
		row := balanceMap[acct.ID]
		debitSum := decimal.Zero
		creditSum := decimal.Zero
		if row != nil {
			debitSum = row.DebitSum
			creditSum = row.CreditSum
		}

		var balance decimal.Decimal
		switch acct.NormalBalance {
		case NormalBalanceDebit:
			balance = debitSum.Sub(creditSum)
		case NormalBalanceCredit:
			balance = creditSum.Sub(debitSum)
		}

		items = append(items, AccountBalanceResponse{
			AccountID:  acct.ID,
			Code:       acct.Code,
			Name:       acct.Name,
			Category:   acct.Category,
			DebitSum:   debitSum,
			CreditSum:  creditSum,
			Balance:    balance,
			IsLeaf:     acct.Level >= 2,
		})

		totalDebit = totalDebit.Add(debitSum)
		totalCredit = totalCredit.Add(creditSum)
	}

	return &TrialBalanceResponse{
		Items:       items,
		TotalDebit:  totalDebit,
		TotalCredit: totalCredit,
		IsBalanced:  totalDebit.Equal(totalCredit),
		PeriodID:    periodID,
		Year:        period.Year,
		Month:       period.Month,
	}, nil
}

// GetAccountBalance returns all journal entries for an account up to a period,
// including the opening balance and period entries with running balance.
func (s *BookService) GetAccountBalance(ctx context.Context, orgID, periodID, accountID int64) (*LedgerResponse, error) {
	acct, err := s.accountRepo.GetByID(orgID, accountID)
	if err != nil {
		return nil, fmt.Errorf("获取科目失败: %w", err)
	}
	period, err := s.periodRepo.GetByID(orgID, periodID)
	if err != nil {
		return nil, fmt.Errorf("获取期间失败: %w", err)
	}

	entries, err := s.journalRepo.GetByAccountUpToPeriod(orgID, accountID, periodID)
	if err != nil {
		return nil, fmt.Errorf("获取账户明细失败: %w", err)
	}

	openingBal, periodEntries := s.splitOpeningAndPeriodEntries(entries, periodID, acct.NormalBalance)

	ledgerEntries := make([]LedgerEntryResponse, 0, len(periodEntries))
	runningBalance := openingBal
	var periodDebit, periodCredit decimal.Decimal

	for _, e := range periodEntries {
		if e.DC == DCDebit {
			runningBalance = runningBalance.Add(e.Amount)
			periodDebit = periodDebit.Add(e.Amount)
		} else {
			runningBalance = runningBalance.Sub(e.Amount)
			periodCredit = periodCredit.Add(e.Amount)
		}
		ledgerEntries = append(ledgerEntries, LedgerEntryResponse{
			VoucherNo:    e.VoucherNo,
			VoucherDate:  e.VoucherDate.Format("2006-01-02"),
			Description:  e.Summary,
			AccountName:  acct.Name,
			DC:           string(e.DC),
			Amount:       e.Amount,
			BalanceAfter: runningBalance,
		})
	}

	return &LedgerResponse{
		AccountID:        acct.ID,
		AccountCode:      acct.Code,
		AccountName:      acct.Name,
		PeriodID:         periodID,
		Year:             period.Year,
		Month:            period.Month,
		Entries:          ledgerEntries,
		PeriodDebitSum:   periodDebit,
		PeriodCreditSum:  periodCredit,
		EndingBalance:    runningBalance,
		OpeningBalance:   openingBal,
	}, nil
}

// GetLedger returns paginated ledger entries for an account in a period.
func (s *BookService) GetLedger(ctx context.Context, orgID, periodID, accountID int64, page, limit int) (*LedgerResponse, error) {
	acct, err := s.accountRepo.GetByID(orgID, accountID)
	if err != nil {
		return nil, fmt.Errorf("获取科目失败: %w", err)
	}
	period, err := s.periodRepo.GetByID(orgID, periodID)
	if err != nil {
		return nil, fmt.Errorf("获取期间失败: %w", err)
	}

	entries, err := s.journalRepo.GetByAccountUpToPeriod(orgID, accountID, periodID)
	if err != nil {
		return nil, fmt.Errorf("获取账户明细失败: %w", err)
	}

	openingBal, periodEntries := s.splitOpeningAndPeriodEntries(entries, periodID, acct.NormalBalance)

	runningBalance := openingBal
	type entryBal struct {
		JournalEntryWithVoucher
		RunningBalance decimal.Decimal
	}
	withBal := make([]entryBal, 0, len(periodEntries))
	for _, e := range periodEntries {
		if e.DC == DCDebit {
			runningBalance = runningBalance.Add(e.Amount)
		} else {
			runningBalance = runningBalance.Sub(e.Amount)
		}
		withBal = append(withBal, entryBal{e, runningBalance})
	}

	total := int64(len(withBal))
	start := (page - 1) * limit
	if start > int(total) {
		start = int(total)
	}
	end := start + limit
	if end > int(total) {
		end = int(total)
	}
	paged := withBal[start:end]

	ledgerEntries := make([]LedgerEntryResponse, 0, len(paged))
	for _, e := range paged {
		ledgerEntries = append(ledgerEntries, LedgerEntryResponse{
			VoucherNo:    e.VoucherNo,
			VoucherDate:  e.VoucherDate.Format("2006-01-02"),
			Description:  e.Summary,
			AccountName:  acct.Name,
			DC:           string(e.DC),
			Amount:       e.Amount,
			BalanceAfter: e.RunningBalance,
		})
	}

	var periodDebit, periodCredit decimal.Decimal
	for _, e := range periodEntries {
		if e.DC == DCDebit {
			periodDebit = periodDebit.Add(e.Amount)
		} else {
			periodCredit = periodCredit.Add(e.Amount)
		}
	}

	return &LedgerResponse{
		AccountID:        acct.ID,
		AccountCode:      acct.Code,
		AccountName:      acct.Name,
		PeriodID:         periodID,
		Year:             period.Year,
		Month:            period.Month,
		Entries:          ledgerEntries,
		PeriodDebitSum:   periodDebit,
		PeriodCreditSum:  periodCredit,
		EndingBalance:    runningBalance,
		OpeningBalance:   openingBal,
	}, nil
}

// splitOpeningAndPeriodEntries separates entries into opening balance and current period entries.
func (s *BookService) splitOpeningAndPeriodEntries(
	entries []JournalEntryWithVoucher,
	periodID int64,
	normalBal NormalBalance,
) (openingBal decimal.Decimal, periodEntries []JournalEntryWithVoucher) {
	for _, e := range entries {
		if e.PeriodID < periodID {
			// Add to opening balance using normal balance direction
			switch normalBal {
			case NormalBalanceDebit:
				if e.DC == DCDebit {
					openingBal = openingBal.Add(e.Amount)
				} else {
					openingBal = openingBal.Sub(e.Amount)
				}
			case NormalBalanceCredit:
				if e.DC == DCCredit {
					openingBal = openingBal.Add(e.Amount)
				} else {
					openingBal = openingBal.Sub(e.Amount)
				}
			}
		} else if e.PeriodID == periodID {
			periodEntries = append(periodEntries, e)
		}
	}
	return
}
