---
phase: 9
slug: 待办中心
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-19
---

# Phase 9 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Backend Framework** | `go test` + `testify` (已配置，Phase 8 已验证) |
| **Frontend Framework** | `vitest` (已配置，package.json scripts.test:unit) |
| **Backend Config** | `go.mod` 无需额外安装；`internal/todo/` 目录新建 |
| **Frontend Config** | `frontend/vite.config.ts` 无需修改；`frontend/src/api/` 新建 |
| **Quick run** | `go test ./internal/todo/... -v` / `npm run test:unit` |
| **Full suite** | `go test ./... -race -cover` / `npm run test:unit` |
| **Estimated runtime** | ~60 seconds (backend 20 files + frontend 5 files) |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/todo/... -v` for backend; `npm run test:unit` for frontend
- **After every plan wave:** Run `go test ./... -race -cover` + `npm run test:unit`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 9-01-01 | 01 | 1 | TODO-19 | T-9-01 | 完成率 API 只返回 org 内数据 | unit | `go test ./internal/dashboard/... -run TestGetTodoStats` | ✅ W0 | ⬜ pending |
| 9-01-02 | 01 | 1 | TODO-20 | T-9-01 | 同上 | unit | `go test ./internal/dashboard/... -run TestGetTimeLimitedStats` | ✅ W0 | ⬜ pending |
| 9-01-03 | 01 | 1 | TODO-09 | T-9-03 | 轮播图 OSS URL 校验，不执行 XSS | unit | `go test ./internal/todo/... -run TestListCarousels` | NEW | ⬜ pending |
| 9-01-04 | 01 | 1 | TODO-10 | — | 快捷入口路径有效 | smoke | `grep -c "新入职" frontend/src/views/home/HomeView.vue` | ✅ W0 | ⬜ pending |
| 9-02-01 | 02 | 1 | TODO-01 | T-9-01 | 列表 API org 隔离 | unit | `go test ./internal/todo/... -run TestListTodos` | NEW | ⬜ pending |
| 9-02-02 | 02 | 1 | TODO-02 | T-9-02 | 搜索 SQL 注入防护 | unit | `go test ./internal/todo/... -run TestSearchTodos` | NEW | ⬜ pending |
| 9-02-03 | 02 | 1 | TODO-03 | T-9-01 | 时间段筛选上限 60 天 | unit | `go test ./internal/todo/... -run TestFilterByDateRange` | NEW | ⬜ pending |
| 9-02-04 | 02 | 1 | TODO-06 | — | 置顶排序字段正确更新 | unit | `go test ./internal/todo/... -run TestPinTodo` | NEW | ⬜ pending |
| 9-02-05 | 02 | 1 | TODO-07 | T-9-02 | 列表字段完整输出 | integration | `go test ./internal/todo/... -run TestTodoListFields` | NEW | ⬜ pending |
| 9-02-06 | 02 | 1 | TODO-08 | T-9-04 | Excel 无公式注入 | integration | `go test ./internal/todo/... -run TestExportTodos` | NEW | ⬜ pending |
| 9-03-01 | 03 | 2 | TODO-11 | T-9-01 | 合同任务幂等创建 | unit | `go test ./internal/employee/... -run TestContractTriggersTodo` | NEW | ⬜ pending |
| 9-03-02 | 03 | 2 | TODO-12 | T-9-01 | 合同续签任务幂等创建 | unit | `go test ./internal/employee/... -run TestContractRenewTriggersTodo` | NEW | ⬜ pending |
| 9-03-03 | 03 | 2 | TODO-13 | T-9-01 | 个税申报任务幂等创建 | unit | `go test ./internal/tax/... -run TestTaxDeclarationTodo` | NEW | ⬜ pending |
| 9-03-04 | 03 | 2 | TODO-14 | T-9-01 | 社保缴费任务幂等创建 | unit | `go test ./internal/socialinsurance/... -run TestSIPaymentTodo` | NEW | ⬜ pending |
| 9-03-05 | 03 | 2 | TODO-15 | T-9-01 | 社保增减员任务幂等创建 | unit | `go test ./internal/socialinsurance/... -run TestSIChangeTodo` | NEW | ⬜ pending |
| 9-03-06 | 03 | 2 | TODO-16 | T-9-01 | 年度社保基数调整任务 | unit | `go test ./internal/socialinsurance/... -run TestSIAnnualBaseTodo` | NEW | ⬜ pending |
| 9-03-07 | 03 | 2 | TODO-17 | T-9-01 | 年度公积金基数调整任务 | unit | `go test ./internal/socialinsurance/... -run TestFundAnnualBaseTodo` | NEW | ⬜ pending |
| 9-04-01 | 04 | 2 | TODO-04 | T-9-05 | Token 安全生成，不可枚举 | unit | `go test ./internal/todo/... -run TestInviteTodo` | NEW | ⬜ pending |
| 9-04-02 | 04 | 2 | TODO-05 | T-9-01 | 终止保留数据，状态正确 | unit | `go test ./internal/todo/... -run TestTerminateTodo` | NEW | ⬜ pending |
| 9-04-03 | 04 | 2 | TODO-18 | — | urgency_status 状态计算正确 | unit | `go test ./internal/todo/... -run TestUrgencyStatus` | NEW | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/todo/model.go` — TodoItem 扩展字段（deadline/is_time_limited/urgency_status/status/source_type/source_id/is_pinned/sort_order）
- [ ] `internal/todo/carousel_model.go` — CarouselItem 模型
- [ ] `internal/todo/todo_invite_model.go` — TodoInvite Token 表
- [ ] `internal/todo/repository.go` — TodoRepository（ListTodos/SearchTodos/TerminateTodo/PinTodo）
- [ ] `internal/todo/carousel_repository.go` — CarouselRepository（ListCarousels）
- [ ] `internal/todo/service.go` — TodoService（urgency_status 状态机逻辑）
- [ ] `internal/todo/handler.go` — Gin Handler（ListTodos/ExportTodos/InviteTodo/TerminateTodo）
- [ ] `internal/todo/excel.go` — Excel 导出（excelize）
- [ ] `internal/todo/scheduler.go` — asynq/gocron 定时任务（urgency_status 扫描 + 轮播图激活）
- [ ] `internal/todo/router.go` — 路由注册
- [ ] `internal/dashboard/model.go` — 扩展 TodoItem 字段
- [ ] `internal/dashboard/service.go` — 扩展 GetTodoStats（环形图）
- [ ] `internal/dashboard/repository.go` — 扩展 todo 统计查询
- [ ] `frontend/src/api/todo.ts` — 待办 API 客户端
- [ ] `frontend/src/api/carousel.ts` — 轮播图 API 客户端
- [ ] `frontend/src/views/home/HomeView.vue` — 环形图 + 轮播图 + 快捷入口
- [ ] `frontend/src/views/todo/TodoListView.vue` — 待办完整列表页
- [ ] `frontend/src/views/todo/InviteFillPage.vue` — 协办填写页
- [ ] `frontend/src/views/home/components/TodoRingChart.vue` — 环形图组件
- [ ] `frontend/src/views/home/components/HomeCarousel.vue` — 轮播图组件

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| 轮播图 OSS 上传图片可见性 | TODO-09 | 需要登录阿里云 OSS 控制台验证图片可见性 | 1. 上传图片到 OSS 测试 bucket；2. 访问 image_url 确认可访问 |
| 协办填写页 Token 链接邮件/消息发送 | TODO-04 | 消息发送依赖外部渠道，无法自动化验证 | 1. 生成测试 Token；2. 手动访问链接；3. 确认页面正常渲染 |

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

---

## Security Threat Model

| ID | Threat | Mitigation | ASVS |
|----|--------|-----------|------|
| T-9-01 | 跨租户访问待办数据 | org_id GORM Scope 自动注入，Token 验证时校验 org_id | V4.1 |
| T-9-02 | 搜索 SQL 注入 | GORM 参数化查询 | V5.2 |
| T-9-03 | 轮播图 XSS | OSS URL 白名单校验，不在页面执行用户输入 | V5.3 |
| T-9-04 | Excel 公式注入 | excelize 输出 xlsx 格式，不使用 CSV | V5.3 |
| T-9-05 | Token 枚举/暴力破解 | crypto/rand 32字节熵，7天过期 | V6.1 |
