---
phase: 03-social-insurance
plan: 03
status: complete
started: "2026-04-07T08:00:00Z"
completed: "2026-04-07T08:45:00Z"
---

# Plan 03-03: 缴费提醒 + PDF/Excel导出 + 离职回调集成

## Objective
实现缴费到期自动提醒（gocron 定时任务）、参保材料 PDF 导出、缴费凭证 Excel 导出、离职停缴回调集成，以及薪资变动基数调整预留接口。

## What Was Built

### Models
- **Reminder**: 提醒记录模型（payment_due/stop_reminder/base_adjust 三种类型）

### Repositories
- **ReminderRepository**: 提醒记录 CRUD + 去重检查 + 按企业分组查询

### Scheduler
- **StartScheduler**: gocron 定时任务启动器
  - 每日 08:00 CST 扫描
  - 缴费前 3 天生成提醒（daysUntilDue in [0,3]）
  - 按企业汇总 (D-10)
  - 支持 Redis 分布式锁（开发环境自动降级）
  - 检查重复提醒，避免重复生成

### Service Extensions
- `CheckPaymentDueReminders`: 定时任务扫描逻辑（D-09/D-10/D-11）
- `CreateStopReminder`: 离职后创建停缴提醒（D-07）
- `SuggestBaseAdjustment`: 基数调整预留接口（D-13/SOCL-06）
- `GenerateEnrollmentPDF`: 生成参保材料 PDF（SOCL-02）
- `ExportPaymentDetailExcel`: 导出缴费明细 Excel（SOCL-05）

### PDF Generation
- EnrollmentPDFData 填充结构
- fpdf 实现：员工信息区 + 险种明细表 + 合计行
- Helvetica 字体（V1.0 英文/拼音标注）

### Excel Export
- 16列：员工姓名、城市、参保月份、基数、各险种金额（企业/个人）、合计
- 表头样式、求和公式、自适应列宽
- xlsx 格式，Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet

### Handler Routes (新增 4 个)
- `GET /social-insurance/reminders` — 查询提醒列表
- `PUT /social-insurance/reminders/:id/dismiss` — 关闭提醒
- `GET /social-insurance/records/export` (owner, admin) — 导出 Excel
- `GET /social-insurance/records/:id/pdf` (owner, admin) — 生成参保材料 PDF

### Integration
- employee.OffboardingService 集成 `siSvc.OnEmployeeResigned` 回调
- CompleteOffboarding 触发停缴提醒
- main.go: 完整依赖注入（ReminderRepository + 社保作为 Offboarding 事件处理器）

## Key Decisions
- 提醒去复用 `FindByTypeAndRecordID` 检查（避免重复）
- 缴费截止日固定为每月 15 日（D-11）
- 汇总提醒使用虚拟 `record_id = 0`（不关联具体记录）
- 定时任务在 Redis 不可用时自动降级为单机模式
- 薪资变动基数调整预留接口（偏差 >10% 触发）

## Commits
- `a87b005`: feat(03-03): 缴费提醒、PDF/Excel导出、离职回调集成

## Test Results
- `go test ./internal/socialinsurance/... -count=1`: PASS
- `go build ./cmd/server/`: SUCCESS
- `go test ./... -count=1`: ALL PASS (no regressions)
