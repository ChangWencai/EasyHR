---
status: issues
files_reviewed: 14
phase: 12
phase_name: 考勤合规报表
reviewer_note: "WR-01 经人工复核为误报——实际代码使用 `if weekday == 0 || weekday == 6`（整数比较），非 switch/case 字符串比较，逻辑正确。"

## Findings

### WR-01: ~~ClassifyOvertimeCategory 字符串与整数比较永远不匹配~~ [已核实：无此问题]
**File:** `internal/attendance/rule_engine.go:89-93`
**Severity:** ~~WARNING~~ — **FALSE POSITIVE（已人工复核）**
**Issue:** ~~`weekday` 是 `int` 类型，但 switch case 比较字符串...~~ 实际代码使用 `if weekday == 0 || weekday == 6 { return "weekend" }`（整数比较），**完全正确**。reviewer 误将 plan 中的注释当作实际代码，plan 写的是 `if weekday == 0 || ...` 但 reviewer 在分析时描述为 switch/case 字符串比较。结论：**无需修复**，WR-01 作废。

---

### WR-02: year_month 缺少格式校验，可接受无效月份值
**File:** `internal/attendance/dto.go:182-185`
**Severity:** WARNING
**Issue:** `ComplianceReportRequest` 的 `year_month` 字段只有 `binding:"required"`，没有格式校验正则表达式。`handler.go` 的各合规接口也仅检查非空。无效值如 `2024-13`、`2024-00`、`abc` 均可通过校验，最终传入数据库查询或 `time.Parse("2006-01", ...)` 时才失败，返回 500 错误而非友好的参数错误提示。

```go
// 当前代码
type ComplianceReportRequest struct {
    YearMonth string `form:"year_month" binding:"required"` // 无格式校验
    DeptIDs   string `form:"dept_ids"`
}
```

**Fix:** 添加正则校验（使用 go-playground/validator 的 `regex` 标签）：
```go
type ComplianceReportRequest struct {
    YearMonth string `form:"year_month" binding:"required,regexp=^\\d{4}-(0[1-9]|1[0-2])$"`
    DeptIDs   string `form:"dept_ids"`
}
```

同时在 handler 层捕获校验错误，返回 400 而非 500：
```go
if err := c.ShouldBindQuery(&req); err != nil {
    response.BadRequest(c, "year_month 格式错误，请使用 YYYY-MM（如 2024-03）")
    return
}
```

同类问题存在于 `GetDailyRecords`（dto.go 无校验，handler.go 同样只检查 required）和 `GetMonthlyReport`（同样无格式校验）。

---

### WR-03: roundHalf 函数使用 int 截断，负数值计算错误
**File:** `internal/attendance/service.go:430-432`
**Severity:** WARNING
**Issue:** `roundHalf` 使用 `int(val/0.5+0.5)` 实现取整，但 `int()` 向零截断（而非向下取整）。对于负数：

- `val = -0.5`: `int(-0.5/0.5 + 0.5) = int(-1.0 + 0.5) = int(-0.5) = 0` → 返回 `0.0`（应为 `-0.5`）
- `val = -1.0`: `int(-1.0/0.5 + 0.5) = int(-2.0 + 0.5) = int(-1.5) = -1` → 返回 `-0.5`（应为 `-1.0`）

虽然实际业务中年假天数和加班时长通常不为负，但如果数据异常或手动修正输入负值，Excel 导出的合计行会出现数据错误。

**Fix:**
```go
func roundHalf(val float64) float64 {
    return math.Floor(val/0.5+0.5) * 0.5
}
```
需要 import "math"。或者使用 `math.Round`：
```go
func roundHalf(val float64) float64 {
    return math.Round(val/0.5) * 0.5
}
```

---

### WR-04: GetComplianceMonthly 中新员工无月度记录时显示全零数据
**File:** `internal/attendance/service.go:988`
**Severity:** WARNING
**Issue:** `GetComplianceMonthly` 中，如果某员工当月没有 `AttendanceMonthly` 记录（map 中不存在该 key），则 `m.RequiredDays`、`m.ActualDays` 等全部为零值。业务上这是"无记录"而非"零出勤"的语义，但前端和 Excel 导出都显示为零，可能造成误解。

```go
m := monthlyMap[emp.EmployeeID]  // 新员工：zero-value AttendanceMonthly{0,0,0,...}
```

**Fix:** 在循环中添加判断，无记录时跳过或标注：
```go
m, ok := monthlyMap[emp.EmployeeID]
if !ok {
    // 新员工无当月考勤记录，可选择跳过或显示 -
    list = append(list, MonthlyComplianceItem{
        EmployeeID: emp.EmployeeID,
        EmployeeName: emp.EmployeeName,
        DepartmentName: emp.DepartmentName,
        IsAnomaly: false,
        // 其他字段保持默认值 0，前端可选择不显示或显示 "-"
    })
    continue
}
```

---

### WR-05: ListApprovalsByMonth 和 ListClockRecordsByMonth 未应用 orgScope 安全模式
**File:** `internal/attendance/repository.go:348-363, 368-382`
**Severity:** WARNING
**Issue:** 这两个函数绕过了 `orgScope()` 辅助函数，直接在 WHERE 子句中手动拼接 `org_id = ?`。虽然功能正确（等效于 orgScope），但与代码库其他函数使用 `r.db.Scopes(r.orgScope(orgID))` 的模式不一致，增加了维护风险和遗漏 org_id 过滤的可能性。

**Fix:** 统一使用 `Scopes`：
```go
func (r *AttendanceRepository) ListApprovalsByMonth(orgID int64, employeeIDs []int64, yearMonth string) ([]Approval, error) {
    if len(employeeIDs) == 0 {
        return nil, nil
    }
    parsed, err := time.Parse("2006-01", yearMonth)
    if err != nil {
        return nil, err
    }
    startTime := time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.UTC)
    endTime := startTime.AddDate(0, 1, 0)
    var approvals []Approval
    err = r.db.Scopes(r.orgScope(orgID)).
        Where("employee_id IN ? AND status = ? AND start_time >= ? AND start_time < ?",
            employeeIDs, ApprovalStatusApproved, startTime, endTime).
        Find(&approvals).Error
    return approvals, err
}
```

同类问题存在于 `ListClockRecordsByMonth`（同样绕过了 orgScope）。

---

### WR-06: ComplianceTable 组件未被任何页面使用
**File:** `frontend/src/components/compliance/ComplianceTable.vue`
**Severity:** INFO
**Issue:** `ComplianceTable.vue` 组件定义了通用的表格封装（带分页），但四个合规报表页面（ComplianceOvertime、ComplianceLeave、ComplianceAnomaly、ComplianceMonthly）全部手写表格代码，未使用此组件。组件存在但未被消费，属于死代码。

**Fix:** 评估是否保留：
- 如果后续重构需统一表格样式和分页行为 → 保留并使用它
- 如果确认不需要 → 删除该文件

---

### WR-07: GetClockLive 的 departmentID 参数从未使用
**File:** `internal/attendance/service.go:199` 和 `internal/attendance/handler.go:202-207`
**Severity:** INFO
**Issue:** handler 解析了 `department_id` 查询参数并传给 service，但 service 的 `GetClockLive` 方法完全忽略该参数，始终查询所有在职员工。功能上"按部门筛选打卡实况"未实现。

```go
// handler.go: 解析了 departmentID
var departmentID *int64
if deptIDStr := c.Query("department_id"); deptIDStr != "" {
    if id, err := strconv.ParseInt(deptIDStr, 10, 64); err == nil {
        departmentID = &id
    }
}
result, err := h.svc.GetClockLive(..., departmentID, ...)

// service.go: departmentID 参数被忽略
func (s *AttendanceService) GetClockLive(ctx context.Context, orgID int64, date string, departmentID *int64, page, pageSize int) (*ClockLiveResponse, error) {
    allEmps, err := s.repo.ListAllActiveEmployees(orgID)  // 无 dept 过滤
```

**Fix:** 要么移除参数（如果功能不需要），要么在 repository 中添加按部门过滤：
```go
func (r *AttendanceRepository) ListAllActiveEmployees(orgID int64, departmentID *int64) ([]EmployeeBrief, error) {
    // ...
    query := r.db.Table("employees").Where("employees.org_id = ? AND employees.status IN ? AND employees.deleted_at IS NULL", orgID, []string{"active", "probation"})
    if departmentID != nil {
        query = query.Where("employees.department_id = ?", *departmentID)
    }
    // ...
}
```

---

### WR-08: ExportComplianceMonthly 导出 Blob 类型断言不安全
**File:** `frontend/src/views/compliance/ComplianceMonthly.vue:241-244`
**Severity:** INFO
**Issue:** 导出函数直接断言响应为 `Blob`，未验证响应内容类型或状态码。如果后端返回非 2xx 状态码（如 400 参数错误），axios 仍会抛出异常，但响应体不是 Blob 时会导致运行时错误。

```go
const blob = await attendanceApi.exportComplianceMonthly({...}) as unknown as Blob
```

**Fix:** 添加响应状态检查和类型验证：
```typescript
async function handleExport() {
  try {
    const deptIds = selectedDepts.value.length ? selectedDepts.value.join(',') : undefined
    const response = await attendanceApi.exportComplianceMonthly({
      year_month: selectedMonth.value,
      dept_ids: deptIds,
    })
    const blob = response as unknown as Blob
    if (!blob || blob.type !== 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet') {
      ElMessage.error('导出失败，请重试')
      return
    }
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `月度考勤汇总_${selectedMonth.value}.xlsx`
    a.click()
    URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch {
    ElMessage.error('导出失败')
  }
}
```

---

### WR-09: ListEmployeesByOrgWithDept 只查询 status='active'，排除试用期员工
**File:** `internal/attendance/repository.go:331`
**Severity:** INFO
**Issue:** 四个合规报表共用 `ListEmployeesByOrgWithDept`，但该函数只查询 `status = 'active'` 的员工。相比之下，`ListAllActiveEmployees`（打卡实况用）查询 `status IN ('active', 'probation')`。试用期员工被排除在合规报表之外，可能导致合规数据不完整（试用期员工也有加班、请假等数据）。

**Fix:** 根据业务需求决定是否包含试用期员工。建议与 `ListAllActiveEmployees` 保持一致：
```go
Where("employees.org_id = ? AND employees.status IN ? AND employees.deleted_at IS NULL", orgID, []string{"active", "probation"})
```

---

## Security Summary

**org_id 隔离:** 正确。所有合规报表 API 均通过 `getOrgID(c)` 从 JWT token 提取 org_id，传递给 service/repository 层。repository 中使用 `orgScope(orgID)` 或手动 `WHERE org_id = ?` 确保数据隔离。✓

**输入校验:** 有缺陷（见 WR-02）。`year_month` 格式未校验，无效值可到达数据库层。

**硬编码密码/密钥:** 未发现。✓

**SQL 注入:** 未发现。所有查询均使用 GORM 参数化查询（`?` 占位符）。✓

**Excel 导出路径遍历:** 未发现。文件名由 `fmt.Sprintf` 构造，不含用户输入。✓

---

## Positive Notes

- 四个合规报表 API 的 org_id 隔离实现一致且正确
- 所有分页逻辑正确处理了边界情况（start > total 时 clamp 到末尾）
- 空员工列表时返回空数组而非 nil，前端处理健壮
- Excel 导出异常处理完善（try-catch，错误提示）
- 前端 dept_ids 参数为空时正确传递 `undefined`（不包含该参数）
- `ComplianceStatCard` 组件设计合理，支持 hover 动效和动态 icon

---

_Reviewed: 2026-04-20T00:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
