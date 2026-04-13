---
gsd_state_version: 1.0
milestone: v1.2
milestone_name: H5 管理后台 UI 重构
status: planning
stopped_at: Milestone v1.2 roadmap created
last_updated: "2026-04-14"
last_activity: 2026-04-14
progress:
  total_phases: 4
  completed_phases: 0
  total_plans: 0
  completed_plans: 0
  percent: 0
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-04-14)

**Core value:** 简单、好用、省时间 -- 老板3步完成核心人事操作，无需专业知识
**Current focus:** v1.2 — H5 管理后台 UI 重构

## Current Position

Phase: Not started (planning)
Plan: —
Status: Planning
Last activity: 2026-04-14 — Roadmap created

Progress: [░░░░░░░░░░] 0% (0/4 phases, 0/0 plans)

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
- [Phase 04]: TaxBracket uses OrgID=0 global data pattern for nationally standardized tax rates
- [Phase 04]: Special deductions: 6 monthly types only (excludes serious illness per D-07 research)
- [Phase 04]: TaxCalculator interface accepts grossIncome parameter for Phase 5 decoupling
- [Phase 04]: ContractRepo DI ordering moved before tax module to resolve compile dependency
- [Phase 04]: TaxReminder deduplication by org_id + year + month (one reminder per org per month)
- [Phase 04]: GetMyTaxRecords returns error stub, needs Phase 5 user-employee mapping
- [v1.2]: H5 UI 重构遵循 web-design/EasyHR-web.pen 原型图设计
- [v1.2]: 主色调商务蓝 #4F6EF7，卡片圆角 12px，侧边栏固定 220px

### Roadmap Evolution

- v1.0: 后端核心 API + 原生 APP MVP
- v1.1: H5 管理后台基础框架（路由/布局/登录/员工管理）
- v1.2: H5 管理后台 UI 重构（全面对齐原型图）

### Phase 1 Deliverables (v1.1)

- 40 Go files, 7 test packages all PASS
- Complete auth flow: SMS code login → auto-register → onboarding → token refresh → logout
- RBAC middleware with OWNER/ADMIN/MEMBER roles
- Multi-tenant isolation via GORM TenantScope
- Audit logging via GORM callback
- City list API (37 cities)
- JWT/SMS/OSS pkg libraries
- Docker dev environment (PostgreSQL 16 + Redis 7)

### Blockers/Concerns

- v1.2 仅改 UI，不动后端 API
- v1.2 保持现有路由结构不变
- 考勤打卡设备对接/手机定位打卡 属于 V2.0 范围

## Session Continuity

Last session: 2026-04-14T00:00:00.000Z
Stopped at: v1.2 roadmap created
Next step: /gsd-plan-phase 1
