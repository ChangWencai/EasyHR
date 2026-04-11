---
gsd_state_version: 1.0
milestone: v1.1
milestone_name: 老板登录界面优化
status: executing
stopped_at: Phase 01a-login-boss complete, Phase 03a-web context ready
last_updated: "2026-04-11T13:00:00.000Z"
last_activity: 2026-04-11 -- Phase 01a-login-boss execution complete (2/2 plans)
progress:
  total_phases: 1
  completed_phases: 1
  total_plans: 2
  completed_plans: 2
  percent: 100
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-04-06)

**Core value:** 简单、好用、省时间 -- 老板3步完成核心人事操作，无需专业知识
**Current focus:** Phase 01-login-boss (v1.1) — 老板专属登录页，Plans ready

## Current Position

Phase: 01-login-boss (v1.1) — 老板专属登录页
Plan: 0 of 2 (01-01: LoginView.vue + Auth Guard, 01-02: Password login + /auth/me)
Status: Executing
Last activity: 2026-04-11 -- Phase 01-login-boss planning complete

Progress: [▓░░░░░░░░░] 0% (1/2 phases, 0/0 plans)
v1.1: Phase 01 Context ✅ | Phase 02 ⏸ deferred

## Performance Metrics

**Velocity:**

- Total plans completed: 27
- Average duration: ~15 min/plan
- Total execution time: ~7 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| Phase 1 | 4 | 4 | ~15 min |
| Phase 2 | 4 | 4 | ~15 min |
| Phase 3 | 3 | 3 | ~20 min |
| Phase 4 | 2 | 2 | ~12 min |
| Phase 5 | 3 | 3 | ~1h |
| Phase 6 | 4 | 4 | ~30 min |
| Phase 7 | 2 | 2 | ~18 min |
| Phase 8 | 2 | 2 | ~15 min |
| Phase 9 | 0 | 3 | - |

**Recent Trend:**

- Last 4 plans: all PASS
- Trend: steady

| Phase 02 P01 | 10min | 2 tasks | 9 files |
| Phase 03 P01 | 23 min | 2 tasks | 10 files |
| Phase 04 P01 | 12 | 2 tasks | 8 files |
| Phase 04 P02 | 12min | 2 tasks | 10 files |
| Phase 05-salary P01 | 1h | 2 tasks | 12 files |
| Phase 05-salary P02 | 0 | 2 tasks | 3 files |
| Phase 05-salary P02 | 1200 | 2 tasks | 3 files |
| Phase 08 P01 | ~15min | 2 tasks | 8 files |
| Phase 08 P02 | ~15min | 2 tasks | 10 files |

## Accumulated Context

### Decisions

- Go 1.25.0 (gin v1.12.0 要求)
- GORM AutoMigrate（V1.0 够用，不用 golang-migrate）
- 软删除使用部分唯一索引 `WHERE deleted_at IS NULL`
- 多设备并发登录，不限制设备数
- Refresh Token 轮转策略（每次刷新颁发新 token，旧 token 黑名单）
- 逻辑多租户（org_id），GORM Scope 自动注入
- 敏感字段双列模式（AES-256-GCM 加密值 + SHA-256 哈希索引）
- 三级 RBAC：OWNER（全部）/ ADMIN（大部分）/ MEMBER（只读）
- 审计日志 GORM Hook 自动记录，INSERT ONLY
- [Phase 02]: Repository层事务校验唯一性（兼容SQLite测试和PostgreSQL生产）
- [Phase 02]: LIKE替代ILIKE用于姓名/岗位搜索（SQLite兼容）
- [Phase 02]: StatusProbation常量补充完整员工生命周期
- [Phase 03]: 政策库为全局共享数据(OrgID=0)，不使用TenantScope — 社保政策是全国统一数据，所有企业共用同一套政策库，参保记录才按org_id隔离
- [Phase 03]: 政策库为全局共享数据(OrgID=0)，不使用TenantScope — 社保政策是全国统一数据，所有企业共用同一套政策库，参保记录才按org_id隔离
- [Phase 04]: TaxBracket uses OrgID=0 global data pattern for nationally standardized tax rates
- [Phase 04]: Special deductions: 6 monthly types only (excludes serious illness per D-07 research)
- [Phase 04]: TaxCalculator interface accepts grossIncome parameter for Phase 5 decoupling
- [Phase 04]: ContractRepo DI ordering moved before tax module to resolve compile dependency
- [Phase 04]: TaxReminder deduplication by org_id + year + month (one reminder per org per month)
- [Phase 04]: GetMyTaxRecords returns error stub, needs Phase 5 user-employee mapping

### Roadmap Evolution

- Phase 01 added: 新增登陆界面，该登陆界面只运行老板登陆
- Phase 02 added: 新增登陆界面，该登陆界面只允许老板账户登陆 (2026-04-11)
- Phase 03 added: web登陆界面中添加注册按钮以及流程 (2026-04-11)
- Phase 01 discussed: 3种登录方式（手机+验证码/密码/微信OAuth），OWNER+ADMIN允许，MEMBER拒绝，Auth Guard，首次引导分流 (2026-04-11)
- Phase 02 deferred: Phase 01 完成后重新定义范围（当前与 Phase 01 描述重复）

### Phase 1 Deliverables

- 40 Go files, 7 test packages all PASS
- Complete auth flow: SMS code login → auto-register → onboarding → token refresh → logout
- RBAC middleware with OWNER/ADMIN/MEMBER roles
- Multi-tenant isolation via GORM TenantScope
- Audit logging via GORM callback
- City list API (37 cities)
- JWT/SMS/OSS pkg libraries
- Docker dev environment (PostgreSQL 16 + Redis 7)

### Blockers/Concerns

- Phase 1 (新增登陆界面): 需设计老板专属登录页，员工通过微信小程序登录（H5管理后台仅限老板）
- 集成测试需要 Redis 运行（docker-compose up）

## Session Continuity

Last session: 2026-04-11T11:32:00.000Z
Stopped at: Phase 01 added to v1.1
Next step: /gsd-plan-phase 01
