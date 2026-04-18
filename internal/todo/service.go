package todo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Service 待办事项业务逻辑层
type Service struct {
	repo *Repository
}

// NewService 创建 Service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ListTodosRequest 列表查询请求
type ListTodosRequest struct {
	Keyword   string `form:"keyword"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Status    string `form:"status"` // pending/completed/terminated, empty=all
	Page      int    `form:"page,default=1"`
	PageSize  int    `form:"page_size,default=20"`
}

// ListTodosResponse 列表响应
type ListTodosResponse struct {
	Items    []TodoItem `json:"items"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

// ListTodos 查询待办列表
func (s *Service) ListTodos(ctx context.Context, orgID int64, req *ListTodosRequest) (*ListTodosResponse, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var items []TodoItem
	var total int64
	var err error

	if req.Keyword != "" {
		items, total, err = s.repo.SearchTodos(ctx, orgID, req.Keyword, req.Status, page, pageSize)
	} else if req.StartDate != "" && req.EndDate != "" {
		startDate, parseErr := time.Parse("2006-01-02", req.StartDate)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid start_date format, use YYYY-MM-DD")
		}
		endDate, parseErr2 := time.Parse("2006-01-02", req.EndDate)
		if parseErr2 != nil {
			return nil, fmt.Errorf("invalid end_date format, use YYYY-MM-DD")
		}
		// endDate end of day
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		items, total, err = s.repo.FilterByDateRange(ctx, orgID, startDate, endDate, req.Status, page, pageSize)
	} else {
		items, total, err = s.repo.ListTodos(ctx, orgID, req.Status, page, pageSize)
	}

	if err != nil {
		return nil, err
	}

	return &ListTodosResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// PinTodo 置顶/取消置顶
func (s *Service) PinTodo(ctx context.Context, orgID int64, id int64, pinned bool) error {
	return s.repo.PinTodo(ctx, orgID, id, pinned)
}

// ExportTodos 导出待办列表全量数据
func (s *Service) ExportTodos(ctx context.Context, orgID int64) ([]TodoItem, error) {
	return s.repo.ListAllForExport(ctx, orgID)
}

// CreateTodo 创建待办（供各模块调用）
func (s *Service) CreateTodo(ctx context.Context, item *TodoItem) error {
	// 幂等检查
	if item.SourceType != "" && item.SourceID != nil {
		exists, err := s.repo.ExistsBySource(ctx, item.OrgID, item.SourceType, *item.SourceID)
		if err != nil {
			return err
		}
		if exists {
			return nil // 已存在，跳过
		}
	}
	return s.repo.CreateTodo(ctx, item)
}

// ListCarousels 查询启用的轮播图
func (s *Service) ListCarousels(ctx context.Context, orgID int64) ([]CarouselItem, error) {
	return s.repo.ListCarousels(ctx, orgID)
}

// GenerateInviteToken 生成协办邀请Token（32字节hex）
func GenerateInviteToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// ComputeUrgencyStatus 计算紧迫状态（D-09-04规则）
// deadline - now > 7 days: normal
// deadline - now in [-15, 7]: overdue
// deadline - now < -15: expired
func ComputeUrgencyStatus(deadline time.Time, currentStatus string) string {
	if currentStatus == TodoStatusCompleted || currentStatus == TodoStatusTerminated {
		return currentStatus
	}

	now := time.Now()
	daysUntil := int(deadline.Sub(now).Hours() / 24)

	if daysUntil < -15 {
		return UrgencyExpired
	}
	// 已超期(负数) 或 剩余<=7天 -> overdue
	if daysUntil <= 7 {
		return UrgencyOverdue
	}
	return UrgencyNormal
}
