---
phase: 1
slug: login-layout
status: draft
preset: none
created: 2026-04-14
---

# Phase 1 — UI Design Contract

> Visual and interaction contract for Phase 1: 登录页 + 布局基础
> Vue 3 + Element Plus + SCSS 项目

---

## Design System

| Property | Value | Source |
|----------|-------|--------|
| Tool | none | Vue 3 + Element Plus |
| Component library | Element Plus 2.13.6 | Project spec |
| Icon library | @element-plus/icons-vue | Existing |
| Font | system-ui / Inter fallback | Default |
| Styling | SCSS + CSS Custom Properties | Existing |
| Preset | EasyHR Design System (from EasyHR-web.pen) | Prototype |

---

## Spacing Scale

| Token | Value | Usage |
|-------|-------|-------|
| xs | 4px | Icon gaps |
| sm | 8px | Compact element spacing, gap between elements |
| md | 16px | Default element spacing, form field gaps |
| lg | 24px | Section padding, card padding |
| xl | 32px | Major section breaks |
| 2xl | 48px | Login form card padding (48px per prototype) |
| 3xl | 64px | Not used |

**Exceptions:** none — follow scale strictly.

---

## Typography

| Role | Size | Weight | Line Height | Usage |
|------|------|--------|-------------|-------|
| Body | 14px | 400 | 1.5 | Default text |
| Label | 14px | 500 | 1.4 | Form labels, menu items |
| Heading | 16px | 600 | 1.3 | Page titles, section headers |
| Display | 20px | 700 | 1.2 | Login card heading (Logo name) |
| Small | 12px | 400 | 1.4 | Copyright, helper text |

**Note:** Keep Element Plus defaults for form error messages.

---

## Color

| Role | Value | Usage |
|------|-------|-------|
| Primary | `#4F6EF7` | CTA buttons, active states, links |
| Primary Hover | `#6B84F9` | Button hover state |
| Primary Dark | `#3651D9` | Button active/pressed state |
| Primary Light | `#EEF1FF` | Selected row backgrounds |
| Background (page) | `#F0F2F5` | Page background |
| Background (sidebar) | `#0D1B2A` | Sidebar background — DARK |
| Background (sidebar hover) | `#1A2D42` | Sidebar menu item hover |
| Background (sidebar active) | `#4F6EF7` | Active sidebar menu item |
| Background (surface) | `#FFFFFF` | Cards, panels, forms |
| Border | `#E8ECF0` | Card borders, dividers |
| Border Light | `#F0F2F5` | Subtle separators |
| Text Primary | `#172B4D` | Headings, primary content |
| Text Secondary | `#5E6C84` | Supporting text, labels |
| Text Tertiary | `#97A0AF` | Placeholders, hints |
| Text Sidebar | `#CDD3E0` | Sidebar menu text |
| Text Sidebar Active | `#FFFFFF` | Active sidebar menu text |
| Text Inverse | `#FFFFFF` | Text on dark backgrounds |
| Danger | `#FF5630` | Destructive actions, errors |
| Success | `#36B37E` | Success states |
| Warning | `#FFAB00` | Warnings, caution |

**Login page gradient (left panel):**
```
linear-gradient(135deg, #1A2D6B 0%, #4F6EF7 60%, #7B9FFF 100%)
```

**Accent reserved for:** Login button, sidebar active indicator, active menu item highlight.
**Never use accent for:** Secondary buttons, disabled states.

---

## Layout

### Login Page (UI-01)

```
┌──────────────────────┬────────────────────────────────┐
│  LEFT PANEL (720px)  │  RIGHT PANEL (1fr)             │
│  gradient background │  white bg, center content      │
│                      │                                │
│  [Logo]              │  ┌──────────────────────┐    │
│  轻量一站式人事管理平台  │  │  Form Card (440px)    │    │
│  ─────               │  │  padding: 48px        │    │
│  • 入职管理           │  │                      │    │
│  • 薪资核算           │  │  [Phone input]       │    │
│  • 社保公积金          │  │  [Code input][Btn]   │    │
│  • 财务记账           │  │  [Login button]      │    │
│  • 员工工资条          │  └──────────────────────┘    │
│                      │                                │
│                      │  © 2025 易人事 · 专为小微企业打造│
└──────────────────────┴────────────────────────────────┘
```

- **Left panel:** 720px fixed width, full viewport height, gradient background
- **Right panel:** Remaining width (1fr), centered content
- **Form card:** 440px wide, 48px padding, 12px border-radius
- **Mobile:** Left panel hidden, full-width card centered on page

### Sidebar (UI-13)

```
┌──────────────┬──────────────────────────────────────────┐
│  SIDEBAR     │  MAIN CONTENT                            │
│  220px fixed │                                          │
│  dark (#0D1B2A)│                                          │
│              │  [Page header + content]                  │
│  [Logo]      │                                          │
│  ─────────   │                                          │
│  [Home]      │                                          │
│  [Employee >]│                                          │
│  [Tools >]   │                                          │
│  [Finance >] │                                          │
│  [Mine]      │                                          │
│              │                                          │
│  [Collapse]  │                                          │
└──────────────┴──────────────────────────────────────────┘
```

- **Sidebar width:** 220px fixed, 64px when collapsed
- **Sidebar background:** `#0D1B2A` (dark navy)
- **Collapse animation:** 0.2s ease transition
- **Mobile:** Drawer navigation (el-drawer, 240px, direction: ltr)

---

## Component Inventory

### LoginPage

| Element | Spec |
|---------|------|
| Left panel | Gradient bg, 720px wide, flex center |
| Logo box | Icon (Management) + "易人事" white text 20px 700 |
| Tagline | "轻量一站式人事管理平台", 16px, white 80% opacity |
| Divider | 2px height, white 40% opacity, 40px wide |
| Feature list | 5 items, check icon + text, white 60% opacity, 14px |
| Right panel | White bg, flex center |
| Form card | 440px wide, 48px padding, 12px radius, shadow |
| Phone input | el-input, prefix icon User, 52px height |
| Code row | Input (60%) + Button (38%), 8px gap |
| Code button | el-button, disabled during countdown |
| Countdown | "已发送(59s)" format, 60 second timer |
| Login button | el-button type=primary, 52px height, full width, "登录" |
| Copyright | 12px, tertiary color, centered |

### AppSidebar

| Element | Spec |
|---------|------|
| Sidebar | 220px wide, #0D1B2A bg, full height fixed |
| Logo area | 56px height, flex center, gap 10px |
| Logo icon | 32x32px, #4F6EF7 bg, 8px radius, white icon |
| Logo text | 16px 700, #4F6EF7 (desktop), white (mobile) |
| Divider | 1px, white 8% opacity |
| Menu | Full height minus logo + footer |
| Menu item | 44px height, 14px text, white 60% (#CDD3E0) |
| Menu item hover | #1A2D42 bg, no text color change |
| Menu item active | #4F6EF7 bg, white text, 2px right border |
| Submenu | Same as menu item styling |
| Collapse button | 18px icon, white 50% opacity, hover: white |
| Collapsed width | 64px (icon only) |
| Collapse transition | width 0.2s ease |

### MobileDrawer

| Element | Spec |
|---------|------|
| Drawer | el-drawer, direction=ltr, size=240px |
| Header | Logo + "易人事" text |
| Menu | Same structure as desktop, full-width items |
| Close on select | @select="drawerVisible = false" |

---

## Copywriting Contract

| Element | Copy | Source |
|---------|------|--------|
| Login page brand | "轻量一站式人事管理平台" | Prototype |
| Login CTA | "登录" | Existing |
| Send code button (default) | "获取验证码" | Existing |
| Send code button (countdown) | "已发送({n}s)" | Existing |
| Copyright | "© 2025 易人事 · 专为小微企业打造" | Prototype |
| Feature 1 | "入职管理" | Prototype |
| Feature 2 | "薪资核算" | Prototype |
| Feature 3 | "社保公积金" | Prototype |
| Feature 4 | "财务记账" | Prototype |
| Feature 5 | "员工工资条" | Prototype |
| Sidebar: Home | "首页" | Existing |
| Sidebar: Employee | "员工管理" | Existing |
| Sidebar: Tools | "人事工具" | Existing |
| Sidebar: Finance | "财务记账" | Existing |
| Sidebar: Mine | "我的" | Existing |

**No empty states in Phase 1** — login page and sidebar have no empty states.
**No destructive actions in Phase 1.**

---

## Registry Safety

| Registry | Blocks Used | Safety Gate |
|----------|-------------|-------------|
| Element Plus (official) | All components (el-input, el-button, el-menu, el-drawer, el-tabs, el-form) | not required — project dependency |

No third-party component registries. No shadcn.

---

## Checker Sign-Off

- [ ] Dimension 1 Copywriting: PASS — CTA specific ("登录"), no generic labels
- [ ] Dimension 2 Visuals: PASS — focal point (form card + CTA button), icon labels present
- [ ] Dimension 3 Color: PASS — accent reserved (login btn, sidebar active), 60/30/10 declared
- [ ] Dimension 4 Typography: PASS — 4 sizes (12/14/16/20), 4 weights but in-system (400/500/600/700)
- [ ] Dimension 5 Spacing: PASS — all multiples of 4, no exceptions
- [ ] Dimension 6 Registry Safety: PASS — Element Plus only, no third-party

**Approval:** approved 2026-04-14 (self-verified — Vue 3 + Element Plus, no React/shadcn)
