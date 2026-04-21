---
gsd_state_version: 1.0
milestone: v1.0
milestone_name: milestone
status: completed
stopped_at: Phase 14 context gathered (2026-04-21)
last_updated: "2026-04-21T00:00:00.000Z"
last_activity: 2026-04-21
progress:
  total_phases: 4
  completed_phases: 4
  total_plans: 10
  completed_plans: 10
  percent: 100
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-04-17)

**Core value:** 简单，好用，省时间 -- 老板3步完成核心人事操作，无需专业知识
**Current focus:** Phase 14 — 组织架构图（部门+岗位管理）— Context gathered

## Current Position

Phase: 14 (组织架构图) — **Context gathered (discuss-phase complete)**

Status: Phase 10 complete | Phase 11 complete | Phase 12 complete | Phase 13 complete | Phase 14 context done
Last activity: 2026-04-21

Progress: [▌▌▌▌▌▌▌▌▌▌] 100%

## v1.4 Phase Overview

| Phase | Goal | Requirements | Status |
|-------|------|--------------|--------|
| 10 | UX 基础 - 流程简化与引导体系 | UX-01~09 (9个) | **COMPLETE** |
| 11 | 合同合规 | COMP-01~04 (4个) | **COMPLETE** |
| 12 | 考勤合规报表 | COMP-05~08 (4个) | **COMPLETE** |
| 13 | 工资合规 | COMP-09~11 (3个) | **COMPLETE** |
| 14 | 组织架构图（部门+岗位管理） | ORG-01~04 (4个) | **Context ready** |

## v1.4 Phase 14 Summary

| Decision | Value |
|----------|-------|
| 岗位建模 | 新建 Position 表（独立管理） |
| 岗位归属 | 通用岗位（department_id=NULL，跨部门复用） |
| 架构图交互 | 支持拖拽调整部门层级 |
| 部门删除 | 引导转移员工后再删除 |
| 员工岗位 | el-select 下拉选择（部门联动过滤） |

## v1.4 Phase 13 Summary

| Plan | Type | Key Features |
|------|------|-------------|
| 13-01 | Backend | confirm API, confirmed_at, LEFT JOIN logs, asynq reminder worker, gocron scheduler |
| 13-02 | Frontend | H5 confirm button, send log confirmation column, salary.ts types |

## Performance Metrics

**Velocity:**

- Total plans completed: 38
- Average duration: ~15 min/plan
- Total execution time: ~9.2 hours

**By Phase:**
| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| v1.0 (9 phases) | 27 | ~7h | ~15min |
| v1.1 (3 phases) | 5 | ~1h | ~12min |
| v1.2 (4 phases) | 4 | ~1h | ~15min |
| v1.4 (Phase 10) | 3 | ~6min | ~2min |

**Recent Trend:**

- Last 5 plans: all PASS
- Trend: Stable

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
- [v1.4 Phase 10]: UX 增强统一前端基础设施，不涉及后端 API 变更（除必要的数据结构支持）
- [Phase 10 P01]: 员工向导创建模式使用StepWizard，编辑模式保持原有表单
- [Phase 10 P01]: 确认发送采用手动触发：Step2完成创建，员工手动点击发送短信
- [Phase 10 P03]: Tour高亮使用非scoped全局CSS `.tour-highlight` + `!important`，gridItems通过dataTour属性而非DOM操作
- [Phase 10 P03]: request.ts用ERROR_MESSAGES映射表+useMessage，可重试错误(500/502/503/timeout/network)传showActions:true
- [Phase 13]: 工资条确认使用 `confirmed` 状态替代 `signed`；gocron每日9:00 CST入队asynq任务；TodoCenter幂等通过 SourceType+SourceID 实现
- [Phase 14]: Position 表（通用岗位 department_id=NULL）；Employee.position_id FK；架构图支持拖拽；部门删除引导转移；员工岗位 el-select 下拉

### Roadmap Evolution

- v1.0: 后端核心 API + 原生 APP MVP
- v1.1: H5 管理后台基础框架
- v1.2: H5 管理后台 UI 重构
- v1.3: 产品功能全面优化（待办/考勤/薪资/社保/员工管理）
- v1.4: 用户体验优化（流程简化/引导/错误处理）+ 合规增强（合同/考勤报表/工资条回执）+ 组织架构图（部门+岗位管理）

### Blockers/Concerns

- [v1.3 research]: 薪资计算公式变更可能覆盖历史数据 -- 已确认月份强制只读保护（D-SAL-DATA-01 已实现）
- [v1.3 research]: 病假系数按城市配置 -- 初期只支持一线城市（北上广深）
- [v1.3 research]: 个税 Excel 模板需标准化 -- 非标准格式容错策略待 Phase 7 细化
- [v1.4]: 合同 PDF 生成依赖模板格式，需确认 PDF 库选型（go-pdf/fpdf 或 excelize 导出）
- [v1.4]: 员工签署短信验证码依赖阿里云 SMS，需确认模板 ID 配置

## Session Continuity

Last session: 2026-04-20T14:58:42.405Z
Stopped at: context exhaustion at 91% (2026-04-20)

## Deferred Items

Items acknowledged and deferred at milestone close on milestone completion:

| Category | Item | Status |
|----------|------|--------|
| debug | auth-me-org-id-mismatch | fixed |
| debug | dashboard-expiry-date-column | fixed |
| debug | dashboard-invalid-org-id-type | fixed |
| debug | dashboard-payment-month-missing | fixed |
| debug | login-success-no-redirect | fixed |
| debug | onboarding-redirect-fails | fixed |
| debug | org-onboarding-null-org-id | fixed |
| debug | orgs-current-404 | fixed |
| debug | register-calls-login-api | fixed |
| debug | register-error-message-concat | fixed |
