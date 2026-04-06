# Phase 1 Plan 01-04 Summary

**Plan:** 01-04 — Token 刷新/退出 + 集成测试
**Wave:** 4
**Completed:** 2026-04-06

## Files Created
| File | Purpose |
|------|---------|
| `test/testutil/db.go` | 测试数据库初始化（SQLite 内存模式）+ 数据工厂 |
| `test/testutil/redis.go` | 测试 Redis 初始化 + 等待/清理 |
| `test/integration/auth_test.go` | 集成测试（认证流程/城市列表/认证保护/RBAC） |

## Tests
- `go test ./test/integration/... -v` — 编译通过，运行时 SKIP（Redis 未启动）
- 当 `docker-compose up` 启动 Redis 后，集成测试将自动执行

## Integration Test Coverage
| Test | What it verifies |
|------|-----------------|
| TestIntegrationAuthFlow | 验证码登录 → 自动注册 → 返回 token + onboarding_required |
| TestIntegrationCityList | 城市列表 API 返回完整数据 |
| TestIntegrationAuthRequired | 未认证访问受保护路由 → 401 |
| TestIntegrationRBACMemberCannotCreateUser | MEMBER 创建子账号 → 403 |
| TestIntegrationRBACMemberCanListUsers | MEMBER 查看用户列表 → 200 |
| TestIntegrationOwnerCanCreateUser | OWNER 创建子账号 → 200 |

## Build Verification
- `go build ./...` — BUILD OK
- `go test ./internal/common/... ./pkg/... ./internal/city/... -short` — 7/7 packages PASS
