---
phase: 07-薪资管理增强
plan: 02
subsystem: api, database
tags: [go, shopspring/decimal, gorm, attendance-salary-integration, overtime-calculation, sick-leave]

# Dependency graph
requires:
  - phase: 06-考勤管理
    provides: AttendanceMonthly, Approval, AttendanceRule models and repository
  - phase: 07-01
    provides: SalaryAdjustment, PerformanceCoefficient, SickLeavePolicy models and services
provides:
  - AttendanceProvider interface and MonthlyAttendance struct for salary integration
  - Enhanced calculator functions (billing days, overtime pay, sick leave deduction)
  - CalculatePayroll with attendance-driven salary adjustments
affects: [07-03, 07-04]

# Tech tracking
tech-stack:
  added: []
  patterns: [interface-based cross-module dependency injection, decimal precision for financial calculations]

key-files:
  created:
    - internal/attendance/adapter.go
    - internal/salary/calculator_enhanced.go
  modified:
    - internal/salary/service.go
    - internal/salary/salary_test.go
    - cmd/server/main.go

key-decisions:
  - "Overtime classification derived from StartTime + RuleEngine (weekday/weekend/holiday) instead of non-existent overtime_type field"
  - "SickLeavePolicyService reused from salary package (not attendance) for coefficient lookup"
  - "Sick leave deduction = dailyWage * sickDays * (1 - coefficient), representing the shortfall from normal pay"

patterns-established:
  - "Cross-module adapter pattern: attendance.AttendanceProvider interface consumed by salary.Service via DI"
  - "Financial calculations use shopspring/decimal for all intermediate steps, round to 2 decimals only at output"

requirements-completed: [SAL-13, SAL-14, SAL-15, SAL-16]

# Metrics
duration: 5min
completed: 2026-04-18
---

# Phase 07 Plan 02: 薪资算法增强（考勤联动 + 病假 + 加班费）Summary

**AttendanceProvider 接口 + 考勤联动薪资核算：计薪天数/三档加班费/病假扣款集成到 CalculatePayroll，全链路 shopspring/decimal 精度控制**

## Performance

- **Duration:** 5 min
- **Started:** 2026-04-18T10:03:54Z
- **Completed:** 2026-04-18T10:09:13Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- AttendanceProvider 接口和 MonthlyAttendance 数据结构，跨模块（考勤->薪资）数据桥接
- 加班分档推导：根据加班审批的 StartTime 判断工作日/双休日/法定节假日，分别按 1.5x/2.0x/3.0x 计费
- 计薪天数 = 实际出勤 + 法定节假日 + 带薪假，据此调整基本工资
- 病假扣款 = 日工资 * 病假天数 * (1 - 病假系数)，系数从 SickLeavePolicy 按城市+工龄获取
- CalculatePayroll 全流程集成：考勤联动 -> 加班费 -> 病假扣款 -> 社保 -> 个税

## Task Commits

Each task was committed atomically:

1. **Task 1: AttendanceProvider Interface + Implementation** - `dae0870` (feat)
2. **Task 2: Calculator Enhanced + Service Integration** - `5259b8a` (feat)

## Files Created/Modified
- `internal/attendance/adapter.go` - AttendanceProvider interface, MonthlyAttendance struct, GetMonthlyAttendance implementation with overtime breakdown derivation
- `internal/salary/calculator_enhanced.go` - CalculateBillingDays, CalculateSalaryByBillingDays, CalculateSickLeaveWage, CalculateSickLeaveDeduction, CalculateOvertimePay (all with shopspring/decimal)
- `internal/salary/service.go` - Added attendanceProvider/sickLeavePolicySvc fields, enhanced CalculatePayroll with attendance integration
- `internal/salary/salary_test.go` - Fixed NewService call signatures for new parameters
- `cmd/server/main.go` - Wired AttendanceProvider and SickLeavePolicyService into salary service

## Decisions Made
- 加班类型不依赖不存在的 overtime_type 字段，而是利用加班审批的 StartTime 配合 RuleEngine.IsHoliday/isWeekend 推导分类（Rule 2: 添加缺失关键功能）
- 病假扣款公式为差额模式（正常日工资 - 病假日工资），而非直接工资替代，符合薪资核算逻辑
- 复用已创建的 SickLeavePolicyService 实例，避免 main.go 中重复初始化
- 病假系数默认城市"北京"，后续可从员工档案获取实际工作城市

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Overtime type classification without overtime_type field**
- **Found during:** Task 1 (AttendanceProvider implementation)
- **Issue:** Plan referenced `overtime_type` field on Approval model, but this field does not exist
- **Fix:** Derived overtime classification from StartTime date using RuleEngine.IsHoliday() and isWeekend() helper
- **Files modified:** internal/attendance/adapter.go
- **Verification:** Build passes, classifyOvertime function correctly categorizes by date
- **Committed in:** dae0870 (Task 1 commit)

**2. [Rule 3 - Blocking] Fixed salary_test.go NewService signature mismatch**
- **Found during:** Task 2 (build verification)
- **Issue:** NewService signature changed (added 2 parameters), test file still used old 8-arg signature
- **Fix:** Added nil, nil for attendanceProvider and sickLeavePolicySvc parameters in all test NewService calls
- **Files modified:** internal/salary/salary_test.go
- **Verification:** go test ./internal/salary/... passes
- **Committed in:** 5259b8a (Task 2 commit)

---

**Total deviations:** 2 auto-fixed (1 missing critical, 1 blocking)
**Impact on plan:** Both auto-fixes necessary for correctness and buildability. No scope creep.

## Issues Encountered
- None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- 薪资核算引擎完整支持考勤联动，07-03（工资条发送）和 07-04（前端）可直接使用
- 病假城市字段目前硬编码为"北京"，后续需从员工档案读取

---
*Phase: 07-薪资管理增强*
*Completed: 2026-04-18*

## Self-Check: PASSED

- internal/attendance/adapter.go: FOUND
- internal/salary/calculator_enhanced.go: FOUND
- internal/salary/service.go: FOUND
- internal/salary/salary_test.go: FOUND
- cmd/server/main.go: FOUND
- Commit dae0870: FOUND
- Commit 5259b8a: FOUND
