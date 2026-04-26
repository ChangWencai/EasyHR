---
status: complete
date: 2026-04-26
quick_id: 260426-001
---

# Quick Task: 薪资管理和社保管理提升为一级菜单

## Summary

将"薪资管理"和"社保管理"从"人事工具"二级菜单提升为独立的一级菜单，移除"人事工具"菜单分组，重新调整整体侧边栏菜单结构。

## Changes

### Menu Structure (Before → After)

Before:
- 人事工具 (二级菜单)
  - 工具概览、薪资管理、邮箱模板、社保管理、个税申报

After:
- 薪资管理 (一级菜单) — 子项：薪资概览、工资条发放、个税申报
- 社保管理 (一级菜单) — 子项：社保概览

### Files Modified

1. **frontend/src/views/layout/AppLayout.vue**
   - Desktop sidebar: replaced 人事工具 with 薪资管理 + 社保管理
   - Mobile drawer: synced with desktop
   - Updated pageTitleMap with new paths
   - Updated icon imports (removed Tools, added Wallet + Umbrella)

2. **frontend/src/router/index.ts**
   - Added new route groups: /salary/*, /social-insurance
   - Old /tool/* routes redirect to new paths
   - Updated auth guard to include /salary and /social-insurance
   - Preserved /tool/email-templates as standalone route

3. **frontend/src/views/home/HomeView.vue**
   - Updated grid items and route map paths

4. **frontend/src/views/employee/OffboardingList.vue**
   - Updated social insurance path reference

5. **frontend/src/views/tool/ToolOverview.vue**
   - Updated tool card paths

6. **frontend/src/views/tool/ToolHome.vue**
   - Updated sub-menu paths

## New Route Structure

| Path | Component |
|------|-----------|
| /salary | SalaryTool.vue |
| /salary/dashboard | SalaryDashboard.vue |
| /salary/slip-send | SalarySlipSend.vue |
| /salary/tax-upload | TaxUpload.vue |
| /salary/tax | TaxTool.vue |
| /social-insurance | SITool.vue |
