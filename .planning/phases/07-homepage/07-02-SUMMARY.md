---
phase: "07-homepage"
plan: "02"
subsystem: ui
tags: [vue3, vite, element-plus, pinia, vue-router, typescript, sass]

requires:
  - phase: "07-01"
    provides: Go dashboard REST API at /api/v1/dashboard
provides:
  - Vue 3 H5 SPA scaffold with Vite 5 + TypeScript
  - Bottom tab navigation (首页/员工/工具/财务/我的)
  - Home dashboard with todo cards, 5-grid, collapsible overview
  - JWT auth interceptor in Axios (Bearer token)
  - Pinia stores for auth and dashboard state
  - First-time onboarding overlay
affects: [08-employee, 09-tool, 10-finance, 11-mine]

tech-stack:
  added: [vue@3.5.32, vue-router@5, pinia@3, element-plus@2.13.6, @element-plus/icons-vue, axios, dayjs, sass, unplugin-auto-import, unplugin-vue-components, vite@5, eslint@10, prettier, vitest]
  patterns: [API response interceptor, Pinia setup stores, Vue Router lazy loading, SCSS scoped styles, Element Plus auto-import]

key-files:
  created:
    - frontend/src/api/request.ts (Axios JWT interceptor)
    - frontend/src/api/dashboard.ts (Dashboard API client)
    - frontend/src/stores/auth.ts (Pinia auth store)
    - frontend/src/stores/dashboard.ts (Pinia dashboard store)
    - frontend/src/router/index.ts (Vue Router 5 with 5-tab routes)
    - frontend/src/views/layout/BottomTabBar.vue (Fixed bottom tab nav)
    - frontend/src/views/layout/AppLayout.vue (Layout wrapper)
    - frontend/src/views/home/HomeView.vue (Home dashboard)
    - frontend/src/App.vue (Root with onboarding overlay)
    - frontend/src/styles/global.scss (Global reset + safe-area)
    - frontend/src/styles/variables.scss (Element Plus theme overrides)
    - frontend/vite.config.ts (Vite with auto-import + proxy)
    - frontend/eslint.config.js (ESLint flat config)
  modified:
    - frontend/package.json (added all deps + scripts)
    - frontend/tsconfig.app.json (path alias + ignoreDeprecations)
    - frontend/index.html (mobile viewport + iOS meta tags)

key-decisions:
  - "Vite 5 instead of Vite 8 (Vite 8 crashes with Bus error on macOS Node 22)"
  - "ESLint flat config (eslint.config.js) for ESLint v10 compatibility"
  - "Element Plus icons auto-imported via unplugin-vue-components"
  - "Hash history router (createWebHashHistory) for GitHub Pages compatibility"
  - "Wallet icon instead of non-existent MoneyCollect for financial module"

patterns-established:
  - "Axios interceptor pattern for JWT injection + 401 redirect"
  - "Pinia setup store pattern (defineStore with composition API)"
  - "Vue Router lazy loading with dynamic import()"
  - "SCSS scoped styles with lang='scss' attribute"

requirements-completed: [HOME-01, HOME-02, HOME-03, HOME-04, HOME-05, HOME-06]

duration: 6min
completed: 2026-04-10
---

# Phase 07 Plan 02: Vue 3 H5 Frontend — Project Init + Home Dashboard Summary

**Vue 3 H5 SPA scaffold with home dashboard, bottom tab navigation, JWT auth interceptor, and first-time onboarding overlay**

## Performance

- **Duration:** 6 min
- **Started:** 2026-04-10T23:21:22+08:00
- **Completed:** 2026-04-10T23:26:48+08:00
- **Tasks:** 5 (Scaffold + Infrastructure + Home UI + Onboarding + Verification)
- **Files created/modified:** ~25

## Accomplishments
- Complete Vue 3 + Vite 5 + TypeScript H5 SPA scaffold with mobile-first config
- Bottom tab bar with 5 tabs (首页/员工/工具/财务/我的) fixed at bottom with iOS safe-area
- Home dashboard with: blue header, API-driven todo cards, 5-grid core functions, collapsible data overview
- JWT Bearer token injection via Axios interceptor, 401 auto-redirect to /login
- First-time user onboarding overlay (3 steps, localStorage dismissal flag)
- All npm scripts: dev, build, type-check, lint, test:unit

## Task Commits

1. **Task 1: Initialize Vite project scaffold** - `19e78f6` (feat) — Vite 5 + deps + mobile index.html
2. **Task 2: Build core frontend infrastructure** - `7e871ea` (feat) + `ba74ef4` (refactor) — API client, stores, router, styles, main.ts, App.vue
3. **Task 3: Implement home dashboard UI** - `9627ae8` (feat) — BottomTabBar, AppLayout, HomeView
4. **Task 3b: First-time onboarding** - `4cdf267` (feat) — Onboarding overlay in App.vue
5. **Task 4: Verification + lint + cleanup** - `f28d741` (chore) + `013c482` (refactor) — ESLint flat config, unused scaffold removal

## Decisions Made

- **Vite 5 vs Vite 8:** Vite 8 crashes with Bus error on macOS Node 22; Vite 5 works reliably
- **ESLint flat config:** ESLint v10 requires `eslint.config.js` (flat format); `.eslintrc.cjs` no longer supported
- **Hash history:** Using `createWebHashHistory()` for GitHub Pages compatibility (Vite dev server works with both)
- **Icon replacement:** `MoneyCollect` not in `@element-plus/icons-vue` — replaced with `Wallet`
- **Unused scaffold cleanup:** Removed HelloWorld.vue, hero.png, vite.svg, vue.svg, style.css

## Deviations from Plan

None - plan executed exactly as written.

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Vite 8 Bus error on macOS**
- **Found during:** Task 1 (scaffold verification)
- **Issue:** `vite build` with Vite 8.0.8 causes "Bus error: 10" crash on macOS Node 22.14.0
- **Fix:** Downgraded to `vite@5` which works correctly
- **Files modified:** frontend/package.json
- **Verification:** `npm run build` succeeds with vite 5
- **Committed in:** `013c482` (cleanup commit)

**2. [Rule 3 - Blocking] Missing AppLayout and HomeView stubs**
- **Found during:** Task 2 (build verification)
- **Issue:** Router references `@/views/layout/AppLayout.vue` and `@/views/home/HomeView.vue` which didn't exist yet
- **Fix:** Created stub placeholder files so router compiles, replaced with full implementation in Task 3
- **Files created:** frontend/src/views/layout/AppLayout.vue, frontend/src/views/home/HomeView.vue
- **Verification:** `npm run build` passes after stubs added
- **Committed in:** `ba74ef4` (Task 2)

**3. [Rule 3 - Blocking] MoneyCollect icon not in @element-plus/icons-vue**
- **Found during:** Task 3 (build verification)
- **Issue:** `MoneyCollect` is not exported from `@element-plus/icons-vue`
- **Fix:** Replaced with `Wallet` icon (verified available via `node -e "require('@element-plus/icons-vue')"`)
- **Files modified:** frontend/src/views/home/HomeView.vue
- **Verification:** `npm run build` passes
- **Committed in:** `9627ae8` (Task 3)

**4. [Rule 1 - Bug] ESLint v10 requires flat config**
- **Found during:** Task 4 (lint verification)
- **Issue:** `npm run lint` failed because ESLint v10 dropped `.eslintrc.*` support
- **Fix:** Created `eslint.config.js` flat config with plugin-vue, @vue/eslint-config-typescript, @vue/eslint-config-prettier
- **Files created:** frontend/eslint.config.js
- **Verification:** `npm run lint` passes with no output
- **Committed in:** `f28d741` (Task 4)

---

**Total deviations:** 4 auto-fixed (4 blocking)
**Impact on plan:** All auto-fixes were necessary for the project to compile and pass lint. No scope creep.

## Issues Encountered
- **Vite 8 crash:** See deviation #1 above. Solved by downgrading to Vite 5.
- **ESLint v10 migration:** See deviation #4 above. Solved by creating flat config.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Frontend scaffold is fully functional and committed to git
- Home dashboard UI implemented, connects to Go backend at `http://localhost:8080` via Vite proxy
- Bottom tab navigation in place with placeholder routes for employee/finance/tools/mine
- Ready for next phase (08-employee management)
- Go backend dashboard handler from phase 07-01 should be running for full integration testing

---
*Phase: 07-homepage / 07-02*
*Completed: 2026-04-10*
