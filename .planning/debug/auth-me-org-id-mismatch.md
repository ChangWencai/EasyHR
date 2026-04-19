---
status: fixed
trigger: "GET /api/v1/auth/me 返回 10013，用户存在但查不到"
created: 2026-04-11T00:00:00+08:00
updated: 2026-04-11T00:00:00+08:00
---

## Current Focus
root_cause: "pkg/jwt/jwt.go RefreshTokens 生成新 access token 时硬编码 orgID=0，导致 onboarding 后用户的旧 refresh token 被刷新时，生成的新 access token 携带 orgID=0，后续所有需要 tenant scope 的请求都会失败"
fix: "1. pkg/jwt: RefreshTokens 增加 orgID/role/userID 参数\n2. internal/user/service.go: RefreshToken 先通过 repo.FindByID(userID) 查询用户最新信息，再传给 jwt\n3. internal/user/repository.go: 新增 FindByID（不使用 tenant scope）和 UpdateUserOrgID"
next_action: "验证 build + test 通过"

## Symptoms
expected: "GetMe 应返回 user_id=4 的用户信息（org_id=5, role=owner）"
actual: "返回 10013 \"获取用户信息失败\"，GetMe 报错 \"用户不存在, user_id=4\""
errors: "code: 10013, message: \"获取用户信息失败\""
reproduction: "登录后访问 /api/v1/auth/me"
started: "Token 存了 org_id=0，但数据库中用户 org_id=5，导致按 org_id 查询找不到用户"

## Eliminated
- hypothesis: "Token 发放时 org_id 就写错了"
  evidence: "Login/CompleteOnboarding/PasswordLogin 都正确传入了 user.OrgID，只有 RefreshTokens 硬编码 orgID=0"
  timestamp: 2026-04-11

## Evidence
- timestamp: 2026-04-11
  checked: "pkg/jwt/jwt.go:102"
  found: "GenerateAccessToken(claims.UserID, 0, \"\", secret, accessTTL) — orgID 硬编码为 0，role 为空字符串"
  implication: "ROOT CAUSE FOUND: RefreshTokens 生成的 access token org_id=0"
- timestamp: 2026-04-11
  checked: "pkg/jwt/jwt.go:45-56"
  found: "GenerateRefreshToken 只含 UserID，不含 OrgID/Role"
  implication: "refresh token 本身无法携带 org 信息，必须在刷新时从数据库查询"
- timestamp: 2026-04-11
  checked: "internal/user/service.go:188-193"
  found: "RefreshToken 方法只传 refreshTokenStr，orgID/role 无法从 refresh token 获取"
  implication: "需要在 service 层查库获取用户信息后再传给 jwt"
- timestamp: 2026-04-11
  checked: "internal/user/repository.go:71"
  found: "FindUserByID(orgID, userID) 使用 TenantScope(orgID)，orgID=0 时查不到任何用户"
  implication: "确认查询失败原因：Token 中 orgID=0 → TenantScope(orgID=0) → WHERE org_id=0 → user_id=4 的用户在 org_id=5 下，找不到"

## Resolution
root_cause: "pkg/jwt/jwt.go RefreshTokens() 生成新 access token 时硬编码 orgID=0 和空 role。用户在 onboarding 前拿到的 refresh token（只有 userID）无法携带 org 信息，刷新后生成的新 token 丢失了 org_id=5，导致 GetMe 按 org_id=0 查询找不到用户"
fix: |
  1. pkg/jwt/jwt.go: RefreshTokens 增加 userID/orgID/role 参数，
     GenerateAccessToken(claims.UserID, orgID, role, ...)
  2. internal/user/service.go: RefreshToken 先通过 jwt.ParseRefreshToken 获取 userID，
     再调用 repo.FindByID(claims.UserID) 查询用户最新 orgID/role，
     最后传入 jwt.RefreshTokens(user.ID, user.OrgID, user.Role)
  3. internal/user/repository.go: 新增 FindByID（不做 tenant scope）和 UpdateUserOrgID
verification: "go build ./... 通过；go test ./pkg/jwt/... 通过"
files_changed:
  - pkg/jwt/jwt.go
  - pkg/jwt/jwt_test.go
  - internal/user/repository.go
  - internal/user/service.go
