---
status: awaiting_human_verify
trigger: "Dashboard 接口报错 'column payment_month does not exist'"
created: 2026-04-11T00:00:00Z
updated: 2026-04-11T00:10:00Z
---

## Current Focus

hypothesis: "social_insurance_records 表中缺少 `payment_month` 字段，导致 dashboard 查询失败"
test: "检查 internal/socialinsurance/model.go 中 SocialInsuranceRecord 模型定义，以及 internal/dashboard/repository.go 中的 SQL 查询"
expecting: "如果模型中有 payment_month 字段但查询中用了其他名称，或者相反，即可确认根因"
next_action: "等待用户验证 /api/v1/dashboard 是否正常返回"

## Symptoms

expected: Dashboard 接口返回正常数据
actual: 返回 500 错误，"failed to get dashboard: get social insurance total: ERROR: column \"payment_month\" does not exist (SQLSTATE 42703)"
errors: SQLSTATE 42703 - 列不存在
reproduction: 访问 /api/v1/dashboard
started: 2026-04-11

## Eliminated

- hypothesis: 数据库表缺少 payment_month 列
  evidence: 数据库表确实有 start_month 字段，不是表的问题，而是 repository.go 中使用的字段名与实际不匹配
  timestamp: 2026-04-11T00:05:00Z

## Evidence

- timestamp: 2026-04-11T00:03:00Z
  checked: internal/socialinsurance/model.go - SocialInsuranceRecord 模型定义
  found: 模型字段为 `StartMonth string` (映射到 `start_month` 列)，以及 `EndMonth *string` (映射到 `end_month` 列)，没有 `payment_month` 字段
  implication: 实际表结构是 start_month/end_month，不是 payment_month

- timestamp: 2026-04-11T00:04:00Z
  checked: internal/dashboard/repository.go - SIRecord 结构和查询
  found: SIRecord 定义了 `PaymentMonth string` 字段，查询用 `Where("payment_month = ?", paymentMonth)`
  implication: Repository 使用了不存在的列名，与实际模型不匹配

- timestamp: 2026-04-11T00:08:00Z
  checked: 编译验证
  found: go build 编译通过，无语法错误
  implication: 代码修改正确，可以部署

## Resolution

root_cause: "Dashboard repository.go 中的 SIRecord 结构体和查询使用了 `payment_month` 字段，但实际的 `social_insurance_records` 表中该字段名为 `start_month`（参保起始月份）"
fix: "修改 SIRecord 结构体将 `PaymentMonth` 改为 `StartMonth`，查询条件从 `payment_month` 改为 `start_month`"
verification: "编译通过，go build 无错误"
files_changed: ["internal/dashboard/repository.go"]
