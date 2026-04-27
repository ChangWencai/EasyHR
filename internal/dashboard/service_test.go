package dashboard

import (
	"context"
	"testing"
)

func TestGetDashboard_AllZero(t *testing.T) {
	mock := &MockDashboardRepository{
		EmployeeCount:     0,
		JoinedThisMonth:   0,
		LeftThisMonth:     0,
		PayrollTotal:      "0",
		SocialInsuranceAmt: "0",
		DashboardTodos:    nil,
	}
	svc := NewService(mock)
	result, err := svc.GetDashboard(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Todos) != 0 {
		t.Errorf("expected 0 todos, got %d", len(result.Todos))
	}
	if result.Overview.EmployeeCount != 0 {
		t.Errorf("expected employee count 0, got %d", result.Overview.EmployeeCount)
	}
}

func TestGetDashboard_EmployeeData(t *testing.T) {
	mock := &MockDashboardRepository{
		EmployeeCount:     5,
		JoinedThisMonth:   1,
		LeftThisMonth:     0,
		PayrollTotal:      "0",
		SocialInsuranceAmt: "0",
		DashboardTodos:    nil,
	}
	svc := NewService(mock)
	result, err := svc.GetDashboard(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Todos) != 0 {
		t.Errorf("expected 0 todos with no pending items, got %d", len(result.Todos))
	}
	if result.Overview.EmployeeCount != 5 {
		t.Errorf("expected employee count 5, got %d", result.Overview.EmployeeCount)
	}
}

func TestGetDashboard_PayrollData(t *testing.T) {
	mock := &MockDashboardRepository{
		EmployeeCount:     3,
		JoinedThisMonth:   0,
		LeftThisMonth:     0,
		PayrollTotal:      "15000.00",
		SocialInsuranceAmt: "0",
		DashboardTodos:    nil,
	}
	svc := NewService(mock)
	result, err := svc.GetDashboard(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Todos) != 0 {
		t.Errorf("expected 0 todos with no pending items, got %d", len(result.Todos))
	}
	if result.Overview.PayrollTotal != "15000.00" {
		t.Errorf("expected payroll total '15000.00', got %q", result.Overview.PayrollTotal)
	}
}

func TestGetDashboard_AllModulesWithData(t *testing.T) {
	mock := &MockDashboardRepository{
		EmployeeCount:     10,
		JoinedThisMonth:   2,
		LeftThisMonth:     1,
		PayrollTotal:      "50000.00",
		SocialInsuranceAmt: "8000.00",
		DashboardTodos: []DashboardTodoStat{
			{SourceType: "si_payment", Count: 3},
			{SourceType: "tax_declaration", Count: 1},
			{SourceType: "employee", Count: 2},
			{SourceType: "contract_renew", Count: 5},
			{SourceType: "expense", Count: 2},
			{SourceType: "voucher", Count: 3},
		},
	}
	svc := NewService(mock)
	result, err := svc.GetDashboard(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Todos) != 6 {
		t.Errorf("expected 6 todos, got %d", len(result.Todos))
	}

	for i := 0; i < len(result.Todos)-1; i++ {
		if result.Todos[i].Priority > result.Todos[i+1].Priority {
			t.Errorf("todos not sorted by priority: [%d]=%d vs [%d]=%d",
				i, result.Todos[i].Priority, i+1, result.Todos[i+1].Priority)
		}
	}
}

func TestGetDashboard_RepositoryError(t *testing.T) {
	mock := &ErrorMockRepository{}
	svc := NewService(mock)
	_, err := svc.GetDashboard(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error from error mock repository, got nil")
	}
}

func TestGetDashboard_TodoPriorities(t *testing.T) {
	mock := &MockDashboardRepository{
		EmployeeCount:     5,
		JoinedThisMonth:   0,
		LeftThisMonth:     0,
		PayrollTotal:      "0",
		SocialInsuranceAmt: "1000.00",
		DashboardTodos: []DashboardTodoStat{
			{SourceType: "si_payment", Count: 1},
			{SourceType: "tax_declaration", Count: 1},
			{SourceType: "employee", Count: 2},
			{SourceType: "contract_renew", Count: 1},
			{SourceType: "expense", Count: 1},
			{SourceType: "voucher", Count: 1},
		},
	}
	svc := NewService(mock)
	result, err := svc.GetDashboard(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	priorityMap := make(map[TodoType]int)
	for _, todo := range result.Todos {
		priorityMap[todo.Type] = todo.Priority
	}

	expected := map[TodoType]int{
		TodoSocialInsurance: 1,
		TodoTax:            2,
		TodoEmployee:       3,
		TodoContract:       4,
		TodoExpense:        5,
		TodoVoucher:        6,
	}

	for typ, expectedPriority := range expected {
		if got, ok := priorityMap[typ]; !ok {
			t.Errorf("missing todo type %q", typ)
		} else if got != expectedPriority {
			t.Errorf("todo %q: expected priority %d, got %d", typ, expectedPriority, got)
		}
	}
}

func TestGetDashboard_UnknownSourceType(t *testing.T) {
	mock := &MockDashboardRepository{
		EmployeeCount:     5,
		JoinedThisMonth:   0,
		LeftThisMonth:     0,
		PayrollTotal:      "0",
		SocialInsuranceAmt: "0",
		DashboardTodos: []DashboardTodoStat{
			{SourceType: "unknown_type", Count: 10},
			{SourceType: "tax_declaration", Count: 2},
		},
	}
	svc := NewService(mock)
	result, err := svc.GetDashboard(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Todos) != 1 {
		t.Errorf("expected 1 todo (tax only), got %d", len(result.Todos))
	}
	if result.Todos[0].Type != TodoTax {
		t.Errorf("expected todo type tax, got %q", result.Todos[0].Type)
	}
}

func TestGetDashboard_ZeroCount(t *testing.T) {
	mock := &MockDashboardRepository{
		EmployeeCount:     5,
		JoinedThisMonth:   0,
		LeftThisMonth:     0,
		PayrollTotal:      "0",
		SocialInsuranceAmt: "0",
		DashboardTodos: []DashboardTodoStat{
			{SourceType: "tax_declaration", Count: 0},
			{SourceType: "employee", Count: 2},
		},
	}
	svc := NewService(mock)
	result, err := svc.GetDashboard(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Todos) != 1 {
		t.Errorf("expected 1 todo, got %d", len(result.Todos))
	}
}
