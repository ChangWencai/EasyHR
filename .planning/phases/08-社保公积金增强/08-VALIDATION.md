---
phase: 8
slug: social-insurance-enhancement
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-18
---

# Phase 08 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `go test` + `stretchr/testify` |
| **Config file** | none — 标准 Go 测试 |
| **Quick run command** | `go test ./internal/socialinsurance/... -run TestMonthlyPayment -v` |
| **Full suite command** | `go test ./internal/socialinsurance/... -race -cover` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/socialinsurance/... -run <TestName> -v`
- **After every plan wave:** Run `go test ./internal/socialinsurance/... -race -cover`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 08-01 | 01 | 1 | SI-01~SI-04 | T-08-01 | org_id isolation on dashboard | unit | `go test ... -run TestSIDashboard -v` | ❌ W0 | ⬜ pending |
| 08-02 | 01 | 1 | SI-05~SI-08 | T-08-02 | org_id isolation on enroll | integration | `go test ... -run TestEnroll -v` | ❌ W0 | ⬜ pending |
| 08-03 | 02 | 2 | SI-09~SI-13 | T-08-03 | org_id isolation on stop | unit | `go test ... -run TestStop -v` | ❌ W0 | ⬜ pending |
| 08-04 | 02 | 2 | SI-14~SI-16 | T-08-04 | webhook auth token | integration | `go test ... -run TestPaymentChannel -v` | ❌ W0 | ⬜ pending |
| 08-05 | 03 | 2 | SI-17~SI-18 | T-08-05 | status transition idempotency | unit | `go test ... -run TestStatusTransition -v` | ❌ W0 | ⬜ pending |
| 08-06 | 03 | 2 | SI-19~SI-20 | T-08-06 | overdue banner data filtering | unit | `go test ... -run TestFiveInsDetail -v` | ❌ W0 | ⬜ pending |
| 08-07 | 04 | 3 | SI-21 | T-08-07 | Excel injection prevention | unit | `go test ... -run TestExport -v` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/socialinsurance/monthly_payment_test.go` — covers SI-01~SI-04, SI-09~SI-13, SI-17~SI-18, SI-19~SI-20
- [ ] `internal/socialinsurance/asynq_worker_test.go` — covers SI-14~SI-16, SI-17~SI-18
- [ ] Framework install: `go test ./internal/socialinsurance/...` — verify existing tests pass before adding new

*If none: "Existing infrastructure covers all phase requirements."*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| UI 红色横幅渲染 | SI-19 | 前端渲染层，grep 检查 Vue 组件存在即可 | `grep -r "overdueItems" frontend/src/views/socialinsurance/` |
| 五险分项弹窗数字格式 | SI-20 | 千分位分隔符依赖浏览器 toLocaleString | `grep -r "toLocaleString" frontend/src/components/socialinsurance/` |
| 导出 Excel 列宽 | SI-21 | Excel 样式无法自动化验证 | 人工验收：打开导出文件检查列宽 |
| 横幅超量折叠策略 | SI-19 | UI 交互逻辑 | `grep "overdueItems.length" frontend/src/` |

*If none: "All phase behaviors have automated verification."*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending

---
