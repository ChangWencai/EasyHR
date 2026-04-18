package socialinsurance

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

// SIDashboardService 社保数据看板聚合服务（D-SI-05/D-SI-06）
type SIDashboardService struct {
	db              *gorm.DB
	paymentRepo     *SIMonthlyPaymentRepository
	recordRepo      *Repository
}

// NewSIDashboardService 创建社保数据看板服务
func NewSIDashboardService(db *gorm.DB, paymentRepo *SIMonthlyPaymentRepository, recordRepo *Repository) *SIDashboardService {
	return &SIDashboardService{db: db, paymentRepo: paymentRepo, recordRepo: recordRepo}
}

// siIndicator 看板指标聚合结果
type siIndicator struct {
	current  decimal.Decimal
	previous decimal.Decimal
}

// GetDashboard 获取社保看板数据（4 指标 + 环比 + 欠缴列表）
// D-SI-05: 应缴总额/单位部分/个人部分/欠缴金额
// D-SI-06: 环比上月百分比
func (s *SIDashboardService) GetDashboard(ctx context.Context, orgID int64, yearMonth string) (*SIDashboardResponse, error) {
	prevYearMonth := prevYM(yearMonth)

	// 活跃状态：normal + pending + overdue（D-SI-06 仅统计活跃记录）
	activeStatuses := []PaymentStatus{PaymentStatusNormal, PaymentStatusPending, PaymentStatusOverdue}
	overdueStatuses := []PaymentStatus{PaymentStatusOverdue}

	var total, company, personal, overdue siIndicator

	g, _ := errgroup.WithContext(ctx)

	// 并发查询 4 个指标
	g.Go(func() error {
		curr, err := s.paymentRepo.SumFieldByOrgAndYearMonth(ctx, orgID, yearMonth, "total_amount", activeStatuses)
		if err != nil {
			return fmt.Errorf("查询当月应缴总额失败: %w", err)
		}
		prev, _ := s.paymentRepo.SumFieldByOrgAndYearMonth(ctx, orgID, prevYearMonth, "total_amount", activeStatuses)
		total = siIndicator{current: curr, previous: prev}
		return nil
	})

	g.Go(func() error {
		curr, err := s.paymentRepo.SumFieldByOrgAndYearMonth(ctx, orgID, yearMonth, "company_amount", activeStatuses)
		if err != nil {
			return fmt.Errorf("查询当月单位部分失败: %w", err)
		}
		prev, _ := s.paymentRepo.SumFieldByOrgAndYearMonth(ctx, orgID, prevYearMonth, "company_amount", activeStatuses)
		company = siIndicator{current: curr, previous: prev}
		return nil
	})

	g.Go(func() error {
		curr, err := s.paymentRepo.SumFieldByOrgAndYearMonth(ctx, orgID, yearMonth, "personal_amount", activeStatuses)
		if err != nil {
			return fmt.Errorf("查询当月个人部分失败: %w", err)
		}
		prev, _ := s.paymentRepo.SumFieldByOrgAndYearMonth(ctx, orgID, prevYearMonth, "personal_amount", activeStatuses)
		personal = siIndicator{current: curr, previous: prev}
		return nil
	})

	g.Go(func() error {
		curr, err := s.paymentRepo.SumFieldByOrgAndYearMonth(ctx, orgID, yearMonth, "total_amount", overdueStatuses)
		if err != nil {
			return fmt.Errorf("查询当月欠缴金额失败: %w", err)
		}
		prev, _ := s.paymentRepo.SumFieldByOrgAndYearMonth(ctx, orgID, prevYearMonth, "total_amount", overdueStatuses)
		overdue = siIndicator{current: curr, previous: prev}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	stats := []SIStatItem{
		buildSIStatItem("当月应缴总额", total.current, total.previous, false),
		buildSIStatItem("单位部分合计", company.current, company.previous, false),
		buildSIStatItem("个人部分合计", personal.current, personal.previous, false),
		buildSIStatItem("欠缴金额", overdue.current, overdue.previous, true),
	}

	// 查询欠缴列表（D-SI-10 红色横幅）
	overduePayments, err := s.paymentRepo.GetOverdueByOrg(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("查询欠缴列表失败: %w", err)
	}

	overdueItems := make([]OverdueItem, 0, len(overduePayments))
	for _, p := range overduePayments {
		employeeName := ""
		// 从参保记录获取员工姓名和城市
		record, rErr := s.recordRepo.FindActiveRecordByEmployee(orgID, int64(p.EmployeeID))
		if rErr == nil && record != nil {
			employeeName = record.EmployeeName
		}
		overdueItems = append(overdueItems, OverdueItem{
			ID:           p.ID,
			EmployeeID:   p.EmployeeID,
			EmployeeName: employeeName,
			YearMonth:    p.YearMonth,
			Amount:       formatDecimal(p.TotalAmount),
		})
	}

	return &SIDashboardResponse{
		Stats:        stats,
		OverdueItems: overdueItems,
	}, nil
}

// buildSIStatItem 构建单个看板指标项（复用 salary/buildStatItem 模式）
func buildSIStatItem(label string, current, previous decimal.Decimal, invertTrend bool) SIStatItem {
	item := SIStatItem{
		Label:          label,
		Value:          formatDecimal(current),
		TrendPercent:   nil,
		TrendDirection: "neutral",
	}

	if !previous.IsZero() {
		diff := current.Sub(previous)
		pct := diff.Div(previous).Mul(decimal.NewFromInt(100))

		pctStr := pct.StringFixed(2)
		item.TrendPercent = &pctStr

		if pct.GreaterThan(decimal.Zero) {
			item.TrendDirection = "up"
		} else if pct.LessThan(decimal.Zero) {
			item.TrendDirection = "down"
		}
	}

	// 欠缴金额：increase = bad（invertTrend=true 时，up 实际是恶化）
	// 前端根据 invertTrend 决定颜色
	_ = invertTrend

	return item
}

// formatDecimal 格式化 decimal 为金额字符串
func formatDecimal(d decimal.Decimal) string {
	return d.StringFixed(2)
}

// prevYM 计算上月 YYYYMM
func prevYM(yearMonth string) string {
	if len(yearMonth) != 6 {
		return yearMonth
	}
	year := int(yearMonth[0]-'0')*1000 + int(yearMonth[1]-'0')*100 + int(yearMonth[2]-'0')*10 + int(yearMonth[3]-'0')
	month := int(yearMonth[4]-'0')*10 + int(yearMonth[5]-'0')

	if month == 1 {
		year--
		month = 12
	} else {
		month--
	}

	return fmt.Sprintf("%04d%02d", year, month)
}
