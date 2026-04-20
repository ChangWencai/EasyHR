# Phase 11 代码审查报告（合同合规）

**审查文件：**
- Backend: `internal/employee/pdf.go`, `contract_model.go`, `contract_service.go`, `contract_handler.go`, `contract_dto.go`, `internal/todo/scheduler.go`
- Frontend: `src/api/contract.ts`, `src/views/employee/EmployeeDrawer.vue`, `src/views/employee/components/ContractList.vue`, `src/components/contract/*.vue`, `src/views/sign/SignPage.vue`

---

## CRITICAL

无。

---

## HIGH

### [HIGH-1] `FindBySignToken` 缺少租户隔离 — 跨租户信息泄露

**文件：** `internal/employee/contract_repository.go:122-130`

```go
// FindBySignToken 根据 SignToken 查找验证码记录
func (r *ContractRepository) FindBySignToken(signToken string) (*ContractSignCode, error) {
    var signCode ContractSignCode
    err := r.db.Where("sign_token = ?", signToken).First(&signCode).Error  // 无 org_id 过滤
    ...
}
```

**问题：** `ConfirmSign` 通过 `SignToken` 验证签署，但 `FindBySignToken` 在全表中查询，没有任何租户隔离。攻击者持有任意租户的 `SignToken`（如通过手机号注册到其他企业）即可确认签署，修改其他企业的合同状态。

**背景：** 其他公开端点（`SendSignCode`、`VerifySignCode`）通过手机号哈希全局查找员工（`FindByPhoneHashGlobal`）并验证 `contract.EmployeeID` 关联关系实现了跨租户身份验证。`ConfirmSign` 应该遵循同样的模式 — 通过 `signCode.Phone` 反查员工 orgID，再验证合同属于该 orgID。

**建议修复：** 在 `ConfirmSign` 中，查询到 `signCode` 后，通过 `signCode.Phone` 哈希反查员工 orgID，然后用 orgID 调用带租户隔离的 `FindByID`：

```go
// ConfirmSign 中替换：
signCode, err := s.contractRepo.FindBySignToken(signToken)
// ...
emp, _ := s.empRepo.FindByPhoneHashGlobal(crypto.HashSHA256(signCode.Phone))
if emp == nil {
    return nil, fmt.Errorf("签署验证失败")
}
contract, err := s.contractRepo.FindByID(emp.OrgID, contractID) // 租户隔离
if err != nil {
    return nil, fmt.Errorf("合同不存在")
}
```

---

### [HIGH-2] `uploadPdfToOss` 忽略 OSS 上传错误 — PDF 丢失静默

**文件：** `internal/employee/contract_service.go:706-720`

```go
func (s *ContractService) uploadPdfToOss(...) (string, error) {
    if s.ossClient == nil {
        objectKey := fmt.Sprintf("contracts/org_%d/contract_%d_%d.pdf", ...)
        return objectKey, nil   // ossClient 为 nil 时返回 key，但合同状态已是 pending_sign
    }
    putURL, err := s.ossClient.GeneratePutURL(...)
    if err != nil {
        return objectKey, nil   // 上传失败也返回 nil 错误，SendSignLink 继续执行
    }
    _ = putURL // 前端直传，或后端上传
    return objectKey, nil
}
```

**问题：** 当 `ossClient.GeneratePutURL` 失败时，函数返回 `nil` error，调用方 `SendSignLink` 继续执行，发送短信给员工，告知"签署链接已发送"，但 PDF 实际没有上传到 OSS。员工点击链接打开的是空 PDF。

**建议修复：** 上传失败时返回 error：

```go
putURL, err := s.ossClient.GeneratePutURL(...)
if err != nil {
    return "", fmt.Errorf("生成上传URL失败: %w", err)
}
// 如果前端直传模式：返回 putURL，让前端上传
// 如果后端上传模式：这里上传后再返回
return objectKey, nil
```

---

## MEDIUM

### [MEDIUM-1] `ConfirmSign` 中 `contractRepo.Update` 忽略返回值和错误

**文件：** `internal/employee/contract_service.go:572-576`

```go
s.contractRepo.Update(orgID, contractID, map[string]interface{}{
    "status":         status,
    "sign_date":      now,
    "signed_pdf_url": signedPdfUrl,
})
```

**问题：** `Update` 方法有 error 返回值但被完全忽略。如果更新失败，函数仍返回"签署成功"，但数据库状态未更新。

**建议修复：** 检查并返回错误。

---

### [MEDIUM-2] `SignPage.vue` 中 `window.open` 未作为 `void` 调用

**文件：** `frontend/src/views/sign/SignPage.vue:225`

```vue
@click="window.open(signedPdfUrl, '_blank')"
```

**问题：** 在模板表达式中直接调用 `window.open` 是非标准用法，可能在某些构建工具下报错。应通过 methods 中的处理函数调用。

**建议修复：** 在 `<script setup>` 中定义处理函数：

```ts
function openPdf() {
  window.open(signedPdfUrl.value, '_blank')
}
```

```vue
@click="openPdf"
```

---

### [MEDIUM-3] `ContractList.vue` 终止合同理由为硬编码字符串

**文件：** `frontend/src/views/employee/components/ContractList.vue:63`

```ts
await contractApi.terminate(contract.id, '老板主动终止', new Date().toISOString().split('T')[0])
```

**问题：** 终止理由"老板主动终止"是硬编码的，用户无法自定义。如合同列表页增加终止原因输入框体验会更好。

**建议修复：** 调用 `ElMessageBox.prompt` 让用户输入终止原因。

---

### [MEDIUM-4] `contract_dto.go` 字段命名风格不一致

**文件：** `internal/employee/contract_dto.go:79-86`

```go
type VerifySignCodeResponse struct {
    SignToken    string `json:"sign_token"`   // 下划线
    ExpiresIn   int    `json:"expires_in"`
    EmployeeName string `json:"employee_name"`
    ContractType string `json:"contract_type"`
    StartDate   string `json:"start_date"`
    EndDate     string `json:"end_date,omitempty"`
    OrgName     string `json:"org_name"`
}
```

**问题：** Go 字段名中 `SignToken` 和 `ExpiresIn` 是驼峰，其他是全小写下划线。JSON tag 也混用风格（`sign_token` vs `expires_in`）。

**建议：** 统一为全小写下划线（Go 风格）或驼峰（与项目其他 DTO 一致）。

---

### [MEDIUM-5] `SendSignCode` 中的短信内容注入潜在 XSS

**文件：** `internal/employee/contract_service.go:472`

```go
templateParam := fmt.Sprintf(`{"name":"%s","link":"%s","days":"7"}`, emp.Name, signLink)
```

**问题：** `emp.Name` 直接插入 JSON 字符串，如果名字中包含引号会导致 JSON 格式错误，可能被短信网关截断或产生意外行为。`signLink` 中的 URL 参数 token 如果包含特殊字符也可能有问题。

**建议：** 对 `emp.Name` 和 `signLink` 进行 JSON 转义：

```go
import "encoding/json"
param := map[string]string{"name": emp.Name, "link": signLink, "days": "7"}
paramBytes, _ := json.Marshal(param)
templateParam := string(paramBytes)
```

---

## LOW

### [LOW-1] `scheduler.go` 中月度社保待办循环内每次 `CreateTodo` 无事务

**文件：** `internal/todo/scheduler.go:116-126`

```go
for _, orgID := range orgIDs {
    todo := &TodoItem{...}
    _ = s.repo.CreateTodo(ctx, todo)  // 错误被静默忽略
}
```

**问题：** 如果 N 个企业中部分创建失败，错误被忽略且继续处理剩余企业。循环内无事务，如中途失败部分已创建的待办无法回滚。

**建议：** 记录失败的企业 ID，最后统一 log 告警。

---

### [LOW-2] `pdf.go` 字体文件大小未校验

**文件：** `internal/employee/pdf.go:40-50`

```go
regularFont, err := fontFiles.ReadFile("NotoSansSC-Regular.ttf")
pdf.AddUTF8FontFromBytes("NotoSansSC", "", regularFont)
```

**问题：** 嵌入字体文件大小未校验，恶意构建环境下如果字体文件被替换为超大文件可能导致内存问题。实际风险极低（embed.FS 是编译期嵌入）。

---

### [LOW-3] `contract_service.go` 多处 `_` 忽略错误

**文件：** `internal/employee/contract_service.go` 多处

```go
emp, _ := s.empRepo.FindByID(...)
s.db.Where("id = ?", contract.OrgID).First(&org)
foundEmp, _ := s.empRepo.FindByID(emp.OrgID, contract.EmployeeID)
```

**问题：** 多处 `emp`/`org` 查询失败时返回空对象，可能导致后续逻辑基于零值继续执行（显示空白信息）。应至少检查关键字段是否存在。

---

### [LOW-4] `scheduler.go` 定义了两次中国时区

**文件：** `internal/todo/scheduler.go:32,109`

```go
var cstZone = time.FixedZone("CST", 8*3600)
...
cst := time.FixedZone("CST", 8*3600)  // 重复定义
```

**建议：** 使用已定义的 `cstZone` 变量。

---

### [LOW-5] `ContractList.vue` 组件 emit 未使用

**文件：** `frontend/src/views/employee/components/ContractList.vue:18-20`

```ts
const emit = defineEmits<{
  'open-wizard': []
}>()
```

**问题：** `emit('open-wizard')` 从未被调用，`open-wizard` 事件在父组件 `EmployeeDrawer.vue` 中注册但无实际作用。

---

## Review Summary

| Severity | Count | Status |
|----------|-------|--------|
| CRITICAL | 0     | pass   |
| HIGH     | 2     | warn   |
| MEDIUM   | 5     | info   |
| LOW      | 5     | note   |

**Verdict: WARNING** — 2 个 HIGH 问题应在合并前修复：
1. `FindBySignToken` 缺少租户隔离（CRITICAL 级别跨租户风险）
2. `uploadPdfToOss` 静默忽略 OSS 上传失败，导致用户体验"签署链接已发"但 PDF 实际未上传

**重点关注：** `ConfirmSign` 的租户隔离问题（HIGH-1）是签署流程中最严重的安全缺陷——任何持有 `SignToken` 的用户可跨租户确认签署。
