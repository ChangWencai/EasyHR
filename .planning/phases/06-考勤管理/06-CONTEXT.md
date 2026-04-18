# Phase 06: 考勤管理 - Context

**Gathered:** 2026-04-18
**Status:** Ready for planning

<domain>
## Phase Boundary

管理员可配置灵活的打卡规则（固定时间/按排班/自由工时三种模式），员工可提交11类假勤审批申请（补卡/调班/出差/外出/7种请假），管理员可查看今日打卡实况和出勤月报。

具体包含：
- 打卡规则设置（3种模式 + 工作日/节假日/打卡位置/方式）
- 今日打卡实况（全员打卡状态列表 + 假勤统计）
- 审批流引擎（11种审批类型，统一状态机管理）
- 出勤月报（双视图：统计行+日历 / 格子矩阵表）

**Scope:** 仅 H5 管理后台，后端 API 配合新增。手机定位打卡、APP推送不在范围内。
**Depends on:** Phase 5（Department 模型用于排班分配）

</domain>

<decisions>
## Implementation Decisions

### 审批流实现
- **D-01:** 使用 `qmuntal/stateless` v1.8.0 状态机库，11种审批类型统一状态机管理
- **D-02:** 状态转换：draft → pending → approved / rejected / cancelled / timeout
- **D-03:** 守卫条件（Guard）：仅审批人在 pending 状态下可操作 approve/reject；申请人可取消自己的申请

### 出勤月报展示
- **D-04:** 默认视图：出勤率统计行 + 应/实/加班时长统计卡片（风格与员工/薪资看板一致），点击员工展开当月日历打卡详情
- **D-05:** 可选视图：格子矩阵表（员工行 × 日期列，横屏滚动，类似 Excel）
- **D-06:** 视图切换：顶部切换按钮，默认折叠到统计行视图

### 加班时长精度
- **D-07:** 加班时长按 0.5h（半小时）取整。1小时23分钟 → 1.5小时。钉钉/飞书行业惯例
- **D-08:** 存储精确到 0.01h，显示时四舍五入到 0.5h 展示

### 打卡模式 UI
- **D-09:** 打卡规则设置页：顶部 Tab（固定时间 / 按排班 / 自由工时），三个 Tab 平等展示
- **D-10:** 打卡实况页：顶部 Tab 切换不同模式下的打卡数据，三模式同等重要
- **D-11:** 默认打开 Tab 顺序：固定时间 → 按排班 → 自由工时

### 跨天班次归属
- **D-12:** 跨天班次（如夜班 22:00-06:00）的打卡记录归属班次起始日（不是结束日）
- **D-13:** Shift 模型必须包含 `work_date_offset int` 字段（研究已确认），用于区分是否跨天
- **D-14:** ClockRecord 存储 `work_date`（归属工作日）和 `clock_time`（实际打卡时间）两个独立字段

### Claude's Discretion
- 打卡规则设置的每个 Tab 内部的具体字段布局和默认值
- 今日打卡实况的筛选器（按部门/按状态）和排序
- 出勤月报格子矩阵表的列宽、颜色标注规则
- 审批流超时机制的具体触发时间（建议：审批通过/驳回后自动超时清理）
- 请假附件拍照上传的具体实现（复用现有文件上传组件）
- 法定节假日来源：初期管理员手动录入节假日日期，支持后续扩展节假日API对接

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements
- `.planning/REQUIREMENTS.md` §2 — Phase 6 需求定义（ATT-01~ATT-21）
- `.planning/ROADMAP.md` §Phase 6 — 阶段目标、成功标准、依赖关系
- `.planning/REQUIREMENTS.md` §3 — Phase 7 考勤联动需求（SAL-13~SAL-16），提前了解下游依赖

### Research
- `.planning/research/ARCHITECTURE.md` §2 — 考勤管理模块架构设计（6个数据模型、API路由、考勤→薪资联动接口）
- `.planning/research/SUMMARY.md` — v1.3 研究综合报告（状态机推荐、加班精度、跨天班次陷阱）

### Existing Code Patterns
- `internal/dashboard/service.go` — Dashboard 聚合模式（errgroup 并发），直接复用
- `internal/finance/service_expense.go` — 现有审批流参考（简单状态模式），新建 attendance 模块时替换为状态机
- `internal/employee/offboarding_service.go` — 审批通过触发跨模块事件（社保减员），attendance 审批通过触发薪资联动
- `internal/department/model.go` — Department 邻接表模型，Schedule 排班分配依赖 DepartmentID
- `internal/employee/model.go` — Employee.DepartmentID 字段，Phase 05 已建好
- `frontend/src/views/employee/EmployeeDashboard.vue` — 看板统计卡片模式，attendance 月报统计行复用此模式
- `frontend/src/views/employee/OffboardingList.vue` — 审批列表行内按钮模式，attendance 审批列表复用

### Project Decisions
- `.planning/PROJECT.md` — Key Decisions 表（多租户 org_id、RBAC、加密策略）
- `.planning/STATE.md` — 项目状态（v1.3 进行中）
- `05-CONTEXT.md` — Phase 05 决策（花名册/Drawer/组织架构），Phase 06 复用 Department 模型

### Go Module Dependency
- `qmuntal/stateless` v1.8.0 — 审批流状态机（需添加到 go.mod）
- `shopspring/decimal` — 已在 go.mod，用于加班费精确计算

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/dashboard/service.go`: errgroup 并发聚合模式，可直接用于考勤数据看板（出勤率/加班时长/缺勤统计）
- `internal/finance/service_expense.go`: 审批流字段结构（EmployeeID/Status/ApproverID/ApprovedAt/RejectedNote），attendance/Approval 模型直接扩展这些字段
- `frontend/src/views/employee/EmployeeDashboard.vue`: 纯数字卡片模式，attendance 月报统计行复用（4张卡片：出勤天数/应出勤/加班时长/缺勤天数）
- `frontend/src/views/employee/OffboardingList.vue`: 审批列表行内按钮 + el-popconfirm 确认，attendance 审批列表复用

### Established Patterns
- Handler → Service → Repository 三层架构（所有模块统一）
- `org_id` 逻辑多租户隔离，GORM Scope 自动注入
- 状态机统一用 qmuntal/stateless，审批类型用 `approval_type` 字符串区分（不是多张表）
- ClockRecord 的 `work_date` 与 `clock_time` 分离（跨天班次归属的关键设计）
- Adapter 接口模式：attendance 模块实现 `AttendanceProvider` 供 salary 模块调用

### Integration Points
- attendance → salary: `AttendanceProvider.GetMonthlyAttendance()` 供薪资计算读取出勤数据
- attendance → todo: 审批申请创建时推送 TodoItem 到待办中心（Phase 09）
- employee → attendance: Employee.DepartmentID 用于排班分配员工
- department → attendance: Schedule 排班记录关联 Department（跨天排班用 work_date_offset）

### Critical Anti-Patterns to Avoid
- ❌ ClockRecord 不存储 `work_date`，导致跨天班次无法归属 → 必须分离 work_date 和 clock_time
- ❌ UPDATE 历史 AttendanceMonthly 记录 → 月报预计算后禁止修改，只能通过补卡审批更新
- ❌ 状态机遗漏 cancelled/timeout 状态 → 申请人可主动取消，超时由 gocron 驱动

</code_context>

<specifics>
## Specific Ideas

- 法定节假日默认不用打卡（AttendanceRule.Holidays JSON），初期管理员手动录入
- 今日打卡实况的打卡时间列用颜色标注：正常（绿色）/ 迟到（黄色）/ 缺勤（红色）
- 出勤月报格子矩阵表节假日列底色高亮（浅红），法定节假日自动标注
- 审批列表Tab：全部 / 待我审批 / 我发起的（与现有 OffboardingList 风格一致）
- 请假时长自动计算：输入开始/结束时间，自动计算天数，7种请假类型各有不同说明文字

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 06-考勤管理*
*Context gathered: 2026-04-18*
