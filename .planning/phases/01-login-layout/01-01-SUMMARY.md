---
phase: 01-login-layout
plan: 01
subsystem: frontend
tags: [ui, login, design-system, css-token]
dependency_graph:
  requires: []
  provides:
    - css-design-tokens
    - login-left-right-layout
  affects:
    - frontend/src/views/layout/LoginView.vue
    - frontend/src/styles/variables.scss
    - frontend/src/styles/global.scss
tech_stack:
  added:
    - CSS Custom Properties (--primary, --bg-sidebar, --text-sidebar, --radius-lg)
    - Element Plus color override (--el-color-primary: #4F6EF7)
  patterns:
    - Design Token 体系（CSS 自定义属性）
    - 左右分栏 Grid 布局
    - 移动端响应式断点
key_files:
  created: []
  modified:
    - frontend/src/styles/variables.scss
    - frontend/src/styles/global.scss
    - frontend/src/views/layout/LoginView.vue
decisions:
  - "Element Plus 主色从 #1677ff 迁移到 #4F6EF7（统一商务蓝品牌色）"
  - "reset 样式移入 variables.scss，global.scss 仅保留 @import"
  - "LoginView.vue template/style 完全重写，script 逻辑零改动"
  - "剩余 #1677ff 残留（App.vue/HomeView.vue/ToolHome.vue/FinanceHome.vue）留待后续计划"
metrics:
  duration_seconds: 128
  completed_date: "2026-04-14"
  tasks_completed: 3
  files_modified: 3
  commits: 2
---

# Phase 01 Plan 01 Summary: 登录页左右分栏 + CSS 设计 Token 体系

## 一句话

建立 CSS 设计 Token 体系（--primary, --bg-sidebar 等），将登录页从单卡片重构为左侧 720px 渐变品牌区 + 右侧 440px 白色表单卡。

## 做了什么

### Task 1+2: 设计 Token 体系（commit: 0f797b1）

**frontend/src/styles/variables.scss** — 完全重写，新增：
- CSS 自定义属性（--primary, --primary-hover, --primary-dark, --primary-light）
- 背景色（--bg-page, --bg-sidebar, --bg-surface）
- 边框色（--border）
- 文字色（--text-primary, --text-secondary, --text-tertiary, --text-sidebar, --text-sidebar-active）
- 状态色（--danger, --success, --warning）
- 圆角（--radius-lg: 12px, --radius-md: 8px）
- 间距（--spacing-lg: 24px, --spacing-md: 16px）
- Element Plus 颜色覆盖（--el-color-primary: #4F6EF7 替代 #1677ff）

**frontend/src/styles/global.scss** — 简化为仅 `@import './variables.scss'`（reset 样式移入 variables.scss）

**frontend/src/main.ts** — 无需修改（`import '@/styles/global.scss'` 已存在）

### Task 3: 登录页左右分栏（commit: 1995a3e）

**frontend/src/views/layout/LoginView.vue** — 完全重写 template + style，script 逻辑零改动：
- 左侧品牌区 `<aside class="login-brand">`（720px，渐变背景 #1A2D6B → #4F6EF7 → #7B9FFF）：
  - Logo + "易人事" 品牌名 + 标语
  - 5个功能特性列表（入职管理、薪资核算、社保公积金、财务记账、员工工资条）
  - 装饰圆伪元素（::before / ::after）
  - 底部版权文字
- 右侧表单区 `<main class="login-form-panel">`（白色背景，居中显示）：
  - 品牌标题 "易人事"（primary 色）+ "老板管理后台" 副标题
  - 3 个 el-tab-pane（手机验证码 / 密码登录 / 注册），逻辑完全保留
  - 60s 倒计时功能保持
- 移动端响应式（≤768px）：品牌区隐藏，表单占满屏幕，顶部 48px padding
- 动画入场：slideUp 0.4s ease

## 验收清单

- [x] 登录页左右分栏布局正常渲染
- [x] 渐变背景 linear-gradient(135deg, #1A2D6B 0%, #4F6EF7 60%, #7B9FFF 100%) 正确渲染
- [x] 登录 API 调用（login, send-code, register, login/password）保持不变
- [x] 60s 倒计时功能正常
- [x] 移动端响应式正确（品牌区隐藏，表单占满屏幕）
- [x] LoginView.vue 无 #1677ff 硬编码残留

## 已知残留（Deferred）

| 文件 | 残留数量 | 说明 |
|------|---------|------|
| frontend/src/views/home/HomeView.vue | 4 处 #1677ff | 首页 UI，将在 01-02 AppLayout 计划中统一处理 |
| frontend/src/views/tool/ToolHome.vue | 3 处 #1677ff | 工具首页，将在 01-02 计划中统一处理 |
| frontend/src/views/finance/FinanceHome.vue | 3 处 #1677ff | 财务首页，将在 01-02 计划中统一处理 |
| frontend/src/App.vue | 2 处 #1677ff | App 根组件，将在 01-02 计划中统一处理 |

> Plan 01-01 scope 仅覆盖 LoginView.vue、variables.scss、global.scss。剩余 #1677ff 硬编码由后续计划（01-02 AppLayout）统一迁移到 CSS Token。

## Commits

| Hash | 描述 |
|------|------|
| 0f797b1 | feat(01-login-layout): 建立 CSS 设计 Token 体系 |
| 1995a3e | feat(01-login-layout): 重构登录页为左右分栏布局 |
