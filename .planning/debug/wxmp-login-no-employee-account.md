---
name: wxmp-login-no-employee-account
description: WeChat MP login fails with "手机号未关联员工账号" despite employee existing in roster
type: bug
status: investigating
trigger: 员工列表有数据，但是使用微信登录员工账号提示"该手机号未关联员工账号"
created: 2026-04-24
updated: 2026-04-24

## Symptoms

1. **Expected behavior**: Employee with phone 150****5678 should be able to log in via WeChat MP
2. **Actual behavior**: POST /api/v1/wxmp/auth/login → 500 "该手机号未关联员工账号"
3. **Roster API returns**: Employee id=2, name="员工2", phone="150****5678", status="pending"
4. **Key observation**: Roster shows employee with phone but wxmp login fails — possible mismatch between which field/table is used for phone lookup

## Current Focus

**Root Cause Identified**: Employee creation does NOT set `UserID` field. wxmp login queries `employees` by `user_id` (via `employees.user_id = users.id`), but `user_id` is always NULL since employees are created without binding a user account.

**Evidence Chain**:
1. `employees` table has `UserID *int64` field with comment "关联用户ID（绑定账号后填写）" — meant to be set AFTER account binding
2. `employee.CreateEmployee` (service.go:46-179) creates employee WITHOUT setting `UserID`
3. `employee.RegistrationService.SubmitRegistration` creates employee WITHOUT setting `UserID`
4. `user.CreateSubAccount` (service.go:300-319) creates `User` record but does NOT update the corresponding `Employee.UserID`
5. wxmp `GetMemberByPhone` queries `users` table first (by phone_hash), then joins to `employees` by `employees.user_id = users.id` — but `user_id` is NULL

**Next action**: Verify the fix approach — modify `GetMemberByPhone` to query employees directly by phone_hash instead of joining via user_id.

## Evidence

- timestamp: 2026-04-24
  checked: "internal/wxmp/repository.go GetMemberByPhone"
  found: "Queries `users` table by phone_hash, then joins `employees` by `employees.user_id = users.id` WHERE user_id matches"
  implication: "If Employee.UserID is never set, this query will always return no employee"

- timestamp: 2026-04-24
  checked: "internal/employee/model.go Employee struct"
  found: "UserID field exists with comment '关联用户ID（绑定账号后填写）' — explicitly meant to be set AFTER account binding"
  implication: "Employee creation intentionally does not set UserID; it's set later during account binding"

- timestamp: 2026-04-24
  checked: "internal/employee/service.go CreateEmployee"
  found: "Creates Employee record without setting UserID field"
  implication: "All employees created via admin panel have UserID=NULL"

- timestamp: 2026-04-24
  checked: "internal/employee/registration_service.go SubmitRegistration"
  found: "Creates Employee record without setting UserID field"
  implication: "All employees created via registration link have UserID=NULL"

- timestamp: 2026-04-24
  checked: "internal/user/service.go CreateSubAccount"
  found: "Creates User record with phone_hash, does NOT update Employee.UserID"
  implication: "Sub-account creation does not link the User to the existing Employee"

- timestamp: 2026-04-24
  checked: "internal/employee/repository.go ListRoster"
  found: "Roster queries `employees` table directly by `phone_hash` field"
  implication: "Roster works because it queries employees directly; wxmp login fails because it queries employees by user_id"

## Eliminated

## Root Cause

**Root Cause**: `WXMPRepositoryImpl.GetMemberByPhone` 在 `employees` 表中通过 `user_id` 字段关联查找员工，而不是通过 `phone_hash` 直接查找。当管理员通过后台创建员工时，`Employee.UserID` 字段从未被设置（该字段设计为"绑定账号后填写"），导致 wxmp 登录时 JOIN 查询找不到对应的员工记录。

**Data Flow**:
1. 管理员后台创建员工 → `employees` 表记录 `phone_hash` 有值，`user_id` 为 NULL
2. 管理员创建子账户（User） → `users` 表记录 `phone_hash` 与员工一致，但 `employees.user_id` 未被更新
3. 员工 wxmp 登录 → `GetMemberByPhone` 查询 `users` 表（找到用户），然后 JOIN `employees` WHERE `user_id = ?`（找不到，因为 user_id 是 NULL）

**Fix**: 修改 `GetMemberByPhone` 方法，直接通过 `phone_hash` 在 `employees` 表中查找员工，移除通过 `user_id` 的关联逻辑。同时需要确保 `employees.user_id` 在创建子账户时被正确设置。

## Fix

**Option A (推荐)**: 修改 `GetMemberByPhone`，直接通过 `phone_hash` 在 `employees` 表中查找员工（不依赖 `user_id` 关联）

**Option B**: 在创建子账户时，同时更新 `employees.user_id` 字段，将 User 和 Employee 关联起来

## Verification

(tbd)

## Files Changed

(tbd)
