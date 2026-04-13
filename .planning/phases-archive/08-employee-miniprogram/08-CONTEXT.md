# Phase 8: 员工微信小程序 - Context

**Gathered:** 2026-04-10
**Status:** Ready for planning

<domain>
## Phase Boundary

员工通过微信小程序查看工资条（含明细）、合同状态、社保记录，提交费用报销（拍照上传票据）。覆盖 WXMP-01 ~ WXMP-06 全部需求。员工端独立于老板 H5 管理后台，是独立用户角色（MEMBER）。

</domain>

<decisions>
## Implementation Decisions

### WXMP-01: 登录与身份认证
- **D-01:** 员工通过**手机号+验证码**登录（与 AUTH-01 类似但独立路径）。无需预先注册，由老板在管理后台添加员工记录时自动创建账号。
- **D-02:** 登录后可**绑定微信头像昵称**（可选，非强制）。绑定后下次可使用微信一键登录（微信授权获取 openid → 关联已有手机号）。
- **D-03:** 小程序 JWT token 有效期与老板端一致（1小时 access + 7天 refresh）。Token 存于小程序 storage。
- **D-04:** 员工账号属于 MEMBER 角色（Phase 1 RBAC），无管理后台权限。

### WXMP-02: 工资条查看（含短信验证）
- **D-05:** 员工可查看各月工资条（按月份列表，含应发合计、实发合计）。
- **D-06:** 查看工资条**明细**（各项收入、各项扣款、实发）时必须短信验证身份：输入手机号后4位 + 短信验证码（Phase 5 D-10 已实现此模式），验证通过后当月无需再验证。
- **D-07:** 数据源：复用 Phase 5 PayrollSlip 模型，所有员工工资单数据已存在后端，无需重复存储。
- **D-08:** 员工只能查看**自己**的工资条（按当前登录员工的 user_id 过滤，与管理后台查看他人工资条不同）。

### WXMP-03: 合同状态查看
- **D-09:** 员工可查看自己关联的合同列表，展示：合同类型（劳动合同/实习协议）、状态（待签署/已签署/已过期）、签署日期、合同开始/结束日期。
- **D-10:** V1.0 降级：员工可查看合同 PDF 文件（Phase 2/Phase 8 后端提供 PDF URL），线下签署后由老板上传，不支持在线电子签。
- **D-11:** 合同数据源：Phase 2 Employee 模块的 contract_id 关联。

### WXMP-04: 社保记录查看
- **D-12:** 员工可查看自己当前参保状态和缴费明细。
- **D-13:** 展示内容：参保城市、参保基数、**个人缴费额**（明细到各险种：养老/医疗/失业）。**单位缴费额不对员工展示**（视为企业用人成本保密）。
- **D-14:** 历史记录：展示近12个月缴费明细（按月），可展开各险种分项金额。
- **D-15:** 无参保记录时（刚入职尚未参保）：显示"暂无社保缴纳记录"空状态。
- **D-16:** 数据源：Phase 3 socialinsurance 模块，按当前登录员工过滤。

### WXMP-05: 费用报销提交
- **D-17:** 员工可提交费用报销，填写：报销类型（差旅费/交通费/招待费/办公费/其他）、金额（decimal.Decimal）、说明、凭证照片（最多9张）。
- **D-18:** 票据上传：后端生成 OSS 预签名 URL → 小程序前端直接上传到 OSS，不经过服务端中转（节省流量和服务器压力）。
- **D-19:** 提交后状态：pending → approved → paid（老板审批通过后）或 rejected（老板驳回）。
- **D-20:** 报销提交后生成 ExpenseReimbursement 记录（Phase 6 D-23 已定义模型），老板在 H5 管理后台审批。

### WXMP-06: 报销状态查看
- **D-21:** 员工可查看自己所有报销单的当前状态（待审批/已通过/已支付/已驳回）。
- **D-22:** 报销驳回时展示驳回原因（老板填写的 remark）。
- **D-23:** 已支付的报销展示支付方式和支付日期（字段来自 Phase 6 ExpenseReimbursement.paid_at）。

### 技术架构
- **D-24:** 小程序采用**原生微信小程序**开发（基础库 3.x+），不使用 Taro/uni-app（CLAUDE.md 明确要求）。
- **D-25:** UI 使用 **WeUI**（微信官方设计规范）+ 内置组件。
- **D-26:** 项目建于 `miniprogram/` 目录，与 `frontend/` H5 管理后台完全独立，独立部署。
- **D-27:** 网络请求：小程序原生 `wx.request`（无需引入额外网络库），封装统一请求拦截器（自动注入 JWT token）。
- **D-27b:** 敏感页面（工资条明细）验证流程：前端请求 → 后端返回需要验证 → 前端发起短信发送 → 用户填验证码 → 前端携验证码重试 → 验证通过返回数据。

### RBAC 与权限
- **D-28:** 员工（MEMBER）只能访问自己数据：我的工资条、我的合同、我的社保、我的报销。
- **D-29:** 员工无法访问：管理后台、账簿报表、其他员工数据。
- **D-30:** Token 刷新机制：与 Phase 1 一致，access token 过期时使用 refresh token 换取新 token。

</decisions>

<deferred>
## Deferred to V2.0 or Later

- **微信模板消息推送**（报销审批结果通知）— V2.0，需要微信认证服务号
- **电子签 API 集成**（合同在线签署）— V2.0，需要对接上上签/e签宝
- **打卡记录查看**（员工查看自己考勤）— V2.0，Phase 9+ 考勤模块
- **多级审批流**（报销先经同事再经老板）— V2.0+

</deferred>

<prior_decisions>
## From Prior Phases

### Phase 1: 基础框架
- JWT token 通过 `Authorization: Bearer <token>` header 传递
- 多租户隔离：所有查询必须包含 `org_id`（TenantScope middleware）
- RBAC 三级：OWNER / ADMIN / MEMBER
- 敏感字段 AES-256-GCM 加密 + SHA-256 哈希索引

### Phase 2: 员工管理
- Employee 模型含 contract_id，关联合同
- 员工通过 H5 链接邀请入职

### Phase 3: 社保管理
- 社保参保记录按 org_id + employee_id 隔离
- 缴费记录存储各险种明细

### Phase 5: 工资核算
- PayrollSlip 模型含 token，用于 H5 查看链接
- H5 工资单 token + 短信验证模式（D-10）
- MEMBER 角色仅能查看自己工资单（不含 H5 RBAC 限制）
- 工资表状态机：draft → calculated → confirmed → paid

### Phase 6: 财务记账
- ExpenseReimbursement 模型：employee_id、amount、expense_type、description、attachments（最多9张）、status（pending/approved/rejected/paid）
- 报销审批通过后调用 financeSvc.GenerateExpenseVoucher() 生成凭证

### Phase 7: 首页工作台
- 前端 H5 项目建于 `frontend/` 目录
- 微信小程序建于 `miniprogram/` 目录（H5 和小程序各自独立）

### Project-Level (PROJECT.md)
- Core Value: 简单，好用，省时间 — 员工端同样适用，操作步骤 ≤ 3步
- 小程序使用原生开发，不用 Taro/uni-app

</prior_decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### 项目规范
- `.planning/REQUIREMENTS.md` — WXMP-01 ~ WXMP-06 需求定义
- `.planning/ROADMAP.md` — Phase 8 定义和成功标准

### 前序 Phase 上下文
- `.planning/phases/01-foundation-auth/1-CONTEXT.md` — Phase 1 全部决策（JWT、RBAC、多租户）
- `.planning/phases/02-employee-management/02-CONTEXT.md` — Phase 2 决策，Employee 模型
- `.planning/phases/03-social-insurance/03-CONTEXT.md` — Phase 3 决策，社保数据模型
- `.planning/phases/05-salary/05-CONTEXT.md` — Phase 5 决策，PayrollSlip 模型
- `.planning/phases/06-finance/06-CONTEXT.md` — Phase 6 决策，ExpenseReimbursement 模型
- `.planning/phases/07-homepage/07-CONTEXT.md` — Phase 7 决策，前端项目结构

</canonical_refs>

