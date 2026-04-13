---
phase: 03-social-insurance
plan: 01
subsystem: social-insurance
tags: [jsonb, gorm, datatypes, social-insurance, policy, calculation-engine]

# Dependency graph
requires:
  - phase: 01-foundation-auth
    provides: BaseModel, RBAC middleware, TenantScope, response helpers
  - phase: 02-employee-management
    provides: 三层架构模式, city module (37 cities), Repository pattern

provides:
  - SocialInsurancePolicy 数据模型（JSONB存储五险一金配置）
  - Repository CRUD + FindByCityAndYear（取 effective_year<=year 最新记录）
  - CalculateInsuranceAmounts 基数计算引擎（clamp到上下限，6险种金额计算）
  - 6个 HTTP 端点（政策 CRUD + 计算接口）
  - 预留 GetSocialInsuranceDeduction 接口（Plan 02 实现）

affects: [03-02, 03-03, 05-salary]

# Tech tracking
tech-stack:
  added: [gocron/v2 v2.19.1, gocron-redis-lock/v2 v2.2.1]
  patterns: [datatypes.JSONType[T] 强类型JSONB封装, 全局数据OrgID=0不使用TenantScope, clamp基数计算函数]

key-files:
  created:
    - internal/socialinsurance/model.go
    - internal/socialinsurance/dto.go
    - internal/socialinsurance/repository.go
    - internal/socialinsurance/service.go
    - internal/socialinsurance/handler.go
    - internal/socialinsurance/service_test.go
    - internal/socialinsurance/repository_test.go
  modified:
    - cmd/server/main.go
    - go.mod
    - go.sum

key-decisions:
  - "政策库为全局共享数据（OrgID=0），不使用 TenantScope，所有企业共用同一套政策"
  - "使用 datatypes.JSONType[FiveInsurances] 强类型封装 JSONB，避免手动 json.Marshal"
  - "基数计算使用 clamp(salary, baseLower, baseUpper) 函数，各险种独立 clamp"
  - "金额四舍五入到分（乘100后round再除100），避免浮点精度问题"
  - "GetSocialInsuranceDeduction 预留空方法签名，Plan 02 实现"

patterns-established:
  - "全局数据模式：OrgID=0 + 查询条件显式 AND org_id = 0"
  - "JSONB强类型：datatypes.JSONType[T] + newJSONType辅助函数"
  - "金额精度：math.Round(x*100)/100 四舍五入到分"

requirements-completed: [SOCL-01]

# Metrics
duration: 23min
completed: 2026-04-07
---

# Phase 03 Plan 01: 社保政策库+基数计算引擎 Summary

**JSONB强类型社保政策库（datatypes.JSONType[FiveInsurances]）+ 基数clamp计算引擎 + 6个HTTP端点**

## Performance

- **Duration:** 23 min
- **Started:** 2026-04-07T06:47:55Z
- **Completed:** 2026-04-07T07:11:32Z
- **Tasks:** 2
- **Files modified:** 9

## Accomplishments
- 社保政策库数据模型：SocialInsurancePolicy 使用 JSONB 存储五险一金配置（6个险种，每个包含企业/个人比例、基数上下限）
- 基数计算引擎：CalculateInsuranceAmounts 根据城市+薪资自动匹配政策，clamp基数到上下限，计算6个险种的企业和个人缴费金额
- Repository 层：CRUD + FindByCityAndYear 取 effective_year<=year 的最新政策记录
- Handler 层：6个 HTTP 端点，OWNER 权限控制政策增删改，所有角色可查询和计算
- 工伤/生育保险个人缴费比例为0已验证
- 10个单元测试全部 PASS，全项目测试无回归

## Task Commits

Each task was committed atomically:

1. **Task 1: 社保政策库模型 + Repository + 基数计算 Service** - `6e932fc` (feat)
2. **Task 2: 政策管理 Handler + 路由注册 + main.go 集成** - `d9425a1` (feat)

## Files Created/Modified
- `internal/socialinsurance/model.go` - SocialInsurancePolicy 模型，InsuranceItem/FiveInsurances 结构体
- `internal/socialinsurance/dto.go` - 请求/响应 DTO（CreatePolicyRequest, CalculateRequest, CalculateResponse 等）
- `internal/socialinsurance/repository.go` - Repository CRUD + FindByCityAndYear（全局数据 OrgID=0）
- `internal/socialinsurance/service.go` - Service 业务逻辑，CalculateInsuranceAmounts 计算引擎
- `internal/socialinsurance/handler.go` - 6个 HTTP 端点，RBAC 权限控制
- `internal/socialinsurance/service_test.go` - Service 单元测试（10个测试用例）
- `internal/socialinsurance/repository_test.go` - Repository 单元测试（CRUD/分页/软删除）
- `cmd/server/main.go` - AutoMigrate 添加 SocialInsurancePolicy，依赖注入和路由注册
- `go.mod` / `go.sum` - 新增 gocron v2.19.1 依赖

## Decisions Made
- 政策库为全局共享数据（OrgID=0），不使用 TenantScope，所有企业共用同一套政策数据
- 使用 datatypes.JSONType[T] 强类型封装 JSONB，避免手动序列化/反序列化
- 金额四舍五入到分（math.Round(x*100)/100），避免浮点精度累积误差
- GetSocialInsuranceDeduction 预留空方法签名（D-12），Plan 02 实现员工社保扣款查询
- 测试中总计期望值从 3850/2205 修正为 3930/2250（原始计算遗漏了失业保险50/50和工伤保险20）

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] 测试期望值计算错误**
- **Found during:** Task 1 (TDD GREEN 阶段)
- **Issue:** 测试中 TotalCompany/TotalPersonal 期望值 3850/2205 未包含失业保险(50+50)和工伤保险(20)的全部金额
- **Fix:** 重新计算所有6个险种的金额：养老(1600+800) + 医疗(980+200) + 失业(50+50) + 工伤(20+0) + 生育(80+0) + 公积金(1200+1200) = 企业3930/个人2250
- **Files modified:** internal/socialinsurance/service_test.go
- **Verification:** go test ./internal/socialinsurance/... -count=1 -v 全部 PASS
- **Committed in:** 6e932fc (Task 1 commit)

**2. [Rule 3 - Blocking] repository_test.go 未使用的 import**
- **Found during:** Task 1 (首次测试运行)
- **Issue:** gorm.io/gorm 被 import 但未使用导致编译失败
- **Fix:** 删除未使用的 import
- **Files modified:** internal/socialinsurance/repository_test.go
- **Verification:** go test 编译通过
- **Committed in:** 6e932fc (Task 1 commit)

---

**Total deviations:** 2 auto-fixed (1 bug, 1 blocking)
**Impact on plan:** 均为测试代码修正，无功能影响。计划执行完全符合预期。

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- 社保政策库 CRUD API 和基数计算引擎已完成，Plan 02 可在此基础上实现参保/停缴操作
- GetSocialInsuranceDeduction 接口已预留，Plan 02 需要实现员工社保扣款查询
- gocron v2.19.1 已安装，Plan 03 缴费提醒定时任务可直接使用
- handler.go 中 RegisterRoutes 模式已建立，后续 Plan 的端点可追加注册

---
*Phase: 03-social-insurance*
*Completed: 2026-04-07*
