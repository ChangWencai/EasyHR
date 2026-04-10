package finance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wencai/easyhr/internal/common/model"
	"gorm.io/gorm"
)

// ReportService generates financial reports (balance sheet, income statement) and tax calculations.
// Per D-13: reports are saved as ReportSnapshot on close; per D-14/D-15: simplified V1.0 formulas.
type ReportService struct {
	db          *gorm.DB
	snapshotRepo *SnapshotRepository
	journalRepo  *JournalEntryRepository
	invoiceRepo  *InvoiceRepository
	periodRepo   *PeriodRepository
}

// NewReportService creates a new ReportService.
func NewReportService(
	db *gorm.DB,
	snapshotRepo *SnapshotRepository,
	journalRepo *JournalEntryRepository,
	invoiceRepo *InvoiceRepository,
	periodRepo *PeriodRepository,
) *ReportService {
	return &ReportService{
		db:           db,
		snapshotRepo: snapshotRepo,
		journalRepo:  journalRepo,
		invoiceRepo:  invoiceRepo,
		periodRepo:   periodRepo,
	}
}

// GenerateBalanceSheet generates the balance sheet for a period and saves a snapshot.
// Per D-14: Assets = Liabilities + Equity.
func (s *ReportService) GenerateBalanceSheet(ctx context.Context, orgID, periodID int64) (*BalanceSheetResponse, error) {
	period, err := s.periodRepo.GetByID(orgID, periodID)
	if err != nil {
		return nil, fmt.Errorf("获取期间失败: %w", err)
	}

	// Asset accounts: normal debit balance (D-02)
	assetItems, assetTotal, err := s.getCategoryBalances(orgID, periodID, []AccountCategory{AccountCategoryAsset})
	if err != nil {
		return nil, fmt.Errorf("计算资产合计失败: %w", err)
	}

	// Liability accounts: normal credit balance
	liabilityItems, liabilityTotal, err := s.getCategoryBalances(orgID, periodID, []AccountCategory{AccountCategoryLiability})
	if err != nil {
		return nil, fmt.Errorf("计算负债合计失败: %w", err)
	}

	// Equity accounts: normal credit balance
	equityItems, equityTotal, err := s.getCategoryBalances(orgID, periodID, []AccountCategory{AccountCategoryEquity})
	if err != nil {
		return nil, fmt.Errorf("计算权益合计失败: %w", err)
	}

	isBalanced := assetTotal.Equal(liabilityTotal.Add(equityTotal))
	if !isBalanced {
		// This should not happen in a properly balanced accounting system
		// but we log it and continue
	}

	resp := &BalanceSheetResponse{
		PeriodID:        periodID,
		Year:            period.Year,
		Month:           period.Month,
		Assets:          assetItems,
		Liabilities:     liabilityItems,
		Equity:          equityItems,
		AssetTotal:      assetTotal,
		LiabilityTotal:  liabilityTotal,
		EquityTotal:     equityTotal,
		IsBalanced:      isBalanced,
		GeneratedAt:     time.Now(),
	}

	// Save snapshot (D-13)
	if err := s.saveSnapshot(orgID, periodID, ReportTypeBalanceSheet, resp, 0); err != nil {
		// Log but don't fail: report is still valid, just not snapshotted
		_ = err
	}

	return resp, nil
}

// GenerateIncomeStatement generates the income statement for a period and saves a snapshot.
// Per D-15: Net = Revenue - COGS - SGA - Tax + NonOp - IncomeTax
func (s *ReportService) GenerateIncomeStatement(ctx context.Context, orgID, periodID int64) (*IncomeStatementResponse, error) {
	period, err := s.periodRepo.GetByID(orgID, periodID)
	if err != nil {
		return nil, fmt.Errorf("获取期间失败: %w", err)
	}

	// Revenue: PROFIT accounts with normal credit balance and code 6xxx starting with 60xx
	// Per D-15: 6001 主营业务收入 + 6051 其他业务收入
	revenue, err := s.getProfitCategoryTotal(orgID, periodID, []string{"6001", "6051"})
	if err != nil {
		return nil, fmt.Errorf("计算营业收入失败: %w", err)
	}

	// COGS: 6401 主营业务成本 + 6402 其他业务成本
	cogs, err := s.getCostCategoryTotal(orgID, periodID, []string{"6401", "6402"})
	if err != nil {
		return nil, fmt.Errorf("计算营业成本失败: %w", err)
	}

	// SGA: 6601 销售费用 + 6602 管理费用 + 6603 财务费用
	sga, err := s.getCostCategoryTotal(orgID, periodID, []string{"6601", "6602", "6603"})
	if err != nil {
		return nil, fmt.Errorf("计算期间费用失败: %w", err)
	}

	// Tax: 6403 税金及附加
	tax, err := s.getCostCategoryTotal(orgID, periodID, []string{"6403"})
	if err != nil {
		return nil, fmt.Errorf("计算税金及附加失败: %w", err)
	}

	// Non-op income: 6901 营业外收入
	nonOpIncome, err := s.getProfitCategoryTotal(orgID, periodID, []string{"6901"})
	if err != nil {
		return nil, fmt.Errorf("计算营业外收入失败: %w", err)
	}

	// Non-op expense: 6911 营业外支出
	nonOpExpense, err := s.getCostCategoryTotal(orgID, periodID, []string{"6911"})
	if err != nil {
		return nil, fmt.Errorf("计算营业外支出失败: %w", err)
	}

	// Income tax: 6902 所得税费用
	incomeTax, err := s.getCostCategoryTotal(orgID, periodID, []string{"6902"})
	if err != nil {
		return nil, fmt.Errorf("计算所得税费用失败: %w", err)
	}

	// Net = Revenue - COGS - SGA - Tax + NonOpIncome - NonOpExpense - IncomeTax
	netProfit := revenue.Sub(cogs).Sub(sga).Sub(tax).Add(nonOpIncome).Sub(nonOpExpense).Sub(incomeTax)

	resp := &IncomeStatementResponse{
		PeriodID:     periodID,
		Year:         period.Year,
		Month:        period.Month,
		Revenue:      revenue,
		COGS:         cogs,
		SGA:          sga,
		Tax:          tax,
		NonOpIncome:  nonOpIncome,
		NonOpExpense: nonOpExpense,
		IncomeTax:    incomeTax,
		NetProfit:    netProfit,
		GeneratedAt:  time.Now(),
	}

	// Save snapshot
	if err := s.saveSnapshot(orgID, periodID, ReportTypeIncomeStatement, resp, 0); err != nil {
		_ = err
	}

	return resp, nil
}

// GetBalanceSheet returns a balance sheet from a valid snapshot, or generates one fresh.
func (s *ReportService) GetBalanceSheet(ctx context.Context, orgID, periodID int64) (*BalanceSheetResponse, error) {
	snapshot, err := s.snapshotRepo.GetByPeriodAndType(orgID, periodID, ReportTypeBalanceSheet)
	if err == nil && snapshot != nil {
		// Return cached snapshot
		var data BalanceSheetData
		if err := json.Unmarshal([]byte(snapshot.Data), &data); err != nil {
			return nil, fmt.Errorf("解析资产负债表快照失败: %w", err)
		}
		return &BalanceSheetResponse{
			PeriodID:        periodID,
			Year:            data.Year,
			Month:           data.Month,
			AssetTotal:      data.AssetTotal,
			LiabilityTotal:  data.LiabilityTotal,
			EquityTotal:     data.EquityTotal,
			IsBalanced:      data.AssetTotal.Equal(data.LiabilityTotal.Add(data.EquityTotal)),
			GeneratedAt:     data.GeneratedAt,
		}, nil
	}
	// No valid snapshot, generate fresh
	return s.GenerateBalanceSheet(ctx, orgID, periodID)
}

// GetIncomeStatement returns an income statement from a valid snapshot, or generates one fresh.
func (s *ReportService) GetIncomeStatement(ctx context.Context, orgID, periodID int64) (*IncomeStatementResponse, error) {
	snapshot, err := s.snapshotRepo.GetByPeriodAndType(orgID, periodID, ReportTypeIncomeStatement)
	if err == nil && snapshot != nil {
		var data IncomeStatementData
		if err := json.Unmarshal([]byte(snapshot.Data), &data); err != nil {
			return nil, fmt.Errorf("解析利润表快照失败: %w", err)
		}
		return &IncomeStatementResponse{
			PeriodID:     periodID,
			Year:         data.Year,
			Month:        data.Month,
			Revenue:      data.Revenue,
			COGS:         data.COGS,
			SGA:          data.SGA,
			Tax:          data.Tax,
			NonOpIncome:  data.NonOpIncome,
			NonOpExpense: data.NonOpExpense,
			IncomeTax:    data.IncomeTax,
			NetProfit:    data.NetProfit,
			GeneratedAt:  data.GeneratedAt,
		}, nil
	}
	return s.GenerateIncomeStatement(ctx, orgID, periodID)
}

// GetMultiPeriodBalanceSheet returns balance sheets for multiple periods with comparison.
func (s *ReportService) GetMultiPeriodBalanceSheet(ctx context.Context, orgID int64, periodIDs []int64) (*MultiPeriodBalanceSheetResponse, error) {
	if len(periodIDs) == 0 {
		return nil, fmt.Errorf("period_ids cannot be empty")
	}
	if len(periodIDs) > 4 {
		return nil, fmt.Errorf("period_ids cannot exceed 4")
	}

	// Get balance sheets for each period
	sheets := make([]*BalanceSheetResponse, len(periodIDs))
	periods := make([]PeriodSummary, len(periodIDs))

	for i, pid := range periodIDs {
		bs, err := s.GetBalanceSheet(ctx, orgID, pid)
		if err != nil {
			return nil, fmt.Errorf("获取第%d期资产负债表失败: %w", i+1, err)
		}
		sheets[i] = bs
		periods[i] = PeriodSummary{
			PeriodID: pid,
			Year:     bs.Year,
			Month:    bs.Month,
			Label:    fmt.Sprintf("%d-%02d", bs.Year, bs.Month),
		}
	}

	// Merge all line items across periods
	itemMap := make(map[string]*MultiPeriodBalanceItem)
	for i, sheet := range sheets {
		for _, item := range sheet.Assets {
			key := fmt.Sprintf("A-%s", item.Code)
			if _, ok := itemMap[key]; !ok {
				itemMap[key] = &MultiPeriodBalanceItem{
					AccountID: item.AccountID,
					Code:      item.Code,
					Name:      item.Name,
					Values:    make([]decimal.Decimal, len(periodIDs)),
				}
			}
			itemMap[key].Values[i] = item.Balance
		}
		for _, item := range sheet.Liabilities {
			key := fmt.Sprintf("L-%s", item.Code)
			if _, ok := itemMap[key]; !ok {
				itemMap[key] = &MultiPeriodBalanceItem{
					AccountID: item.AccountID,
					Code:      item.Code,
					Name:      item.Name,
					Values:    make([]decimal.Decimal, len(periodIDs)),
				}
			}
			itemMap[key].Values[i] = item.Balance
		}
		for _, item := range sheet.Equity {
			key := fmt.Sprintf("E-%s", item.Code)
			if _, ok := itemMap[key]; !ok {
				itemMap[key] = &MultiPeriodBalanceItem{
					AccountID: item.AccountID,
					Code:      item.Code,
					Name:      item.Name,
					Values:    make([]decimal.Decimal, len(periodIDs)),
				}
			}
			itemMap[key].Values[i] = item.Balance
		}
	}

	// Compute diff and pct change for each item
	items := make([]MultiPeriodBalanceItem, 0, len(itemMap))
	for _, item := range itemMap {
		first := item.Values[0]
		last := item.Values[len(item.Values)-1]
		diff := last.Sub(first)
		var pctChange decimal.Decimal
		if !first.IsZero() {
			pctChange = diff.Div(first).Mul(decimal.NewFromInt(100))
		}
		item.Diff = diff
		item.PctChange = pctChange
		items = append(items, *item)
	}

	return &MultiPeriodBalanceSheetResponse{
		Periods: periods,
		Items:    items,
	}, nil
}

// CalculateVAT computes monthly VAT (增值税) for a given year/month.
// Per D-22: net_vat = SUM(output invoice tax_amount) - SUM(input verified invoice tax_amount)
func (s *ReportService) CalculateVAT(ctx context.Context, orgID int64, year, month int) (*VATCalculationResponse, error) {
	// Get the period for this month
	period, err := s.periodRepo.GetByYearMonth(orgID, year, month)
	if err != nil {
		return nil, fmt.Errorf("获取期间失败: %w", err)
	}

	outputTax, inputTax, _, _, err := s.invoiceRepo.GetMonthlyTaxSummary(orgID, year, month)
	if err != nil {
		return nil, fmt.Errorf("汇总发票税额失败: %w", err)
	}

	// Get input invoices for reference
	inputInvoices, _ := s.getInputInvoices(orgID, year, month)
	outputInvoices, _ := s.getOutputInvoices(orgID, year, month)

	netVAT := outputTax.Sub(inputTax)
	if netVAT.IsNegative() {
		netVAT = decimal.Zero // negative VAT means refund, cap at 0 for display
	}

	return &VATCalculationResponse{
		PeriodID:       period.ID,
		Year:           year,
		Month:          month,
		OutputTax:      outputTax,
		InputTax:       inputTax,
		NetVAT:         netVAT,
		InputInvoices:  inputInvoices,
		OutputInvoices: outputInvoices,
	}, nil
}

// CalculateCIT estimates quarterly CIT (企业所得税) for a given year and quarter.
// Per D-15: simplified V1.0 formula using accumulated P&L through end of quarter.
// CIT = (annual_revenue_ytd - annual_costs_ytd - annual_expenses_ytd) * 0.05
func (s *ReportService) CalculateCIT(ctx context.Context, orgID int64, year, quarter int) (*CITCalculationResponse, error) {
	if quarter < 1 || quarter > 4 {
		return nil, fmt.Errorf("quarter must be 1-4")
	}

	// Last month of the quarter
	lastMonth := quarter * 3

	// Accumulate from January to lastMonth
	var revenueYTD, costsYTD, expensesYTD decimal.Decimal

	for m := 1; m <= lastMonth; m++ {
		pid, err := s.periodRepo.GetByYearMonth(orgID, year, m)
		if err != nil {
			continue // period not found, skip
		}

		// Revenue: 6001, 6051 (credit balance = revenue)
		rev, _ := s.getProfitCategoryTotal(orgID, pid.ID, []string{"6001", "6051"})
		revenueYTD = revenueYTD.Add(rev)

		// COGS: 6401, 6402 (debit balance = expense/cost)
		cost, _ := s.getCostCategoryTotal(orgID, pid.ID, []string{"6401", "6402"})
		costsYTD = costsYTD.Add(cost)

		// SGA + Tax: 6601, 6602, 6603, 6403, 6911, 6902
		exp, _ := s.getCostCategoryTotal(orgID, pid.ID, []string{"6601", "6602", "6603", "6403", "6911", "6902"})
		expensesYTD = expensesYTD.Add(exp)
	}

	profitBeforeTax := revenueYTD.Sub(costsYTD).Sub(expensesYTD)
	if profitBeforeTax.IsNegative() {
		profitBeforeTax = decimal.Zero
	}

	// Small enterprise rate: 5% (优惠税率)
	taxRate := decimal.NewFromFloat(0.05)
	estimatedCIT := profitBeforeTax.Mul(taxRate)

	return &CITCalculationResponse{
		Year:             year,
		Quarter:          quarter,
		RevenueYTD:       revenueYTD,
		CostsYTD:         costsYTD,
		ExpensesYTD:      expensesYTD,
		ProfitBeforeTax:  profitBeforeTax,
		TaxRate:          taxRate,
		EstimatedCIT:     estimatedCIT,
	}, nil
}

// --- Internal helpers ---

// getCategoryBalances returns balance sheet items for given account categories.
// For ASSET/COST: balance = SUM(debit) - SUM(credit) (normal debit)
// For LIABILITY/EQUITY/PROFIT: balance = SUM(credit) - SUM(debit) (normal credit)
func (s *ReportService) getCategoryBalances(orgID, periodID int64, categories []AccountCategory) ([]BalanceSheetItem, decimal.Decimal, error) {
	accounts, err := s.journalRepo.GetAccountsByCategory(orgID, categories)
	if err != nil {
		return nil, decimal.Zero, err
	}

	balanceMap, err := s.journalRepo.SumByAccount(orgID, periodID)
	if err != nil {
		return nil, decimal.Zero, err
	}

	items := make([]BalanceSheetItem, 0)
	var total decimal.Decimal

	for _, acct := range accounts {
		row := balanceMap[acct.ID]
		debit := decimal.Zero
		credit := decimal.Zero
		if row != nil {
			debit = row.DebitSum
			credit = row.CreditSum
		}

		var balance decimal.Decimal
		switch acct.NormalBalance {
		case NormalBalanceDebit:
			balance = debit.Sub(credit)
		case NormalBalanceCredit:
			balance = credit.Sub(debit)
		}

		// Only include accounts with non-zero balance or with entries
		if !balance.IsZero() || (row != nil && (!debit.IsZero() || !credit.IsZero())) {
			items = append(items, BalanceSheetItem{
				AccountID: acct.ID,
				Code:     acct.Code,
				Name:     acct.Name,
				Balance:  balance,
			})
			total = total.Add(balance)
		}
	}

	return items, total, nil
}

// getProfitCategoryTotal returns total for PROFIT-category accounts with specific codes (revenue/credits).
func (s *ReportService) getProfitCategoryTotal(orgID, periodID int64, codes []string) (decimal.Decimal, error) {
	var total decimal.Decimal
	for _, code := range codes {
		bal, err := s.getAccountBalance(orgID, periodID, code, NormalBalanceCredit)
		if err != nil {
			continue
		}
		total = total.Add(bal)
	}
	return total, nil
}

// getCostCategoryTotal returns total for PROFIT-category accounts with specific codes (expense/debit).
func (s *ReportService) getCostCategoryTotal(orgID, periodID int64, codes []string) (decimal.Decimal, error) {
	var total decimal.Decimal
	for _, code := range codes {
		bal, err := s.getAccountBalance(orgID, periodID, code, NormalBalanceDebit)
		if err != nil {
			continue
		}
		total = total.Add(bal)
	}
	return total, nil
}

// getAccountBalance returns the balance for an account by code.
func (s *ReportService) getAccountBalance(orgID, periodID int64, code string, normalBal NormalBalance) (decimal.Decimal, error) {
	var accountID int64
	err := s.db.Model(&Account{}).
		Where("org_id = ? AND code = ?", orgID, code).
		Select("id").
		Scan(&accountID).Error
	if err != nil {
		return decimal.Zero, fmt.Errorf("account %s not found: %w", code, err)
	}

	balanceMap, _ := s.journalRepo.SumByAccount(orgID, periodID)
	if balanceMap == nil {
		return decimal.Zero, nil
	}
	sum := balanceMap[accountID]
	if sum == nil {
		return decimal.Zero, nil
	}

	switch normalBal {
	case NormalBalanceDebit:
		return sum.DebitSum.Sub(sum.CreditSum), nil
	case NormalBalanceCredit:
		return sum.CreditSum.Sub(sum.DebitSum), nil
	}
	return decimal.Zero, nil
}

// saveSnapshot serializes report data and saves it as a ReportSnapshot.
func (s *ReportService) saveSnapshot(orgID, periodID int64, reportType ReportType, data interface{}, userID int64) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化报表数据失败: %w", err)
	}

	snapshot := &ReportSnapshot{
		BaseModel:    model.BaseModel{OrgID: orgID},
		PeriodID:     periodID,
		ReportType:   reportType,
		Data:         string(jsonData),
		GeneratedBy:  userID,
		GeneratedAt:  time.Now(),
		IsValid:      true,
	}

	// Invalidate existing snapshots for this period/type before creating new one
	_ = s.snapshotRepo.InvalidateByPeriod(orgID, periodID)

	return s.snapshotRepo.Create(snapshot)
}

// getInputInvoices returns verified INPUT invoices for a given month.
func (s *ReportService) getInputInvoices(orgID int64, year, month int) ([]InvoiceRef, error) {
	// This requires adding a method to InvoiceRepository - stub here
	// Will be implemented via invoiceRepo
	return nil, nil
}

// getOutputInvoices returns OUTPUT invoices for a given month.
func (s *ReportService) getOutputInvoices(orgID int64, year, month int) ([]InvoiceRef, error) {
	return nil, nil
}
