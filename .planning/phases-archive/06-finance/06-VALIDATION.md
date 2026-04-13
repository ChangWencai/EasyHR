---
phase: 06
slug: finance
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-10
---

# Phase 06 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test + testify |
| **Config file** | 无 — 共享 testutil (参考 internal/salary/calculator_test.go) |
| **Quick run command** | `go test ./internal/finance/... -v -short` |
| **Full suite command** | `go test ./internal/finance/... -v -race` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/finance/... -run TestVoucherBalance -v`
- **After every plan wave:** Run `go test ./internal/finance/... -v -short`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 06-00-01 | 06-00 | 1 | Framework install | setup | `go get github.com/shopspring/decimal` | ✅ | ⬜ pending |
| 06-01-01 | 06-01 | 2 | 数据模型定义 | unit | `go test ./internal/finance/... -run TestAccountModel -x` | W0 | ⬜ pending |
| 06-01-02 | 06-01 | 2 | 科目体系 CRUD | unit | `go test ./internal/finance/... -run TestAccountManagement -x` | W0 | ⬜ pending |
| 06-01-03 | 06-01 | 2 | 凭证 CRUD | unit | `go test ./internal/finance/... -run TestVoucher -x` | W0 | ⬜ pending |
| 06-01-04 | 06-01 | 2 | PayrollAdapter | unit | `go test ./internal/finance/... -run TestPayrollAdapter -x` | W0 | ⬜ pending |
| 06-02-01 | 06-02 | 3 | 发票管理 | unit | `go test ./internal/finance/... -run TestInvoice -x` | W0 | ⬜ pending |
| 06-02-02 | 06-02 | 3 | 费用报销审批 | unit | `go test ./internal/finance/... -run TestExpense -x` | W0 | ⬜ pending |
| 06-03-01 | 06-03 | 4 | 账簿生成 | unit | `go test ./internal/finance/... -run TestAccountBalance -x` | W0 | ⬜ pending |
| 06-03-02 | 06-03 | 4 | 财务报表 | unit | `go test ./internal/finance/... -run TestFinancialReport -x` | W0 | ⬜ pending |
| 06-03-03 | 06-03 | 4 | 结账/反结账 | unit | `go test ./internal/finance/... -run TestClosing -x` | W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/finance/model_test.go` — 核心数据模型单元测试
- [ ] `internal/finance/voucher_service_test.go` — 凭证业务逻辑测试（借贷平衡、红冲）
- [ ] `internal/finance/service_test.go` — 账簿/报表计算测试
- [ ] Framework install: `go get github.com/shopspring/decimal` — decimal 库引入

*06-00-PLAN.md creates the Wave 0 test scaffold and installs decimal dependency.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| 借贷平衡实时校验（UI交互） | FINC-01 | 前端实时校验无法在单元测试验证 | 在 H5 页面手动测试：输入借贷不平衡金额，验证是否阻止提交 |
| 凭证号自动生成 | FINC-04 | 需要完整数据库状态 | 在测试环境手动验证：连续创建3张凭证，检查编号连续性 |
| 报表多期对比（UI展示） | FINC-15 | UI交互测试 | 在 H5 页面选择两个期间，验证对比数据正确 |

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

## Key Test Commands (Phase 06)

### 凭证借贷平衡测试
```bash
go test ./internal/finance/... -run TestVoucherBalance -v -x
```

### 凭证红冲测试
```bash
go test ./internal/finance/... -run TestVoucherRedFlash -v -x
```

### 科目余额表测试
```bash
go test ./internal/finance/... -run TestAccountBalance -v -x
```

### 财务报表生成测试
```bash
go test ./internal/finance/... -run TestFinancialReport -v -x
```

### 结账测试
```bash
go test ./internal/finance/... -run TestClosing -v -x
```

### 完整测试套件
```bash
go test ./internal/finance/... -v -race
```

---

*Phase: 06-finance*
*Validation strategy created: 2026-04-10*
