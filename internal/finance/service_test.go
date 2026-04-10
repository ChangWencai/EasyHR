package finance

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wencai/easyhr/internal/common/model"
	"github.com/wencai/easyhr/test/testutil"
	"gorm.io/gorm"
)

// createTestAccountForService is a local helper to create a test account.
func createTestAccountForService(db *gorm.DB, orgID int64, code, name, category string, normalBalance NormalBalance) (*Account, error) {
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

// createTestPeriodForService is a local helper to create a test accounting period.
func createTestPeriodForService(db *gorm.DB, orgID int64, year, month int) (*Period, error) {
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

// createTestVoucherForService is a local helper to create a test voucher with journal entries.
func createTestVoucherForService(db *gorm.DB, orgID int64, periodID int64, status VoucherStatus, entries []JournalEntry) (*Voucher, error) {
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

func TestTrialBalance_CalculatesCorrectly(t *testing.T) {
	// TDD RED: BookService not yet implemented
	// Placeholder test — implementation will be added in plan 06-03
	t.Errorf("BookService not yet implemented")
}

func TestBalanceSheet_EquationHolds(t *testing.T) {
	// TDD RED: ReportService not yet implemented
	// Placeholder test — implementation will be added in plan 06-03
	// Verifies: Assets = Liabilities + OwnersEquity for a simple balanced dataset
	db := setupFinanceDB(t)
	org, _ := testutil.CreateTestOrg(db, "Test Org BS", "91110000123456001X", "Beijing")
	period, err := createTestPeriodForService(db, org.ID, 2026, 4)
	if err != nil {
		t.Fatalf("failed to create period: %v", err)
	}

	// Asset account (normal debit balance)
	assetAcct, err := createTestAccountForService(db, org.ID, "1001", "库存现金", "asset", NormalBalanceDebit)
	if err != nil {
		t.Fatalf("failed to create asset account: %v", err)
	}
	// Liability account (normal credit balance)
	liabilityAcct, err := createTestAccountForService(db, org.ID, "2201", "应付账款", "liability", NormalBalanceCredit)
	if err != nil {
		t.Fatalf("failed to create liability account: %v", err)
	}
	// Owners equity account (normal credit balance)
	equityAcct, err := createTestAccountForService(db, org.ID, "4001", "实收资本", "equity", NormalBalanceCredit)
	if err != nil {
		t.Fatalf("failed to create equity account: %v", err)
	}

	// Initial balance: Asset=1000, Liability=400, Equity=600  =>  1000=1000
	entries := []JournalEntry{
		{BaseModel: model.BaseModel{OrgID: org.ID}, AccountID: assetAcct.ID, DC: DCTypeDebit, Amount: decimal.NewFromInt(1000), Summary: "资产借方"},
		{BaseModel: model.BaseModel{OrgID: org.ID}, AccountID: liabilityAcct.ID, DC: DCTypeCredit, Amount: decimal.NewFromInt(400), Summary: "负债贷方"},
		{BaseModel: model.BaseModel{OrgID: org.ID}, AccountID: equityAcct.ID, DC: DCTypeCredit, Amount: decimal.NewFromInt(600), Summary: "权益贷方"},
	}
	voucher, err := createTestVoucherForService(db, org.ID, period.ID, VoucherStatusAudited, entries)
	if err != nil {
		t.Fatalf("failed to create voucher: %v", err)
	}
	_ = voucher // ReportService will query journal entries by period
	t.Errorf("ReportService not yet implemented")
}
