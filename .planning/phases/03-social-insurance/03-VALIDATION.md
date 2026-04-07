---
phase: 3
slug: social-insurance
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-07
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing + testify v1.11.1 |
| **Config file** | none (SQLite内存数据库) |
| **Quick run command** | `go test ./internal/socialinsurance/... -count=1 -v` |
| **Full suite command** | `go test ./... -count=1 -race` |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/socialinsurance/... -count=1 -v`
- **After every plan wave:** Run `go test ./... -count=1 -race`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 03-01-01 | 01 | 1 | SOCL-01 | unit | `go test ./internal/socialinsurance/... -run TestPolicyCRUD -v` | Wave 0 | pending |
| 03-01-02 | 01 | 1 | SOCL-01 | unit | `go test ./internal/socialinsurance/... -run TestCalculateInsurance -v` | Wave 0 | pending |
| 03-02-01 | 02 | 1 | SOCL-02, SOCL-04 | unit | `go test ./internal/socialinsurance/... -run TestEnrollEmployees -v` | Wave 0 | pending |
| 03-02-02 | 02 | 1 | SOCL-07 | unit | `go test ./internal/socialinsurance/... -run TestStopInsurance -v` | Wave 0 | pending |
| 03-03-01 | 03 | 2 | SOCL-03 | unit | `go test ./internal/socialinsurance/... -run TestCheckPaymentDue -v` | Wave 0 | pending |
| 03-03-02 | 03 | 2 | SOCL-05 | unit | `go test ./internal/socialinsurance/... -run TestExportExcel -v` | Wave 0 | pending |
| 03-04-01 | 04 | 2 | SOCL-06, SOCL-07 | unit | `go test ./internal/socialinsurance/... -run TestSalaryChangeReminder -v` | Wave 0 | pending |
| 03-04-02 | 04 | 2 | SOCL-04 | unit | `go test ./internal/socialinsurance/... -run TestChangeHistory -v` | Wave 0 | pending |

*Status: pending / green / red / flaky*

---

## Wave 0 Requirements

- [ ] `internal/socialinsurance/service_test.go` — stubs for SOCL-01 through SOCL-07
- [ ] `internal/socialinsurance/repository_test.go` — covers policy CRUD, record CRUD
- [ ] `internal/socialinsurance/scheduler_test.go` — covers payment due scan logic
- [ ] Framework install: `go get github.com/go-co-op/gocron/v2@v2.19.1 && go get github.com/go-co-op/gocron-redis-lock/v2@v2.2.1`

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| 参保材料PDF格式检查 | SOCL-02 | 需人工检查PDF排版、中文显示是否正常 | 生成PDF后打开检查排版、表头、数据完整性 |
| 缴费凭证Excel格式检查 | SOCL-05 | 需人工检查Excel格式、列头、数据对齐 | 导出Excel后打开检查列头、金额格式、合计行 |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 15s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
