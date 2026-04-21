---
name: mine-edit-org-city-dropdown
description: 我的页面-编辑企业信息中城市输入框接入/api/v1/cities下拉接口
status: resolved
trigger: "前端界面 我的-编辑企业信息中的城市输入框要接入/api/v1/cities接口"
created: 2026-04-21
updated: 2026-04-21
resolved: 2026-04-21
symptoms:
  expected: "城市下拉选择，调用 /api/v1/cities 接口，只保存名称"
  actual: "当前使用 el-input 文本输入框，未接入任何接口，保存城市名称"
  impact: "编辑企业信息页面（MineView.vue）"
  current: "使用 el-input，保存城市名称"
---

## Current Focus

**Hypothesis:** MineView.vue 中城市字段使用 el-input 文本输入，未接入 /api/v1/cities 接口

**Next action:** ROOT CAUSE FOUND — 修复完成

---

## Evidence

- timestamp: 2026-04-21T00:00:00Z
  source: MineView.vue line 227-228
  detail: "城市字段使用 el-input 文本输入框 `<el-input v-model="editOrgForm.city">`，未接入任何 API"

---

## Eliminated

- 后端 /api/v1/cities 接口正常（调试会话触发器说明接口已存在）
- 保存逻辑正常，只是前端 UI 没用下拉

---

## Root Cause

**Root Cause:** MineView.vue 编辑企业信息弹窗中，城市字段使用 `el-input` 文本输入框（第227-228行），未接入后端 `/api/v1/cities` 接口。

**Fix:** 将 `el-input` 替换为 `el-select`（filterable + remote），增加 `fetchCityOptions` 方法调用 cities API 搜索城市，下拉列表选择后只保存城市名称。

---

## Fix

**Changes to frontend/src/views/mine/MineView.vue:**

1. 将 `el-input v-model="editOrgForm.city"` 替换为 `el-select`（filterable + remote + reserve-keyword）
2. 增加 `cityOptions` ref 存储城市下拉列表，`cityLoading` 控制加载状态
3. 增加 `fetchCityOptions` 方法（防抖 300ms），调用 `GET /cities?search=query` 接口，映射响应字段（code/name）
4. 表单验证规则保持 required，不额外限制（用户可从下拉选或手动输入）

---

## Verification

- 打开编辑企业信息弹窗，城市字段为下拉选择器（可搜索）
- 输入"北京"，下拉显示北京市等选项
- 选择城市后保存，企业信息卡片城市字段正确显示
- 未选择（输入空白）提交时触发"请输入所在城市"验证错误

---

## Files Changed

- frontend/src/views/mine/MineView.vue