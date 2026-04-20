# Roadmap: 易人事（EasyHR）

## Milestones

- ✅ **v1.0 MVP** — 9 phases, 27 plans (shipped 2026-04-11)
- ✅ **v1.1** — 老板登录界面优化 (shipped 2026-04-13)
- ✅ **v1.2** — H5 管理后台 UI 重构 (shipped 2026-04-14)
- ✅ **v1.3** — 产品功能全面优化（基于 PRD 1.1）(shipped 2026-04-19)
- 🔄 **v1.4** — 用户体验 + 合规增强 (in progress)

## Phases

<details>
<summary>✅ v1.0 MVP — SHIPPED 2026-04-11</summary>

- [x] Phase 1: 基础框架与用户认证 (4 plans) — 2026-04-06
- [x] Phase 2: 员工管理 (4 plans) — 2026-04-07
- [x] Phase 3: 社保管理 (3 plans) — 2026-04-07
- [x] Phase 4: 个税计算 (2 plans) — 2026-04-07
- [x] Phase 5: 工资核算 (3 plans) — 2026-04-09
- [x] Phase 6: 财务记账 (4 plans) — 2026-04-10
- [x] Phase 7: 首页工作台 (2 plans) — 2026-04-10
- [x] Phase 8: 员工微信小程序 (2 plans) — 2026-04-11
- [x] Phase 9: v1.0 收尾 (3 plans) — 2026-04-11

详细归档见: [.planning/milestones/v1.0-ROADMAP.md](.planning/milestones/v1.0-ROADMAP.md)

</details>

<details>
<summary>✅ v1.1 老板登录界面优化 — SHIPPED 2026-04-13</summary>

- [x] Phase 01: 新增老板专属 H5 登录页 (4 plans) — 2026-04-11
- [x] Phase 02: 老板账户登陆 (deferred，范围待定)
- [x] Phase 03a: Web 注册界面 (1 plan) — 2026-04-11

详细归档见: [.planning/milestones/v1.1-ROADMAP.md](.planning/milestones/v1.1-ROADMAP.md)

</details>

<details>
<summary>✅ v1.2 H5 管理后台 UI 重构 — SHIPPED 2026-04-14</summary>

- [x] Phase 1: 登录页 + 布局基础 (2 plans) — 2026-04-14
- [x] Phase 2: 首页仪表盘 + 员工管理 — shipped
- [x] Phase 3: 薪资管理 + 社保管理 — shipped
- [x] Phase 4: 考勤 + 审批 + 补充页面 — shipped

详细归档见: [.planning/milestones/v1.2-ROADMAP.md](.planning/milestones/v1.2-ROADMAP.md)

</details>

<details>
<summary>✅ v1.3 产品功能全面优化（基于 PRD 1.1）— SHIPPED 2026-04-19</summary>

- [x] Phase 5: 员工管理增强 + 组织架构基础 (5 plans) — 2026-04-18
- [x] Phase 6: 考勤管理 (4 plans) — 2026-04-18
- [x] Phase 7: 薪资管理增强 (4 plans) — 2026-04-18
- [x] Phase 8: 社保公积金增强 (4 plans) — 2026-04-19
- [x] Phase 9: 待办中心 (3 plans) — 2026-04-19

详细归档见: [.planning/milestones/v1.3-ROADMAP.md](.planning/milestones/v1.3-ROADMAP.md)

</details>

<details>
<summary>🔄 v1.4 用户体验 + 合规增强 — IN PROGRESS (2/4 phases)</summary>

- [x] **Phase 10: UX 基础 - 流程简化与引导体系** — 2026-04-20
- [x] **Phase 11: 合同合规** — 2026-04-20
- [ ] **Phase 12: 考勤合规报表** — 待开始
- [ ] **Phase 13: 工资合规** — 待开始

</details>

---

## Phase Details

### Phase 10: UX 基础 - 流程简化与引导体系

**Goal**: 用户操作步骤减少到 3 步以内，获得清晰的首次引导和错误处理

**Depends on**: Phase 9 (v1.3 last phase)

**Requirements**: UX-01, UX-02, UX-03, UX-04, UX-05, UX-06, UX-07, UX-08, UX-09

**Success Criteria** (what must be TRUE):

1. 新员工入职：老板从"新增员工"到员工收到邀请短信，步骤不超过 3 步（填写基本信息 -> 选择入职日期 -> 确认发送）
2. 批量操作：支持员工批量入职/转正/离职，同一批次可处理 50 人以上
3. 首次用户：首次登录时自动触发引导流程，60 秒内让用户知道第一个任务是什么
4. 表单填写：输入过程中实时校验，错误提示包含具体原因和修正建议（如"手机号格式错误，请输入11位数字"）
5. 操作失败：网络错误/系统异常等场景，提供一键重试或切换解决方案的操作引导
6. 空状态：每个模块（员工/考勤/薪资/社保）在无数据时，显示引导性空状态插画 + 下一步行动按钮

**Plans**: 3 plans

**Plan list**:
- [x] 10-01-PLAN.md — 员工向导3步改造 + 核心组件（StepWizard/EmptyState/useMessage/ErrorActions）— **SHIPPED**
- [x] 10-02-PLAN.md — Excel批量导入向导（3步：模板→预览→确认）— **SHIPPED**
- [x] 10-03-PLAN.md — Tour首次引导 + API错误映射 + 工具提示 — **SHIPPED**

**UI hint**: yes

### Phase 11: 合同合规

**Goal**: 劳动合同全生命周期管理，从模板创建到员工签署并存档

**Depends on**: Phase 10

**Requirements**: COMP-01, COMP-02, COMP-03, COMP-04

**Success Criteria** (what must be TRUE):

1. 用户可创建/编辑合同模板，支持标准模板（劳动合同/实习协议）和自定义模板，模板可预览和版本管理
2. 选择员工后一键生成 PDF，自动填充员工姓名/身份证/职位/薪资等关键信息
3. 员工通过短信验证码完成签署，签署过程无需额外 APP
4. 合同签署完成后自动归档，支持按员工/日期/类型检索，可下载 PDF

**Plans**: TBD
**UI hint**: yes

### Phase 12: 考勤合规报表

**Goal**: 合规要求的考勤统计报表，支持导出用于留存备查

**Depends on**: Phase 10

**Requirements**: COMP-05, COMP-06, COMP-07, COMP-08

**Success Criteria** (what must be TRUE):

1. 加班统计报表：按法定节假日加班/工作日加班/周末加班三档分类统计，区分加班时长和可调休时长
2. 请假合规报表：年假/病假/事假分类统计，标注剩余额度，支持查看历史请假记录
3. 考勤异常报表：迟到/早退/旷工按月统计，支持按部门/员工筛选，显示异常次数和时长
4. 月度考勤汇总：Excel 导出含所有员工每日打卡情况、请假、加班、异常汇总，可直接打印或留存

**Plans**: TBD
**UI hint**: yes

### Phase 13: 工资合规

**Goal**: 工资条发放确认全流程管理，确保工资发放有据可查

**Depends on**: Phase 10

**Requirements**: COMP-09, COMP-10, COMP-11

**Success Criteria** (what must be TRUE):

1. 工资条明细中包含"确认回执"按钮，员工查看工资条后可一键确认"已收到"
2. 工资条确认记录自动存档，含员工姓名/月份/确认时间/IP地址
3. 未确认工资条自动提醒：系统每日检测未确认工资条，通过待办中心推送提醒老板

**Plans**: TBD
**UI hint**: yes

---

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. 基础框架与用户认证 | v1.0 | 4/4 | Complete | 2026-04-06 |
| 2. 员工管理 | v1.0 | 4/4 | Complete | 2026-04-07 |
| 3. 社保管理 | v1.0 | 3/3 | Complete | 2026-04-07 |
| 4. 个税计算 | v1.0 | 2/2 | Complete | 2026-04-07 |
| 5. 工资核算 | v1.0 | 3/3 | Complete | 2026-04-09 |
| 6. 财务记账 | v1.0 | 4/4 | Complete | 2026-04-10 |
| 7. 首页工作台 | v1.0 | 2/2 | Complete | 2026-04-10 |
| 8. 员工微信小程序 | v1.0 | 2/2 | Complete | 2026-04-11 |
| 9. v1.0 收尾 | v1.0 | 3/3 | Complete | 2026-04-11 |
| 01. 新增老板专属 H5 登录页 | v1.1 | 4/4 | Complete | 2026-04-11 |
| 03a. Web 注册界面 | v1.1 | 1/1 | Complete | 2026-04-11 |
| 1. 登录页 + 布局基础 | v1.2 | 2/2 | Complete | 2026-04-14 |
| 2. 首页仪表盘 + 员工管理 | v1.2 | TBD | Complete | - |
| 3. 薪资管理 + 社保管理 | v1.2 | TBD | Complete | - |
| 4. 考勤 + 审批 + 补充页面 | v1.2 | TBD | Complete | - |
| 5. 员工管理增强 + 组织架构基础 | v1.3 | 5/5 | Complete | 2026-04-18 |
| 6. 考勤管理 | v1.3 | 4/4 | Complete | 2026-04-18 |
| 7. 薪资管理增强 | v1.3 | 4/4 | Complete | 2026-04-18 |
| 8. 社保公积金增强 | v1.3 | 4/4 | Complete | 2026-04-19 |
| 9. 待办中心 | v1.3 | 3/3 | Complete | 2026-04-19 |
| 10. UX 基础 - 流程简化与引导体系 | v1.4 | 3/3 | Complete | 2026-04-20 |
| 11. 合同合规 | v1.4 | 2/2 | Complete | 2026-04-20 |
| 12. 考勤合规报表 | v1.4 | 0/TBD | Not started | - |
| 13. 工资合规 | v1.4 | 0/TBD | Not started | - |

---

_v1.0 + v1.1 + v1.2 + v1.3 + Phase 10 (v1.4) 全部交付。Roadmap created: 2026-04-06_
