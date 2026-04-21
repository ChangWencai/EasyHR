# Milestones

## v1.4 用户体验 + 合规增强 (Shipped: 2026-04-21)

**Phases completed:** 5 phases, 12 plans
**Known deferred items at close:** 11 (see STATE.md Deferred Items)

**Key accomplishments:**

1. **员工向导 3 步改造**: StepWizard 组件 + Excel 批量导入向导 + Tour 首次引导
2. **中文 PDF 合同电子签署**: go-pdf/fpdf 中文渲染 + 手机验证码签署 + 合同存档
3. **考勤合规报表**: 加班/请假/出勤异常/月度汇总 4 类报表 + Excel 导出
4. **工资条确认回执**: confirmed 状态 + gocron 每日提醒 + asynq 异步任务
5. **组织架构图增强**: Position 岗位 CRUD + ECharts 拖拽 + 右键菜单 + 员工岗位下拉

---

## v1.3 产品功能全面优化 (Shipped: 2026-04-19)

**Phases completed:** 5 phases, 20 plans

**Key accomplishments:**

1. 员工管理增强（看板视图、组织架构图、登记表、离职流程）
2. 考勤管理（打卡、审批流状态机、出勤月报）
3. 薪资管理增强（调薪 INSERT ONLY、个税引擎、绩效、工资条发送）
4. 社保公积金增强（增减员、缴费渠道、状态流转）
5. 待办中心（限时任务、轮播图、协办邀请）

---

## v1.2 H5 管理后台 UI 重构 (Shipped: 2026-04-14)

**Phases completed:** 4 phases

**Key accomplishments:**

1. H5 管理后台按 EasyHR-web.pen 原型图全面重构
2. 主色调 #4F6EF7，卡片圆角 12px，侧边栏固定 220px
3. 员工/薪资/社保/考勤/审批等核心页面完成

---

## v1.1 老板登录界面优化 (Shipped: 2026-04-13)

**Phases completed:** 3 phases, 9 plans, 9 tasks

**Key accomplishments:**

1. LoginView.vue Tab 3 改为「注册」Tab，手机号+验证码表单，复用已有后端接口

---

## v1.0 MVP — 易人事（EasyHR）

**Shipped:** 2026-04-11
**Phases:** 9 | **Plans:** 27

### Key Accomplishments

1. **Go 后端完整交付**: 认证/员工/社保/个税/工资/财务 6 大模块，40+ Go 文件，7 测试包全部 PASS
2. **Vue 3 H5 管理后台**: 底部 5-Tab 导航、首页工作台、员工/工具/财务/我的 完整 Tab 实现
3. **微信小程序员工端**: 5-tab 员工端，登录、工资条（含短信验证）、合同、社保、报销
4. **Excel 导出**: 账簿导出 + 纳税申报表导出（excelize）
5. **城市自动定位**: OrgSetup 页接入 ipapi.co，支持 IP 地理位置自动填充城市

### What Was Shipped

- 三级 RBAC + JWT 认证 + 多租户隔离
- 员工全生命周期管理（入职邀请/档案/离职）
- 社保政策库 + 参保/停缴/提醒
- 累计预扣法个税引擎 + 专项附加扣除
- 薪资结构自定义 + 一键核算 + 工资单推送
- 财务记账完整流程（凭证/发票/报销/账簿/报表/结账）
- 首页工作台（待办卡片 + 5 宫格入口 + 数据概览）

### Stats

- **Timeline**: 2026-04-06 → 2026-04-11 (~5 天)
- **Execution time**: ~7 hours total
- **Phases**: 9 (全部完成)
- **Plans**: 27/27
- **Requirements**: 57/57 v1 全部达成
