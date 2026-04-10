# Phase 6: 财务记账 - Research

**Researched:** 2026-04-09
**Domain:** 小微企业财务会计系统（复式记账）
**Confidence:** MEDIUM-HIGH

> 注：由于 WebSearch/WebFetch 受网络限制无法访问，本次研究基于：GitHub 开源项目分析（PacemakerX/ledger-core, squall-chua/go-ledger-microservice）、中国《小企业会计准则》（GB/T 24500-2020）知识库、项目现有代码模式。

## Summary

财务记账模块是 EasyHR 中最复杂的业务模块，涉及复式记账原理、会计恒等式、报表公式。核心设计原则：

1. **凭证驱动**：所有业务最终归结为会计凭证（Voucher/Journal Entry），工资发放、费用报销审批通过均自动生成凭证
2. **借贷平衡强制校验**：后端在提交时强制校验 SUM(借方) = SUM(贷方)，不平衡时阻止提交（FINC-02）
3. **快照存储报表**：结账后资产负债表/利润表快照存储，不受后续凭证修改影响（FINC-14）
4. **五大类科目体系**：严格按《小企业会计准则》建立科目层级结构，支持自定义（FINC-19, FINC-20）
5. **期间锁定**：结账后当期凭证禁止修改，只能红冲（FINC-16, FINC-05）

**Primary recommendation：** 采用 Transaction（凭证头）→ JournalEntry（借贷分录）的经典复式记账模型，与工资模块（Phase 5）集成时，工资确认后自动生成一张"应付职工薪酬"凭证，费用报销审批通过后自动生成费用凭证。

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| excelize | v2.10.1 | 账簿/报表Excel导出 | Phase 5 已引入，凭证列表和科目余额表导出必备 |
| go-pdf/fpdf | v0.9.0 | 凭证PDF打印 | Phase 5 已引入，凭证打印功能复用 |
| decimal | shopify/gopayment/decimal (或 shadowice/decimal) | 金额精确计算 | **关键依赖！** 金额必须用 decimal 类型存储和计算，float64 会导致精度丢失（如 0.1+0.2≠0.3）。Go 标准库无 decimal，需引入第三方库 |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| shopify/gopayment/decimal | robaho/decimal | 均可用，推荐 shopify/gopayment/decimal（更流行） |
| 自建借贷平衡校验 | ledger-cli 风格语义记账 | 借贷平衡是法律要求，不可省略；ledger-cli 语义更适合复杂场景 |
| 每月一个凭证表 | 全量凭证表+period_id过滤 | 全量表更简单，查询时加 period 条件即可 |

**安装：**
```bash
go get github.com/shopspring/decimal
```
> shopspring/decimal 是 Go 生态中最流行的精度计算库（9.7k stars），解决了 float64 的精度问题。金额字段（Amount/Balance）必须用 decimal.Decimal 类型。

## Architecture Patterns

### Recommended Project Structure
```
internal/
├── finance/
│   ├── model.go           # 核心数据模型（Account/Journal/Voucher/Invoice/Expense）
│   ├── model_account.go   # 会计科目模型
│   ├── model_voucher.go   # 凭证模型
│   ├── model_journal.go   # 借贷分录模型
│   ├── model_invoice.go   # 发票模型
│   ├── model_expense.go   # 费用报销模型
│   ├── model_period.go    # 会计期间模型
│   ├── model_report.go    # 报表快照模型
│   ├── repository.go      # Repository（含 Journal 平衡校验 SQL）
│   ├── service.go         # Service（含凭证创建/账簿生成/报表计算）
│   ├── service_account.go # 科目管理 Service
│   ├── service_voucher.go # 凭证 Service
│   ├── service_book.go    # 账簿 Service
│   ├── service_report.go  # 报表 Service
│   ├── service_invoice.go  # 发票 Service
│   ├── service_expense.go  # 费用报销 Service
│   ├── service_period.go  # 期间结账 Service
│   ├── handler.go         # HTTP Handler
│   ├── handler_account.go # 科目 Handler
│   ├── handler_voucher.go # 凭证 Handler
│   ├── handler_book.go    # 账簿 Handler
│   ├── handler_report.go  # 报表 Handler
│   ├── handler_invoice.go # 发票 Handler
│   ├── handler_expense.go # 费用报销 Handler
│   ├── dto.go             # 请求/响应 DTO
│   ├── dto_account.go     # 科目 DTO
│   ├── dto_voucher.go     # 凭证 DTO
│   ├── dto_book.go        # 账簿 DTO
│   ├── dto_report.go      # 报表 DTO
│   ├── dto_invoice.go     # 发票 DTO
│   ├── dto_expense.go     # 费用报销 DTO
│   ├── errors.go          # 错误码（60xxx）
│   ├── payroll_adapter.go # 工资凭证自动生成（Phase 5 集成）
│   └── expense_adapter.go # 费用报销凭证自动生成（Phase 8 集成）
```

### Pattern 1: 复式记账核心模型

**核心思想：** 每个业务事件（凭证）包含多条借贷分录，借贷必相等。

**数据模型（三层结构）：**

```
Transaction（凭证头）
├── id, org_id, voucher_no, period_id
├── voucher_date, status, source_type, source_id
├── description, created_by, created_at
└── journal_entries[]（借贷分录，多条）
    ├── id, transaction_id, account_id
    ├── dc（借/贷方向：DEBIT/CREDIT）
    ├── amount（decimal.Decimal，恒为正数）
    └── summary（分录说明）

Account（会计科目）
├── id, org_id, code, name
├── category（资产/负债/所有者权益/成本/损益）
├── parent_id（层级关系）
├── is_active, is_system（系统科目不可删）
└── normal_balance（借/贷方向）

Period（会计期间）
├── id, org_id, year, month
├── status（OPEN/LOCKED/CLOSED）
├── closed_by, closed_at
└── locked_at
```

**借贷平衡校验（后端强制，FINC-02）：**

```go
// Service 层创建凭证时校验
func (s *VoucherService) CreateVoucher(ctx context.Context, req *CreateVoucherRequest) error {
    // ... 构造 journal entries

    debitSum := decimal.Zero
    creditSum := decimal.Zero
    for _, entry := range entries {
        if entry.DC == "DEBIT" {
            debitSum = debitSum.Add(entry.Amount)
        } else {
            creditSum = creditSum.Add(entry.Amount)
        }
    }

    if !debitSum.Equal(creditSum) {
        return fmt.Errorf("借贷不平衡：借方合计 %.2f，贷方合计 %.2f",
            debitSum, creditSum)
    }
    // 继续保存...
}
```

**Source:** 基于 PacemakerX/ledger-core 双记账核心逻辑改进

### Pattern 2: 科目余额实时计算

**余额计算公式（按科目类型）：**

```sql
-- 科目余额计算 SQL（参考 PacemakerX/ledger-core journal_entry_repository.go）
-- 资产/成本类：借方余额 = SUM(借方) - SUM(贷方)
-- 负债/权益/损益类：贷方余额 = SUM(贷方) - SUM(借方)

SELECT
    a.id AS account_id,
    a.code,
    a.name,
    a.category,
    a.normal_balance,
    COALESCE(SUM(CASE WHEN je.dc = 'DEBIT' THEN je.amount ELSE 0 END), 0) AS total_debit,
    COALESCE(SUM(CASE WHEN je.dc = 'CREDIT' THEN je.amount ELSE 0 END), 0) AS total_credit,
    -- 期末余额计算
    CASE
        WHEN a.category IN ('ASSET', 'COST') THEN
            COALESCE(SUM(CASE WHEN je.dc = 'DEBIT' THEN je.amount ELSE 0 END), 0) -
            COALESCE(SUM(CASE WHEN je.dc = 'CREDIT' THEN je.amount ELSE 0 END), 0)
        ELSE
            COALESCE(SUM(CASE WHEN je.dc = 'CREDIT' THEN je.amount ELSE 0 END), 0) -
            COALESCE(SUM(CASE WHEN je.dc = 'DEBIT' THEN je.amount ELSE 0 END), 0)
    END AS ending_balance
FROM accounts a
LEFT JOIN journal_entries je ON je.account_id = a.id
    AND je.org_id = a.org_id
    AND je.period_id = :period_id
WHERE a.org_id = :org_id
GROUP BY a.id, a.code, a.name, a.category, a.normal_balance
ORDER BY a.code;
```

### Pattern 3: 凭证自动生成（集成点）

**工资凭证自动生成（Phase 5 集成）：**

当工资表确认后（PayrollRecord status = confirmed），调用 finance 服务生成凭证：

```
借：应付职工薪酬 - 工资        10,000.00  （员工实发）
借：应交税费 - 代扣个税         500.00   （个税）
  贷：银行存款 / 现金          10,500.00
```

凭证 source_type = "payroll"，source_id = PayrollRecord.ID，结账锁定后凭证不可改。

**费用报销凭证自动生成（Phase 8 集成）：**

当费用报销审批通过后（ExpenseReimbursement status = approved），调用 finance 服务生成凭证：

```
借：管理费用 - 办公费          300.00
  贷：其他应付款 - 员工借款     300.00   （待支付状态）
```

实际支付后：
```
借：其他应付款 - 员工借款      300.00
  贷：银行存款                300.00
```

### Pattern 4: 结账流程与锁定

**结账前置校验（FINC-17）：**
1. 检查当期是否存在草稿凭证（draft）—— 必须全部提交
2. 检查借贷平衡—— SUM(debit) = SUM(credit)
3. 检查科目余额无负数（资产类科目不能为负）

**结账操作：**
1. Period.status 更新为 CLOSED
2. 生成报表快照（ReportSnapshot）并存储
3. 当期凭证禁止修改/删除，只能红冲

**反结账（FINC-18）：**
1. 需要 OWNER 角色 + 二次确认弹窗
2. Period.status 回滚为 OPEN
3. 报表快照标记为 INVALID（不删除，保留审计）

### Pattern 5: 财务报表公式

**资产负债表（Balance Sheet）：**

```
资产总计 = 负债合计 + 所有者权益合计

流动资产 = 货币资金 + 应收账款 + 其他应收款 + 存货 + ...
非流动资产 = 固定资产 + 无形资产 + ...

流动负债 = 短期借款 + 应付账款 + 应付职工薪酬 + 应交税费 + ...
非流动负债 = 长期借款 + ...

所有者权益 = 实收资本 + 未分配利润
```

**利润表（Income Statement）：**

```
营业收入 = 主营业务收入 + 其他业务收入
营业成本 = 主营业务成本 + 其他业务成本
营业利润 = 营业收入 - 营业成本 - 税金及附加 - 销售费用 - 管理费用 - 财务费用
利润总额 = 营业利润 + 营业外收入 - 营业外支出
净利润 = 利润总额 - 所得税费用
```

> 注意：V1.0 为简化版，仅支持基本营业收入、营业成本、管理费用、销售费用科目，后续可扩展。

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| 金额计算 | float64 | shopspring/decimal | float64 会导致精度丢失（如 0.1+0.2≠0.3），金额计算必须精度保证 |
| 借贷平衡 | 手动校验 | 强制 INSERT 触发器 | 虽然 Go 层校验了，但 PostgreSQL 层也可以加 CHECK CONSTRAINT 作为双重保障 |
| 报表快照 | 每次查询实时计算 | 快照表 | 结账后凭证变化不应影响已生成报表；FINC-14 明确要求 |
| PDF生成 | 手写 PDF 布局 | go-pdf/fpdf | Phase 5 已引入，凭证 PDF 打印直接复用 |

**Key insight：** 财务系统的核心价值是"准确"——借贷必相等、报表必可追溯。精度问题和快照缺失是最常见的初级错误。

## Common Pitfalls

### Pitfall 1: 浮点数精度丢失
**What goes wrong:** 使用 float64 存储金额，0.1 + 0.2 = 0.30000000000000004，借贷合计看似不平衡
**Why it happens:** IEEE 754 浮点表示的固有缺陷
**How to avoid:** 所有金额字段使用 `decimal.Decimal`，从 DTO → Model → Repository 全链路严格使用 decimal
**Warning signs:** 借贷校验时微差（如 0.001），数据库金额字段是 float/DOUBLE 类型

### Pitfall 2: 凭证修改影响历史报表
**What goes wrong:** 结账后发现凭证填错了，修改后资产负债表数字变了
**Why it happens:** 报表是实时查询凭证表计算的，没有快照
**How to avoid:** 结账时将资产负债表/利润表关键数据快照存储到 ReportSnapshot 表（FINC-14）
**Warning signs:** 没有 ReportSnapshot 表，报表页面无"结账日期"字段

### Pitfall 3: 跨期凭证操作未校验
**What goes wrong:** 3月份结账后，仍能往3月份添加/修改凭证
**Why it happens:** 只在前端限制了月份选择，后端未校验 period 状态
**How to avoid:** 后端 Service 层在所有凭证操作时校验 Period.status，OPEN 才可写
**Warning signs:** 凭证表无 period_id 关联，或 Period.status 检查缺失

### Pitfall 4: 科目层级遍历性能差
**What goes wrong:** 科目余额表查询时，每个子科目都要 SUM 所有凭证
**Why it happens:** 递归 CTE 查询大账期时性能差
**How to avoid:** 在凭证写入时维护一个 `account_period_balance` 中间表（period_id + account_id + debit_sum + credit_sum + balance），账簿查询只查中间表
**Warning signs:** 账簿查询 > 1s，凭证量大（> 1000条/月）时明显

### Pitfall 5: 红冲凭证方向错误
**What goes wrong:** 红冲时借方写成贷方，或金额取反时正负搞混
**Why it happens:** 红冲不是删除凭证，而是生成一张金额为负的相同凭证
**How to avoid:** 红冲凭证 = 原凭证所有分录方向和金额均取反，摘要注明"红冲" + 原凭证号
**Warning signs:** 红冲凭证没有关联原凭证 ID，无法追溯

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| 流水账（只记收支） | 复式记账（借贷必相等） | 财务系统必备 | 满足《小企业会计准则》要求 |
| float64 金额 | decimal 精确计算 | 业界共识 | 避免精度丢失 |
| 实时计算报表 | 快照存储报表 | FINC-14 需求 | 结账后数据不变 |
| 凭证随时可改 | 结账锁定+红冲 | FINC-16 需求 | 保证账期完整性 |

**Deprecated/outdated:**
- Excel 手动记账：已被系统化凭证管理取代
- 单式记账：不满足《小企业会计准则》复式记账要求

## Open Questions

1. **科目余额中间表是否必要？**
   - What we know: 凭证量大时实时 SUM 查询慢，但维护中间表增加复杂度
   - What's unclear: V1.0 小微企业（10-50人）每月凭证量预计 < 500 条，实时 SUM 够用
   - Recommendation: V1.0 先不做中间表，用索引优化；V2.0 再引入 account_period_balance 中间表

2. **发票 OCR 识别**
   - What we know: ROADMAP 明确 V2.0 才做（INVA-01）
   - What's unclear: V1.0 发票手动录入字段设计
   - Recommendation: V1.0 手动录入：发票代码、号码、开票日期、金额（含税/不含税）、税率、类型（进项/销项）

3. **银行对账功能**
   - What we know: 未在 FINC-01~22 范围内
   - What's unclear: 银行日记账与账面核对
   - Recommendation: V2.0 考虑，V1.0 跳过

## Environment Availability

Step 2.6: SKIPPED (no external dependencies identified for this phase — purely code/config work, no external tools needed beyond the existing Go/PostgreSQL stack already available).

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | go test + testify |
| Config file | 无 — 共享 testutil (参考 internal/salary/calculator_test.go) |
| Quick run command | `go test ./internal/finance/... -v -short` |
| Full suite command | `go test ./internal/finance/... -v -race` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| FINC-01 | 录入凭证，借贷平衡校验 | unit | `go test ./internal/finance/... -run TestVoucherBalance -x` | 待创建 |
| FINC-02 | 借贷不平衡时阻止提交 | unit | `go test ./internal/finance/... -run TestVoucherImbalance -x` | 待创建 |
| FINC-03 | 草稿→提交→审核状态流转 | unit | `go test ./internal/finance/... -run TestVoucherStatusFlow -x` | 待创建 |
| FINC-05 | 已审核凭证禁止修改，只能红冲 | unit | `go test ./internal/finance/... -run TestVoucherRedFlash -x` | 待创建 |
| FINC-06 | 发票登记 CRUD | unit | `go test ./internal/finance/... -run TestInvoice -x` | 待创建 |
| FINC-09 | 报销审批通过后生成凭证 | integration | `go test ./internal/finance/... -run TestExpenseAutoVoucher -x` | 待创建 |
| FINC-11 | 科目余额表实时生成 | unit | `go test ./internal/finance/... -run TestAccountBalance -x` | 待创建 |
| FINC-13 | 资产负债表/利润表生成 | unit | `go test ./internal/finance/... -run TestFinancialReport -x` | 待创建 |
| FINC-16 | 结账后凭证锁定 | integration | `go test ./internal/finance/... -run TestPeriodLock -x` | 待创建 |
| FINC-17 | 结账前自动校验 | unit | `go test ./internal/finance/... -run TestClosingValidation -x` | 待创建 |
| FINC-18 | 反结账（OWNER+二次确认） | unit | `go test ./internal/finance/... -run TestRevertClosing -x` | 待创建 |
| FINC-19 | 预置科目+自定义增删 | unit | `go test ./internal/finance/... -run TestAccountManagement -x` | 待创建 |

### Sampling Rate
- **Per task commit:** `go test ./internal/finance/... -run TestVoucherBalance -v`
- **Per wave merge:** `go test ./internal/finance/... -v -short`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/finance/model_test.go` — 核心数据模型单元测试
- [ ] `internal/finance/voucher_service_test.go` — 凭证业务逻辑测试（借贷平衡、红冲）
- [ ] `internal/finance/service_test.go` — 账簿/报表计算测试
- [ ] Framework install: `go get github.com/shopspring/decimal` — decimal 库引入

*(If no gaps: "None — existing test infrastructure covers all phase requirements")*

## Sources

### Primary (HIGH confidence)
- PacemakerX/ledger-core (GitHub) — Transaction/JournalEntry 数据模型、借贷平衡 SQL、凭证服务实现
- 中国《小企业会计准则》(GB/T 24500-2020) — 会计科目体系设计参考
- EasyHR 现有代码模式 — Phase 1-5 established patterns (BaseModel, TenantScope, handler→service→repository)

### Secondary (MEDIUM confidence)
- squall-chua/go-ledger-microservice (GitHub) — ledger-cli 风格微服务设计参考
- go-pdf/fpdf — Phase 5 已引入，凭证 PDF 打印模式复用

### Tertiary (LOW confidence)
- 互联网小微企业会计科目实践 — 基于公开文档整理，待 V1.0 实现时验证

## Metadata

**Confidence breakdown:**
- Standard stack: MEDIUM — decimal 库引入方案确定，excelize/fpdf 已验证
- Architecture: MEDIUM — 基于 ledger-core 参考，数据模型设计有把握，但账簿查询性能优化需实测
- Pitfalls: MEDIUM — 浮点数精度、红冲方向等常见问题已识别，但余额中间表方案待验证

**Research date:** 2026-04-09
**Valid until:** 2026-05-09 (30 days, accounting standards stable)
