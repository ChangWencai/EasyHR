# Phase 09: Code Review Report

**Reviewed:** 2026-04-19T00:00:00Z
**Depth:** quick
**Files Reviewed:** 20
**Status:** issues_found

## Summary

Reviewed 20 files across the todo center phase (backend Go services, frontend Vue/TypeScript, and server wiring). Found 2 critical issues and 5 warnings.

The most serious issue is a missing idempotency check in `CheckContractRenewalReminders` that can cause duplicate contract renewal todos every time the daily scheduler runs. The `/upload/image` endpoint has no auth guard. Several other issues include silent error swallowing in todo creation and the invite-fill route being protected by the frontend auth guard despite being designed as a public endpoint.

---

## Critical Issues

### CR-01: Missing idempotency check causes duplicate contract renewal todos

**File:** `internal/employee/contract_service.go:354-391`
**Issue:** `CheckContractRenewalReminders` creates a contract renewal todo on every invocation without checking whether one already exists. The `todo` service's `CreateTodo` method uses `ExistsBySource` for idempotency, but `CheckContractRenewalReminders` calls `CreateTodoFromEmployee` which bypasses that check entirely (it directly creates the item). This means every time the scheduler runs at 02:05 CST, a duplicate `contract_renew` todo will be created for each expiring contract, flooding the user's todo list.

**Fix:**
```go
// Before creating, check if a contract_renew todo already exists for this contract
exists, _ := s.todoSvc.ExistsBySource(ctx, contract.OrgID, "contract", &contractID)
if exists {
    continue
}
_ = s.todoSvc.CreateTodoFromEmployee(...)
```

---

### CR-02: `/upload/image` endpoint has no authentication

**File:** `internal/upload/router.go:11` and `cmd/server/main.go:291`
**Issue:** `RegisterRouter(v1.Group(""), "./uploads", "")` registers the upload endpoint at the root of the API group, outside any auth middleware. Additionally, the handler itself (`internal/upload/handler.go`) performs no session or token validation. Any unauthenticated user can upload arbitrary files up to 5MB, which could be abused for spam or storage exhaustion.

**Fix:**
Wrap the upload route with `authMiddleware`:
```go
// In router.go
func RegisterRouter(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, uploadDir, baseURL string) {
    handler := NewHandler(uploadDir, baseURL)
    authGroup := rg.Group("")
    authGroup.Use(authMiddleware)
    authGroup.POST("/upload/image", handler.UploadImage)
}
```
Then update `main.go:291` to pass the middleware:
```go
upload.RegisterRouter(v1.Group(""), authMiddleware, "./uploads", "")
```

---

## Warnings

### WR-01: Silent error swallowing in `CreateTodoFromEmployee`

**File:** `internal/employee/service.go:126` and `internal/employee/contract_service.go:377`
**Issue:** Both `CreateTodoFromEmployee` call sites use `_ = s.todoSvc.CreateTodoFromEmployee(...)`, silently discarding any error returned by the todo service. If todo creation fails, the caller has no way to know.

**Fix:** Log the error instead of silently discarding it:
```go
if err := s.todoSvc.CreateTodoFromEmployee(...); err != nil {
    logger.Logger.Warn("failed to create contract todo", zap.Error(err))
}
```

---

### WR-02: `SubmitInvite` discards submitted data

**File:** `internal/todo/service.go:282-284`
**Issue:** `SubmitInvite` accepts `name`, `phone`, and `remark` from the request body, marks the invite as used, but then discards all three fields (`_ = req.Name`, `_ = req.Phone`, `_ = req.Remark`). The协办人's submission data is permanently lost. The code comment acknowledges this: "当前仅记录协办人已提交，后续可在此扩展实际业务逻辑". Until this is implemented, users filling out the form see a success message but their data is not stored.

**Fix:** Persist the submission data or, if intentionally left as a placeholder, remove the fields from the request struct to avoid false expectations.

---

### WR-03: File extension check does not validate MIME type or magic bytes

**File:** `internal/upload/handler.go:44-48`
**Issue:** The handler only checks `filepath.Ext(file.Filename)` against an allowlist. It does not verify the file's actual content type (MIME) or magic bytes. An attacker can upload a file named `malicious.jpg` containing PHP code or other executable content. While `filepath.Ext` is case-insensitive on this platform, the lack of content validation is a defense-in-depth concern.

**Fix:** Add MIME type checking after reading the file header:
```go
// Read first 512 bytes for MIME detection
buffer := make([]byte, 512)
n, _ := src.Read(buffer)
mime := http.DetectContentType(buffer[:n])
allowedMimes := map[string]bool{"image/jpeg": true, "image/png": true, "image/gif": true, "image/webp": true}
if !allowedMimes[mime] {
    response.Error(c, http.StatusBadRequest, 40002, "文件类型不允许")
    return
}
```

---

### WR-04: Frontend `/todo/:id/invite` route is protected by auth guard

**File:** `frontend/src/router/index.ts:215-232`
**Issue:** The router guard logic treats `/todo/:id/invite` as a protected route because it `startsWith('/todo')`. The backend correctly exposes `VerifyInviteToken` and `SubmitInvite` as public endpoints (no auth middleware), but the frontend guard blocks unauthenticated users from ever reaching the `InviteFillPage.vue`. A user clicking an invite link while logged out will be redirected to `/login` instead of seeing the invite page.

**Fix:** Add `/todo` invite route exclusion in the guard, before the protected route check:
```ts
// Line 215: extend the public path exclusion
if (to.path === '/login' ||
    to.path.startsWith('/todo/') && to.path.includes('/invite') ||  // invite pages are public
    to.path === '/onboarding/org-setup' ||
    to.path.startsWith('/register') ||
    to.path.startsWith('/salary/slip/')) {
  return
}
```

---

### WR-05: `CarouselItem` and `TodoItem` types duplicated across API files

**File:** `frontend/src/api/carousel.ts:1-11` and `frontend/src/api/todo.ts:25-34`
**Issue:** The `CarouselItem` interface is defined in both `carousel.ts` (lines 3-11) and `todo.ts` (lines 25-34). `TodoItem` is defined in `todo.ts` but a separate variant also appears in `carousel.ts`. This duplication creates a maintenance risk: if the backend API response shape changes, both copies must be updated in sync.

**Fix:** Consolidate shared types into a single file such as `frontend/src/types/todo.ts` and import from there:
```ts
// frontend/src/types/todo.ts
export interface CarouselItem { ... }
export interface TodoItem { ... }
```

---

## Info

### IN-01: Duplicate `CarouselItem` type across API modules

**File:** `frontend/src/api/carousel.ts:3-11` and `frontend/src/api/todo.ts:25-34`
**Issue:** Same as WR-05 above. Consolidate into a shared types file.

### IN-02: `TodoItem` interface field `org_id` present in backend model but absent in frontend type

**File:** `frontend/src/api/todo.ts:3-23`
**Issue:** The backend `TodoItem` model (internal/todo/model.go) includes `OrgID` (via `model.BaseModel`), but the frontend `TodoItem` interface does not include an `org_id` field. This is acceptable if the frontend never renders org info, but if it is needed, the field should be added.

### IN-03: `todoRepoForDI` instantiated twice in `main.go`

**File:** `cmd/server/main.go:148-149` and `cmd/server/main.go:325`
**Issue:** `todo.NewRepository(db)` is called twice: once for employee module DI (line 148) and once for the scheduler (line 325). This is not a bug, but it means two separate repository instances point to the same DB. If they diverge (e.g., one gets a transaction), their state may be inconsistent. Low risk in current code but worth noting.

### IN-04: `UpdateCarousel` silently ignores missing fields

**File:** `internal/todo/service.go:330-348`
**Issue:** `UpdateCarousel` always sets `image_url` and `active` in the updates map, even when the request body does not include them (they are required/present by the `CarouselRequest` struct definition). This means `image_url` is always overwritten, even with the same value. No functional bug, but a consistency issue with the create path which checks for required fields.

---

_Reviewed: 2026-04-19_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: quick_
