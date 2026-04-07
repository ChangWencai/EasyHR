---
phase: 3
slug: social-insurance
status: draft
shadcn_initialized: false
preset: none
created: 2026-04-07
---

# Phase 3 -- UI Design Contract

> Visual and interaction contract for the social insurance management module. Phase 3 is primarily a Go backend API phase. The H5 admin panel (Vue 3 + Element Plus) will need policy management pages. The native APP and WeChat Mini Program will consume these APIs in later phases.

---

## Design System

| Property | Value |
|----------|-------|
| Tool | none (Go backend phase; H5 admin uses Vue 3 + Element Plus) |
| Preset | not applicable |
| Component library | Element Plus 2.13.6 (H5 admin panel) |
| Icon library | Element Plus built-in icons + linear minimalist icons (2px rounded, per ui-ux.md) |
| Font | System sans-serif (mobile default); Source: ui-ux.md 1.3 |

---

## Phase Scope -- UI Surfaces

This phase produces **backend API endpoints** consumed by three frontend surfaces. The UI contract defines what the API must serve to support each surface's visual requirements.

### Surface 1: H5 Admin Panel -- Policy Management (D-02)
Admins manually input/edit social insurance policy data for 30+ cities via the H5 backend. This is the only new UI surface in Phase 3.

### Surface 2: Native APP -- Social Insurance Module (consumed in Phase 7)
APP pages that will consume Phase 3 APIs. UI contract here ensures API response shape supports the declared layouts.

### Surface 3: WeChat Mini Program -- Employee Social Insurance View (consumed in Phase 8)
Read-only employee view of social insurance records. UI contract here ensures API provides sufficient data for the declared layout.

---

## Spacing Scale

Declared values (must be multiples of 4). Source: ui-ux.md 1.3 "页面间距8px/16px层级分明".

| Token | Value | Usage |
|-------|-------|-------|
| xs | 4px | Icon gaps, inline padding |
| sm | 8px | Compact element spacing, mobile page margins |
| md | 16px | Default element spacing, card internal padding, section gaps |
| lg | 24px | Section padding, card margins |
| xl | 32px | Layout gaps, module breaks |
| 2xl | 48px | Major section breaks |
| 3xl | 64px | Page-level spacing |

Exceptions: Touch targets for mobile interactive elements must be minimum 44px per mobile accessibility guidelines. This is the only exception to the 4px scale.

---

## Typography

Source: ui-ux.md 1.3 "标题18px/正文16px/辅助文字14px，行高1.5倍".

| Role | Size | Weight | Line Height | Usage |
|------|------|--------|-------------|-------|
| Body | 16px | 400 (regular) | 1.5 | Main content text, form labels, data values |
| Label | 14px | 400 (regular) | 1.5 | Secondary text, helper text, timestamps, table secondary columns |
| Heading | 18px | 600 (semibold) | 1.2 | Page titles, section headers, card titles |
| Display | 20px | 600 (semibold) | 1.2 | Key financial figures (total social insurance amount, base amount) |

Declared font weights: exactly 2 -- regular (400) and semibold (600).

---

## Color

Source: ui-ux.md 1.3 "商务蓝 #1677FF + 辅助绿 #52C41A + 红 #FF4D4F".

| Role | Value | Usage |
|------|-------|-------|
| Dominant (60%) | #FFFFFF | Page background, surface areas, form backgrounds |
| Secondary (30%) | #F5F5F5 | Cards, list item backgrounds, section containers, sidebar |
| Accent (10%) | #1677FF | Primary actions: "确认参保" button, active tab indicator, selected employee checkbox, link text, navigation active state |
| Destructive | #FF4D4F | Risk/warning actions: "停缴" button, overdue payment alerts, deletion confirmation |
| Success | #52C41A | Operation success feedback, "参保中" status badge, completed actions |

Accent reserved for:
- "确认参保" / "确认" primary CTA button
- Active tab / selected state indicator
- Interactive link text
- Employee selection checkbox (checked state)
- Bottom navigation active icon

Destructive reserved for:
- "停缴" (stop enrollment) action button
- Overdue payment warning badge
- Delete policy confirmation
- Error state text in form validation

Success reserved for:
- "参保中" (active) status badge
- Operation success toast/banner
- "已导出" (exported) confirmation

---

## Copywriting Contract

Source: CONTEXT.md D-05 (3-step flow), D-07 (stop triggers), ui-ux.md 4.2 (status feedback).

| Element | Copy |
|---------|------|
| Primary CTA (enrollment) | 确认参保 |
| Primary CTA (stop enrollment) | 确认停缴 |
| Primary CTA (policy edit) | 保存政策 |
| Primary CTA (export) | 导出凭证 |
| Empty state heading (no enrolled employees) | 暂无参保员工 |
| Empty state body (no enrolled employees) | 点击"参保办理"按钮，为员工办理社保参保 |
| Empty state heading (no payment records) | 暂无缴费记录 |
| Empty state body (no payment records) | 参保成功后将自动生成缴费记录 |
| Empty state heading (no policies for city) | 暂未配置该城市社保政策 |
| Empty state body (no policies for city) | 请联系管理员在后台添加社保政策数据 |
| Error state (policy not found) | 未找到该城市社保政策，请联系管理员配置 |
| Error state (batch partial failure) | 部分员工参保失败，请查看失败原因后重试 |
| Error state (network) | 网络连接失败，请检查网络后重试 |
| Destructive confirmation (stop enrollment) | 停缴确认：停缴后该员工社保将自下月起停止缴纳，确定继续？ |
| Destructive confirmation (delete policy) | 删除政策：删除后该城市社保政策数据将不可恢复，确定删除？ |
| Success toast (enrollment) | 参保成功，共为 {N} 名员工办理社保 |
| Success toast (stop) | 停缴成功，共 {N} 名员工 |
| Success toast (export) | 凭证导出成功 |
| Payment reminder (3 days before due) | 社保缴费提醒：{N}名员工社保将于{date}到期，请及时缴费 |
| Base adjustment reminder | 社保基数调整建议：{employee_name}薪资变动，建议调整社保基数 |
| Resignation stop reminder | 社保停缴提醒：{employee_name}已离职，请及时办理社保停缴 |

---

## Component Inventory -- H5 Admin Policy Management

These are the UI components needed for the H5 admin policy management pages (D-02). All use Element Plus components.

| Component | Element Plus Base | Purpose |
|-----------|-------------------|---------|
| Policy city selector | ElSelect with ElOption | Select city from 37-city list |
| Policy year input | ElInputNumber | Effective year (e.g. 2025, 2026) |
| Insurance item form (x6) | ElForm with ElFormItem | Pension, Medical, Unemployment, Work Injury, Maternity, Housing Fund |
| Rate input (company/personal) | ElInputNumber (percentage, step 0.01) | Company rate and personal rate per insurance type |
| Base limit input (upper/lower) | ElInputNumber (currency, step 1) | Base upper and lower limits per insurance type |
| Policy list table | ElTable with ElTableColumn | List all policies by city and year |
| Save/Update button | ElButton type="primary" | Save or update policy |
| Delete policy button | ElButton type="danger" | Delete policy with ElMessageBox confirm |

---

## API Response Shape -- Visual Requirements

The API must return data in shapes that support the following mobile UI layouts (defined in ui-ux.md 2.3.2).

### Enrollment Preview (Step 2 of D-05 3-step flow)
The API must return per-employee insurance breakdown for the confirmation screen:

```
{
  employee_id: int64,
  employee_name: string,
  city_name: string,
  base_amount: float64,
  total_company: float64,
  total_personal: float64,
  items: [
    {
      name: "养老保险",
      base: float64,
      company_rate: float64,
      company_amount: float64,
      personal_rate: float64,
      personal_amount: float64
    },
    // ... 5 more insurance types
  ]
}
```

### Payment Detail List (SOCL-04)
Must support card-based list display with: employee name, enrollment month, base amount, total company amount, total personal amount, status badge (active/stopped/pending).

### Change History (SOCL-07)
Must support timeline display with: change type (enrollment/base adjustment/stop), change date, before/after values, operator name.

### Export (SOCL-05)
Must return Excel file with headers: employee name, city, enrollment month, base amount, each insurance type company/personal amounts, totals.

---

## Registry Safety

| Registry | Blocks Used | Safety Gate |
|----------|-------------|-------------|
| shadcn official | none | not applicable -- project uses Vue 3 + Element Plus, not React/shadcn |
| Third-party | none | not applicable |

No third-party registries declared. No vetting required.

---

## Interaction Flow Contract

Source: CONTEXT.md D-05, ui-ux.md 2.3.2.

### Enrollment Flow (3 steps, D-05)
1. **Select employees** -- Multi-select employee list with checkboxes. Default sort: recent hire date. Show employee name, position, city.
2. **Preview amounts** -- Auto-calculated insurance breakdown per selected employee. Show base amount (matched by city policy), each insurance type with company/personal amounts. Highlight total company cost and total personal deduction.
3. **Confirm enrollment** -- "确认参保" button triggers batch enrollment. Show confirmation dialog with count. On success, show success toast with count.

### Stop Enrollment Flow
1. Select employee(s) from active enrollment list
2. Show confirmation dialog with destructive color (#FF4D4F): "停缴后该员工社保将自下月起停止缴纳"
3. On confirm, batch stop. Show success toast.

### Policy Management Flow (H5 Admin, D-02)
1. Select city from dropdown
2. Select effective year
3. Fill in 6 insurance type forms (company rate, personal rate, base lower, base upper)
4. Save. Show success toast.

### Payment Reminder Flow (D-09, D-10)
- gocron daily scan at 08:00 CST
- 3 days before due: generate reminder record
- Reminder consumed by Phase 7 homepage as card: "{N}名员工社保将于{date}到期，请及时缴费"
- No SMS, no WeChat template message (V1.0 cost control, D-10)

---

## Checker Sign-Off

- [ ] Dimension 1 Copywriting: PASS
- [ ] Dimension 2 Visuals: PASS
- [ ] Dimension 3 Color: PASS
- [ ] Dimension 4 Typography: PASS
- [ ] Dimension 5 Spacing: PASS
- [ ] Dimension 6 Registry Safety: PASS

**Approval:** pending
