# Phase 8: 社保公积金增强 - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-18
**Phase:** 08-社保公积金增强
**Areas discussed:** 缴费状态模型 + 流转机制, 数据看板 UI 风格, 欠缴提醒呈现方式, 参保记录列表增强

---

## 缴费状态模型 + 流转机制

| Option | Description | Selected |
|--------|-------------|----------|
| 轻量：字段扩展 | 在 SocialInsuranceRecord 加 payment_status 列，payment_channel 加到 Organization 表，asynq 定时任务流转 | |
| 完整：月度缴费表 | 新建 SIMonthlyPayment 表（employee_id + year_month + status + channel + amount），Organization.payment_channel 作为默认值，asynq 定时任务更新月度状态 | ✓ |

**User's choice:** 完整：月度缴费表
**Notes:** 清晰区分参保状态和缴费状态，月度独立追踪

---

### 状态自动流转

| Option | Description | Selected |
|--------|-------------|----------|
| 定时任务自动流转 | asynq 定时任务每天凌晨运行，当月26日后自动流转（正常→待缴→欠缴） | ✓ |
| 按需计算 | 每次打开参保记录页面时计算当月状态 | |
| 手动确认 | 管理员手动确认，定时任务只发提醒通知 | |

**User's choice:** 定时任务自动流转
**Notes:** asynq 定时任务检查所有记录，代理缴费 webhook 更新扣缴状态

---

## 数据看板 UI 风格

| Option | Description | Selected |
|--------|-------------|----------|
| 与薪资数据看板一致 | 4张纯数字卡片，无月度筛选器，只显示当月 | ✓ |
| 加月度筛选器 | 顶部加月度选择器（默认当月），可切换查看历史月份 | |

**User's choice:** 与薪资数据看板一致
**Notes:** 简单统一，4卡片（应缴总额/单位/个人/欠缴）带环比百分比

---

## 欠缴提醒呈现方式

| Option | Description | Selected |
|--------|-------------|----------|
| 红色横幅在参保操作 Tab 顶部 | 横幅滚动展示所有未处理欠缴，政策通知在横幅下方 | |
| 横幅 + 列表行内红色标注 | 横幅展示最大欠缴项，欠缴行标红背景 | ✓ |
| 轻量提示条 + 列表红色高亮 | 字号小提示条，欠缴行红色背景高亮 | |

**User's choice:** 横幅 + 列表行内红色标注
**Notes:** 横幅展示当前最大欠缴项（员工姓名 + 城市 + 欠缴月 + 金额）

---

## 参保记录列表增强

| Option | Description | Selected |
|--------|-------------|----------|
| 状态标签 + 渠道列 + 展开详情弹窗 | 5种状态标签（正常-绿/待缴-黄/欠缴-红/已转出-灰/未转出-蓝），增加缴费渠道列，点击行展开五险分项弹窗 | ✓ |
| 全列平铺不折叠 | 所有信息平铺在列表中 | |

**User's choice:** 状态标签 + 渠道列 + 展开详情弹窗
**Notes:** 五险分项弹窗：养老/医疗/失业/工伤/生育/公积金各自展示单位+个人，底部合计

---

## Deferred Ideas

None — discussion stayed within phase scope
