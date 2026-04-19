---
status: fixed
trigger: "注册接口返回消息被拼接：expected '该手机号已注册，请直接登录', actual '登录已过期，请重新登录，该手机号已注册，请直接登录'"
created: 2026-04-11T14:00:00+08:00
updated: 2026-04-11T14:15:00+08:00
---

## Current Focus
status: fixed

## Symptoms
expected: "注册接口返回 {\"code\": 10014, \"message\": \"该手机号已注册，请直接登录\"}，前端显示该消息"
actual: "前端显示 \"登录已过期，请重新登录，该手机号已注册，请直接登录\" — 两个错误消息被拼接"
errors: ["后端返回 code=10014, message=\"该手机号已注册，请直接登录\"", "前端显示了额外的 \"登录已过期，请重新登录\""]
reproduction: "在注册 tab 输入已注册的手机号和验证码，点击注册"
started: "注册功能修复后出现（2026-04-11）"

## Eliminated
- 前端 handleRegister 有多个 catch 块: 否 - 只有一个 catch 块
- 注册接口返回了两次响应: 否 - 后端只返回一次

## Evidence
- timestamp: 2026-04-11T14:05:00+08:00
  checked: "handler.go:68"
  found: "Register handler 在手机号已注册时返回 http.StatusUnauthorized(401)，code=10014"
  implication: "HTTP 401 会触发前端 request.ts 拦截器显示 '登录已过期'"
- timestamp: 2026-04-11T14:05:00+08:00
  checked: "request.ts:37-42"
  found: "响应拦截器对所有 401 状态码显示 '登录已过期，请重新登录'，然后 return Promise.reject(error)"
  implication: "错误继续向上传播，handleRegister 的 catch 也显示业务错误 → 两条消息"
- timestamp: 2026-04-11T14:08:00+08:00
  checked: "request.ts:10 PUBLIC_AUTH_PATHS"
  found: "'/auth/register' 在 PUBLIC_AUTH_PATHS 中 → 请求不带 token → 不存在 'token 过期' 的前置条件"
  implication: "返回 401 是语义错误，手机号已注册是业务错误，应返回 409 Conflict"

## Resolution
root_cause: "Register handler 返回 HTTP 401，而 401 是认证失败的语义。前端 request.ts 拦截器将所有 401 解释为 'token 过期' 并显示 '登录已过期'，同时业务 handler 的 catch 也显示原始错误，导致两条消息拼接"
fix: "将 handler.go:68 Register 的 http.StatusUnauthorized 改为 http.StatusConflict(409)"
verification: "已编译通过，请重启后端服务并测试"
files_changed:
  - "internal/user/handler.go"
