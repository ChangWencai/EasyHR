package finance

import (
	"fmt"
	"strings"
	"time"

	"github.com/wencai/easyhr/internal/common/model"
)

// AccountService handles business logic for accounting accounts.
type AccountService struct {
	repo        *AccountRepository
	periodRepo  *PeriodRepository
}

// NewAccountService creates a new AccountService.
func NewAccountService(repo *AccountRepository) *AccountService {
	return &AccountService{repo: repo}
}

// NewAccountServiceWithPeriod creates a new AccountService with a PeriodRepository.
func NewAccountServiceWithPeriod(repo *AccountRepository, periodRepo *PeriodRepository) *AccountService {
	return &AccountService{repo: repo, periodRepo: periodRepo}
}

// GetTree returns the full account tree for an org, grouped by category and nested by parent.
func (s *AccountService) GetTree(orgID int64) ([]*AccountTreeResponse, error) {
	accounts, err := s.repo.GetActiveByOrg(orgID)
	if err != nil {
		return nil, fmt.Errorf("获取科目列表失败: %w", err)
	}

	// Build lookup map by ID
	nodeMap := make(map[int64]*AccountTreeResponse)
	for _, acct := range accounts {
		nodeMap[acct.ID] = &AccountTreeResponse{
			ID:            acct.ID,
			Code:          acct.Code,
			Name:          acct.Name,
			Category:      acct.Category,
			ParentID:      acct.ParentID,
			Level:         acct.Level,
			NormalBalance: acct.NormalBalance,
			IsActive:      acct.IsActive,
			IsSystem:      acct.IsSystem,
			Children:      []*AccountTreeResponse{},
		}
	}

	// Build tree by linking children to parents
	var roots []*AccountTreeResponse
	for _, acct := range accounts {
		node := nodeMap[acct.ID]
		if acct.ParentID == nil {
			roots = append(roots, node)
		} else {
			if parent, ok := nodeMap[*acct.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			} else {
				// Orphan node, treat as root
				roots = append(roots, node)
			}
		}
	}

	return roots, nil
}

// CreateCustomAccount creates a new non-system account.
// Custom account codes must start with "8" per D-08.
func (s *AccountService) CreateCustomAccount(orgID int64, req *CreateAccountRequest) (*Account, error) {
	// D-08: custom accounts must use 8xxxx code range
	if !strings.HasPrefix(req.Code, "8") {
		return nil, &FinanceError{Code: 60211, Err: fmt.Errorf("自定义科目代码必须以8开头")}
	}

	// Check for duplicate code
	existing, err := s.repo.GetByCode(orgID, req.Code)
	if err == nil && existing != nil {
		return nil, &FinanceError{Code: 60212, Err: fmt.Errorf("科目代码 %s 已存在", req.Code)}
	}

	// Determine normal balance from category
	normalBalance := s.normalBalanceForCategory(req.Category)

	account := &Account{
		BaseModel:     model.BaseModel{OrgID: orgID},
		Code:          req.Code,
		Name:          req.Name,
		Category:      req.Category,
		NormalBalance: normalBalance,
		IsActive:      true,
		IsSystem:      false,
		ParentID:      req.ParentID,
		Level:         1,
	}
	if req.ParentID != nil {
		parent, err := s.repo.GetByID(orgID, *req.ParentID)
		if err != nil {
			return nil, &FinanceError{Code: CodeAccountNotFound, Err: fmt.Errorf("父级科目不存在")}
		}
		account.Level = parent.Level + 1
		if account.Level > 3 {
			return nil, &FinanceError{Code: 60213, Err: fmt.Errorf("科目最多支持3级")}
		}
	}

	if err := s.repo.Create(account); err != nil {
		return nil, fmt.Errorf("创建科目失败: %w", err)
	}
	return account, nil
}

// UpdateAccount updates name and/or is_active for an account.
// System accounts cannot be deactivated; is_active cannot be set to false for is_system=true.
func (s *AccountService) UpdateAccount(orgID int64, accountID int64, req *UpdateAccountRequest) (*Account, error) {
	account, err := s.repo.GetByID(orgID, accountID)
	if err != nil {
		return nil, &FinanceError{Code: CodeAccountNotFound, Err: fmt.Errorf("科目不存在或无权访问")}
	}

	if req.Name != "" {
		account.Name = req.Name
	}
	if req.IsActive != nil {
		if account.IsSystem && !*req.IsActive {
			return nil, &FinanceError{Code: 60214, Err: fmt.Errorf("系统预置科目不能禁用")}
		}
		account.IsActive = *req.IsActive
	}

	if err := s.repo.Update(account); err != nil {
		return nil, fmt.Errorf("更新科目失败: %w", err)
	}
	return account, nil
}

// SeedIfEmpty seeds preset accounts if the org has none.
func (s *AccountService) SeedIfEmpty(orgID int64) error {
	return s.repo.SeedIfEmpty(orgID)
}

// GetOrCreateCurrentPeriod returns the current period or creates it.
func (s *AccountService) GetOrCreateCurrentPeriod(orgID int64) (*Period, error) {
	year, month := currentYearMonth()
	return s.periodRepo.GetOrCreate(orgID, year, month)
}

func (s *AccountService) normalBalanceForCategory(cat AccountCategory) NormalBalance {
	switch cat {
	case AccountCategoryAsset, AccountCategoryCost:
		return NormalBalanceDebit
	default:
		return NormalBalanceCredit
	}
}

// currentYearMonth returns the current year and month.
func currentYearMonth() (int, int) {
	now := currentTime()
	return now.Year(), int(now.Month())
}

// currentTime is overridable for testing.
var currentTime = func() time.Time { return time.Now() }

// findAccountByCode is package-internal; used by payroll_adapter.go.
func (s *AccountService) findAccountByCode(orgID int64, code string) (*Account, error) {
	return s.repo.GetByCode(orgID, code)
}
