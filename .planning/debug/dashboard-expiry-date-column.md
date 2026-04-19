---
status: fixed
trigger: "GET /api/v1/dashboard 返回 500，错误为 column \"expiry_date\" does not exist (SQLSTATE 42703)，位于 internal/dashboard/repository.go:182"
created: 2026-04-11T00:00:00+08:00
updated: 2026-04-11T00:00:00+08:00
---

## Current Focus
hypothesis: "repository.go 中引用的 `expiry_date` 列名与数据库实际列名不匹配"
test: "读取 repository.go 和 contract_model.go，定位问题"
expecting: "找到正确的列名（contract_model.go 映射为 end_date）"
next_action: "修复并验证编译"
---

## Symptoms
expected: "Dashboard 返回合同即将到期的统计数据（未来30天内到期的合同数量）"
actual: "500 Internal Server Error，\"failed to get dashboard: get contract expirations: ERROR: column \\\"expiry_date\\\" does not exist\""
errors: "SQLSTATE 42703, column \"expiry_date\" does not exist"
reproduction: "登录后访问 dashboard 页面（自动触发 /api/v1/dashboard）"
timeline: "数据库 schema 可能与 Go model 定义不一致"

## Evidence
- timestamp: 2026-04-11
  checked: "internal/dashboard/repository.go:180 和 internal/employee/contract_model.go:17"
  found: "repository.go WHERE 子句使用 `expiry_date` 列，但 contract_model.go 定义的字段为 `EndDate`，GORM tag 为 `column:end_date`"
  implication: "SQL 查询中的列名 `expiry_date` 与数据库实际列名 `end_date` 不匹配，导致 42703 错误"

## Resolution
root_cause: "repository.go 中硬编码了错误的列名 `expiry_date`，而 Contract model 实际的数据库列名是 `end_date`"
fix: "将 repository.go 中的 `expiry_date` 改为 `end_date`（两处：WHERE 子句和 ContractRecord 结构体字段名）"
verification: "go build ./internal/dashboard/ 和 go vet ./internal/dashboard/ 均通过"
files_changed:
  - "internal/dashboard/repository.go"
