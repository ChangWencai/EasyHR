# 易人事（EasyHR）

## What This Is

为小微企业/个体户老板（10-50人规模）打造的轻量化、一站式人事管理APP。解决无专职HR的小老板在员工入职、社保、工资、个税、财务记账等基础人事事务中的操作痛点。核心体验：3步内完成核心操作，零学习成本。

产品包含：老板端原生APP（Android/iOS）、H5管理后台（Vue 3）、员工端微信小程序。

## Core Value

**简单、好用、省时间** — 老板打开APP第一时间知道要做什么，3步完成核心人事操作，无需专业知识。

## Current Milestone: v1.3 产品功能全面优化（基于 PRD 1.1）

**Goal:** 根据 PRD 1.1 对现有产品进行功能优化和补全，新增待办中心、考勤管理、完善薪资/社保/员工管理模块，仅 H5 管理后台。

**Target features:**
- 待办中心：事项汇总、快捷办事、限时任务、完成率环形图
- 考勤管理：打卡设置（3种模式）、今日打卡、审批流（7种假类型）、出勤月报
- 薪资管理：数据看板、调薪/普调、个税上传、绩效系数、发工资条
- 社保公积金：数据看板、增减员优化、缴费渠道、欠缴状态管理
- 员工管理：数据看板、组织架构可视化、员工信息登记、办离职优化、花名册增强

**Scope:** 仅 H5 管理后台（Vue 3），后端 API 配合新增

## Active Requirements

(None yet — will be defined in step 9)

## Validated Requirements

- [x] **UI-01**: 登录页重构为左右分栏布局，蓝色渐变背景+品牌区+表单区 (Phase 01)
- [x] **UI-13**: AppLayout 侧边栏按原型风格优化（折叠/展开/Logo/菜单） (Phase 01)

### In This Milestone (v1.2)

- [ ] **UI-02**: 首页重构为仪表盘布局：4统计卡片+图表+待办事项+数据表格
- [ ] **UI-03**: 员工管理列表重构为筛选栏+表格卡片风格
- [ ] **UI-04**: 员工详情页重构为左侧信息卡+右侧详情面板
- [ ] **UI-05**: 新增员工表单重构为3步骤条引导
- [ ] **UI-06**: 薪资管理页重构：汇总行+工资表卡片
- [ ] **UI-07**: 薪资配置页重构为双栏配置表单
- [ ] **UI-08**: 薪资明细页重构：月份导航+员工明细+实发卡片
- [ ] **UI-09**: 社保管理页重构：警告横幅+多行布局
- [ ] **UI-10**: 考勤管理页重构：统计行+日历表格
- [ ] **UI-11**: 审批管理页重构：Tab切换+审批列表
- [ ] **UI-12**: 审批详情页重构：状态徽章+详情卡
- [ ] **UI-13**: AppLayout 侧边栏按原型风格优化（折叠/展开/Logo/菜单）
- [ ] **UI-14**: 工具首页、个人中心等未覆盖页面按原型风格补充设计

---

*Last updated: 2026-04-14 — Phase 01 complete, UI-01/UI-13 validated*

### Out of Scope

- 考勤管理（打卡设备对接/手机定位打卡） — V2.0
- 电子签API自动签署 — V2.0（V1.0降级为PDF模板+手动签署）
- 社保对接第三方数据服务商 — V2.0（V1.0自建政策库）
- 社保在线办理对接政务接口 — V3.0
- 发票OCR识别 + 税务局查验接口 — V2.0（V1.0手动录入）
- 个税一键申报对接 — V2.0（V1.0手动提交至自然人电子税务局）
- 招聘模块 — V3.0
- 付费增值服务（人事咨询、社保代办、合规服务） — V3.0
- 深色模式 — 后续迭代
- 人事报表/数据分析 — V2.0

## Context

- **市场背景**：钉钉/飞书以工作协同为核心，人事功能非专门化、操作复杂，小微企业老板难以使用
- **目标用户特征**：无专业HR知识、时间精力有限、对成本敏感、追求高效便捷
- **次要用户**：兼职人事/行政，需快速完成基础操作
- **商业模式**：V1.0-V2.0核心功能完全免费，通过免费获客沉淀数据；V3.0起通过增值服务和付费模块变现
- **技术架构**：模块化单体（Go + PostgreSQL），逻辑多租户（org_id），前端原生APP + 微信小程序 + H5管理后台
- **数据安全**：敏感字段AES-256-GCM加密 + SHA-256哈希索引，符合《个人信息保护法》《劳动合同法》

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
| 模块化单体架构（非微服务） | V1.0用户量小，降低运维复杂度；按业务边界划分模块，后续可拆 | — Pending |
| 逻辑多租户（org_id） | 单DB共享降低成本，org_id隔离满足数据安全要求 | — Pending |
| 原生APP（Kotlin + Swift） | 性能和用户体验优先，目标用户对流畅度敏感 | — Pending |
| 所有第三方依赖设计降级方案 | 核心业务不依赖外部接口可用性，降低风险 | — Pending |
| V1.0 自建社保政策库 | 30+城市数据，管理员手动更新，避免V1.0对接成本 | — Pending |
| Go + PostgreSQL 后端 | 高性能、编译为单二进制、ACID事务保障、JSONB灵活存储 | — Pending |
| RBAC三级权限 | OWNER/ADMIN/MEMBER，简单明确，满足小微企业需求 | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd:transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd:complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-04-17 — v1.3 milestone started*
