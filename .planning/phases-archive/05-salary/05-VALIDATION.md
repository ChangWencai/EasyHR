---
phase: 05
slug: salary
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-08
---

# Phase 05 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (标准库) + testify v1.11.1 |
| **Config file** | none — 已在前序 Phase 建立 |
| **Quick run command** | `go test -race -count=1 ./internal/salary/...` |
| **Full suite command** | `go test -race -count=1 ./internal/...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test -race -count=1 ./internal/salary/...`
- **After every plan wave:** Run `go test -race -count=1 ./internal/...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | Status |
|---------|------|------|-------------|-----------|-------------------|--------|
| 05-01-01 | 01 | 1 | PAYR-01 | unit | `go test ./internal/salary/... -run TestSalaryTemplate` | ⬜ pending |
| 05-01-02 | 01 | 1 | PAYR-01 | unit | `go test ./internal/salary/... -run TestSalaryItem` | ⬜ pending |
| 05-02-01 | 02 | 1 | PAYR-02 | unit | `go test ./internal/salary/... -run TestCalculate` | ⬜ pending |
| 05-02-02 | 02 | 1 | PAYR-02,03 | unit | `go test ./internal/salary/... -run TestCopyFrom` | ⬜ pending |
| 05-02-03 | 02 | 2 | PAYR-04,08,09 | integration | `go test ./internal/salary/... -run "TestAttendanceImport\|TestPayConfirm"` | ⬜ pending |
| 05-03-01 | 03 | 3 | PAYR-05,06 | integration | `go test ./internal/salary/... -run TestSlipToken` | ⬜ pending |
| 05-03-02 | 03 | 3 | PAYR-07 | integration | `go test ./internal/salary/... -run TestExportExcel` | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing test infrastructure covers all phase requirements (SQLite in-memory testing from test/testutil/).

*No Wave 0 needed — existing infrastructure sufficient.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| 工资单 H5 页面中文显示 | PAYR-05 | 需要 PDF/HTML 渲染验证 | 导出工资单后打开查看 |
| Excel 导出格式对齐 | PAYR-07 | 需要与 Excel 软件对比 | 导出后用 Excel 打开验证格式 |
| 考勤导入边界情况 | PAYR-04 | 需要真实 Excel 文件 | 导入含特殊字符/空行的 Excel |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
