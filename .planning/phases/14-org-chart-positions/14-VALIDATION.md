---
phase: 14
slug: org-chart-positions
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-21
---

# Phase 14 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (backend) + vitest (frontend) |
| **Config file** | none — existing infrastructure |
| **Quick run command** | `go test ./internal/position/... ./internal/department/... -v -count=1 2>&1 | tail -20` |
| **Full suite command** | `go test ./internal/position/... ./internal/department/... ./internal/employee/... -v -count=1 && cd frontend && npm run test -- --run 2>&1 | tail -20` |
| **Estimated runtime** | ~45 seconds (backend 30s + frontend 15s) |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/position/... ./internal/department/... -v -count=1`
- **After every plan wave:** Run full suite
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 14-01-01 | 01 | 1 | ORG-01 | T-14-01 | Position names sanitized, org isolation enforced | unit | `go test ./internal/position/... -v -run TestPosition` | W0 stub needed | ⬜ pending |
| 14-01-02 | 01 | 1 | ORG-01 | T-14-02 | Unique constraint prevents duplicates per org | unit | `go test ./internal/position/... -v -run TestUnique` | W0 stub needed | ⬜ pending |
| 14-01-03 | 01 | 1 | ORG-01 | — | Migration maps existing text positions to FK | unit | `go test ./internal/position/... -v -run TestMigration` | W0 stub needed | ⬜ pending |
| 14-02-01 | 01 | 1 | ORG-02 | — | Tree rebuilds with real Position IDs, no virtual nodes | unit | `go test ./internal/department/... -v -run TestBuildTree` | W0 stub needed | ⬜ pending |
| 14-02-02 | 01 | 1 | ORG-02 | — | Cycle detection prevents parent loops | unit | `go test ./internal/department/... -v -run TestCycle` | W0 stub needed | ⬜ pending |
| 14-02-03 | 01 | 1 | ORG-02 | — | Move-to updates parent_id, triggers tree rebuild | unit | `go test ./internal/department/... -v -run TestUpdateParent` | W0 stub needed | ⬜ pending |
| 14-03-01 | 01 | 1 | ORG-03 | T-14-03 | Transfer moves employees, preserves position_id | unit | `go test ./internal/department/... -v -run TestTransferDelete` | W0 stub needed | ⬜ pending |
| 14-04-01 | 02 | 2 | ORG-04 | — | Position dropdown filters correctly by department | unit | `cd frontend && npm run test -- --run position` | W0 stub needed | ⬜ pending |
| 14-04-02 | 02 | 2 | ORG-04 | — | Employee form position select renders with groups | unit | `cd frontend && npm run test -- --run employee-position` | W0 stub needed | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/position/position_test.go` — stubs for TestPosition, TestUnique, TestMigration
- [ ] `internal/department/department_test.go` — stubs for TestBuildTree, TestCycle, TestUpdateParent, TestTransferDelete
- [ ] `frontend/src/__tests__/position.test.ts` — stubs for position dropdown tests
- [ ] `frontend/src/__tests__/employee-position.test.ts` — stubs for employee form position tests

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| ECharts tree visual drag-and-drop UX | ORG-02 | Visual interaction requires browser rendering | Open OrgChart.vue, right-click dept node → "移动到..." → select new parent → tree reflows correctly |
| Inline edit activates on click | ORG-03 | DOM interaction with input focus | Click dept name in tree → input appears → type → blur saves |
| Delete transfer dialog renders employees | ORG-03 | Dialog shows live employee list | Delete dept with employees → dialog lists all employees → select target → confirm |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending