---
status: fixed
trigger: "GET /api/v1/orgs/current 返回 404 Not Found，从未成功过（新接口）"
created: 2026-04-11T00:00:00+08:00
updated: 2026-04-11T00:00:00+08:00
---

## Current Focus
hypothesis: "后端没有注册 /api/v1/orgs/current 路由，前端应改调用已存在的 /api/v1/auth/me"
test: "将 MineView.vue 中的 /orgs/current 改为 /auth/me，并调整响应映射"
expecting: "/auth/me 返回 MeResponse {id, name, phone, role, org: {id, name, credit_code, city}}"
next_action: "等待用户在\"我的\"页面验证修复效果"

## Symptoms
expected: "返回当前登录用户的组织信息（org details）和用户信息"
actual: "返回 404 Not Found，\"我的\"页面无法显示登录状态和企业信息"
errors: "404 Not Found"
reproduction: "切换到\"我的\"页面时，前端自动调用 /api/v1/orgs/current"
started: "从未成功过（新接口）"

## Eliminated
- hypothesis: "后端路由路径写错了（路径不匹配）"
  evidence: "后端从未注册过 /orgs/current 路由，不存在路径写错的问题，是完全缺失"
  timestamp: 2026-04-11
- hypothesis: "需要新增后端 /orgs/current handler"
  evidence: "GetMe handler 已存在并返回完整 user+org 信息，前端只需改调用路径"
  timestamp: 2026-04-11

## Evidence
- timestamp: 2026-04-11
  checked: "frontend/src/api/request.ts"
  found: "baseURL: '/api/v1'，调用 /orgs/current 实际请求 GET /api/v1/orgs/current"
  implication: "前端发送的请求路径是 /api/v1/orgs/current"
- timestamp: 2026-04-11
  checked: "cmd/server/main.go 路由注册"
  found: "注册了 user/emp/inv/ob/contract/si/tax/salary/finance/city/audit/dashboard/wxmp 路由，没有 /orgs 相关路由"
  implication: "后端没有 /api/v1/orgs/current 路由，返回 404"
- timestamp: 2026-04-11
  checked: "internal/user/handler.go"
  found: "已注册 authGroup.GET(\"/auth/me\", h.GetMe)，即 GET /api/v1/auth/me"
  implication: "后端有 /auth/me 接口，但前端没用它"
- timestamp: 2026-04-11
  checked: "internal/user/service.go GetMe() 和 dto.go MeResponse"
  found: "GetMe 返回 MeResponse{ID, Name, Phone, Role, Org: *OrgInfo{id,name,credit_code,city}, OnboardingRequired}"
  implication: "后端已有完整 user+org 数据，前端改调用 /auth/me 即可"
- timestamp: 2026-04-11
  checked: "frontend/src/views/mine/MineView.vue"
  found: "调用 request.get('/orgs/current')，期望 res.user 和 res.org"
  implication: "前端需改为调用 /auth/me，且 MeResponse 结构不同于预期(res.user/res.org 而是扁平结构)"
- timestamp: 2026-04-11
  checked: "frontend/src/views/mine/MineView.vue (修改后)"
  found: "已修改 loadOrgInfo() 调用 /auth/me，正确映射 res.id/res.name/res.phone/res.role 到 user，res.org 到 org"
  implication: "前端修复已完成"

## Resolution
root_cause: "前端调用了不存在的路由 /api/v1/orgs/current。后端已有 GET /api/v1/auth/me 接口（通过 GetMe handler），返回完整的用户信息和企业信息，与前端需求完全一致，但前端错误地调用了一个从未注册过的路由"
fix: "将 frontend/src/views/mine/MineView.vue 中 request.get('/orgs/current') 改为 request.get('/auth/me')，并将响应映射从 res.user/res.org 扁平化为 res.id/res.name/res.phone/res.role 和 res.org"
verification: "修复已提交，等待用户手动验证\"我的\"页面"
files_changed: ["frontend/src/views/mine/MineView.vue"]
