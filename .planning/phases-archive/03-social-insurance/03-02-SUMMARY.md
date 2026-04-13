---
phase: 03-social-insurance
plan: 02
status: complete
started: "2026-04-07T07:30:00Z"
completed: "2026-04-07T08:00:00Z"
---

# Plan 03-02: 参保/停缴核心操作 + 变更历史

## Objective
实现社保参保/停缴核心操作、缴费明细查询和变更历史记录。老板可批量选择员工一键参保，手动停缴，查看缴费明细，所有操作自动记录变更历史。

## What Was Built

### Models
- **SocialInsuranceRecord**: 参保记录模型（pending/active/stopped 三态），一条记录存储所有险种明细（JSONB Details 字段）
- **ChangeHistory**: 变更历史模型，自动记录 enroll/stop/base_adjust 三种变更类型

### Service Methods
- `EnrollPreview`: 参保预览，按城市自动匹配基数显示各险种金额
- `BatchEnroll`: 批量参保，返回部分成功报告（已有 active 记录的员工跳过）
- `BatchStopEnrollment`: 批量停缴，更新状态和结束月份
- `ListRecords`: 参保记录列表，支持按状态/姓名筛选
- `GetMyRecords`: MEMBER 角色通过 user_id 查询自己记录
- `GetChangeHistory`: 变更历史时间线查询
- `GetSocialInsuranceDeduction`: 社保扣款查询（D-12 接口，供 Phase 5 调用）

### Infrastructure
- **EmployeeAdapter**: 解耦社保和员工模块的适配器
- **EmployeeQuerier**: 员工查询接口（在 socialinsurance 包定义）
- employee.Repository 新增 `FindByIDs` 和 `FindByUserID` 方法

### HTTP Endpoints (7 new)
- `POST /social-insurance/enroll/preview` (owner, admin)
- `POST /social-insurance/enroll` (owner, admin)
- `POST /social-insurance/stop` (owner, admin)
- `GET /social-insurance/records` (all roles)
- `GET /social-insurance/my-records` (all roles)
- `GET /social-insurance/records/:id/history` (all roles)
- `GET /social-insurance/deduction` (all roles)

## Key Decisions
- 社保模块通过 EmployeeQuerier 接口解耦员工模块，避免循环依赖
- 批量参保逐个处理（per Pitfall 5），不用整体事务
- Details JSONB 存储各险种金额快照，参保时冻结
- 变更历史自动记录所有操作（BeforeValue/AfterValue）

## Commits
- `0499a71`: feat(03-02): 实现参保/停缴核心业务逻辑
- `aeee3e3`: fix(03-02): 修复main.go依赖注入缺少EmployeeAdapter参数

## Test Results
- `go test ./internal/socialinsurance/... -count=1`: PASS
- `go build ./cmd/server/`: SUCCESS
- `go test ./... -count=1`: ALL PASS (no regressions)
