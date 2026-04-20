# Phase 12: 考勤合规报表 - Context

**Gathered:** 2026-04-20
**Status:** Ready for planning

<domain>
## Phase Boundary

合规要求的历史统计报表（加班/请假/异常/月度汇总），用于留存备查。

具体包含：
- 加班统计报表（法定节假日/延时加班/周末加班分3档统计）
- 请假合规报表（按员工维度，年假剩余/已用、病假、事假）
- 出勤异常报表（按异常次数统计，迟到/早退/缺勤）
- 月度考勤汇总 Excel 导出（统计表格式，按部门筛选，支持近12个月）

**Scope:** H5管理后台（新增独立合规报表菜单），后端 API 配合。
**Depends on:** Phase 06（考勤管理已有打卡/月报基础）

</domain>

<decisions>
## Implementation Decisions

### 报表入口（D-12-01）
- **D-12-01:** 新增「合规报表」独立一级菜单，不复用考勤月报页面；4个子页面：加班统计/请假合规/出勤异常/月度汇总
- **D-12-02:** 所有报表操作步骤 ≤ 3步（选择月份/部门 → 查看报表 → 导出 Excel）

### 加班统计报表（COMP-05）
- **D-12-03:** 按场景分3档加班：法定节假日加班（3倍工资）、工作日延时加班（1.5倍）、周末加班（2倍）
- **D-12-04:** 加班时长按 0.5h 取整（复用 Phase 06 D-07 决策）
- **D-12-05:** 法定节假日来源：复用 AttendanceRule.Holidays（Phase 06 管理员手动录入）
- **D-12-06:** 加班时长存储精确到 0.01h，显示四舍五入到 0.5h

### 请假合规报表（COMP-06）
- **D-12-07:** 按员工维度汇总：每人一行，显示年假剩余/已用、病假天数、事假天数
- **D-12-08:** 年假额度来源：管理员手动配置（不自动计算）；可按月份筛选

### 出勤异常报表（COMP-07）
- **D-12-09:** 按异常次数统计：每人显示迟到次数、早退次数、缺勤天数、累计异常时长
- **D-12-10:** 异常次数多的员工用红色高亮标注（异常次数 > 阈值可配置）

### 月度考勤汇总 Excel 导出（COMP-08）
- **D-12-11:** 导出格式：统计汇总表（每员工一行），列：应出勤天数/实际出勤天数/迟到次数/早退次数/缺勤天数/加班小时数/请假天数（分年假/病假/事假）
- **D-12-12:** 导出支持按部门筛选（单选/多选/全选）
- **D-12-13:** 支持近12个月的历史数据导出

### Claude's Discretion
- 异常红色高亮的具体阈值（默认：迟到>3次 or 缺勤>1天）
- 加班费率（3倍/1.5倍/2倍）的具体计算逻辑由薪资模块后续实现，本报表仅统计时长
- 统计表的具体列宽和样式

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements
- `.planning/REQUIREMENTS.md` §COMP-05~COMP-08 — 考勤合规报表需求定义
- `.planning/ROADMAP.md` §Phase 12 — 阶段目标，成功标准

### Prior Phase Context
- `.planning/phases/06-考勤管理/06-CONTEXT.md` — Phase 06 决策（加班0.5h取整、ClockRecord分离work_date/clock_time、出勤月报格子矩阵、法定节假日管理员录入）
- `.planning/phases/11-合同合规/11-CONTEXT.md` — Phase 11 决策（前端UI组件模式）

### Existing Attendance Code
- `internal/attendance/model.go` — ClockRecord（含 work_date/clock_time 分离 per D-14）、AttendanceRule（含 Holidays JSONB）
- `internal/attendance/service.go` — 出勤月报服务（复用此模式扩展合规报表）
- `frontend/src/api/attendance.ts` — 现有考勤 API（ClockLive、LeaveStats）
- `internal/attendance/dto.go` — DTO 结构（LeaveStats、ClockRecordResponse）

### Tech Stack
- excelize v2.10.1 — Excel 导出（backend）
- xlsx (SheetJS) — 浏览器端 Excel 预览（如需要）

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/attendance/service.go`: 出勤月报服务，可扩展统计逻辑；复用 errgroup 并发聚合模式
- `frontend/src/views/employee/EmployeeDashboard.vue`: 看板数字卡片模式，报表顶部统计卡片复用
- `frontend/src/api/attendance.ts`: 已有 ClockLiveResponse、LeaveStats 类型，可扩展
- `internal/attendance/model.go` ClockRecord: clock_time 已存储，可叠加 Holidays 判断节假日加班

### Established Patterns
- 统计汇总表格式（员工行×统计列）— 与 Phase 07 薪资统计风格一致
- 部门多选复用 Phase 07 的 `__all__` sentinel + toggleSelectAllDepts 模式
- Excel 导出复用 Phase 08 的 Excel handler 模式（`export=full` 参数）

### Integration Points
- attendance → compliance: ClockRecord + AttendanceRule.Holidays 计算节假日加班
- attendance → approval: 请假申请数据来自现有 approval_service，统计维度按 employee_id 聚合
- attendance → excel: excelize 后端生成，写入 gin.Context.Data

</code_context>

<specifics>
## Specific Ideas

- 异常红色高亮标注：迟到>3次 or 缺勤>1天 → 用 el-tag type="danger"
- 加班统计表格列：员工姓名/部门/法定节假日加班小时数/工作日延时加班小时数/周末加班小时数/合计小时数
- 请假合规表格列：员工姓名/部门/应享年假天数/已用年假/剩余年假/病假天数/事假天数
- 月度考勤汇总表格列：应出勤/实出勤/迟到次数/早退次数/缺勤天数/加班时数/年假/病假/事假

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 12-考勤合规报表*
*Context gathered: 2026-04-20*
