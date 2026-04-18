# Phase 7: 薪资管理增强 - Context

**Gathered:** 2026-04-18
**Status:** Ready for planning

<domain>
## Phase Boundary

管理员获得完整的薪资数据洞察，调薪/普调按部门或个人灵活操作，薪资算法集成考勤数据自动核算。

具体包含：
- 薪资数据看板（应发/实发/社保公积金/个税总额 + 环比上月百分比）
- 调薪/普调（员工调薪/部门普调，金额或比例，生效期限自动应用）
- 个税 Excel 上传（自动抓取字段，更新当月工资表）
- 绩效系数设置（0%-100%，独立月度表）
- 薪资算法增强（考勤数据集成：计薪天数/病假系数/加班费）
- 发工资条（全员/选定员工）+ Excel 导出

**Scope:** 仅 H5 管理后台，后端 API 配合新增。
**Depends on:** Phase 5（Department 模型用于普调），Phase 6（考勤数据驱动薪资计算）

</domain>

<decisions>
## Implementation Decisions

### 薪资数据看板
- **D-SAL-DASH-01:** 4张纯数字卡片（应发总额/实发总额/社保公积金/个税总额），纯数字卡片风格，与员工数据看板/考勤月报统计行一致
- **D-SAL-DASH-02:** 环比仅展示已确认月份（confirmed/paid）的数据；下月无数据时显示"—"

### 调薪/普调 INSERT ONLY 策略
- **D-SAL-ADJ-01:** INSERT ONLY + 月份只读保护。新调薪记录 INSERT 新行（禁止 UPDATE 历史）；当月 draft 工资表自动重新核算；confirmed/paid 月份的 PayrollRecord 保持不变。调薪历史通过 SalaryItem.effective_month 控制自然生效
- **D-SAL-ADJ-02:** 绩效系数使用独立月度表：performance_coefficients 表（employee_id + year_month + coefficient），默认值 1.0（100%），在 CalculatePayroll 时动态读取并与绩效工资项相乘

### 个税 Excel 上传
- **D-SAL-TAX-01:** 员工姓名精确匹配为主（精确匹配 → 模糊匹配 → 跳过），无法匹配的行记录在错误日志中并提示"3行无法匹配"
- **D-SAL-TAX-02:** 部分成功提示策略：部分匹配成功则成功，列出未匹配行；全部失败则整体失败并显示失败原因
- **D-SAL-TAX-03:** 上传成功后自动更新当月工资表个税字段（PayrollItem 或 PayrollRecord.tax），并标记工资表状态回 draft（需重新核算）

### 薪资算法增强：考勤联动
- **D-SAL-ATT-01:** 计薪天数 = 实际出勤 + 法定节假日 + 带薪假天数；基本工资 = 基本工资项/应出勤 × 计薪天数（按应出勤比例计算）
- **D-SAL-ATT-02:** 病假工资 = 基本工资 × 病假系数，系数存于 sick_leave_policies 表（工龄档位 × 城市 × 系数），初期仅支持北上广深；不得低于当地最低工资 80%
- **D-SAL-ATT-03:** 加班费按法定系数计算（工作日 150%/双休日 200%/节假日 300%），Phase 6 已决定加班时长按 0.5h 取整，存储精确到 0.01h
- **D-SAL-ATT-04:** AttendanceProvider.GetMonthlyAttendance() 返回：actual_days（实际出勤）、should_attend（应出勤）、overtime_hours（加班时长）、paid_leave_days（带薪假天数）、legal_holiday_days（法定节假日天数）、sick_leave_days（病假天数）

### 工资条发送
- **D-SAL-SLIP-01:** 发送通道优先级：微信小程序通知（wx.request）→ 短信（阿里云 SMS，小程序未绑定时降级）→ H5 工资条链接（PayrollSlip.Token，最简降级）
- **D-SAL-SLIP-02:** 向全员发送当月工资条时，使用后台批量发送（asynq 队列），避免超时

### 历史数据保护
- **D-SAL-DATA-01:** draft/calculated 状态可重新编辑和重算；confirmed/paid 禁止任何字段修改，需管理员输入解锁码（企业主手机验证码）才能临时解锁，解锁后重新核算并需再次确认
- **D-SAL-DATA-02:** 解锁操作必须记录审计日志（解锁人/时间/原因），不可逆

### Claude's Discretion
- 薪资数据看板卡片的排列顺序和具体样式细节
- 普调按部门选择的具体 UI（多选部门？全选按钮？）
- 个税 Excel 列名识别算法（支持哪些别名映射）
- 计薪天数中"法定节假日"的来源（AttendanceRule.Holidays JSON 还是独立 holidays 表）
- 工资条 H5 页面具体样式和内容结构
- sick_leave_policies 表的初始数据（北上广深各档位系数）

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements
- `.planning/REQUIREMENTS.md` §3 — Phase 7 需求定义（SAL-01~SAL-19）
- `.planning/ROADMAP.md` §Phase 7 — 阶段目标、成功标准、依赖关系
- `.planning/REQUIREMENTS.md` §2（ATT）— 考勤数据联动需求（SAL-13~SAL-16）

### Research
- `.planning/research/SUMMARY.md` — v1.3 研究综合报告（关键陷阱：月份只读保护、病假系数配置）
- `.planning/research/FEATURES.md` §2.3 — 请假类型与薪资影响（病假/年假/调休的计算规则）
- `.planning/research/ARCHITECTURE.md` — v1.3 架构设计（adapter 接口扩展）

### Existing Code Patterns
- `internal/salary/model.go` — PayrollRecord/PayrollItem/SalaryItem 模型（已有 gross_income/net_income/tax/si_deduction 字段）
- `internal/salary/calculator.go` — 现有 CalculatePayroll 函数骨架，需扩展考勤联动逻辑
- `internal/salary/slip.go` — PayrollSlip + Token 模式，可复用发送逻辑
- `internal/salary/excel.go` — excelize Excel 导出，个税上传复用此模式
- `internal/attendance/adapter.go` — AttendanceProvider 接口，薪资模块依赖此接口读取考勤数据
- `internal/dashboard/service.go` — Dashboard 聚合模式（errgroup 并发），薪资看板复用
- `frontend/src/views/tool/SalaryTool.vue` — 现有薪资工具页面，需扩展新 Tab

### Project Decisions
- `.planning/PROJECT.md` — Key Decisions 表（多租户 org_id、RBAC、加密策略）
- `.planning/STATE.md` — 项目状态（陷阱标注：月份强制只读保护、病假系数按城市配置）
- `05-CONTEXT.md` — Phase 05 决策（Department 模型、花名册增强），复用 Department 用于普调
- `06-CONTEXT.md` — Phase 06 决策（加班 0.5h 取整、AttendanceMonthly 预计算、AttendanceProvider 接口）

### Go Module Dependency
- `shopspring/decimal`（已安装）— 薪资精确计算
- `excelize`（已安装）— Excel 读写
- `asynq`（已安装）— 工资条批量发送队列

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/salary/calculator.go`: 现有 CalculatePayroll 纯函数骨架，扩展考勤联动时在 PayrollItemInput 中加入 attendance 相关薪资项
- `internal/salary/slip.go`: PayrollSlip.Token 已存在，扩展 sendSlip 时复用 Token 验证逻辑
- `internal/attendance/adapter.go`: AttendanceProvider 接口，GetMonthlyAttendance(employeeID, yearMonth) 供 salary 模块调用
- `internal/dashboard/service.go`: errgroup 并发聚合模式，薪资看板复用此模式同时查询 4 个指标

### Established Patterns
- Handler → Service → Repository 三层架构（所有模块统一）
- org_id 逻辑多租户隔离，GORM Scope 自动注入
- Adapter 接口模式：salary 是 consumer，attendance/si/tax 是 provider
- INSERT ONLY 策略：调薪历史不 UPDATE，通过 effective_month 自然生效

### Integration Points
- attendance → salary: AttendanceProvider.GetMonthlyAttendance() 读取出勤数据
- department → salary: 普调按部门维度选择员工，Department 模型已建好（Phase 05）
- tax → salary: TaxProvider.UploadTaxFile() 更新当月个税（复用现有 TaxProvider 接口扩展）
- employee → salary: EmployeeProvider.GetEmployees() 普调时获取部门员工列表
- asynq → salary: 工资条批量发送使用 asynq 队列（SlipSendTask）

### Critical Anti-Patterns to Avoid
- ❌ UPDATE 历史 PayrollRecord → confirmed/paid 月份禁止修改，只能 INSERT 新调薪记录
- ❌ 重新核算 confirmed/paid 月份 → 必须先解锁，解锁后重算并重新确认
- ❌ 浮点计算薪资 → 必须使用 shopspring/decimal，所有金额计算

</code_context>

<specifics>
## Specific Ideas

- 薪资数据看板与员工数据看板风格保持一致（蓝色主题、卡片圆角 12px）
- 普调按部门：部门多选后预览影响人数和预估金额，再确认提交
- 工资条 H5 页面：展示员工姓名、月份、分项工资明细（应发合计/扣除合计/实发），可截图保存
- 计薪天数中法定节假日天数从 AttendanceRule.Holidays JSON 读取

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 07-薪资管理增强*
*Context gathered: 2026-04-18*
