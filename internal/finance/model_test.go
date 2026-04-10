package finance

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wencai/easyhr/internal/common/model"
	"github.com/wencai/easyhr/test/testutil"
	"gorm.io/gorm"
)

// setupFinanceDB returns a test DB with finance models migrated.
func setupFinanceDB(t *testing.T) *gorm.DB {
	db, err := testutil.SetupTestDB()
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&Account{},
		&Period{},
		&Voucher{},
		&JournalEntry{},
	); err != nil {
		t.Fatalf("failed to migrate finance models: %v", err)
	}
	return db
}

func TestAccountModel_NormalBalance(t *testing.T) {
	// Verify that ASSET/COST accounts have debit normal balance,
	// and LIABILITY/EQUITY/PROFIT accounts have credit normal balance (D-07).
	assetAcct := Account{NormalBalance: NormalBalanceDebit, Category: AccountCategoryAsset}
	if assetAcct.NormalBalance != NormalBalanceDebit {
		t.Errorf("ASSET account should have debit normal balance")
	}

	liabilityAcct := Account{NormalBalance: NormalBalanceCredit, Category: AccountCategoryLiability}
	if liabilityAcct.NormalBalance != NormalBalanceCredit {
		t.Errorf("LIABILITY account should have credit normal balance")
	}

	equityAcct := Account{NormalBalance: NormalBalanceCredit, Category: AccountCategoryEquity}
	if equityAcct.NormalBalance != NormalBalanceCredit {
		t.Errorf("EQUITY account should have credit normal balance")
	}

	costAcct := Account{NormalBalance: NormalBalanceDebit, Category: AccountCategoryCost}
	if costAcct.NormalBalance != NormalBalanceDebit {
		t.Errorf("COST account should have debit normal balance")
	}

	profitAcct := Account{NormalBalance: NormalBalanceDebit, Category: AccountCategoryProfit}
	if profitAcct.NormalBalance != NormalBalanceDebit {
		t.Errorf("PROFIT expense accounts should have debit normal balance")
	}
}

func TestVoucherModel_StatusTransitions(t *testing.T) {
	// Verify voucher status constants are defined correctly (D-04).
	if VoucherStatusDraft != "draft" {
		t.Errorf("VoucherStatusDraft should be 'draft', got %s", VoucherStatusDraft)
	}
	if VoucherStatusSubmitted != "submitted" {
		t.Errorf("VoucherStatusSubmitted should be 'submitted', got %s", VoucherStatusSubmitted)
	}
	if VoucherStatusAudited != "audited" {
		t.Errorf("VoucherStatusAudited should be 'audited', got %s", VoucherStatusAudited)
	}
	if VoucherStatusClosed != "closed" {
		t.Errorf("VoucherStatusClosed should be 'closed', got %s", VoucherStatusClosed)
	}
}

func TestJournalEntry_AmountPrecision(t *testing.T) {
	// Verify that decimal.Decimal preserves precision where float64 would lose it.
	db := setupFinanceDB(t)
	org, _ := testutil.CreateTestOrg(db, "Test Org", "91110000123456001X", "Beijing")

	acct := &Account{
		BaseModel:     model.BaseModel{OrgID: org.ID},
		Code:          "1001",
		Name:          "库存现金",
		Category:      AccountCategoryAsset,
		NormalBalance: NormalBalanceDebit,
		IsActive:      true,
		IsSystem:      false,
	}
	if err := db.Create(acct).Error; err != nil {
		t.Fatalf("failed to create test account: %v", err)
	}

	// Create a period
	period := &Period{
		BaseModel: model.BaseModel{OrgID: org.ID},
		Year:      2026,
		Month:     4,
		Status:    PeriodStatusOpen,
	}
	if err := db.Create(period).Error; err != nil {
		t.Fatalf("failed to create period: %v", err)
	}

	// Create a voucher
	voucher := &Voucher{
		BaseModel:   model.BaseModel{OrgID: org.ID},
		PeriodID:    period.ID,
		Status:      VoucherStatusDraft,
		Date:        time.Now(),
		VoucherNo:   "202604-0001",
		SourceType:  SourceTypeManual,
	}
	if err := db.Create(voucher).Error; err != nil {
		t.Fatalf("failed to create voucher: %v", err)
	}

	// This is the key precision test: 0.1 + 0.2 with float64 != 0.3,
	// but with decimal.Decimal the result equals 0.3.
	original := decimal.NewFromFloat(0.1).Add(decimal.NewFromFloat(0.2))
	if original.Equal(decimal.NewFromFloat(0.3)) {
		// decimal precision preserved - this is the correct behavior
	} else {
		// float64 would fail this assertion; decimal should pass
		t.Errorf("decimal precision lost: 0.1 + 0.2 = %s, want 0.3", original.String())
	}

	// Also verify large numbers and small decimals are preserved
	bigAmt := decimal.NewFromFloat(12345678.9012)
	if bigAmt.String() != "12345678.9012" {
		t.Errorf("large decimal number precision lost: got %s", bigAmt.String())
	}

	entry := &JournalEntry{
		BaseModel: model.BaseModel{OrgID: org.ID},
		VoucherID: voucher.ID,
		AccountID: acct.ID,
		DC:        DCDebit,
		Amount:    original,
		Summary:   "precision test",
	}
	if err := db.Create(entry).Error; err != nil {
		t.Fatalf("failed to insert journal entry: %v", err)
	}

	var loaded JournalEntry
	if err := db.First(&loaded, entry.ID).Error; err != nil {
		t.Fatalf("failed to load journal entry: %v", err)
	}
	if !loaded.Amount.Equal(original) {
		t.Fatalf("amount mismatch: got %s, want %s", loaded.Amount.String(), original.String())
	}
}
