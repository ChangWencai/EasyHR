package finance

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wencai/easyhr/internal/common/model"
	"github.com/wencai/easyhr/test/testutil"
	"gorm.io/gorm"
)

// createTestAccount is a local helper to create a test account.
func createTestAccount(db *gorm.DB, orgID int64, code, name, category string, normalBalance NormalBalance) (*Account, error) {
	acct := &Account{
		BaseModel:     model.BaseModel{OrgID: orgID},
		Code:          code,
		Name:          name,
		Category:      category,
		NormalBalance: normalBalance,
		IsActive:      true,
		IsSystem:      false,
	}
	if err := db.Create(acct).Error; err != nil {
		return nil, err
	}
	return acct, nil
}

// createTestPeriod is a local helper to create a test accounting period.
func createTestPeriod(db *gorm.DB, orgID int64, year, month int) (*Period, error) {
	period := &Period{
		BaseModel: model.BaseModel{OrgID: orgID},
		Year:      year,
		Month:     month,
		Status:    PeriodStatusOpen,
	}
	if err := db.Create(period).Error; err != nil {
		return nil, err
	}
	return period, nil
}

// createTestVoucher is a local helper to create a test voucher with journal entries.
func createTestVoucher(db *gorm.DB, orgID int64, periodID int64, status VoucherStatus, entries []JournalEntry) (*Voucher, error) {
	voucher := &Voucher{
		BaseModel:  model.BaseModel{OrgID: orgID},
		PeriodID:   periodID,
		Status:     status,
		Date:       time.Now(),
		VoucherNo:  "",
		SourceType: SourceTypeManual,
		Entries:    entries,
	}
	if err := db.Create(voucher).Error; err != nil {
		return nil, err
	}
	return voucher, nil
}

func TestCreateVoucher_BalancedEntries(t *testing.T) {
	// TDD RED: VoucherService.CreateVoucher not yet implemented
	// Placeholder test — implementation will be added in plan 06-01
	t.Errorf("VoucherService.CreateVoucher not yet implemented")
}

func TestCreateVoucher_UnbalancedEntries_ReturnsError(t *testing.T) {
	// TDD RED: VoucherService.CreateVoucher not yet implemented
	// Placeholder test — implementation will be added in plan 06-01
	// Verifies that SUM(debit) != SUM(credit) returns error code 60201
	t.Errorf("VoucherService.CreateVoucher not yet implemented")
}

func TestReverseVoucher_FlipsDC(t *testing.T) {
	// TDD RED: VoucherService.ReverseVoucher not yet implemented
	// Placeholder test — implementation will be added in plan 06-01
	// Verifies: reversal entry flips DEBIT↔CREDIT, amount unchanged, summary annotated
	db := setupFinanceDB(t)
	org, _ := testutil.CreateTestOrg(db, "Test Org Reverse", "91110000123456001X", "Beijing")
	period, err := createTestPeriod(db, org.ID, 2026, 4)
	if err != nil {
		t.Fatalf("failed to create period: %v", err)
	}
	acct1, _ := createTestAccount(db, org.ID, "1001", "库存现金", "asset", NormalBalanceDebit)
	acct2, _ := createTestAccount(db, org.ID, "6001", "主营业务收入", "revenue", NormalBalanceCredit)

	amt := decimal.NewFromInt(1000)
	entries := []JournalEntry{
		{BaseModel: model.BaseModel{OrgID: org.ID}, AccountID: acct1.ID, DC: DCTypeDebit, Amount: amt, Summary: "借方"},
		{BaseModel: model.BaseModel{OrgID: org.ID}, AccountID: acct2.ID, DC: DCTypeCredit, Amount: amt, Summary: "贷方"},
	}
	voucher, err := createTestVoucher(db, org.ID, period.ID, VoucherStatusAudited, entries)
	if err != nil {
		t.Fatalf("failed to create test voucher: %v", err)
	}
	_ = voucher // ReverseVoucher will use this voucher ID
	t.Errorf("VoucherService.ReverseVoucher not yet implemented")
}
