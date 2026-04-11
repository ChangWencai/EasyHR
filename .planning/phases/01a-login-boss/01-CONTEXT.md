# Phase 01: 新增登陆界面，该登陆界面只允许老板账户登陆 - Context

**Gathered:** 2026-04-11
**Status:** Ready for planning

<domain>
## Phase Boundary

为 H5 管理后台创建老板专属登录页，替换当前 `/login` 的占位页"功能开发中"。实现 3 种登录方式，按角色过滤（OWNER+ADMIN 允许，MEMBER 拒绝），首次登录引导录入企业信息。

Phase 01 交付完整登录功能。Phase 02 将在 Phase 01 完成后重新定义范围。

</domain>

<decisions>
## Implementation Decisions

### 登录方式
- **D-01:** 3 种登录方式并行入口：
  1. 手机号 + 短信验证码（首选，与 v1.0 Phase 1 后端实现一致）
  2. 手机号 + 密码登录（新增，需后端扩展 `/api/v1/auth/login/password` 接口）
  3. 微信授权登录（备选，适合老板有微信即可扫码登录，无需记忆密码）
- **D-02:** 3 种方式平等展示为 Tab 或入口按钮，不分主次
- **D-03:** 微信授权走静默授权（已绑定手机的微信号）或手机号绑定流程（首次授权后要求绑定手机号）

### 非老板账户处理
- **D-04:** OWNER（老板）→ 允许登录，正常进入
- **D-05:** ADMIN（管理员）→ 允许登录，正常进入（H5 后台管理员可执行大部分操作）
- **D-06:** MEMBER（普通成员）→ 拒绝登录，显示提示："您的账号为员工账号，请使用员工端微信小程序登录"，不跳转

### 登录后跳转逻辑
- **D-07:** 首次登录（用户无企业信息）→ 跳转 `/onboarding/org-setup` 录入企业信息
- **D-08:** 后续登录（已完成企业信息录入）→ 跳转 `/home` 首页
- **D-09:** 登录成功后同步更新 `stores/user.ts`（user info + org info），Auth Guard 读取 store 判断是否已登录

### Auth Guard 路由守卫
- **D-10:** 全局路由守卫（`router/index.ts` 或 `main.ts`）检查登录状态
- **D-11:** 未登录状态下访问 AppLayout 子路由 → 重定向到 `/login`
- **D-12:** 登录页本身不做守卫（无需重定向）

### 登录页视觉风格
- **D-13:** 极简商务风：纯色背景（品牌蓝 #1677ff 或浅灰 #F5F7FA）+ 居中卡片布局
- **D-14:** 顶部品牌 Logo "易人事"，底部版权信息
- **D-15:** 表单元素使用 Element Plus 组件（el-input、el-button、el-tabs）
- **D-16:** 移动端全屏覆盖，桌面端居中卡片（max-width ~400px）

### 短信验证码
- **D-17:** 后端已有 Redis 存储验证码逻辑（`sms:code:{phone}`，TTL 5分钟）
- **D-18:** 前端获取验证码按钮：点击后 60 秒倒计时，显示"已发送(60s)"，倒计时结束后恢复"获取验证码"

### MEMBER 拒绝提示
- **D-19:** 拒绝文案："您的账号为员工账号，请使用员工端微信小程序登录"
- **D-20:** 拒绝方式：后端返回 403 Forbidden，前端显示 Element Plus Message 提示，不跳转

### Phase 02 定位
- **D-21:** Phase 02 将在 Phase 01 交付后重新定义范围（当前 ROADMAP 描述与 Phase 01 重复）

### Claude's Discretion
- 登录页具体 CSS 细节（阴影、圆角、间距）
- 微信授权的具体 UI（弹窗 vs 新标签页）
- 短信验证码倒计时 UI（按钮样式变化）
- 错误提示的具体文案措辞
- WeChat OAuth 的静默授权 vs 强制授权策略细节

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### 现有代码
- `frontend/src/stores/auth.ts` — 当前 auth store：token 管理、setToken、logout（基于 localStorage）
- `frontend/src/stores/user.ts` — user store：user info（id/name/phone/role）、org info（name/credit_code/city）
- `frontend/src/api/request.ts` — axios 封装：token 自动注入、401 → /login 重定向
- `frontend/src/router/index.ts` — 路由配置：`/login` → PlaceholderView，`/onboarding/org-setup` → OrgSetup.vue
- `frontend/src/views/layout/AppLayout.vue` — 侧边栏布局（已完成的完整导航，品牌色 #1677ff）
- `frontend/src/views/onboarding/OrgSetup.vue` — 企业信息录入引导页（首次登录目标页）

### 后端（参考）
- v1.0 Phase 1 CONTEXT.md — 认证流程决策（D-05 手机验证码、D-27 多设备登录、D-40 RBAC）
- `.planning/phases/01-foundation-auth/1-CONTEXT.md` — 认证层实现决策

### 产品规范
- `prd.md` — 产品需求文档（参考 AUTH-01 ~ AUTH-04）
- `frontend/src/App.vue` — 新手引导 Overlay（首次进入时触发，品牌蓝色 #1677ff）

</canonical_refs>

<codebase_context>
## Existing Code Insights

### Reusable Assets
- `stores/auth.ts` — token 管理基础已就绪，登录成功后调用 `setToken` 即可
- `stores/user.ts` — user + org info 结构已定义，登录后填充
- `request.ts` — 401 拦截器已实现，会自动跳转 /login
- Element Plus — 全局已引入（el-input/button/tabs/message/loading 等直接使用）

### Established Patterns
- 品牌色：#1677ff（Element Plus 蓝），与 AppLayout 保持一致
- 移动优先 + 桌面适配：AppLayout 已实现响应式，登录页应保持一致
- Pinia store：useAuthStore / useUserStore 已建立，登录后填充

### Integration Points
- `/login` 路由 → 新建 `LoginView.vue` 替换 PlaceholderView
- AppLayout 路由 → 需加 Auth Guard（检查 isLoggedIn）
- `/onboarding/org-setup` → 已有，企业信息录入引导页（Phase 1 已完成）
- `stores/user.ts` → 登录后需调用 `setUser` + `setOrg` 填充用户信息
- 微信授权 → 需调用微信 OAuth2 接口，需后端配合（/api/v1/auth/login/wechat）

</codebase_context>

<specifics>
## Specific Ideas

No specific references from discussion — open to standard approaches

</specifics>

<deferred>
## Deferred Ideas

### Phase 02 重新定义
- Phase 02 在 ROADMAP 中描述与 Phase 01 相同（均为"老板专属登录页"）。Phase 01 交付完整登录后，需重新评估 Phase 02 的实际需求再定义范围。

</deferred>

---
*Phase: 01-login-boss*
*Context gathered: 2026-04-11*
