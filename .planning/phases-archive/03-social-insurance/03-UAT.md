---
status: testing
phase: 03-social-insurance
source: 03-01-SUMMARY.md, 03-02-SUMMARY.md, 03-03-SUMMARY.md
started: 2026-04-07T09:00:00Z
updated: 2026-04-07T09:00:00Z
---

## Current Test
<!-- OVERWRITE each test - shows where we are -->

number: 2
name: 社保政策创建与查询
expected: |
  用 OWNER 角色调用 POST /social-insurance/policies 创建一条社保政策（如上海 2025），返回完整政策数据含五险一金比例。再调用 GET /social-insurance/policies 按城市查询，返回刚创建的政策。
awaiting: user response

## Tests

### 1. Cold Start Smoke Test
expected: 停止所有运行中的服务，从零启动服务器。服务器启动无报错，数据库迁移（SocialInsurancePolicy、SocialInsuranceRecord、ChangeHistory、Reminder 四张新表）完成，基础 API 调用返回正常。
result: issue
reported: "go run cmd/server/main.go exit status 1 — panic: nil pointer dereference in go-redis Ping(nil)"
severity: blocker

### 2. 社保政策创建与查询
expected: 用 OWNER 角色调用 POST /social-insurance/policies 创建一条社保政策（如上海 2025），返回完整政策数据含五险一金比例。再调用 GET /social-insurance/policies 按城市查询，返回刚创建的政策。
result: [pending]

### 3. 基数计算引擎
expected: 调用 POST /social-insurance/calculate，传入城市+月薪（如上海，月薪 10000）。返回 6 个险种的企业和个人缴费金额，基数被 clamp 到政策上下限内，金额四舍五入到分。
result: [pending]

### 4. 参保预览
expected: 调用 POST /social-insurance/enroll/preview，选择几个员工。返回每个员工按其城市自动匹配的社保基数和各险种金额预览。
result: [pending]

### 5. 批量参保
expected: 调用 POST /social-insurance/enroll，批量选择员工参保。成功创建参保记录（状态 active），已参保的员工被跳过并返回部分成功报告。变更历史自动记录 enroll 操作。
result: [pending]

### 6. 批量停缴
expected: 调用 POST /social-insurance/stop，选择 active 状态的参保记录停缴。状态更新为 stopped，记录结束月份。变更历史自动记录 stop 操作。
result: [pending]

### 7. 参保记录查询与筛选
expected: 调用 GET /social-insurance/records，返回参保记录列表。可按状态（active/stopped）和姓名筛选，支持分页。
result: [pending]

### 8. 员工自查询参保记录
expected: 用 MEMBER 角色调用 GET /social-insurance/my-records，返回该员工自己的参保记录（仅限本人数据）。
result: [pending]

### 9. 变更历史时间线
expected: 调用 GET /social-insurance/records/:id/history，返回该参保记录的所有变更事件（enroll/stop/base_adjust），包含变更前后的值。
result: [pending]

### 10. 缴费提醒生成与查询
expected: 定时任务在缴费到期前 3 天自动生成提醒（或手动触发）。调用 GET /social-insurance/reminders 返回提醒列表。调用 PUT /social-insurance/reminders/:id/dismiss 可关闭提醒。
result: [pending]

### 11. 参保材料 PDF 导出
expected: 调用 GET /social-insurance/records/:id/pdf，返回 PDF 文件。PDF 包含员工信息区、险种明细表和合计行。
result: [pending]

### 12. 缴费明细 Excel 导出
expected: 调用 GET /social-insurance/records/export，返回 xlsx 文件。Excel 包含 16 列（员工姓名、城市、参保月份、基数、各险种金额），有表头样式和求和公式。
result: [pending]

### 13. 离职停缴回调
expected: 员工办理离职时，社保模块自动收到回调，为该员工创建停缴提醒（stop_reminder）。提醒可在提醒列表中查到。
result: [pending]

### 14. 社保扣款查询（预留接口）
expected: 调用 GET /social-insurance/deduction，传入员工和月份，返回该员工的社保扣款明细。此接口供 Phase 5 工资核算模块调用。
result: [pending]

### 15. 全项目测试回归
expected: 运行 go test ./... -count=1，所有测试通过，无回归错误。社保模块测试覆盖政策 CRUD、基数计算、参保/停缴、变更历史、提醒等核心流程。
result: [pending]

## Summary

total: 15
passed: 0
issues: 0
pending: 15
skipped: 0

## Gaps

- truth: "服务器从零启动无报错，数据库迁移完成，健康检查返回正常"
  status: fixed
  reason: "rdb.Ping(nil) 传 nil context 导致 go-redis v9 panic，端口 5432/6379 被占用"
  severity: blocker
  test: 1
  root_cause: "main.go:68 调用 rdb.Ping(nil)，go-redis v9 要求非 nil context"
  artifacts:
    - path: "cmd/server/main.go"
      issue: "Ping(nil) → Ping(context.Background())"
    - path: "docker-compose.yml"
      issue: "端口 5432→5433, 6379→6380 避免冲突"
    - path: "config/config.yaml"
      issue: "端口同步更新"
  missing: []
