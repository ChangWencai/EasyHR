# Phase 08 Plan 02: WeChat Mini Program Frontend - Summary

**Plan:** 08-02
**Phase:** 08 — Employee WeChat Mini Program
**Status:** Complete
**Completed:** 2026-04-10

---

## One-liner

Complete WeChat Mini Program with phone+SMS login, 5-tab navigation (工资/合同/社保/报销/我的), payslip SMS verification gate, contract PDF preview, personal-only social insurance display, expense OSS upload, and expense status filtering.

---

## Tasks Completed

| # | Task | Commit | Files |
|---|------|--------|-------|
| 1 | Project scaffold (Wave 0) | 65181e8 | app.json, app.js, app.wxss, project.config.json, sitemap.json, pages.json, 10 icon PNGs |
| 2 | Core infrastructure (Wave 1) | 65181e8 | utils/request.js, utils/auth.js, utils/sms.js, utils/util.js |
| 3 | All 7 pages (Wave 2) | 65181e8 | login, payslips, payslips-detail, contracts, social, expense, expense-list, mine (32 page files) |

---

## Key Implementation Details

### 5-Tab TabBar (app.json)
- 我的工资 / pages/payslips/payslips
- 我的合同 / pages/contracts/contracts
- 社保记录 / pages/social/social
- 费用报销 / pages/expense/expense
- 我的 / pages/mine/mine

### Network Layer (utils/request.js)
- JWT injected via `Authorization: Bearer <token>` header on every request
- 401 response: clears token + member_info, redirects to `/pages/login/login`
- Exports: `request`, `get`, `post`, `put`, `del`

### Auth Utilities (utils/auth.js)
- `getToken`, `setToken`, `clearToken`, `getMemberInfo`, `setMemberInfo`, `requireAuth`
- Token stored in WeChat `wx.StorageSync`

### SMS Flow (utils/sms.js)
- `sendVerifyCode(phone)` → POST `/wxmp/auth/send-code`
- `countDown(seconds, setter)` → interval timer, returns cancel function

### Pages Implemented

| Page | API Endpoints | Key Feature |
|------|--------------|-------------|
| login | POST /wxmp/auth/send-code, POST /wxmp/auth/login | Phone+SMS, 60s countdown, JWT storage |
| payslips | GET /wxmp/payslips | Monthly list, status badges |
| payslips-detail | GET /wxmp/payslips/:id, POST /wxmp/payslips/:id/verify, POST /wxmp/payslips/:id/sign | SMS gate + sign action |
| contracts | GET /wxmp/contracts | PDF via wx.openDocument |
| social | GET /wxmp/social-insurance | Personal amounts only, no employer data |
| expense | POST /wxmp/oss/upload-url, POST /wxmp/expenses | OSS pre-signed upload, photo grid |
| expense-list | GET /wxmp/expenses | Status filter tabs, expandable detail |
| mine | - | Profile display + logout |

---

## Design System (per 08-UI-SPEC.md)

- Accent: #1677FF (商务蓝)
- Background: #F5F5F5
- White cards with left blue border on list items
- Typography: 18px headings, 16px body, 14px caption, 12px mini
- Status badges: pending=#FAAD14, success=#52C41A, danger=#FF4D4F, accent=#1677FF

---

## Deviations from Plan

None — plan executed exactly as written.

---

## Artifacts Created

- `miniprogram/app.json` — 5-tab tabBar, 8 pages registered
- `miniprogram/project.config.json` — WeChat DevTools config
- `miniprogram/utils/request.js` — JWT interceptor + 401 handler
- `miniprogram/utils/auth.js` — token/member_info CRUD
- `miniprogram/utils/sms.js` — SMS verification flow
- `miniprogram/utils/util.js` — formatAmount, formatMonth, formatDate
- `miniprogram/pages/login/` — login.js + .wxml + .wxss + .json
- `miniprogram/pages/payslips/` — payslips list + .wxml + .wxss + .json
- `miniprogram/pages/payslips-detail/` — detail with SMS gate + .wxml + .wxss + .json
- `miniprogram/pages/contracts/` — contract list + PDF + .wxml + .wxss + .json
- `miniprogram/pages/social/` — personal social insurance + .wxml + .wxss + .json
- `miniprogram/pages/expense/` — expense form + OSS upload + .wxml + .wxss + .json
- `miniprogram/pages/expense-list/` — expense status list + .wxml + .wxss + .json
- `miniprogram/pages/mine/` — profile + logout + .wxml + .wxss + .json

---

## Self-Check

- [x] miniprogram/app.json tabBar has exactly 5 entries
- [x] All 8 page paths registered in app.json
- [x] All 32 page files exist (.js, .wxml, .wxss, .json)
- [x] All 4 utils files exist with correct exports
- [x] 10 icon PNG placeholder files created
- [x] All files committed to git (no untracked in miniprogram/)

## Commit

`65181e8` — feat(08-02): implement WeChat Mini Program frontend scaffold and all 7 pages (52 files, 730 lines)
