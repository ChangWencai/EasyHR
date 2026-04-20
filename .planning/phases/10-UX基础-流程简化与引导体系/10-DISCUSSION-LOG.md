# Phase 10: UX 基础 - 流程简化与引导体系 - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-20
**Phase:** 10-UX基础-流程简化与引导体系
**Areas discussed:** 员工入职流程、批量操作模式、首次使用引导、错误提示+Toast规范

---

## 员工入职流程（UX-01核心）

| Option | Description | Selected |
|--------|-------------|----------|
| 步骤向导（Step Wizard） | 向导式，用户按顺序操作。当前表单字段不变，分组分页展示，步骤条提示进度，每步提交前校验。 | ✓ |
| 极简单页（可折叠区块） | 单页完成，默认折叠薪资/紧急联系人为可选区块，主表单只保留4个必填字段，底部一键提交+发送邀请短信。 | |
| 你来决定 | 由Claude决定如何在保留现有功能完整性的同时实现3步目标。 | |

**User's choice:** 步骤向导（Step Wizard）
**Notes:** 用户明确选择了向导式交互，并进一步细化了3步分配方案

### 向导步骤分配

| Option | Description | Selected |
|--------|-------------|----------|
| 3步：基本信息 → 入职信息 → 确认发送 | Step1: 姓名+手机号+身份证 | ✓ |
| 3步：基础+日期 → 薪资 → 确认发送 | Step1: 姓名+手机号+身份证+入职日期+岗位 | |
| 3步：最小集 → 核心信息 → 确认发送 | Step1: 姓名+手机号+身份证（最小集） | |

**User's choice:** 3步：基本信息 → 入职信息 → 确认发送

### 提交后短信处理

| Option | Description | Selected |
|--------|-------------|----------|
| 自动发送（推荐） | 员工创建成功后，自动触发邀请短信。老板无需额外操作。 | |
| 手动确认发送（推荐） | 创建成功后，显示确认页，由老板手动点击「发送邀请短信」按钮。老板可在此修改手机号再发送。 | ✓ |

**User's choice:** 手动确认发送

---

## 批量操作模式（UX-02）

| Option | Description | Selected |
|--------|-------------|----------|
| Excel模板导入（推荐） | 员工列表页顶部/底部放置「批量入职」按钮，点击弹出向导（Step1: 上传Excel模板 / Step2: 数据预览+错误标注 / Step3: 一键确认发送邀请短信）。 | ✓ |
| 页面内多选批量处理 | 员工列表页顶部放置「批量入职」按钮，点击弹出Dialog/侧边栏，页面内勾选多名员工或输入多名员工信息。 | |

**User's choice:** Excel模板导入

### Excel导入错误处理

| Option | Description | Selected |
|--------|-------------|----------|
| 允许部分导入（推荐） | Excel预览页显示所有行，错误行红色标注，合格行绿色标注，底部一键「仅导入合格项」或「全部导入」。 | ✓ |
| 全部正确才能导入 | Excel预览页检测到任何错误行，全部阻止导入，用户必须修正Excel后重新上传。 | |

**User's choice:** 允许部分导入

---

## 首次使用引导（UX-03）

| Option | Description | Selected |
|--------|-------------|----------|
| Overlay点位引导（推荐） | 首页顶部依次出现标注引导点（Overlay高亮+文字说明），覆盖快捷入口→待办→下一步指引。首次登录触发，完成后不再显示。 | ✓ |
| 独立引导任务卡片页 | 首次登录跳转至独立引导页，按顺序展示「添加员工/配置考勤/发首月工资」等关键任务卡片，点击直达对应页面。 | |

**User's choice:** Overlay点位引导

### Tour引导点覆盖范围

| Option | Description | Selected |
|--------|-------------|----------|
| 3个核心引导点（推荐） | 引导覆盖：快捷入口(新增员工)→首页待办→快速上手面板。「跳过引导」随时可点。 | ✓ |
| 5个完整引导点 | 引导覆盖：快捷入口→考勤打卡→发薪记录→社保参保→首页待办→快速上手面板。 | |

**User's choice:** 3个核心引导点

### 引导触发判断

| Option | Description | Selected |
|--------|-------------|----------|
| 本地 localStorage 标记（推荐） | 首次登录（localStorage标记hasSeenTour）触发。完成后标记，永不重复显示。简单可靠，无需后端。 | ✓ |
| 后端字段标记 | 后端 User表增加 has_completed_onboarding 字段，首次登录后端检测触发。 | |

**User's choice:** 本地 localStorage 标记

---

## 错误提示+Toast规范（UX-04~06）

### API错误拦截位置

| Option | Description | Selected |
|--------|-------------|----------|
| 统一拦截（推荐） | request.ts拦截器统一处理所有API错误，根据HTTP状态码+后端error_code映射为用户友好消息。各页面只处理业务逻辑特有的错误。 | ✓ |
| 各页面自行处理 | request.ts只做通用错误处理，其他错误在各API调用处处理。灵活性高但代码分散。 | |

**User's choice:** 统一拦截

### 操作失败引导

| Option | Description | Selected |
|--------|-------------|----------|
| 显示重试+联系管理员（推荐） | 操作失败时，错误消息下方显示「重试」按钮和「联系管理员」按钮。网络错误/超时时优先显示重试。 | ✓ |
| 仅显示友好消息 | 操作失败时，仅显示友好错误消息，不提供额外操作入口。 | |

**User's choice:** 显示重试+联系管理员

### Toast统一方式

| Option | Description | Selected |
|--------|-------------|----------|
| 统一封装 useMessage（推荐） | 统一封装 useMessage() composable，内部调用ElMessage/ElNotification，duration统一规范（成功2秒/警告3秒/错误不关闭+操作按钮）。各页面替换现有零散的ElMessage调用。 | ✓ |
| 仅规范 ElMessage 调用 | 保持现有ElMessage调用，只需在项目内规范duration和类型使用规则。 | |

**User's choice:** 统一封装 useMessage

---

## Claude's Discretion

- 向导的具体动画（步骤切换时滑入/淡入）
- 引导气泡的具体文案内容（根据实际页面决定）
- 空状态SVG插画的具体设计风格
- Toast操作按钮的具体文案（重试/关闭/查看详情）
- 批量Excel模板的具体字段列定义

## Deferred Ideas

无 — 讨论保持在阶段范围内
