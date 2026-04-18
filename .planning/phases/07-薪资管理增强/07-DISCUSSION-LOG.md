# Phase 7: 薪资管理增强 - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-18
**Phase:** 07-薪资管理增强
**Areas discussed:** 薪资数据看板, 调薪/普调, 个税上传, 考勤联动, 工资条发送, 历史数据保护

---

## Area 1: 薪资数据看板

| Option | Description | Selected |
|--------|-------------|----------|
| 4张纯数字卡片（推荐） | 4张纯数字卡片（应发总额/实发总额/社保公积金/个税总额），纯数字卡片风格 | ✓ |
| 6张细分卡片 | 应发/实发分开，社保再分单位+个人，指标更细致但密度更高 | |
| 你来决定 | 自由描述 | |

**User's choice:** 4张纯数字卡片（推荐）
**Notes:** 与员工数据看板、考勤月报风格一致，纯数字卡片交付最快

| Option | Description | Selected |
|--------|-------------|----------|
| 仅已确认月份（推荐） | 当月已确认/已支付（confirmed/paid）的 PayrollRecord 聚合数据；下月无数据时显示"—" | ✓ |
| 包含草稿月份 | 即使当月工资表未确认，也展示草稿数据，标注状态 | |

**User's choice:** 仅已确认月份（推荐）
**Notes:** —

---

## Area 2: 调薪/普调

| Option | Description | Selected |
|--------|-------------|----------|
| INSERT ONLY + 月份只读（推荐） | 新调薪记录 INSERT 新行（不 UPDATE 历史）；当月 draft 工资表自动重新核算；confirmed/paid 月份的 PayrollRecord 保持不变 | ✓ |
| 其他 | 自由描述 | |

**User's choice:** INSERT ONLY + 月份只读（推荐）
**Notes:** effective_month 自然控制生效月份

| Option | Description | Selected |
|--------|-------------|----------|
| 独立月度表（推荐） | performance_coefficients 表（employee_id + year_month + coefficient），每月一条记录，默认值 1.0 | ✓ |
| 纯计算字段 | 无需新增字段，在 CalculatePayroll 时动态读取系数与绩效工资项相乘 | |

**User's choice:** 独立月度表（推荐）
**Notes:** —

---

## Area 3: 个税上传

| Option | Description | Selected |
|--------|-------------|----------|
| 姓名精确匹配为主（推荐） | 员工姓名列直接匹配（精确匹配 > 模糊匹配 > 跳过）；跳过无法匹配的行并提示"3行无法匹配" | ✓ |
| 多字段备选匹配 | 支持按手机号或身份证匹配；匹配方式可配置 | |

**User's choice:** 姓名精确匹配为主（推荐）
**Notes:** —

| Option | Description | Selected |
|--------|-------------|----------|
| 部分成功提示（推荐） | 匹配失败的行记录在错误日志中；全部失败则整体失败；部分匹配成功则成功并列出未匹配行 | ✓ |
| 严格校验，全部失败 | 上传失败并显示具体失败原因，需修正后重新上传 | |

**User's choice:** 部分成功提示（推荐）
**Notes:** —

---

## Area 4: 考勤联动

| Option | Description | Selected |
|--------|-------------|----------|
| 按应出勤比例计算（推荐） | 基本工资项 × (实际出勤+法定节假日+带薪假)/应出勤 × 法定系数；SAL-13 的精确实现 | ✓ |
| 其他 | 自由描述 | |

**User's choice:** 按应出勤比例计算（推荐）
**Notes:** 计薪天数 = 实际出勤 + 法定节假日 + 带薪假天数

| Option | Description | Selected |
|--------|-------------|----------|
| sick_leave_policies 表（推荐） | sick_leave_policies 表：工龄档位 × 城市 × 系数；初期只北上广深；参照当地最低工资80%校验 | ✓ |
| 硬编码固定系数 | 前期用固定系数：6个月医疗期内=60%/超6个月=40%，不按城市区分 | |

**User's choice:** sick_leave_policies 表（推荐）
**Notes:** —

---

## Area 5: 工资条发送

| Option | Description | Selected |
|--------|-------------|----------|
| 微信小程序优先（推荐） | 优先微信小程序通知（wx.request）→ 小程序未绑定则发短信（阿里云 SMS）→ 短信发送失败则降级 H5 链接 | ✓ |
| 仅 H5 链接 | 只发 H5 工资条链接（PayrollSlip.Token）；最简单实现 | |

**User's choice:** 微信小程序优先（推荐）
**Notes:** —

---

## Area 6: 历史数据保护

| Option | Description | Selected |
|--------|-------------|----------|
| confirmed/paid 需解锁码（推荐） | draft/calculated 可重新编辑和重算；confirmed/paid 禁止任何修改，需管理员输入解锁码（如：企业主手机验证码）才能解锁后重新编辑 | ✓ |
| confirmed/paid 永久只读 | draft/calculated 可编辑；confirmed/paid 只能查看，解锁需要联系客服 | |
| 其他 | 自由描述 | |

**User's choice:** confirmed/paid 需解锁码（推荐）
**Notes:** —

---

## Claude's Discretion

- 薪资数据看板卡片的排列顺序和具体样式细节
- 普调按部门选择的具体 UI（多选部门？全选按钮？）
- 个税 Excel 列名识别算法（支持哪些别名映射）
- 计薪天数中"法定节假日"的来源（AttendanceRule.Holidays JSON 还是独立 holidays 表）
- 工资条 H5 页面具体样式和内容结构
- sick_leave_policies 表的初始数据（北上广深各档位系数）

## Deferred Ideas

None — discussion stayed within phase scope
