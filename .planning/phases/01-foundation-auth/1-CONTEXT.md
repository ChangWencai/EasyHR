# Phase 1: 基础框架与用户认证 - Context

**Gathered:** 2026-04-06
**Status:** Ready for planning

<domain>
## Phase Boundary

项目脚手架搭建、多租户数据隔离、用户认证（手机号+验证码）、企业信息录入引导、RBAC权限管理（OWNER/ADMIN/MEMBER）、审计日志全程记录。API统一响应封装。敏感数据加密。文件上传至OSS。

</domain>

<decisions>
## Implementation Decisions

### 项目结构
- **D-01:** Go 模块化单体，采用 `internal/` 下按业务模块组织，`cmd/server/main.go` 入口，公共包放 `pkg/`。
- **D-02:** 每个业务模块遵循 handler → service → repository 三层架构。
- **D-03:** 使用 Go 1.22+ + Gin v1.12+ + GORM v1.31.1（研究确认选型）。
- **D-04:** 配置使用 `config/config.yaml`，YAML格式，环境变量覆盖敏感配置。

### 认证流程
- **D-05:** 登录方式为手机号+短信验证码，无密码登录。验证码6位数字，有效期5分钟。
- **D-06:** JWT token认证。Token 包含 `user_id`、`org_id`、`role`。有效期7天。Refresh token 30天。
- **D-07:** 验证码存储于 Redis。Key: `sms:code:{phone}`，TTL 5分钟。限流：同一手机号60秒内最多发1次。
- **D-08:** 首次登录自动注册（手机号不存在则创建用户+企业）。引导流程：登录 → 录入企业信息 → 进入首页。

### 多租户隔离
- **D-09:** 逻辑多租户。所有业务表包含 `org_id` 字段。GORM 全局 Scope 自动注入 `WHERE org_id = ?` 过滤。
- **D-10:** JWT 从 token 中提取 `org_id`，中间件将 `org_id` 注入请求上下文（`gin.Context`）。Repository 层从上下文获取 `org_id`，不信任客户端传入的 `org_id`。
- **D-11:** 集成测试必须包含多租户隔离验证（创建两个企业，验证数据互不可见）。

### RBAC 权限
- **D-12:** 三级角色：OWNER（老板，全部权限）、ADMIN（管理员，大部分权限）、MEMBER（普通成员，只读为主）。
- **D-13:** 权限检查在中间件层统一处理。通过注解 `@RequireRole("OWNER")` 标注需要的角色。
- **D-14:** OWNER 角色不可删除。每个企业有且仅有一个 OWNER。

### 数据安全
- **D-15:** 敏感字段（手机号、身份证号）使用 AES-256-GCM 加密存储。同时存储 SHA-256 哈希值用于精确查询。
- **D-16:** 密码哈希使用 bcrypt（cost=10）。JWT secret 从环境变量读取。
- **D-17:** API 响应中敏感字段（手机号、身份证号）返回脱敏数据（如 `138****5678`）。

### API 设计
- **D-18:** RESTful API 风格。统一响应格式：`{"code": 0, "message": "success", "data": {...}}`。
- **D-19:** 错误码规范：0=成功，4xx=客户端错误。5xx=服务端错误。分模块定义错误码（如 10xxx=用户模块，20xxx=员工模块）。
- **D-20:** 分页参数统一：`page`（从1开始）、`page_size`（默认20，最大100）。
- **D-21:** API 版本前缀：`/api/v1/`。所有接口挂在 v1 下。

### 宰计日志
- **D-22:** 通过 GORM 钩子（Hook）自动记录所有写操作。记录字段：`org_id`, `user_id`, `module`, `action`, `target_type`, `target_id`, `detail`(JSONB), `ip_address`, `created_at`。
- **D-23:** 日志只增不改（INSERT ONLY）。不提供删除/修改日志的接口。

### 数据库
- **D-24:** PostgreSQL 15+。使用 GORM AutoMigrate 管理迁移。
- **D-25:** 软删除使用 `deleted_at` 字段。涉及唯一约束的表使用部分唯一索引（`WHERE deleted_at IS NULL`）。
- **D-26:** 审计字段统一：`created_by`, `created_at`, `updated_by`, `updated_at`。

### Claude's Discretion
- 具体目录结构细节
- 中间件执行顺序
- 错误处理包装方式
- 日志格式和输出目标（stdout/file）
- Docker 配置细节

</decisions>

<canonical_refs>
## Canonical References

### 项目规范
- `prd.md` — 产品需求文档，V1.0功能范围、验收标准
- `ui-ux.md` — UI/UX设计原型，登录流程、首页布局、视觉规范
- `tech-architecture.md` — 技术架构设计、数据模型、模块结构、API设计

### 研究报告
- `.planning/research/STACK.md` — 技术栈选型建议（Gin v1.12.0、GORM v1.31.1、go-redis等）
- `.planning/research/ARCHITECTURE.md` — 架构设计建议、模块边界、构建顺序
- `.planning/research/PITFALLS.md` — 常见陷阱（多租户隔离、敏感数据加密、界面复杂度）

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- 无（Greenfield项目，无现有代码）

### Established Patterns
- 无（Phase 1 建立基础模式）

### Integration Points
- Phase 1 是所有后续模块的基础。所有业务模块将继承：
  - `internal/common/middleware/` — 鉴权、限流、日志、CORS 中间件
  - `internal/common/response/` — 统一响应封装
  - `internal/common/crypto/` — 加密工具
  - `pkg/jwt/` — JWT 工具
  - `pkg/oss/` — OSS 客户端
  - `pkg/sms/` — 短信客户端

</code_context>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---
*Phase: 01-foundation-auth*
*Context gathered: 2026-04-06*
