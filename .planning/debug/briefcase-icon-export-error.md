---
name: briefcase-icon-export-error
description: Vue Router fails to start due to missing BriefCase icon export
type: debug
status: resolved
trigger: "SyntaxError: The requested module '/node_modules/.vite/deps/@element-plus_icons-vue.js' does not provide an export named 'BriefCase' (at LoginView.vue:187:34)"
created: 2026-04-19
updated: 2026-04-19
---

## Symptoms

- **Expected**: App loads normally, login page renders at localhost:5173/#/
- **Actual**: Page is completely blank; router fails to start
- **Error**: `SyntaxError: The requested module '/node_modules/.vite/deps/@element-plus_icons-vue.js' does not provide an export named 'BriefCase' (at LoginView.vue:187:34)`
- **Timeline**: Started after recent changes
- **Reproduction**: Visit `http://localhost:5173/#/` — blank page shown

## Evidence

- Error occurs at `LoginView.vue:187:34` — `BriefCase` icon import from `@element-plus/icons-vue`
- Vue Router fails to install due to this uncaught syntax error
- Additional Sass warning: `@import './variables.scss'` in `src/styles/global.scss` line 1:9 (may be unrelated)
- Confirmed: `node -e "const icons = require('@element-plus/icons-vue'); console.log(Object.keys(icons).filter(k => k.toLowerCase().includes('brief')))"` → `Briefcase` (correct name, with lowercase 'c'), `BriefCase` does not exist

## Current Focus

- **Hypothesis**: `@element-plus/icons-vue` does not export a `BriefCase` icon — likely a naming mismatch (should be `Briefcase`, `OfficeBuilding`, or another variant)
- **Next action**: Check actual exports of `@element-plus/icons-vue` and fix the import in LoginView.vue

## Eliminated

## Root Cause

`@element-plus/icons-vue` does not export a `BriefCase` icon. The correct export name is `Briefcase` (capital B, lowercase 'c'). The import in `LoginView.vue:187` and template usage at line 20 both used the incorrect `BriefCase` name.

## Fix

Two changes in `frontend/src/views/layout/LoginView.vue`:
1. Line 20 (template): `<BriefCase />` → `<Briefcase />`
2. Line 187 (script import): `BriefCase` → `Briefcase`

Fixed file: `frontend/src/views/layout/LoginView.vue`

## Verification

Dev server starts successfully at http://localhost:5174/ — no more `BriefCase` syntax error.

## Files Changed

- `frontend/src/views/layout/LoginView.vue` (2 changes: template line 20, import line 187)
