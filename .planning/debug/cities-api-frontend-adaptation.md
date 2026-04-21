---
name: cities-api-frontend-adaptation
description: 前端城市选择器适配后端 /api/v1/cities 新返回的 area_code 表结构
status: verifying
trigger: "/api/v1/cities接口现在返回的是area_code表结构数据，前端没有适配"
created: 2026-04-21
updated: 2026-04-21
symptoms:
  expected: 前端期望城市列表数据（如 id、name 等字段）
  actual: "后端返回了 area_code 表结构：{code, name, level, pcode, category, created_at}"
  impact: "城市选择器"
  backend_response:
    code: 0
    data:
      - code: 110000000000
        name: "北京市"
        level: 1
        pcode: 0
        category: 0
        created_at: "0001-01-01T00:00:00Z"
---

## Current Focus

**Root Cause Confirmed:** 前端城市选择器组件（OrgSetup.vue）使用旧的字段名 `id`，而后端 city 模块重构后返回的是 `AreaCode` 结构，其主键字段是 `code` 而非 `id`。

**Fix Applied:** 修改了 4 处代码，将所有 `city.id` 改为 `city.code`

---

## Evidence

- timestamp: 2026-04-21
  checked: frontend/src/views/onboarding/OrgSetup.vue
  found: |
    cityList 类型为 { id: number; name: string }[]
    el-option 使用 city.id 作为 key 和 value
    loadCities 调用 /cities 接口，将 res.data 直接赋值给 cityList
  implication: 前端期望 id/name 结构，但后端返回 code/name 结构

- timestamp: 2026-04-21
  checked: internal/city/area_code.go
  found: AreaCode 结构体使用 code 作为主键字段
  implication: 后端已重构为 area_code 表结构

- timestamp: 2026-04-21
  checked: vite.config.ts proxy 配置
  found: /api 代理到 localhost:8089，前端 /cities 映射到后端 /api/v1/cities
  implication: 确认前端调用的就是后端新的 cities 接口

---

## Eliminated

---

## Root Cause

前端城市选择器组件（OrgSetup.vue）使用旧的字段名 `id`，而后端 city 模块重构后返回的是 `AreaCode` 结构，其主键字段是 `code` 而非 `id`。

## Fix

修改 `frontend/src/views/onboarding/OrgSetup.vue`:
1. 将 `cityList` 类型从 `{ id: number; name: string }[]` 改为 `{ code: number; name: string }[]`
2. 将模板中的 `city.id` 改为 `city.code`（el-option 的 key 和 value）
3. 将 `detectCity` 函数中的 `matched.id` 改为 `matched.code`
4. 将 `handleSubmit` 函数中查找城市的 `c.id` 改为 `c.code`

## Verification

- [ ] 启动后端服务
- [ ] 启动前端服务
- [ ] 访问 /org-setup 页面
- [ ] 验证城市下拉列表正确显示省份数据

## Files Changed
- frontend/src/views/onboarding/OrgSetup.vue
