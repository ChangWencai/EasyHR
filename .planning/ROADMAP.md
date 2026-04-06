# Roadmap: 易人事（EasyHR）

## Overview

从项目基础设施出发，依次交付用户认证、员工管理、社保管理、个税计算、工资核算、财务记账、首页工作台和员工微信小程序，构建小微企业HR+财务一体化闭环。依赖链驱动阶段顺序：用户/组织 -> 员工 -> 社保 -> 个税+工资 -> 财务 -> 首页 -> 小程序。

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: 基础框架与用户认证** - 项目脚手架、多租户隔离、认证、企业信息、RBAC权限、审计日志
- [ ] **Phase 2: 员工管理** - 员工入职/离职/档案/合同全生命周期管理
- [ ] **Phase 3: 社保管理** - 30+城市社保政策库、参保/停缴/变更、缴费提醒
- [ ] **Phase 4: 个税计算** - 累计预扣法个税引擎、专项附加扣除、申报提醒
- [ ] **Phase 5: 工资核算** - 薪资结构自定义、一键核算、工资单推送、发放记录
- [ ] **Phase 6: 财务记账** - 会计科目、凭证、发票、费用报销、账簿、报表、结账
- [ ] **Phase 7: 首页工作台** - 待办事项、功能入口、数据概览、导航
- [ ] **Phase 8: 员工微信小程序** - 工资条/合同/社保查看、费用报销提交

## Phase Details

### Phase 1: 基础框架与用户认证
**Goal**: 用户可通过手机号登录系统，企业管理员可管理子账号和权限，多租户数据完全隔离
**Depends on**: Nothing (first phase)
**Requirements**: AUTH-01, AUTH-02, AUTH-03, AUTH-04, PLAT-01, PLAT-02, PLAT-03, PLAT-04, PLAT-05, PLAT-06, PLAT-07
**Success Criteria** (what must be TRUE):
  1. 老板通过手机号+验证码一键登录/注册，无需密码
  2. 首次登录自动进入企业信息录入引导，完成后可正常使用系统
  3. 企业管理员可添加子账号并分配OWNER/ADMIN/MEMBER权限，不同权限看到的功能和数据范围不同
  4. 所有写操作自动记录审计日志（谁、什么时间、做了什么），可追溯
  5. 两个不同企业的用户无法看到对方的数据（多租户隔离验证通过）
**Plans**: 4 plans

**Phase 1**: ✅ COMPLETE — 4/4 plans executed, 40 Go files, 7 test packages PASS
**Plans**:
- [x] 01-01-PLAN.md — 项目脚手架+基础设施(统一响应/加密/多租户Scope/中间件)
- [x] 01-02-PLAN.md — JWT认证工具+短信客户端+OSS客户端(pkg包)
- [x] 01-03-PLAN.md — 用户认证流程+RBAC权限+审计日志(核心业务逻辑)
- [x] 01-04-PLAN.md — Token刷新/退出+集成测试(认证/多租户隔离/审计)

### Phase 2: 员工管理
**Goal**: 老板可在3步内完成员工入职，集中管理员工档案，办理离职并自动触发后续流程
**Depends on**: Phase 1
**Requirements**: EMPL-01, EMPL-02, EMPL-03, EMPL-04, EMPL-05, EMPL-06, EMPL-07, EMPL-08
**Success Criteria** (what must be TRUE):
  1. 老板一键创建入职邀请，员工通过链接在线填写信息完成入职
  2. 老板可手动录入员工信息，档案支持按姓名/岗位快速检索并导出Excel
  3. 离职办理后自动生成交接清单，员工状态更新为"离职"，同步触发社保停缴提醒
  4. 合同可生成PDF模板、手动签署上传，并关联至员工档案
**Plans**: 4 plans

Plans:
- [x] 02-01-PLAN.md — 员工模型+档案管理+搜索+Excel导出 (EMPL-02, EMPL-03, EMPL-04)
- [x] 02-02-PLAN.md — 入职邀请流程 (EMPL-01)
- [x] 02-03-PLAN.md — 离职管理 (EMPL-05, EMPL-06, EMPL-07)
- [x] 02-04-PLAN.md — 合同管理+PDF生成 (EMPL-08)

### Phase 3: 社保管理
**Goal**: 老板可根据员工城市和岗位自动匹配社保基数，一键办理参保/停缴，缴费到期自动提醒
**Depends on**: Phase 2
**Requirements**: SOCL-01, SOCL-02, SOCL-03, SOCL-04, SOCL-05, SOCL-06, SOCL-07
**Success Criteria** (what must be TRUE):
  1. 根据员工城市+岗位自动匹配社保参保基数，30+城市政策库可用
  2. 老板勾选员工并确认后一键生成参保材料PDF
  3. 社保缴费到期前3天自动提醒老板，缴费明细可查询、可导出凭证
  4. 员工岗位或薪资变动时自动触发社保基数调整提醒
**Plans**: TBD

### Phase 4: 个税计算
**Goal**: 基于工资数据自动匹配专项附加扣除并精准计算个税，申报截止前自动提醒
**Depends on**: Phase 2
**Requirements**: TAX-01, TAX-02, TAX-03, TAX-04, TAX-05, TAX-06
**Success Criteria** (what must be TRUE):
  1. 基于工资核算数据自动匹配个税专项附加扣除项（子女教育、房贷等）
  2. 按中国累计预扣预缴法精准计算个税，正确处理税率跳档
  3. 个税申报截止前3天自动提醒老板，并生成申报表供手动提交
  4. 个税申报明细可查询状态、可导出凭证
**Plans**: TBD

### Phase 5: 工资核算
**Goal**: 老板可自定义薪资结构，一键核算月度工资（自动关联社保和个税扣款），生成电子工资单推送至员工
**Depends on**: Phase 3, Phase 4
**Requirements**: PAYR-01, PAYR-02, PAYR-03, PAYR-04, PAYR-05, PAYR-06, PAYR-07, PAYR-08, PAYR-09
**Success Criteria** (what must be TRUE):
  1. 老板可自定义薪资结构（基本工资、绩效、补贴、扣款等），一键核算自动关联社保和个税扣款
  2. 支持"复制上月工资表"快速核算，支持导入考勤表Excel辅助核算
  3. 自动生成电子工资单并推送至员工手机，员工可在线确认签收
  4. 工资条可导出Excel，每月发放状态/金额/方式有记录，异常发放自动提醒
**Plans**: TBD

### Phase 6: 财务记账
**Goal**: 老板可完成小微企业完整财务记账流程：录入凭证、管理发票、费用报销审批、查看账簿报表、月度结账
**Depends on**: Phase 2, Phase 5
**Requirements**: FINC-01, FINC-02, FINC-03, FINC-04, FINC-05, FINC-06, FINC-07, FINC-08, FINC-09, FINC-10, FINC-11, FINC-12, FINC-13, FINC-14, FINC-15, FINC-16, FINC-17, FINC-18, FINC-19, FINC-20, FINC-21, FINC-22
**Success Criteria** (what must be TRUE):
  1. 老板可手动录入会计凭证，借贷平衡实时校验，不平衡时阻止提交；已审核凭证只能红冲
  2. 员工提交费用报销后老板在线审批，审批通过后自动生成费用凭证
  3. 基于凭证数据实时生成总账、明细账、科目余额表，支持导出Excel
  4. 月末结账后自动生成资产负债表和利润表，支持多期对比；结账锁定当期凭证
  5. 基于发票和凭证数据自动计算增值税和企业所得税，生成纳税申报辅助数据
**Plans**: TBD
**UI hint**: yes

### Phase 7: 首页工作台
**Goal**: 老板打开APP第一时间知道要做什么，待办事项和核心功能1步可达
**Depends on**: Phase 1, Phase 2, Phase 3, Phase 4, Phase 5, Phase 6
**Requirements**: HOME-01, HOME-02, HOME-03, HOME-04, HOME-05, HOME-06
**Success Criteria** (what must be TRUE):
  1. 首页工作台展示所有待办事项卡片（社保缴费/个税申报/入职离职/费用报销/凭证审核），完成后自动消失
  2. 核心功能5宫格入口（员工/社保/工资/个税/财务），1步直达
  3. 数据概览区展示在职员工数、本月入职/离职数、社保总额、工资总额
  4. 底部5Tab导航（首页/员工/工具/财务/我的）贯穿所有页面
**Plans**: TBD
**UI hint**: yes

### Phase 8: 员工微信小程序
**Goal**: 员工通过微信小程序查看工资条、合同、社保记录，提交费用报销
**Depends on**: Phase 2, Phase 3, Phase 5
**Requirements**: WXMP-01, WXMP-02, WXMP-03, WXMP-04, WXMP-05, WXMP-06
**Success Criteria** (what must be TRUE):
  1. 员工通过微信授权登录小程序，查看各月工资条（含明细），敏感信息需短信验证
  2. 员工可查看合同状态和社保记录
  3. 员工可提交费用报销（上传票据、填写金额/类型），查看报销状态
**Plans**: TBD
**UI hint**: yes

## Progress

**Execution Order:**
Phases execute in numeric order: 1 -> 2 -> 3 -> 4 -> 5 -> 6 -> 7 -> 8

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. 基础框架与用户认证 | 4/4 | Complete | 2026-04-06 |
| 2. 员工管理 | 1/4 | In Progress|  |
| 3. 社保管理 | 0/TBD | Not started | - |
| 4. 个税计算 | 0/TBD | Not started | - |
| 5. 工资核算 | 0/TBD | Not started | - |
| 6. 财务记账 | 0/TBD | Not started | - |
| 7. 首页工作台 | 0/TBD | Not started | - |
| 8. 员工微信小程序 | 0/TBD | Not started | - |
