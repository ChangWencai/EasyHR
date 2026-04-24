---
name: pending-sign-shows-no-contract
description: Employee list shows "无合同" despite pending_sign contract existing
type: bug
status: resolved
trigger: 有合同，合同是待签署状态，但是在员工列表中的合同到期状态显示是无合同
created: 2026-04-24
updated: 2026-04-24

## Symptoms

1. **Expected behavior**: Employee list should show contract status (e.g. "待签署" or the contract's expiry info) when a contract exists, even if it's in pending_sign status
2. **Actual behavior**: Employee list shows "无合同" (no contract) even though the employee has a contract in pending_sign status
3. **Reproduction**: Create a contract in pending_sign status for an employee, then check the employee list — contract status column shows "无合同"

## Current Focus

**FIXED**: Added `'pending_sign'` to the `status IN` filter in `GetContractExpiryDays` SQL query.

## Evidence

- timestamp: 2026-04-24
  checked: "internal/employee/repository.go" `GetContractExpiryDays` function (lines 360-408)
  found: "SQL query uses `WHERE ... AND status IN ('active', 'signed')` — explicitly excludes `pending_sign` contracts"
  implication: "Employees with `pending_sign` contracts will not appear in the contract expiry map, causing frontend to show '无合同'"

- timestamp: 2026-04-24
  checked: "internal/employee/contract_model.go" ContractStatus constants (lines 35-42)
  found: "Contract lifecycle: draft -> pending_sign -> signed -> active -> terminated/expired. pending_sign contracts have valid end_date and should be displayed"
  implication: "pending_sign contracts have valid end_date and should be included in roster contract expiry display"

- timestamp: 2026-04-24
  checked: "frontend/src/views/employee/EmployeeList.vue" contract column (lines 136-158)
  found: "Frontend correctly checks `contract_expiry_days !== null && contract_expiry_days !== undefined`. If map has no entry, renders '无合同'"
  implication: "Frontend logic is correct — it reflects whatever the backend returns. Bug is entirely in backend query"

- timestamp: 2026-04-24
  checked: "internal/employee/service.go" `ListRoster` and `ExportExcel` functions
  found: "Both call `GetContractExpiryDays` — fix in repository method fixes both roster display and Excel export"

## Eliminated

- Hypothesis: Frontend checks wrong field for contract status
  evidence: "Frontend correctly checks `contract_expiry_days`. When backend returns undefined, it shows '无合同' as expected."
- Hypothesis: Employee list API (ListEmployees) does not return contract info
  evidence: "The employee list page uses `getRoster` API (ListRoster service). Contract info comes from GetContractExpiryDays repository method."

## Root Cause

`GetContractExpiryDays` in `internal/employee/repository.go` line 386 uses SQL filter `status IN ('active', 'signed')` which explicitly excludes `pending_sign` contracts. When an employee's only contract is `pending_sign`, `contractMap` has no entry for that employee. The frontend receives `contract_expiry_days = undefined` and renders "无合同".

## Fix

In `internal/employee/repository.go`, added `'pending_sign'` to the `status IN` clause in `GetContractExpiryDays` SQL query (two places: subquery and outer query).

Before: `AND status IN ('active', 'signed')`
After:  `AND status IN ('active', 'signed', 'pending_sign')`

## Verification

Fix applied. No frontend changes needed — the frontend was already correctly rendering the `contract_expiry_days` field. The bug was entirely in the backend query excluding `pending_sign` contracts from the contract expiry map.

## Files Changed

- `internal/employee/repository.go`: Add `'pending_sign'` to status filter in `GetContractExpiryDays` SQL query (lines 386 and 389)
