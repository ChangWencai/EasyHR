package salary

// ========== 薪资看板 DTO ==========

// SalaryDashboardResponse 薪资看板响应
type SalaryDashboardResponse struct {
	Stats []StatItem `json:"stats"`
}

// StatItem 单个看板指标
type StatItem struct {
	Label          string   `json:"label"`            // 指标名称：应发总额/实发总额/社保公积金/个税总额
	Value          string   `json:"value"`            // 格式化后的金额（如 "12,345.67"）
	TrendPercent   *string  `json:"trend_percent"`    // 环比百分比（如 "+5.23"），无上月数据时为 nil
	TrendDirection string   `json:"trend_direction"`  // up/down/neutral
}
