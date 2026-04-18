package dashboard

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGetDashboardHandler_ValidToken(t *testing.T) {
	// Setup mock service
	mockSvc := &MockDashboardService{
		Result: &DashboardResult{
			Todos: []TodoItem{
				{Type: TodoSocialInsurance, Title: "社保缴费提醒", Count: 2, Priority: 1},
			},
			Overview: Overview{
				EmployeeCount:       5,
				JoinedThisMonth:     1,
				LeftThisMonth:       0,
				SocialInsuranceTotal: "2000.00",
				PayrollTotal:        "15000.00",
			},
		},
	}

	// Create handler
	handler := NewHandler(mockSvc)

	// Setup gin router
	router := gin.New()
	router.GET("/api/v1/dashboard", func(c *gin.Context) {
		// Simulate JWT middleware injecting org_id
		c.Set("org_id", int64(1))
		c.Next()
	}, handler.GetDashboard)

	// Make request
	req, _ := http.NewRequest("GET", "/api/v1/dashboard", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if resp["code"] != float64(0) {
		t.Errorf("expected code 0, got %v", resp["code"])
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data field in response")
	}

	todos, ok := data["todos"].([]interface{})
	if !ok {
		t.Fatal("expected todos array in data")
	}
	if len(todos) != 1 {
		t.Errorf("expected 1 todo, got %d", len(todos))
	}

	overview, ok := data["overview"].(map[string]interface{})
	if !ok {
		t.Fatal("expected overview in data")
	}
	if overview["employee_count"] != float64(5) {
		t.Errorf("expected employee_count 5, got %v", overview["employee_count"])
	}
}

func TestGetDashboardHandler_NoOrgID(t *testing.T) {
	mockSvc := &MockDashboardService{}
	handler := NewHandler(mockSvc)

	router := gin.New()
	router.GET("/api/v1/dashboard", handler.GetDashboard)

	req, _ := http.NewRequest("GET", "/api/v1/dashboard", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Without org_id set, should return error response
	if w.Code == http.StatusOK {
		t.Error("expected non-200 status when org_id not set")
	}
}

func TestGetDashboardHandler_ServiceError(t *testing.T) {
	mockSvc := &MockDashboardService{
		Err: errors.New("service error"),
	}
	handler := NewHandler(mockSvc)

	router := gin.New()
	router.GET("/api/v1/dashboard", func(c *gin.Context) {
		c.Set("org_id", int64(1))
		c.Next()
	}, handler.GetDashboard)

	req, _ := http.NewRequest("GET", "/api/v1/dashboard", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Error("expected non-200 status on service error")
	}
}

// MockDashboardService is a simple mock for testing the handler.
type MockDashboardService struct {
	Result              *DashboardResult
	EmployeeResult      *EmployeeDashboardResult
	Err                  error
}

func (m *MockDashboardService) GetDashboard(ctx context.Context, orgID int64) (*DashboardResult, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Result, nil
}

func (m *MockDashboardService) GetEmployeeDashboard(ctx context.Context, orgID int64) (*EmployeeDashboardResult, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.EmployeeResult, nil
}

func (m *MockDashboardService) GetTodoStats(ctx context.Context, orgID int64) (*GetTodoStatsResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return &GetTodoStatsResponse{}, nil
}
func (m *MockDashboardService) GetTimeLimitedStats(ctx context.Context, orgID int64) (*GetTimeLimitedStatsResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return &GetTimeLimitedStatsResponse{}, nil
}

// Verify Handler works with ServiceInterface
var _ ServiceInterface = (*MockDashboardService)(nil)

func TestGetEmployeeDashboard_Success(t *testing.T) {
	mockSvc := &MockDashboardService{
		EmployeeResult: &EmployeeDashboardResult{
			ActiveCount:     3,
			JoinedThisMonth: 1,
			LeftThisMonth:   1,
			TurnoverRate:    25.00,
		},
	}

	handler := NewHandler(mockSvc)

	router := gin.New()
	router.GET("/api/v1/dashboard/employee-dashboard", func(c *gin.Context) {
		c.Set("org_id", int64(1))
		c.Next()
	}, handler.GetEmployeeDashboard)

	req, _ := http.NewRequest("GET", "/api/v1/dashboard/employee-dashboard", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data field in response")
	}

	if data["active_count"] != float64(3) {
		t.Errorf("expected active_count 3, got %v", data["active_count"])
	}
	if data["joined_this_month"] != float64(1) {
		t.Errorf("expected joined_this_month 1, got %v", data["joined_this_month"])
	}
	if data["left_this_month"] != float64(1) {
		t.Errorf("expected left_this_month 1, got %v", data["left_this_month"])
	}
	if data["turnover_rate"] != 25.00 {
		t.Errorf("expected turnover_rate 25.00, got %v", data["turnover_rate"])
	}
}

func TestGetEmployeeDashboard_ZeroDenominator(t *testing.T) {
	mockSvc := &MockDashboardService{
		EmployeeResult: &EmployeeDashboardResult{
			ActiveCount:     0,
			JoinedThisMonth: 0,
			LeftThisMonth:   0,
			TurnoverRate:    0.0,
		},
	}

	handler := NewHandler(mockSvc)

	router := gin.New()
	router.GET("/api/v1/dashboard/employee-dashboard", func(c *gin.Context) {
		c.Set("org_id", int64(1))
		c.Next()
	}, handler.GetEmployeeDashboard)

	req, _ := http.NewRequest("GET", "/api/v1/dashboard/employee-dashboard", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data field in response")
	}

	if data["active_count"] != float64(0) {
		t.Errorf("expected active_count 0, got %v", data["active_count"])
	}
	if data["turnover_rate"] != 0.0 {
		t.Errorf("expected turnover_rate 0.0 when denominator is zero, got %v", data["turnover_rate"])
	}
}
