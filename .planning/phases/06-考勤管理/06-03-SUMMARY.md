---
phase: "06"
plan: "03"
status: complete
type: execute
wave: 2
---

# 06-03 考勤审批流 — SUMMARY

## What was built

考勤审批流完整实现，支持 11 种审批类型的全生命周期管理。

### Backend
- **Approval 模型** (`model.go`): 审批记录表，含 11 种审批类型常量、状态常量、类型名称映射
- **状态机** (`approval_service.go`): 基于 qmuntal/stateless 实现 draft→pending→approved/rejected/cancelled/timeout 转换
- **CRUD API** (`handler.go`): GET/POST/PUT 端点（列表/创建/审批/驳回/撤回/待办计数）
- **Repository** (`repository.go`): Approval 数据访问层（创建/查询/列表/按员工月度查询/待办统计）
- **DTO** (`dto.go`): ApprovalResponse/CreateApprovalRequest/RejectApprovalRequest

### Frontend
- **AttendanceApproval.vue**: 3-Tab 审批页面（全部/待我审批/我发起的）+ 行内同意/驳回/撤回
- **ApprovalApplyDialog.vue**: 新建申请弹窗（11 种类型 + 自动时长计算）
- **ApprovalTypeTag.vue**: 审批类型标签组件（不同颜色区分类型）
- **API 层** (`attendance.ts`): approvalApi 完整 CRUD

### Key Decisions
- D-02: 状态机管理审批生命周期
- D-03: 仅申请人可取消自己的申请
- D-07: 显示时长按 0.5h 取整
- D-08: 存储精确到 0.01h

## Self-Check: PASSED

- [x] Approval 模型含 11 种审批类型
- [x] 状态机管理 draft→pending→终态转换
- [x] 管理员可同意或驳回申请
- [x] 审批列表显示待办条数
- [x] 请假时长按 0.5h 取整显示，0.01h 存储
- [x] Go 编译通过
- [x] 依赖注入已更新 main.go

## key-files

### created
- internal/attendance/approval_service.go
- frontend/src/views/attendance/AttendanceApproval.vue
- frontend/src/components/attendance/ApprovalApplyDialog.vue
- frontend/src/components/attendance/ApprovalTypeTag.vue

### modified
- internal/attendance/model.go
- internal/attendance/dto.go
- internal/attendance/repository.go
- internal/attendance/handler.go
- cmd/server/main.go
- frontend/src/api/attendance.ts
- frontend/src/router/index.ts
