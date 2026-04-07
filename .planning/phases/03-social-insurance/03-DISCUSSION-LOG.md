# Phase 3: 社保管理 - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-07
**Phase:** 03-social-insurance
**Areas discussed:** 社保政策库数据结构, 参保/停缴操作流程, 缴费提醒机制, 社保与工资联动

---

## 社保政策库数据结构

### 存储方案

| Option | Description | Selected |
|--------|-------------|----------|
| 单表+JSONB | 每城市一条记录，JSONB存各险种配置 | ✓ |
| 双表关联 | 城市表+险种配置表，每城市5条记录 | |
| 三表规范化 | 城市表+险种定义表+比例表 | |

**User's choice:** 单表+JSONB
**Notes:** 利用PostgreSQL JSONB查询能力，管理员更新政策直接编辑JSON字段，简单灵活

### 初始数据录入方式

| Option | Description | Selected |
|--------|-------------|----------|
| 管理员手动录入 | H5后台手动编辑政策数据，每年更新一次 | ✓ |
| Excel模板导入 | 填模板后批量上传 | |
| 代码内硬编码 | Go代码中作为初始数据 | |

**User's choice:** 管理员手动录入
**Notes:** V1.0最简单，不需要额外导入功能

### 公积金范围

| Option | Description | Selected |
|--------|-------------|----------|
| 不含公积金 | V1.0只做五险 | |
| 包含公积金 | 五险一金全覆盖 | ✓ |

**User's choice:** 包含公积金
**Notes:** 六险全覆盖，JSONB中每个险种一个key

---

## 参保/停缴操作流程

### 参保流程

| Option | Description | Selected |
|--------|-------------|----------|
| 自动匹配+确认 | 选员工→自动匹配基数→确认→生效（3步） | ✓ |
| 自动匹配+可调整 | 选员工→匹配→可手动调整基数→确认→生效 | |

**User's choice:** 自动匹配+确认
**Notes:** 符合"3步内完成核心操作"原则

### 批量操作

| Option | Description | Selected |
|--------|-------------|----------|
| 仅单个操作 | 一次一个员工 | |
| 支持批量操作 | 勾选多员工批量参保/停缴 | ✓ |

**User's choice:** 支持批量操作
**Notes:** 前端传 employee_ids 数组，后端批量处理

### 停缴触发方式（多选）

| Option | Description | Selected |
|--------|-------------|----------|
| 老板手动停缴 | 在社保模块手动操作 | ✓ |
| 离职自动触发提醒 | onEmployeeResigned 回调 | ✓ |
| 转正自动检查 | 试用期转正时检查社保状态 | ✓ |

**User's choice:** 三种全部支持
**Notes:** 离职后只生成提醒待办，不自动执行停缴

---

## 缴费提醒机制

### 定时任务技术

| Option | Description | Selected |
|--------|-------------|----------|
| gocron定时检查 | 每日扫描到期情况，生成提醒记录 | ✓ |
| asynq异步队列 | 精确调度每个企业的提醒时间 | |

**User's choice:** gocron定时检查
**Notes:** 简单可靠，不需要额外依赖

### 提醒渠道

| Option | Description | Selected |
|--------|-------------|----------|
| APP内消息+待办 | 不需要外部服务，开发简单 | ✓ |
| APP内+微信模板消息 | 需微信API对接 | |
| APP内+短信 | 触达率高但有成本 | |

**User's choice:** APP内消息+待办
**Notes:** V1.0成本控制，不用短信或微信模板消息

### 缴费截止日

| Option | Description | Selected |
|--------|-------------|----------|
| 每月固定日期 | 如每月15日或25日，系统默认值 | ✓ |
| 企业自定义日期 | 老板可配置每月缴费日期 | |

**User's choice:** 每月固定日期
**Notes:** 中国社保各地缴费日期固定，系统设定默认值即可

---

## 社保与工资联动

### 联动接口

| Option | Description | Selected |
|--------|-------------|----------|
| 社保提供查询接口 | GetSocialInsuranceDeduction，Phase 5 调用 | ✓ |
| 接口注入解耦 | 定义 SocialDeductionProvider 接口 | |

**User's choice:** 社保提供查询接口
**Notes:** 单向依赖，简单清晰。V1.0不需要过度解耦

### 薪资变动处理

| Option | Description | Selected |
|--------|-------------|----------|
| 只提醒不自动调整 | 检测薪资变动后生成提醒待办 | ✓ |
| 自动调整基数 | 检测后自动更新社保基数 | |

**User's choice:** 只提醒不自动调整
**Notes:** 老板确认后才操作，避免误操作风险

---

## Claude's Discretion

- 社保政策库 JSONB 内部具体字段命名
- 参保记录的数据模型（按险种拆分行 vs 一条记录存所有险种）
- 缴费明细的记录粒度
- 导出凭证的 Excel/PDF 格式细节
- 社保模块内部目录结构

## Deferred Ideas

None — discussion stayed within phase scope
