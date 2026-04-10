package finance

import (
	"github.com/wencai/easyhr/internal/common/model"
)

// Account represents an accounting account.
type Account struct {
	model.BaseModel
	OrgID         int64           `gorm:"column:org_id;not null;index:idx_account_org_code,priority:1" json:"-"`
	Code          string          `gorm:"type:varchar(20);not null;index:idx_account_org_code,priority:2" json:"code"`
	Name          string          `gorm:"type:varchar(100);not null" json:"name"`
	Category      AccountCategory `gorm:"type:varchar(20);not null" json:"category"`
	NormalBalance NormalBalance   `gorm:"type:varchar(10);not null" json:"normal_balance"`
	IsActive      bool            `gorm:"default:true" json:"is_active"`
	IsSystem      bool            `gorm:"default:false" json:"is_system"`
	ParentID      *int64          `gorm:"index" json:"parent_id,omitempty"`
	Level         int             `gorm:"default:1" json:"level"`
}

// TableName returns the table name for Account.
func (Account) TableName() string {
	return "accounts"
}

// AccountTreeNode represents an account with nested children for tree display.
type AccountTreeNode struct {
	ID            int64             `json:"id"`
	Code          string            `json:"code"`
	Name          string            `json:"name"`
	Category      AccountCategory   `json:"category"`
	ParentID      *int64           `json:"parent_id,omitempty"`
	Level         int               `json:"level"`
	NormalBalance NormalBalance     `json:"normal_balance"`
	IsActive      bool              `json:"is_active"`
	IsSystem      bool              `json:"is_system"`
	Children      []*AccountTreeNode `json:"children,omitempty"`
}

// PresetAccounts returns the 40+ preset accounts per D-07 of 06-CONTEXT.md.
// These are organized by the five categories of the小企业会计准则:
// ASSET (1xxx, normal debit), LIABILITY (2xxx, normal credit),
// EQUITY (3xxx, normal credit), COST (5xxx, normal debit), PROFIT (6xxx, normal credit).
func PresetAccounts(orgID int64) []Account {
	accounts := []struct {
		code         string
		name         string
		category     AccountCategory
		normal       NormalBalance
		parentCode   string
		level        int
		isSystem     bool
	}{
		// Asset (1001-1999, debit normal)
		{"1001", "库存现金", AccountCategoryAsset, NormalBalanceDebit, "", 1, true},
		{"1002", "银行存款", AccountCategoryAsset, NormalBalanceDebit, "", 1, true},
		{"1012", "其他货币资金", AccountCategoryAsset, NormalBalanceDebit, "", 1, true},
		{"1122", "应收账款", AccountCategoryAsset, NormalBalanceDebit, "", 1, true},
		{"1123", "预付账款", AccountCategoryAsset, NormalBalanceDebit, "", 1, true},
		{"1221", "其他应收款", AccountCategoryAsset, NormalBalanceDebit, "", 1, true},
		{"1405", "原材料", AccountCategoryAsset, NormalBalanceDebit, "", 1, true},
		{"1407", "发出商品", AccountCategoryAsset, NormalBalanceDebit, "", 1, true},
		{"1601", "固定资产", AccountCategoryAsset, NormalBalanceDebit, "", 1, true},
		{"1602", "累计折旧", AccountCategoryAsset, NormalBalanceDebit, "", 1, true},
		{"1701", "无形资产", AccountCategoryAsset, NormalBalanceDebit, "", 1, true},
		{"1901", "长期待摊费用", AccountCategoryAsset, NormalBalanceDebit, "", 1, true},

		// Liability (2001-2999, credit normal)
		{"2001", "短期借款", AccountCategoryLiability, NormalBalanceCredit, "", 1, true},
		{"2201", "应付票据", AccountCategoryLiability, NormalBalanceCredit, "", 1, true},
		{"2202", "应付账款", AccountCategoryLiability, NormalBalanceCredit, "", 1, true},
		{"2203", "预收账款", AccountCategoryLiability, NormalBalanceCredit, "", 1, true},
		{"2241", "其他应付款", AccountCategoryLiability, NormalBalanceCredit, "", 1, true},
		{"2501", "长期借款", AccountCategoryLiability, NormalBalanceCredit, "", 1, true},
		{"2801", "应付职工薪酬", AccountCategoryLiability, NormalBalanceCredit, "", 1, true},
		{"2221", "应交税费", AccountCategoryLiability, NormalBalanceCredit, "", 1, true},

		// Equity (3001-3999, credit normal)
		{"4001", "实收资本", AccountCategoryEquity, NormalBalanceCredit, "", 1, true},
		{"4002", "资本公积", AccountCategoryEquity, NormalBalanceCredit, "", 1, true},
		{"4101", "盈余公积", AccountCategoryEquity, NormalBalanceCredit, "", 1, true},
		{"4103", "本年利润", AccountCategoryEquity, NormalBalanceCredit, "", 1, true},
		{"4104", "利润分配", AccountCategoryEquity, NormalBalanceCredit, "", 1, true},

		// Cost (5001-5999, debit normal)
		{"5001", "生产成本", AccountCategoryCost, NormalBalanceDebit, "", 1, true},
		{"5101", "制造费用", AccountCategoryCost, NormalBalanceDebit, "", 1, true},
		{"5301", "研发支出", AccountCategoryCost, NormalBalanceDebit, "", 1, true},

		// Profit (6001-6999)
		{"6001", "主营业务收入", AccountCategoryProfit, NormalBalanceCredit, "", 1, true},
		{"6051", "其他业务收入", AccountCategoryProfit, NormalBalanceCredit, "", 1, true},
		{"6401", "主营业务成本", AccountCategoryProfit, NormalBalanceDebit, "", 1, true},
		{"6402", "其他业务成本", AccountCategoryProfit, NormalBalanceDebit, "", 1, true},
		{"6403", "税金及附加", AccountCategoryProfit, NormalBalanceDebit, "", 1, true},
		{"6601", "销售费用", AccountCategoryProfit, NormalBalanceDebit, "", 1, true},
		{"6602", "管理费用", AccountCategoryProfit, NormalBalanceDebit, "", 1, true},
		{"6603", "财务费用", AccountCategoryProfit, NormalBalanceDebit, "", 1, true},
		{"6901", "营业外收入", AccountCategoryProfit, NormalBalanceCredit, "", 1, true},
		{"6911", "营业外支出", AccountCategoryProfit, NormalBalanceDebit, "", 1, true},
		{"6902", "所得税费用", AccountCategoryProfit, NormalBalanceDebit, "", 1, true},
	}

	// Add sub-accounts for 6602 管理费用 (D-07 sub-account for payroll)
	subAccts6602 := []struct {
		code       string
		name       string
		normal     NormalBalance
		level      int
	}{
		{"660201", "管理费用-工资", NormalBalanceDebit, 2},
		{"660202", "管理费用-社保", NormalBalanceDebit, 2},
		{"660203", "管理费用-公积金", NormalBalanceDebit, 2},
		{"660204", "管理费用-办公费", NormalBalanceDebit, 2},
		{"660205", "管理费用-差旅费", NormalBalanceDebit, 2},
	}

	// Add sub-accounts for 2801 应付职工薪酬
	subAccts2801 := []struct {
		code       string
		name       string
		normal     NormalBalance
		level      int
	}{
		{"280101", "应付职工薪酬-工资", NormalBalanceCredit, 2},
		{"280102", "应付职工薪酬-社保", NormalBalanceCredit, 2},
		{"280103", "应付职工薪酬-公积金", NormalBalanceCredit, 2},
	}

	result := make([]Account, 0, len(accounts)+len(subAccts6602)+len(subAccts2801))

	// Add level-1 accounts
	for _, a := range accounts {
		result = append(result, Account{
			BaseModel:     model.BaseModel{OrgID: orgID},
			Code:          a.code,
			Name:          a.name,
			Category:      a.category,
			NormalBalance: a.normal,
			IsActive:      true,
			IsSystem:      a.isSystem,
			ParentID:      nil,
			Level:         a.level,
		})
	}

	// Sub-accounts will be resolved by the repository's SeedIfEmpty using code lookup.
	return result
}
