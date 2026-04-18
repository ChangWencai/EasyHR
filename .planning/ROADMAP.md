# Roadmap: 易人事（EasyHR）

## Milestones

- ✅ **v1.0 MVP** — 9 phases, 27 plans (shipped 2026-04-11)
- ✅ **v1.1** — 老板登录界面优化 (shipped 2026-04-13)
- ✅ **v1.2** — H5 管理后台 UI 重构 (shipped 2026-04-14)
- 🚧 **v1.3** — 产品功能全面优化（基于 PRD 1.1）(active)

## Phases

**Phase Numbering:**
- Phases 1-9: v1.0 MVP (shipped)
- Phases 01-03a: v1.1 (shipped)
- Phases 1-4: v1.2 (shipped)
- Phases 5-9: v1.3 (this milestone)
- Decimal phases (5.1, 5.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 5: 员工管理增强 + 组织架构基础** - 数据看板、花名册增强、组织架构可视化、员工信息登记、办离职优化、下载导出
- [ ] **Phase 6: 考勤管理** - 打卡设置（3种模式）、今日打卡实况、审批流引擎（11种审批类型）、出勤月报
- [ ] **Phase 7: 薪资管理增强** - 数据看板、调薪/普调、个税上传、绩效系数、薪资算法增强、发工资条、下载导出
- [ ] **Phase 8: 社保公积金增强** - 数据看板、增减员优化、缴费渠道与状态管理、提醒与明细
- [ ] **Phase 9: 待办中心** - 待办事项汇总、轮播图与快捷办事、限时任务引擎、完成率统计

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

---

### 🚧 v1.3 产品功能全面优化（基于 PRD 1.1）(In Progress)

**Milestone Goal:** 根据 PRD 1.1 对现有产品进行功能优化和补全，新增待办中心、考勤管理、完善薪资/社保/员工管理模块

## Phase Details

### Phase 5: 员工管理增强 + 组织架构基础
**Goal**: 管理员获得完整的员工数据洞察和组织管理能力，部门模型为后续薪资普调和考勤排班提供基础维度
**Depends on**: v1.2 (shipped)
**Requirements**: EMP-01, EMP-02, EMP-03, EMP-04, EMP-05, EMP-06, EMP-07, EMP-08, EMP-09, EMP-10, EMP-11, EMP-12, EMP-13, EMP-14, EMP-15, EMP-16
**Success Criteria** (what must be TRUE):
  1. 管理员可在员工数据看板看到在职人数、当月新入职/离职人数和离职率
  2. 管理员可通过组织架构可视化图表按部门/岗位/员工层级浏览和检索
  3. 管理员可创建员工信息登记表并转发给员工填写，提交后自动更新员工档案
  4. 管理员可审批离职申请，通过后一键跳转社保减员，减员完成自动更新离职状态
  5. 花名册展示完整信息（状态/岗位薪资/在职年限/合同到期/手机号），支持搜索和 Excel 导出
**Plans**: 5 plans

Plans:
- [x] 05-01: 员工数据看板（4张纯数字卡片 + 离职率计算）+ Phase 5 全部前端路由注册
- [x] 05-02: 组织架构可视化（Department 模块 + ECharts tree）
- [x] 05-03: 员工信息登记（Registration Token + H5 填写页 + 转发）
- [ ] 05-04: 办离职优化（驳回 + 行内审批 + 去减员跳转）
- [ ] 05-05: 花名册增强（多列聚合 + Drawer 详情 + Excel 导出增强）

**UI hint**: yes

### Phase 6: 考勤管理
**Goal**: 管理员可配置灵活的打卡规则，员工可提交各类假勤审批，管理员获得出勤月报数据驱动薪资核算
**Depends on**: Phase 5 (部门模型用于排班分配)
**Requirements**: ATT-01, ATT-02, ATT-03, ATT-04, ATT-05, ATT-06, ATT-07, ATT-08, ATT-09, ATT-10, ATT-11, ATT-12, ATT-13, ATT-14, ATT-15, ATT-16, ATT-17, ATT-18, ATT-19, ATT-20, ATT-21
**Success Criteria** (what must be TRUE):
  1. 管理员可设置 3 种打卡模式（固定时间/按排班/自由工时），配置工作日、节假日、打卡位置和方式
  2. 员工可提交 5 种审批申请（补卡/调班/出差/外出/请假），请假支持 7 种类型并自动计算时长
  3. 管理员可在今日打卡实况查看全员打卡状态，点击员工查看假勤统计并手动修正
  4. 管理员可审批考勤相关申请（同意/驳回），审批列表显示待办条数
  5. 管理员可查看出勤月报（实际/应出勤/加班时长），查看每日打卡详情，并导出 Excel
**Plans**: TBD

Plans:
- [ ] 06-01: TBD
- [ ] 06-02: TBD
- [ ] 06-03: TBD
- [ ] 06-04: TBD

**UI hint**: yes

### Phase 7: 薪资管理增强
**Goal**: 管理员获得完整的薪资数据洞察，调薪/普调按部门或个人灵活操作，薪资算法集成考勤数据自动核算
**Depends on**: Phase 5 (部门模型用于普调), Phase 6 (考勤数据驱动薪资计算)
**Requirements**: SAL-01, SAL-02, SAL-03, SAL-04, SAL-05, SAL-06, SAL-07, SAL-08, SAL-09, SAL-10, SAL-11, SAL-12, SAL-13, SAL-14, SAL-15, SAL-16, SAL-17, SAL-18, SAL-19
**Success Criteria** (what must be TRUE):
  1. 管理员可在薪资数据看板看到当月应发总额、实发总额、社保公积金总额、个税总额，均带环比上月百分比
  2. 管理员可对选定员工调薪或选定部门普调（支持金额或比例），按生效期限自动应用于工资核算
  3. 管理员可上传个税 Excel 文件，系统自动抓取关键字段并更新当月工资表
  4. 管理员可为员工设置绩效系数（0%-100%），自动挂钩绩效工资计算
  5. 薪资算法自动集成考勤数据：基本工资按计薪天数计算、病假工资按工龄系数、加班费按法定系数，管理员可查看税前工资明细
  6. 管理员可向全员或选定员工发送工资条，薪资列表支持 Excel 导出
**Plans**: TBD

Plans:
- [ ] 07-01: TBD
- [ ] 07-02: TBD
- [ ] 07-03: TBD
- [ ] 07-04: TBD

**UI hint**: yes

### Phase 8: 社保公积金增强
**Goal**: 管理员获得社保公积金数据洞察，增减员流程优化，缴费渠道和欠缴状态自动管理
**Depends on**: Phase 5 (离职减员联动)
**Requirements**: SI-01, SI-02, SI-03, SI-04, SI-05, SI-06, SI-07, SI-08, SI-09, SI-10, SI-11, SI-12, SI-13, SI-14, SI-15, SI-16, SI-17, SI-18, SI-19, SI-20, SI-21
**Success Criteria** (what must be TRUE):
  1. 管理员可在社保数据看板看到应缴总额（单位+个人+欠缴）、单位/个人分项、欠缴金额，均带环比百分比
  2. 管理员在增员弹窗可按姓名检索，设置起始月份、缴费城市、社保基数、公积金比例和基数
  3. 管理员在减员弹窗可按姓名检索，设置终止月份、减员原因、转出/封存日期，系统提示生效规则
  4. 管理员可选择缴费渠道（自主/代理），缴费状态按月自动流转（正常/待缴/欠缴/已转出/未转出）
  5. 管理员可看到关键节点红字提醒和政策通知，社保明细展示五险分项金额，列表支持 Excel 导出
**Plans**: TBD

Plans:
- [ ] 08-01: TBD
- [ ] 08-02: TBD
- [ ] 08-03: TBD

**UI hint**: yes

### Phase 9: 待办中心
**Goal**: 管理员拥有统一的事项聚合入口，系统自动生成限时任务，快捷办事直达核心操作
**Depends on**: Phase 5 (员工数据), Phase 6 (考勤审批数据), Phase 7 (薪资任务), Phase 8 (社保任务)
**Requirements**: TODO-01, TODO-02, TODO-03, TODO-04, TODO-05, TODO-06, TODO-07, TODO-08, TODO-09, TODO-10, TODO-11, TODO-12, TODO-13, TODO-14, TODO-15, TODO-16, TODO-17, TODO-18, TODO-19, TODO-20
**Success Criteria** (what must be TRUE):
  1. 管理员可在待办中心查看全部待办汇总列表，支持关键字搜索、时间段筛选（60天内）、置顶排序和 Excel 导出
  2. 管理员可邀请协办（转发给内部或外部人），可终止待办任务
  3. 首页展示 1-3 张轮播图和快捷办事入口（新入职/调薪/考勤/个税/社保公积金），点击直达对应功能
  4. 系统自动生成 7 种限时任务（合同签约/续签、个税申报、社保缴费/增减员/基数调整），显示剩余时间和超时/失效状态
  5. 待办中心展示全部事项完成率和限时任务完成率的环形图
**Plans**: TBD

Plans:
- [ ] 09-01: TBD
- [ ] 09-02: TBD
- [ ] 09-03: TBD

**UI hint**: yes

## Progress

**Execution Order:**
Phases execute in numeric order: 5 → 6 → 7 → 8 → 9

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 5. 员工管理增强 + 组织架构基础 | v1.3 | 0/5 | Planning | - |
| 6. 考勤管理 | v1.3 | 0/? | Not started | - |
| 7. 薪资管理增强 | v1.3 | 0/? | Not started | - |
| 8. 社保公积金增强 | v1.3 | 0/? | Not started | - |
| 9. 待办中心 | v1.3 | 0/? | Not started | - |

---
_v1.0 + v1.1 + v1.2 已完整交付。v1.3 roadmap created: 2026-04-17_
