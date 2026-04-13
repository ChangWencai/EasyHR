# Phase 1: 基础框架与用户认证 - Research

**Researched:** 2026-04-06
**Domain:** Go 后端项目脚手架、多租户隔离、认证、RBAC、审计日志、OSS文件上传
**Confidence:** HIGH

## Summary

Phase 1 是整个 EasyHR 系统的基础，需要搭建 Go 模块化单体项目脚手架，实现多租户数据隔离（org_id 全链路透传）、手机号+短信验证码登录、企业信息引导录入、OWNER/ADMIN/MEMBER 三级 RBAC 权限、GORM 钩子驱动的审计日志、AES-256-GCM 敏感字段加密，以及阿里云 OSS 文件上传。所有后续业务模块将继承本阶段建立的中间件、统一响应、加密工具等基础设施。

**Primary recommendation:** 使用 Gin v1.12 + GORM v1.31 + PostgreSQL 16 搭建模块化单体，通过 GORM Scope 实现自动多租户隔离，Gin 中间件链处理认证/RBAC/审计，Redis 存储验证码和 Token 黑名单。

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** Go 模块化单体，`internal/` 下按业务模块组织，`cmd/server/main.go` 入口，公共包放 `pkg/`。
- **D-02:** 每个业务模块遵循 handler -> service -> repository 三层架构。
- **D-03:** Go 1.22+ + Gin v1.12+ + GORM v1.31.1。
- **D-04:** 配置使用 `config/config.yaml`，YAML 格式，环境变量覆盖敏感配置。
- **D-05:** 手机号+短信验证码登录，无密码。验证码6位数字，有效期5分钟。
- **D-06:** JWT token 认证。Token 含 `user_id`、`org_id`、`role`。有效期7天。Refresh token 30天。
- **D-07:** 验证码存储 Redis。Key: `sms:code:{phone}`，TTL 5分钟。限流：同一手机号60秒内最多发1次。
- **D-08:** 首次登录自动注册。引导流程：登录 -> 录入企业信息 -> 进入首页。
- **D-09:** 逻辑多租户。所有业务表含 `org_id`。GORM 全局 Scope 自动注入 `WHERE org_id = ?`。
- **D-10:** JWT 从 token 提取 `org_id`，中间件注入 `gin.Context`。Repository 层从上下文获取，不信任客户端传入。
- **D-11:** 集成测试必须含多租户隔离验证。
- **D-12:** 三级角色：OWNER/ADMIN/MEMBER。权限检查在中间件层统一处理。
- **D-13:** 通过注解方式标注角色需求（Go 中用 Gin 中间件函数包裹路由）。
- **D-14:** OWNER 不可删除。每个企业仅一个 OWNER。
- **D-15:** 敏感字段 AES-256-GCM 加密存储 + SHA-256 哈希索引。
- **D-16:** 密码哈希 bcrypt（cost=10）。JWT secret 从环境变量读取。
- **D-17:** API 响应中敏感字段返回脱敏数据。
- **D-18:** RESTful API，统一响应：`{"code": 0, "message": "success", "data": {...}}`。
- **D-19:** 错误码：0=成功，4xx/5xx。分模块定义（10xxx=用户，20xxx=员工）。
- **D-20:** 分页：`page`(从1开始)、`page_size`(默认20，最大100)。
- **D-21:** API 前缀 `/api/v1/`。
- **D-22:** GORM Hook 自动记录写操作。审计字段：org_id, user_id, module, action, target_type, target_id, detail(JSONB), ip_address, created_at。
- **D-23:** 审计日志 INSERT ONLY，不提供删除/修改接口。
- **D-24:** PostgreSQL 15+，GORM AutoMigrate 管理迁移。
- **D-25:** 软删除 `deleted_at`，唯一约束使用部分索引 `WHERE deleted_at IS NULL`。
- **D-26:** 审计字段：created_by, created_at, updated_by, updated_at。

### Claude's Discretion
- 具体目录结构细节
- 中间件执行顺序
- 错误处理包装方式
- 日志格式和输出目标（stdout/file）
- Docker 配置细节

### Deferred Ideas (OUT OF SCOPE)
None
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| AUTH-01 | 老板通过手机号+验证码一键登录/注册 | D-05/D-07/D-08: Redis验证码存储+限流+自动注册 |
| AUTH-02 | 首次登录自动进入企业信息录入引导页 | D-08: 首次登录检测+企业引导流程 |
| AUTH-03 | 城市选择器自动定位当前城市 | 前端功能，后端提供城市列表API |
| AUTH-04 | JWT token认证，会话持久化，多设备登录 | D-06: JWT含user_id/org_id/role，7天+30天刷新 |
| PLAT-01 | 多子账号管理，OWNER/ADMIN/MEMBER三级RBAC | D-12/D-13/D-14: 中间件角色校验+RBAC矩阵 |
| PLAT-02 | 操作日志全程记录 | D-22/D-23: GORM Hook审计日志 |
| PLAT-03 | 统一响应封装、错误处理、请求日志 | D-18/D-19: 统一JSON响应+错误码规范 |
| PLAT-04 | 限流/鉴权/CORS/SSL中间件 | Gin中间件链: RateLimit+Auth+CORS+TLS |
| PLAT-05 | 敏感字段AES-256-GCM加密+SHA-256哈希索引 | D-15/D-17: crypto标准库双列模式 |
| PLAT-06 | 数据库逻辑多租户隔离(org_id全链路透传) | D-09/D-10/D-11: GORM Scope自动注入 |
| PLAT-07 | 文件上传至阿里云OSS | 阿里云OSS SDK v3签名URL直传 |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go | 1.23.4 | 编程语言 | 环境已安装，满足1.22+要求 |
| Gin | v1.12.0 | HTTP框架 | 中国最流行Go Web框架，中间件生态丰富 |
| GORM | v1.31.1 | ORM | Scope机制天然适合多租户，软删除/Hook内建 |
| gorm.io/driver/postgres | v1.5.x | PostgreSQL驱动 | GORM官方PostgreSQL驱动 |
| go-redis | v9.18.0 | Redis客户端 | 验证码存储、Token黑名单、限流计数 |
| golang-jwt/jwt/v5 | v5.3.1 | JWT认证 | 官方维护，v5 API简洁安全 |
| Viper | v1.21.0 | 配置管理 | YAML配置+环境变量覆盖 |
| Zap | v1.27.1 | 结构化日志 | Uber出品，性能极高 |
| validator | v10.30.2 | 参数校验 | Gin生态标配，struct tag声明式 |
| Aliyun OSS SDK | v3.0.2 | 文件存储 | 签名URL直传 |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| testify | v1.11.1 | 测试断言 | 所有单元/集成测试 |
| golang-migrate | v4.19.1 | 数据库迁移 | 可选，GORM AutoMigrate也可满足V1.0 |
| resty | v2.17.2 | HTTP客户端 | 调用阿里云短信API |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| casbin RBAC | 自定义中间件 | V1.0只有3个固定角色，中间件足够简单直接；casbin增加依赖和复杂度 |
| golang-migrate | GORM AutoMigrate | D-24已锁定AutoMigrate，迁移工具可后续引入 |

**Installation (go.mod 核心依赖):**
```bash
go get github.com/gin-gonic/gin@v1.12.0
go get gorm.io/gorm@v1.31.1
go get gorm.io/driver/postgres
go get github.com/redis/go-redis/v9
go get github.com/golang-jwt/jwt/v5
go get github.com/spf13/viper@v1.21.0
go get go.uber.org/zap@v1.27.1
go get github.com/go-playground/validator/v10
go get github.com/aliyun/aliyun-oss-go-sdk/v3
go get github.com/go-resty/resty/v2
go get github.com/stretchr/testify
```

## Architecture Patterns

### Recommended Project Structure
```
easyhr/
├── cmd/server/main.go           # 入口：初始化配置/DB/Redis/路由/中间件
├── config/config.yaml           # 配置文件
├── internal/
│   ├── common/
│   │   ├── middleware/           # auth.go, rbac.go, cors.go, ratelimit.go, audit.go, tenant.go
│   │   ├── response/            # response.go (统一响应封装)
│   │   ├── crypto/              # aes.go, hash.go, mask.go (加密/哈希/脱敏)
│   │   └── logger/              # zap初始化
│   └── user/                    # 用户/组织/权限模块
│       ├── handler.go           # HTTP Handler (Gin路由处理)
│       ├── service.go           # 业务逻辑
│       ├── repository.go        # 数据访问 (GORM)
│       ├── model.go             # 数据模型 (GORM Model)
│       └── dto.go               # 请求/响应DTO
├── pkg/
│   ├── jwt/                     # JWT生成/验证/刷新
│   ├── sms/                     # 阿里云短信客户端
│   └── oss/                     # 阿里云OSS客户端
├── migrations/                  # 数据库迁移文件
├── Dockerfile
├── docker-compose.yml           # PostgreSQL + Redis 开发环境
├── go.mod
└── go.sum
```

### Pattern 1: GORM Scope 多租户隔离 (CRITICAL)
**What:** Repository层所有查询自动追加 org_id 条件
**When:** 所有业务模块数据访问
```go
// internal/common/middleware/tenant.go
func TenantScope(orgID int64) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("org_id = ?", orgID)
    }
}

// 使用：所有 Repository 查询
db.Scopes(TenantScope(orgID)).Where("status = ?", "ACTIVE").Find(&employees)
```

### Pattern 2: 加密字段双列模式
**What:** 敏感字段存储加密值+哈希值
```go
// model.go
type User struct {
    ID        int64  `gorm:"primaryKey"`
    Phone     string `gorm:"column:phone;type:varchar(200)"`      // AES加密值
    PhoneHash string `gorm:"column:phone_hash;type:varchar(64);uniqueIndex"` // SHA-256
    // ...
}

// 查找时用哈希
func (r *userRepo) FindByPhone(phone string) (*User, error) {
    hash := crypto.HashSHA256(phone)
    return r.db.Where("phone_hash = ?", hash).First(&User{}).Error
}
```

### Pattern 3: Gin 中间件 RBAC
**What:** 路由级别角色检查
```go
// internal/common/middleware/rbac.go
func RequireRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := c.GetString("role")
        for _, r := range roles {
            if userRole == r { c.Next(); return }
        }
        response.Forbidden(c, "权限不足")
        c.Abort()
    }
}

// 路由注册
v1.POST("/users", middleware.RequireRole("OWNER", "ADMIN"), handler.CreateUser)
```

### Pattern 4: 统一响应封装
```go
// internal/common/response/response.go
func Success(c *gin.Context, data interface{}) {
    c.JSON(200, gin.H{"code": 0, "message": "success", "data": data})
}
func Error(c *gin.Context, httpStatus int, code int, msg string) {
    c.JSON(httpStatus, gin.H{"code": code, "message": msg, "data": nil})
}
func PageSuccess(c *gin.Context, list interface{}, total int64, page, pageSize int) {
    c.JSON(200, gin.H{
        "code": 0, "message": "success",
        "data": list,
        "meta": gin.H{"total": total, "page": page, "page_size": pageSize},
    })
}
```

### Pattern 5: 审计日志 GORM Hook
**What:** 通过 GORM Callback 自动记录写操作
```go
// 注册 GORM Callback
db.Callback().Create().After("gorm:create").Register("audit:log", auditLogCallback)
db.Callback().Update().After("gorm:update").Register("audit:log", auditLogCallback)
db.Callback().Delete().After("gorm:delete").Register("audit:log", auditLogCallback)
```

### Anti-Patterns to Avoid
- **跨模块直接访问 Repository:** 破坏模块边界，走 Service 接口
- **Handler 包含业务逻辑:** Handler 只做参数解析+调Service+返回响应
- **忘记 org_id 过滤:** 任何查询遗漏即数据泄露，必须用 Scope 自动注入
- **信任客户端传入 org_id:** 必须从 JWT 提取，不信任请求参数

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| JWT生成/验证 | 手写JWT逻辑 | golang-jwt/jwt/v5 | 安全性要求高，边缘情况多 |
| AES-256-GCM加密 | 自定义加密实现 | crypto/aes + crypto/cipher 标准库 | 标准库经过审计，nonce/padding处理正确 |
| 参数校验 | 手写if/else校验 | go-playground/validator | 声明式、可扩展、中文错误消息 |
| 限流 | 自定义计数器 | 基于 Redis + Gin中间件 | 需要分布式计数，Redis原子操作 |
| 短信发送 | 直接HTTP调用 | resty + 阿里云SDK | 重试/超时/签名处理 |
| OSS签名URL | 自定义签名算法 | 阿里云OSS SDK | 签名算法复杂，SDK处理V4签名 |

## Common Pitfalls

### Pitfall 1: 多租户数据泄露 (CRITICAL)
**What goes wrong:** 某条查询遗漏 org_id 过滤，A企业看到B企业数据
**Why it happens:** 手动拼接 org_id 容易遗漏，尤其复杂查询
**How to avoid:** GORM 全局 Scope 自动注入 + 集成测试覆盖双租户场景 (D-11)
**Warning signs:** Repository 有裸 `db.Where()` 而非 `db.Scopes(TenantScope())`

### Pitfall 2: 验证码可暴力破解
**What goes wrong:** 攻击者短时间内大量尝试6位验证码
**Why it happens:** 缺少尝试次数限制
**How to avoid:** 验证码错误5次后删除Redis key，需重新发送；发送间隔60秒限制 (D-07)
**Warning signs:** 无 Redis 计数器或计数器无 TTL

### Pitfall 3: JWT Token 无法主动失效
**What goes wrong:** 用户退出登录或被禁用后 Token 仍有效
**Why it happens:** JWT 无状态设计不支持服务端撤销
**How to avoid:** Redis 维护 Token 黑名单（jti -> TTL），退出时将 jti 加入黑名单

### Pitfall 4: GORM 软删除与唯一约束冲突
**What goes wrong:** 软删除后重新创建同名记录违反唯一约束
**Why it happens:** 默认唯一约束不排除 deleted_at IS NOT NULL 的记录
**How to avoid:** 使用部分唯一索引 `WHERE deleted_at IS NULL` (D-25)

### Pitfall 5: AES加密密钥硬编码
**What goes wrong:** 密钥写在代码或配置文件中提交到 Git
**Why it happens:** 开发便利性压过安全意识
**How to avoid:** 密钥从环境变量读取，Docker 部署时注入 (D-04/D-16)

## Code Examples

### JWT Token 生成与中间件验证
```go
// pkg/jwt/jwt.go
type Claims struct {
    UserID int64  `json:"user_id"`
    OrgID  int64  `json:"org_id"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

func GenerateToken(userID, orgID int64, role, secret string, ttl time.Duration) (string, error) {
    claims := Claims{
        UserID: userID, OrgID: orgID, Role: role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

// internal/common/middleware/auth.go
func Auth(jwtSecret string, rdb *redis.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenStr := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
        claims, err := jwt.ParseToken(tokenStr, jwtSecret)
        if err != nil { response.Unauthorized(c, "无效Token"); c.Abort(); return }
        // 检查黑名单
        if exists, _ := rdb.Exists(ctx, "token:blacklist:"+claims.ID).Result(); exists > 0 {
            response.Unauthorized(c, "Token已失效"); c.Abort(); return
        }
        c.Set("user_id", claims.UserID)
        c.Set("org_id", claims.OrgID)
        c.Set("role", claims.Role)
        c.Next()
    }
}
```

### AES-256-GCM 加密/解密
```go
// internal/common/crypto/aes.go
func Encrypt(plaintext string, key []byte) (string, error) {
    block, _ := aes.NewCipher(key)
    gcm, _ := cipher.NewGCM(block)
    nonce := make([]byte, gcm.NonceSize())
    io.ReadFull(rand.Reader, nonce)
    ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(encoded string, key []byte) (string, error) {
    data, _ := base64.StdEncoding.DecodeString(encoded)
    block, _ := aes.NewCipher(key)
    gcm, _ := cipher.NewGCM(block)
    nonceSize := gcm.NonceSize()
    plaintext, _ := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
    return string(plaintext), nil
}
```

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | 后端编译 | ✓ | 1.23.4 | -- |
| PostgreSQL | 数据存储 | ✗ | -- | docker-compose 开发环境 |
| Redis | 验证码/缓存 | ✗ | -- | docker-compose 开发环境 |
| Docker | 容器化 | ✓ | 29.3.1 | -- |
| Node.js | 前端(H5) | ✓ | 22.14.0 | -- |
| psql CLI | DB管理 | ✗ | -- | Docker exec |
| redis-cli | 缓存调试 | ✗ | -- | Docker exec |

**Missing dependencies with fallback:**
- PostgreSQL 和 Redis 通过 docker-compose 启动开发环境（Phase 1 必须创建 docker-compose.yml）

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + testify v1.11.1 |
| Config file | none -- Go 原生 testing |
| Quick run command | `go test ./internal/... -v -short` |
| Full suite command | `go test ./... -v -cover` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| AUTH-01 | 手机号验证码登录/注册 | integration | `go test ./internal/user/... -run TestLogin -v` | Wave 0 |
| AUTH-02 | 首次登录企业引导 | integration | `go test ./internal/user/... -run TestOnboarding -v` | Wave 0 |
| AUTH-04 | JWT认证多设备 | unit | `go test ./pkg/jwt/... -v` | Wave 0 |
| PLAT-01 | RBAC角色校验 | unit | `go test ./internal/common/middleware/... -run TestRBAC -v` | Wave 0 |
| PLAT-02 | 审计日志记录 | integration | `go test ./internal/common/... -run TestAudit -v` | Wave 0 |
| PLAT-03 | 统一响应封装 | unit | `go test ./internal/common/response/... -v` | Wave 0 |
| PLAT-05 | AES加密/脱敏 | unit | `go test ./internal/common/crypto/... -v` | Wave 0 |
| PLAT-06 | 多租户隔离 | integration | `go test ./internal/... -run TestTenantIsolation -v` | Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/... -v -short`
- **Per wave merge:** `go test ./... -v -cover`
- **Phase gate:** Full suite green before verify

### Wave 0 Gaps
- [ ] `internal/user/handler_test.go` -- AUTH-01, AUTH-02
- [ ] `internal/user/service_test.go` -- AUTH-01 业务逻辑
- [ ] `internal/user/repository_test.go` -- PLAT-06 租户隔离
- [ ] `pkg/jwt/jwt_test.go` -- AUTH-04 Token生成/验证
- [ ] `internal/common/middleware/rbac_test.go` -- PLAT-01
- [ ] `internal/common/middleware/audit_test.go` -- PLAT-02
- [ ] `internal/common/response/response_test.go` -- PLAT-03
- [ ] `internal/common/crypto/aes_test.go` -- PLAT-05
- [ ] `test/testutil/db.go` -- 测试DB初始化fixture

## Sources

### Primary (HIGH confidence)
- tech-architecture.md -- 完整技术架构、数据模型、API设计、安全策略
- prd.md -- 产品需求文档
- .planning/research/STACK.md -- 技术栈选型及版本验证
- .planning/research/ARCHITECTURE.md -- 模块化单体架构模式
- .planning/research/PITFALLS.md -- 常见陷阱（多租户隔离#5、敏感数据#6）
- CONTEXT.md -- 用户锁定决策（D-01到D-26）

### Secondary (MEDIUM confidence)
- Go 标准库 crypto/aes + crypto/cipher -- AES-256-GCM 用法（基于训练数据）
- GORM v2 Scope/Hook/SoftDelete 文档（基于训练数据）
- Gin 中间件模式（基于训练数据）

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- 版本已在 STACK.md 中通过 GitHub Release API 验证
- Architecture: HIGH -- ARCHITECTURE.md 提供详细模块边界和数据流
- Pitfalls: HIGH -- PITFALLS.md 覆盖多租户和加密陷阱
- Environment: HIGH -- 已实际验证本机工具版本

**Research date:** 2026-04-06
**Valid until:** 2026-05-06 (30 days, stable Go ecosystem)
