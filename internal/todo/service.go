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

// InviteResult 邀请结果
type InviteResult struct {
	URL string `json:"url"`
}

// InviteTodo 创建协办邀请（生成Token）
func (s *Service) InviteTodo(ctx context.Context, orgID int64, todoID int64, userID int64) (*InviteResult, error) {
	// 验证待办存在
	_, err := s.repo.FindTodoByID(ctx, orgID, todoID)
	if err != nil {
		return nil, fmt.Errorf("todo not found")
	}

	token, err := GenerateInviteToken()
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	now := time.Now()
	invite := &TodoInvite{}
	invite.OrgID = orgID
	invite.TodoID = todoID
	invite.Token = token
	invite.Status = InviteStatusPending
	invite.ExpiresAt = now.Add(InviteExpiryDuration)
	invite.CreatedBy = userID
	if err := s.repo.CreateInvite(ctx, invite); err != nil {
		return nil, fmt.Errorf("create invite: %w", err)
	}

	// 返回邀请URL（前端通过 VITE_API_BASE_URL 拼接完整地址）
	inviteURL := fmt.Sprintf("/todo/%d/invite?token=%s", todoID, token)
	return &InviteResult{URL: inviteURL}, nil
}

// TerminateTodo 终止待办任务（状态变为 terminated）
func (s *Service) TerminateTodo(ctx context.Context, orgID int64, todoID int64) error {
	return s.repo.UpdateTodoStatus(ctx, orgID, todoID, TodoStatusTerminated)
}

// VerifyResult Token验证结果
type VerifyResult struct {
	Valid   bool   `json:"valid"`
	Expired bool   `json:"expired"`
	Title   string `json:"title"`
	TodoID  int64  `json:"todo_id"`
}

// VerifyInviteToken 验证协办邀请Token
func (s *Service) VerifyInviteToken(ctx context.Context, token string) (*VerifyResult, error) {
	invite, err := s.repo.FindInviteByToken(ctx, token)
	if err != nil {
		return &VerifyResult{Valid: false}, nil // 找不到Token -> 无效
	}

	if invite.Status == InviteStatusUsed {
		return &VerifyResult{Valid: false}, nil
	}
	if time.Now().After(invite.ExpiresAt) {
		return &VerifyResult{Valid: false, Expired: true}, nil
	}

	// 查询关联待办标题
	todo, err := s.repo.FindTodoByID(ctx, invite.OrgID, invite.TodoID)
	if err != nil {
		return &VerifyResult{Valid: false}, nil
	}

	return &VerifyResult{
		Valid:  true,
		Title:  todo.Title,
		TodoID: invite.TodoID,
	}, nil
}

// SubmitInviteRequest 协办提交请求
type SubmitInviteRequest struct {
	Name   string `json:"name" binding:"required"`
	Phone  string `json:"phone"`
	Remark string `json:"remark"`
}

// SubmitInviteResponse 协办提交响应
type SubmitInviteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SubmitInvite 协办人提交信息（公开接口，无需登录）
func (s *Service) SubmitInvite(ctx context.Context, token string, req *SubmitInviteRequest) (*SubmitInviteResponse, error) {
	// 验证Token
	invite, err := s.repo.FindInviteByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}
	if invite.Status == InviteStatusUsed {
		return nil, fmt.Errorf("link already used")
	}
	if time.Now().After(invite.ExpiresAt) {
		return nil, fmt.Errorf("link expired")
	}

	// 标记邀请已使用
	if err := s.repo.MarkInviteUsed(ctx, invite.ID); err != nil {
		return nil, fmt.Errorf("mark used: %w", err)
	}

	// 当前仅记录协办人已提交，后续可在此扩展实际业务逻辑
	_ = req.Name
	_ = req.Phone
	_ = req.Remark

	return &SubmitInviteResponse{
		Success: true,
		Message: "信息已提交，感谢您的配合",
	}, nil
}

// ScanUrgencyStatus 扫描限时任务更新紧迫状态
func (s *Service) ScanUrgencyStatus(ctx context.Context) (int, error) {
	return s.repo.ScanUrgencyStatus(ctx)
}

// UpdateCarouselActivation 更新轮播图激活状态
func (s *Service) UpdateCarouselActivation(ctx context.Context) (int, error) {
	return s.repo.UpdateCarouselActivation(ctx)
}

// CarouselRequest 创建/更新轮播图请求
type CarouselRequest struct {
	ImageURL  string     `json:"image_url" binding:"required"`
	LinkURL   string     `json:"link_url"`
	SortOrder int        `json:"sort_order"`
	Active    bool       `json:"active"`
	StartAt   *time.Time `json:"start_at"`
	EndAt     *time.Time `json:"end_at"`
}

// CreateCarousel 创建轮播图（最多3张）
func (s *Service) CreateCarousel(ctx context.Context, orgID int64, req *CarouselRequest) (*CarouselItem, error) {
	item := &CarouselItem{
		ImageURL:  req.ImageURL,
		LinkURL:   req.LinkURL,
		SortOrder: req.SortOrder,
		Active:    req.Active,
		StartAt:   req.StartAt,
		EndAt:     req.EndAt,
	}
	item.OrgID = orgID
	if err := s.repo.CreateCarousel(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

// UpdateCarousel 更新轮播图
func (s *Service) UpdateCarousel(ctx context.Context, orgID int64, id int64, req *CarouselRequest) error {
	updates := map[string]interface{}{
		"image_url": req.ImageURL,
		"active":    req.Active,
	}
	if req.LinkURL != "" {
		updates["link_url"] = req.LinkURL
	}
	if req.SortOrder > 0 {
		updates["sort_order"] = req.SortOrder
	}
	if req.StartAt != nil {
		updates["start_at"] = req.StartAt
	}
	if req.EndAt != nil {
		updates["end_at"] = req.EndAt
	}
	return s.repo.UpdateCarousel(ctx, orgID, id, updates)
}

// DeleteCarousel 删除轮播图
func (s *Service) DeleteCarousel(ctx context.Context, orgID int64, id int64) error {
	return s.repo.DeleteCarousel(ctx, orgID, id)
}

// ListAllCarousels 查询所有轮播图（管理用）
func (s *Service) ListAllCarousels(ctx context.Context, orgID int64) ([]CarouselItem, error) {
	return s.repo.ListAllCarousels(ctx, orgID)
}
