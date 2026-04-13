# Phase 1: 基础框架与用户认证 - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-06
**Phase:** 01-foundation-auth
**Areas discussed:** 矛盾修复+Go实现方式, 多设备登录策略, OSS文件上传策略, RBAC权限矩阵

---

## 矛盾修复 + Go实现方式

### D-16 bcrypt 矛盾

| Option | Description | Selected |
|--------|-------------|----------|
| 删除 D-16 | 系统无密码，bcrypt 行无意义 | ✓ |
| 保留但标记未来 | D-16 改为"未来可选" | |
| V1.0 支持密码 | 添加密码注册/登录流程 | |

**User's choice:** 删除 D-16
**Notes:** V1.0 纯短信验证码登录，未来如需密码登录再添加 bcrypt 相关代码

### D-13 Go RBAC 实现方式

| Option | Description | Selected |
|--------|-------------|----------|
| 路由组中间件 | 按 Gin 路由组统一配置 RequireRole | ✓ |
| Handler 内调用 | 每个 handler 显式调用 checkRole | |
| 路由声明式中间件 | 路由定义时声明 RequireRole | |

**User's choice:** 路由组中间件
**Notes:** 如 `adminRoutes.Use(RequireRole("OWNER", "ADMIN"))`，Go 原生风格，声明式且不易遗漏

---

## 多设备登录策略

### 并发会话策略

| Option | Description | Selected |
|--------|-------------|----------|
| 允许多设备并发 | 每设备独立 token，手机+H5可同时用 | ✓ |
| 单设备限制 | 新登录踢掉旧会话 | |
| 限制设备数 | 最多2-3台同时在线 | |

**User's choice:** 允许多设备并发

### Refresh token 安全策略

| Option | Description | Selected |
|--------|-------------|----------|
| Refresh token 轮转 | 每次刷新颁发新 token 并作废旧 token | ✓ |
| 固定 refresh token | 30天内不变，简单但安全性低 | |

**User's choice:** Refresh token 轮转

### 退出登录策略

| Option | Description | Selected |
|--------|-------------|----------|
| 仅当前设备 | 退出时不影响其他设备 | ✓ |
| 全部设备 | 退出时吊销所有 token | |

**User's choice:** 仅当前设备

---

## OSS文件上传策略

### 上传方式

| Option | Description | Selected |
|--------|-------------|----------|
| 客户端签名直传 | 服务端签名URL，客户端直连OSS | ✓ |
| 服务端代理转发 | 客户端→服务端→OSS | |
| 混合模式 | 小文件服务端，大文件直传 | |

**User's choice:** 客户端签名直传

### 文件大小限制

| Option | Description | Selected |
|--------|-------------|----------|
| 图片5MB / 文档20MB | 按类型区分，兼顾实用和安全 | ✓ |
| 统一10MB上限 | 简单但合同PDF可能超限 | |
| 不限制 | 灵活但OSS成本和滥用风险 | |

**User's choice:** 图片5MB / 文档20MB

### 允许文件类型

| Option | Description | Selected |
|--------|-------------|----------|
| 图片+PDF/Excel | 白名单：jpg/png/jpeg + pdf/xlsx | ✓ |
| 不限制类型 | 只校验大小 | |

**User's choice:** 图片(jpg/png/jpeg) + 文档(pdf/xlsx)

### OSS 存储结构

| Option | Description | Selected |
|--------|-------------|----------|
| 按业务类型+org_id | contracts/{org_id}/{date}/{file} | ✓ |
| 单一bucket+前缀 | 统一前缀区分 | |
| 每企业一个bucket | 隔离最强但bucket数量有限 | |

**User's choice:** 按业务类型 + org_id

---

## RBAC权限矩阵

### 精细程度

| Option | Description | Selected |
|--------|-------------|----------|
| 通用规则+后续补充 | Phase 1 定义通用三级规则，各Phase逐步补充 | ✓ |
| Phase 1 全量定义 | 一次性列出所有8个模块权限 | |

**User's choice:** 通用规则 + 后续补充

### 通用权限边界

| Option | Description | Selected |
|--------|-------------|----------|
| 三级权限分层 | OWNER全部、ADMIN增删改查、MEMBER只读 | ✓ |
| 三级+ADMIN可配置 | OWNER可自定义ADMIN权限 | |

**User's choice:** 三级权限分层
**Notes:** OWNER全部权限；ADMIN增删改查但不可管理子账号和修改企业信息；MEMBER只读不可导出敏感数据

---

## Claude's Discretion

- 具体目录结构细节
- 中间件执行顺序
- 错误处理包装方式
- 日志格式和输出目标
- Docker 配置细节
- Refresh token 黑名单的 Redis key 设计
- OSS 签名 URL 有效期

## Deferred Ideas

None — discussion stayed within phase scope
