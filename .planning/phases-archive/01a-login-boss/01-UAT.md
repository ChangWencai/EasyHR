---
status: testing
phase: 01-login-boss
source: 01-01-PLAN.md, 01-02-PLAN.md
started: 2026-04-11T12:00:00Z
updated: 2026-04-11T12:00:00Z
---

## Current Test
<!-- OVERWRITE each test - shows where we are -->

number: 3
name: 手机号+密码登录
expected: |
  在"密码登录" tab 输入手机号 + 密码，点击登录。
  成功时跳转到 /home（已有企业）或 /onboarding/org-setup（首次登录），页面无报错。
awaiting: user response

## Tests

### 1. 登录页 UI 与3种登录方式
expected: 访问 /login 显示品牌蓝背景 + 居中白色卡片，"易人事" 标题可见。el-tabs 显示3个 tab：手机验证码 / 密码登录 / 微信登录。Tab 1：输入手机号 + 验证码，点击登录按钮。Tab 2：输入手机号 + 密码，点击登录按钮。Tab 3：点击微信登录按钮，显示"微信登录功能开发中"提示。
result: pass

### 2. 验证码发送与60秒倒计时
expected: 在"手机验证码" tab 输入手机号，点击"获取验证码"。按钮立即变为"已发送(60s)"并每秒递减，到0时恢复"获取验证码"。60秒内按钮禁用，无法重复发送。
result: skipped
reason: 阿里云短信凭证未配置（MissingAccessKeyId），API 无法实际发送验证码

### 3. 手机号+密码登录
expected: 在"密码登录" tab 输入手机号 + 密码，点击登录。成功时跳转到 /home（已有企业）或 /onboarding/org-setup（首次登录），页面无报错。
result: [pending]

### 4. Auth Guard 未登录重定向
expected: 未登录状态下直接访问 http://localhost:5173/home，被自动重定向到 /login。访问 /login 不重定向，可正常显示登录页。
result: [pending]

### 5. MEMBER 账号 403 拒绝
expected: 用一个 role=member 的账号登录（任意方式），页面显示提示："您的账号为员工账号，请使用员工端微信小程序登录"，不跳转。
result: [pending]

### 6. 登录后按 onboarding 状态分流
expected: 有企业的 OWNER/ADMIN 登录成功后跳转到 /home。无企业的账号登录成功后跳转到 /onboarding/org-setup。登录后 URL 与页面内容匹配。
result: [pending]

### 7. 后端密码登录 API
expected: POST /api/v1/auth/login/password {"phone":"...", "password":"..."} 返回 {"access_token": "...", "refresh_token": "...", "onboarding_required": bool}，HTTP 200。
result: [pending]

### 8. 后端 /auth/me 接口
expected: GET /api/v1/auth/me（带 Bearer token）返回 {id, name, phone, role, org{id,name,credit_code,city}, onboarding_required}，HTTP 200。
result: [pending]

## Summary

total: 8
passed: 1
issues: 0
pending: 6
skipped: 1
blocked: 0

## Gaps

[none yet]
