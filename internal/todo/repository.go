package todo

import (
	"context"
	"fmt"
	"time"

	"github.com/wencai/easyhr/internal/common/middleware"
	"gorm.io/gorm"
)

// Repository 数据访问层
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建 Repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ListTodos 查询待办列表（按 is_pinned DESC, sort_order DESC, created_at DESC）
func (r *Repository) ListTodos(ctx context.Context, orgID int64, status string, page, pageSize int) ([]TodoItem, int64, error) {
	var items []TodoItem
	var total int64

	query := r.db.Model(&TodoItem{}).Scopes(middleware.TenantScope(orgID))

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count todos: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := query.
		Order("is_pinned DESC, sort_order DESC, created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("list todos: %w", err)
	}

	return items, total, nil
}

// SearchTodos 关键字搜索（LIKE title 或 creator_name 或 employee_name）
func (r *Repository) SearchTodos(ctx context.Context, orgID int64, keyword string, status string, page, pageSize int) ([]TodoItem, int64, error) {
	var items []TodoItem
	var total int64

	keywordPattern := "%" + keyword + "%"
	query := r.db.Model(&TodoItem{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("title LIKE ? OR creator_name LIKE ? OR employee_name LIKE ?", keywordPattern, keywordPattern, keywordPattern)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count search todos: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := query.
		Order("is_pinned DESC, sort_order DESC, created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("search todos: %w", err)
	}

	return items, total, nil
}

// FilterByDateRange 按时间段筛选（不超过60天）
func (r *Repository) FilterByDateRange(ctx context.Context, orgID int64, startDate, endDate time.Time, status string, page, pageSize int) ([]TodoItem, int64, error) {
	var items []TodoItem
	var total int64

	// Limit to 60 days per TODO-03
	maxDays := int64(60)
	if endDate.Sub(startDate).Hours()/24 > float64(maxDays) {
		endDate = startDate.AddDate(0, 0, int(maxDays))
	}

	query := r.db.Model(&TodoItem{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count date-range todos: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := query.
		Order("is_pinned DESC, sort_order DESC, created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("filter todos by date: %w", err)
	}

	return items, total, nil
}

// PinTodo 置顶/取消置顶待办
func (r *Repository) PinTodo(ctx context.Context, orgID int64, id int64, pinned bool) error {
	result := r.db.Model(&TodoItem{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_pinned":  pinned,
			"sort_order": 0,
		})
	if result.Error != nil {
		return fmt.Errorf("pin todo: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("todo not found")
	}
	return nil
}

// ListAllForExport 导出全部待办（不分页，全量）
func (r *Repository) ListAllForExport(ctx context.Context, orgID int64) ([]TodoItem, error) {
	var items []TodoItem
	err := r.db.Model(&TodoItem{}).
		Scopes(middleware.TenantScope(orgID)).
		Order("is_pinned DESC, sort_order DESC, created_at DESC").
		Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("list all for export: %w", err)
	}
	return items, nil
}

// ListCarousels 查询启用的轮播图（按 sort_order ASC）
func (r *Repository) ListCarousels(ctx context.Context, orgID int64) ([]CarouselItem, error) {
	var items []CarouselItem
	now := time.Now()
	err := r.db.Model(&CarouselItem{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("active = ?", true).
		Where("(start_at IS NULL OR start_at <= ?)", now).
		Where("(end_at IS NULL OR end_at >= ?)", now).
		Order("sort_order ASC").
		Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("list carousels: %w", err)
	}
	return items, nil
}

// CreateTodo 创建待办
func (r *Repository) CreateTodo(ctx context.Context, item *TodoItem) error {
	return r.db.Create(item).Error
}

// FindTodoByID 根据ID查询
func (r *Repository) FindTodoByID(ctx context.Context, orgID int64, id int64) (*TodoItem, error) {
	var item TodoItem
	err := r.db.Model(&TodoItem{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("id = ?", id).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// UpdateTodoStatus 更新待办状态
func (r *Repository) UpdateTodoStatus(ctx context.Context, orgID int64, id int64, status string) error {
	result := r.db.Model(&TodoItem{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("todo not found")
	}
	return nil
}

// ExistsBySource 根据 source 检查是否已存在（幂等创建）
func (r *Repository) ExistsBySource(ctx context.Context, orgID int64, sourceType string, sourceID int64) (bool, error) {
	var count int64
	err := r.db.Model(&TodoItem{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("source_type = ? AND source_id = ?", sourceType, sourceID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateInvite 创建协办邀请
func (r *Repository) CreateInvite(ctx context.Context, invite *TodoInvite) error {
	return r.db.Create(invite).Error
}

// FindInviteByToken 根据Token查找邀请
func (r *Repository) FindInviteByToken(ctx context.Context, token string) (*TodoInvite, error) {
	var invite TodoInvite
	err := r.db.Where("token = ?", token).First(&invite).Error
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

// MarkInviteUsed 标记邀请已使用
func (r *Repository) MarkInviteUsed(ctx context.Context, id int64) error {
	now := time.Now()
	return r.db.Model(&TodoInvite{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":  InviteStatusUsed,
		"used_at": now,
	}).Error
}

// ScanUrgencyStatus 扫描限时任务更新 urgency_status
// D-09-04 规则（精确实现）：
//   expired  : deadline < now - 15 days  (deadline - now < -15)
//   overdue  : deadline in [now - 15 days, now + 7 days] (deadline - now in [-15, 7])
//   normal   : deadline > now + 7 days   (deadline - now > 7)
// 仅更新限时、未完成的记录（排除 completed/terminated）
func (r *Repository) ScanUrgencyStatus(ctx context.Context) (int, error) {
	now := time.Now()
	expiredThreshold := now.AddDate(0, 0, -15)
	overdueStart := now.AddDate(0, 0, -15)
	overdueEnd := now.AddDate(0, 0, 7)
	normalThreshold := now.AddDate(0, 0, 7)

	// 基础过滤器：限时 + 未完成
	activeFilter := "is_time_limited = ? AND status NOT IN ?"
	activeArgs := []interface{}{true, []string{TodoStatusCompleted, TodoStatusTerminated}}

	// 1. expired: deadline < now - 15 days
	result1 := r.db.Model(&TodoItem{}).
		Where(activeFilter+" AND deadline IS NOT NULL AND deadline < ?", append(activeArgs, expiredThreshold)...).
		Update("urgency_status", UrgencyExpired)
	if result1.Error != nil {
		return 0, fmt.Errorf("update expired: %w", result1.Error)
	}

	// 2. overdue: deadline in [now-15, now+7] range
	result2 := r.db.Model(&TodoItem{}).
		Where(activeFilter+" AND deadline IS NOT NULL AND deadline >= ? AND deadline <= ?",
			append(activeArgs, overdueStart, overdueEnd)...).
		Update("urgency_status", UrgencyOverdue)
	if result2.Error != nil {
		return 0, fmt.Errorf("update overdue: %w", result2.Error)
	}

	// 3. normal: deadline > now + 7 days
	result3 := r.db.Model(&TodoItem{}).
		Where(activeFilter+" AND deadline IS NOT NULL AND deadline > ?",
			append(activeArgs, normalThreshold)...).
		Update("urgency_status", UrgencyNormal)
	if result3.Error != nil {
		return 0, fmt.Errorf("update normal: %w", result3.Error)
	}

	return int(result1.RowsAffected + result2.RowsAffected + result3.RowsAffected), nil
}

// UpdateCarouselActivation 根据 start_at/end_at 更新轮播图 active 状态
func (r *Repository) UpdateCarouselActivation(ctx context.Context) (int, error) {
	now := time.Now()
	// 激活：时间区间内且 active=false
	activate := r.db.Model(&CarouselItem{}).
		Where("active = ? AND (start_at IS NULL OR start_at <= ?) AND (end_at IS NULL OR end_at >= ?)",
			false, now, now).
		Update("active", true)
	if activate.Error != nil {
		return 0, activate.Error
	}

	// 停用：超出时间区间
	deactivate := r.db.Model(&CarouselItem{}).
		Where("(start_at IS NOT NULL AND start_at > ?) OR (end_at IS NOT NULL AND end_at < ?)",
			now, now).
		Update("active", false)
	if deactivate.Error != nil {
		return 0, deactivate.Error
	}

	return int(activate.RowsAffected + deactivate.RowsAffected), nil
}

// GenerateMonthlyTodos 生成每月例行待办（个税申报+社保缴费，每月1日触发）
func (r *Repository) GenerateMonthlyTodos(ctx context.Context) error {
	cst := time.FixedZone("CST", 8*3600)
	today := time.Now().In(cst)
	deadline := time.Date(today.Year(), today.Month(), 15, 23, 59, 59, 0, cst)

	// 获取所有活跃组织
	var orgIDs []int64
	r.db.Model(&struct{ ID int64 }{}).Table("organizations").Pluck("id", &orgIDs)

	for _, orgID := range orgIDs {
		// 个税申报
		taxDeadline := deadline
		taxTodo := &TodoItem{}
		taxTodo.OrgID = orgID
		taxTodo.Title = fmt.Sprintf("%d月个税申报，请于15日前完成", today.Month())
		taxTodo.Type = TodoTypeTaxDeclaration
		taxTodo.Deadline = &taxDeadline
		taxTodo.IsTimeLimited = true
		taxTodo.Status = TodoStatusPending
		taxTodo.UrgencyStatus = UrgencyNormal
		taxTodo.SourceType = "tax"
		_ = r.CreateTodo(ctx, taxTodo) // 幂等由 CreateTodo 处理

		// 社保缴费
		siDeadline := deadline
		siTodo := &TodoItem{}
		siTodo.OrgID = orgID
		siTodo.Title = fmt.Sprintf("%d月社保公积金缴费，请于15日前完成", today.Month())
		siTodo.Type = TodoTypeSIPayment
		siTodo.Deadline = &siDeadline
		siTodo.IsTimeLimited = true
		siTodo.Status = TodoStatusPending
		siTodo.UrgencyStatus = UrgencyNormal
		siTodo.SourceType = "socialinsurance"
		_ = r.CreateTodo(ctx, siTodo)
	}

	return nil
}

// GenerateAnnualBaseTodos 生成年度基数调整待办（每年6月15日触发）
func (r *Repository) GenerateAnnualBaseTodos(ctx context.Context) error {
	cst := time.FixedZone("CST", 8*3600)
	today := time.Now().In(cst)

	// 社保基数调整截止日：7月15日
	siDeadline := time.Date(today.Year(), 7, 15, 23, 59, 59, 0, cst)
	// 公积金基数调整截止日：7月15日
	fundDeadline := siDeadline

	var orgIDs []int64
	r.db.Model(&struct{ ID int64 }{}).Table("organizations").Pluck("id", &orgIDs)

	for _, orgID := range orgIDs {
		// 年度社保基数调整
		siTodo := &TodoItem{}
		siTodo.OrgID = orgID
		siTodo.Title = "年度社保基数调整，请于7月15日前完成"
		siTodo.Type = TodoTypeSIAnnualBase
		siTodo.Deadline = &siDeadline
		siTodo.IsTimeLimited = true
		siTodo.Status = TodoStatusPending
		siTodo.UrgencyStatus = UrgencyNormal
		siTodo.SourceType = "socialinsurance"
		_ = r.CreateTodo(ctx, siTodo)

		// 年度公积金基数调整
		fundTodo := &TodoItem{}
		fundTodo.OrgID = orgID
		fundTodo.Title = "年度公积金基数调整，请于7月15日前完成"
		fundTodo.Type = TodoTypeFundAnnualBase
		fundTodo.Deadline = &fundDeadline
		fundTodo.IsTimeLimited = true
		fundTodo.Status = TodoStatusPending
		fundTodo.UrgencyStatus = UrgencyNormal
		fundTodo.SourceType = "socialinsurance"
		_ = r.CreateTodo(ctx, fundTodo)
	}

	return nil
}

// ListAllCarousels 查询企业所有轮播图（管理用，不过滤active状态）
func (r *Repository) ListAllCarousels(ctx context.Context, orgID int64) ([]CarouselItem, error) {
	var items []CarouselItem
	err := r.db.Model(&CarouselItem{}).
		Scopes(middleware.TenantScope(orgID)).
		Order("sort_order DESC, created_at DESC").
		Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("list all carousels: %w", err)
	}
	return items, nil
}

// CreateCarousel 创建轮播图
func (r *Repository) CreateCarousel(ctx context.Context, item *CarouselItem) error {
	// 限制最多3张
	var count int64
	if err := r.db.Model(&CarouselItem{}).Scopes(middleware.TenantScope(item.OrgID)).Count(&count).Error; err != nil {
		return err
	}
	if count >= 3 {
		return fmt.Errorf("轮播图最多3张，当前已有%d张", count)
	}
	return r.db.Create(item).Error
}

// UpdateCarousel 更新轮播图
func (r *Repository) UpdateCarousel(ctx context.Context, orgID int64, id int64, updates map[string]interface{}) error {
	result := r.db.Model(&CarouselItem{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("id = ?", id).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("carousel not found")
	}
	return nil
}

// DeleteCarousel 删除轮播图
func (r *Repository) DeleteCarousel(ctx context.Context, orgID int64, id int64) error {
	result := r.db.Model(&CarouselItem{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("id = ?", id).
		Delete(&CarouselItem{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("carousel not found")
	}
	return nil
}
