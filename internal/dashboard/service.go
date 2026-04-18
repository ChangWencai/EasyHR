package dashboard

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"

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

// GetDashboard returns the dashboard for the given org.
func (s *DashboardService) GetDashboard(ctx context.Context, orgID int64) (*DashboardResult, error) {
	var (
		empStats        struct{ active, joined, left int }
		payrollTotal    string
		siTotal         string
		pendingVouchers int
		pendingExpenses int
		taxReminders    int
		contractExp     int
		pendingOffboard int
		pendingInvites  int
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
		count, err := s.repo.GetPendingVouchers(ctx, orgID)
		if err != nil {
			return err
		}
		pendingVouchers = count
		return nil
	})

	g.Go(func() error {
		count, err := s.repo.GetPendingExpenses(ctx, orgID)
		if err != nil {
			return err
		}
		pendingExpenses = count
		return nil
	})

	g.Go(func() error {
		count, err := s.repo.GetTaxReminders(ctx, orgID)
		if err != nil {
			return err
		}
		taxReminders = count
		return nil
	})

	g.Go(func() error {
		count, err := s.repo.GetContractExpirations(ctx, orgID)
		if err != nil {
			return err
		}
		contractExp = count
		return nil
	})

	g.Go(func() error {
		count, err := s.repo.GetPendingOffboardings(ctx, orgID)
		if err != nil {
			return err
		}
		pendingOffboard = count
		return nil
	})

	g.Go(func() error {
		count, err := s.repo.GetPendingInvitations(ctx, orgID)
		if err != nil {
			return err
		}
		pendingInvites = count
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Build todos only for types with count > 0
	todos := []TodoItem{}

	if siTotal != "" {
		if f, err := strconv.ParseFloat(siTotal, 64); err == nil && f > 0 {
			todos = append(todos, TodoItem{
				Type:     TodoSocialInsurance,
				Title:    "社保缴费提醒",
				Count:    1,
				Priority: 1,
			})
		}
	}

	if taxReminders > 0 {
		todos = append(todos, TodoItem{
			Type:     TodoTax,
			Title:    "个税申报提醒",
			Count:    taxReminders,
			Priority: 2,
		})
	}

	employeeCount := pendingOffboard + pendingInvites
	if employeeCount > 0 {
		todos = append(todos, TodoItem{
			Type:     TodoEmployee,
			Title:    "员工入离职待审核",
			Count:    employeeCount,
			Priority: 3,
		})
	}

	if contractExp > 0 {
		todos = append(todos, TodoItem{
			Type:     TodoContract,
			Title:    "合同到期提醒",
			Count:    contractExp,
			Priority: 4,
		})
	}

	if pendingExpenses > 0 {
		todos = append(todos, TodoItem{
			Type:     TodoExpense,
			Title:    "费用报销待审批",
			Count:    pendingExpenses,
			Priority: 5,
		})
	}

	if pendingVouchers > 0 {
		todos = append(todos, TodoItem{
			Type:     TodoVoucher,
			Title:    "凭证待审核",
			Count:    pendingVouchers,
			Priority: 6,
		})
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
