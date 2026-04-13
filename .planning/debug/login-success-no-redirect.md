---
status: verifying
trigger: "登录接口正确返回，但页面不跳转"
created: 2026-04-11T00:00:00+08:00
updated: 2026-04-11T00:00:00+08:00
---

## Current Focus

hypothesis: "前端 `handleSmsLogin` 中错误地将 `resp.data.data`（undefined）传给了 `handleLoginSuccess`，而不是 `resp.data`。"
test: "trace 整个 axios interceptor + response 处理链"
expecting: "`handleLoginSuccess` 收到的 `resp` 应为 `{ access_token, refresh_token, onboarding_required }`，但实际收到 `undefined`"
next_action: "已修复 3 处 `resp.data.data` -> `resp.data`，待人工验证"

## Symptoms

expected: 登录成功后根据 `onboarding_required` 跳转到 `/onboarding/org-setup` 或 `/home`
actual: 接口返回成功（code=0），页面不跳转
errors: ""
reproduction: 输入正确手机号+验证码，点击登录，接口返回成功
started: 2026-04-11

## Eliminated

## Evidence

- timestamp: 2026-04-11
  checked: "frontend/src/api/request.ts 的 axios response interceptor"
  found: "interceptor 返回 `response.data` — 即 `{ code: 0, message: 'success', data: { access_token, refresh_token, onboarding_required } }`"
  implication: "`handleSmsLogin` 中的 `resp` 已经是 API Body，不再是 axios response"

- timestamp: 2026-04-11
  checked: "internal/common/response/response.go 的 Success()"
  found: "`Success(c, data)` 渲染 `{ code: 0, message: 'success', data: data }`"
  implication: "HTTP Body 为三层嵌套：{ code, message, data: { access_token, refresh_token, onboarding_required } }"

- timestamp: 2026-04-11
  checked: "LoginView.vue 的 `handleSmsLogin`"
  found: "`handleLoginSuccess(resp.data.data)` — 访问了 `resp.data.data`"
  implication: "BUG！`resp` = { code, message, data: LoginResponse }，所以 `resp.data` = LoginResponse，`resp.data.data` = undefined。`setToken(undefined)` 将 undefined 存入 localStorage，但更重要的是 `resp.onboarding_required` 也变成了 undefined，导致 router.push 永远不执行"

- timestamp: 2026-04-11
  checked: "LoginView.vue 的 `handlePasswordLogin` 和 `handleRegister`"
  found: "同样的错误：`handleLoginSuccess(resp.data.data)`"
  implication: "三个登录入口全部受影响"

- timestamp: 2026-04-11
  checked: "修复后的 LoginView.vue"
  found: "3 处 `handleLoginSuccess(resp.data.data)` 已全部改为 `handleLoginSuccess(resp.data)`"
  implication: "修复已完成，等待人工验证"

## Resolution

root_cause: "axios response interceptor 将 `response.data`（即 API Body `{ code, message, data: LoginResponse }`）返回给调用方。`handleLoginSuccess(resp.data.data)` 错误地多访问了一层 `.data`，最终 `resp` 参数为 undefined，导致 `setToken(undefined)` 和 `onboarding_required` 检查失败，router.push 不执行。"
fix: "将 3 处 `handleLoginSuccess(resp.data.data)` 改为 `handleLoginSuccess(resp.data)`"
verification: ""
files_changed: ["frontend/src/views/layout/LoginView.vue"]
