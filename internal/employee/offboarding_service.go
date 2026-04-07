package employee

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/wencai/easyhr/internal/common/logger"
	"go.uber.org/zap"
	"gorm.io/datatypes"
)

// SocialInsuranceEventHandler 社保事件处理器接口
// 在 employee 包中定义，由 socialinsurance.Service 实现（避免循环依赖）
type SocialInsuranceEventHandler interface {
	OnEmployeeResigned(orgID, employeeID int64)
}

// OffboardingService 离职管理业务逻辑层
type OffboardingService struct {
	obRepo    *OffboardingRepository
	empRepo   *Repository
	siHandler SocialInsuranceEventHandler
}

// NewOffboardingService 创建离职 Service
func NewOffboardingService(obRepo *OffboardingRepository, empRepo *Repository, siHandler SocialInsuranceEventHandler) *OffboardingService {
	return &OffboardingService{
		obRepo:    obRepo,
		empRepo:   empRepo,
		siHandler: siHandler,
	}
}

// BossResign 老板直接办理离职（立即生效）
// 1. 验证员工存在且未离职
// 2. 创建 Offboarding 记录（type=involuntary）
// 3. 立即更新 Employee.status = resigned
func (s *OffboardingService) BossResign(orgID, userID, employeeID int64, req *BossResignRequest) error {
	// 验证员工存在且未离职
	emp, err := s.empRepo.FindByID(orgID, employeeID)
	if err != nil {
		return fmt.Errorf("员工不存在")
	}
	if emp.Status == StatusResigned {
		return fmt.Errorf("该员工已离职，不可重复办理")
	}

	// 解析离职日期
	resignationDate, err := time.Parse("2006-01-02", req.ResignationDate)
	if err != nil {
		return fmt.Errorf("离职日期格式错误: %w", err)
	}

	// 生成默认交接清单
	checklistItems := defaultChecklistItems()

	// 创建 Offboarding 记录
	ob := &Offboarding{}
	ob.OrgID = orgID
	ob.CreatedBy = userID
	ob.UpdatedBy = userID
	ob.EmployeeID = employeeID
	ob.Type = OffboardingTypeInvoluntary
	ob.ResignationDate = resignationDate
	ob.Reason = req.Reason
	ob.Status = OffboardingStatusPending
	ob.ChecklistItems = checklistItems

	if err := s.obRepo.Create(ob); err != nil {
		return fmt.Errorf("创建离职记录失败: %w", err)
	}

	// 立即更新 Employee 状态为 resigned（老板直接办理立即生效）
	updates := map[string]interface{}{
		"status":             StatusResigned,
		"resignation_date":   resignationDate,
		"resignation_reason": req.Reason,
		"updated_by":         userID,
	}
	if err := s.empRepo.Update(orgID, employeeID, updates); err != nil {
		return fmt.Errorf("更新员工状态失败: %w", err)
	}

	return nil
}

// EmployeeApplyResign 员工申请离职（等待老板审批）
// 1. 验证员工存在且未离职
// 2. 创建 Offboarding 记录（type=voluntary, status=pending）
// 3. 员工状态不变，等待审批通过后才更新
func (s *OffboardingService) EmployeeApplyResign(orgID, employeeID int64, req *EmployeeResignRequest) error {
	// 验证员工存在且未离职
	emp, err := s.empRepo.FindByID(orgID, employeeID)
	if err != nil {
		return fmt.Errorf("员工不存在")
	}
	if emp.Status == StatusResigned {
		return fmt.Errorf("该员工已离职，不可重复申请")
	}

	// 解析离职日期
	resignationDate, err := time.Parse("2006-01-02", req.ResignationDate)
	if err != nil {
		return fmt.Errorf("离职日期格式错误: %w", err)
	}

	// 生成默认交接清单
	checklistItems := defaultChecklistItems()

	// 创建 Offboarding 记录
	ob := &Offboarding{}
	ob.OrgID = orgID
	ob.CreatedBy = employeeID // 员工自己创建
	ob.UpdatedBy = employeeID
	ob.EmployeeID = employeeID
	ob.Type = OffboardingTypeVoluntary
	ob.ResignationDate = resignationDate
	ob.Reason = req.Reason
	ob.Status = OffboardingStatusPending
	ob.ChecklistItems = checklistItems

	if err := s.obRepo.Create(ob); err != nil {
		return fmt.Errorf("创建离职申请失败: %w", err)
	}

	return nil
}

// ApproveResign 审批通过员工离职申请
// 1. 查找 Offboarding 记录
// 2. 验证状态为 pending
// 3. 更新 Offboarding 状态为 approved
// 4. 更新 Employee 状态为 resigned
func (s *OffboardingService) ApproveResign(orgID, approverID, offboardingID int64) error {
	ob, err := s.obRepo.FindByID(orgID, offboardingID)
	if err != nil {
		return fmt.Errorf("离职记录不存在")
	}
	if ob.Status != OffboardingStatusPending {
		return fmt.Errorf("当前状态不可审批（状态: %s）", ob.Status)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":      OffboardingStatusApproved,
		"approved_by": approverID,
		"approved_at": now,
		"updated_by":  approverID,
	}
	if err := s.obRepo.Update(orgID, offboardingID, updates); err != nil {
		return fmt.Errorf("审批离职失败: %w", err)
	}

	// 更新 Employee 状态为 resigned
	empUpdates := map[string]interface{}{
		"status":             StatusResigned,
		"resignation_date":   ob.ResignationDate,
		"resignation_reason": ob.Reason,
		"updated_by":         approverID,
	}
	if err := s.empRepo.Update(orgID, ob.EmployeeID, empUpdates); err != nil {
		return fmt.Errorf("更新员工状态失败: %w", err)
	}

	return nil
}

// UpdateChecklist 更新交接清单
func (s *OffboardingService) UpdateChecklist(orgID, offboardingID int64, items datatypes.JSON) error {
	ob, err := s.obRepo.FindByID(orgID, offboardingID)
	if err != nil {
		return fmt.Errorf("离职记录不存在")
	}
	if ob.Status == OffboardingStatusCompleted {
		return fmt.Errorf("已完成的离职记录不可修改交接清单")
	}

	updates := map[string]interface{}{
		"checklist_items": items,
	}
	if err := s.obRepo.Update(orgID, offboardingID, updates); err != nil {
		return fmt.Errorf("更新交接清单失败: %w", err)
	}

	return nil
}

// CompleteOffboarding 完成交接流程
// 1. 验证状态为 pending 或 approved
// 2. 更新 Offboarding 状态为 completed
// 3. 触发 onEmployeeResigned 事件
func (s *OffboardingService) CompleteOffboarding(orgID, offboardingID int64) error {
	ob, err := s.obRepo.FindByID(orgID, offboardingID)
	if err != nil {
		return fmt.Errorf("离职记录不存在")
	}
	if ob.Status == OffboardingStatusCompleted {
		return fmt.Errorf("离职已完成，不可重复操作")
	}
	if ob.Status != OffboardingStatusPending && ob.Status != OffboardingStatusApproved {
		return fmt.Errorf("当前状态不可完成交接（状态: %s）", ob.Status)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       OffboardingStatusCompleted,
		"completed_at": now,
	}
	if err := s.obRepo.Update(orgID, offboardingID, updates); err != nil {
		return fmt.Errorf("完成交接失败: %w", err)
	}

	// 触发员工离职事件
	s.onEmployeeResigned(orgID, ob.EmployeeID)

	return nil
}

// GetOffboarding 获取离职详情（含员工姓名）
func (s *OffboardingService) GetOffboarding(orgID, offboardingID int64) (*OffboardingDetailResponse, error) {
	ob, err := s.obRepo.FindByID(orgID, offboardingID)
	if err != nil {
		return nil, fmt.Errorf("离职记录不存在")
	}

	emp, err := s.empRepo.FindByID(orgID, ob.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("关联员工不存在")
	}

	return &OffboardingDetailResponse{
		ID:              ob.ID,
		EmployeeID:      ob.EmployeeID,
		EmployeeName:    emp.Name,
		Type:            ob.Type,
		ResignationDate: ob.ResignationDate,
		Reason:          ob.Reason,
		Status:          ob.Status,
		ChecklistItems:  ob.ChecklistItems,
		CompletedAt:     ob.CompletedAt,
		ApprovedBy:      ob.ApprovedBy,
		ApprovedAt:      ob.ApprovedAt,
		CreatedAt:       ob.CreatedAt,
	}, nil
}

// ListOffboardings 离职列表（分页）
func (s *OffboardingService) ListOffboardings(orgID int64, params OffboardingListQueryParams) ([]OffboardingDetailResponse, int64, error) {
	page := params.Page
	pageSize := params.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offboardings, total, err := s.obRepo.List(orgID, params.Status, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("查询离职列表失败: %w", err)
	}

	var responses []OffboardingDetailResponse
	for _, ob := range offboardings {
		emp, empErr := s.empRepo.FindByID(orgID, ob.EmployeeID)
		if empErr != nil {
			continue
		}
		responses = append(responses, OffboardingDetailResponse{
			ID:              ob.ID,
			EmployeeID:      ob.EmployeeID,
			EmployeeName:    emp.Name,
			Type:            ob.Type,
			ResignationDate: ob.ResignationDate,
			Reason:          ob.Reason,
			Status:          ob.Status,
			ChecklistItems:  ob.ChecklistItems,
			CompletedAt:     ob.CompletedAt,
			ApprovedBy:      ob.ApprovedBy,
			ApprovedAt:      ob.ApprovedAt,
			CreatedAt:       ob.CreatedAt,
		})
	}

	return responses, total, nil
}

// onEmployeeResigned 员工离职事件触发（per D-17）
// 调用社保模块创建停缴提醒
func (s *OffboardingService) onEmployeeResigned(orgID, employeeID int64) {
	if s.siHandler != nil {
		s.siHandler.OnEmployeeResigned(orgID, employeeID)
	}
	if logger.Logger != nil {
		logger.Logger.Info("employee resigned event",
			zap.Int64("org_id", orgID),
			zap.Int64("employee_id", employeeID))
	}
}

// defaultChecklistItems 生成默认交接清单（3分类：资产归还/工作交接/权限回收）
func defaultChecklistItems() datatypes.JSON {
	categories := []ChecklistCategory{
		{
			Category: "资产归还",
			Items: []ChecklistItem{
				{Name: "笔记本电脑", Completed: false},
				{Name: "门禁卡", Completed: false},
			},
		},
		{
			Category: "工作交接",
			Items: []ChecklistItem{
				{Name: "项目/任务清单", Completed: false},
				{Name: "文档资料", Completed: false},
			},
		},
		{
			Category: "权限回收",
			Items: []ChecklistItem{
				{Name: "系统账号", Completed: false},
				{Name: "钥匙", Completed: false},
			},
		},
	}

	data, _ := json.Marshal(categories)
	return datatypes.JSON(data)
}
