# Phase 13: 工资合规 - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-20
**Phase:** 13-工资合规
**Areas discussed:** 确认机制, 提醒触发, 存档内容, 老板视图

---

## 确认机制

| Option | Description | Selected |
|--------|-------------|----------|
| 主动确认回执 | 员工查看工资条后显示「确认已收到」按钮，点击后完成确认。员工需要主动操作，存档有明确时间戳和 IP。 | ✓ |
| 查看即确认（隐式） | 员工打开工资条即视为确认，无需额外操作。操作最简，但部分员工可能无意识确认。 | |
| 两步确认 | 工资条查看和确认分开两步：先查看明细，再弹窗要求点击「确认」才存档。 | |

**User's choice:** 主动确认回执
**Notes:** 不可抵赖的主动行为比隐式确认更有法律效力，符合合规目标。

---

## 提醒触发

| Option | Description | Selected |
|--------|-------------|----------|
| 自动每日检测（推荐） | asynq 定时任务（如每日早9点）自动扫描上月未确认工资条，发现后通过 TodoCenter 自动推送老板。无需老板手动操作。 | ✓ |
| 老板手动催确认 | 老板可在工资条列表页手动点击「催确认」按钮，给所有未确认员工发送短信/H5 链接提醒。老板掌控提醒时机。 | |
| 两者并存 | 自动检测 + 手动催确认并存。系统自动检测推送，老板也可随时手动催。 | |

**User's choice:** 自动每日检测（推荐）
**Notes:** 与合同合规的自动提醒模式一致（Phase 11 D-11-08：合同签署第3天自动提醒）。

---

## 存档内容

| Option | Description | Selected |
|--------|-------------|----------|
| 扩展 PayrollSlip 表 | 在现有 PayrollSlip 模型新增字段：ConfirmedAt (timestamp)、ConfirmedIP (string)、ConfirmedDevice (string)。不改表结构，但需 db migration。 | ✓ |
| 新建确认记录表 | 新建 PayrollSlipConfirmation 表：employee_id/employee_name/year/month/confirmed_at/ip/device。表结构更清晰，但多一张表。 | |

**User's choice:** 扩展 PayrollSlip 表
**Notes:** 保持单一模型，字段扩展最小化。

---

## 老板视图

| Option | Description | Selected |
|--------|-------------|----------|
| 在工资条发送页展示（推荐） | 在 SalarySlipSend.vue 的发送记录表格中新增「确认状态」列，显示每条工资条的已确认/未确认状态。老板无需跳转到新页面。 | ✓ |
| 独立合规报表菜单 | 在 H5 管理后台「财务记账」菜单下新增独立「工资条合规」页面，显示所有员工的工资条发放和确认状态汇总。独立视图更清晰。 | |

**User's choice:** 在工资条发送页展示（推荐）
**Notes:** 与 Phase 12 考勤合规报表模式相反——考勤是独立菜单，工资合规是嵌入现有页面。

---

## Deferred Ideas

None — all scope items resolved in this discussion.
