---
phase: "12-考勤合规报表"
plan: "01"
subsystem: "attendance"
tags: ["backend", "compliance", "reports", "attendance"]
dependency_graph:
  requires: []
  provides:
    - "GET /api/v1/attendance/compliance/overtime"
    - "GET /api/v1/attendance/compliance/leave"
    - "GET /api/v1/attendance/compliance/anomaly"
    - "GET /api/v1/attendance/compliance/monthly"
    - "GET /api/v1/attendance/compliance/monthly/export"
  affects:
    - "Phase 12 Plan 02 (frontend)"
tech_stack:
  added:
    - "excelize/v2 (Excel export)"
  patterns:
    - "RuleEngine 复用 (ClassifyOvertimeCategory)"
    - "GORM Scope org_id 隔离"
    - "0.5h 取整 (roundHalf)"
    - "异常阈值: late>3 OR absent>1"
    - "部门多选 __all__ sentinel"
key_files:
  created: []
  modified:
    - "internal/attendance/rule_engine.go"
    - "internal/attendance/dto.go"
    - "internal/attendance/repository.go"
    - "internal/attendance/service.go"
    - "internal/attendance/handler.go"
decisions:
  - "加班分类: holiday(法定节假日)/weekday(工作日延时)/weekend(周末) 按 ClassifyOvertimeCategory 划分"
  - "年假额度: AnnualLeaveQuota 表独立建模, 管理员配置, 按 year 过滤"
  - "异常高亮: late_count > 3 OR absent_days > 1 (D-12-10)"
  - "Excel 导出限制 5000 行防止内存耗尽 (D-12-03 威胁缓解)"
metrics:
  duration: "~3min"
  completed: "2026-04-20"
  tasks_completed: 5
  files_modified: 5
---

# Phase 12 Plan 01: 考勤合规报表后端 API

## One-liner

考勤合规报表后端 API: 加班/请假/异常/月度汇总报表, 支持部门筛选, Excel 导出。

## Commits

| # | Task | Commit | Files |
|---|------|--------|-------|
| 1 | ClassifyOvertimeCategory 方法 | `28c0b33` | rule_engine.go |
| 2 | DTO 类型定义 | `28c0b33` | dto.go |
| 3 | Repository 查询方法 | `c59b9f4` | repository.go |
| 4 | Service 业务逻辑 | `68e55e7` | service.go |
| 5 | Handler 路由注册 | `bb6d9dc` | handler.go |
| fix | normalStyle 类型修复 | `8074e6c` | service.go |

## Deviations from Plan

None - plan executed exactly as written.

## Artifacts

### API Routes (5 endpoints)

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/attendance/compliance/overtime | 加班统计报表 (holiday/weekday/weekend 分类) |
| GET | /api/v1/attendance/compliance/leave | 请假合规报表 (年假额度/已用/剩余) |
| GET | /api/v1/attendance/compliance/anomaly | 考勤异常报表 (迟到/早退/缺勤, 异常高亮) |
| GET | /api/v1/attendance/compliance/monthly | 月度综合合规报表 |
| GET | /api/v1/attendance/compliance/monthly/export | 月度考勤汇总 Excel 导出 |

### Key Implementation Details

**加班分类 (D-12-03):**
- `RuleEngine.ClassifyOvertimeCategory(start, end)` 基于 `StartTime` 判断
- holiday: `IsHoliday()` 为 true
- weekend: `Weekday == 0 (周日) || 6 (周六)`
- weekday: 其余工作日

**0.5h 取整 (D-12-04):**
- 所有时长字段使用 `roundHalf()` (原 service.go 已定义)

**异常阈值 (D-12-10):**
- `late_count > 3 || absent_days > 1` → `is_anomaly = true`
- Excel 导出时异常行红色高亮

**AnnualLeaveQuota (D-12-08):**
- 独立模型, 表名 `attendance_annual_leave_quotas`
- 复合唯一索引: `(employee_id, year)`
- 管理员通过 Plan 02/03 配置额度

### Verification

```
go build ./cmd/server/...  # OK
go vet ./internal/attendance/...  # OK
```

## Self-Check

- [x] ClassifyOvertimeCategory exists in rule_engine.go
- [x] 5 DTO types in dto.go (OvertimeItem, LeaveItem, AnomalyItem, MonthlyComplianceItem, ComplianceReportRequest)
- [x] AnnualLeaveQuota model defined in dto.go
- [x] 5 repository methods (ListEmployeesByOrgWithDept, ListClockRecordsByMonth, ListApprovalsByMonth, GetAnnualLeaveQuotas, ListMonthlyAttendanceForCompliance)
- [x] 5 service methods (GetComplianceOvertime, GetComplianceLeave, GetComplianceAnomaly, GetComplianceMonthly, ExportComplianceMonthlyExcel)
- [x] 5 handler methods (GetComplianceOvertime, GetComplianceLeave, GetComplianceAnomaly, GetComplianceMonthly, ExportComplianceMonthly)
- [x] 5 routes registered under /attendance/compliance/
- [x] go build succeeds
- [x] go vet passes
