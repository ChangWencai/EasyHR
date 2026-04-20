# Phase 10: UX 基础 - 流程简化与引导体系 - Context

**Gathered:** 2026-04-20
**Status:** Ready for planning

<domain>
## Phase Boundary

通过流程精简（核心操作≤3步）、首次使用引导、友好错误处理、空状态设计，提升老板操作体验。

具体包含：
- 员工入职流程改造（步骤向导，3步完成）
- 批量操作框架（Excel模板导入，支持批量入职/离职/发薪/参保）
- 首次使用Tour（Overlay点位引导，3个核心引导点）
- 表单实时校验 + API错误统一映射 + 操作失败解决方案引导
- 各模块空状态设计
- Toast通知统一优化
- 关键页面工具提示（el-tooltip）

**Scope:** H5管理后台，前端为主，后端API配合。
**Depends on:** Phase 9（待办中心）

</domain>

<decisions>
## Implementation Decisions

### 员工入职向导（UX-01）
- **D-10-01:** 员工新增改为步骤向导（Step Wizard），每步单独显示一张卡片，顶部步骤条指示进度
- **D-10-02:** 步骤分配：Step1=姓名+手机号+身份证（必填基本信息）/ Step2=入职日期+岗位+薪资（入职信息）/ Step3=紧急联系人（可选）+确认提交+发送短信
- **D-10-03:** 第3步「确认发送」采用手动确认模式：创建成功后，显示确认页，由老板手动点击「发送邀请短信」按钮

### 批量操作框架（UX-02）
- **D-10-04:** 批量操作统一采用Excel模板导入模式，在员工列表/薪资/社保等页面顶部/底部放置「批量XXX」入口按钮
- **D-10-05:** 批量导入流程：Step1=下载模板上传Excel / Step2=数据预览+错误标注（合格行绿标，错误行红标）/ Step3=一键确认导入
- **D-10-06:** Excel预览页支持「仅导入合格项」，允许部分导入；错误信息标注到具体行和字段

### 首次使用Tour（UX-03）
- **D-10-07:** 采用Overlay点位引导形式（高亮遮罩+文字气泡），在首页覆盖快捷入口→待办→下一步指引
- **D-10-08:** Tour引导点共3个：快捷入口(新增员工)、首页待办、快捷入口的整体说明
- **D-10-09:** 首次登录触发Tour（localStorage标记hasSeenTour=true），完成后标记，永不重复显示；「跳过引导」随时可点

### 表单校验（UX-04）
- **D-10-10:** 实时校验策略：el-form rules trigger='blur'（输入完成时校验）；el-input maxlength即时限制；手机号/身份证等使用正则pattern
- **D-10-11:** 校验消息包含具体原因和修正建议（如「手机号格式错误，请输入11位数字」）

### API错误统一处理（UX-05）
- **D-10-12:** API错误在request.ts拦截器统一处理（集中式），根据HTTP状态码+后端error_code映射为用户友好消息；各页面只处理业务逻辑特有的错误

### 操作失败引导（UX-06）
- **D-10-13:** 操作失败时，错误消息下方显示「重试」按钮和「联系管理员」按钮；网络错误/超时时优先显示重试

### 空状态设计（UX-07）
- **D-10-14:** 各模块（员工/考勤/薪资/社保）设计统一空状态组件，含引导性插画（SVG）+ 描述文案 + 下一步行动按钮

### Toast统一优化（UX-08）
- **D-10-15:** 统一封装 `useMessage()` composable，替换所有零散的ElMessage调用；duration规范：success/info=2秒 / warning=3秒 / 错误=不自动关闭+操作按钮
- **D-10-16:** 成功消息显示2秒自动关闭；普通错误不自动关闭，提供操作按钮

### 工具提示（UX-09）
- **D-10-17:** 关键页面使用el-tooltip添加hover提示，tooltip内容简洁，聚焦操作指引而非概念解释

### Claude's Discretion
- 向导的具体动画（步骤切换时滑入/淡入）
- 引导气泡的具体文案内容（根据实际页面决定）
- 空状态SVG插画的具体设计风格
- Toast操作按钮的具体文案（重试/关闭/查看详情）
- 批量Excel模板的具体字段列定义

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements
- `.planning/REQUIREMENTS.md` §v1.4 — UX-01~UX-09 需求定义（9个UX需求）
- `.planning/ROADMAP.md` §Phase 10 — 阶段目标、成功标准、依赖关系

### Existing Code Patterns
- `frontend/src/views/employee/EmployeeCreate.vue` — 现有员工创建表单，需改造为步骤向导
- `frontend/src/api/request.ts` — API拦截器，需扩展错误映射逻辑
- `frontend/src/views/onboarding/OrgSetup.vue` — 现有注册引导（非Tour，可参考其样式）
- `frontend/src/views/home/HomeView.vue` — 首页，Tour引导点在此覆盖
- `frontend/src/stores/dashboard.ts` — DashboardStore，待办数据用于首页引导
- `.planning/phases/09-待办中心/09-CONTEXT.md` — Phase 09决策（环形图/轮播图/asynq定时任务模式）
- `.planning/phases/08-社保公积金增强/08-CONTEXT.md` — Phase 08决策（asynq批量操作框架）

### Project Decisions
- `.planning/PROJECT.md` — 核心价值：3步内完成核心操作，零学习成本
- `.planning/PROJECT.md` — 技术栈：Vue 3 + Element Plus + asynq + Go + Gin + GORM

### Tech Stack
- `xlsx (SheetJS)`（前端已安装）— Excel导入解析
- `Element Plus el-tooltip`（已使用）— 工具提示
- `ElMessage/ElNotification`（已使用）— Toast通知，需统一封装

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `EmployeeCreate.vue`: 已有表单结构和样式，可复用section-header/glass-card样式，改造为多步骤向导
- `OrgSetup.vue`: 步骤感较弱但有el-form基础结构，可参考表单验证模式
- `request.ts`: 现有拦截器只有401处理，是扩展统一错误映射的最佳位置
- `HomeView.vue`: 快捷入口grid和待办列表是Tour引导的目标元素
- `asynq批量操作模式`（Phase 08）: 批量导入后端处理可复用Phase 08的scheduler框架

### Established Patterns
- Element Plus el-form + rules: 表单校验已有实现
- ElMessage: toast已有使用，但零散不规范
- glass-card: 样式组件已建立，向导卡片可复用
- section-header: 分组标题样式已建立

### Integration Points
- `EmployeeCreate.vue` → 员工列表（创建后跳转）
- `EmployeeCreate.vue` → 短信服务（Step3发送邀请）
- 各模块列表页 → 批量导入入口
- `HomeView.vue` → Tour引导覆盖层
- `request.ts` → 全局API错误拦截

</code_context>

<specifics>
## Specific Ideas

- 员工向导Step1: 姓名、手机号、身份证（必填）；Step2: 入职日期、岗位、薪资（必填）；Step3: 紧急联系人（可选）、确认提交、手动发送短信
- Tour引导点1: 快捷入口（标注"新增员工"按钮）；引导点2: 首页待办（标注待办数量）；引导点3: 底部说明（"60秒内完成第一个任务"）
- 批量Excel导入支持50人+：预览页使用虚拟滚动，错误行红色标注，合格行绿色标注
- useMessage() composable 签名：useMessage().success(msg)/error(msg, {showActions: true})/warning(msg)/info(msg)
- 空状态组件：`EmptyState.vue`，props: {title, description, actionText, actionRoute}

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 10-UX基础-流程简化与引导体系*
*Context gathered: 2026-04-20*
