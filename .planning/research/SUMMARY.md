# Project Research Summary

**Project:** EasyHR (小微企业人事管理系统)
**Domain:** 中国小微企业HR+财务一体化SaaS
**Researched:** 2026-04-06
**Confidence:** HIGH

## Executive Summary

EasyHR 是一款面向中国小微企业（10-50人）的人事管理系统，核心定位是"轻量化HR全流程 + 基础财务记账 + 3步完成核心操作 + 完全免费"。市场空白明确：没有产品同时做到这四点。钉钉和飞书的HR功能分散且面向中大型企业，2号人事部功能较重，代理记账公司成本高且数据不在自己手中。EasyHR 通过极简交互和HR+财务一体化切入这个空白市场。

技术方案推荐 Go 模块化单体（Modular Monolith）架构，后端使用 Gin + GORM + PostgreSQL，前端管理后台使用 Vue 3 + Element Plus，员工端使用原生微信小程序，移动端分别使用 Kotlin (Android Jetpack Compose) 和 Swift (iOS SwiftUI)。多端通过统一的 REST API 通信。模块边界清晰定义（user/employee/social/payroll/tax/finance/notification），V1.0 使用进程内事件总线，为未来微服务拆分预留接口抽象。7个模块预计12周完成。

核心风险集中在三个领域：**多租户数据隔离泄露**（org_id遗漏导致跨企业数据泄露）、**个税累计预扣法计算错误**（中国个税算法复杂度高）、**社保政策变动导致数据错误**（30+城市政策库维护成本）。三者都有明确的预防策略，但必须在对应模块开发前建立测试框架和校验机制，而非事后修补。

## Key Findings

### Recommended Stack

后端采用 Go 1.23+ 单二进制架构，选择 Gin（HTTP）、GORM（ORM）、PostgreSQL（主库）、Redis（缓存/会话）为核心。前端 Vue 3 + Element Plus + Vite + TypeScript 构建管理后台。员工端使用原生微信小程序 + WeUI（不使用跨端框架，因为功能极简只有5-6个页面）。移动端 Kotlin/Jetpack Compose 和 Swift/SwiftUI 分别开发。部署在阿里云 ECS + Docker + Nginx 上。

**Core technologies:**
- **Go + Gin:** 后端语言和HTTP框架 -- 高性能单二进制，中国Go生态成熟，Gin中间件生态最丰富
- **GORM + PostgreSQL:** ORM和数据库 -- GORM的Scope机制天然适合多租户隔离，PostgreSQL的JSONB适合社保政策灵活存储
- **Vue 3 + Element Plus + Pinia:** 前端三件套 -- Vue 3 Composition API成熟，Element Plus是后台管理标配
- **asynq + gocron:** 异步任务和定时任务 -- 基于Redis，避免引入消息中间件，适合工资核算和社保提醒场景
- **casbin:** RBAC权限框架 -- 支持OWNER/ADMIN/MEMBER三级权限，与Gin集成成熟
- **excelize:** Excel读写 -- 工资条导出和考勤导入必备，Go生态最强Excel库

### Expected Features

V1.0 围绕三个闭环构建：人事基础闭环（员工入职到离职全生命周期）、月度运营闭环（社保+工资+个税每月必做操作）、财务闭环（凭证到报表的小微企业记账）。

**Must have (table stakes):**
- 员工入职/离职/档案管理 -- 所有HR系统的基础，必须3步内完成
- 劳动合同管理（PDF模板） -- 《劳动合同法》合规要求
- 社保参保基数自动匹配（30+城市政策库） -- 用户最大痛点，不懂社保政策
- 薪资结构自定义 + 一键工资核算 -- 替代Excel手动计算的核心痛点
- 个税自动计算（累计预扣法） -- 个税计算复杂，用户不懂专项附加扣除
- 电子工资单推送（微信小程序） -- 员工最关心的功能
- 手机号验证码登录 -- 降低注册门槛，中国用户习惯

**Should have (differentiators):**
- HR+财务一站式 -- 市面没有产品同时做好HR和记账，这是关键差异化
- 费用报销闭环（员工提交 -> 老板审批 -> 自动生成凭证） -- HR和财务的联动点
- 工资-个税-社保联动 -- 全流程数据打通，不需要重复录入
- 核心操作3步完成 -- 与钉钉/飞书的核心体验区别

**Defer (v2+):**
- 考勤管理 -- 硬件对接复杂，V2.0
- 电子签API自动签署 -- 成本问题，V2.0
- 发票OCR识别 -- 技术复杂度，V2.0
- 个税一键申报 -- 政务接口复杂，V2.0
- 招聘管理 -- 非核心，V3.0

### Architecture Approach

推荐模块化单体（Modular Monolith）架构，单进程、单数据库、按业务边界划分8个模块。模块间通过 Service 接口通信（禁止跨模块直接访问 Repository），V1.0 使用进程内同步事件总线。模块边界即未来微服务边界。

**Major components:**
1. **user** -- 认证、用户管理、组织管理、RBAC权限（所有模块依赖此模块）
2. **employee** -- 员工入职/离职/档案/合同管理（核心实体，social/payroll/tax依赖其数据）
3. **social** -- 社保参保/停缴/变更/政策匹配（维护30+城市政策库，是工资核算上游）
4. **payroll** -- 薪资结构/工资核算/发放/工资单推送（跨模块数据聚合点，最复杂）
5. **tax** -- 个税计算/专项附加扣除/申报提醒/增值税企业所得税（与payroll双向依赖）
6. **finance** -- 会计科目/凭证/账簿/发票/费用报销/会计期间/报表（最复杂模块，依赖前面所有模块）
7. **notification** -- 站内消息/短信/微信推送（横切模块，随业务模块逐步接入）
8. **common** -- 中间件/加密/统一响应/审计日志/工具函数（基础设施层）

**Key architecture patterns:**
- 模块接口契约（interface） -- 可测试、可替换、预留微服务拆分
- 加密字段双列模式（加密值 + SHA-256哈希索引） -- 满足个人信息保护法要求
- 租户隔离Scope（自动注入org_id） -- 防止多租户数据泄露
- 事务脚本（复杂写操作封装在Service层事务中） -- 单库事务保障一致性
- 进程内事件总线（V1.0同步调用，接口抽象好随时可替换） -- 避免过早引入消息队列

### Critical Pitfalls

1. **多租户数据隔离泄露** -- 在GORM层面实现全局Scope自动注入org_id，所有Repository方法自动带租户过滤，集成测试必须包含多租户隔离验证
2. **个税累计预扣法计算错误** -- 严格按照《个人所得税扣缴申报管理办法》实现，每月必须读取年度历史数据，编写至少10种边界场景的单元测试
3. **社保政策变动导致数据错误** -- 政策采用配置表（城市+年份+险种+比例），每条政策记录effective_date/expiry_date，变更时自动标记受影响员工
4. **工资核算跨模块依赖顺序错误** -- 建立明确的月度核算流程（社保核算 -> 个税计算 -> 工资核算），工资表设置前置条件校验
5. **财务凭证借贷不平衡** -- 后端API层强制校验借贷平衡，已审核凭证禁止修改（只能红冲），数据库层增加约束
6. **界面复杂度过高导致用户流失** -- 坚持"3步完成核心操作"原则，表单拆分为多步骤，专业术语附带通俗解释

## Implications for Roadmap

Based on research, suggested phase structure (8 phases over ~12 weeks):

### Phase 1: Foundation & Project Scaffold
**Rationale:** 所有后续模块的基石，必须最先搭建。中间件、加密、统一响应、数据库连接等横切关注点在此阶段完成。
**Delivers:** 可运行的空项目框架 + 健康检查API + CI/CD流水线
**Addresses:** 项目基础设施
**Avoids:** Pitfall #5 多租户隔离（在此阶段建立TenantScope全局机制）、#6 敏感数据加密（建立双列模式）

### Phase 2: User & Organization Module
**Rationale:** user模块是所有模块的依赖基础，认证和权限不完成则其他模块无法开发。
**Delivers:** 手机号验证码登录 + 企业信息管理 + OWNER/ADMIN/MEMBER三级RBAC + 审计日志
**Addresses:** 必备功能：手机号验证码登录、企业信息设置、多子账号权限
**Avoids:** Pitfall #5 多租户隔离（在此阶段验证隔离机制）

### Phase 3: Employee Module
**Rationale:** 员工管理是核心实体，social/payroll/tax/finance都依赖员工数据。员工入职到离职是产品第一触点。
**Delivers:** 员工入职（3步完成）/ 信息编辑 / 合同管理（PDF模板）/ 离职办理 / 搜索导出
**Addresses:** 必备功能：员工入职登记、档案管理、合同管理、离职办理
**Avoids:** Pitfall #9 界面复杂度（必须3步完成入职）、#10 业务规则（试用期与合同期限法定关系）
**Uses:** excelize（导出）、go-pdf/fpdf（合同PDF）、阿里云OSS（文件存储）

### Phase 4: Social Insurance Module
**Rationale:** 社保管理依赖员工数据，也是工资核算的上游（提供社保扣款数据）。社保基数自动匹配是用户最大痛点。
**Delivers:** 30+城市社保政策库 + 参保/停缴/变更 + 自动计算缴费金额 + 缴费提醒 + 参保材料PDF
**Addresses:** 必备功能：社保参保基数自动匹配、社保缴费提醒、社保变更联动
**Avoids:** Pitfall #1 社保政策变动（配置表+生效日期）、#10 业务规则（五险比例）
**Research flag:** 社保政策库数据收集（30+城市五险一金基数和比例）需要额外研究，建议在Phase规划时使用 /gsd:research-phase

### Phase 5: Tax Module (parallel with Phase 6 start)
**Rationale:** 个税模块与工资模块紧密耦合但核心计算逻辑独立，可与工资模块前期并行开发。
**Delivers:** 个税计算引擎（累计预扣法）+ 专项附加扣除管理 + 申报提醒 + 申报表PDF + 税务计算
**Addresses:** 必备功能：个税自动计算、个税申报提醒、个税申报表生成
**Avoids:** Pitfall #2 个税累计预扣法（严格按法规实现）、#10 业务规则（起征点、专项扣除）
**Research flag:** 2026年最新个税税率和专项附加扣除政策需要验证，建议 /gsd:research-phase

### Phase 6: Payroll Module
**Rationale:** 工资核算需要聚合员工、社保、个税三个模块数据，是最复杂的跨模块聚合点，必须在前三个业务模块完成后开发。
**Delivers:** 薪资结构自定义 + 月度工资表创建 + 一键核算引擎 + 工资单推送 + 工资发放记录 + Excel导入导出
**Addresses:** 必备功能：薪资结构自定义、一键工资核算、电子工资单推送、工资发放记录、导出工资条Excel
**Avoids:** Pitfall #3 跨模块依赖顺序（建立核算状态机）、#9 界面复杂度（一键核算流程引导）
**Uses:** asynq（异步工资核算）、excelize（Excel导入导出）、notification（工资单推送）

### Phase 7: Finance Module
**Rationale:** 财务模块是最复杂的模块，依赖前面所有模块的数据（工资凭证、社保凭证、报销人信息）。是产品关键差异化。
**Delivers:** 会计科目 + 凭证管理 + 费用报销 + 发票登记 + 账簿查询 + 会计期间/结账 + 财务报表 + 税务计算
**Addresses:** 差异化功能：财务记账集成、费用报销闭环、工资-个税-社保联动
**Avoids:** Pitfall #4 借贷平衡（后端强制校验）、#8 结账并发冲突（Redis分布式锁）、#9 界面复杂度（术语通俗化）
**Research flag:** 小微企业会计科目预置模板和中国小企业会计准则需要额外研究，建议 /gsd:research-phase

### Phase 8: WeChat Mini Program (Employee Portal)
**Rationale:** 员工端功能简单（查看工资条/合同/社保记录/提交报销），放在最后开发。需要微信小程序审核资质。
**Delivers:** 员工微信小程序（5-6个页面）+ CI/CD自动上传
**Addresses:** 必备功能：员工查看工资条/合同/社保记录、费用报销提交
**Avoids:** Pitfall #7 小程序审核被拒（提前准备资质、完善隐私政策）
**Research flag:** 微信小程序2026年审核政策和人力资源类目资质要求需要验证，建议 /gsd:research-phase

### Phase Ordering Rationale

- **依赖关系驱动顺序：** user -> employee -> social/tax -> payroll -> finance，每个模块依赖前序模块的数据，不可跳序
- **复杂度递增：** 从简单CRUD（employee）到复杂聚合（payroll）到专业领域（finance），团队逐步积累经验
- **跨模块接口先行：** Phase 4/5/6之间存在循环依赖（payroll <-> tax），通过接口注入解决，必须在架构阶段约定好接口契约
- **Notification贯穿开发：** 不是独立Phase，而是随各业务模块逐步接入通知类型
- **并行机会：** Tax模块可与Payroll模块前期并行开发，节省1-2周

### Research Flags

**Phases likely needing deeper research during planning:**
- **Phase 4 (Social Insurance):** 30+城市五险一金基数和比例数据收集，这是纯数据工作而非技术工作，数据质量直接决定产品可信度
- **Phase 5 (Tax):** 2026年最新个税税率表、专项附加扣除政策、累计预扣法边界场景确认
- **Phase 7 (Finance):** 小微企业会计科目预置模板、中国小企业会计准则报表格式、结账流程最佳实践
- **Phase 8 (WeChat Mini Program):** 微信小程序审核政策、人力资源类目资质要求、隐私政策模板

**Phases with standard patterns (skip research-phase):**
- **Phase 1 (Foundation):** Go项目脚手架、中间件、GORM配置等都是成熟模式
- **Phase 2 (User/Org):** JWT认证、RBAC权限、短信验证码是标准实现
- **Phase 3 (Employee):** 标准CRUD + 加密字段 + PDF生成，模式成熟
- **Phase 6 (Payroll):** 跨模块数据聚合模式在架构文档中已明确，无需额外研究

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | 所有推荐技术版本通过GitHub Release API和npm Registry验证（2026-04-06），版本号精确到具体release |
| Features | MEDIUM | 竞品分析基于训练数据而非实时爬取，WebSearch/WebReader不可用。竞品功能列表和定价需要实施时验证 |
| Architecture | HIGH | 架构方案基于项目已有的tech-architecture.md（1374行）和PROJECT.md，模块边界和数据流定义清晰 |
| Pitfalls | HIGH | 基于PRD分析和领域经验，6个CRITICAL/HIGH级别陷阱都有明确的预防策略和代码示例 |

**Overall confidence:** HIGH

### Gaps to Address

- **竞品实时数据:** WebSearch/WebReader不可用导致竞品分析基于训练数据。实施前应访问钉钉开放平台、2号人事部、蚂蚁HR确认其HR模块具体功能列表和定价
- **社保政策库数据源:** 30+城市五险一金基数和比例需要权威数据源，建议研究各城市人社局官网或第三方社保数据服务API
- **2026年政策变动:** 个税税率、专项附加扣除标准、社保缴费基数可能在2026年有调整，开发前需验证最新政策
- **微信小程序审核:** 人力资源类小程序审核要求可能变化，建议在Phase 8开始前研究最新审核指南
- **小微企业会计准则:** 预置科目体系和报表格式需要参考最新版《小企业会计准则》，确认是否有2026年更新
- **swag v2.0稳定性:** API文档工具swag v2.0仍在RC阶段，生产环境需要关注稳定性或退回v1.x

## Sources

### Primary (HIGH confidence)
- GitHub Release API -- Go后端所有依赖库版本验证（gin v1.12.0, gorm v1.31.1, asynq v0.26.0等）
- npm Registry -- 前端所有依赖库版本验证（vue 3.5.32, element-plus 2.13.6, vite 8.0.3等）
- tech-architecture.md -- 项目完整技术架构文档（1374行），模块边界和数据流定义
- PRD文档（prd.md） -- 产品需求详细定义，功能优先级和依赖关系
- PROJECT.md -- 项目上下文和约束

### Secondary (MEDIUM confidence)
- Go社区模块化单体最佳实践（threedots.tech, gobeyond.dev） -- 架构模式参考
- GORM v2官方文档 -- ORM使用模式和Scope机制
- Gin框架官方文档 -- 中间件生态和路由设计
- 训练数据中的竞品知识（钉钉、飞书、2号人事部、蚂蚁HR） -- 竞品功能对比

### Tertiary (LOW confidence)
- 微信小程序审核政策 -- 需要Phase 8开始前实时验证
- 2026年社保/个税最新政策 -- 需要开发前实时查询各城市人社局和税务局官网
- 小微企业会计准则最新版 -- 需要确认是否有2026年更新

---
*Research completed: 2026-04-06*
*Ready for roadmap: yes*
