---
phase: 05
slug: 员工管理增强-组织架构基础
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-04-17
---

# Phase 05 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing + testify (后端), Vitest (前端) |
| **Config file** | 无独立 vitest.config（Vite 内联配置） |
| **Quick run command** | `go test -race ./internal/employee/... ./internal/department/... -v` |
| **Full suite command** | `go test -race ./... -cover && cd frontend && npm run test:unit` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test -race ./internal/employee/... ./internal/department/... -v`
- **After every plan wave:** Run `go test -race ./... -cover`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 05-01-01 | 01 | 1 | EMP-01 | T-05-01 | N/A | unit | `go test ./internal/dashboard/... -run TestGetEmployeeDashboard -v` | in-task | ⬜ pending |
| 05-01-02 | 01 | 1 | EMP-02 | T-05-02 | 分母为0时返回0%而非NaN | unit | `go test ./internal/dashboard/... -run TestTurnoverRate -v` | in-task | ⬜ pending |
| 05-02-01 | 02 | 1 | EMP-03 | — | N/A | unit | `go test ./internal/department/... -run TestBuildTree -v` | in-task | ⬜ pending |
| 05-02-02 | 02 | 1 | EMP-04 | — | N/A | unit | `go test ./internal/department/... -run TestSearchTree -v` | in-task | ⬜ pending |
| 05-03-01 | 03 | 2 | EMP-05 | T-05-03 | Token 使用 crypto/rand 32-byte hex | unit | `go test ./internal/employee/... -run TestCreateRegistration -v` | in-task | ⬜ pending |
| 05-03-02 | 03 | 2 | EMP-06 | — | 过期 Token 拒绝访问 | unit | `go test ./internal/employee/... -run TestRegistrationExpiry -v` | in-task | ⬜ pending |
| 05-03-03 | 03 | 2 | EMP-07 | — | N/A | unit | `go test ./internal/employee/... -run TestRegistrationOverwrite -v` | in-task | ⬜ pending |
| 05-03-04 | 03 | 2 | EMP-08 | — | 提交后员工档案数据一致 | unit | `go test ./internal/employee/... -run TestSubmitRegistration -v` | in-task | ⬜ pending |
| 05-04-01 | 04 | 2 | EMP-09 | — | N/A | unit | `go test ./internal/employee/... -run TestListOffboardingsWithRejected -v` | in-task | ⬜ pending |
| 05-04-02 | 04 | 2 | EMP-10 | — | rejected 状态转换正确 | unit | `go test ./internal/employee/... -run TestRejectResign -v` | in-task | ⬜ pending |
| 05-04-03 | 04 | 2 | EMP-11 | — | N/A | manual | 手动验证跳转参数传递 | N/A | ⬜ pending |
| 05-04-04 | 04 | 2 | EMP-12 | — | 减员完成状态联动 | integration | 手动验证 | N/A | ⬜ pending |
| 05-05-01 | 05 | 3 | EMP-13 | — | 薪资/年限/合同到期/手机号返回 | unit | `go test ./internal/employee/... -run TestListRoster -v` | in-task | ⬜ pending |
| 05-05-02 | 05 | 3 | EMP-14 | — | N/A | unit (前端) | `cd frontend && npx vitest run --reporter=verbose` | in-task | ⬜ pending |
| 05-05-03 | 05 | 3 | EMP-15 | — | N/A | unit | `go test ./internal/employee/... -run TestRosterSearch -v` | in-task | ⬜ pending |
| 05-05-04 | 05 | 3 | EMP-16 | — | N/A | unit | `go test ./internal/employee/... -run TestExportExcelEnhanced -v` | in-task | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Wave 0 测试 stub 和依赖安装已整合到各 Plan 的 Task action 中，无需独立 Wave 0 任务：

- Plan 05-01 Task 1: 包含 `handler_test.go` 中 TestGetEmployeeDashboard 测试编写
- Plan 05-02 Task 1 前置: 包含 `npm install echarts vue-echarts` 安装
- Plan 05-03 Task 1: 包含 Registration 模块完整测试（model/service/handler）
- Plan 05-04 Task 1: 包含 RejectResign 测试

以下独立测试文件由各 Plan Task 在实现过程中同步创建：

- [ ] `internal/department/model_test.go` — stubs for EMP-03/EMP-04
- [ ] `internal/department/service_test.go` — Department CRUD + Tree 构建
- [ ] `internal/employee/registration_model_test.go` — Registration Token 生成
- [ ] `internal/employee/registration_service_test.go` — EMP-05~EMP-08
- [ ] `internal/employee/offboarding_service_test.go` — EMP-10 (RejectResign)
- [ ] `internal/dashboard/service_test.go` — EMP-01/EMP-02
- [ ] 前端: `npm install echarts vue-echarts` — ECharts 依赖安装

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| 审批通过后跳转社保减员页面参数传递 | EMP-11 | 跨模块路由跳转需浏览器环境验证 | 在 H5 管理后台点击"去减员"，验证 URL 参数含 employee_id 和 name |
| 减员完成后离职状态自动更新 | EMP-12 | 跨模块联动（社保+离职）需端到端验证 | 在社保减员页面完成减员操作，返回离职列表确认状态已更新 |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references (integrated into plan tasks)
- [x] No watch-mode flags
- [x] Feedback latency < 30s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
