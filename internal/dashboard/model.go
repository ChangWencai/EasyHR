package dashboard

// TodoType represents the type of todo item.
type TodoType string

const (
	TodoSocialInsurance TodoType = "social_insurance"
	TodoTax            TodoType = "tax"
	TodoEmployee       TodoType = "employee"
	TodoContract       TodoType = "contract"
	TodoExpense        TodoType = "expense"
	TodoVoucher        TodoType = "voucher"
)

// TodoItem represents a single todo card on the dashboard.
type TodoItem struct {
	Type     TodoType `json:"type"`
	Title    string   `json:"title"`
	Count    int      `json:"count"`
	Deadline string   `json:"deadline,omitempty"`
	Priority int      `json:"priority"`
}

// Overview holds the summary statistics for the dashboard.
type Overview struct {
	EmployeeCount        int    `json:"employee_count"`
	JoinedThisMonth      int    `json:"joined_this_month"`
	LeftThisMonth        int    `json:"left_this_month"`
	SocialInsuranceTotal string `json:"social_insurance_total"`
	PayrollTotal         string `json:"payroll_total"`
}

// DashboardResult is the full dashboard response.
type DashboardResult struct {
	Todos    []TodoItem `json:"todos"`
	Overview Overview   `json:"overview"`
}

// EmployeeDashboardResult holds the employee-specific dashboard statistics.
type EmployeeDashboardResult struct {
	ActiveCount     int     `json:"active_count"`
	JoinedThisMonth int     `json:"joined_this_month"`
	LeftThisMonth   int     `json:"left_this_month"`
	TurnoverRate    float64 `json:"turnover_rate"` // 离职率百分比，保留2位小数
}
