# 易人事（EasyHR）

## What This Is

为小微企业/个体户老板（10-50人规模）打造的轻量化、一站式人事管理APP。解决无专职HR的小老板在员工入职、社保、工资、个税、财务记账等基础人事事务中的操作痛点。核心体验：3步内完成核心操作，零学习成本。

产品包含：老板端原生APP（Android/iOS）、H5管理后台（Vue 3）、员工端微信小程序。

## Core Value

**简单、好用、省时间** — 老板打开APP第一时间知道要做什么，3步完成核心人事操作，无需专业知识。

## Current State

**v1.3 shipped (2026-04-19)** — H5 管理后台功能全面优化

已交付：待办中心、考勤管理（打卡/审批流/出勤月报）、薪资增强（调薪/个税/绩效/工资条）、社保增强（增减员/渠道/状态）、员工管理增强（看板/架构/登记/离职）

技术栈：Go 1.25 + Gin v1.12 + GORM + PostgreSQL + Vue 3 + Element Plus + ECharts + asynq + gocron

详见：[.planning/milestones/v1.3-ROADMAP.md](.planning/milestones/v1.3-ROADMAP.md)

## Current Milestone: v1.4 用户体验 + 合规增强

**Goal:** 优化操作体验（精简步骤/完善提示/帮助引导）并增强合规能力（合同电子签/考勤报表/工资条签收）

**Target features:**
- 操作步骤精简：核心功能 ≤ 3 步完成
- 错误提示完善：友好错误消息 + 解决方案引导
- 帮助引导：首次使用引导、空状态提示、操作提示
- 劳动合同电子签：电子合同生成 + 签署流程
- 考勤合规报表：加班/请假/出勤统计合规报表
- 工资条签收确认：员工确认收到工资条

## Validated Requirements

### v1.0 MVP (shipped 2026-04-11)

- [x] 用户注册/登录（H5+APP）
- [x] 员工管理 CRUD
- [x] 社保管理（城市/基数/参保城市/增减员）
- [x] 个税计算（税率表配置/算税引擎）
- [x] 工资核算（应发/实发/工资条/导出）
- [x] 财务记账（收支/余额/对账）
- [x] 首页工作台（统计卡片/快捷入口）
- [x] 员工微信小程序（工资条/合同/社保）

### v1.1 (shipped 2026-04-13)

- [x] **UI-01**: 登录页左右分栏布局 (#1A2D6B → #4F6EF7 → #7B9FFF 渐变)
- [x] **UI-13**: AppLayout 侧边栏（折叠/展开/Logo/菜单）

### v1.2 (shipped 2026-04-14)

- [x] 员工管理/薪资/社保/考勤/审批 H5 页面按原型图重构
- [x] 工具首页、个人中心等补充页面

### v1.3 (shipped 2026-04-19)

- [x] **COMP-05**: 加班统计报表（法定节假日/工作日延时/周末加班分3档，0.5h取整）
- [x] **COMP-06**: 请假合规报表（年假额度/已用/剩余，病假，事假）
- [x] **COMP-07**: 出勤异常报表（迟到/早退/缺勤，异常行红色高亮 late>3 or absent>1）
- [x] **COMP-08**: 月度考勤汇总 Excel 导出（Blob download，12列统计表）

## Out of Scope

| 功能 | 原因 |
|------|------|
| 手机定位打卡/GPS打卡 | 需要 APP 端支持，v1.3 仅做 H5 |
| 电子签API自动签署 | V2.0 范围，v1.0 降级为 PDF 模板+手动签署 |
| 社保对接第三方数据服务商 | V2.0 范围 |
| 社保在线办理对接政务接口 | V3.0 范围 |
| 发票OCR识别 | V2.0 范围 |
| 个税一键申报对接 | V2.0 范围 |
| 招聘模块 | V3.0 范围 |
| 付费增值服务 | V3.0 范围 |
| 深色模式 | 后续迭代 |
| 微信小程序员工端审批 | 员工端后续迭代 |
| 多级审批流 | 小微企业只需老板审批，单级足够 |
| HMAC webhook 签名验证 | Phase 08 已知技术债，待 V2.0 前端安全加固 |
| SMS 转发（阿里云模板配置） | Phase 05 已知技术债，待短信服务配置 |

## Context

- **市场背景**：钉钉/飞书以工作协同为核心，人事功能非专门化、操作复杂，小微企业老板难以使用
- **目标用户特征**：无专业HR知识、时间精力有限、对成本敏感、追求高效便捷
- **次要用户**：兼职人事/行政，需快速完成基础操作
- **商业模式**：V1.0-V2.0 核心功能完全免费，通过免费获客沉淀数据；V3.0 起通过增值服务和付费模块变现
- **技术架构**：模块化单体（Go + PostgreSQL），逻辑多租户（org_id），前端原生APP + 微信小程序 + H5管理后台
- **数据安全**：敏感字段AES-256-GCM加密 + SHA-256哈希索引，符合《个人信息保护法》《劳动合同法》
- **代码规模**：~50+ Go packages, ~30+ Vue 组件，v1.3 新增 5 个模块（attendance/department/todo/salary/socialinsurance dashboard）

## Constraints

- **操作体验**: 核心功能操作步骤 ≤ 3步 — 产品核心差异化，必须坚持
- **性能**: API 响应 ≤ 500ms（95%请求），APP端操作响应 ≤ 2s — 小微企业老板不会等待
- **并发**: 支持 ≥ 1000 同时在线用户，单企业 ≤ 50人同时操作无卡顿
- **可用性**: 全年 ≥ 99.9%，7×24小时稳定运行
- **兼容性**: Android 8.0+ / iOS 12.0+
- **合规**: 符合《个人信息保护法》《劳动合同法》等法律法规
- **成本**: V1.0 核心功能完全免费，技术成本需可控
- **部署**: Docker + 阿里云 ECS，单二进制部署

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| 模块化单体架构（非微服务） | V1.0用户量小，降低运维复杂度；按业务边界划分模块，后续可拆 | ✅ V1.3 验证，attendance/department/todo 等独立模块工作良好 |
| 逻辑多租户（org_id） | 单DB共享降低成本，org_id隔离满足数据安全要求 | ✅ 继续 |
| 原生APP（Kotlin + Swift） | 性能和用户体验优先，目标用户对流畅度敏感 | ⏳ v1.3 聚焦 H5，APP 端待 V2.0 |
| 所有第三方依赖设计降级方案 | 核心业务不依赖外部接口可用性，降低风险 | ✅ asynq/gocron 定时任务均有 fallback；Redis 不可用时解锁不阻塞 |
| V1.0 自建社保政策库 | 30+城市数据，管理员手动更新，避免V1.0对接成本 | ✅ 继续 |
| Go + PostgreSQL 后端 | 高性能、编译为单二进制、ACID事务保障、JSONB灵活存储 | ✅ v1.3 继续使用，GORM AutoMigrate 够用 |
| RBAC三级权限 | OWNER/ADMIN/MEMBER，简单明确，满足小微企业需求 | ✅ 继续 |
| H5 UI 重构遵循 EasyHR-web.pen 原型 | 统一视觉风格，主色调 #4F6EF7，卡片圆角 12px | ✅ v1.2+v1.3 遵循 |
| 审批流使用 qmuntal/stateless 状态机 | 11种审批类型需要状态机管理流转 | ✅ Phase 06 |
| 组织架构复用 ECharts tree 图表 | 部门→岗位→员工三层结构树 | ✅ Phase 05 |
| 调薪 INSERT ONLY，禁止 UPDATE 历史 | 薪资历史数据不可篡改 | ✅ Phase 07 |
| 考勤班次模型必须包含 workDateOffset | 跨天班次（如22:00-06:00）需要日期偏移 | ✅ Phase 06 |
| SIMonthlyPayment 月度缴费表独立建模 | 月度状态流转（正常→待缴→欠缴→已转出）独立于参保生命周期 | ✅ Phase 08 |
| TodoItem 扩展字段而非新建表 | 限时任务字段（deadline/is_time_limited/urgency_status）扩展到现有 TodoItem | ✅ Phase 09 |
| 协办邀请复用 Token 机制 | 纯填写无需登录，和 Registration 模块共享 generateToken 模式 | ✅ Phase 09 |
| 加班分类: ClassifyOvertimeCategory 复用 AttendanceRule.Holidays | 法定节假日加班按 IsHoliday 判断，周末按 Weekday 判断 | ✅ Phase 12 |
| 异常阈值: late>3 OR absent>1 → 红色高亮 | 异常员工数和异常时长双维度统计 | ✅ Phase 12 |

## Evolution

This document evolves at phase transitions and milestone boundaries.

---
*Last updated: 2026-04-20 — v1.4 Phase 12 考勤合规报表 shipped*
