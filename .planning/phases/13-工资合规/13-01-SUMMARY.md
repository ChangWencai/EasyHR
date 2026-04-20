# Phase 13 Plan 01 Summary — 后端工资条确认回执

## 完成状态：✅ 已完成

## 实现内容

### D-13-01: 员工确认按钮（`/salary/slip/:token/confirm`）

- **Handler** (`handler.go:559-575`): 新增 `ConfirmSlip` 路由，POST `/salary/slip/:token/confirm`（公开，无需认证）
- **Service** (`slip.go:327-351`): `ConfirmSlip(token, clientIP)` 方法，更新 `confirmed_at` 和 `confirmed_ip`
- **Model** (`model.go`): PayrollSlip 新增 `ConfirmedAt *time.Time`、`ConfirmedIP string`、`ExpiresAt time.Time` 字段（AutoMigrate 自动添加列）
- **DTO** (`dto.go`): SlipDetailResponse 新增 `confirmed_at` 格式化返回

### D-13-03: H5 页面显示确认状态

- `SlipDetailResponse` 已在 D-13-01 中一并实现（`confirmed_at` 字段格式化）
- 前端 SalarySlipH5.vue 显示逻辑见 Plan 02

### D-13-11: 发送记录显示员工确认状态

- **Repository** (`repository.go:333-343`): `FindSlipSendLogs` 使用 `LEFT JOIN payroll_slips` 联查 `confirmed_at`
- **Model** (`sick_leave_policy_model.go`): SalarySlipSendLog 新增 `ConfirmedAt *time.Time`（`gorm:"-"` 跳过写入，SELECT 别名填充）
- **Service** (`slip_send_service.go:271-310`): `GetSlipLogsWithConfirmation` 方法

### D-13-08/09: asynq Worker 每日提醒

- **Handler** (`slip_send_service.go:334-347`): `HandleRemindUnconfirmedTask` — asynq worker handler
- **Service** (`slip_send_service.go:351-387`): `processRemindUnconfirmed` — 查询未确认工资条，创建 TodoCenter 通知（幂等 SourceType=salary_unconfirmed）
- **Scheduler** (`scheduler.go`): `SalaryScheduler` — gocron 每日 9:00 CST 查询所有企业，为每个企业入队 `salary:remind-unconfirmed` 任务
- **注册**: main.go 中 `asynqMux.HandleFunc(salary.TypeRemindUnconfirmed, salary.HandleRemindUnconfirmedTask)` + salary scheduler Start()

### 依赖注入

- `salary.Service` 新增 `todoSvc TodoSvc` 字段（避免循环依赖）
- `salary.NewService` 新增 `todoSvc` 参数
- `main.go`: `salarySvc` 构造传入 `todoSvcForDI`

## 验证

- `go build ./...` ✅
- `go test ./internal/salary/...` ✅
