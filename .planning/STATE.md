---
gsd_state_version: 1.0
milestone: v1.3
milestone_name: 产品功能全面优化（基于 PRD 1.1）
status: shipped
shipped_at: 2026-04-19
last_updated: "2026-04-19T12:51:00Z"
last_activity: 2026-04-19

## Deferred Items

Items acknowledged and deferred at milestone close on 2026-04-19:

| Category | Item | Status |
|----------|------|--------|
| debug | auth-me-org-id-mismatch | verifying |
| debug | dashboard-expiry-date-column | verifying |
| debug | dashboard-invalid-org-id-type | verifying |
| debug | dashboard-payment-month-missing | awaiting_human_verify |
| debug | login-success-no-redirect | verifying |
| debug | orgs-current-404 | verifying |
| debug | register-calls-login-api | verifying |
| debug | register-error-message-concat | awaiting_human_verify |
| verification | Phase 05: SMS forwarding placeholder | gaps_found |
| verification | Phase 08: SIDetailDialog API path mismatch | gaps_found |
progress:
  total_phases: 5
  completed_phases: 5
  total_plans: 20
  completed_plans: 20
  percent: 100
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-04-17)

**Core value:** 简单、好用、省时间 -- 老板3步完成核心人事操作，无需专业知识
**Current focus:** Phase 09 — 待办中心

## Current Position

Phase: 09 (待办中心) -- Complete (all 3 plans done)
Plan: 3 of 3 planned
Status: Complete
Last activity: 2026-04-19

Progress: [██████████] 100%

## Phase 09 Key Decisions

- **环形图**: ECharts + HomeView 顶部
- **限时任务**: 扩展 TodoItem 字段（deadline/is_time_limited/urgency_status），1-7天超期=超时，15天+=失效
- **轮播图**: CarouselItem 表，管理员配置 1-3 张
- **快捷入口**: 保留现有 6 个 + 追加 3 个新入口
- **协办邀请**: 复用 Token 机制，纯填写，无登录
- **终止任务**: 保留数据，状态改为"已终止"

## Performance Metrics

**Velocity:**

- Total plans completed: 32
- Average duration: ~15 min/plan
- Total execution time: ~8.5 hours

**By Phase:**
| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| v1.0 (9 phases) | 27 | ~7h | ~15min |
| v1.1 (3 phases) | 5 | ~1h | ~12min |
| v1.2 (4 phases) | 4 | ~1h | ~15min |

**Recent Trend:**

- Last 5 plans: all PASS
- Trend: Stable

| Phase 07 P01 | 13 | 5 tasks | 19 files |
| Phase 07 P02 | 5 | 2 tasks | 4 files |
| Phase 07 P03 | 22 | 4 tasks | 15 files |
| Phase 07 P04 | 18 | 4 tasks | 12 files |
| Phase 05 P04 | 8 | 2 tasks | 8 files |
| Phase 08 P01 | 10 | 3 tasks | 9 files |
| Phase 08 P02 | 3 | 3 tasks | 3 files |
| Phase 08 P03 | 3 | 3 tasks | 3 files |
| Phase 08 P04 | 3 | 2 tasks | 2 files |
| Phase 09 P01 | 8 | 3 tasks | 10 files |
| Phase 09 P02 | 5 | 2 tasks | 9 files |
| Phase 09 P03 | 16 | 2 tasks | 16 files |

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- Go 1.25.0 (gin v1.12.0 要求)
- GORM AutoMigrate（V1.0 够用，不用 golang-migrate）
- 软删除使用部分唯一索引 `WHERE deleted_at IS NULL`
- 多设备并发登录，不限制设备数
- Refresh Token 轮转策略
- 逻辑多租户（org_id），GORM Scope 自动注入
- 敏感字段双列模式（AES-256-GCM + SHA-256）
- 三级 RBAC：OWNER/ADMIN/MEMBER
- 审计日志 GORM Hook 自动记录，INSERT ONLY
- [v1.2]: H5 UI 重构遵循 web-design/EasyHR-web.pen 原型图设计
- [v1.2]: 主色调商务蓝 #4F6EF7，卡片圆角 12px，侧边栏固定 220px
- [v1.3]: 仅 H5 管理后台实现新功能，后端 API 配合新增
- [v1.3 research]: 审批流使用 qmuntal/stateless 状态机
- [v1.3 research]: 个税计算自建引擎，税率表用 JSON 配置
- [v1.3 research]: 组织架构复用 ECharts tree 图表
- [v1.3 research]: 调薪 INSERT ONLY，禁止 UPDATE 历史
- [v1.3 research]: 考勤班次模型必须包含 workDateOffset
- [Phase 05]: Registration 模块复用 Invitation 的 generateToken 模式（crypto/rand 32-byte hex），SubmitRegistration 事务性 upsert 员工记录
- [Phase 05 P04]: 前端状态值从 pending_review 修正为 pending，与后端模型保持一致
- [Phase 05 P04]: CompleteOffboardingFromSI 接收 employeeID 而非 offboardingID，因社保模块只持有员工 ID
- [Phase 05 P04]: 前端 API 调用从 POST 改为 PUT，与后端 approve/reject/complete 路由对齐
- [Phase 05 P05]: 花名册使用独立 ListRoster API 保持接口兼容，批量关联查询 HasTable 降级
- [Phase 07 P01]: DashboardHandler 方法合并到主 salary Handler 简化路由注册
- [Phase 07 P01]: SalaryAdjustment 使用 plain struct（无 BaseModel）因 INSERT ONLY 无需 soft-delete
- [Phase 07 P01]: SickLeavePolicy + SalarySlipSendLog 共享模型文件
- [Phase 07 P02]: 加班分档从 Approval.StartTime + RuleEngine 推导（无 overtime_type 字段）
- [Phase 07 P02]: 病假扣款 = 日工资 * 病假天数 * (1 - 系数)，差额模式
- [Phase 07 P04]: 部门多选用 __all__ sentinel value 模拟全选，toggleSelectAllDepts 处理
- [Phase 07 P04]: 解锁降级：Redis 不可用时打印 fallback code 到日志，不阻塞解锁流程
- [Phase 07 P04]: SalaryListHandler 和 PayrollHandler 路由分开（/salary/list vs /salary/payroll）避免重复注册
- [Phase 08]: SIMonthlyPayment 月度缴费表（employee_id + year_month + status + payment_channel），asynq 定时任务流转状态，Organization.payment_channel 作为默认值
- [Phase 08 P04]: Excel 导出 handler 直接调用 repo.ListRecords，export=full 控制含明细导出，写入 gin.Context.Data 避免双重 buffer
- [Phase 09]: TodoItem 扩展 deadline/is_time_limited/urgency_status 字段，不新建表；环形图 ECharts + HomeView 顶部；CarouselItem 表存轮播图配置；协办复用 Token 机制，纯填写无需登录；终止保留数据+标记状态

### Roadmap Evolution

- v1.0: 后端核心 API + 原生 APP MVP
- v1.1: H5 管理后台基础框架
- v1.2: H5 管理后台 UI 重构
- v1.3: 产品功能全面优化（待办/考勤/薪资/社保/员工管理）

### Blockers/Concerns

- [v1.3 research]: 薪资计算公式变更可能覆盖历史数据 -- 已确认月份强制只读保护（D-SAL-DATA-01 已实现）
- [v1.3 research]: 病假系数按城市配置 -- 初期只支持一线城市（北上广深）
- [v1.3 research]: 个税 Excel 模板需标准化 -- 非标准格式容错策略待 Phase 7 细化

## Session Continuity

Last session: 2026-04-19T04:38:50.038Z
Stopped at: context exhaustion at 90% (2026-04-19)
Resume file: None
