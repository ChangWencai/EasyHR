package finance

import (
	"fmt"
	"time"

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
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Preload("Entries").
		Where(q.Statement).
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
		OrgID:       original.OrgID,
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
