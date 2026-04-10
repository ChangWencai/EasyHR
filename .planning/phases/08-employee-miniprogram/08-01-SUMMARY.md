# 08-01-SUMMARY: Go Backend — Employee WeChat Mini Program API

**Plan:** 08-01 | **Phase:** 08 | **Status:** ✓ Complete

## What was built

WXMP backend package at `internal/wxmp/` providing all employee-facing API endpoints.

**Packages:**
- `model.go` — DTOs: LoginRequest/Response, PayslipSummary, PayslipDetail, ContractDTO, SocialInsuranceDTO, ExpenseRequest/Response, OssUploadResponse
- `repository.go` — WXMPRepository, queries Phase 2/3/5/6 data filtered by user_id+org_id
- `service.go` — WXMPService, coordinates all operations, SMS verification logic
- `middleware.go` — WXMPMemberAuth, JWT + MEMBER role enforcement
- `handler.go` — 12 HTTP handlers for all WXMP endpoints
- `router.go` — Route registration at `/api/v1/wxmp/*`
- `*_test.go` — Test scaffold with Redis skip logic

**API Endpoints:**
- `POST /api/v1/wxmp/auth/login` — 手机号+验证码登录
- `POST /api/v1/wxmp/auth/wechat/bind` — 绑定微信openid
- `GET /api/v1/wxmp/payslips` — 月度工资单列表
- `POST /api/v1/wxmp/payslips/:id/verify` — 短信验证工资条明细
- `GET /api/v1/wxmp/payslips/:id` — 工资条明细（含明细，需验证）
- `POST /api/v1/wxmp/payslips/:id/sign` — 确认签收
- `GET /api/v1/wxmp/contracts` — 合同列表
- `GET /api/v1/wxmp/contracts/:id/pdf` — 合同PDF URL
- `GET /api/v1/wxmp/social-insurance` — 社保缴费记录（仅个人缴费）
- `POST /api/v1/wxmp/expenses` — 提交费用报销
- `GET /api/v1/wxmp/expenses` — 报销列表
- `GET /api/v1/wxmp/expenses/:id` — 报销详情
- `GET /api/v1/wxmp/oss/upload-url` — OSS预签名上传URL

**Integration:** Registered in `cmd/server/main.go`

## Verification

- `go build ./cmd/server/...` ✓
- `go vet ./internal/wxmp/...` ✓

## Key Decisions

- Employee data filtered by JWT user_id, not org_id alone (employee→user mapping)
- Social insurance: only personal contributions shown (company amount hidden)
- SMS verification: 5-minute code stored in Redis, verified before payslip detail access
- All routes require MEMBER token except login

## Deviation from Plan

- Tests use Redis with skip-if-unavailable pattern instead of miniredis mock
- WXMPRepository adapted to actual model fields (assumed field names didn't match)
