# Phase 1 Plan 01-02 Summary

**Plan:** 01-02 — JWT 认证工具 + 短信客户端 + OSS 客户端 + Auth 中间件
**Wave:** 2
**Completed:** 2026-04-06

## Files Created
| File | Purpose |
|------|---------|
| `pkg/jwt/jwt.go` | JWT 生成/解析/刷新（AccessToken + RefreshToken 轮转） |
| `pkg/jwt/jwt_test.go` | 7 个 JWT 测试（生成/解析/过期/篡改/刷新/黑名单） |
| `pkg/sms/client.go` | 阿里云短信客户端（HMAC-SHA1 签名 + resty HTTP） |
| `pkg/sms/client_test.go` | 3 个 SMS 测试（初始化/成功/失败，httptest mock） |
| `pkg/oss/client.go` | 阿里云 OSS 签名 URL 客户端（文件大小/类型校验） |
| `pkg/oss/client_test.go` | 4 个 OSS 测试（类型/大小/路径/初始化） |
| `internal/common/middleware/auth.go` | JWT Auth 中间件（Bearer token 解析 + Redis 黑名单检查） |
| `internal/common/middleware/auth_test.go` | 6 个 Auth 测试（缺 header/格式错误/有效/过期/篡改/黑名单） |

## Tests
- `go test ./pkg/jwt/... -v` — 7/7 PASS
- `go test ./pkg/sms/... -v` — 3/3 PASS
- `go test ./pkg/oss/... -v` — 4/4 PASS
- `go test ./internal/common/middleware/... -v` — 全通过（1 SKIP: Redis 不可用）

## Key Design
- Refresh Token 轮转：每次刷新返回新 access + refresh，旧 jti 加入 Redis 黑名单
- Auth 中间件：从 `Authorization: Bearer {token}` 提取，注入 user_id/org_id/role 到 gin.Context
- OSS 文件类型白名单：image/jpeg/png/jpg + application/pdf/xlsx
