---
name: onboarding-redirect-fails
description: 企业信息录入后没有跳转到主页面
trigger: 企业信息录入后，没有跳转到主页面
status: fixed
created: 2026-04-19
updated: 2026-04-19
next_action: "human verify"
root_cause: "HomeView.vue imports UserAdd from @element-plus/icons-vue, which does not export this icon name. This causes a Vite module resolution error preventing HomeView from loading, blocking the router.push('/home') call from OrgSetup.vue."
fix: "Removed UserAdd from import, replaced with Plus icon for 新入职 shortcut"
verification: "Reload Vite dev server, test onboarding flow end-to-end"
files_changed: ["frontend/src/views/home/HomeView.vue"]
---

## Symptoms

1. **Expected behavior**: 企业信息录入提交后，跳转到主页
2. **Actual behavior**: 没有跳转，页面停留
3. **Error messages**:
   - Frontend: `SyntaxError: The requested module '/node_modules/.vite/deps/@element-plus_icons-vue.js?v=9c6d1f65' does not provide an export named 'UserAdd' (at HomeView.vue:142:3)`
   - Backend (2nd call): `ERROR: duplicate key value violates unique constraint "idx_org_credit_code" (SQLSTATE 23505)`
4. **Timeline**: 2026-04-19 13:30-13:31，两次调用，第一次成功（200），第二次报错（重复）
5. **Reproduction**: 调用 PUT /api/v1/org/onboarding 完成企业信息录入

## Evidence

- 第一次调用（13:30:16）：user_id=1, org_id=0, 返回 200，6ms
- 第二次调用（13:31:42）：user_id=1, org_id=2, 返回 200，然后 DB 报错重复键
- HomeView.vue:142 引用了不存在的 `UserAdd` 图标组件
- 前端错误阻止了页面跳转逻辑执行

## Current Focus

**hypothesis**: HomeView.vue 导入的 UserAdd 图标在 @element-plus/icons-vue 中不存在，导致 JS 解析错误，阻止了后续的路由跳转逻辑执行

**next_action**: 检查 HomeView.vue 第 142 行附近的 UserAdd 导入，验证 @element-plus/icons-vue 中该图标名称