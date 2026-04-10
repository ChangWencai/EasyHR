# Phase 6: 财务记账 - Context

**Gathered:** 2026-04-09
**Status:** Ready for planning

<domain>
## Phase Boundary

老板可完成小微企业完整财务记账流程：录入凭证、管理发票、费用报销审批、查看账簿报表、月度结账。覆盖 FINC-01 ~ FINC-22 全部需求。依赖 Phase 2（员工管理）和 Phase 5（工资核算）。

</domain>

<decisions>
## Implementation Decisions

### 复式记账核心模型
- **D-01:** 采用 Transaction（凭证头）→ JournalEntry（借贷分录）双层结构。Transaction 包含凭证号（按期间自增）、日期、摘要、来源类型（manual/payroll/expense）、审核状态。JournalEntry 每条包含科目ID、借/贷方向（DEBIT/CREDIT）、金额（decimal.Decimal）、分录摘要。
- **D-02:** 金额全程使用 `decimal.Decimal` 类型（shopspring/decimal）。DTO → Model → Repository 全链路禁止 float64。所有金额字段存为字符串（`varchar`），数据库层防止精度丢失。
- **D-03:** 借贷平衡校验：Service 层创建凭证时强制校验 SUM(debit) = SUM(credit)，不平衡时返回错误码 60201，阻止保存（FINC-02）。同时在 PostgreSQL 层加 CHECK CONSTRAINT：`CHECK (debit_sum = credit_sum)`。

### 凭证状态与操作权限
- **D-04:** 凭证状态流转：draft（草稿）→ submitted（已提交待审核）→ audited（已审核）→ closed（已结账）。已审核凭证禁止修改/删除，只能红冲（FINC-05）。
- **D-05:** 红冲凭证生成：原凭证所有分录方向取反、金额不变、摘要注明"红冲凭证 原凭证号 XXX"。红冲凭证与原凭证通过 `reversal_of` 字段关联。
- **D-06:** 凭证号编码规则：`{YYYYMM}-{序号}`，如 `202604-0001`，每期独立自增。

### 会计科目体系（预置+自定义）
- **D-07:** 预置科目表（基于《小企业会计准则》），分五大类：

  **资产类（1开头，借方余额）：**
  - 1001 库存现金
  - 1002 银行存款
  - 1012 其他货币资金
  - 1122 应收账款
  - 1123 预付账款
  - 1221 其他应收款
  - 1405 原材料
  - 1407 发出商品
  - 1601 固定资产
  - 1602 累计折旧
  - 1701 无形资产
  - 1901 长期待摊费用

  **负债类（2开头，贷方余额）：**
  - 2001 短期借款
  - 2201 应付票据
  - 2202 应付账款
  - 2203 预收账款
  - 2241 其他应付款
  - 2501 长期借款
  - 2801 应付职工薪酬
  - 2221 应交税费

  **所有者权益类（3开头，贷方余额）：**
  - 4001 实收资本
  - 4002 资本公积
  - 4101 盈余公积
  - 4103 本年利润
  - 4104 利润分配

  **成本类（5开头，借方余额）：**
  - 5001 生产成本
  - 5101 制造费用
  - 5301 研发支出

  **损益类（6开头）：**
  - 6001 主营业务收入
  - 6051 其他业务收入
  - 6401 主营业务成本
  - 6402 其他业务成本
  - 6403 税金及附加
  - 6601 销售费用
  - 6602 管理费用
  - 6603 财务费用
  - 6901 营业外收入
  - 6911 营业外支出
  - 6901 所得税费用

- **D-08:** 系统预置科目（is_system=true）不可删除，可禁用（is_active=false）。企业可新增自定义科目，code 允许 8xxxx 段。
- **D-09:** 科目按层级展示：父科目 → 子科目，最多支持 3 级（FINC-20）。

### 账簿生成逻辑
- **D-10:** 实时生成账簿（FINC-11）。科目余额表 = 实时 SUM(period 内 journal_entries)。明细账 = 按科目过滤的凭证列表。V1.0 暂不维护 account_period_balance 中间表（每月凭证量 < 500 时实时 SUM 足够）。
- **D-11:** 账簿查询参数：期间（year/month）、科目ID（可选）、凭证号（可选）。分页展示。
- **D-12:** 账簿导出 Excel：使用 excelize，按列（凭证号、日期、摘要、借方、贷方、余额）导出。

### 财务报表（快照存储）
- **D-13:** 结账时生成报表快照（FINC-14）。ReportSnapshot 表存储：period_id、report_type（balance_sheet/income_statement）、data（JSONB，存储计算结果）、generated_by、generated_at。后续凭证修改不影响已生成快照。
- **D-14:** 资产负债表公式：资产总计 = 负债合计 + 所有者权益合计。V1.0 简化版只支持主要科目合计（货币资金=库存现金+银行存款、应付职工薪酬=2201等）。
- **D-15:** 利润表公式：净利润 = 营业收入 - 营业成本 - 税金及附加 - 销售费用 - 管理费用 - 财务费用 + 营业外收支 - 所得税。
- **D-16:** 报表多期对比（FINC-15）：查询 ReportSnapshot，支持选择2-4期数据并排展示。

### 会计期间管理
- **D-17:** Period 表：year、month、status（OPEN/LOCKED/CLOSED）。V1.0 自动创建未来12个月期间，初始状态 OPEN。结账后状态更新为 CLOSED。
- **D-18:** 结账前置校验（FINC-17）：① 当期无 draft/submitted 状态凭证；② 借贷总额平衡；③ 关键科目余额无负数（资产类、成本类科目余额 >= 0）。
- **D-19:** 反结账（FINC-18）：Period.status 回滚为 OPEN，ReportSnapshot 标记 invalid。必须 OWNER 角色，且前端二次确认弹窗。
- **D-20:** 结账后凭证操作限制：CLOSED 状态的 period 内，audited 凭证禁止修改/删除，只能红冲。

### 发票管理
- **D-21:** Invoice 表：invoice_type（input/output 进项/销项）、code、number、date、amount（含税）、tax_rate、tax_amount、status（未认证/已认证/已抵扣）、remark。关联凭证（可选）。
- **D-22:** 月末增值税自动计算（FINC-07, FINC-21）：销项税额 = SUM(output invoice tax_amount)，进项税额 = SUM(input invoice 已认证 tax_amount)，应纳税额 = 销项 - 进项。生成纳税申报辅助数据，支持导出 Excel。

### 费用报销审批
- **D-23:** 报销单字段：employee_id、amount、expense_type（差旅费/交通费/招待费/办公费/其他）、description、attachments（最多9张凭证照片）、status（pending/approved/rejected/paid）、approver_id、approved_at。
- **D-24:** 报销状态流转（FINC-10）：pending（员工提交）→ approved（老板审批通过）→ paid（已支付）/ rejected（驳回）。审批通过后自动生成费用凭证（FINC-09）。
- **D-25:** 自动生成的报销凭证分录：借：管理费用-XX费（按报销类型），贷：其他应付款-员工借款。实际支付后：借：其他应付款-员工借款，贷：银行存款。
- **D-26:** 报销凭证由 Finance 模块生成，Expense 模块调用 Finance 模块的接口，不直接写 journal_entries。

### 工资凭证集成（Phase 5 集成点）
- **D-27:** 工资确认后（PayrollRecord confirmed），Finance 模块生成工资凭证。凭证来源：source_type="payroll"，source_id=PayrollRecord.ID。
- **D-28:** 工资凭证分录（简化版）：借：管理费用-工资 10000.00，贷：应付职工薪酬-工资 10000.00。注意：V1.0 简化处理，社保和个税不单独走凭证（Phase 5 已有 PayrollItem 明细记录）。
- **D-29:** PayrollRecord 确认时，通过 adapter 调用 FinanceService.CreateVoucher()，传入工资总额和摘要，Finance 服务负责借贷平衡和凭证保存。

### RBAC 权限（财务模块）
- **D-30:** 财务模块权限分配：
  - OWNER：全部操作（科目管理、凭证录入/审核/红冲、发票管理、报销审批、结账/反结账、报表导出）
  - ADMIN：凭证录入/审核、发票管理、报销审批、查看账簿报表
  - MEMBER：无财务模块权限（Phase 8 员工端仅能提交报销）
- **D-31:** Phase 8 员工端：MEMBER 可通过 WXMP-05 提交费用报销单，但不能审核、不能查看账簿。

### 数据模型设计
- **D-32:** 核心模型：
  - `Account`（会计科目，OrgID 隔离，is_system 标识预置科目）
  - `Period`（会计期间，OrgID 隔离）
  - `Voucher`（凭证，OrgID 隔离，含 voucher_no、status、source_type）
  - `JournalEntry`（借贷分录，OrgID 隔离，外键 voucher_id + account_id）
  - `Invoice`（发票，OrgID 隔离，关联 voucher_id 可选）
  - `ExpenseReimbursement`（费用报销，OrgID 隔离，关联 voucher_id 可选）
  - `ReportSnapshot`（报表快照，OrgID 隔离，结账时生成）
- **D-33:** Voucher.JournalEntries 一对多关系，GORM preload 预加载。
- **D-34:** 错误码范围：60xxx（财务模块），子范围：60201=借贷不平衡，60202=凭证已审核禁止修改，60203=期间已结账禁止操作，60204=结账校验失败。

### Claude's Discretion
- 账簿查询的具体 SQL 优化策略（是否引入 account_period_balance 中间表）
- 发票 PDF/凭证打印的 UI 布局细节
- 纳税申报辅助数据的具体格式（Excel 列结构）
- 科目辅助核算维度（V1.0 是否支持项目核算/部门核算）
- 银行存款日记账（V1.0 不支持，V2.0 考虑）
- 财务报表的 UI 多期对比展示方式

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### 项目规范
- `.planning/REQUIREMENTS.md` — FINC-01 ~ FINC-22 需求定义
- `.planning/ROADMAP.md` — Phase 6 定义和成功标准

### 前序 Phase 上下文
- `.planning/phases/01-foundation-auth/1-CONTEXT.md` — Phase 1 全部决策（三层架构、多租户、加密、RBAC 等）
- `.planning/phases/02-employee-management/02-CONTEXT.md` — Phase 2 决策，关键：Employee 模型
- `.planning/phases/05-salary/05-CONTEXT.md` — Phase 5 决策，关键：PayrollRecord 确认后集成点

### 研究报告
- `.planning/phases/06-finance/06-RESEARCH.md` — 本 phase 研究报告
- `.planning/research/STACK.md` — 技术栈选型（decimal 库、excelize 等）

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/common/model/base.go` — BaseModel（ID、OrgID、审计字段、软删除），财务所有模型嵌入此基类
- `internal/common/response/` — 统一响应封装
- `internal/common/middleware/` — 认证、RBAC（RequireRole）、多租户（TenantScope）
- `internal/salary/service.go` — PayrollRecord 确认后调用 FinanceService 的集成点参考
- `internal/salary/repository.go` — 复杂 Repository 查询参考（多表关联、分页）
- `internal/tax/service.go` — TaxCalculator 接口模式，FinanceService 可参考 adapter 模式
- `internal/employee/excel.go` — Excel 导出参考（excelize）
- `test/testutil/` — 测试工具（SQLite 内存测试、CreateTestOrg、CreateTestUser、CreateTestEmployee）

### Established Patterns
- 三层架构：handler → service → repository，财务模块严格遵循
- 跨模块解耦：定义接口 + adapter 实现（finance/payroll_adapter.go）
- DTO 模式：独立请求/响应结构体，binding tag 做参数校验
- decimal 类型：Model 层金额字段用 decimal.Decimal，JSON 序列化用 String()
- 快照存储：结账时写入 ReportSnapshot，凭证修改不更新快照

### Integration Points
- `cmd/server/main.go` AutoMigrate — 新增 Account、Period、Voucher、JournalEntry、Invoice、ExpenseReimbursement、ReportSnapshot 模型
- `cmd/server/main.go` 路由注册 — 新增 financeHandler.RegisterRoutes(v1, authMiddleware)
- `cmd/server/main.go` 依赖注入 — salarySvc → financeSvc（工资确认后调用）
- Phase 5 PayrollRecord 确认时 — 调用 financeSvc.GeneratePayrollVoucher()
- Phase 8 ExpenseReimbursement 审批时 — 调用 financeSvc.GenerateExpenseVoucher()

### 新增依赖
- shopspring/decimal — 金额精确计算（必须引入）

</code_context>

<specifics>
## Specific Ideas

- 凭证录入 UI：左侧科目选择（树形结构），右侧借/贷金额输入，底部实时显示借贷合计和不平衡差额
- 科目选择：支持搜索（输入"银行"显示所有含"银行"的科目），避免小老板不知道科目代码的问题
- 报表多期对比：3列并排展示（本期、上期、上年同期），差异金额和百分比自动计算
- 结账操作：老板点击"结账"后，显示校验结果列表（如"存在3条未提交凭证"），全部通过后才可结账
- 红冲凭证：在凭证列表页，红冲按钮需二次确认，显示"此操作将生成一张借贷方向相反的凭证，是否继续？"

</specifics>

<deferred>
## Deferred Ideas

- 科目辅助核算（项目核算/部门核算）— V2.0 需求
- 银行日记账与账面核对 — V2.0 考虑
- 发票 OCR 识别和税务局查验接口对接 — V2.0 需求（INVA-01, INVA-02）
- 自动化固定资产折旧计算 — V2.0 考虑
- 预算管理 — V2.0+ 考虑
- 现金流量表 — V2.0+ 考虑

</deferred>

---
*Phase: 06-finance*
*Context gathered: 2026-04-09*
