---
phase: 07-薪资管理增强
plan: "04"
subsystem: ui
tags: [vue3, element-plus, golang, gin, redis, excelize]

# Dependency graph
requires:
  - phase: 07-01
    provides: adjustment/preview APIs, performance APIs, SalaryDashboard component
  - phase: 07-03
    provides: TaxUpload, SalarySlipSend, SalarySlipH5 components
provides:
  - SalaryTool.vue 扩展至 6 Tab（数据看板/薪资模板/工资核算/调薪管理/导出/薪资列表）
  - SalaryAdjustment.vue：员工调薪 + 部门普调（含预览和提交）
  - PerformanceCoefficient.vue：绩效系数滑块批量设置
  - SalaryList.vue：薪资列表（含筛选、锁定解锁、导出选项）
  - GET /salary/list API（含部门/姓名筛选 + 分页）
  - GET /salary/export 增强（含 include_details 参数导出税前明细）
  - POST /salary/unlock + POST /salary/unlock/send-code（Redis 验证码解锁 confirmed/paid）
affects:
  - Phase 07 薪资管理增强

# Tech tracking
tech-stack:
  added: []
  patterns:
    - el-slider 绩效系数实时计算
    - 调薪 INSERT ONLY 预览/提交双阶段
    - SMS 验证码解锁 D-SAL-DATA-01
    - Excel 导出可选明细列（动态表头）

key-files:
  created:
    - frontend/src/views/tool/SalaryAdjustment.vue
    - frontend/src/views/tool/PerformanceCoefficient.vue
    - frontend/src/views/tool/SalaryList.vue
    - internal/salary/salary_list_service.go
    - internal/salary/salary_list_handler.go
    - internal/salary/salary_unlock_service.go
    - internal/salary/salary_unlock_handler.go
  modified:
    - frontend/src/views/tool/SalaryTool.vue
    - frontend/src/api/salary.ts
    - internal/salary/excel.go
    - cmd/server/main.go

key-decisions:
  - "部门多选用 __all__ sentinel value 模拟全选，toggleSelectAllDepts 处理"
  - "salary_list_handler 和 salary_unlock_handler 使用 authMiddleware 统一鉴权"
  - "解锁降级模式：Redis 不可用时打印 fallback code 到日志，不阻塞解锁流程"
  - "salaryListHandler 和现有 PayrollHandler 使用不同路由前缀（/salary/list vs /salary/payroll）避免冲突"

patterns-established:
  - "调薪双 Tab 表单：员工调薪 / 部门普调，共用调整类型/方式/生效月份字段"
  - "dirty tracking 批量保存：仅提交系数 != 100% 的行"

requirements-completed: [SAL-05, SAL-06, SAL-07, SAL-11, SAL-12, SAL-16, SAL-19]

# Metrics
duration: 18min
completed: 2026-04-18
---

# Phase 07-04: 调薪管理 UI + 薪资列表导出增强 + 解锁

**6 Tab 薪资工具页面（数据看板/模板/核算/调薪/导出/列表）+ 调薪管理/绩效系数/薪资列表 Vue 组件 + salary list API + confirmed/paid 解锁服务**

## Performance

- **Duration:** 18 min
- **Started:** 2026-04-18T10:20:00Z
- **Completed:** 2026-04-18T10:38:00Z
- **Tasks:** 4 (Task 1-4 合并为 3 个 commit，后端 Task 5-6 合并为 1 个 commit)
- **Files modified:** 12 个（4 前端修改 + 3 新前端 + 4 新后端 + 1 main.go）

## Accomplishments

- SalaryTool.vue 从 5 Tab 扩展为 7 Tab，新增数据看板、调薪管理、薪资列表
- SalaryAdjustment.vue：员工调薪（搜索选择员工 + 预览 + 提交）和部门普调（全选部门 + 批量预览 + 提交）
- PerformanceCoefficient.vue：el-slider 绩效系数（0-200%，步进 5%）实时计算实际绩效工资，dirty tracking 批量保存
- SalaryList.vue：月份/部门/姓名筛选 + 状态标签 + confirmed/paid 行锁定图标 + 解锁弹窗（含 SMS 验证码）+ 导出选项（当前页/含税前明细）
- GET /salary/list API：JOIN employees/departments 返回部门名称，支持分页和筛选
- GET /salary/export 增强：include_details=true 时导出所有 PayrollItem 明细列（动态表头）
- POST /salary/unlock + POST /salary/unlock/send-code：Redis 存储 6 位验证码，5 分钟有效期，解锁后状态回退 calculated

## Task Commits

1. **Task 1: SalaryTool.vue 扩展 6 Tab** - `f457f68` (feat)
2. **Tasks 2+3: SalaryAdjustment + PerformanceCoefficient** - `88593ba` (feat)
3. **Task 4: SalaryList（含导出和解锁弹窗）** - `76a8d58` (feat)
4. **Tasks 5+6: Backend list/export/unlock** - `d9da6ae` (feat)

## Files Created/Modified

- `frontend/src/views/tool/SalaryTool.vue` - 扩展 7 Tab，默认 dashboard，整合新组件
- `frontend/src/views/tool/SalaryAdjustment.vue` - 员工调薪 + 部门普调，含预览和提交
- `frontend/src/views/tool/PerformanceCoefficient.vue` - 绩效系数 el-slider + dirty tracking 批量保存
- `frontend/src/views/tool/SalaryList.vue` - 薪资列表筛选 + 状态 + 解锁弹窗 + 导出选项
- `frontend/src/api/salary.ts` - 添加 previewAdjustment/getSalaryList/sendUnlockCode/unlockRecord/exportWithDetails 方法
- `internal/salary/salary_list_service.go` - ListSalaryRecords（JOIN 部门名称，支持筛选分页）
- `internal/salary/salary_list_handler.go` - GET /salary/list + GET /salary/export
- `internal/salary/salary_unlock_service.go` - UnlockPayroll（Redis 验证码 + 状态回退 calculated）
- `internal/salary/salary_unlock_handler.go` - POST /salary/unlock + POST /salary/unlock/send-code
- `internal/salary/excel.go` - ExportPayrollExcelWithDetails（动态表头，支持明细列）
- `cmd/server/main.go` - 注册 SalaryListHandler 和 UnlockHandler 路由

## Decisions Made

- "部门多选 __all__ sentinel value 实现全选，避免 el-select 无原生全选支持"
- "salaryListHandler 和 PayrollHandler 路由分开（/salary/list vs /salary/payroll），避免重复注册"
- "解锁降级：Redis 不可用时打印 fallback code 到日志，不阻塞流程"
- "department_api.list() 无参数版本，直接 .map 取 id/name 填充下拉"

## Deviations from Plan

**None - plan executed exactly as written**

## Issues Encountered

- `cmd/server` 目录被 `.gitignore` 中的 `server` pattern 匹配（编译产物二进制），需要 `git add -f` 强制添加 main.go
- `ExportPayrollExcelWithDetails` 列数字格式设置中 `lastColStr` 变量 scope 错误（定义为 `rune` 与 `string` 混用），重构为 `totalCols` 和 `excelize.CoordinatesToCellName` 计算列名
- `salary_list_service.go` 和 `salary_list_handler.go` 意外 import 了未使用的 middleware 包，清理后编译通过

## Next Phase Readiness

- 所有 07-04 计划功能已完成，前端 6 Tab 和后端 3 个新端点（list/export/unlock）均可用
- 下一步：07-04 完成，Phase 07 所有 4 个计划均已完成

---
*Phase: 07-薪资管理增强 04*
*Completed: 2026-04-18*
