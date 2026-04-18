---
phase: 08-社保公积金增强
verified: 2026-04-19T12:00:00Z
status: gaps_found
score: 9/10 must-haves verified
overrides_applied: 0
gaps:
  - truth: "管理员可点击详情打开五险分项弹窗（6险x2列 + 其他缴费 + 合计）"
    status: partial
    reason: "SIDetailDialog.vue API 路径与后端路由不匹配。前端调用 /api/v1/socialinsurance/records/${id}/detail，后端注册 /api/v1/social-insurance/monthly-records/:id。两处差异：(1) socialinsurance vs social-insurance (2) records/:id/detail vs monthly-records/:id"
    artifacts:
      - path: "frontend/src/components/socialinsurance/SIDetailDialog.vue"
        issue: "第180行 API 路径 /api/v1/socialinsurance/records/${recordId}/detail 与后端路由不匹配"
    missing:
      - "将 SIDetailDialog.vue 第180行 API 路径改为 /api/v1/social-insurance/monthly-records/${recordId}，或后端新增匹配路由"
---

# Phase 08: 社保公积金增强 Verification Report

**Phase Goal:** 管理员获得社保公积金数据洞察，增减员流程优化，缴费渠道和欠缴状态自动管理
**Verified:** 2026-04-19
**Status:** gaps_found
**Re-verification:** No -- initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | 管理员可在社保数据看板看到应缴总额/单位部分/个人部分/欠缴金额，均带环比百分比 | VERIFIED | SIDashboard.vue 4卡片(#4F6EF7/#FF5630) + dashboard_service.go errgroup 并发4指标 + 环比计算 |
| 2 | SIMonthlyPayment 表记录月度缴费状态，不与参保生命周期 conflate | VERIFIED | model.go SIMonthlyPayment 结构体 + PaymentStatus 5种状态 + decimal 金额字段 |
| 3 | 缴费渠道写入成功，代理缴费 webhook 可更新状态 | VERIFIED | handler.go /payment-callback 路由 + /confirm-payment 路由 + PaymentCallback handler |
| 4 | 管理员可在增员弹窗按姓名检索，设置起始月份、城市、社保基数、公积金比例和基数 | VERIFIED | EnrollDialog.vue remote-method 搜索 + 全字段(el-date-picker/el-input-number) + POST /enroll/single |
| 5 | 参保操作 Tab 顶部显示红色欠缴横幅 | VERIFIED | SITool.vue overdue-banner + WarningFilled + overdueItems 条件渲染 + dismissBanner |
| 6 | 管理员可在减员弹窗按姓名检索员工，设置终止月份、原因、转出日期 | VERIFIED | StopDialog.vue 姓名 remote 搜索 + 终止月份(date-picker) + 减员原因三选一 + 转出生效规则 tooltip |
| 7 | 减员弹窗显示转出生效规则提示 | VERIFIED | StopDialog.vue 第75行 转出生效规则三档提示(5日前/5-25日/25日后) |
| 8 | 参保记录列表显示5种状态标签 + 缴费渠道列 + 详情弹窗 | VERIFIED | SIRecordsTable.vue statusTagType 5色 + statusLabelMap + 缴费渠道列 + openDetail + openStopDialog |
| 9 | 点击详情打开五险分项弹窗（6险x2列 + 其他缴费 + 合计） | PARTIAL | SIDetailDialog.vue 组件完整（养老保险~住房公积金 + toLocaleString + el-descriptions 其他缴费 + 合计行），但 API 路径与后端不匹配 |
| 10 | 参保记录列表支持 Excel 格式导出（当前页/含五险分项） | VERIFIED | excel.go ExportSIRecordsWithDetails 22列 + handler.go /records/export 路由 + SIRecordsTable.vue showExportDialog + blob 下载 |

**Score:** 9/10 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/socialinsurance/model.go` | SIMonthlyPayment 模型 | VERIFIED | 7KB, SIMonthlyPayment struct + PaymentStatus 5种 + decimal 金额 |
| `internal/socialinsurance/dashboard_service.go` | DashboardService 4指标 | VERIFIED | 5KB, GetDashboard + errgroup 并发 |
| `internal/socialinsurance/repository.go` | SIMonthlyPayment CRUD | VERIFIED | 11KB, BatchUpsert + UpdateOverduePayments |
| `internal/socialinsurance/handler.go` | API 路由 | VERIFIED | 22KB, 7+新路由 (dashboard/enroll/single/stop/single/payment-callback/monthly-records/confirm-payment/records/export) |
| `internal/socialinsurance/scheduler.go` | asynq 定时任务 | VERIFIED | 6KB, TypeGenerateMonthlyPayments + TypeCheckPaymentStatus |
| `internal/socialinsurance/excel.go` | Excel 导出 | VERIFIED | 10KB, ExportSIRecordsWithDetails 22列 |
| `internal/socialinsurance/dto.go` | DTO 定义 | VERIFIED | 8KB, SIDashboardResponse/SIStatItem/OverdueItem |
| `internal/common/model/org.go` | Organization.SIPaymentChannel | VERIFIED | SIPaymentChannel 字段存在 |
| `frontend/src/views/socialinsurance/SIDashboard.vue` | 4卡片数据看板 | VERIFIED | 6KB, 4卡片 + trend + axios GET |
| `frontend/src/components/socialinsurance/EnrollDialog.vue` | 增员弹窗 | VERIFIED | 6KB, remote-method + 全字段 + POST |
| `frontend/src/views/tool/SITool.vue` | 参保操作 Tab + 红色横幅 | VERIFIED | 15KB, overdue-banner + EnrollDialog 集成 |
| `frontend/src/components/socialinsurance/StopDialog.vue` | 减员弹窗 | VERIFIED | 7KB, 终止月份 + 原因 + 生效规则 + confirm |
| `frontend/src/views/socialinsurance/SIRecordsTable.vue` | 参保记录列表 | VERIFIED | 8KB, 5色标签 + 渠道列 + 导出弹窗 |
| `frontend/src/components/socialinsurance/SIDetailDialog.vue` | 五险分项弹窗 | PARTIAL | 6KB, 组件完整但 API 路径不匹配 |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| SIDashboard.vue | /api/v1/social-insurance/dashboard | axios GET | WIRED | 路径匹配 |
| EnrollDialog.vue | /api/v1/social-insurance/enroll/single | axios POST | WIRED | 路径匹配 |
| StopDialog.vue | /api/v1/social-insurance/stop/single | axios POST | WIRED | 路径匹配 |
| SITool.vue | SIDashboard.vue | 组件引入 | WIRED | import EnrollDialog |
| SITool.vue | /api/v1/social-insurance/dashboard | axios GET (overdueItems) | WIRED | 路径匹配 |
| SIRecordsTable.vue | /api/v1/social-insurance/monthly-records | axios GET | WIRED | 路径匹配 |
| SIRecordsTable.vue | /api/v1/social-insurance/records/export | axios GET blob | WIRED | 路径匹配 |
| SIRecordsTable.vue | StopDialog.vue | 行内减员按钮 | WIRED | openStopDialog |
| SIRecordsTable.vue | SIDetailDialog.vue | 行内详情按钮 | WIRED | openDetail |
| SIDetailDialog.vue | /api/v1/social-insurance/monthly-records/:id | axios GET | NOT_WIRED | 前端路径 /api/v1/socialinsurance/records/${id}/detail 与后端不匹配 |
| handler.go | dashboard_service.go | SIDashboardService.GetDashboard | WIRED | DI 注入 |
| scheduler.go | repository.go | asynq Worker | WIRED | HandleGenerate/HandleCheck 调用 repo |
| handler.go | excel.go | ExportSIRecordsWithDetails | WIRED | 直接调用 |
| excel.go | model.go | SIMonthlyPayment 字段映射 | WIRED | 引用 SocialInsuranceRecord 字段 |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|--------------|--------|--------------------|--------|
| SIDashboard.vue | statCards (computed) | axios GET /dashboard | 后端 errgroup 并发查询 SIMonthlyPayment | FLOWING |
| SIRecordsTable.vue | records (ref) | axios GET /monthly-records | 后端 repo.ListRecords 查询 | FLOWING |
| SIDetailDialog.vue | detail (ref) | axios GET (路径不匹配) | 后端 GetMonthlyRecordDetail | DISCONNECTED |
| SIRecordsTable.vue export | blob | axios GET /records/export | 后端 excelize 生成 | FLOWING |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| SI-01 | 01,02 | 应缴总额环比 | SATISFIED | SIDashboard.vue 应缴总额卡片 + dashboard_service.go |
| SI-02 | 01,02 | 单位部分环比 | SATISFIED | SIDashboard.vue 单位部分卡片 |
| SI-03 | 01,02 | 个人部分环比 | SATISFIED | SIDashboard.vue 个人部分卡片 |
| SI-04 | 01,02 | 欠缴金额 | SATISFIED | SIDashboard.vue 欠缴卡片(#FF5630) + scheduler >=26日 overdue |
| SI-05 | 02 | 增员姓名检索 | SATISFIED | EnrollDialog.vue remote-method |
| SI-06 | 02 | 起始月份 | SATISFIED | EnrollDialog.vue startYearMonth 默认当月 |
| SI-07 | 02 | 缴费城市和基数 | SATISFIED | EnrollDialog.vue cityCode + siBase |
| SI-08 | 02 | 公积金比例和基数 | SATISFIED | EnrollDialog.vue hfRatio + hfBase |
| SI-09 | 03 | 减员姓名检索 | SATISFIED | StopDialog.vue remote-method |
| SI-10 | 03 | 终止月份不早于当月 | SATISFIED | StopDialog.vue disableStopDate |
| SI-11 | 03 | 减员原因三选一 | SATISFIED | StopDialog.vue reason (job_change/retirement/other) |
| SI-12 | 03 | 转出日期和封存日期 | SATISFIED | StopDialog.vue transferDate + hfFreezeDate |
| SI-13 | 03 | 转出生效规则提示 | SATISFIED | StopDialog.vue 三档 tooltip |
| SI-14 | 01 | 缴费渠道选择 | SATISFIED | org.go SIPaymentChannel + handler PaymentChannel |
| SI-15 | 01 | 自主缴费确认 | SATISFIED | handler.go /confirm-payment |
| SI-16 | 01 | 代理缴费通知 | SATISFIED | handler.go /payment-callback |
| SI-17 | 01,03 | 缴费状态自动流转 | SATISFIED | scheduler.go HandleCheckPaymentStatus + 5种 PaymentStatus |
| SI-18 | 03 | 5种状态标签 | SATISFIED | SIRecordsTable.vue statusTagType + statusLabelMap |
| SI-19 | 02,03 | 红字提醒+政策通知 | SATISFIED | SITool.vue overdue-banner 红色横幅 |
| SI-20 | 03 | 五险分项明细 | PARTIAL | SIDetailDialog.vue 组件完整但 API 路径不匹配 |
| SI-21 | 04 | Excel 导出 | SATISFIED | excel.go 22列 + handler /records/export + SIRecordsTable.vue blob |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| handler.go | 529 | TODO: HMAC 签名验证(T-08-04) | Warning | webhook 安全性依赖简单验证，生产环境需补充 HMAC |

### Human Verification Required

### 1. SIDashboard 四卡片视觉效果

**Test:** 打开社保数据看板页面，检查4张卡片布局
**Expected:** 4卡片并排(grid 4列)，应缴/单位/个人蓝色(#4F6EF7)，欠缴红色(#FF5630)，环比箭头方向正确
**Why human:** CSS 布局和颜色渲染需要视觉确认

### 2. 红色欠缴横幅交互

**Test:** 在参保操作 Tab 中，当有欠缴记录时检查横幅显示
**Expected:** 红色背景横幅，显示最大欠缴信息，超过5条显示"还有N项"折叠，可关闭
**Why human:** 条件渲染和交互行为需要视觉确认

### 3. 增减员弹窗表单交互

**Test:** 分别打开增员和减员弹窗，测试姓名搜索、表单填写和提交流程
**Expected:** 姓名搜索 debounce 300ms 响应，身份证号自动填充，表单校验提示正确
**Why human:** 表单交互和验证提示需要端到端体验确认

### Gaps Summary

Phase 08 整体完成度很高，13 个文件全部存在且内容实质性完整。唯一问题是 SIDetailDialog.vue 的 API 路径与后端不匹配：

- 前端调用: `/api/v1/socialinsurance/records/${recordId}/detail`
- 后端路由: `/api/v1/social-insurance/monthly-records/:id`

差异有两处：(1) `socialinsurance` 缺少连字符应为 `social-insurance`；(2) `records/:id/detail` 应为 `monthly-records/:id`。这会导致五险分项详情弹窗无法加载后端数据。

修复方案：将 SIDetailDialog.vue 第180行改为 `await axios.get(\`/api/v1/social-insurance/monthly-records/${recordId}\`)` 即可。

另外 handler.go 有一处 TODO（HMAC 签名验证），属于安全性增强项，不影响核心功能但建议后续补充。

---

_Verified: 2026-04-19_
_Verifier: Claude (gsd-verifier)_
