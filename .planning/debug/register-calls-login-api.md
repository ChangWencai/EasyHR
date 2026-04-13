---
status: verifying
trigger: "点击注册按钮，调用了 POST /api/v1/auth/login (401 Unauthorized)，而不是注册接口"
created: 2026-04-11T00:00:00+08:00
updated: 2026-04-11T00:00:00+08:00
---

## Current Focus
hypothesis: "注册 tab 绑定到 handleSmsLogin，调用 /auth/login 而非注册接口；后端缺少 /auth/register 端点"
test: "1. 修改 LoginView.vue：注册 tab 使用独立 registerForm 和 handleRegister；2. 添加后端 /auth/register 端点"
expecting: "注册按钮调用 /auth/register，成功后跳转 onboarding 或 home"
next_action: "修改 LoginView.vue 前端注册表单和 handler，同时添加后端 /auth/register 端点"

## Symptoms
expected: 点击注册 tab 的"注册"按钮，应调用 /api/v1/auth/register（手机号+验证码注册新账号）
actual: 调用了 POST /api/v1/auth/login，返回 401 Unauthorized
errors: [堆栈: handleSmsLogin @ LoginView.vue:177]
reproduction: 切换到"注册" tab，输入手机号+验证码，点击"注册"按钮
started: 注册功能实现时引入

## Eliminated
- 后端路由缺失: 否 - 前端根本调用了错误的 handler
- send-code 接口问题: 否 - send-code 是通用的，发往同一手机号
- 验证码校验逻辑: 否 - handleSmsLogin 本身逻辑正确，问题是注册 tab 错误使用它

## Evidence
- timestamp: 2026-04-11
  checked: "LoginView.vue:80,109"
  found: "注册 tab form 的 @submit.prevent 和 button 的 @click 都绑定到 handleSmsLogin"
  implication: "点击注册按钮 → handleSmsLogin → POST /auth/login → 401（用户不存在时）"
- timestamp: 2026-04-11
  checked: "internal/user/handler.go:25-30"
  found: "没有 /auth/register 路由，仅有 /auth/send-code, /auth/login, /auth/login/password"
  implication: "即使前端正确调用，也需要先添加 /auth/register 端点"

## Resolution
root_cause: "注册 tab 的 form(@submit.prevent)和 button(@click)都绑定到 handleSmsLogin，handleSmsLogin 调用 POST /auth/login；后端也没有 /auth/register 端点"
fix: |
  1. LoginView.vue: 注册 tab 改为独立的 registerForm 和 handleRegister 函数，调用 POST /auth/register
  2. LoginView.vue: handleSendCode 根据 activeTab 判断从 registerForm 还是 smsForm 获取手机号
  3. internal/user/dto.go: 添加 RegisterRequest = LoginRequest 类型别名
  4. internal/user/handler.go: 添加 /auth/register 路由和 Register handler
  5. internal/user/service.go: 添加 Register service 方法，校验验证码并检查手机号是否已注册
verification: "切换到注册 tab，输入手机号+验证码，点击注册 → 调用 POST /auth/register → 成功创建企业+owner账号 → 跳转 onboarding"
files_changed:
  - "frontend/src/views/layout/LoginView.vue"
  - "internal/user/dto.go"
  - "internal/user/handler.go"
  - "internal/user/service.go"
