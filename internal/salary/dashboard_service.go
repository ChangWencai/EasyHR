package salary

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

// SalaryDashboardService 薪资看板聚合服务
type SalaryDashboardService struct {
	db *gorm.DB
}

// NewDashboardService 创建薪资看板服务
func NewDashboardService(db *gorm.DB) *SalaryDashboardService {
	return &SalaryDashboardService{db: db}
}

// dashboardIndicator 看板指标聚合结果
type dashboardIndicator struct {
	current float64
	previous float64
}

// GetDashboard 获取薪资看板数据（4 指标 + 环比）
func (s *SalaryDashboardService) GetDashboard(ctx context.Context, orgID int64, year, month int) (*SalaryDashboardResponse, error) {
	prevYear, prevMonth := prevYearMonth(year, month)

	var gross, net, si, tax dashboardIndicator

	g, _ := errgroup.WithContext(ctx)

	// 并发查询 4 个指标
	g.Go(func() error {
		curr, err := s.sumPayrollField(orgID, year, month, "gross_income")
		if err != nil {
			return fmt.Errorf("查询应发总额失败: %w", err)
		}
		prev, _ := s.sumPayrollField(orgID, prevYear, prevMonth, "gross_income")
		gross = dashboardIndicator{current: curr, previous: prev}
		return nil
	})

	g.Go(func() error {
		curr, err := s.sumPayrollField(orgID, year, month, "net_income")
		if err != nil {
			return fmt.Errorf("查询实发总额失败: %w", err)
		}
		prev, _ := s.sumPayrollField(orgID, prevYear, prevMonth, "net_income")
		net = dashboardIndicator{current: curr, previous: prev}
		return nil
	})

	g.Go(func() error {
		curr, err := s.sumPayrollField(orgID, year, month, "si_deduction")
		if err != nil {
			return fmt.Errorf("查询社保公积金总额失败: %w", err)
		}
		prev, _ := s.sumPayrollField(orgID, prevYear, prevMonth, "si_deduction")
		si = dashboardIndicator{current: curr, previous: prev}
		return nil
	})

	g.Go(func() error {
		curr, err := s.sumPayrollField(orgID, year, month, "tax")
		if err != nil {
			return fmt.Errorf("查询个税总额失败: %w", err)
		}
		prev, _ := s.sumPayrollField(orgID, prevYear, prevMonth, "tax")
		tax = dashboardIndicator{current: curr, previous: prev}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	stats := []StatItem{
		buildStatItem("应发总额", gross.current, gross.previous),
		buildStatItem("实发总额", net.current, net.previous),
		buildStatItem("社保公积金", si.current, si.previous),
		buildStatItem("个税总额", tax.current, tax.previous),
	}

	return &SalaryDashboardResponse{Stats: stats}, nil
}

// sumPayrollField 聚合查询某字段总和
func (s *SalaryDashboardService) sumPayrollField(orgID int64, year, month int, field string) (float64, error) {
	var result float64
	err := s.db.Model(&PayrollRecord{}).
		Where("org_id = ? AND year = ? AND month = ? AND status IN ?",
			orgID, year, month,
			[]string{PayrollStatusCalculated, PayrollStatusConfirmed, PayrollStatusPaid}).
		Select(fmt.Sprintf("COALESCE(SUM(%s), 0)", field)).
		Scan(&result).Error
	return result, err
}

// buildStatItem 构建单个看板指标项
func buildStatItem(label string, current, previous float64) StatItem {
	item := StatItem{
		Label:          label,
		Value:          formatMoney(current),
		TrendPercent:   nil,
		TrendDirection: "neutral",
	}

	if previous > 0 {
		curr := decimal.NewFromFloat(current)
		prev := decimal.NewFromFloat(previous)
		diff := curr.Sub(prev)
		pct := diff.Div(prev).Mul(decimal.NewFromInt(100))

		pctStr := pct.StringFixed(2)
		item.TrendPercent = &pctStr

		if pct.GreaterThan(decimal.Zero) {
			item.TrendDirection = "up"
		} else if pct.LessThan(decimal.Zero) {
			item.TrendDirection = "down"
		}
	}

	return item
}

// formatMoney 格式化金额为千分位字符串
func formatMoney(val float64) string {
	d := decimal.NewFromFloat(val)
	return d.StringFixed(2)
}

// prevYearMonth 计算上月年份和月份
func prevYearMonth(year, month int) (int, int) {
	if month == 1 {
		return year - 1, 12
	}
	return year, month - 1
}

// init 确保时间包被引用
var _ = time.Second
