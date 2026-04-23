---
name: 260423-6ump0-admin-employee-form
description: 在新增员工页面添加管理员手动输入所有员工信息的功能
type: quick
status: complete
created: 2026-04-23
completed: 2026-04-23
---

# Quick Task Summary: 管理员手动录入完整员工信息

## Completed

- 在新增员工的 Step 0 中增加了完整信息录入字段：
  - 试用期薪资（probation_salary）
  - 工资卡号（bank_card）
  - 紧急联系人姓名（emergency_contact）
  - 紧急联系人电话（emergency_phone）

- 简化了 Step 1，移除中间确认状态，直接在完成时创建员工并发送邀请

- 移除了未使用的 `employeeCreated` 变量

- 创建成功后自动跳转到员工列表页

## Files Changed

- `frontend/src/views/employee/EmployeeCreate.vue`

## Verification

- npm run build 编译通过
