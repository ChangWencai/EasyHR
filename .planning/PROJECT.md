# 易人事（EasyHR）

## What This Is

为小微企业/个体户老板（10-50人规模）打造的轻量化、一站式人事管理APP。解决无专职HR的小老板在员工入职、社保、工资、个税、财务记账等基础人事事务中的操作痛点。核心体验：3步内完成核心操作，零学习成本。

产品包含：老板端原生APP（Android/iOS）、H5管理后台（Vue 3）、员工端微信小程序。

## Core Value

**简单、好用、省时间** — 老板打开APP第一时间知道要做什么，3步完成核心人事操作，无需专业知识。

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] 用户通过手机号验证码一键登录/注册，首次登录自动引导录入企业信息
- [ ] 老板可创建入职邀请，员工在线填写信息并签署合同（V1.0 PDF模板+手动签署）
- [ ] 老板可办理员工离职，自动生成交接清单，同步触发社保停缴提醒
- [ ] 集中管理员工档案，支持按姓名/岗位搜索、导出Excel
- [ ] 根据员工城市/岗位自动匹配社保基数（自建30+城市政策库），一键生成参保材料
- [ ] 社保缴费到期前3天自动提醒，记录缴费明细，支持打印凭证
- [ ] 员工岗位/薪资变动自动触发社保基数调整提醒
- [ ] 支持自定义薪资结构，一键核算工资，支持导入考勤表、复制上月工资表
- [ ] 自动生成电子工资单并推送至员工，支持导出工资条Excel
- [ ] 记录每月工资发放状态/金额/方式，异常发放自动提醒
- [ ] 基于工资数据自动匹配个税专项附加扣除，精准计算个税
- [ ] 个税申报截止前3天自动提醒，生成申报表供手动提交
- [ ] 记录个税申报明细，支持查询状态和导出凭证
- [ ] 手动录入会计凭证，支持借贷平衡实时校验、草稿保存、提交审核
- [ ] 手动登记进项/销项发票，月末自动汇总计算增值税
- [ ] 员工通过微信小程序提交费用报销，老板在线审批，审批通过自动生成费用凭证
- [ ] 基于凭证数据实时生成总账、明细账、科目余额表，支持导出Excel
- [ ] 月末结账后自动生成资产负债表、利润表，支持多期对比
- [ ] 按月度管理会计期间，结账锁定当期凭证，支持反结账（OWNER权限+二次确认）
- [ ] 预置小微企业常用会计科目，支持自定义增删，五大类分层展示
- [ ] 基于发票和凭证数据自动计算增值税和企业所得税，生成纳税申报辅助数据
- [ ] 企业基础信息设置，同步至各功能模块
- [ ] 多子账号权限管理（OWNER/ADMIN/MEMBER三级RBAC）
- [ ] 操作日志全程记录，可追溯
- [ ] 员工端微信小程序：查看工资条/合同/社保记录、提交费用报销

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
*Last updated: 2026-04-06 after initialization*
