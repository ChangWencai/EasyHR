# Phase 03a: web注册界面 - Context

**Gathered:** 2026-04-11
**Status:** Ready for planning

<domain>
## Phase Boundary

在 H5 登录页（`LoginView.vue`）将 Tab 3 从占位"微信登录"改为"注册"Tab，点击注册后通过手机号+验证码完成账号创建，并跳转到 `/onboarding/org-setup` 录入企业信息。

</domain>

<decisions>
## Implementation Decisions

### 注册入口 UI
- **D-01:** 替换 LoginView.vue Tab 3「微信登录」为「注册」Tab
- **D-02:** 注册 Tab 表单：手机号输入框 + 验证码输入框（与 Tab 1「手机验证码」完全相同）
- **D-03:** 注册 Tab 提交按钮文案：「注册」（区别于 Tab 1 的「登录」）
- **D-04:** 注册成功后：POST /auth/login 返回 `onboarding_required=true` → 前端跳转 `/onboarding/org-setup`

### 注册流程（登录即注册）
- **D-05:** 点击「注册」→ 发送验证码 → 用户输入验证码 → POST /auth/login → 后端自动创建用户（`onboarding_required=true`）→ 跳 `/onboarding/org-setup`
- **D-06:** 无需新增后端接口，复用已有 `POST /auth/send-code` 和 `POST /auth/login` 接口
- **D-07:** 注册 Tab 的 handleSendCode 和 handleSubmit 逻辑与 Tab 1「手机验证码」相同

### Onboarding 页面
- **D-08:** `/onboarding/org-setup` 无需修改（已在 Phase 1 完成，Auth Guard 已豁免该路径）
- **D-09:** Onboarding 提交 `POST /orgs` → 创建企业 → 跳转 `/home`（现有逻辑不变）

### 已注册手机号处理（已知局限）
- **D-10:** 若用户输入已存在的手机号发送验证码，后端将其视为登录而非注册，登录成功后 `onboarding_required=false`，直接跳转 `/home`（而非注册流程）
- **D-11:** Phase 03a 不处理"手机号已被注册"的明确提示（需要新增 `/auth/check-phone` 接口，延后处理）

### Claude's Discretion
- 注册 Tab 的具体文案（确认按钮、提示文案）
- 注册 Tab 的视觉样式（与 Tab 1 一致即可）
- 是否需要在注册成功后清除 URL 中的 code 参数

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### 登录页代码
- `frontend/src/views/layout/LoginView.vue` — 现有登录页实现，需将 Tab 3 改为注册
- `frontend/src/stores/auth.ts` — authStore：setToken、logout
- `frontend/src/api/request.ts` — axios 封装：POST /auth/login、POST /auth/send-code

### Onboarding 代码
- `frontend/src/views/onboarding/OrgSetup.vue` — 企业信息录入页（注册流程终点）
- `frontend/src/router/index.ts` — `/onboarding/org-setup` 已在 Auth Guard 豁免列表（line 134）

### 登录 Phase Context
- `.planning/phases/01a-login-boss/01-CONTEXT.md` — 登录页 D-01~D-20 决策（D-07 首次登录跳转 /onboarding/org-setup）

</canonical_refs>

<codebase_context>
## Existing Code Insights

### Reusable Assets
- LoginView.vue Tab 1（手机验证码）：表单结构、发送验证码逻辑、60秒倒计时 → 直接复用
- handleLoginSuccess()：已处理 `onboarding_required` 分流逻辑 → 注册流程复用
- Auth Guard `/onboarding/org-setup` 豁免：无需修改路由守卫

### Established Patterns
- 品牌色：#1677ff，LoginView 样式保持一致
- Element Plus 组件：el-tabs、el-input、el-button（与 LoginView 一致）
- 移动端适配：已有 `@media (max-width: 480px)` 样式，注册 Tab 自然继承

### Integration Points
- POST /auth/send-code（手机号发送验证码）
- POST /auth/login（手机号+验证码登录/自动注册）
- POST /orgs（Onboarding 提交企业信息）

</codebase_context>

<specifics>
## Specific Ideas

- 注册 Tab 按钮文案：与 Tab 1 保持一致结构，按钮文字从"登录"改为"注册"
- 无需额外 UX 设计，替换现有占位 Tab 即可

</specifics>

<deferred>
## Deferred Ideas

### 手机号已被注册的明确提示
- 目前点击注册时，若手机号已存在，后端将其视为登录。用户体验是直接跳转到 /home，而非看到"该手机号已注册"提示。
- 解决方案：后端新增 `/auth/check-phone` 接口，注册前先检查手机号是否已注册，返回相应提示。
- 归属：建议作为 Phase 02a 或后续迭代的一部分

</deferred>

---

*Phase: 03a-web*
*Context gathered: 2026-04-11*
