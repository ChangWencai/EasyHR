---
status: resolved
trigger: "Mine page: handleEditOrg and handleChangePassword only have empty ElMessage.info stubs"
created: 2026-04-11T00:00:00+08:00
updated: 2026-04-11T00:10:00+08:00
---

## Current Focus
next_action: "Done"

## Symptoms
expected: "点击'编辑企业信息'能编辑企业信息，点击'修改密码'能修改密码"
actual: "handleEditOrg 和 handleChangePassword 只有 ElMessage.info 空实现"
reproduction: "打开'我的'页面，点击编辑或修改密码按钮"
started: "从未实现"

## Eliminated

## Evidence
- timestamp: 2026-04-11
  checked: internal/user/handler.go
  found: "No PUT /auth/password or PUT /org endpoint exists"
  implication: "Need to add both endpoints"

- timestamp: 2026-04-11
  checked: internal/user/repository.go
  found: "UpdateOrg method exists (line 53-55) for partial org updates"
  implication: "Can use this for org update without modifying CompleteOnboarding"

- timestamp: 2026-04-11
  checked: OrgSetup.vue + CompleteOnboardingRequest DTO
  found: "Frontend sends snake_case (contact_phone) but Go struct has ContactPhone field with json:\"ContactPhone\" tag — mismatch causes phone to not bind"
  implication: "Need to fix DTO json tags AND frontend field names"

## Resolution
root_cause: "Two features never implemented: no backend password-change API, no backend org-update API, no frontend dialogs"
fix: |
  1. Added ChangePasswordRequest + UpdateOrgRequest DTOs (dto.go)
  2. Added ChangePassword + UpdateOrg service methods (service.go)
  3. Added PUT /auth/password + PUT /org handler endpoints (handler.go)
  4. Fixed CompleteOnboardingRequest json tags (snake_case compatible)
  5. Full el-dialog implementations in MineView.vue for both features
verification: "go build ./... passes, npx tsc --noEmit passes, go test ./internal/common/crypto/... passes"
files_changed:
  - internal/user/dto.go
  - internal/user/service.go
  - internal/user/handler.go
  - frontend/src/views/mine/MineView.vue
