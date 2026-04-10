---
phase: 07
slug: homepage
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-10
---

# Phase 07 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

**Frontend (Vue 3 H5)**

| Property | Value |
|----------|-------|
| **Framework** | Vitest (Vue 3 + Vite 标准测试框架) |
| **Config file** | `vite.config.ts` (vitest plugin) + `frontend/src/**/*.test.ts` |
| **Quick run command** | `npm run test:unit -- run` (no watch mode) |
| **Full suite command** | `npm run test:unit -- coverage` |
| **Estimated runtime** | ~30 seconds (full suite) |

**Backend (Go Dashboard)**

| Property | Value |
|----------|-------|
| **Framework** | `go test` (标准库) |
| **Config file** | `internal/dashboard/*_test.go` |
| **Quick run command** | `go test ./internal/dashboard/... -v` |
| **Full suite command** | `go test ./internal/dashboard/... -v -cover` |
| **Estimated runtime** | ~10 seconds |

---

## Sampling Rate

- **After every task commit:** Run frontend quick + backend quick
- **After every plan wave:** Run full suite (coverage)
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

*(Tasks populated by planner after PLAN.md creation)*

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| — | — | — | HOME-01~06 | — | — | — | — | — | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

**Frontend (frontend/ is new — Wave 0 scaffold):**
- [ ] `frontend/src/stores/auth.test.ts` — auth store JWT token 注入逻辑测试
- [ ] `frontend/src/stores/dashboard.test.ts` — dashboard store 聚合逻辑测试
- [ ] `frontend/src/api/__mocks__/axios.ts` — Axios mock
- [ ] `frontend/vite.config.ts` — Vitest plugin 配置
- [ ] `frontend/src/views/home/__tests__/HomeView.test.ts` — 首页组件测试 stubs
- [ ] `frontend/src/views/layout/__tests__/AppLayout.test.ts` — Tab 导航测试 stubs
- [ ] `npm install -D vitest @vue/test-utils` — 测试依赖安装

**Backend (Go — Wave 0 scaffold):**
- [ ] `internal/dashboard/service_test.go` — DashboardService 聚合逻辑单元测试
- [ ] `internal/dashboard/handler_test.go` — HTTP handler 测试 ( gin test helper)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| 移动端 Tab 栏在不同机型显示 | HOME-04 | 需要真机/模拟器验证 | iOS Safari + Android Chrome 各测 1 次 |
| 待办卡片点击跳转正确页面 | HOME-01/07 | 路由集成测试 | 在 H5 Dev Server 中手动点击 6 张卡片 |
| 数据概览数字与后端一致 | HOME-03 | DB 真实数据 | 对照 DB 查询结果验证 API 返回 |
| 无网络时错误提示 | UI/UX | 网络异常模拟 | Chrome DevTools Network → Offline 测试 |

*If none: "All phase behaviors have automated verification."*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending

