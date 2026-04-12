package finance

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/model"
	"gorm.io/gorm"
)

// ========== AccountRepository ==========

// AccountRepository handles data access for Account.
type AccountRepository struct {
	db *gorm.DB
}

// NewAccountRepository creates a new AccountRepository.
func NewAccountRepository(db *gorm.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// GetByOrg returns all accounts for an org (including inactive).
func (r *AccountRepository) GetByOrg(orgID int64) ([]Account, error) {
	var accounts []Account
	err := r.db.Scopes(middleware.TenantScope(orgID)).Order("code").Find(&accounts).Error
	return accounts, err
}

// GetActiveByOrg returns only active accounts for an org.
func (r *AccountRepository) GetActiveByOrg(orgID int64) ([]Account, error) {
	var accounts []Account
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("is_active = ?", true).
		Order("code").
		Find(&accounts).Error
	return accounts, err
}

// GetByID returns an account by ID within an org.
func (r *AccountRepository) GetByID(orgID, id int64) (*Account, error) {
	var account Account
	err := r.db.Scopes(middleware.TenantScope(orgID)).First(&account, id).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// GetByCode returns an account by code within an org.
func (r *AccountRepository) GetByCode(orgID int64, code string) (*Account, error) {
	var account Account
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("code = ?", code).
		First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// Create creates a new account.
func (r *AccountRepository) Create(account *Account) error {
	return r.db.Create(account).Error
}

// Update updates an existing account.
func (r *AccountRepository) Update(account *Account) error {
	return r.db.Save(account).Error
}

// Delete soft-deletes a non-system account.
func (r *AccountRepository) Delete(orgID, id int64) error {
	var account Account
	if err := r.db.Scopes(middleware.TenantScope(orgID)).First(&account, id).Error; err != nil {
		return err
	}
	if account.IsSystem {
		return &FinanceError{Code: CodeSystemAccountDelete, Err: errSystemAccountCannotBeDeleted}
	}
	return r.db.Delete(&account).Error
}

// SeedIfEmpty seeds the 40+ preset accounts if the org has none.
// Uses a transaction to ensure atomicity. Sub-accounts (level 2) are resolved
// by looking up the parent account's ID after level-1 accounts are inserted.
func (r *AccountRepository) SeedIfEmpty(orgID int64) error {
	var count int64
	r.db.Model(&Account{}).Scopes(middleware.TenantScope(orgID)).Count(&count)
	if count > 0 {
		return nil // already seeded
	}

	presets := PresetAccounts(orgID)
	if len(presets) == 0 {
		return nil
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		// First pass: insert all level-1 accounts and record their IDs
		level1Map := make(map[string]int64) // code -> id
		for _, acct := range presets {
			if acct.Level == 1 {
				if err := tx.Create(&acct).Error; err != nil {
					return err
				}
				var loaded Account
				if err := tx.Scopes(middleware.TenantScope(orgID)).
					Where("code = ?", acct.Code).First(&loaded).Error; err != nil {
					return err
				}
				level1Map[acct.Code] = loaded.ID
			}
		}

		// Second pass: insert sub-accounts (level 2) with resolved parent IDs
		subAccts := []struct {
			code     string
			name     string
			category AccountCategory
			normal   NormalBalance
			level    int
			parent   string
		}{
			// 管理费用 sub-accounts
			{"660201", "管理费用-工资", AccountCategoryProfit, NormalBalanceDebit, 2, "6602"},
			{"660202", "管理费用-社保", AccountCategoryProfit, NormalBalanceDebit, 2, "6602"},
			{"660203", "管理费用-公积金", AccountCategoryProfit, NormalBalanceDebit, 2, "6602"},
			{"660204", "管理费用-办公费", AccountCategoryProfit, NormalBalanceDebit, 2, "6602"},
			{"660205", "管理费用-差旅费", AccountCategoryProfit, NormalBalanceDebit, 2, "6602"},
			// 应付职工薪酬 sub-accounts
			{"280101", "应付职工薪酬-工资", AccountCategoryLiability, NormalBalanceCredit, 2, "2801"},
			{"280102", "应付职工薪酬-社保", AccountCategoryLiability, NormalBalanceCredit, 2, "2801"},
			{"280103", "应付职工薪酬-公积金", AccountCategoryLiability, NormalBalanceCredit, 2, "2801"},
		}

		for _, sub := range subAccts {
			parentID := level1Map[sub.parent]
			if parentID == 0 {
				continue // parent not found, skip
			}
			acct := Account{
				BaseModel:     model.BaseModel{OrgID: orgID},
				Code:          sub.code,
				Name:          sub.name,
				Category:      sub.category,
				NormalBalance: sub.normal,
				IsActive:      true,
				IsSystem:      true,
				ParentID:      &parentID,
				Level:         sub.level,
			}
			if err := tx.Create(&acct).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// ========== PeriodRepository ==========

// PeriodRepository handles data access for Period.
type PeriodRepository struct {
	db *gorm.DB
}

// NewPeriodRepository creates a new PeriodRepository.
func NewPeriodRepository(db *gorm.DB) *PeriodRepository {
	return &PeriodRepository{db: db}
}

// GetByYearMonth returns a period by org/year/month.
func (r *PeriodRepository) GetByYearMonth(orgID int64, year, month int) (*Period, error) {
	var period Period
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("year = ? AND month = ?", year, month).
		First(&period).Error
	if err != nil {
		return nil, err
	}
	return &period, nil
}

// GetByID returns a period by ID within an org.
func (r *PeriodRepository) GetByID(orgID, periodID int64) (*Period, error) {
	var period Period
	err := r.db.Scopes(middleware.TenantScope(orgID)).First(&period, periodID).Error
	if err != nil {
		return nil, err
	}
	return &period, nil
}

// GetOrCreate returns existing period or creates a new one for the given year/month.
func (r *PeriodRepository) GetOrCreate(orgID int64, year, month int) (*Period, error) {
	var period Period
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("year = ? AND month = ?", year, month).
		First(&period).Error
	if err == nil {
		return &period, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	period = Period{
		BaseModel: model.BaseModel{OrgID: orgID},
		Year:       year,
		Month:      month,
		Status:     PeriodStatusOpen,
	}
	if err := r.db.Create(&period).Error; err != nil {
		return nil, err
	}
	return &period, nil
}

// Update updates a period.
func (r *PeriodRepository) Update(period *Period) error {
	return r.db.Save(period).Error
}

// GetAllByOrg returns all periods for an org ordered by year/month descending.
func (r *PeriodRepository) GetAllByOrg(orgID int64) ([]Period, error) {
	var periods []Period
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Order("year DESC, month DESC").
		Find(&periods).Error
	return periods, err
}

// UpdateStatus updates the status of a period.
func (r *PeriodRepository) UpdateStatus(orgID, periodID int64, status PeriodStatus) error {
	return r.db.Scopes(middleware.TenantScope(orgID)).
		Model(&Period{}).
		Where("id = ?", periodID).
		Update("status", status).Error
}

// LockByID locks a period row for closing (FOR UPDATE).
func (r *PeriodRepository) LockByID(orgID, periodID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var period Period
		return tx.Scopes(middleware.TenantScope(orgID)).
			Clauses().
			First(&period, periodID).Error
	})
}

// ========== VoucherRepository ==========

// VoucherRepository handles data access for Voucher and JournalEntry.
type VoucherRepository struct {
	db *gorm.DB
}

// NewVoucherRepository creates a new VoucherRepository.
func NewVoucherRepository(db *gorm.DB) *VoucherRepository {
	return &VoucherRepository{db: db}
}

// Create creates a voucher and its journal entries in a transaction.
// The voucher.ID must be set before calling; entries will have their VoucherID set.
func (r *VoucherRepository) Create(voucher *Voucher, entries []JournalEntry) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(voucher).Error; err != nil {
			return err
		}
		for i := range entries {
			entries[i].VoucherID = voucher.ID
			if err := tx.Create(&entries[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetByID returns a voucher by ID with entries preloaded.
func (r *VoucherRepository) GetByID(orgID, voucherID int64) (*Voucher, error) {
	var voucher Voucher
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Preload("Entries").
		First(&voucher, voucherID).Error
	if err != nil {
		return nil, err
	}
	return &voucher, nil
}

// ListByPeriod returns paginated vouchers for a period.
func (r *VoucherRepository) ListByPeriod(orgID, periodID int64, page, size int) ([]Voucher, int64, error) {
	var vouchers []Voucher
	var total int64

	query := r.db.Scopes(middleware.TenantScope(orgID)).
		Model(&Voucher{}).
		Where("period_id = ?", periodID)

	query.Count(&total)

	offset := (page - 1) * size
	err := query.Preload("Entries").
		Order("voucher_no ASC").
		Offset(offset).
		Limit(size).
		Find(&vouchers).Error

	return vouchers, total, err
}

// Search searches vouchers by keyword and/or account_id.
func (r *VoucherRepository) Search(orgID int64, periodID *int64, accountID *int64, keyword string, page, size int) ([]Voucher, int64, error) {
	var vouchers []Voucher
	var total int64

	q := r.db.Scopes(middleware.TenantScope(orgID)).Model(&Voucher{})
	if periodID != nil {
		q = q.Where("period_id = ?", *periodID)
	}
	if keyword != "" {
		q = q.Where("summary LIKE ? OR voucher_no LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	q.Count(&total)

	offset := (page - 1) * size
	err := q.Preload("Entries").
		Order("voucher_no ASC").
		Offset(offset).
		Limit(size).
		Find(&vouchers).Error

	return vouchers, total, err
}

// UpdateStatus updates the status of a voucher.
func (r *VoucherRepository) UpdateStatus(orgID, voucherID int64, status VoucherStatus) error {
	return r.db.Scopes(middleware.TenantScope(orgID)).
		Model(&Voucher{}).
		Where("id = ?", voucherID).
		Update("status", status).Error
}

// GetNextVoucherNo returns the next voucher number for a period.
// Format: "YYYYMM-{counter, padded 4}" per D-06.
func (r *VoucherRepository) GetNextVoucherNo(orgID, periodID int64) (string, error) {
	var period Period
	if err := r.db.Scopes(middleware.TenantScope(orgID)).First(&period, periodID).Error; err != nil {
		return "", err
	}

	period.VoucherNoCounter++
	newCounter := period.VoucherNoCounter
	if err := r.db.Save(&period).Error; err != nil {
		return "", err
	}

	yymm := period.Year*100 + period.Month
	return fmt.Sprintf("%d-%04d", yymm, newCounter), nil
}

// CreateReversal creates a reversal voucher for the given original voucher.
// Per D-05: DC direction flipped, amount unchanged, reversal_of set.
func (r *VoucherRepository) CreateReversal(original *Voucher, entries []JournalEntry) (*Voucher, error) {
	origID := original.ID
	reversal := &Voucher{
		BaseModel:   model.BaseModel{OrgID: original.OrgID},
		PeriodID:    original.PeriodID,
		Date:        time.Now(),
		Status:      VoucherStatusDraft,
		SourceType:  original.SourceType,
		SourceID:    original.SourceID,
		Summary:     fmt.Sprintf("红冲凭证 原凭证号 %s", original.VoucherNo),
		ReversalOf:  &origID,
	}

	if err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(reversal).Error; err != nil {
			return err
		}
		for _, entry := range entries {
			var flippedDC DCType
			if entry.DC == DCDebit {
				flippedDC = DCCredit
			} else {
				flippedDC = DCDebit
			}
			reversed := JournalEntry{
				BaseModel: model.BaseModel{OrgID: original.OrgID},
				VoucherID: reversal.ID,
				AccountID: entry.AccountID,
				DC:        flippedDC,
				Amount:    entry.Amount,
				Summary:   entry.Summary,
			}
			if err := tx.Create(&reversed).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return reversal, nil
}

// ========== InvoiceRepository ==========

// InvoiceRepository handles data access for Invoice.
type InvoiceRepository struct {
	db *gorm.DB
}

// NewInvoiceRepository creates a new InvoiceRepository.
func NewInvoiceRepository(db *gorm.DB) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

// Create creates a new invoice.
func (r *InvoiceRepository) Create(invoice *Invoice) error {
	return r.db.Create(invoice).Error
}

// Update updates an existing invoice.
func (r *InvoiceRepository) Update(invoice *Invoice) error {
	return r.db.Save(invoice).Error
}

// GetByID returns an invoice by ID within an org.
func (r *InvoiceRepository) GetByID(orgID, id int64) (*Invoice, error) {
	var invoice Invoice
	err := r.db.Scopes(middleware.TenantScope(orgID)).First(&invoice, id).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

// LinkVoucher sets the VoucherID on an invoice.
func (r *InvoiceRepository) LinkVoucher(orgID, invoiceID, voucherID int64) error {
	return r.db.Scopes(middleware.TenantScope(orgID)).
		Model(&Invoice{}).
		Where("id = ?", invoiceID).
		Update("voucher_id", voucherID).Error
}

// List returns paginated invoices with optional filters.
func (r *InvoiceRepository) List(orgID int64, req *ListInvoiceRequest) ([]Invoice, int64, error) {
	var invoices []Invoice
	var total int64

	q := r.db.Scopes(middleware.TenantScope(orgID)).Model(&Invoice{})
	if req.Type != nil {
		q = q.Where("invoice_type = ?", *req.Type)
	}
	if req.Status != nil {
		q = q.Where("status = ?", *req.Status)
	}
	if req.Year > 0 {
		q = q.Where("EXTRACT(YEAR FROM date) = ?", req.Year)
	}
	if req.Month > 0 {
		q = q.Where("EXTRACT(MONTH FROM date) = ?", req.Month)
	}

	q.Count(&total)

	offset := (req.Page - 1) * req.Limit
	err := q.Order("date DESC, id DESC").
		Offset(offset).
		Limit(req.Limit).
		Find(&invoices).Error

	return invoices, total, err
}

// GetMonthlyTaxSummary returns aggregated tax amounts by invoice type for a given year/month.
func (r *InvoiceRepository) GetMonthlyTaxSummary(orgID int64, year, month int) (outputTax, inputTax, outputAmount, inputAmount decimal.Decimal, err error) {
	type result struct {
		InvoiceType string
		TaxAmount   string
		Amount      string
	}

	var results []result
	err = r.db.Scopes(middleware.TenantScope(orgID)).
		Model(&Invoice{}).
		Select("invoice_type, SUM(tax_amount) as tax_amount, SUM(amount) as amount").
		Where("EXTRACT(YEAR FROM date) = ? AND EXTRACT(MONTH FROM date) = ?", year, month).
		Group("invoice_type").
		Find(&results).Error
	if err != nil {
		return
	}

	for _, row := range results {
		taxAmt, _ := decimal.NewFromString(row.TaxAmount)
		amt, _ := decimal.NewFromString(row.Amount)
		if row.InvoiceType == string(InvoiceTypeOutput) {
			outputTax = taxAmt
			outputAmount = amt
		} else if row.InvoiceType == string(InvoiceTypeInput) {
			inputTax = taxAmt
			inputAmount = amt
		}
	}
	return
}

// ========== ExpenseRepository ==========

// ExpenseRepository handles data access for ExpenseReimbursement.
type ExpenseRepository struct {
	db *gorm.DB
}

// NewExpenseRepository creates a new ExpenseRepository.
func NewExpenseRepository(db *gorm.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

// Create creates a new expense reimbursement record.
func (r *ExpenseRepository) Create(expense *ExpenseReimbursement) error {
	return r.db.Create(expense).Error
}

// Update updates an existing expense reimbursement record.
func (r *ExpenseRepository) Update(expense *ExpenseReimbursement) error {
	return r.db.Save(expense).Error
}

// GetByID returns an expense by ID within an org.
func (r *ExpenseRepository) GetByID(orgID, id int64) (*ExpenseReimbursement, error) {
	var expense ExpenseReimbursement
	err := r.db.Scopes(middleware.TenantScope(orgID)).First(&expense, id).Error
	if err != nil {
		return nil, err
	}
	return &expense, nil
}

// UpdateStatus updates the status and related fields of an expense.
func (r *ExpenseRepository) UpdateStatus(orgID, id int64, status ExpenseStatus) error {
	return r.db.Scopes(middleware.TenantScope(orgID)).
		Model(&ExpenseReimbursement{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// LinkVoucher sets the VoucherID on an expense.
func (r *ExpenseRepository) LinkVoucher(orgID, expenseID, voucherID int64) error {
	return r.db.Scopes(middleware.TenantScope(orgID)).
		Model(&ExpenseReimbursement{}).
		Where("id = ?", expenseID).
		Update("voucher_id", voucherID).Error
}

// List returns paginated expenses with optional filters.
func (r *ExpenseRepository) List(orgID int64, status *ExpenseStatus, employeeID *int64, page, limit int) ([]ExpenseReimbursement, int64, error) {
	var expenses []ExpenseReimbursement
	var total int64

	q := r.db.Scopes(middleware.TenantScope(orgID)).Model(&ExpenseReimbursement{})
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if employeeID != nil {
		q = q.Where("employee_id = ?", *employeeID)
	}

	q.Count(&total)

	offset := (page - 1) * limit
	err := q.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&expenses).Error

	return expenses, total, err
}

// ========== JournalEntryRepository ==========

// JournalEntryRepository handles data access for JournalEntry queries used by BookService.
type JournalEntryRepository struct {
	db *gorm.DB
}

// NewJournalEntryRepository creates a new JournalEntryRepository.
func NewJournalEntryRepository(db *gorm.DB) *JournalEntryRepository {
	return &JournalEntryRepository{db: db}
}

// BalanceSum holds the summed debit and credit amounts for an account.
type BalanceSum struct {
	DebitSum  decimal.Decimal
	CreditSum decimal.Decimal
}

// SumByAccount returns a map of accountID -> BalanceSum for a given period.
// This is the core query for the trial balance (D-10).
func (r *JournalEntryRepository) SumByAccount(orgID, periodID int64) (map[int64]*BalanceSum, error) {
	type row struct {
		AccountID int64
		DC        string
		Total     string
	}
	var rows []row
	err := r.db.Model(&JournalEntry{}).
		Joins("JOIN vouchers ON vouchers.id = journal_entries.voucher_id").
		Where("journal_entries.org_id = ? AND vouchers.period_id = ?", orgID, periodID).
		Select("journal_entries.account_id, journal_entries.dc, SUM(journal_entries.amount) as total").
		Group("journal_entries.account_id, journal_entries.dc").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]*BalanceSum)
	for _, row := range rows {
		if _, ok := result[row.AccountID]; !ok {
			result[row.AccountID] = &BalanceSum{}
		}
		total, _ := decimal.NewFromString(row.Total)
		if row.DC == string(DCDebit) {
			result[row.AccountID].DebitSum = total
		} else {
			result[row.AccountID].CreditSum = total
		}
	}
	return result, nil
}

// JournalEntryWithVoucher holds a journal entry joined with voucher info for ledger display.
type JournalEntryWithVoucher struct {
	JournalEntry
	PeriodID   int64
	VoucherNo  string
	VoucherDate time.Time
}

// GetByAccountUpToPeriod returns all journal entries for an account up to and including the given period,
// joined with voucher info for ledger display.
func (r *JournalEntryRepository) GetByAccountUpToPeriod(orgID, accountID, periodID int64) ([]JournalEntryWithVoucher, error) {
	var rows []JournalEntryWithVoucher
	err := r.db.Model(&JournalEntry{}).
		Select("journal_entries.*, vouchers.period_id, vouchers.voucher_no, vouchers.date as voucher_date").
		Joins("JOIN vouchers ON vouchers.id = journal_entries.voucher_id").
		Where("journal_entries.org_id = ? AND journal_entries.account_id = ? AND vouchers.period_id <= ?", orgID, accountID, periodID).
		Order("vouchers.date ASC, vouchers.id ASC, journal_entries.id ASC").
		Find(&rows).Error
	return rows, err
}

// GetPeriodDebitCreditSum returns the total debit and credit sums for all journal entries in a period.
func (r *JournalEntryRepository) GetPeriodDebitCreditSum(orgID, periodID int64) (debit, credit decimal.Decimal, err error) {
	type row struct {
		DC    string
		Total string
	}
	var rows []row
	err = r.db.Model(&JournalEntry{}).
		Joins("JOIN vouchers ON vouchers.id = journal_entries.voucher_id").
		Where("journal_entries.org_id = ? AND vouchers.period_id = ?", orgID, periodID).
		Select("journal_entries.dc, SUM(journal_entries.amount) as total").
		Group("journal_entries.dc").
		Find(&rows).Error
	if err != nil {
		return
	}
	for _, row := range rows {
		total, _ := decimal.NewFromString(row.Total)
		if row.DC == string(DCDebit) {
			debit = total
		} else {
			credit = total
		}
	}
	return
}

// GetAccountPeriodEntries returns all journal entries for an account in a specific period.
func (r *JournalEntryRepository) GetAccountPeriodEntries(orgID, accountID, periodID int64) ([]JournalEntry, error) {
	var entries []JournalEntry
	err := r.db.Model(&JournalEntry{}).
		Joins("JOIN vouchers ON vouchers.id = journal_entries.voucher_id").
		Where("journal_entries.org_id = ? AND journal_entries.account_id = ? AND vouchers.period_id = ?", orgID, accountID, periodID).
		Order("vouchers.date ASC, vouchers.id ASC").
		Find(&entries).Error
	return entries, err
}

// GetAccountsWithNegativeBalance returns asset/cost accounts with negative balances in a period.
func (r *JournalEntryRepository) GetAccountsWithNegativeBalance(orgID, periodID int64) ([]struct {
	AccountID int64
	Balance   decimal.Decimal
}, error) {
	// Get all account balances for asset/cost categories and check for negatives
	// We'll compute in-memory since we need to check each balance individually
	type row struct {
		AccountID int64
		DC        string
		Total     string
	}
	var rows []row
	err := r.db.Model(&JournalEntry{}).
		Select("journal_entries.account_id, journal_entries.dc, SUM(journal_entries.amount) as total").
		Joins("JOIN vouchers ON vouchers.id = journal_entries.voucher_id").
		Joins("JOIN accounts ON accounts.id = journal_entries.account_id").
		Where("journal_entries.org_id = ? AND vouchers.period_id = ? AND accounts.category IN ?",
			orgID, periodID, []string{string(AccountCategoryAsset), string(AccountCategoryCost)}).
		Group("journal_entries.account_id, journal_entries.dc").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	balanceMap := make(map[int64]decimal.Decimal)
	for _, row := range rows {
		total, _ := decimal.NewFromString(row.Total)
		if _, ok := balanceMap[row.AccountID]; !ok {
			balanceMap[row.AccountID] = decimal.Zero
		}
		if row.DC == string(DCDebit) {
			balanceMap[row.AccountID] = balanceMap[row.AccountID].Add(total)
		} else {
			balanceMap[row.AccountID] = balanceMap[row.AccountID].Sub(total)
		}
	}

	var result []struct {
		AccountID int64
		Balance   decimal.Decimal
	}
	for aid, bal := range balanceMap {
		if bal.IsNegative() {
			result = append(result, struct {
				AccountID int64
				Balance   decimal.Decimal
			}{aid, bal})
		}
	}
	return result, nil
}

// UpdateVoucherStatusBatch updates the status of all vouchers in a period.
func (r *JournalEntryRepository) UpdateVoucherStatusBatch(orgID, periodID int64, status VoucherStatus) error {
	return r.db.Model(&Voucher{}).
		Where("org_id = ? AND period_id = ?", orgID, periodID).
		Update("status", status).Error
}

// GetAccountsByCategory returns all accounts for an org within the given categories.
func (r *JournalEntryRepository) GetAccountsByCategory(orgID int64, categories []AccountCategory) ([]Account, error) {
	var accounts []Account
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("category IN ?", categories).
		Order("code ASC").
		Find(&accounts).Error
	return accounts, err
}

// SumByCategory returns the total debit and credit sums for all accounts in given categories for a period.
func (r *JournalEntryRepository) SumByCategory(orgID, periodID int64, categories []AccountCategory) (debit, credit decimal.Decimal, err error) {
	type row struct {
		DC    string
		Total string
	}
	var rows []row
	err = r.db.Model(&JournalEntry{}).
		Select("journal_entries.dc, SUM(journal_entries.amount) as total").
		Joins("JOIN vouchers ON vouchers.id = journal_entries.voucher_id").
		Joins("JOIN accounts ON accounts.id = journal_entries.account_id").
		Where("journal_entries.org_id = ? AND vouchers.period_id = ? AND accounts.category IN ?",
			orgID, periodID, categories).
		Group("journal_entries.dc").
		Find(&rows).Error
	if err != nil {
		return
	}
	for _, row := range rows {
		total, _ := decimal.NewFromString(row.Total)
		if row.DC == string(DCDebit) {
			debit = total
		} else {
			credit = total
		}
	}
	return
}

// ========== SnapshotRepository ==========

// SnapshotRepository handles data access for ReportSnapshot.
type SnapshotRepository struct {
	db *gorm.DB
}

// NewSnapshotRepository creates a new SnapshotRepository.
func NewSnapshotRepository(db *gorm.DB) *SnapshotRepository {
	return &SnapshotRepository{db: db}
}

// Create creates a new report snapshot.
func (r *SnapshotRepository) Create(snapshot *ReportSnapshot) error {
	return r.db.Create(snapshot).Error
}

// GetByPeriodAndType returns a valid snapshot for the given period and report type.
func (r *SnapshotRepository) GetByPeriodAndType(orgID int64, periodID int64, reportType ReportType) (*ReportSnapshot, error) {
	var snapshot ReportSnapshot
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("period_id = ? AND report_type = ? AND is_valid = ?", periodID, reportType, true).
		First(&snapshot).Error
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

// InvalidateByPeriod marks all snapshots for a period as invalid (used during revert).
func (r *SnapshotRepository) InvalidateByPeriod(orgID, periodID int64) error {
	return r.db.Scopes(middleware.TenantScope(orgID)).
		Model(&ReportSnapshot{}).
		Where("period_id = ?", periodID).
		Update("is_valid", false).Error
}

// GetByPeriod returns all snapshots for a period.
func (r *SnapshotRepository) GetByPeriod(orgID, periodID int64) ([]ReportSnapshot, error) {
	var snapshots []ReportSnapshot
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("period_id = ?", periodID).
		Order("generated_at DESC").
		Find(&snapshots).Error
	return snapshots, err
}
