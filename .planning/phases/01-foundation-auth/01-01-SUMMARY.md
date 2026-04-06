# Phase 1 Plan 01-01 Summary

**Plan:** 01-01 — 项目脚手架 + 配置 + 统一响应封装 + 加密工具 + 多租户Scope + 基础中间件
**Wave:** 1
**Completed:** 2026-04-06

## Files Created
| File | Purpose |
|------|---------|
| `go.mod` / `go.sum` | Go 模块 + 15 个核心依赖 |
| `config/config.yaml` | YAML 配置（server/database/redis/jwt/oss/sms/crypto） |
| `internal/common/config/config.go` | Viper 加载配置 + 环境变量覆盖 |
| `internal/common/logger/logger.go` | Zap 结构化日志 |
| `internal/common/response/response.go` | 统一响应封装（Success/Error/PageSuccess/Unauthorized/Forbidden/BadRequest） |
| `internal/common/response/response_test.go` | 6 个响应测试，100% 覆盖率 |
| `internal/common/model/base.go` | BaseModel（ID/OrgID/CreatedBy/CreatedAt/UpdatedBy/UpdatedAt/DeletedAt） |
| `internal/common/model/org.go` | Organization 模型 |
| `internal/common/model/user.go` | User 模型（加密双列模式） |
| `internal/common/database/database.go` | PostgreSQL 连接 + 连接池 |
| `internal/common/crypto/aes.go` | AES-256-GCM 加密/解密 |
| `internal/common/crypto/hash.go` | SHA-256 哈希 |
| `internal/common/crypto/mask.go` | 手机号/身份证脱敏 |
| `internal/common/crypto/crypto_test.go` | 7 个加密测试，79.5% 覆盖率 |
| `internal/common/middleware/tenant.go` | TenantScope GORM Scope |
| `internal/common/middleware/cors.go` | CORS 中间件 |
| `internal/common/middleware/ratelimit.go` | Redis 限流中间件 |
| `internal/common/middleware/logger.go` | 请求日志中间件 |
| `internal/common/middleware/middleware_test.go` | 中间件测试 |
| `cmd/server/main.go` | 应用入口 |
| `docker-compose.yml` | PostgreSQL 16 + Redis 7 开发环境 |
| `Dockerfile` | 多阶段构建 |

## Tests
- `go test ./internal/common/response/... -v` — 6/6 PASS, 100% coverage
- `go test ./internal/common/crypto/... -v` — 7/7 PASS, 79.5% coverage
- `go test ./internal/common/middleware/... -v` — 5/5 PASS, 56.2% coverage
- `go build ./cmd/server` — BUILD OK

## Decisions Made
- Go 1.25.0 (gin v1.12.0 要求)
- GORM AutoMigrate（不用 golang-migrate，V1.0 够用）
- 软删除使用部分唯一索引 `WHERE deleted_at IS NULL`
