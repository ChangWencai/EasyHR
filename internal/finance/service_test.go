package finance

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wencai/easyhr/internal/common/model"
	"github.com/wencai/easyhr/test/testutil"
	"gorm.io/gorm"
)

// createTestAccountForService is a local helper to create a test account.
func createTestAccountForService(db *gorm.DB, orgID int64, code, name string, category AccountCategory, normalBalance NormalBalance) (*Account, error) {
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
	// Test that TrialBalance: total debit SUM = total credit SUM across all entries
	db := setupFinanceDB(t)
	org, _ := testutil.CreateTestOrg(db, "Test Org TB", "91110000123456002X", "Beijing")
	period, err := createTestPeriodForService(db, org.ID, 2026, 5)
	if err != nil {
		t.Fatalf("failed to create period: %v", err)
	}

	// Create two asset accounts
	asset1, err := createTestAccountForService(db, org.ID, "1001", "库存现金", AccountCategoryAsset, NormalBalanceDebit)
	if err != nil {
		t.Fatalf("failed to create asset1: %v", err)
	}
	asset2, err := createTestAccountForService(db, org.ID, "1002", "银行存款", AccountCategoryAsset, NormalBalanceDebit)
	if err != nil {
		t.Fatalf("failed to create asset2: %v", err)
	}
	// Create one liability account
	liability, err := createTestAccountForService(db, org.ID, "2202", "应付账款", AccountCategoryLiability, NormalBalanceCredit)
	if err != nil {
		t.Fatalf("failed to create liability: %v", err)
	}

	// Voucher 1: DEBIT 1001=500, DEBIT 1002=300, CREDIT 2202=800
	entries1 := []JournalEntry{
		{BaseModel: model.BaseModel{OrgID: org.ID}, AccountID: asset1.ID, DC: DCDebit, Amount: decimal.NewFromInt(500), Summary: "提现"},
		{BaseModel: model.BaseModel{OrgID: org.ID}, AccountID: asset2.ID, DC: DCDebit, Amount: decimal.NewFromInt(300), Summary: "转账"},
		{BaseModel: model.BaseModel{OrgID: org.ID}, AccountID: liability.ID, DC: DCCredit, Amount: decimal.NewFromInt(800), Summary: "应付"},
	}
	v1, err := createTestVoucherForService(db, org.ID, period.ID, VoucherStatusAudited, entries1)
	if err != nil {
		t.Fatalf("failed to create voucher1: %v", err)
	}
	_ = v1

	// Voucher 2: DEBIT 2202=200, CREDIT 1001=200 (repayment)
	entries2 := []JournalEntry{
		{BaseModel: model.BaseModel{OrgID: org.ID}, AccountID: liability.ID, DC: DCDebit, Amount: decimal.NewFromInt(200), Summary: "还款"},
		{BaseModel: model.BaseModel{OrgID: org.ID}, AccountID: asset1.ID, DC: DCCredit, Amount: decimal.NewFromInt(200), Summary: "还款"},
	}
	v2, err := createTestVoucherForService(db, org.ID, period.ID, VoucherStatusAudited, entries2)
	if err != nil {
		t.Fatalf("failed to create voucher2: %v", err)
	}
	_ = v2

	// Build services manually (no gorm.DB scoping in tests)
	accountRepo := NewAccountRepository(db)
	periodRepo := NewPeriodRepository(db)
	journalRepo := NewJournalEntryRepository(db)
	bookSvc := NewBookService(db, accountRepo, journalRepo, periodRepo)

	result, err := bookSvc.GetTrialBalance(t.Context(), org.ID, period.ID)
	if err != nil {
		t.Fatalf("GetTrialBalance failed: %v", err)
	}

	if !result.IsBalanced {
		t.Errorf("trial balance is not balanced: totalDebit=%s, totalCredit=%s",
			result.TotalDebit.String(), result.TotalCredit.String())
	}

	if !result.TotalDebit.Equal(result.TotalCredit) {
		t.Errorf("SUM(debit)=%s != SUM(credit)=%s", result.TotalDebit.String(), result.TotalCredit.String())
	}

	// After voucher2, 1001 balance should be 300 (500-200), 1002=300, 2202=600
	for _, item := range result.Items {
		switch item.Code {
		case "1001":
			if !item.Balance.Equal(decimal.NewFromInt(300)) {
				t.Errorf("1001 balance = %s, want 300", item.Balance.String())
			}
		case "1002":
			if !item.Balance.Equal(decimal.NewFromInt(300)) {
				t.Errorf("1002 balance = %s, want 300", item.Balance.String())
			}
		case "2202":
			if !item.Balance.Equal(decimal.NewFromInt(600)) {
				t.Errorf("2202 balance = %s, want 600", item.Balance.String())
			}
		}
	}
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
	assetAcct, err := createTestAccountForService(db, org.ID, "1001", "库存现金", AccountCategoryAsset, NormalBalanceDebit)
	if err != nil {
		t.Fatalf("failed to create asset account: %v", err)
	}
	// Liability account (normal credit balance)
	liabilityAcct, err := createTestAccountForService(db, org.ID, "2201", "应付账款", AccountCategoryLiability, NormalBalanceCredit)
	if err != nil {
		t.Fatalf("failed to create liability account: %v", err)
	}
	// Owners equity account (normal credit balance)
	equityAcct, err := createTestAccountForService(db, org.ID, "4001", "实收资本", AccountCategoryEquity, NormalBalanceCredit)
	if err != nil {
		t.Fatalf("failed to create equity account: %v", err)
	}

	// Initial balance: Asset=1000, Liability=400, Equity=600  =>  1000=1000
	entries := []JournalEntry{
		{BaseModel: model.BaseModel{OrgID: org.ID}, AccountID: assetAcct.ID, DC: DCDebit, Amount: decimal.NewFromInt(1000), Summary: "资产借方"},
		{BaseModel: model.BaseModel{OrgID: org.ID}, AccountID: liabilityAcct.ID, DC: DCCredit, Amount: decimal.NewFromInt(400), Summary: "负债贷方"},
		{BaseModel: model.BaseModel{OrgID: org.ID}, AccountID: equityAcct.ID, DC: DCCredit, Amount: decimal.NewFromInt(600), Summary: "权益贷方"},
	}
	voucher, err := createTestVoucherForService(db, org.ID, period.ID, VoucherStatusAudited, entries)
	if err != nil {
		t.Fatalf("failed to create voucher: %v", err)
	}
	_ = voucher

	// Build services
	journalRepo := NewJournalEntryRepository(db)
	snapshotRepo := NewSnapshotRepository(db)
	periodRepo := NewPeriodRepository(db)
	reportSvc := NewReportService(db, snapshotRepo, journalRepo, nil, periodRepo)

	bs, err := reportSvc.GenerateBalanceSheet(context.Background(), org.ID, period.ID)
	if err != nil {
		t.Fatalf("GenerateBalanceSheet failed: %v", err)
	}

	if !bs.IsBalanced {
		t.Errorf("balance sheet equation failed: Assets=%s != Liabilities+Equity=%s",
			bs.AssetTotal.String(), bs.LiabilityTotal.Add(bs.EquityTotal).String())
	}

	// Verify actual amounts
	if !bs.AssetTotal.Equal(decimal.NewFromInt(1000)) {
		t.Errorf("AssetTotal=%s, want 1000", bs.AssetTotal.String())
	}
	if !bs.LiabilityTotal.Equal(decimal.NewFromInt(400)) {
		t.Errorf("LiabilityTotal=%s, want 400", bs.LiabilityTotal.String())
	}
	if !bs.EquityTotal.Equal(decimal.NewFromInt(600)) {
		t.Errorf("EquityTotal=%s, want 600", bs.EquityTotal.String())
	}
}
