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
