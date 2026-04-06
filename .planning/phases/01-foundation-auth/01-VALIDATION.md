---
phase: 1
slug: foundation-auth
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-06
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing + testify v1.11.1 |
| **Config file** | none — Go native testing |
| **Quick run command** | `go test ./internal/... -v -short` |
| **Full suite command** | `go test ./... -v -cover` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/... -v -short`
- **After every plan wave:** Run `go test ./... -v -cover`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 01-01-01 | 01 | 1 | PLAT-03 | unit | `go test ./internal/common/response/... -v` | ❌ W0 | ⬜ pending |
| 01-01-02 | 01 | 1 | PLAT-05 | unit | `go test ./internal/common/crypto/... -v` | ❌ W0 | ⬜ pending |
| 01-01-03 | 01 | 1 | PLAT-06 | unit | `go test ./internal/common/middleware/... -v` | ❌ W0 | ⬜ pending |
| 01-01-04 | 01 | 1 | PLAT-04 | unit | `go test ./internal/common/middleware/... -v` | ❌ W0 | ⬜ pending |
| 01-02-01 | 02 | 1 | AUTH-04 | unit | `go test ./pkg/jwt/... -v` | ❌ W0 | ⬜ pending |
| 01-02-02 | 02 | 1 | PLAT-07 | unit | `go test ./pkg/oss/... -v` | ❌ W0 | ⬜ pending |
| 01-03-01 | 03 | 2 | AUTH-01 | integration | `go test ./internal/user/... -run TestLogin -v` | ❌ W0 | ⬜ pending |
| 01-03-02 | 03 | 2 | AUTH-02 | integration | `go test ./internal/user/... -run TestOnboarding -v` | ❌ W0 | ⬜ pending |
| 01-03-03 | 03 | 2 | PLAT-01 | unit | `go test ./internal/common/middleware/... -run TestRBAC -v` | ❌ W0 | ⬜ pending |
| 01-03-04 | 03 | 2 | PLAT-02 | integration | `go test ./internal/common/... -run TestAudit -v` | ❌ W0 | ⬜ pending |
| 01-04-01 | 04 | 2 | AUTH-04 | integration | `go test ./internal/user/... -run TestTokenRefresh -v` | ❌ W0 | ⬜ pending |
| 01-04-02 | 04 | 2 | PLAT-06 | integration | `go test ./internal/... -run TestTenantIsolation -v` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/user/handler_test.go` — stubs for AUTH-01, AUTH-02
- [ ] `internal/user/service_test.go` — stubs for AUTH-01 business logic
- [ ] `internal/user/repository_test.go` — stubs for PLAT-06 tenant isolation
- [ ] `internal/common/response/response_test.go` — stubs for PLAT-03
- [ ] `internal/common/crypto/crypto_test.go` — stubs for PLAT-05
- [ ] `internal/common/middleware/auth_test.go` — stubs for PLAT-04, PLAT-01
- [ ] `pkg/jwt/jwt_test.go` — stubs for AUTH-04
- [ ] `pkg/oss/oss_test.go` — stubs for PLAT-07

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| SMS verification code delivery | AUTH-01 | Requires 阿里云 SMS API credentials + real phone number | Send test SMS to test phone number, verify 6-digit code received |
| OSS signed URL upload | PLAT-07 | Requires 阿里云 OSS credentials + real bucket | Upload test file via signed URL, verify file in OSS console |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
