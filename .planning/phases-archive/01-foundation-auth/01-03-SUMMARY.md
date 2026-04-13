# Phase 1 Plan 01-03 Summary

**Plan:** 01-03 — 用户认证流程 + RBAC 权限 + 审计日志 + 城市列表 API
**Wave:** 3
**Completed:** 2026-04-06

## Files Created
| File | Purpose |
|------|---------|
| `internal/user/dto.go` | 请求/响应 DTO（SendCode/Login/Refresh/Onboarding/SubAccount/UserInfo） |
| `internal/user/repository.go` | 用户数据访问（FindByPhoneHash/CreateUser/CreateOrg/ListUsers/UpdateRole/DeleteUser） |
| `internal/user/service.go` | 用户业务逻辑（SendCode/Login/RefreshToken/Logout/CompleteOnboarding/CreateSubAccount/ListSubAccounts/UpdateRole/DeleteSubAccount） |
| `internal/user/handler.go` | HTTP Handler（12 个路由注册） |
| `internal/common/middleware/rbac.go` | RequireRole 中间件 |
| `internal/common/middleware/rbac_test.go` | 3 个 RBAC 测试（允许/禁止/无角色） |
| `internal/city/model.go` | 城市种子数据（37 个城市，覆盖 34 省级行政区） |
| `internal/city/handler.go` | 城市列表 API（支持 province 筛选） |
| `internal/city/handler_test.go` | 2 个城市 API 测试 |
| `internal/audit/model.go` | AuditLog 模型（JSONB detail） |
| `internal/audit/handler.go` | 审计日志查询 API（分页/按 module/action 筛选） |
| `cmd/server/main.go` | 更新：注册所有路由 + AutoMigrate + SMS/OSS 初始化 |

## Routes Registered
| Method | Path | Auth | Role |
|--------|------|------|------|
| POST | /api/v1/auth/send-code | No | - |
| POST | /api/v1/auth/login | No | - |
| POST | /api/v1/auth/refresh | No | - |
| POST | /api/v1/auth/logout | Yes | - |
| PUT | /api/v1/org/onboarding | Yes | - |
| GET | /api/v1/users | Yes | owner, admin |
| POST | /api/v1/users | Yes | owner only |
| PUT | /api/v1/users/:id/role | Yes | owner only |
| DELETE | /api/v1/users/:id | Yes | owner only |
| GET | /api/v1/cities | No | - |
| GET | /api/v1/audit-logs | Yes | owner, admin |

## Tests
- `go test ./internal/city/... -v` — 2/2 PASS
- `go test ./internal/common/middleware/... -v` — RBAC 3/3 PASS

## Key Design
- Login 自动注册：手机号不存在 → 创建 Organization(inactive) + User(owner) → onboarding_required=true
- 验证码错误 5 次 → 删除 Redis key，需重新发送
- OWNER 不可删除/不可降级
- 所有 Repository 查询使用 TenantScope(orgID) 自动注入 org_id
