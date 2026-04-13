# Phase 5: 工资核算 - Context

**Gathered:** 2026-04-08
**Status:** Ready for planning

<domain>
## Phase Boundary

老板可自定义薪资结构（预置模板启用/禁用），一键核算月度工资（自动关联社保和个税扣款），支持复制上月和考勤导入快速核算，生成电子工资单通过H5链接推送至员工（短信验证签收），工资条可导出Excel，记录发放状态和方式，异常发放自动提醒。覆盖 PAYR-01 ~ PAYR-09 全部需求。

</domain>

<decisions>
## Implementation Decisions

### 薪资结构设计
- **D-01:** 预置固定薪资项模板，企业启用/禁用。预置项包括：固定收入项（基本工资[必须启用]、绩效工资、岗位补贴、餐补、交通补、通讯补、其他补贴）和固定扣款项（事假扣款、病假扣款、其他扣款）。老板只需勾选启用哪些项，无需自定义薪资项名称。
- **D-02:** 每个员工独立填写各项金额。启用薪资项后，每个员工可填不同金额（如张三绩效3000、李四绩效2000）。工资表中逐行填写。
- **D-03:** 薪资项独立存储（新增 SalaryItem 表），与 Contract.Salary 解耦。Contract.Salary 保持不变（合同层面基本工资），SalaryItem 的基本工资填写时自动从合同同步但不强制。核算时以 SalaryItem 数据为准。

### 工资核算流程
- **D-04:** 表格式编辑（Excel风格）+ 一键核算。工资表展示所有在职员工和各项薪资金额，老板修改变动项后点击"一键核算"，系统自动计算社保扣款、个税、实发工资。
- **D-05:** 核算状态流转：draft（草稿，可编辑）→ calculated（已核算，可修改重算）→ confirmed（已确认，锁定不可修改）。支持反复修改重算，确认后锁定当月工资表。
- **D-06:** 一键核算计算逻辑：税前收入 = sum(收入项)，调用社保接口 GetSocialInsuranceDeduction 获取个人扣款，调用个税接口 CalculateTax(orgID, empID, year, month, grossIncome) 计算个税，实发 = 税前收入 - 社保个人扣款 - 个税 - sum(扣款项)。
- **D-07:** "复制上月工资表"功能：新建当月工资表时自动填入上月各项金额（固定项直接复制，变动项如绩效/事假扣款也复制作为参考），老板只需修改变动项后一键核算。
- **D-08:** 核算确认后调用 socialinsurance.SuggestBaseAdjustment() 检查薪资变动是否需要调整社保基数（Phase 3 已实现此接口）。

### 工资单推送与签收
- **D-09:** V1.0 通过 H5 链接推送工资单给员工。老板确认工资表后可选择发送工资单给指定员工，生成带唯一 token 的 H5 链接，老板通过微信分享或短信发送给员工。
- **D-10:** H5 页面需要短信验证身份（与 Phase 2 邀请链接模式类似：token + 短信验证码）。员工输入手机号 → 收到验证码 → 验证通过后查看工资单明细。
- **D-11:** 员工查看工资单后可点击"确认签收"，系统记录签收状态和签收时间。签收后仍可查看但不可更改签收状态。
- **D-12:** Phase 8 微信小程序上线后复用同一套工资单数据（PayrollSlip 模型），小程序端直接查看签收，H5 链接仍作为备选。

### 考勤Excel导入
- **D-13:** 固定模板导入，仅支持事假天数和病假天数。模板字段：员工姓名、事假（天）、病假（天）、备注。导入后自动计算扣款金额并填入工资表。
- **D-14:** 日薪计算公式：日薪 = 基本工资 / 21.75（中国劳动法标准月工作日）。事假扣款 = 日薪 × 事假天数，病假扣款 = 日薪 × 病假天数 × 病假扣薪比例（V1.0 按100%扣，简化处理）。
- **D-15:** 导入结果自动填入工资表的事假扣款和病假扣款项。如果工资表中已有手动填写的值，导入时覆盖。

### 工资发放管理
- **D-16:** 工资表确认后进入"待发放"状态。老板记录发放信息：发放方式（银行转账/现金/其他）、发放日期、发放备注。状态：待发放 → 已发放。
- **D-17:** 异常发放自动提醒（PAYR-09）：当月实发工资与上月偏差超过30%的员工标记为"异常"，在确认工资表时高亮提醒老板注意。不使用定时任务，在确认操作时实时检查。

### RBAC 权限（工资模块）
- **D-18:** 工资模块权限分配：
  - OWNER：全部操作（薪资结构配置、核算、确认、发放、导出、推送工资单）
  - ADMIN：同 OWNER（核算、确认、发放、导出、推送）
  - MEMBER：仅查看自己工资单（通过 H5 链接或 Phase 8 小程序，不走管理后台 RBAC）

### 数据模型设计
- **D-19:** 核心模型：
  - `SalaryTemplate`（薪资项模板，OrgID 隔离，记录企业启用了哪些薪资项）
  - `SalaryItem`（员工薪资项金额，每员工每项一条，OrgID 隔离）
  - `PayrollRecord`（月度工资核算主表，每员工每月一条，OrgID 隔离，含状态/总金额/发放信息）
  - `PayrollItem`（工资核算明细，每员工多条，记录各项名称/类型/金额）
  - `PayrollSlip`（工资单，含唯一 token 用于 H5 查看链接，含签收状态/时间）
- **D-20:** PayrollRecord 存储完整核算快照：税前收入、社保个人扣款、个税、各项扣款、实发工资。确认后不可修改（快照性质），确保历史数据可追溯。
- **D-21:** 错误码范围：50xxx（工资模块）。

### Claude's Discretion
- 工资模块内部目录结构（可拆分为 salary/ 主包）
- PayrollItem 是冗余存储薪资项还是只存计算结果
- 工资表 API 设计（批量 vs 逐员工）
- H5 工资单页面的具体实现方式（是否复用邀请链接模式的 token 机制）
- 考勤 Excel 导入的匹配策略（按姓名匹配的容错处理）
- 异常发放提醒的具体阈值和展示方式
- 日薪计算中的小数处理规则（截断 vs 四舍五入）
- PayrollSlip token 的有效期设计

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### 项目规范
- `.planning/REQUIREMENTS.md` — PAYR-01 ~ PAYR-09 需求定义
- `.planning/ROADMAP.md` — Phase 5 定义和成功标准

### 前序 Phase 上下文
- `.planning/phases/01-foundation-auth/1-CONTEXT.md` — Phase 1 全部决策（三层架构、多租户、加密、RBAC 等）
- `.planning/phases/02-employee-management/02-CONTEXT.md` — Phase 2 决策，关键：Contract.Salary 字段、Employee 模型
- `.planning/phases/03-social-insurance/03-CONTEXT.md` — Phase 3 决策，关键：D-12 GetSocialInsuranceDeduction 接口、SuggestBaseAdjustment
- `.planning/phases/04-tax-calculation/04-CONTEXT.md` — Phase 4 决策，关键：D-08 TaxCalculator 接口（接受 grossIncome 参数）、D-11 Phase 5 调用方式

### 研究报告
- `.planning/research/STACK.md` — 技术栈选型（excelize, go-pdf/fpdf, gocron 等）
- `.planning/research/ARCHITECTURE.md` — 架构设计建议、模块边界
- `.planning/research/PITFALLS.md` — 常见陷阱

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/common/model/base.go` — BaseModel（ID、OrgID、审计字段、软删除），工资所有模型嵌入此基类
- `internal/common/response/` — 统一响应封装（Success、Error、PageSuccess）
- `internal/common/middleware/` — 认证、RBAC（RequireRole）、多租户（TenantScope）
- `internal/employee/contract_model.go` — Contract.Salary (decimal(10,2))，工资模块获取基本工资的数据源之一
- `internal/tax/calculator.go` — TaxCalculator 接口：CalculateTax(orgID, empID, year, month, grossIncome float64)，工资模块调用此接口计算个税
- `internal/tax/dto.go` — TaxResult 结构体，包含 MonthlyTax/CumulativeIncome 等
- `internal/tax/model.go` — TaxRecord.Source 字段已预留 "salary" 值
- `internal/tax/employee_adapter.go` — EmployeeAdapter 模式参考（GetActiveSalary 从 Contract.Salary 获取）
- `internal/socialinsurance/service.go` — GetSocialInsuranceDeduction(orgID, empID, month) 返回社保个人扣款
- `internal/socialinsurance/service.go:656` — SuggestBaseAdjustment() 薪资变动后检查社保基数
- `internal/employee/excel.go` — Excel 导出参考实现（excelize）
- `internal/employee/pdf.go` — PDF 生成参考实现（go-pdf/fpdf）
- `internal/employee/invitation_service.go` — 邀请链接 token 模式参考（H5 工资单链接可复用此模式）
- `test/testutil/` — 测试工具（SQLite 内存测试、CreateTestOrg、CreateTestUser、CreateTestEmployee）

### Established Patterns
- 三层架构：handler → service → repository，工资模块遵循相同模式
- 跨模块解耦：定义接口 + adapter 实现（参考 tax/employee_adapter.go）
- DTO 模式：独立请求/响应结构体，binding tag 做参数校验
- 路由注册：RegisterRoutes 方法在 main.go 统一注册
- 审计日志：GORM Hook 自动记录，Module="salary"
- 多租户：Repository 层自动注入 org_id
- 定时提醒：gocron v2 + Reminder 模型（本模块仅异常提醒，不用定时任务）
- H5 外部链接：token + 短信验证模式（参考 employee/invitation）

### Integration Points
- `cmd/server/main.go` AutoMigrate — 新增 SalaryTemplate、SalaryItem、PayrollRecord、PayrollItem、PayrollSlip 模型
- `cmd/server/main.go` 路由注册 — 新增 salaryHandler.RegisterRoutes(v1, authMiddleware)
- `cmd/server/main.go` 依赖注入 — salaryRepo → salarySvc（注入 TaxCalculator + SocialInsuranceDeduction 接口）→ salaryHandler
- `internal/tax/calculator.go` — 调用 taxSvc.CalculateTax() 计算个税
- `internal/socialinsurance/service.go` — 调用 siSvc.GetSocialInsuranceDeduction() 获取社保扣款
- `internal/socialinsurance/service.go:656` — 确认后调用 SuggestBaseAdjustment() 检查社保基数
- Phase 7 首页待办 — 消费异常发放提醒数据
- Phase 8 微信小程序 — 复用 PayrollSlip 数据

### 新增依赖
- 无 — excelize、go-pdf/fpdf、gocron、resty（短信验证码）均已在前面 Phase 引入

</code_context>

<specifics>
## Specific Ideas

- 工资表界面应类似 Excel 表格，行=员工，列=薪资项，最后一列显示实发工资，一目了然
- "复制上月"功能是小老板最常用的操作（大部分月份薪资变动不大），应作为新建工资表的默认选项
- 日薪公式用 21.75（中国劳动法标准月计薪天数），不使用 30 天
- H5 工资单链接复用邀请链接的 token 模式，保持架构一致性
- 确认工资表时高亮显示异常员工（实发偏差 > 30%），不弹窗阻断，只做视觉提醒

</specifics>

<deferred>
## Deferred Ideas

- 完整考勤管理（打卡、加班、调休等）— V2.0 需求（ATTN-01~03）
- 工资自动发放（对接银行API）— V2.0+ 考虑
- 工资条微信模板消息推送 — Phase 8 微信小程序上线后实现
- 薪资变动历史趋势图 — Phase 7 首页数据概览考虑
- 批量银行转账文件导出 — 可作为后续增强

</deferred>

---
*Phase: 05-salary*
*Context gathered: 2026-04-08*
