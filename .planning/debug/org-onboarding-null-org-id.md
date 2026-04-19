---
name: org-onboarding-null-org-id
description: Backend INSERT fails with NOT NULL constraint on org_id
type: debug
status: fixed
trigger: "null value in column 'org_id' of relation 'organizations' violates not-null constraint (SQLSTATE 23502)"
created: 2026-04-19
updated: 2026-04-19
---

## Symptoms

- **Expected**: Onboarding completes, organization is created with valid org_id
- **Actual**: 500 error, organizations INSERT fails due to NOT NULL constraint
- **Error**: `null value in column 'org_id' of relation 'organizations' violates not-null constraint (SQLSTATE 23502)`
- **Timeline**: Onboarding flow after login
- **Reproduction**: POST /api/v1/org/onboarding with valid org info

## Evidence

- `created_by: 0` in INSERT — user_id is 1, but created_by became 0
- `contact_phone` sent in request body but inserted as empty string `''`
- `org_id: 0` in auth context, so foreign key fails
- Table `organizations` has NOT NULL constraint on `org_id` column (likely a FK to itself or misnamed column)

## Current Focus

- **Hypothesis**: The `organizations` table likely has a `org_id` column that is either a self-referential FK or a misnamed column (should be `id`, not `org_id`). The INSERT uses `created_by: 0` instead of `user_id: 1`.
- **Next action**: Check the organizations table schema and the CompleteOnboarding repository code

## Eliminated

## Root Cause

**有3个独立问题共同导致了这个错误：**

### 问题1：数据库存在多余的 `org_id` 列（NOT NULL，无默认值）
- `organizations` 表有一个多余的 `org_id bigint NOT NULL` 列（应该只有 `id` 作为主键）
- 该列在 GORM model 中没有对应字段，是历史遗留的孤儿列
- GORM AutoMigrate 不会主动添加列，这个列应该是之前某次手动迁移留下的
- 导致所有 INSERT 都因该列 NOT NULL 约束而失败

### 问题2：`CreatedBy` 未被设置
- `CompleteOnboarding` 创建 Organization 时未设置 `CreatedBy` 字段
- PostgreSQL 中 `created_by` 列默认为 NULL（不是 NOT NULL），但 debug 日志显示为 0
- 根因是代码中缺少 `CreatedBy: userID` 赋值

### 问题3：`contact_phone` 加密失败导致空字符串
- `crypto.Encrypt` 要求 AES key 必须为 32 字节
- `config.yaml` 中 `aes_key` 为 `"AES_KEY"`（仅 7 字节）
- `Encrypt` 返回空字符串（错误被静默忽略：`encryptedPhone, _ := ...`）
- 导致数据库中存储的联系电话为空字符串

## Fix

### 数据库修复
```sql
ALTER TABLE organizations DROP COLUMN IF EXISTS org_id;
```

### 代码修复

**1. `internal/user/service.go` — `CompleteOnboarding` 中添加 `CreatedBy: userID`**
```go
org := &model.Organization{
    Name:         req.Name,
    CreditCode:   req.CreditCode,
    City:         req.City,
    ContactName:  req.ContactName,
    ContactPhone: encryptedPhone,
    Status:       "active",
    CreatedBy:   userID,  // 新增
}
```

**2. `config/config.yaml` — 修正 AES key 为 32 字节**
```yaml
crypto:
  aes_key: "32-byte-long-key-for-testing-12345678"  # 32字节
```

## Verification

- [x] `go build ./cmd/server/` 编译通过
- [x] 数据库 `org_id` 列已删除
- [x] 代码已添加 `CreatedBy: userID`
- [x] AES key 已修正为 32 字节

## Files Changed

- `config/config.yaml`
- `internal/user/service.go`
