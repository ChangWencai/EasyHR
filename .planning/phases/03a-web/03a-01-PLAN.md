---
phase: "03a"
plan: "01"
type: execute
wave: 1
depends_on: []
files_modified:
  - frontend/src/views/layout/LoginView.vue
autonomous: true
requirements: []
must_haves:
  truths:
    - "用户可以看到「注册」Tab（Tab 3）"
    - "注册 Tab 表单包含手机号输入框和验证码输入框"
    - "点击「注册」按钮触发验证码发送流程"
    - "注册成功后跳转到 /onboarding/org-setup"
  artifacts:
    - path: frontend/src/views/layout/LoginView.vue
      provides: LoginView with register tab
      contains: "el-tab-pane.*label=\"注册\""
      min_lines: 60
    - path: frontend/src/views/layout/LoginView.vue
      provides: Register form structure (phone + code + button)
      contains: "el-input.*placeholder=\"请输入手机号\""
      contains2: "el-input.*placeholder=\"请输入验证码\""
      contains3: "注册"
  key_links:
    - from: frontend/src/views/layout/LoginView.vue
      to: POST /auth/send-code
      via: handleSendCode function
      pattern: "request.post.*send-code"
    - from: frontend/src/views/layout/LoginView.vue
      to: POST /auth/login
      via: handleSmsLogin function
      pattern: "request.post.*login"
    - from: handleLoginSuccess
      to: /onboarding/org-setup
      via: router.push when onboarding_required=true
      pattern: "router.push.*onboarding.*org-setup"
---

<objective>
将 LoginView.vue 的 Tab 3 从「微信登录」占位改为「注册」Tab，实现手机号+验证码注册流程，复用已有后端接口，注册成功后跳转到 /onboarding/org-setup。
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@frontend/src/views/layout/LoginView.vue
@frontend/src/stores/auth.ts
@frontend/src/api/request.ts
@.planning/phases/03a-web/03a-CONTEXT.md
@.planning/phases/03a-web/03A-UI-SPEC.md
</context>

<interfaces>
<!-- 复用接口均来自已有文件，无需新建 -->

<!-- request.post('/auth/send-code') 签名（来自 api/request.ts） -->
// POST /auth/send-code
// body: { phone: string }
// 返回: { message: string }

<!-- request.post('/auth/login') 签名（来自 api/request.ts） -->
// POST /auth/login
// body: { phone: string, code: string }
// 返回: { access_token: string, onboarding_required: boolean }

<!-- handleLoginSuccess 路由分流逻辑（来自 LoginView.vue line 190-197） -->
function handleLoginSuccess(resp: any) {
  authStore.setToken(resp.access_token)
  if (resp.onboarding_required === true) {
    router.push('/onboarding/org-setup')
  } else {
    router.push('/home')
  }
}
</interfaces>

<tasks>

<task type="auto">
  <name>Task 1: 将 Tab 3「微信登录」替换为「注册」Tab</name>
  <files>frontend/src/views/layout/LoginView.vue</files>
  <read_first>frontend/src/views/layout/LoginView.vue</read_first>
  <action>
将 LoginView.vue 的 Tab 3「微信登录」占位替换为「注册」Tab。注册表单结构与 Tab 1「手机验证码」完全相同（el-form > el-form-item > el-input 手机号 + el-input 验证码 + el-button 获取验证码 + el-button 注册按钮）。

具体修改如下：

1. 将 `el-tab-pane label="微信登录" name="wechat"` 改为 `el-tab-pane label="注册" name="register"`

2. 替换 `<div class="wechat-placeholder">...</div>` 为以下 el-form 结构（完全复制 Tab 1 表单）：
```html
<el-form @submit.prevent="handleSmsLogin">
  <el-form-item>
    <el-input
      v-model="smsForm.phone"
      placeholder="请输入手机号"
      maxlength="11"
      type="number"
      :prefix-icon="User"
    />
  </el-form-item>
  <el-form-item>
    <div class="code-row">
      <el-input
        v-model="smsForm.code"
        placeholder="请输入验证码"
        maxlength="6"
        style="width: 60%"
        :prefix-icon="Lock"
      />
      <el-button
        :disabled="countdown > 0"
        style="width: 38%"
        @click="handleSendCode"
      >
        {{ countdown > 0 ? `已发送(${countdown}s)` : '获取验证码' }}
      </el-button>
    </div>
  </el-form-item>
  <el-form-item>
    <el-button type="primary" style="width: 100%; height: 44px; font-size: 16px" @click="handleSmsLogin">
      注册
    </el-button>
  </el-form-item>
</el-form>
```

3. 保留 `.wechat-placeholder` CSS class（在样式块底部保留或删除均可，不影响功能）

关键约束：
- 复用 `smsForm`（ref: { phone: '', code: '' }）、`countdown`、`handleSendCode`、`handleSmsLogin`、`handleLoginSuccess`
- 不新增任何 form ref 或独立的 registerForm
- 按钮文字必须是「注册」（不是「登录」）
- 不修改 Tab 1 和 Tab 2
</action>
  <accept_criteria>
    - LoginView.vue 包含 el-tab-pane label="注册" name="register"（grep 可验证）
    - LoginView.vue 包含 el-button 文字为「注册」（grep 可验证）
    - LoginView.vue 包含 v-model="smsForm.phone" 和 v-model="smsForm.code" 在注册 Tab 内
    - LoginView.vue 不包含「微信登录」Tab（grep 可验证）
    - LoginView.vue 包含 handleSendCode 和 handleSmsLogin 的调用（已复用）
    - 无新增 form ref，无新增独立的 registerForm
  </accept_criteria>
  <verify>
    <automated>
      grep -n 'label="注册"' frontend/src/views/layout/LoginView.vue && \
      grep -n '微信登录' frontend/src/views/layout/LoginView.vue | wc -l | xargs test "0" = && \
      grep -n '注册' frontend/src/views/layout/LoginView.vue | head -5
    </automated>
  </verify>
  <done>Tab 3「注册」Tab 可见；表单可填写手机号+验证码；点击「注册」调用 handleSendCode/handleSmsLogin；onboarding_required=true 跳 /onboarding/org-setup</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| {client -> backend} | 用户输入（手机号、验证码）经 axios 发送到 /auth/send-code 和 /auth/login |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-03a-01 | Tampering | LoginView.vue | N/A | 纯前端UI修改，无数据篡改风险 |
| T-03a-02 | Information Disclosure | 手机号输入 | mitigate | 手机号输入为业务必需，后端已有手机号存储合规要求 |
| T-03a-03 | Denial of Service | 验证码发送 | mitigate | 后端已有 rate limit（前端复用已有 handleSendCode，不新增风险） |
| T-03a-04 | Elevation of Privilege | 注册流程 | accept | D-10/D-11 已知局限：已注册手机号视为登录；Phase 后续迭代处理 |
</threat_model>

<verification>
- Tab 3 标签显示为「注册」
- 点击「注册」Tab 后，表单可见（手机号输入框 + 验证码输入框 + 获取验证码按钮 + 注册按钮）
- 输入手机号并点击「获取验证码」：倒计时 60 秒正常启动（参考 Tab 1 行为）
- 输入正确验证码后点击「注册」：调用 POST /auth/login
- 后端返回 onboarding_required=true 时：跳转到 /onboarding/org-setup
- 视觉样式与 Tab 1 一致（品牌色 #1677ff、44px 高按钮）
</verification>

<success_criteria>
- [ ] Tab 3 标签从「微信登录」改为「注册」
- [ ] 注册表单结构与 Tab 1「手机验证码」完全相同
- [ ] 注册按钮文案为「注册」（非「登录」）
- [ ] 复用 handleSendCode 和 handleSmsLogin（D-07）
- [ ] 复用 smsForm 和 countdown 状态
- [ ] handleLoginSuccess 已含 onboarding_required 分流，注册成功后跳 /onboarding/org-setup
- [ ] 无新增后端接口（D-06）
- [ ] 无新增依赖或包
</success_criteria>

<output>
After completion, create `.planning/phases/03a-web/03a-01-SUMMARY.md`
</output>
