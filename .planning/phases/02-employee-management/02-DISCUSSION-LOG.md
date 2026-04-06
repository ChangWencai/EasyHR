# Phase 2: 员工管理 - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-06
**Phase:** 02-employee-management
**Mode:** auto (assumptions mode with auto-selected defaults)

---

## 入职邀请机制

| Option | Description | Selected |
|--------|-------------|----------|
| H5页面填写，无需下载APP | 邀请链接打开H5网页，员工直接填写信息，门槛最低 | ✓ |
| APP内填写 | 员工需下载APP后填写，体验好但流失率高 | |
| 微信小程序填写 | 员工通过微信扫码填写，依赖小程序已上线 | |

**Auto-selected:** H5页面填写，无需下载APP
**Reason:** 目标用户是小微企业员工，降低填写门槛最重要。H5页面无需安装任何应用，扫码即可填写。

---

## 员工数据模型

| Option | Description | Selected |
|--------|-------------|----------|
| Employee独立于User，可选关联 | 不是所有员工都需要登录，user_id可为空 | ✓ |
| Employee继承User | 员工必须是系统用户，强制关联 | |
| 完全分离，无关联 | 员工和用户系统完全独立 | |

**Auto-selected:** Employee独立于User，可选关联
**Reason:** 老板手动录入的员工不需要登录，邀请入职的员工可能后续开通登录。灵活关联最符合实际场景。

---

## 离职流程

| Option | Description | Selected |
|--------|-------------|----------|
| 双方均可发起 | 老板直接办理或员工申请审批，覆盖所有场景 | ✓ |
| 仅老板发起 | 老板单向操作，流程简单 | |
| 仅员工申请 | 员工发起，老板审批，但小企业可能不需要 | |

**Auto-selected:** 双方均可发起
**Reason:** 小微企业场景中，老板直接办理离职很常见（试用期不合格等），同时也有员工主动辞职的情况。

---

## 合同管理方式

| Option | Description | Selected |
|--------|-------------|----------|
| PDF模板生成+线下签署+上传扫描件 | V1.0降级方案，无需对接电子签平台 | ✓ |
| 纯文本合同，无PDF | 简单但不符合法律规范 | |
| 对接电子签API | V2.0才实现，V1.0成本过高 | |

**Auto-selected:** PDF模板生成+线下签署+上传扫描件
**Reason:** V1.0降级方案已明确在需求中（EMPL-08）。PDF模板保证格式规范，线下签署后上传满足合规要求。

---

## Claude's Discretion

- 员工模块内部目录结构
- 邀请 H5 页面是否复用现有框架
- 交接清单编辑交互细节
- 搜索性能优化策略
- 合同 PDF 模板具体排版
- 员工头像上传（V1.0 可选）

## Deferred Ideas

None — analysis stayed within phase scope
