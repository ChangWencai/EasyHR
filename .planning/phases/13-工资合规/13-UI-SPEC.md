# Phase 13 — UI Design Contract

> Visual and interaction contract for Phase 13: 工资合规. Salary slip confirmation receipt.

---

## Design System

| Property | Value |
|----------|-------|
| Tool | none (Vue 3 + Element Plus, no shadcn) |
| Preset | not applicable |
| Component library | Element Plus |
| Icon library | @element-plus/icons-vue |
| Font | Inter (fallback: -apple-system, BlinkMacSystemFont, Segoe UI, Roboto) |

**Design token source:** `frontend/src/styles/variables.scss` (CSS custom properties)

---

## Visual Hierarchy

Declared focus order (primary to tertiary):

| Level | Element | Role |
|-------|---------|------|
| 1 (primary) | `.slip-header` + `.net-income-card` | 月份 + 实发工资金额 — the two things the employee came to see |
| 2 (secondary) | `.sign-action` button | Primary action — confirm receipt |
| 3 (tertiary) | `.slip-section` table data | Detailed income/deduction breakdown |

---

## Spacing Scale

Declared values (multiples of 4, from `variables.scss`):

| Token | Value | Usage |
|-------|-------|-------|
| xs | 4px | Icon gaps, inline padding |
| sm | 8px | Compact element spacing |
| md | 16px | Default element spacing |
| lg | 24px | Section padding |
| xl | 32px | Layout gaps |
| 2xl | 48px | Major section breaks |

**Exceptions:** None — reuse exactly from `variables.scss`.

---

## Typography

**4 sizes, 2 weights** (consistent with Phase 12):

| Role | Size | Weight | Line Height | Notes |
|------|------|--------|-------------|-------|
| Page title / stat value | 28px | 700 | 1.2 | `.page-title` from variables.scss |
| Body / table / button | 14px | 400 | 1.5 | Default Element Plus body |
| Table header / labels | 13px | 600 | — | `font-weight: 600` in `header-cell-style` |
| Caption / tertiary labels | 12px | 400 | — | confirmation timestamp, error text |

**Font stack:** `'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif`

---

## Color

**Ratio: 60% dominant / 30% secondary / 10% accent** (consistent with Phase 12):

| Role | Value | Ratio | Usage |
|------|-------|-------|-------|
| Dominant (60%) | `#F5F6F8` | 60% | Page background (`.salary-slip-h5`) |
| Secondary (30%) | `#FFFFFF` | 30% | Cards, table, dialogs |
| Accent (10%) | `#4F6EF7` | 10% | Primary action button only |
| Success | `#10B981` | — | Confirmed status, net income |
| Warning | `#F59E0B` | — | Viewed but not confirmed |
| Danger | `#EF4444` | — | Error states, failed sends |

**Accent reserved for (10%):**
- Primary CTA button: "确认已收到" (Phase 13 new action)
- Active sidebar menu item
- Gradient header background (existing)

**Color tokens from `variables.scss`:**
- `--primary: #4F6EF7` / `--primary-hover: #3D5CF5`
- `--success: #10B981`
- `--warning: #F59E0B`
- `--danger: #EF4444`
- `--text-primary: #1F2937` / `--text-secondary: #6B7280` / `--text-tertiary: #9CA3AF`

---

## Component Inventory

### 1. Salary Slip H5 Page — Confirmation Button

**Trigger condition (D-13-01, D-13-02):**
- Show "确认已收到" button ONLY when: `status !== 'signed'` (not already confirmed/signed)
- Do NOT show when already confirmed
- "确认签收" existing button → rename to "确认已收到" per D-13-02

**Button state machine:**
```
status='sent'       → show nothing (wait for viewing)
status='viewed'     → show "确认已收到" primary button
status='signed'     → show "已确认" green badge with timestamp
```

**Button design:**
- `el-button type="primary" size="large"` (consistent with Phase 12 export button)
- Full-width, max-width 320px, centered
- Label: "确认已收到" (not "确认签收")
- Icon: `<Check />` from @element-plus/icons-vue
- Loading state: spinner + "确认中..."
- On success: redirect to confirmation success page

**Confirmation success page (new inline state in SalarySlipH5.vue):**
```
.slip-content → replaced by .confirm-success on confirmation

.confirm-success:
  .success-icon: 72x72 circle, --success color background, checkmark icon
  h2: "工资条已确认" (28px, weight 700)
  p: "您已确认 {year}年{month}月 工资条" (14px, tertiary)
  .back-link: "返回" (text button)
```

**Existing `.sign-action` → rename to `.confirm-action`**
**Existing `.sign-status` → rename to `.confirm-status`**

### 2. Salary Slip H5 — Confirmation Status Badge

After confirmation, show green badge instead of button:

```
.confirm-status (replaces .sign-status):
  background: rgba(16, 185, 129, 0.08)
  border: 1px solid rgba(16, 185, 129, 0.2)
  border-radius: 8px
  padding: 12px 16px
  display: flex, align-items: center, gap: 8px

  el-icon (Check): color --success
  span: "{confirmed_at} 已确认" (14px, color --success)
```

### 3. SalarySlipSend — Confirmation Status Column (D-13-11)

**New column in send log el-table (insert after "发送时间" column):**

| Column | Width | Content |
|--------|-------|---------|
| 确认状态 | 140px | Badge showing confirmation state |

**Column template:**
```vue
<el-table-column prop="confirmed_at" label="确认状态" width="140">
  <template #default="{ row }">
    <span v-if="row.confirmed_at" class="confirm-badge confirm-badge--confirmed">
      <el-icon><Check /></el-icon>
      已确认
    </span>
    <span v-else class="confirm-badge confirm-badge--unconfirmed">
      <el-icon><WarnTriangleFilled /></el-icon>
      未确认
    </span>
  </template>
</el-table-column>
```

**Badge styles:**
```scss
.confirm-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
}
.confirm-badge--confirmed {
  background: rgba(16, 185, 129, 0.1);
  color: #10B981;
}
.confirm-badge--unconfirmed {
  background: rgba(245, 158, 11, 0.1);
  color: #F59E0B;
}
```

**SlipSendLog interface extension (salary.ts):**
```typescript
export interface SlipSendLog {
  // ... existing fields ...
  confirmed_at?: string   // new — ISO timestamp of employee confirmation
}
```

### 4. Send Log Filter Toolbar

No changes to existing toolbar (year/month picker, refresh button). Confirmation status is a data field in the log, not a filter dimension.

---

## Page Layouts

### SalarySlipH5.vue — Confirmation Flow

```
.salary-slip-h5
  [loading/error states — unchanged]

  .slip-content (when slip loaded and not yet confirmed)
    .slip-header         ← level 1: 月份 title
    .employee-card       ← 员工姓名 + status tag
    .slip-section        ← 应发明细
    .slip-section        ← 扣除明细
    .net-income-card     ← level 1: 实发金额 (visual focal point)
    .confirm-action      ← level 2: "确认已收到" primary CTA (NEW)
    .confirm-status      ← level 2: green confirmed badge (if confirmed)
    .slip-footer

  .confirm-success (after successful confirmation — replaces .slip-content)
    .success-icon
    h2: "工资条已确认"
    p: "您已确认 {year}年{month}月 工资条"
    .back-link
```

### SalarySlipSend.vue — Send Log with Confirmation Column

Existing layout unchanged. New column inserted into existing `el-table` between "发送时间" and the final column.

```
.tab-content (logs tab)
  .toolbar-card           ← year/month filter + refresh (unchanged)
  .table-card
    el-table
      employee_id col     (existing)
      channel col         (existing)
      status col          (existing — slip sent status)
      confirmed_at col     (NEW — employee confirmation status)
      error_message col   (existing)
      sent_at col         (existing)
      created_at col      (existing)
```

---

## Copywriting Contract

| Element | Phase 12 Copy | Phase 13 Copy | Notes |
|---------|---------------|---------------|-------|
| Primary CTA (H5) | — | 确认已收到 | D-13-02 confirmation button |
| CTA loading | — | 确认中... | Button loading text |
| CTA success heading | — | 工资条已确认 | Confirmation success page h2 |
| CTA success body | — | 您已确认 {year}年{month}月 工资条 | Dynamic month |
| Confirmation badge confirmed | — | 已确认 | Green badge with check icon |
| Confirmation badge unconfirmed | — | 未确认 | Yellow badge with warning icon |
| Column header (send log) | — | 确认状态 | New column in SalarySlipSend |
| Error message | — | 确认失败，请重试 | API error on confirm |

**DO NOT use:** "签收", "确认签收" — use "确认已收到" per D-13-02.

---

## Routing Contract

No new routes for Phase 13. Existing routes unchanged:
- `/tool/salary-slip/:token` — SalarySlipH5.vue (confirm button added)
- `/tool/salary-slip-send` — SalarySlipSend.vue (confirmation column added)

---

## Component File Locations

| Component | Path | Change |
|-----------|------|--------|
| SalarySlipH5.vue | `frontend/src/views/tool/SalarySlipH5.vue` | Confirm button + success state |
| SalarySlipSend.vue | `frontend/src/views/tool/SalarySlipSend.vue` | Confirmation status column |
| salary.ts (API) | `frontend/src/api/salary.ts` | SlipSendLog interface + confirm API |

---

## API Contract (frontend)

### Confirmation API (new — D-13-03)

```
POST /salary/slip/{token}/confirm
Authorization: none (token in URL is the slip access token)
Response: { success: true }
Side effect: Updates PayrollSlip.confirmed_at + confirmed_ip

GET /salary/slip/{token}
Response: SlipDetail (existing, unchanged)
Note: confirmed_at returned as part of existing SlipDetail response
```

### Send Log API (extend — D-13-11)

```
GET /salary/slip/logs?year=YYYY&month=M
Response: { logs: SlipSendLog[], total: number }
Change: SlipSendLog.confirmed_at is now returned (ISO timestamp or null)
```

---

## Interaction Patterns

### Employee Confirmation Flow (SalarySlipH5.vue)
1. Employee opens salary slip H5 via SMS/miniapp link with token
2. Views salary details (status = 'sent' or 'viewed')
3. Clicks "确认已收到" button
4. Button shows loading state ("确认中...")
5. Frontend POSTs to `/salary/slip/{token}/confirm`
6. On success: replace `.slip-content` with `.confirm-success`, show month + success message
7. On error: show `ElMessage.error('确认失败，请重试')`, re-enable button

### Boss Confirmation Monitoring Flow (SalarySlipSend.vue)
1. Boss navigates to "发送记录" tab
2. Selects year/month to filter
3. Table shows all send logs with new "确认状态" column
4. Yellow "未确认" badge → employee has not yet confirmed
5. Green "已确认" badge with timestamp → employee confirmed
6. No action available for boss in this phase (D-13-08: automatic reminder comes via TodoCenter)

---

## Responsive Breakpoints

From `variables.scss` (unchanged from Phase 12):

| Breakpoint | Width | Layout |
|------------|-------|--------|
| Default | >= 1024px | Full sidebar (260px), full-width table |
| md | < 1024px | Sidebar collapses to 72px |
| sm | < 768px | Mobile header, drawer nav, horizontal scroll for table |

**Phase 13 specific:** Confirmation button is full-width on mobile (max-width: 320px, centered).

---

## Accessibility Notes

- "确认已收到" button: `aria-label="确认已收到工资条"`
- Confirmation success heading: level 2 (`<h2>`) for screen reader structure
- Confirmation badges: `aria-label` on span, e.g., `aria-label="已确认于 2026-04-20 10:30"`
- Color not sole indicator — always pair color with text/icon ("未确认" text + yellow badge)

---

## Registry Safety

| Registry | Blocks Used | Safety Gate |
|----------|-------------|-------------|
| none | none | not applicable |

No new third-party registries — only `@element-plus/icons-vue` (built-in).

---

## Pre-Populated From

| Source | Decisions Used |
|--------|---------------|
| 13-CONTEXT.md | D-13-01 through D-13-12 (all confirmation/design decisions) |
| ROADMAP.md | Phase 13 goal (工资条发放确认全流程管理) |
| REQUIREMENTS.md §COMP-09~COMP-11 | Confirmation requirements |
| `SalarySlipH5.vue` | Existing H5 slip view structure, styles, sign flow |
| `SalarySlipSend.vue` | Existing send log table structure, column layout |
| `frontend/src/styles/variables.scss` | All CSS tokens |
| 12-UI-SPEC.md | Design system consistency (typography, spacing, color ratio) |

---

## Differences from Phase 12 UI Pattern

Phase 13 is a **modification** to existing pages (not new pages), so it follows Phase 12 patterns with these Phase 13-specific additions:

| Pattern | Phase 12 | Phase 13 |
|---------|----------|----------|
| Pages affected | 4 new pages | 2 existing pages modified |
| CTA style | `el-button type="primary" size="large"` | Same — "确认已收到" |
| Primary action location | Page header right | Slip content bottom (below net income) |
| Stat cards | 4-column grid | N/A — single confirmation action |
| Table columns | N/A | New confirmation column added to existing table |
| Empty state | `.empty-state.glass-card` | N/A — no empty state for confirmation |
| Export flow | Blob download | N/A — no export in this phase |

---

## Checker Sign-Off

- [ ] Dimension 1 Copywriting: "确认已收到" used (not "确认签收")
- [ ] Dimension 2 Visuals (Visual Hierarchy): Confirmation button is level 2, below fiscal data
- [ ] Dimension 3 Color (60/30/10): Green for confirmed, yellow for unconfirmed
- [ ] Dimension 4 Typography: 4 sizes, consistent with Phase 12
- [ ] Dimension 5 Spacing: tokens from variables.scss
- [ ] Dimension 6 Registry Safety: No new dependencies
