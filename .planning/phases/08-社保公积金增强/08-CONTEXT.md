# Phase 8: 社保公积金增强 - Context

**Gathered:** 2026-04-18
**Status:** Ready for planning

<domain>
## Phase Boundary

管理员获得社保公积金数据洞察，增减员流程优化，缴费渠道和欠缴状态自动管理。

具体包含：
- 社保数据看板（应缴总额/单位部分/个人部分/欠缴金额 + 环比上月百分比）
- 增员弹窗优化（姓名检索 + 起始月份 + 城市 + 社保基数 + 公积金比例和基数）
- 减员弹窗优化（姓名检索 + 终止月份 + 原因 + 转出/封存日期 + 生效规则提示）
- 缴费渠道管理（自主/代理新客/代理已合作）+ 状态自动流转（正常→待缴→欠缴→已转出）
- 5 种新状态 UI（正常/待缴/欠缴/已转出/未转出）
- 红字横幅提醒 + 政策通知
- 五险分项明细 + Excel 导出

**Scope:** 仅 H5 管理后台，后端 API 配合新增。
**Depends on:** Phase 5（花名册增强用于减员联动）

</domain>

<decisions>
## Implementation Decisions

### 缴费状态模型
- **D-SI-01:** 新建 `SIMonthlyPayment` 表（employee_id + year_month + status + payment_channel + company_amount + personal_amount + total_amount）。status 月度独立追踪，不与参保生命周期 conflate。Organization 表加 payment_channel 字段（si_payment_channel: self/agent_new/agent_existing）作为默认值
- **D-SI-02:** 新表 SIMonthlyPayment 由 asynq 定时任务每月生成（插入下月记录），同时删除跨度过期记录（超过24个月）

### 状态自动流转
- **D-SI-03:** asynq 定时任务触发：每天凌晨运行，检查所有 SIMonthlyPayment 记录：
  - 当月 ≥ 26 日且未缴 → pending → overdue
  - 当月 < 26 日且已确认 → pending → normal
  - 自主缴费：用户手动确认后更新状态
  - 代理已合作：SI-16 扣缴成功/失败 webhook 更新状态
- **D-SI-04:** 代理缴费 webhook（SI-16）：asynq 接收扣缴结果（成功/失败），成功则 SIMonthlyPayment.status = normal，失败则待缴/欠缴

### 数据看板
- **D-SI-05:** 4 张纯数字卡片（应缴总额/单位部分合计/个人部分合计/欠缴金额），带环比上月百分比。与薪资数据看板（SalaryDashboard）风格完全一致，不加月度筛选器
- **D-SI-06:** 环比上月计算：仅统计 confirmed 状态的月份；上月无数据时显示"—"

### 增减员弹窗
- **D-SI-07:** 增员弹窗（EnrollDialog）：姓名输入触发搜索（employeeApi.search by name），显示匹配员工列表选择；起始月份默认当月，可选近3个月（SI-06）；支持单独设置社保基数和公积金基数（SI-07, SI-08）
- **D-SI-08:** 减员弹窗（StopDialog）：姓名输入触发搜索；终止月份默认当月且不可早于当月（SI-10）；原因必选（三选一：跳槽/退休/其他）；转出日期输入后自动提示 SI-13 生效规则
- **D-SI-09:** 增减员 INSERT ONLY：ChangeHistory 表追加记录，不 UPDATE 历史 SI 记录

### 提醒机制
- **D-SI-10:** 横幅 + 列表红色标注：参保操作 Tab 顶部红色横幅展示最大欠缴项（员工姓名 + 城市 + 欠缴月 + 金额），横幅下方滚动展示所有未处理欠缴；参保记录列表行内欠缴项（status=overdue）标红背景

### 参保记录列表增强
- **D-SI-11:** 状态列改为 5 种标签（正常-绿色/待缴-黄色/欠缴-红色/已转出-灰色/未转出-蓝色）；增加缴费渠道列；点击行展开详情弹窗展示五险分项（SI-20）
- **D-SI-12:** 五险分项弹窗（SI-20）：养老/医疗/失业/工伤/生育/公积金各自展示单位缴纳 + 个人缴纳金额，底部合计。另有"其他缴费"行（滞纳金/残保金/漏缴/补缴）

### Excel 导出
- **D-SI-13:** 复用 excelize 导出模式，参保记录导出列：员工/城市/基数/参保月/缴费渠道/状态/五险分项（6险×2列）/合计单位/合计个人。与 SalaryList.vue 导出对话框风格一致

### Claude's Discretion
- 横幅的具体动画效果（静止/滚动/可关闭？）
- 五险分项弹窗的列宽和数字格式（是否加千分位分隔符）
- 政策通知来源（手动录入 vs 同步第三方 API）
- 增员时公积金单独页签还是同表单内切换？
- 欠缴提醒超过多少条时横幅显示策略（全显示/只显示前3条）
- asynq 定时任务的具体 cron 表达式（每天凌晨2点？）
- SI-15 自主缴费跳转至哪个外部页面（需确认第三方缴费平台 URL）

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements
- `.planning/REQUIREMENTS.md` §4 — Phase 8 需求定义（SI-01~SI-21）
- `.planning/ROADMAP.md` §Phase 8 — 阶段目标、成功标准、依赖关系

### Research
- `.planning/research/SUMMARY.md` — v1.3 研究综合报告
- `.planning/research/FEATURES.md` §2.4 — 社保公积金算法（五险计算逻辑）

### Existing Code Patterns
- `internal/socialinsurance/model.go` — SocialInsuranceRecord/ChangeHistory 模型（现有字段需扩展）
- `internal/socialinsurance/service.go` — 现有服务层，新增 Enroll/Stop 方法
- `internal/socialinsurance/repository.go` — 数据访问层
- `internal/socialinsurance/scheduler.go` — 已有 asynq 定时任务机制
- `frontend/src/views/tool/SITool.vue` — 现有社保工具页面（政策库/参保操作/参保记录）
- `frontend/src/views/tool/SalaryDashboard.vue` — 数据看板参考样式（4卡片风格）
- `internal/salary/excel.go` — excelize Excel 导出复用此模式
- `.planning/phases/07-薪资管理增强/07-CONTEXT.md` — Phase 07 决策（D-SAL-DATA-01 月份只读保护可参考）

### Project Decisions
- `.planning/PROJECT.md` — Key Decisions 表（多租户 org_id、RBAC、INSERT ONLY 策略）
- `05-CONTEXT.md` — Phase 05 决策（Department 模型、花名册增强）
- `07-CONTEXT.md` — Phase 07 决策（asynq 批量发送、INSERT ONLY 策略）

### Go Module Dependency
- `asynq`（已安装）— 定时任务 + 批量操作队列
- `excelize`（已安装）— Excel 读写

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/socialinsurance/scheduler.go`: 已有 asynq scheduler，新月度缴费记录生成和状态流转复用此定时任务框架
- `internal/socialinsurance/excel.go`: 已有导出逻辑，扩展五险分项列复用
- `frontend/src/views/tool/SalaryDashboard.vue`: 4卡片风格直接复用
- `frontend/src/views/tool/SalaryList.vue`: 导出对话框（当前页/含明细）可直接复用

### Established Patterns
- Handler → Service → Repository 三层架构
- INSERT ONLY：ChangeHistory 追加记录，不 UPDATE 历史
- asynq 定时任务：每日凌晨触发，检查并更新状态
- org_id 逻辑多租户隔离

### Integration Points
- employee → SI: 花名册减员联动（Phase 05），减员时触发 SIMonthlyPayment 记录
- payroll → SI: PayrollRecord 引用 SI 扣款金额（已有字段）
- scheduler → SI: asynq 定时任务生成月度记录 + 状态流转
- asynq → SI: 代理缴费 webhook 更新扣缴状态

### Critical Anti-Patterns to Avoid
- ❌ UPDATE 历史 SIMonthlyPayment 记录 → 状态变更只能 INSERT 新记录或 UPDATE pending 状态
- ❌ conflate 参保状态（pending/active/stopped）与缴费状态（normal/pending/overdue/transferred/not_transferred）
- ❌ 浮点计算金额 → 使用 shopspring/decimal

</code_context>

<specifics>
## Specific Ideas

- 缴费渠道默认"自主缴费"（Organization.si_payment_channel）
- 减员时终止月份默认当月且不可早于当月
- 增员起始月份可选近3个月（SI-06）
- 红色横幅在参保操作 Tab 顶部，不在其他 Tab 显示
- 政策通知：初期管理员手动录入，存 policy_notices 表（id/title/content/priority/city_id）

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 08-社保公积金增强*
*Context gathered: 2026-04-18*
