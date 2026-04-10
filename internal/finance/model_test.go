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
	// TDD RED: Account.NormalBalance not yet implemented in full
	// Placeholder test — implementation will be added in plan 06-01
	t.Errorf("Account.NormalBalance not yet defined")
}

func TestVoucherModel_StatusTransitions(t *testing.T) {
	// TDD RED: Voucher.Status not yet defined
	// Placeholder test — implementation will be added in plan 06-01
	t.Errorf("Voucher.Status not yet defined")
}

func TestJournalEntry_AmountPrecision(t *testing.T) {
	// TDD RED: JournalEntry.Amount uses decimal.Decimal
	// This test verifies precision is preserved through model operations
	db := setupFinanceDB(t)
	org, _ := testutil.CreateTestOrg(db, "Test Org", "91110000123456001X", "Beijing")

	acct := &Account{
		BaseModel:     model.BaseModel{OrgID: org.ID},
		Code:          "1001",
		Name:          "库存现金",
		Category:      "asset",
		NormalBalance: NormalBalanceDebit,
		IsActive:      true,
		IsSystem:      false,
	}
	if err := db.Create(acct).Error; err != nil {
		t.Fatalf("failed to create test account: %v", err)
	}

	// Create a voucher so we have a valid voucher_id for journal entries
	period := &Period{
		BaseModel: model.BaseModel{OrgID: org.ID},
		Year:      2026,
		Month:     4,
		Status:    PeriodStatusOpen,
	}
	if err := db.Create(period).Error; err != nil {
		t.Fatalf("failed to create period: %v", err)
	}
	voucher := &Voucher{
		BaseModel:  model.BaseModel{OrgID: org.ID},
		PeriodID:   period.ID,
		Status:     VoucherStatusDraft,
		Date:       time.Now(),
		VoucherNo:  "202604-0001",
		SourceType: SourceTypeManual,
	}
	if err := db.Create(voucher).Error; err != nil {
		t.Fatalf("failed to create voucher: %v", err)
	}

	// Verify that a decimal amount is stored with full precision
	original := decimal.NewFromFloat(0.1)
	entry := &JournalEntry{
		BaseModel: model.BaseModel{OrgID: org.ID},
		VoucherID: voucher.ID,
		AccountID: acct.ID,
		DC:        DCTypeDebit,
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

	// Amount precision is verified by decimal.Decimal type — this test will
	// be expanded in plan 06-01 to cover arithmetic operations
	t.Errorf("JournalEntry.Amount not yet defined")
}
