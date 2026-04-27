package dashboard

import (
	"context"
	"fmt"
	"math"
	"sort"

	"golang.org/x/sync/errgroup"
)

// ServiceInterface abstracts DashboardService for handler dependency injection.
type ServiceInterface interface {
	GetDashboard(ctx context.Context, orgID int64) (*DashboardResult, error)
	GetEmployeeDashboard(ctx context.Context, orgID int64) (*EmployeeDashboardResult, error)
	GetTodoStats(ctx context.Context, orgID int64) (*GetTodoStatsResponse, error)
	GetTimeLimitedStats(ctx context.Context, orgID int64) (*GetTimeLimitedStatsResponse, error)
}

// DashboardService aggregates dashboard data from multiple sources.
type DashboardService struct {
	repo DashboardRepository
}

// NewService creates a new DashboardService.
func NewService(repo DashboardRepository) *DashboardService {
	return &DashboardService{repo: repo}
}

// sourceTypeToTodoItem maps a source_type from todo_items table to a dashboard TodoItem.
// Returns nil for unknown source types.
func sourceTypeToTodoItem(sourceType string, count int) *TodoItem {
	if count <= 0 {
		return nil
	}
	switch sourceType {
	case "contract_new", "contract_renew":
		return &TodoItem{Type: TodoContract, Title: "合同到期提醒", Count: count, Priority: 4}
	case "tax_declaration":
		return &TodoItem{Type: TodoTax, Title: "个税申报提醒", Count: count, Priority: 2}
	case "si_payment", "si_change", "si_annual_base", "fund_annual_base":
		return &TodoItem{Type: TodoSocialInsurance, Title: "社保缴费提醒", Count: count, Priority: 1}
	case "employee":
		return &TodoItem{Type: TodoEmployee, Title: "员工入离职待审核", Count: count, Priority: 3}
	case "expense":
		return &TodoItem{Type: TodoExpense, Title: "费用报销待审批", Count: count, Priority: 5}
	case "voucher":
		return &TodoItem{Type: TodoVoucher, Title: "凭证待审核", Count: count, Priority: 6}
	default:
		return nil
	}
}

// GetDashboard returns the dashboard for the given org.
func (s *DashboardService) GetDashboard(ctx context.Context, orgID int64) (*DashboardResult, error) {
	var (
		empStats     struct{ active, joined, left int }
		payrollTotal string
		siTotal      string
		todoStats    []DashboardTodoStat
	)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		active, joined, left, err := s.repo.GetEmployeeStats(ctx, orgID)
		if err != nil {
			return err
		}
		empStats.active = active
		empStats.joined = joined
		empStats.left = left
		return nil
	})

	g.Go(func() error {
		total, err := s.repo.GetPayrollTotal(ctx, orgID)
		if err != nil {
			return err
		}
		payrollTotal = total
		return nil
	})

	g.Go(func() error {
		total, err := s.repo.GetSocialInsuranceTotal(ctx, orgID)
		if err != nil {
			return err
		}
		siTotal = total
		return nil
	})

	g.Go(func() error {
		stats, err := s.repo.GetDashboardTodos(ctx, orgID)
		if err != nil {
			return err
		}
		todoStats = stats
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Map source_type → dashboard TodoItem
	todos := []TodoItem{}
	for _, stat := range todoStats {
		item := sourceTypeToTodoItem(stat.SourceType, stat.Count)
		if item != nil {
			todos = append(todos, *item)
		}
	}

	// Sort by priority ascending
	sort.Slice(todos, func(i, j int) bool {
		return todos[i].Priority < todos[j].Priority
	})

	// Normalize zero strings
	if siTotal == "" || siTotal == "0" || siTotal == "0.00" {
		siTotal = "0"
	}
	if payrollTotal == "" || payrollTotal == "0" || payrollTotal == "0.00" {
		payrollTotal = "0"
	}

	return &DashboardResult{
		Todos: todos,
		Overview: Overview{
			EmployeeCount:        empStats.active,
			JoinedThisMonth:      empStats.joined,
			LeftThisMonth:        empStats.left,
			SocialInsuranceTotal: siTotal,
			PayrollTotal:         payrollTotal,
		},
	}, nil
}

// GetTodoStats returns ring chart stats for all todos.
func (s *DashboardService) GetTodoStats(ctx context.Context, orgID int64) (*GetTodoStatsResponse, error) {
	completed, pending, err := s.repo.GetTodoRingStats(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("get todo ring stats: %w", err)
	}
	total := completed + pending
	percent := 0.0
	if total > 0 {
		percent = math.Round(float64(completed)*100/float64(total)*100) / 100
	}
	return &GetTodoStatsResponse{
		Stats: RingChartStats{
			Completed: completed,
			Pending:   pending,
			Total:     total,
			Percent:   percent,
		},
	}, nil
}

// GetTimeLimitedStats returns ring chart stats for time-limited todos only.
func (s *DashboardService) GetTimeLimitedStats(ctx context.Context, orgID int64) (*GetTimeLimitedStatsResponse, error) {
	completed, pending, err := s.repo.GetTimeLimitedRingStats(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("get time-limited ring stats: %w", err)
	}
	total := completed + pending
	percent := 0.0
	if total > 0 {
		percent = math.Round(float64(completed)*100/float64(total)*100) / 100
	}
	return &GetTimeLimitedStatsResponse{
		Stats: RingChartStats{
			Completed: completed,
			Pending:   pending,
			Total:     total,
			Percent:   percent,
		},
	}, nil
}
// GetEmployeeDashboard returns employee-specific statistics for the employee dashboard.
func (s *DashboardService) GetEmployeeDashboard(ctx context.Context, orgID int64) (*EmployeeDashboardResult, error) {
	active, joined, left, err := s.repo.GetEmployeeStats(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// 离职率 = 离职人数 / (离职人数 + 期末人数) x 100%
	turnoverRate := 0.0
	denominator := float64(left + active)
	if denominator > 0 {
		turnoverRate = float64(left) / denominator * 100
	}

	return &EmployeeDashboardResult{
		ActiveCount:     active,
		JoinedThisMonth: joined,
		LeftThisMonth:   left,
		TurnoverRate:    math.Round(turnoverRate*100) / 100,
	}, nil
}
