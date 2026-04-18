---
phase: "06"
plan: "04"
status: complete
type: execute
wave: 3
---

# 06-04 出勤月报 -- SUMMARY

## What was built

出勤月报功能完整实现，支持统计视图和格子视图双模式切换，含 Excel 导出。

### Backend
- AttendanceMonthly 预计算模型（已在 model.go 中定义）
- ListAttendanceMonthly/UpsertAttendanceMonthly/GetDailyClockRecords 仓库方法
- GetMonthlyReport/GetDailyRecords/ExportMonthlyExcel 服务方法
- /attendance/monthly、/attendance/monthly/export、/attendance/daily-records 三个 API 端点
- Excel 导出使用 excelize 库，含表头样式、合计行

### Frontend
- AttendanceMonthly.vue: 统计卡片 + 月报表格 + Drawer 打卡详情 + Excel 导出
- 视图切换：统计视图（默认）/ 格子视图
- 新增类型：MonthlyReportItem, MonthlyStats, MonthlyReportResponse, DailyRecord, DailyRecordsResponse
- 路由注册：/attendance/monthly

## Self-Check: PASSED

- [x] 管理员可查看出勤月报（默认最近一个月，可选定年月）
- [x] 显示实际出勤/应出勤/加班时长
- [x] 管理员可查看员工每日打卡记录详情
- [x] 支持 Excel 格式导出

## key-files

### created
- frontend/src/views/attendance/AttendanceMonthly.vue

### modified
- internal/attendance/repository.go
- internal/attendance/service.go
- internal/attendance/handler.go
- frontend/src/api/attendance.ts
- frontend/src/router/index.ts
