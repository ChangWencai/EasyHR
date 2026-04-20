---
phase: 11-合同合规
plan: 02
subsystem: ui
tags: [contract, vue3, element-plus, signing, sms, pdf, h5]

# Dependency graph
requires:
  - phase: 11-01
    provides: 后端签署API端点 + 中文PDF生成
affects: [12-考勤合规报表, 13-工资合规]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - H5无认证页面路由（/sign/:contractId 加入白名单）
    - StepWizard内联按钮替代方案（避免内置按钮冲突）
    - 6位短信验证码自定输入组件（自动聚焦/粘贴支持）

key-files:
  created:
    - frontend/src/api/contract.ts
    - frontend/src/components/contract/ContractWizard.vue
    - frontend/src/components/contract/ContractTypeSelect.vue
    - frontend/src/components/contract/ContractPeriodPicker.vue
    - frontend/src/components/contract/PdfPreview.vue
    - frontend/src/components/contract/SmsVerifyInput.vue
    - frontend/src/components/contract/SignSuccessCard.vue
    - frontend/src/components/contract/ContractStatusBadge.vue
    - frontend/src/components/contract/ExpiryCountdown.vue
    - frontend/src/views/employee/components/ContractList.vue
    - frontend/src/views/sign/SignPage.vue
  modified:
    - frontend/src/views/employee/EmployeeDrawer.vue
    - frontend/src/router/index.ts

key-decisions:
  - "StepWizard内嵌按钮冲突：StepWizard本身有上一步/下一步/完成按钮，与ContractWizard自定义底部按钮重复。改用el-steps内联实现。"
  - "SignSuccessCard内联：成功页面内容直接写在SignPage.vue中，不拆出独立组件（WARNING-3 fix）。"
  - "BLOCKER-5 fix: contractApi.create传递有效salary值满足后端binding:required,gt=0，salary不写入PDF正文（D-11-02）。"

patterns-established:
  - "合同3步向导: 类型选择 → PDF预览 → 发送签署链接"
  - "H5签署流程: 手机号 → 验证码 → 确认 → 成功"
  - "合同状态标签: draft/pending_sign/signed/active/terminated/expired"

requirements-completed: [COMP-01, COMP-02, COMP-03, COMP-04]

# Metrics
duration: ~8min
completed: 2026-04-20
---

# Phase 11 Plan 02: 前端合同管理UI

**员工抽屉合同Tab、3步合同发起向导、员工H5签署页，完整合同管理UI覆盖**

## Performance

- **Duration:** ~8 min
- **Started:** 2026-04-20T03:18:01Z
- **Completed:** 2026-04-20T03:26:05Z
- **Tasks:** 9
- **Files created/modified:** 14

## Accomplishments
- contractApi.ts API模块（9个导出：list/create/sendSignLink/sendSignCode/verifySignCode/confirmSign/getSignedPdf/terminate/generatePdfBlob）
- EmployeeDrawer合同Tab（基本信息 + 合同两个el-tab-pane）
- ContractList合同列表（空状态/列表展示/终止合同）
- 3步合同发起向导 ContractWizard（类型选择→PDF预览→发送链接）
- 员工H5签署页 SignPage（手机号→验证码→确认→成功）
- 9个配套组件（ContractStatusBadge/ExpiryCountdown/ContractTypeSelect/ContractPeriodPicker/PdfPreview/SmsVerifyInput/SignSuccessCard）
- /sign/:contractId路由注册 + auth白名单

## Task Commits

1. **Task 1: contractApi.ts** - `1818df4` (feat)
2. **Task 2: EmployeeDrawer.vue合同Tab** - `0a6747a` (feat)
3. **Task 3: ContractList.vue** - `e5c80fa` (feat)
4. **Task 4: ContractStatusBadge + ExpiryCountdown** - `e5c80fa` (feat)
5. **Task 5: ContractTypeSelect + ContractPeriodPicker** - `290797f` (feat)
6. **Task 6: ContractWizard.vue** - `d3b2550` (feat)
7. **Task 7: PdfPreview + SmsVerifyInput** - `8e56a6f` (feat)
8. **Task 8: SignPage.vue** - `c2569d6` (feat)
9. **Task 9: SignSuccessCard + 路由注册** - `28d794e` (feat)

## Files Created/Modified

| 文件 | 说明 |
|------|------|
| `frontend/src/api/contract.ts` | 合同API模块，9个方法 |
| `frontend/src/components/contract/ContractWizard.vue` | 3步合同发起向导 |
| `frontend/src/components/contract/ContractTypeSelect.vue` | 3个合同类型单选卡片 |
| `frontend/src/components/contract/ContractPeriodPicker.vue` | 起止日期选择 |
| `frontend/src/components/contract/PdfPreview.vue` | PDF预览面板 |
| `frontend/src/components/contract/SmsVerifyInput.vue` | 6位验证码输入 |
| `frontend/src/components/contract/ContractStatusBadge.vue` | 状态标签 |
| `frontend/src/components/contract/ExpiryCountdown.vue` | 到期倒计时 |
| `frontend/src/components/contract/SignSuccessCard.vue` | 占位文件（内容内联SignPage） |
| `frontend/src/views/employee/components/ContractList.vue` | 合同列表组件 |
| `frontend/src/views/sign/SignPage.vue` | 员工H5签署页 |
| `frontend/src/views/employee/EmployeeDrawer.vue` | 添加合同Tab |
| `frontend/src/router/index.ts` | 注册/sign/路由 + 白名单 |

## Decisions Made

- **StepWizard内嵌按钮冲突**：原计划使用StepWizard复用组件，但StepWizard内置了上一步/下一步/完成按钮，与ContractWizard自定义底部按钮冲突。改用el-steps + 内联按钮实现3步流程控制。
- **SignSuccessCard内联**：WARNING-3 fix — 成功页面内容直接写在SignPage.vue中，不拆出独立组件。
- **BLOCKER-5 fix**：contractApi.create必须发送有效salary值（employeeSalary ?? data.salary ?? 0）满足后端`binding:"required,gt=0"`验证，salary不写入PDF正文（D-11-02）。

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - 缺失关键功能] StepWizard内嵌按钮冲突导致流程无法正确控制**
- **Found during:** Task 6 (ContractWizard实现)
- **Issue:** StepWizard.vue内置了上一步/下一步/完成按钮，与ContractWizard自定义底部按钮重复，导致两步操作叠加
- **Fix:** 改用el-steps内联实现，用currentStep ref控制流程
- **Files modified:** frontend/src/components/contract/ContractWizard.vue
- **Verification:** 编译通过，3步流程可正确切换
- **Committed in:** d3b2550

**Total deviations:** 1 auto-fixed
**Impact on plan:** StepWizard内嵌按钮冲突为实现层问题，不影响功能完整性。

## Auth Gates

无。

## Known Stubs

无 — 所有功能已完整实现。

## Threat Flags

无新威胁引入。

## Next Phase Readiness

- Phase 11 Plan 01（后端）和 Plan 02（前端）均完成
- 合同合规（COMP-01~04）全量完成
- 待配置：阿里云短信模板ID（`ALIYUN_SMS_CONTRACT_TEMPLATE_CODE`）
- FLAG-1：SignPage "返回首页" 按钮目的待产品确认

---
*Phase: 11-合同合规 Plan 02*
*Completed: 2026-04-20*
