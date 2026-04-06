# Phase 1: 基础框架与用户认证 - Context

**Gathered:** 2026-04-06
**Updated:** 2026-04-06
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
- **D-06:** JWT token认证。Token 包含 `user_id`、`org_id`、`role`。有效期7天。Refresh token 30天，采用轮转策略（每次刷新时颁发新 refresh token 并作废旧 token）。
- **D-07:** 验证码存储于 Redis。Key: `sms:code:{phone}`，TTL 5分钟。限流：同一手机号60秒内最多发1次。
- **D-08:** 首次登录自动注册（手机号不存在则创建用户+企业）。引导流程：登录 → 录入企业信息 → 进入首页。

### 多设备登录策略
- **D-27:** 允许多设备并发登录。老板可能手机+H5后台同时使用，不限制设备数。
- **D-28:** Refresh token 轮转：每次使用 refresh token 获取新 access token 时，同时返回新 refresh token 并将旧 token 加入 Redis 黑名单。防重放攻击。
- **D-29:** 退出登录仅吊销当前设备的 token（不踢其他设备）。提供"退出所有设备"选项但不默认。

### 多租户隔离
- **D-09:** 逻辑多租户。所有业务表包含 `org_id` 字段。GORM 全局 Scope 自动注入 `WHERE org_id = ?` 过滤。
- **D-10:** JWT 从 token 中提取 `org_id`，中间件将 `org_id` 注入请求上下文（`gin.Context`）。Repository 层从上下文获取 `org_id`，不信任客户端传入的 `org_id`。
- **D-11:** 集成测试必须包含多租户隔离验证（创建两个企业，验证数据互不可见）。

### RBAC 权限
- **D-12:** 三级角色：OWNER（老板，全部权限）、ADMIN（管理员，大部分权限）、MEMBER（普通成员，只读为主）。
- **D-13:** 权限检查通过 Gin 路由组中间件统一配置。如 `adminRoutes.Use(RequireRole("OWNER", "ADMIN"))`。Go 原生风格，声明式且不易遗漏。
- **D-14:** OWNER 角色不可删除。每个企业有且仅有一个 OWNER。
- **D-30:** Phase 1 定义通用权限规则，具体模块权限在各 Phase 中逐步补充。通用规则：
  - OWNER：全部权限（增删改查、管理子账号、企业设置、数据导出）
  - ADMIN：增删改查业务数据，不可管理子账号，不可修改企业信息
  - MEMBER：只读查看，不可导出敏感数据（身份证号、薪资等）

### 数据安全
- **D-15:** 敏感字段（手机号、身份证号）使用 AES-256-GCM 加密存储。同时存储 SHA-256 哈希值用于精确查询。
- **D-17:** API 响应中敏感字段（手机号、身份证号）返回脱敏数据（如 `138****5678`）。
- **D-31:** JWT secret 从环境变量读取。

### OSS 文件上传
- **D-32:** 客户端签名直传。服务端生成签名 URL，客户端直连 OSS 上传。节省服务器带宽，上传速度快。
- **D-33:** 文件大小限制：图片 5MB，文档（PDF/Excel）20MB。
- **D-34:** 允许文件类型白名单：图片（jpg/png/jpeg）、文档（pdf/xlsx）。服务端签名时校验。
- **D-35:** OSS 存储按业务类型 + org_id 组织。结构：`{业务类型}/{org_id}/{日期}/{文件名}`。如 `contracts/org_123/2026-04/contract.pdf`。

### API 设计
- **D-18:** RESTful API 风格。统一响应格式：`{"code": 0, "message": "success", "data": {...}}`。
- **D-19:** 错误码规范：0=成功，4xx=客户端错误。5xx=服务端错误。分模块定义错误码（如 10xxx=用户模块，20xxx=员工模块）。
- **D-20:** 分页参数统一：`page`（从1开始）、`page_size`（默认20，最大100）。
- **D-21:** API 版本前缀：`/api/v1/`。所有接口挂在 v1 下。

### 审计日志
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
- Refresh token 黑名单的 Redis key 设计
- OSS 签名 URL 有效期

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
  - `pkg/jwt/` — JWT 工具（含 refresh token 轮转）
  - `pkg/oss/` — OSS 客户端（签名 URL 生成）
  - `pkg/sms/` — 短信客户端

</code_context>

<specifics>
## Specific Ideas

No specific requirements — open to standard approaches

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---
*Phase: 01-foundation-auth*
*Context gathered: 2026-04-06*
*Updated: 2026-04-06*
