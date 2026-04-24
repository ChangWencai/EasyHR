---
name: employee-user-id-null
description: Employee creation does not populate employees.user_id field
type: investigation
status: resolved
trigger: 创建员工信息，表employees的user_id为空，检查创建员工流程
created: 2026-04-24
updated: 2026-04-24

## Root Cause

**`employees.user_id` 字段从未被任何代码路径填充**，是未实现的设计残留。

字段注释：`"关联用户ID（绑定账号后填写）"` — "fill after account binding"。但绑定步骤从未实现。

## Evidence

- `internal/employee/service.go` `CreateEmployee`：创建员工，`user_id` 从未设置
- `internal/employee/registration_service.go` `SubmitRegistration`：创建/更新员工，`user_id` 从未设置
- `internal/user/service.go` `CreateSubAccount`：创建子账户，不回写 `employee.user_id`
- 无任何代码路径写入 `employees.user_id`

## Conclusion

`user_id` 是设计残留字段，**WXMP 登录已通过 `phone_hash` 关联机制绕过此问题**（见 `wxmp-login-no-employee-account.md` 调试会话）。不需要为 `user_id` 字段实现绑定逻辑。
