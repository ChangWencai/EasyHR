# Roadmap: 易人事（EasyHR）

## Milestones

- ✅ **v1.0 MVP** — 9 phases, 27 plans (shipped 2026-04-11)
- 🚧 **v1.1** — 老板登录界面优化

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

---

**v1.0 MVP 已完整交付，包含：** Go 后端（认证/员工/社保/个税/工资/财务）+ Vue 3 H5 管理后台 + 微信小程序员工端

### Phase 01: 新增登陆界面，该登陆界面只运行老板登陆

**Goal:** 实现老板专属 H5 登录页 — 3种登录方式（手机号+验证码/密码/微信授权），OWNER+ADMIN 允许登录，MEMBER 拒绝，首次登录引导录入企业信息
**Requirements**: 手机号+验证码/密码/微信OAuth，角色过滤，Auth Guard，首次引导分流，品牌极简商务风
**Depends on:** v1.0 MVP
**Plans:** 2 plans
**Context:** `.planning/phases/01a-login-boss/01-CONTEXT.md` ✅ 已完成

Plans:
- [x] 01-01-PLAN.md — 前端登录页 + Auth Guard
- [x] 01-02-PLAN.md — 后端登录接口扩展（密码登录 + MEMBER 403 + /auth/me）

### Phase 02: 新增登陆界面，该登陆界面只允许老板账户登陆

**Status:** ⏸ Deferred — Phase 01 交付后重新定义范围（当前描述与 Phase 01 重复）
**Goal:** TBD
**Depends on:** Phase 01

Plans:
- [ ] TBD

### Phase 3: web登陆界面中添加注册按钮以及流程

**Goal:** [To be planned]
**Requirements**: TBD
**Depends on:** Phase 2
**Plans:** 0 plans

Plans:
- [ ] TBD (run /gsd-plan-phase 3 to break down)

---
_Last updated: 2026-04-11 — Phase 01 discussed, Phase 02 deferred pending Phase 01 delivery_
