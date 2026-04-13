---
status: verifying
updated: 2026-04-11T00:00:00+08:00
---

## Current Focus
hypothesis: "已确认两个 bug 并全部修复：1) dashboard handler uint→int64；2) CompleteOnboarding 返回新 token"
next_action: "等待用户在真实环境验证原始问题是否解决"

## Symptoms
<!-- IMMUTABLE -->
expected: "Onboarding 成功后，访问 /api/v1/dashboard 应返回仪表盘数据"
actual: "/api/v1/dashboard 返回 401 Unauthorized {\"code\":40100,\"message\":\"invalid org_id type\"}"
errors: "401, message=\"invalid org_id type\""
reproduction: "完成 onboarding 后访问 dashboard 页面"
started: "2026-04-11"

## Eliminated
<!-- APPEND only -->

## Evidence
<!-- APPEND only -->
- timestamp: 2026-04-11
  checked: "pkg/jwt/jwt.go:13"
  found: "JWT Claims 定义 OrgID 为 int64，middleware 正确存 int64 到 context"
  implication: "middleware 行为正确"
- timestamp: 2026-04-11
  checked: "internal/dashboard/handler.go:28"
  found: "handler 用 orgID, ok := orgIDVal.(uint) 断言 uint，但 context 存的是 int64"
  implication: "类型断言永远失败 → 'invalid org_id type'"
- timestamp: 2026-04-11
  checked: "internal/user/service.go:210-226"
  found: "CompleteOnboarding 创建企业并更新 user.org_id，但返回 nil 不带新 token"
  implication: "onboarding 后旧 token 中 org_id 仍为 0，用户需重新登录才能刷新 token"
- timestamp: 2026-04-11
  checked: "internal/dashboard/repository.go 全链路"
  found: "repository/service 都用 uint，改 handler 后需同步修改全链路否则编译报错"
  implication: "全链路统一为 int64（与 JWT Claims 一致）"
- timestamp: 2026-04-11
  checked: "go test ./internal/dashboard/... -v"
  found: "9 个测试全部 PASS"
  implication: "修复后代码功能正常，无回归"

## Resolution
root_cause: "两个 bug：1) dashboard handler 断言 uint 但 JWT/org_id 是 int64 → 类型断言永远失败；2) CompleteOnboarding 不返回新 token，导致 token 中 org_id 永远是注册时的 0"
fix: |
  1. dashboard/handler.go:28 将 orgIDVal.(uint) 改为 orgIDVal.(int64)
  2. dashboard/service.go/repository.go/repository_mock.go: 全链路 uint→int64
  3. user/service.go: CompleteOnboarding 返回 *LoginResponse（含新 token）
  4. user/handler.go: CompleteOnboarding 调用改为接收 resp 并返回
verification: "编译通过，9 个 dashboard 测试全部 PASS"
files_changed:
  - internal/dashboard/handler.go
  - internal/dashboard/service.go
  - internal/dashboard/repository.go
  - internal/dashboard/repository_mock.go
  - internal/dashboard/handler_test.go
  - internal/user/service.go
  - internal/user/handler.go
