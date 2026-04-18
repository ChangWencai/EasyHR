---
phase: "06"
plan: "02"
status: complete
started: "2026-04-18"
completed: "2026-04-18"
requirements:
  - ATT-05
  - ATT-06
  - ATT-07
  - ATT-08
---

# Plan 06-02: 今日打卡实况 — Summary

## What was built

ClockRecord 打卡记录模型（work_date 与 clock_time 分离 per D-14），ClockLive 打卡实况 API，假勤统计 API，以及前端今日打卡实况页面。

## Key Changes

### Backend
- `internal/attendance/model.go`: ClockRecord 模型、AttendanceManualStats 手动修正模型
- `internal/attendance/dto.go`: ClockRecordResponse、ClockLiveResponse、LeaveStatsResponse DTOs
- `internal/attendance/repository.go`: ClockRecord CRUD、ManualStats UPSERT、ListAllActiveEmployees（JOIN departments）
- `internal/attendance/service.go`: GetClockLive（全员打卡映射 + 迟到判断）、CreateClockRecord、GetLeaveStats、UpdateLeaveStats
- `internal/attendance/handler.go`: clock-live / clock-records / leave-stats 路由注册

### Frontend
- `frontend/src/views/attendance/ClockLive.vue`: 打卡实况页面（表格 + 假勤 Popover + 手动修正弹窗 + 邀请点签）
- `frontend/src/components/attendance/AttendanceStatsCard.vue`: 4张统计卡片组件
- `frontend/src/api/attendance.ts`: ClockLive/LeaveStats API 类型和接口
- `frontend/src/router/index.ts`: /attendance/clock-live 路由
- `frontend/src/views/layout/AppLayout.vue`: 侧边栏菜单项

## Deviations

- GetClockLive 使用 clockPair 结构同时存储上下班记录，修复了原计划 map 只存一条记录的问题
- Repository 使用 EmployeeBrief + JOIN departments 获取部门名称，避免跨包依赖
- ManualStats UPSERT 使用先查后写模式（非 clause.OnConflict），与项目其他模块保持一致

## Self-Check: PASSED

- `go build ./internal/attendance/...` 编译通过
- ClockRecord.work_date 和 clock_time 字段分离（D-14）
- 全员打卡列表含未打卡员工（ListAllActiveEmployees）
- 打卡时间颜色标注（正常绿/迟到黄/缺勤红）
