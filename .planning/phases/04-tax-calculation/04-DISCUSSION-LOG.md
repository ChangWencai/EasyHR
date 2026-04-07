# Phase 4: 个税计算 - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in 04-CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-07
**Phase:** 04-tax-calculation
**Mode:** auto (all decisions auto-selected with recommended defaults)
**Areas discussed:** 个税税率表存储、专项附加扣除管理、工资数据获取与循环依赖解耦、累计预扣预缴计算、申报提醒机制、申报表格式

---

## 个税税率表存储

| Option | Description | Selected |
|--------|-------------|----------|
| Database table (OrgID=0) | 税率表存数据库，全局共享，与社保政策库模式一致。管理员可通过H5后台更新。 | ✓ |
| Config file (YAML) | 税率表写在config.yaml中，更新需重启服务。简单但不灵活。 | |
| Hardcoded constants | Go代码中hardcode税率和速算扣除数。最简单但政策变更需重新编译部署。 | |

**Auto-selected:** Database table (OrgID=0) — 与社保政策库模式一致，管理员可自主更新，无需重新部署

**Notes:** 中国个税七级超额累进税率自2018年改革后相对稳定，但起征点和扣除标准可能调整。数据库存储提供灵活性。

---

## 专项附加扣除管理

| Option | Description | Selected |
|--------|-------------|----------|
| 老板在管理后台录入 | 老板为员工逐项录入扣除信息，V1.0简化方案。核心用户是老板。 | ✓ |
| 员工在小程序自行填写 | 员工通过微信小程序填写自己的扣除项。降低老板工作量但增加员工端复杂度。 | |
| 混合模式 | 员工填写+老板确认。功能完善但实现复杂度高。 | |

**Auto-selected:** 老板在管理后台录入 — V1.0核心用户是老板，操作简洁。扣除标准预置系统，老板只需选类型+填条件。

**Notes:** 支持7项全部扣除类型。扣除标准由系统预置（如子女教育2000元/孩/月），老板选类型+人数即可。

---

## 工资数据获取与循环依赖解耦

| Option | Description | Selected |
|--------|-------------|----------|
| Tax暴露计算接口接收参数 | `CalculateTax(orgID, empID, month, grossIncome)`，工资作为参数传入。彻底解耦。 | ✓ |
| Tax直接查询Contract表 | Tax模块直接读取Contract.Salary获取基本工资。简单但无法处理非固定收入。 | |
| Tax定义SalaryProvider接口 | 定义接口由Phase 5实现。引入Phase 4→5依赖，增加复杂度。 | |

**Auto-selected:** Tax暴露计算接口接收参数 — Tax模块作为独立计算服务，不依赖任何工资模块数据。Phase 5作为调用方。彻底解耦循环依赖。

**Notes:** Tax模块同时提供独立查询能力，此时税前收入来源为Contract.Salary。社保扣款通过SocialInsuranceProvider接口获取。

---

## 累计预扣预缴计算

| Option | Description | Selected |
|--------|-------------|----------|
| 实时累加历史记录 | 每月计算时从TaxRecord表SUM YTD数据。数据来源单一，可靠。 | ✓ |
| 年度累计快照表 | 额外表存储每月累计值。查询快但需维护同步。 | |

**Auto-selected:** 实时累加历史记录 — 简单可靠，数据来源唯一。员工数≤50，性能无忧。

**Notes:** TaxRecord包含完整计算快照。年初（1月）累计数据自动按year字段清零。

---

## 申报提醒机制

| Option | Description | Selected |
|--------|-------------|----------|
| 复用gocron模式 | 与社保模块一致的定时任务模式。已验证的pattern。 | ✓ |
| 使用asynq异步队列 | 基于Redis的异步任务。功能更强但引入新依赖。 | |

**Auto-selected:** 复用gocron模式 — 与社保提醒机制完全一致，已验证的pattern，无额外学习成本。

**Notes:** 截止日为每月15日（法定），APP内消息+待办卡片触达。

---

## 申报表格式

| Option | Description | Selected |
|--------|-------------|----------|
| Excel为主 + PDF为辅 | Excel方便手动录入电子税务局，PDF正式打印存档。 | ✓ |
| 仅Excel | 简单，满足主要需求。 | |
| 仅PDF | 正式但老板复制数据到电子税务局不方便。 | |

**Auto-selected:** Excel为主 + PDF为辅 — 老板需手动提交到自然人电子税务局，Excel格式对齐批量导入模板最实用。

**Notes:** 申报表字段与自然人电子税务局的批量导入格式对齐，减少手动录入量。

---

## Claude's Discretion

- 个税模块内部目录结构
- 税率表 JSONB vs 关系表
- TaxRecord 累计字段冗余设计
- 申报表 Excel 模板格式
- PDF 凭证模板排版
- gocron 调度器共享方式

## Deferred Ideas

None — all discussion stayed within phase scope
