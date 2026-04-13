# Phase 01: 登录页 + 布局基础 - Research

**Researched:** 2026-04-14
**Domain:** Vue 3 + Element Plus UI 重构（CSS 布局、CSS 变量体系、暗色侧边栏）
**Confidence:** HIGH（基于已验证的源代码 + Element Plus 官方文档）

## Summary

Phase 1 的 UI 重构核心是建立统一的设计 Token 体系，然后用它驱动两个页面：登录页（左右分栏）和 AppLayout（暗色侧边栏）。技术路径很清晰：在 `variables.scss` 中用 CSS 自定义属性（`--primary` 等）建立设计变量，然后在 `global.scss` 中覆盖 Element Plus 的 `--el-color-primary`，最后在各组件的 `<style lang="scss">` 中直接消费这些变量。Element Plus 2.13.6 支持 CSS 变量覆盖，无需引入 SCSS 源码包。

**Primary recommendation:** CSS Grid + CSS 自定义属性体系，一次性建立变量，后端组件直接引用。

---

## User Constraints (from REQUIREMENTS.md)

### Locked Decisions
- 登录页左面板 720px 宽，渐变 `#1A2D6B → #4F6EF7 → #7B9FFF`
- 登录页右面板白色表单卡 440px 宽
- 侧边栏宽 220px，折叠后 64px，背景 `#0D1B2A`
- 主色调 `#4F6EF7`（从 `#1677ff` 迁移）
- 版权信息 `© 2025 易人事 · 专为小微企业打造`
- 保持现有登录逻辑（request.ts API 调用）不变
- 现有路由不变

### Out of Scope
- 后端 API 改造
- 路由结构调整
- 性能优化
- 单元测试（UI 重构不涉及业务逻辑）

---

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| UI-01 | 登录页重构为左右分栏布局 | CSS Grid 左右分栏 + CSS 变量体系 |
| UI-13 | AppLayout 侧边栏暗色化 | 自定义 SCSS 覆盖 el-menu + CSS 变量 |

---

## Standard Stack

### Core
| Item | Version | Purpose |
|------|---------|---------|
| Vue 3 | 3.5.32 | 前端框架（已有） |
| Element Plus | 2.13.6 | UI 组件库（已有） |
| Vite | 8.0.3 | 构建工具（已有） |
| Sass | 1.99.0 | CSS 预处理器（已有） |

### Design Token 体系（本次新增）
| File | 职责 | 来源 |
|------|------|------|
| `src/styles/variables.scss` | 定义所有 CSS 自定义属性（`--primary` 等） | 本次新建 |
| `src/styles/global.scss` | Element Plus CSS 变量覆盖 + 全局重置样式 | 已有，本次扩展 |
| 组件 `<style lang="scss">` | 直接使用 `var(--primary)` 等变量 | 各组件修改 |

**Installation:** 无需安装新包，全部利用现有依赖（Sass 已存在）。

---

## Architecture Patterns

### 1. CSS 自定义属性设计 Token（推荐方式）

在 `variables.scss` 中用 `:root` 定义全局 CSS 自定义属性，组件内通过 `var(--xxx)` 引用：

```scss
// src/styles/variables.scss
:root {
  // Primary
  --primary: #4F6EF7;
  --primary-hover: #6B84F9;
  --primary-dark: #3651D9;
  --primary-light: #EEF1FF;

  // Background
  --bg-page: #F0F2F5;
  --bg-sidebar: #0D1B2A;
  --bg-sidebar-active: #4F6EF7;
  --bg-sidebar-hover: #1A2D42;
  --bg-surface: #FFFFFF;

  // Border
  --border: #E8ECF0;

  // Text
  --text-primary: #172B4D;
  --text-secondary: #5E6C84;
  --text-tertiary: #97A0AF;
  --text-sidebar: #CDD3E0;
  --text-sidebar-active: #FFFFFF;

  // Status
  --danger: #FF5630;
  --success: #36B37E;
  --warning: #FFAB00;

  // Radius
  --radius-lg: 12px;
  --radius-md: 8px;

  // Spacing
  --spacing-lg: 24px;
  --spacing-md: 16px;
}
```

然后在 `global.scss` 中覆盖 Element Plus 的 CSS 变量：

```scss
// src/styles/global.scss（在已有 :root 块后追加）
:root {
  // Element Plus primary override
  --el-color-primary: var(--primary);
  --el-color-primary-light-3: #6B84F9;
  --el-color-primary-light-5: #8FA4FB;
  --el-color-primary-light-7: #B3C3FD;
  --el-color-primary-light-8: #C7D3FE;
  --el-color-primary-light-9: #EEF1FF;
  --el-color-primary-dark-2: #3651D9;
  --el-color-success: var(--success);
  --el-color-danger: var(--danger);
  --el-color-warning: var(--warning);
  --el-font-size-base: 14px;

  // Border
  --el-border-color: var(--border);
  --el-border-color-light: #E8ECF0;
  --el-border-radius-base: var(--radius-md);

  // Background
  --el-bg-color-page: var(--bg-page);
  --el-bg-color: var(--bg-surface);
}

*,
*::before,
*::after {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

html,
body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background: var(--bg-page);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
```

### 2. 登录页左右分栏（CSS Grid）

```vue
<!-- LoginView.vue -->
<template>
  <div class="login-layout">
    <!-- 左面板：品牌区 -->
    <aside class="login-brand">
      <!-- Logo + Slogan + Feature list -->
    </aside>

    <!-- 右面板：表单卡 -->
    <main class="login-form-panel">
      <div class="login-card">
        <!-- 表单内容 -->
      </div>
    </main>
  </div>
</template>

<style scoped lang="scss">
.login-layout {
  display: grid;
  grid-template-columns: 720px 1fr; // 左固定720px，右自适应
  min-height: 100vh;

  @media (max-width: 768px) {
    grid-template-columns: 1fr; // 移动端单列
  }
}

.login-brand {
  background: linear-gradient(
    135deg,
    #1A2D6B 0%,
    #4F6EF7 60%,
    #7B9FFF 100%
  );
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 60px 48px;
  position: relative;
  overflow: hidden;

  // 左下角装饰圆
  &::before {
    content: '';
    position: absolute;
    bottom: -80px;
    left: -80px;
    width: 300px;
    height: 300px;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.06);
  }

  // 右上角装饰圆
  &::after {
    content: '';
    position: absolute;
    top: -60px;
    right: -60px;
    width: 200px;
    height: 200px;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.04);
  }

  @media (max-width: 768px) {
    display: none; // 移动端隐藏品牌区
  }
}

.login-form-panel {
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-surface);
  padding: 24px;

  @media (max-width: 768px) {
    min-height: 100vh;
    align-items: flex-start;
    padding-top: 48px;
  }
}

.login-card {
  width: 100%;
  max-width: 440px;
  padding: var(--spacing-lg);
  background: var(--bg-surface);
}
</style>
```

**注意:** 移动端（`< 768px`）品牌区隐藏，表单面板占满屏幕，内容从顶部开始。

### 3. 暗色侧边栏（el-menu 自定义 SCSS）

```scss
// AppLayout.vue — 新增 .sidebar-dark 样式块

// 主容器
.app-layout {
  display: flex;
  min-height: 100vh;
  background: var(--bg-page);
}

// 暗色侧边栏
.sidebar {
  width: 220px;
  min-width: 220px;
  height: 100vh;
  position: fixed;
  left: 0;
  top: 0;
  background: var(--bg-sidebar); // #0D1B2A
  transition: width 0.2s ease, min-width 0.2s ease;
  z-index: 200;
  overflow: hidden;

  &.collapsed {
    width: 64px;
    min-width: 64px;
  }
}

// Logo 区
.sidebar-logo {
  height: 56px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  flex-shrink: 0;

  .logo-icon {
    width: 32px;
    height: 32px;
    background: var(--primary);
    border-radius: var(--radius-md);
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    flex-shrink: 0;
  }

  .logo-text {
    font-size: 16px;
    font-weight: 700;
    color: #fff; // 白色
    white-space: nowrap;
    overflow: hidden;
  }
}

// Element Plus 菜单暗色覆盖
.sidebar-el-menu {
  border-right: none !important;
  background: transparent !important;

  // 菜单项
  .el-menu-item,
  .el-sub-menu__title {
    height: 44px;
    line-height: 44px;
    font-size: 14px;
    color: var(--text-sidebar) !important; // #CDD3E0
    background: transparent !important;

    .el-icon {
      color: var(--text-sidebar) !important;
    }

    &:hover {
      background: var(--bg-sidebar-hover) !important; // #1A2D42
      color: #fff !important;
    }
  }

  // 激活态
  .el-menu-item.is-active {
    background: var(--bg-sidebar-active) !important; // #4F6EF7
    color: var(--text-sidebar-active) !important; // #FFFFFF
    border-right: none !important;

    .el-icon {
      color: #fff !important;
    }
  }

  // 子菜单标题激活态
  .is-active > .el-sub-menu__title {
    color: var(--primary) !important;

    .el-icon {
      color: var(--primary) !important;
    }
  }

  // 子菜单箭头
  .el-sub-menu .el-icon {
    color: var(--text-sidebar) !important;
  }

  // 子菜单弹出面板
  .el-menu--inline {
    background: rgba(0, 0, 0, 0.15) !important;
  }
}

// 折叠按钮
.sidebar-footer {
  flex-shrink: 0;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  padding: 8px 8px 8px 4px;
  display: flex;
  justify-content: flex-end;

  .collapse-btn {
    color: rgba(255, 255, 255, 0.5);
    padding: 8px;

    &:hover {
      color: #fff;
      background: var(--bg-sidebar-hover);
    }
  }
}
```

### 4. 移动端抽屉（已有 el-drawer，保持现状微调）

```scss
// el-drawer 内部菜单也应用暗色（覆盖 scoped 样式限制）
// 在 AppLayout.vue 中添加：
:deep(.el-drawer) {
  background: var(--bg-sidebar) !important;

  .el-drawer__header {
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    margin-bottom: 0;
    padding: 16px 20px;
  }

  .el-menu {
    background: transparent !important;
    border-right: none !important;
  }

  .el-menu-item,
  .el-sub-menu__title {
    color: var(--text-sidebar) !important;

    &:hover {
      background: var(--bg-sidebar-hover) !important;
    }
  }

  .el-menu-item.is-active {
    background: var(--bg-sidebar-active) !important;
    color: #fff !important;
  }
}
```

### 5. 页面切换过渡动画（已有 fade，扩展）

```scss
// AppLayout.vue
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

// 登录页渐入效果（可选）
.login-layout {
  animation: fadeIn 0.4s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}
```

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| 渐变背景 | 手写 vendor prefix | CSS `linear-gradient(135deg, ...)` | 浏览器原生支持，无需 prefix，Vite 生产构建自动处理 |
| 移动端检测 | JS 检测 viewport | CSS `@media (max-width: 768px)` | 媒体查询是标准方式，性能最优 |
| 主色调替换 | 每个组件硬编码 `#1677ff` | CSS 变量 `var(--primary)` | 一次定义，全局生效，改色只需改一处 |
| 暗色菜单文字 | 逐个设置 color | Element Plus `is-active` class 覆盖 | 符合 EP 菜单组件行为 |
| 装饰圆 | 额外 DOM 元素 | CSS `::before`/`::after` 伪元素 | 无语义，保持 HTML 干净 |

---

## Common Pitfalls

### Pitfall 1: Element Plus CSS 变量优先级冲突
**What goes wrong:** 在 `global.scss` 覆盖了 `--el-color-primary`，但组件内部 `<style>` 中的 `color="#1677ff"` 硬编码导致覆盖失效。
**Why it happens:** HTML 属性 `color="#1677ff"` 优先级高于 CSS 变量。
**How to avoid:**
- 全局搜索所有 `1677ff` 硬编码，全部替换为 `var(--primary)`
- 检查 LoginView.vue 中的 `background-color: #1677ff`（当前 login-page）

### Pitfall 2: SCSS 嵌套与 Element Plus class 冲突
**What goes wrong:** SCSS 嵌套 `.sidebar { .el-menu-item { ... } }` 生成 `.sidebar .el-menu-item`，但 Element Plus 内部会用 `!important`。
**Why it happens:** EP 某些样式用 `!important` 强制覆盖。
**How to avoid:** 暗色菜单所有关键属性都加 `!important`（`color`、`background`、`border-right`）。

### Pitfall 3: 暗色侧边栏菜单 hover/active 闪烁
**What goes wrong:** 暗色菜单 hover 时背景从透明跳变，有视觉跳跃。
**Why it happens:** `background: transparent` 初始值和 hover 值之间没有过渡。
**How to avoid:** 给 `.sidebar-el-menu` 的 `.el-menu-item` 加 `transition: background 0.15s ease`。

### Pitfall 4: 移动端抽屉 z-index 被顶部栏遮挡
**What goes wrong:** `.mobile-header` z-index: 100，`el-drawer` 默认 z-index 低于 100。
**Why it happens:** drawer 的遮罩层在 header 下面，点击 header 区域无法触发 drawer 打开。
**How to avoid:** 给 `el-drawer` 的遮罩层设置 `z-index: 300` 或更高，或给 `.mobile-header` 降低到 `z-index: 50`。

### Pitfall 5: CSS Grid 右面板 `1fr` 收缩到 0
**What goes wrong:** 左面板 `720px`，右面板 `1fr`，在小屏幕（< 720px）下右面板被压缩。
**Why it happens:** `1fr` 最小值为 `auto`（内容尺寸），但在某些浏览器会收缩到 0。
**How to avoid:** 给 `login-form-panel` 设置 `min-width: 0` 或 `min-width: 320px`，并配合媒体查询切换 grid 模板。

### Pitfall 6: `variables.scss` 在 `global.scss` 之后 import 导致变量覆盖
**What goes wrong:** 两个文件都定义了 `:root`，后 import 的覆盖先 import 的。
**Why it happens:** CSS 层叠规则：后声明的同名属性覆盖前面的。
**How to avoid:** 确保 `variables.scss` 在 `global.scss` 之前 import，或将所有变量合并到一个文件。只在 `main.ts` 中保留 `@/styles/global.scss` 导入即可（variables.scss 由 global.scss 导入）。

---

## Code Examples

### 全局变量文件结构

```scss
// src/styles/variables.scss（新建）
:root {
  // Design tokens
  --primary: #4F6EF7;
  --primary-hover: #6B84F9;
  --primary-dark: #3651D9;
  --primary-light: #EEF1FF;
  --bg-page: #F0F2F5;
  --bg-sidebar: #0D1B2A;
  --bg-sidebar-active: #4F6EF7;
  --bg-sidebar-hover: #1A2D42;
  --bg-surface: #FFFFFF;
  --border: #E8ECF0;
  --text-primary: #172B4D;
  --text-secondary: #5E6C84;
  --text-tertiary: #97A0AF;
  --text-sidebar: #CDD3E0;
  --text-sidebar-active: #FFFFFF;
  --danger: #FF5630;
  --success: #36B37E;
  --warning: #FFAB00;
  --radius-lg: 12px;
  --radius-md: 8px;
  --spacing-lg: 24px;
  --spacing-md: 16px;
}
```

```scss
// src/styles/global.scss（修改）
@import './variables.scss'; // 第一行

:root {
  // Element Plus overrides
  --el-color-primary: var(--primary);
  --el-color-primary-light-3: var(--primary-hover);
  --el-color-primary-light-5: #8FA4FB;
  --el-color-primary-light-7: #B3C3FD;
  --el-color-primary-light-9: var(--primary-light);
  --el-color-primary-dark-2: var(--primary-dark);
  --el-color-success: var(--success);
  --el-color-danger: var(--danger);
  --el-color-warning: var(--warning);
  --el-border-color: var(--border);
  --el-border-radius-base: var(--radius-md);
}

// 保留原有的全局重置样式...
```

### LoginView.vue 关键样式（完整）

```vue
<style scoped lang="scss">
.login-layout {
  display: grid;
  grid-template-columns: 720px 1fr;
  min-height: 100vh;

  @media (max-width: 768px) {
    grid-template-columns: 1fr;
  }
}

.login-brand {
  background: linear-gradient(135deg, #1A2D6B 0%, #4F6EF7 60%, #7B9FFF 100%);
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 60px 48px;
  position: relative;
  overflow: hidden;

  &::before {
    content: '';
    position: absolute;
    bottom: -80px;
    left: -80px;
    width: 300px;
    height: 300px;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.06);
  }

  &::after {
    content: '';
    position: absolute;
    top: -60px;
    right: -60px;
    width: 200px;
    height: 200px;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.04);
  }

  @media (max-width: 768px) {
    display: none;
  }
}

.login-brand-content {
  position: relative;
  z-index: 1;
  color: #fff;

  .brand-logo {
    font-size: 32px;
    font-weight: 800;
    margin-bottom: 8px;
  }

  .brand-tagline {
    font-size: 20px;
    font-weight: 500;
    margin-bottom: 40px;
    opacity: 0.9;
  }

  .brand-divider {
    width: 40px;
    height: 3px;
    background: rgba(255, 255, 255, 0.5);
    border-radius: 2px;
    margin-bottom: 40px;
  }

  .feature-list {
    list-style: none;
    padding: 0;
    margin: 0;

    li {
      display: flex;
      align-items: center;
      gap: 12px;
      font-size: 15px;
      color: rgba(255, 255, 255, 0.85);
      margin-bottom: 20px;
      opacity: 0.9;

      &::before {
        content: '';
        width: 6px;
        height: 6px;
        border-radius: 50%;
        background: rgba(255, 255, 255, 0.7);
        flex-shrink: 0;
      }
    }
  }
}

.login-form-panel {
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-surface);
  padding: 24px;
  animation: slideUp 0.4s ease;

  @media (max-width: 768px) {
    align-items: flex-start;
    padding-top: 48px;
  }
}

.login-card {
  width: 100%;
  max-width: 440px;

  .brand-header {
    text-align: center;
    margin-bottom: 32px;

    h1 {
      font-size: 24px;
      font-weight: 700;
      color: var(--primary);
      margin: 0 0 6px;
    }

    p {
      font-size: 14px;
      color: var(--text-secondary);
      margin: 0;
    }
  }

  .copyright {
    text-align: center;
    font-size: 12px;
    color: var(--text-tertiary);
    margin-top: 24px;
  }
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
```

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Element Plus 2.13.6 支持通过 CSS 变量 `--el-color-primary` 整体换色 | Standard Stack | MEDIUM — 若 EP CSS 变量覆盖不完整，部分组件需手动逐个覆盖 |
| A2 | 移动端断点使用 `768px` 符合 AppLayout.vue 中已有的断点设计 | Common Pitfalls | LOW — 现有代码使用 768px，遵循一致 |
| A3 | 登录页品牌区左下/右上装饰圆使用伪元素 `::before`/`::after` 满足设计需求 | Code Examples | LOW — 若设计稿要求更复杂装饰需调整 |
| A4 | 暗色菜单 `!important` 覆盖 EP 样式不会引发其他副作用 | Architecture Patterns | MEDIUM — 若 EP 未来版本更新可能需要重新调整 |

---

## Open Questions

1. **版权年份**: REQUIREMENTS.md 写的是 `© 2025 易人事`，当前代码是 `© 2024 易人事`。需要确认是否改为 2025 或动态年份。
   - 建议：使用动态 `<script setup>` 中 `new Date().getFullYear()` 渲染。

2. **Logo 图标**: AppLayout 侧边栏 Logo 当前使用 `<Management />` 图标（Element Plus 内置）。需要确认是否有自定义 Logo 图片，还是继续使用 EP 图标。
   - 当前状态：继续使用 `<Management />` 图标即可，与原型一致。

3. **登录页品牌区内容**: 渐变背景已确认，但 Logo/Slogan/Feature List 的具体文案需要原型图确认。
   - 建议临时使用占位文案，phase plan 中留出确认节点。

---

## Environment Availability

Step 2.6: SKIPPED（Phase 1 为纯前端 UI 重构，无外部依赖）

---

## Validation Architecture

> nyquist_validation 已启用（config.json: `workflow.nyquist_validation: true`）

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Vitest（若已配置）/ 手动视觉验证 |
| Config file | `frontend/vitest.config.ts`（如存在） |
| Quick run | N/A — UI 阶段以人工验收为主 |
| Full suite | N/A |

### Phase Requirements -> Verification Map
| Req ID | Behavior | Test Type | Verification |
|--------|----------|-----------|-------------|
| UI-01 | 登录页左右分栏 | Manual | 视觉检查：左 720px 渐变，右白色卡片 |
| UI-01 | 渐变背景颜色正确 | Manual | 检查 `linear-gradient(135deg, #1A2D6B, #4F6EF7, #7B9FFF)` |
| UI-01 | 登录逻辑未改动 | Code review | 检查 LoginView.vue 中所有 `request.post` 调用未被修改 |
| UI-13 | 侧边栏背景 #0D1B2A | Manual | 视觉检查侧边栏颜色 |
| UI-13 | 菜单激活态 #4F6EF7 | Manual | 点击菜单项检查激活态颜色 |
| UI-13 | 折叠动画 0.2s | Manual | 点击折叠按钮观察过渡 |
| UI-13 | 移动端抽屉 | Manual | 浏览器模拟 375px 宽度测试 |

### Wave 0 Gaps
- 视觉回归测试建议：使用 Playwright 截图登录页和 AppLayout 作为 baseline，后续 UI 变更对比截图。
- 无自动化单元测试需求（UI 重构不涉及业务逻辑）。

---

## Sources

### Primary (HIGH confidence)
- Element Plus 官方文档 - CSS 变量系统 - https://element-plus.org/zh-CN/guide/theming.html
- Element Plus GitHub - SCSS 变量列表 - https://github.com/element-plus/element-plus/blob/dev/packages/theme-chalk/src/common/var.scss
- Vue 3 官方文档 - CSS 自定义属性 - https://vuejs.org/api/built-in-directives.html#v-bind-in-css
- CSS Grid 规范 - MDN - https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_grid_layout

### Secondary (MEDIUM confidence)
- 现有源代码验证（LoginView.vue, AppLayout.vue, main.ts, global.scss, variables.scss）
- Element Plus el-menu dark theme 社区方案

### Tertiary
- 无

---

## Metadata

**Confidence breakdown:**
- Standard Stack: HIGH — CSS 变量 + SCSS 是 Vue 3 + EP 的标准方式
- Architecture: HIGH — 基于已验证的源代码模式
- Pitfalls: MEDIUM — 基于 EP 常见问题社区反馈，部分 !important 使用为经验判断

**Research date:** 2026-04-14
**Valid until:** 2026-05-14（Element Plus 2.13.x 短期内不会有破坏性变更）
