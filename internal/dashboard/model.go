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

// RingChartStats holds ring chart data for todo completion stats.
type RingChartStats struct {
	Completed int     `json:"completed"` // 已完成数
	Pending   int     `json:"pending"`   // 待办数
	Total     int     `json:"total"`     // 总数
	Percent   float64 `json:"percent"`   // 完成率百分比，0-100
}

// GetTodoStatsResponse 全事项完成率响应
type GetTodoStatsResponse struct {
	Stats RingChartStats `json:"stats"`
}

// GetTimeLimitedStatsResponse 限时任务完成率响应
type GetTimeLimitedStatsResponse struct {
	Stats RingChartStats `json:"stats"`
}

// DashboardResult is the full dashboard response.
type DashboardResult struct {
	Todos    []TodoItem `json:"todos"`
	Overview Overview   `json:"overview"`
}
