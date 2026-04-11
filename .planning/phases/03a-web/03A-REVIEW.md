---
phase: 03a-web
reviewed: 2026-04-11T00:00:00Z
depth: standard
files_reviewed: 3
files_reviewed_list:
  - frontend/src/views/layout/LoginView.vue
  - frontend/src/api/request.ts
  - frontend/src/stores/auth.ts
findings:
  critical: 2
  warning: 4
  info: 1
  total: 7
status: issues_found
---

# Phase 03a-web: Code Review Report

**Reviewed:** 2026-04-11
**Depth:** standard
**Files Reviewed:** 3
**Status:** issues_found

## Summary

审查了 `LoginView.vue` 及其依赖的 `request.ts` 和 `auth.ts` 模块。发现注册流程存在严重逻辑错误（调用了登录接口而非注册接口），以及 `console.log` 调试代码遗留。另有定时器未清理、密码登录后未清空表单等潜在问题。

## Critical Issues

### CR-01: 注册 Tab 调用了登录接口

**File:** `frontend/src/views/layout/LoginView.vue:80, 109`
**Issue:** 注册 Tab 的表单提交绑定了 `handleSmsLogin`，该函数 POST 到 `/auth/login` 接口，而非注册接口。这意味着用户点击"注册"实际上是在用同一套验证码逻辑做登录，而非创建新账号。前端 UI 显示"注册"按钮但实际行为是登录，存在严重的 UX 欺骗性。
**Fix:**
```typescript
// 新增独立的注册处理函数
async function handleRegister() {
  if (!smsForm.value.phone || !smsForm.value.code) {
    ElMessage.error('请填写手机号和验证码')
    return
  }
  try {
    const resp = await request.post('/auth/register', {
      phone: smsForm.value.phone,
      code: smsForm.value.code,
    })
    handleLoginSuccess(resp.data)
  } catch (err: any) {
    ElMessage.error(err.response?.data?.message || '注册失败')
  }
}
```
然后将注册 Tab 的 `@submit.prevent` 和按钮 `@click` 从 `handleSmsLogin` 改为 `handleRegister`。

### CR-02: 调试代码 console.log 遗留

**File:** `frontend/src/views/layout/LoginView.vue:233`
**Issue:** `console.log('WeChat code received:', code)` 遗留在前端代码中，违反了项目规范（禁止 console.log）且会在浏览器控制台暴露微信授权码。
**Fix:**
```typescript
// 删除此行，微信回调逻辑如需保留可改为结构化日志
console.log('WeChat code received:', code)
```
改为使用项目日志库或直接删除（该功能为 Phase 1.5 占位）。

## Warnings

### WR-01: 倒计时定时器未在组件卸载时清理

**File:** `frontend/src/views/layout/LoginView.vue:159-168`
**Issue:** `countdownTimer` 是模块级变量，通过 `setInterval` 设置但从未调用 `clearInterval`。如果用户在倒计时期间路由跳转到其他页面，`setInterval` 将持续运行导致内存泄漏。
**Fix:**
```typescript
import { ref, onUnmounted } from 'vue'  // onUnmounted 已导入但未使用

// 在 startCountdown 后或组件级别添加清理
onUnmounted(() => {
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
})
```

### WR-02: 手机号格式校验不足

**File:** `frontend/src/views/layout/LoginView.vue:145`
**Issue:** 校验只检查了长度 `!== 11`，未验证内容是否为纯数字。配合 `type="number"` 的 `el-input`，可以输入 `1234567890a` 这类部分非数字的值通过前端校验。
**Fix:**
```typescript
if (!/^1\d{10}$/.test(smsForm.value.phone)) {
  ElMessage.error('请输入正确的手机号')
  return
}
```

### WR-03: 密码登录成功后未清空密码字段

**File:** `frontend/src/views/layout/LoginView.vue:192-210`
**Issue:** 登录成功后 `passwordForm.password` 仍保留在内存中。密码属于敏感数据，登录完成后应清空。
**Fix:**
```typescript
function handleLoginSuccess(resp: any) {
  authStore.setToken(resp.access_token)
  passwordForm.value.password = ''  // 登录成功后清空密码
  // ... 后续逻辑
}
```

### WR-04: 未使用的 dead code 函数

**File:** `frontend/src/views/layout/LoginView.vue:213-215`
**Issue:** `handleWechatLogin` 函数定义后从未被调用，属于死代码。
**Fix:** 如无近期使用计划，直接删除该函数。

## Info

### IN-01: request.ts 使用了未声明类型的 error 参数

**File:** `frontend/src/api/request.ts:20`
**Issue:** `(error: AxiosError)` 中 `AxiosError` 已导入但 `error` 参数本身可加类型注解以提高可读性。当前写法是合理的，但可考虑将 `catch` 块的 `err: any` 统一改为 `err: unknown` 并做 safe narrowing（见 `LoginView.vue` 中的两处 catch）。
**Fix:** 这是 Info 级别建议，保持当前写法可接受。

---

_Reviewed: 2026-04-11_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
