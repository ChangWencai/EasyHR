# Milestones

## v1.1 老板登录界面优化 (Shipped: 2026-04-13)

**Phases completed:** 3 phases, 9 plans, 9 tasks

**Key accomplishments:**

- Plan:
- Plan:
- Plan:
- Plan:
- 1. [Rule 3 - Blocking] Fixed handler.go CreateEmployee call signature mismatch
- LoginView.vue Tab 3 改为「注册」Tab，手机号+验证码表单，复用已有后端接口

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
