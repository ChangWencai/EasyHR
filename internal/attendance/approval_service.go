package attendance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/qmuntal/stateless"
	"github.com/wencai/easyhr/internal/common/model"
)

// ApprovalStateMachine 基于 qmuntal/stateless 的审批状态机
// 状态转换：draft → pending → approved / rejected / cancelled / timeout
type ApprovalStateMachine struct {
	sm       *stateless.StateMachine
	approval *Approval
}

// NewApprovalStateMachine 创建审批状态机
func NewApprovalStateMachine(approval *Approval) *ApprovalStateMachine {
	sm := stateless.NewStateMachine(stateless.State(approval.Status))

	sm.Configure(stateless.State(ApprovalStatusDraft)).
		Permit("submit", stateless.State(ApprovalStatusPending))

	sm.Configure(stateless.State(ApprovalStatusPending)).
		Permit("approve", stateless.State(ApprovalStatusApproved)).
		Permit("reject", stateless.State(ApprovalStatusRejected)).
		Permit("cancel", stateless.State(ApprovalStatusCancelled)).
		Permit("timeout", stateless.State(ApprovalStatusTimeout))

	sm.Configure(stateless.State(ApprovalStatusApproved))
	sm.Configure(stateless.State(ApprovalStatusRejected))
	sm.Configure(stateless.State(ApprovalStatusCancelled))
	sm.Configure(stateless.State(ApprovalStatusTimeout))

	return &ApprovalStateMachine{sm: sm, approval: approval}
}

// CanCancel 检查申请人是否可以取消
func (asm *ApprovalStateMachine) CanCancel(requesterID int64) bool {
	if asm.approval.Status != ApprovalStatusPending {
		return false
	}
	return asm.approval.EmployeeID == requesterID
}

// Submit 提交申请
func (asm *ApprovalStateMachine) Submit() error {
	canFire, err := asm.sm.CanFire("submit")
	if err != nil {
		return fmt.Errorf("状态检查失败: %w", err)
	}
	if !canFire {
		return fmt.Errorf("当前状态不允许提交")
	}
	if err := asm.sm.Fire("submit"); err != nil {
		return fmt.Errorf("提交失败: %w", err)
	}
	asm.approval.Status = ApprovalStatusPending
	return nil
}

// Approve 审批通过
func (asm *ApprovalStateMachine) Approve(approverID int64) error {
	canFire, err := asm.sm.CanFire("approve")
	if err != nil {
		return fmt.Errorf("状态检查失败: %w", err)
	}
	if !canFire {
		return fmt.Errorf("当前状态不允许审批")
	}
	if err := asm.sm.Fire("approve"); err != nil {
		return fmt.Errorf("审批失败: %w", err)
	}
	now := time.Now()
	asm.approval.Status = ApprovalStatusApproved
	asm.approval.ApproverID = &approverID
	asm.approval.ApprovedAt = &now
	return nil
}

// Reject 审批驳回
func (asm *ApprovalStateMachine) Reject(approverID int64, note string) error {
	canFire, err := asm.sm.CanFire("reject")
	if err != nil {
		return fmt.Errorf("状态检查失败: %w", err)
	}
	if !canFire {
		return fmt.Errorf("当前状态不允许驳回")
	}
	if err := asm.sm.Fire("reject"); err != nil {
		return fmt.Errorf("驳回失败: %w", err)
	}
	now := time.Now()
	asm.approval.Status = ApprovalStatusRejected
	asm.approval.ApproverID = &approverID
	asm.approval.RejectedAt = &now
	asm.approval.RejectedNote = note
	return nil
}

// Cancel 撤回申请
func (asm *ApprovalStateMachine) Cancel(requesterID int64) error {
	if !asm.CanCancel(requesterID) {
		return fmt.Errorf("无权撤回此申请")
	}
	if err := asm.sm.Fire("cancel"); err != nil {
		return fmt.Errorf("撤回失败: %w", err)
	}
	now := time.Now()
	asm.approval.Status = ApprovalStatusCancelled
	asm.approval.CancelledAt = &now
	return nil
}

// ApprovalService 审批流服务
type ApprovalService struct {
	repo *AttendanceRepository
}

func NewApprovalService(repo *AttendanceRepository) *ApprovalService {
	return &ApprovalService{repo: repo}
}

// CalculateDuration 计算审批时长（小时），存储精确到 0.01h（D-08）
func (s *ApprovalService) CalculateDuration(startTime, endTime time.Time) float64 {
	duration := endTime.Sub(startTime).Hours()
	if duration < 0 {
		return 0
	}
	return float64(int(duration*100)) / 100
}

// RoundDurationForDisplay 显示时四舍五入到 0.5h（D-07）
func RoundDurationForDisplay(duration float64) float64 {
	return float64(int(duration/0.5+0.5)) * 0.5
}

// CreateApproval 创建审批申请
func (s *ApprovalService) CreateApproval(ctx context.Context, orgID, employeeID int64, req *CreateApprovalRequest) (*ApprovalResponse, error) {
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("开始时间格式错误: %w", err)
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("结束时间格式错误: %w", err)
	}
	if endTime.Before(startTime) {
		return nil, fmt.Errorf("结束时间不能早于开始时间")
	}

	duration := s.CalculateDuration(startTime, endTime)
	attachmentsJSON, _ := json.Marshal(req.Attachments)
	ccJSON, _ := json.Marshal(req.CCUserIDs)

	approval := &Approval{
		BaseModel:    model.BaseModel{OrgID: orgID, CreatedBy: employeeID, UpdatedBy: employeeID},
		EmployeeID:   employeeID,
		ApprovalType: req.ApprovalType,
		StartTime:    startTime,
		EndTime:      endTime,
		Duration:     duration,
		Reason:       req.Reason,
		LeaveType:    req.LeaveType,
		Status:       ApprovalStatusDraft,
		Attachments:  string(attachmentsJSON),
		CCUserIDs:    string(ccJSON),
	}

	sm := NewApprovalStateMachine(approval)
	if err := sm.Submit(); err != nil {
		return nil, fmt.Errorf("提交申请失败: %w", err)
	}

	if err := s.repo.CreateApproval(approval); err != nil {
		return nil, fmt.Errorf("创建申请记录失败: %w", err)
	}

	return toApprovalResponse(approval), nil
}

// Approve 审批通过
func (s *ApprovalService) Approve(ctx context.Context, orgID, approverID, approvalID int64) (*ApprovalResponse, error) {
	approval, err := s.repo.GetApproval(orgID, approvalID)
	if err != nil {
		return nil, fmt.Errorf("获取申请失败: %w", err)
	}
	if approval == nil {
		return nil, fmt.Errorf("申请不存在")
	}

	sm := NewApprovalStateMachine(approval)
	if err := sm.Approve(approverID); err != nil {
		return nil, err
	}

	approval.UpdatedBy = approverID
	if err := s.repo.UpdateApproval(approval); err != nil {
		return nil, fmt.Errorf("更新申请状态失败: %w", err)
	}

	return toApprovalResponse(approval), nil
}

// Reject 审批驳回
func (s *ApprovalService) Reject(ctx context.Context, orgID, approverID, approvalID int64, note string) (*ApprovalResponse, error) {
	approval, err := s.repo.GetApproval(orgID, approvalID)
	if err != nil {
		return nil, fmt.Errorf("获取申请失败: %w", err)
	}
	if approval == nil {
		return nil, fmt.Errorf("申请不存在")
	}

	sm := NewApprovalStateMachine(approval)
	if err := sm.Reject(approverID, note); err != nil {
		return nil, err
	}

	approval.UpdatedBy = approverID
	if err := s.repo.UpdateApproval(approval); err != nil {
		return nil, fmt.Errorf("更新申请状态失败: %w", err)
	}

	return toApprovalResponse(approval), nil
}

// Cancel 撤回申请
func (s *ApprovalService) Cancel(ctx context.Context, orgID, requesterID, approvalID int64) (*ApprovalResponse, error) {
	approval, err := s.repo.GetApproval(orgID, approvalID)
	if err != nil {
		return nil, fmt.Errorf("获取申请失败: %w", err)
	}
	if approval == nil {
		return nil, fmt.Errorf("申请不存在")
	}

	sm := NewApprovalStateMachine(approval)
	if err := sm.Cancel(requesterID); err != nil {
		return nil, err
	}

	approval.UpdatedBy = requesterID
	if err := s.repo.UpdateApproval(approval); err != nil {
		return nil, fmt.Errorf("更新申请状态失败: %w", err)
	}

	return toApprovalResponse(approval), nil
}

// ListApprovals 查询审批列表
func (s *ApprovalService) ListApprovals(ctx context.Context, orgID int64, status, approvalType string, employeeID *int64, page, pageSize int) (*ApprovalListResponse, error) {
	approvals, total, err := s.repo.ListApprovals(orgID, status, approvalType, employeeID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询申请列表失败: %w", err)
	}
	list := make([]ApprovalResponse, len(approvals))
	for i, a := range approvals {
		list[i] = *toApprovalResponse(&a)
	}
	return &ApprovalListResponse{List: list, Total: total, Page: page}, nil
}

// GetPendingCount 获取待审批条数
func (s *ApprovalService) GetPendingCount(ctx context.Context, orgID int64) (int64, error) {
	return s.repo.CountPendingApprovals(orgID)
}

func toApprovalResponse(a *Approval) *ApprovalResponse {
	var attachments []string
	_ = json.Unmarshal([]byte(a.Attachments), &attachments)
	typeName, ok := ApprovalTypeNameMap[a.ApprovalType]
	if !ok {
		typeName = a.ApprovalType
	}
	return &ApprovalResponse{
		ID:           a.ID,
		EmployeeID:   a.EmployeeID,
		ApprovalType: a.ApprovalType,
		TypeName:     typeName,
		StartTime:    a.StartTime.Format(time.RFC3339),
		EndTime:      a.EndTime.Format(time.RFC3339),
		Duration:     a.Duration,
		LeaveType:    a.LeaveType,
		Reason:       a.Reason,
		Status:       a.Status,
		ApproverID:   a.ApproverID,
		ApprovedAt:   formatTimePtr(a.ApprovedAt),
		RejectedAt:   formatTimePtr(a.RejectedAt),
		RejectedNote: a.RejectedNote,
		CancelledAt:  formatTimePtr(a.CancelledAt),
		Attachments:  attachments,
		CreatedAt:    a.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
