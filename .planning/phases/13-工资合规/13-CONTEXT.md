# Phase 13: 工资合规 - Context

**Gathered:** 2026-04-20
**Status:** Ready for planning

<domain>
## Phase Boundary

工资条发放确认全流程管理——员工主动确认收到工资条，系统自动存档并检测未确认记录推送提醒老板，确保工资发放有据可查。

具体包含：
- 工资条确认回执：员工在 H5 工资条页面点击「确认已收到」按钮，完成主动确认
- 确认记录存档：扩展 PayrollSlip 表，记录 ConfirmedAt + ConfirmedIP
- 未确认自动提醒：asynq 定时任务每日检测上月未确认工资条，通过 TodoCenter 推送老板

**Scope:** 员工端微信小程序工资条页（H5 签署页同款）+ H5 管理后台工资条发送页 + 后端 API + asynq 定时任务
**Depends on:** Phase 07（工资条发送）、Phase 09（待办中心）

</domain>

<decisions>
## Implementation Decisions

### 确认机制（COMP-09）
- **D-13-01:** 员工需**主动点击「确认已收到」**按钮完成确认，查看工资条不等于确认。确认操作是不可抵赖的主动行为。
- **D-13-02:** 确认按钮显示位置：SalarySlipH5.vue 工资条明细页面底部，在所有工资条项目展示完毕后，显示「确认已收到」按钮。
- **D-13-03:** 点击确认后：更新 PayrollSlip.confirmed_at + confirmed_ip，记录确认时间戳和 IP 地址，跳转到确认成功页（显示"您已确认 2026年4月 工资条"）。
- **D-13-04:** 员工确认后状态不再回退（确认不可撤销）。

### 确认记录存档（COMP-10）
- **D-13-05:** 在 PayrollSlip 模型新增字段：ConfirmedAt (timestamp)、ConfirmedIP (string)，扩展 payroll_slips 表结构。
- **D-13-06:** 存档数据：员工姓名/月份/确认时间/IP地址。由后端自动从请求中提取 IP（`c.ClientIP()`）。
- **D-13-07:** 确认记录幂等：员工对同一工资条多次确认只更新 ConfirmedAt，不重复创建记录。

### 未确认提醒（COMP-11）
- **D-13-08:** asynq 定时任务（gocron）每日凌晨检测上月未确认工资条，通过 TodoCenter 推送给老板（Owner 角色）。任务名：`remind-unconfirmed-slips`。
- **D-13-09:** 提醒文案（待细化）："您有 {N} 名员工尚未确认 {年}年{月} 工资条，请及时跟进。"
- **D-13-10:** 提醒幂等：同一月份同一工资条不重复推送（使用 TodoItem.SourceType + SourceID 幂等键）。

### 老板端确认状态展示
- **D-13-11:** 在 SalarySlipSend.vue 的发送日志表格中新增「确认状态」列，显示每条工资条：未确认 / 已确认（时间）/ 已查看未确认。无需新增独立菜单。
- **D-13-12:** 确认状态数据来自 PayrollSlip 表（新增字段）的现有 API，无需新增接口。

### Claude's Discretion
- 确认成功页面的具体文案（简洁即可）
- 定时任务的具体执行时间（建议每日早上9点）
- 表格确认状态列的排序规则（默认按确认状态排序，未确认优先）
- 确认记录查询 API 的分页和筛选参数

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements
- `.planning/REQUIREMENTS.md` §COMP-09~COMP-11 — 工资合规需求定义
- `.planning/ROADMAP.md` §Phase 13 — 阶段目标，成功标准

### Prior Phase Context
- `.planning/phases/07-薪资管理增强/07-CONTEXT.md` — Phase 07 决策（工资条发送通道、INSERT ONLY 调薪、asynq 批量发送）
- `.planning/phases/09-待办中心/09-CONTEXT.md` — Phase 09 决策（TodoItem 字段、CreateTodo 幂等性、asynq 定时任务模式）
- `.planning/phases/11-合同合规/11-CONTEXT.md` — Phase 11 决策（签署流程：短信链接→H5确认→存档模式）

### Existing Code
- `internal/salary/slip.go` — PayrollSlip 模型（含 pending→sent→viewed→signed 状态机）、SendSlip、VerifySlipCode
- `internal/salary/slip_send_service.go` — asynq worker 处理工资条发送、SendAllSlips idempotent 重发逻辑
- `internal/salary/slip_send_handler.go` — /salary/slip/send-all HTTP handler
- `frontend/src/views/tool/SalarySlipH5.vue` — 员工端 H5 工资条页面（需新增确认按钮）
- `frontend/src/views/tool/SalarySlipSend.vue` — 工资条发送页面（需新增确认状态列）
- `frontend/src/api/salary.ts` — salaryApi（现有 sendSlipAll、getSlipLogs）
- `internal/todo/service.go` — CreateTodoFromEmployee（待办创建）、TodoItem 模型

### Tech Stack
- gocron v2.19.1 — 定时任务（参考 Phase 09 用法）
- asynq — 后端异步任务（参考 Phase 07 用法）

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- PayrollSlip 模型（`internal/salary/slip.go`）：可扩展 confirmed_at + confirmed_ip 字段
- TodoItem（`internal/todo/model.go`）：已有 SourceType、SourceID、EmployeeID 等字段，可直接使用
- asynq worker 模式（`slip_send_service.go`）：参考其 asynq task 处理和重试模式
- SalarySlipH5.vue：现有员工端工资条展示页面，在底部加确认按钮

### Established Patterns
- 确认/签署后不可撤销（D-13-04）—— 参考 Phase 11 合同签署模式
- TodoCenter 幂等创建（SourceType+SourceID）—— Phase 09 已建立
- asynq 后台处理 + HTTP handler 立即返回 —— Phase 07 工资条发送模式
- db migration 扩展现有表 —— 参考 Phase 12 AnnualLeaveQuota 扩展 dto.go

### Integration Points
- PayrollSlip 表：新增 ConfirmedAt + ConfirmedIP 字段
- SalarySlipH5.vue：新增「确认已收到」按钮，POST /salary/slip/confirm
- SalarySlipSend.vue：新增确认状态列，读取现有 API 带上新字段
- gocron 定时任务：在 attendance/salary 模块注册定时任务，每日检测未确认工资条

</code_context>

<specifics>
## Specific Ideas

- 确认按钮文案：「确认已收到」（简洁明确，员工无需理解"签收"等术语）
- 确认成功页：显示"您已确认 2026年4月 工资条，感谢确认"
- 定时任务时间：建议每日早上 9:00（gocron）执行
- 提醒标题：待办中心创建新 Type：`salary_unconfirmed`

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 13-工资合规*
*Context gathered: 2026-04-20*
