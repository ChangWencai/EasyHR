# Phase 4: 个税计算 - Context

**Gathered:** 2026-04-07
**Status:** Ready for planning

<domain>
## Phase Boundary

基于工资/收入数据自动匹配个税专项附加扣除项（子女教育、继续教育、大病医疗、住房贷款利息、住房租金、赡养老人、3岁以下婴幼儿照护），按中国累计预扣预缴法精准计算个税（正确处理税率跳档），个税申报截止前3天自动提醒，生成申报表供手动提交至自然人电子税务局。申报明细可查询、可导出凭证。覆盖 TAX-01 ~ TAX-06 全部需求。

</domain>

<decisions>
## Implementation Decisions

### 个税税率表存储
- **D-01:** 税率表存储于数据库（TaxBracket表），OrgID=0 全局共享（国家统一税率政策，与社保政策库模式一致）。
- **D-02:** 税率表包含七级超额累进税率：应纳税所得额区间、税率（3%/10%/20%/25%/30%/35%/45%）、速算扣除数。管理员可通过H5后台更新（年度政策调整时）。
- **D-03:** 起征点（基本减除费用）按月5000元（年度60000元）预置，存储于配置表或税率表关联字段，支持政策调整时更新。

### 专项附加扣除管理
- **D-04:** 支持全部7项专项附加扣除：子女教育、继续教育、大病医疗、住房贷款利息、住房租金、赡养老人、3岁以下婴幼儿照护。每项扣除存储扣除类型、扣除标准金额、生效月份、失效月份。
- **D-05:** V1.0 由老板在管理后台为员工逐个录入专项附加扣除项。老板是核心用户，操作简洁。每员工可有多项扣除，每项独立管理（新增/修改/失效）。
- **D-06:** 扣除标准预置于系统（如子女教育每孩每月2000元、赡养老人每月3000元等），老板只需选择扣除类型和适用人数/条件，系统自动计算月度扣除金额。不需要员工自行填写。
- **D-07:** 专项附加扣除按月生效。支持指定生效月份和失效月份，方便年中变更。计算当月个税时取当月生效的扣除项汇总。

### 工资数据获取与循环依赖解耦
- **D-08:** Tax模块暴露计算接口 `CalculateTax(orgID, employeeID, month, grossIncome float64)` — 工资数据（税前收入）作为参数传入。Tax模块不依赖Phase 5工资核算模块，彻底解耦循环依赖。
- **D-09:** Tax模块依赖Phase 2（Employee模块）获取员工基本信息。通过接口注入：定义 `EmployeeProvider` 接口（获取员工在职状态、入职月份等），由 Employee 模块的 adapter 实现（与社保模块的 EmployeeQuerier 模式一致）。
- **D-10:** Tax模块依赖Phase 3（社保模块）获取个人社保扣款金额。通过接口注入：定义 `SocialInsuranceProvider` 接口（获取指定月份个人社保扣款总额），由社保模块的 adapter 实现。
- **D-11:** Phase 5（工资核算模块）将作为调用方：先计算税前工资 → 调用 Tax 模块获取个税金额 → 计算实发工资。Tax 模块是独立服务提供者，不反向依赖工资模块。
- **D-12:** Tax模块同时提供独立查询接口：老板可按月查看个税计算明细（应税收入、各项扣除、适用税率、应扣税额），不依赖工资模块调用。此时税前收入来源为员工当前生效合同的Salary字段（Contract.Salary）。

### 累计预扣预缴计算
- **D-13:** 采用累计预扣预缴法：当月应扣税额 = (累计应纳税所得额 × 适用税率 - 速算扣除数) - 累计已扣税额。其中累计应纳税所得额 = 累计收入 - 累计基本减除 - 累计专项扣除（社保个人部分）- 累计专项附加扣除。
- **D-14:** 每月计算时从 TaxRecord 历史记录实时累加 YTD（Year-to-Date）数据。不使用额外快照表。TaxRecord 每月每个员工一条记录，包含当月收入、各项扣除、税率、应扣税额、累计数据。
- **D-15:** 正确处理税率跳档：当累计应纳税所得额跨过税率区间边界时，使用对应的新税率级距计算。无需特殊处理，算法本身基于累计值查税率表即可自动处理。
- **D-16:** 年初（1月）累计数据自动清零重置。TaxRecord 包含 year 字段用于区分年度。跨年查询仍可追溯历史。

### 个税申报提醒
- **D-17:** 复用 gocron v2.19.1 定时任务模式，与社保模块提醒机制一致。每日定时检查个税申报截止日期，到期前3天生成提醒。
- **D-18:** 个税申报截止日为每月15日（法定截止日，遇节假日顺延信息不自动处理，V1.0 使用固定15日）。提醒方式：APP内消息 + 首页待办卡片（Phase 7 消费）。不使用短信。
- **D-19:** 提醒内容包含：待申报月份、涉及员工数、预估个税总额。老板点击提醒可跳转到申报页面。

### 申报表与凭证
- **D-20:** 生成 Excel 格式申报表为主（方便老板手动录入到自然人电子税务局），使用 excelize 库。申报表包含：员工姓名、身份证号（脱敏）、累计收入、累计扣除、应纳税额、已扣税额、本次应申报税额。
- **D-21:** 同时支持导出 PDF 格式个税凭证（正式打印存档），使用 go-pdf/fpdf 库。
- **D-22:** 个税申报状态管理：未申报 → 已申报。老板在系统标记"已申报"后更新状态，不对接税务局接口（V1.0 Out of Scope）。

### 数据模型设计
- **D-23:** 核心模型：
  - `TaxBracket`（税率表，OrgID=0，全局共享）
  - `SpecialDeduction`（专项附加扣除，按员工+类型，OrgID隔离）
  - `TaxRecord`（月度个税计算记录，按员工+月份，OrgID隔离）
  - `TaxDeclaration`（申报记录，按月份+企业，OrgID隔离）
- **D-24:** TaxRecord 存储完整计算快照：当月收入、基本减除、社保个人扣款、专项附加扣除总额、适用税率、速算扣除数、当月应扣税额、累计已扣税额、年度累计数据。确保历史可追溯。
- **D-25:** 错误码范围：40xxx（个税模块）。

### RBAC 权限（个税模块）
- **D-26:** 个税模块权限分配：
  - OWNER：全部操作（查看/计算个税、管理专项附加扣除、导出申报表/凭证、标记申报状态）
  - ADMIN：同 OWNER（查看/计算、管理扣除、导出、标记申报）
  - MEMBER：仅查看自己个税记录（通过 user_id → employee_id 关联）

### Claude's Discretion
- 个税模块内部目录结构（可拆分为 tax/ 主包 + calculator.go 计算引擎）
- 税率表 JSONB 还是关系表存储（推荐关系表，结构固定）
- TaxRecord 累计字段的具体设计（是否冗余存储年度累计值）
- 申报表 Excel 模板的具体格式和列定义
- PDF 凭证模板的排版细节
- gocron 调度器与社保模块共享还是独立实例

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### 项目规范
- `prd.md` — 产品需求文档，V1.0功能范围、验收标准
- `ui-ux.md` — UI/UX设计原型
- `tech-architecture.md` — 技术架构设计、数据模型、模块结构、API设计

### Phase 1-3 上下文
- `.planning/phases/01-foundation-auth/1-CONTEXT.md` — Phase 1 全部决策（三层架构、多租户、加密、RBAC 等）
- `.planning/phases/02-employee-management/02-CONTEXT.md` — Phase 2 决策，关键：Contract.Salary 字段、离职事件接口
- `.planning/phases/03-social-insurance/03-CONTEXT.md` — Phase 3 决策，关键：D-12 GetSocialInsuranceDeduction 接口、EmployeeQuerier 接口模式

### 研究报告
- `.planning/research/STACK.md` — 技术栈选型（gocron v2.19.1, excelize, go-pdf/fpdf 等）
- `.planning/research/ARCHITECTURE.md` — 架构设计建议、模块边界
- `.planning/research/PITFALLS.md` — 常见陷阱

### 需求追踪
- `.planning/REQUIREMENTS.md` — TAX-01 ~ TAX-06 需求定义
- `.planning/ROADMAP.md` — Phase 4 定义和成功标准

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/common/model/base.go` — BaseModel（ID、OrgID、审计字段、软删除），Tax所有模型嵌入此基类
- `internal/common/response/` — 统一响应封装（Success、Error、PageSuccess）
- `internal/common/middleware/` — 认证、RBAC（RequireRole）、多租户（TenantScope）
- `internal/employee/contract_model.go` — Contract.Salary (decimal(10,2))，个税模块获取员工基本工资的数据源
- `internal/socialinsurance/dto.go` — DeductionResponse 结构体，个税模块获取社保个人扣款的接口
- `internal/socialinsurance/scheduler.go` — gocron 定时任务模式参考实现
- `internal/socialinsurance/employee_adapter.go` — EmployeeQuerier 接口 adapter 模式参考实现
- `internal/employee/excel.go` / `internal/socialinsurance/excel.go` — Excel 导出参考实现
- `internal/employee/pdf.go` — PDF 生成参考实现
- `test/testutil/` — 测试工具（SQLite 内存测试、CreateTestOrg、CreateTestUser、CreateTestEmployee）

### Established Patterns
- 三层架构：handler → service → repository，个税模块遵循相同模式
- 跨模块解耦：定义接口 + adapter 实现（参考 socialinsurance/employee_adapter.go）
- DTO 模式：独立请求/响应结构体，binding tag 做参数校验
- 路由注册：RegisterRoutes 方法在 main.go 统一注册
- 审计日志：GORM Hook 自动记录，Module="tax"
- 多租户：Repository 层自动注入 org_id（税率表除外，OrgID=0）
- 定时提醒：gocron v2 + Reminder 模型，与社保模块一致

### Integration Points
- `cmd/server/main.go` AutoMigrate — 新增 TaxBracket、SpecialDeduction、TaxRecord、TaxDeclaration 模型
- `cmd/server/main.go` 路由注册 — 新增 taxHandler.RegisterRoutes(v1, authMiddleware)
- `cmd/server/main.go` 依赖注入 — taxRepo → taxSvc（注入 EmployeeProvider + SocialInsuranceProvider）→ taxHandler
- `cmd/server/main.go` gocron 注册 — 新增个税申报提醒定时任务
- Phase 5 调用 — taxSvc.CalculateTax(orgID, employeeID, month, grossIncome)
- Phase 7 首页待办 — 消费 TaxDeclaration 提醒数据

### 新增依赖
- 无 — excelize、go-pdf/fpdf、gocron 均已在前面 Phase 引入

</code_context>

<specifics>
## Specific Ideas

- 个税计算引擎应设计为纯函数式（输入=收入+扣除项，输出=税额），方便单元测试覆盖各种边界情况
- 税率跳档是核心难点，需充分测试跨档场景（如月收入5000→15000时的税率变化）
- 累计预扣预缴法意味着1月和12月的税额差异可能很大，UI展示需注意
- 专项附加扣除标准预置于系统，老板只需选类型+填人数，降低操作成本
- 申报表Excel格式应与自然人电子税务局的批量导入格式对齐，减少老板手动录入工作量

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---
*Phase: 04-tax-calculation*
*Context gathered: 2026-04-07*
