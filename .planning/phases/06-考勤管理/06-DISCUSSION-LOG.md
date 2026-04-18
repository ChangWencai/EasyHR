# Phase 06: 考勤管理 - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-18
**Phase:** 06-考勤管理
**Areas discussed:** 审批流实现方式, 出勤月报展示方式, 加班时长精度, 打卡模式优先级, 跨天班次归属

---

## 审批流实现方式

| Option | Description | Selected |
|--------|-------------|----------|
| 用状态机库 | 新增 qmuntal/stateless v1.8.0，11种审批类型统一状态机，避免遗漏转换路径 | ✓ |
| 状态常量 if-else | 沿用现有 offboarding/expense 的状态常量模式，代码一致但复杂后风险高 | |
| 自封装轻量层 | 自己封装一个轻量审批基类，不用第三方库但保留结构化 | |

**User's choice:** 用状态机库
**Notes:** 11种审批类型复杂度高，状态机库可以统一管理状态转换，避免遗漏 cancelled/timeout 等边缘状态。

---

## 出勤月报展示方式

| Option | Description | Selected |
|--------|-------------|----------|
| 统计行+日历详情 | 出勤率卡片 + 应/实/加班时长统计行，点击员工展开当月日历打卡详情 | |
| 格子矩阵表 | 每个员工一行 × 每月30+列，横屏滚动，类似 Excel | |
| 两种视图可选 | 默认统计行，顶部切换按钮可切换到格子矩阵表 | ✓ |

**User's choice:** 两种视图可选
**Notes:** 满足不同管理员的使用习惯，快速看数据用统计行，详细核对用格子表。

---

## 加班时长精度

| Option | Description | Selected |
|--------|-------------|----------|
| 0.5h 取整 | 加班时长按0.5h（半小时）取整，钉钉/飞书行业惯例，简单直观 | ✓ |
| 精确 0.01h | 精确到0.01小时，计算精确但不够直观 | |

**User's choice:** 0.5h 取整（推荐）
**Notes:** 存储精确到0.01h，显示时四舍五入到0.5h，两全其美。

---

## 打卡模式优先级

| Option | Description | Selected |
|--------|-------------|----------|
| 顶部Tab平等展示 | 打卡规则设置页顶部3个Tab，三者平等展示，切换即时生效 | ✓ |
| 向导式逐步引导 | 先选模式，选完后才出现对应配置表单，一次性决策 | |

**User's choice:** 顶部Tab平等展示
**Notes:** 三种打卡模式同等重要，界面清晰分区，符合产品"简单"核心价值。

---

## 跨天班次归属

| Option | Description | Selected |
|--------|-------------|----------|
| 归班次起始日 | 夜班等跨天班次归属班次起始日，22:00-06:00→归起始日（如周一22:00→归周一） | ✓ |
| 归班次结束日 | 跨天班次归属结束日，22:00-06:00→归结束日 | |
| 手动指定 | 排班时额外指定归属日期，最灵活但配置成本高 | |

**User's choice:** 归班次起始日
**Notes:** 简单直观。Shift 模型包含 work_date_offset 字段区分是否跨天。ClockRecord 存储 work_date（归属）和 clock_time（实际）两个独立字段。

---

## Claude's Discretion

以下领域用户未指定细节，由 Claude 在实现时自行决定：
- 打卡规则设置每个 Tab 内部的具体字段布局和默认值
- 今日打卡实况的筛选器和排序规则
- 出勤月报格子矩阵表的列宽和颜色标注
- 审批超时机制的具体触发时间
- 请假附件拍照上传的具体实现
- 法定节假日来源（初期手动录入）

## Deferred Ideas

无 — 讨论保持在 Phase 06 范围内，未出现范围蔓延。
