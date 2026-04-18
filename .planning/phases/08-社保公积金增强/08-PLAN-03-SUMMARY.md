---
phase: 08
plan: 03
subsystem: social-insurance-enhancement
tags: [frontend, vue3, element-plus, dialog, table, status-tags]
dependency_graph:
  requires: [08-PLAN-01, 08-PLAN-02]
  provides: [StopDialog, SIRecordsTable, SIDetailDialog]
  affects: [SITool.vue, SIRecordsTable.vue]
tech_stack:
  added: []
  patterns: [v-model dialog pattern, remote search select, row-class-name, el-table show-summary]
key_files:
  created:
    - frontend/src/components/socialinsurance/StopDialog.vue
    - frontend/src/views/socialinsurance/SIRecordsTable.vue
    - frontend/src/components/socialinsurance/SIDetailDialog.vue
  modified: []
decisions:
  - StopDialog 使用 v-model 双向绑定 + props (employeeId/employeeName) 预填模式
  - SIRecordsTable 导出弹窗复用 SalaryList.vue 当前页/全部数据模式
  - SIDetailDialog 使用 el-table show-summary + getSummary 计算合计行
  - 减员按钮仅对 normal/pending 状态显示
  - 公积金封存日期通过 watch 自动同步转出日期
metrics:
  duration: 3min
  tasks: 3
  files: 3
  completed: "2026-04-19"
---

# Phase 08 Plan 03: 减员弹窗 + 参保记录增强 + 五险分项弹窗 Summary

减员弹窗（姓名搜索 + 终止月份校验 + 转出规则提示 + 危险确认）、参保记录列表（5色状态标签 + 缴费渠道列 + 欠缴行高亮）、五险分项详情弹窗（6险x2列 + 其他缴费 + 合计行）

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | StopDialog.vue 减员弹窗 | 748658f | frontend/src/components/socialinsurance/StopDialog.vue |
| 2 | SIRecordsTable.vue 参保记录列表 | b3d71f0 | frontend/src/views/socialinsurance/SIRecordsTable.vue |
| 3 | SIDetailDialog.vue 五险分项弹窗 | 7a27ef3 | frontend/src/components/socialinsurance/SIDetailDialog.vue |

## Key Details

### Task 1: StopDialog.vue
- 员工搜索使用 filterable + remote el-select，调用 /api/v1/employees/search
- 终止月份默认当月，disableStopDate 禁止早于当月选择
- 减员原因三选一：跳槽(job_change) / 退休(retirement) / 其他(other)
- 转出生效规则 Tooltip 展示三档规则（5日前/5-25日/25日后）
- 公积金封存日期通过 watch 自动同步转出日期
- 提交前 ElMessageBox.confirm 危险操作确认
- 支持从行数据预填 employeeId/employeeName

### Task 2: SIRecordsTable.vue
- 5色状态标签：正常(success)/待缴(warning)/欠缴(danger)/已转出(info)/未转出(custom blue #4F6EF7)
- 缴费渠道列：自主缴费/代理新客/代理已合作
- 欠缴行高亮背景色 #FF563010
- 操作列：详情按钮 + 减员按钮（仅 normal/pending 状态显示）
- 导出弹窗复用 SalaryList.vue 模式（当前页/全部数据）
- 金额格式化使用 toLocaleString('zh-CN')

### Task 3: SIDetailDialog.vue
- 6险种表格：养老保险/医疗保险/失业保险/工伤保险/生育保险/住房公积金
- 单位缴纳 + 个人缴纳两列，使用 computed insuranceItems
- 合计行通过 el-table show-summary + getSummary 自动计算
- 其他缴费区块（滞纳金/残保金/漏缴/补缴）使用 el-descriptions
- 仅读弹窗，只有"关闭"按钮
- 数据从 /api/v1/socialinsurance/records/:id/detail 加载

## Deviations from Plan

None - plan executed exactly as written.

## Self-Check: PASSED

- FOUND: StopDialog.vue
- FOUND: SIRecordsTable.vue
- FOUND: SIDetailDialog.vue
- FOUND: commit 748658f
- FOUND: commit b3d71f0
- FOUND: commit 7a27ef3
