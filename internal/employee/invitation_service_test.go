package employee

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wencai/easyhr/internal/common/config"
	"github.com/wencai/easyhr/internal/common/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupInvitationTestDB 初始化测试数据库（含 Invitation 和 Organization 表）
func setupInvitationTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Discard,
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Employee{}, &Invitation{}, &model.Organization{}))
	return db
}

// newTestInvitationService 创建测试用的 InvitationService
func newTestInvitationService(db *gorm.DB) *InvitationService {
	invRepo := NewInvitationRepository(db)
	empRepo := NewRepository(db)
	cfg := config.CryptoConfig{AESKey: testAESKey}
	return NewInvitationService(invRepo, empRepo, cfg)
}

// createInvTestOrg 创建测试组织（用于邀请测试，避免与 repository_test.go 中的 createTestOrg 冲突）
func createInvTestOrg(db *gorm.DB, id int64, name string) {
	org := model.Organization{}
	org.ID = id
	org.Name = name
	org.CreditCode = fmt.Sprintf("test_credit_%d", id)
	org.City = "北京"
	org.Status = "active"
	db.Create(&org)
}

// TestCreateInvitation_TokenAndExpiry 验证 token 长度 64 字符，expires_at = now+7天
func TestCreateInvitation_TokenAndExpiry(t *testing.T) {
	db := setupInvitationTestDB(t)
	svc := newTestInvitationService(db)

	beforeCreate := time.Now()
	resp, err := svc.CreateInvitation(1, 1, &CreateInvitationRequest{Position: "前端开发"})
	require.NoError(t, err)
	require.NotNil(t, resp)

	// 验证 token 长度为 64 字符（32 字节 hex 编码）
	assert.Len(t, resp.Token, 64, "token 应为 64 字符 hex 字符串")

	// 验证 InviteURL 格式
	assert.Contains(t, resp.InviteURL, "/invite/")
	assert.Contains(t, resp.InviteURL, resp.Token)

	// 验证过期时间约为 7 天后
	expiresAt, err := time.Parse(time.RFC3339, resp.ExpiresAt)
	require.NoError(t, err)
	expectedExpiry := beforeCreate.Add(InvitationExpiryDuration)
	diff := expiresAt.Sub(expectedExpiry)
	assert.True(t, diff < 5*time.Second && diff > -5*time.Second, "过期时间应为创建时间+7天")

	// 验证数据库记录
	invRepo := NewInvitationRepository(db)
	inv, err := invRepo.FindByToken(resp.Token)
	require.NoError(t, err)
	assert.Equal(t, InvitationStatusPending, inv.Status)
	assert.Equal(t, int64(1), inv.OrgID)
	assert.Equal(t, "前端开发", inv.Position)
}

// TestGetInvitationDetail_Success 返回 org_name 和预设岗位
func TestGetInvitationDetail_Success(t *testing.T) {
	db := setupInvitationTestDB(t)
	svc := newTestInvitationService(db)

	// 创建测试组织
	createInvTestOrg(db, 1, "测试科技有限公司")

	// 创建邀请
	resp, err := svc.CreateInvitation(1, 1, &CreateInvitationRequest{Position: "后端开发"})
	require.NoError(t, err)

	// 获取邀请详情
	detail, err := svc.GetInvitationDetail(resp.Token)
	require.NoError(t, err)
	require.NotNil(t, detail)

	assert.Equal(t, "测试科技有限公司", detail.OrgName)
	assert.Equal(t, "后端开发", detail.Position)
	assert.Equal(t, InvitationStatusPending, detail.Status)
	assert.NotEmpty(t, detail.ExpiresAt)
}

// TestSubmitInvitation_CreatesEmployee Employee status=pending，邀请 status=used
func TestSubmitInvitation_CreatesEmployee(t *testing.T) {
	db := setupInvitationTestDB(t)
	svc := newTestInvitationService(db)

	// 创建测试组织
	createInvTestOrg(db, 1, "测试公司")

	// 创建邀请
	resp, err := svc.CreateInvitation(1, 1, &CreateInvitationRequest{Position: "产品经理"})
	require.NoError(t, err)

	// 员工提交信息
	submitReq := &SubmitInvitationRequest{
		Name:     "张三",
		Phone:    "13800008000",
		IDCard:   "110108199001011234",
		Position: "产品经理",
		HireDate: "2026-01-15",
	}
	err = svc.SubmitInvitation(resp.Token, submitReq)
	require.NoError(t, err)

	// 验证邀请状态已变为 used
	invRepo := NewInvitationRepository(db)
	inv, err := invRepo.FindByToken(resp.Token)
	require.NoError(t, err)
	assert.Equal(t, InvitationStatusUsed, inv.Status)
	require.NotNil(t, inv.EmployeeID)
	require.NotNil(t, inv.UsedAt)

	// 验证 Employee 记录已创建，status=pending
	empRepo := NewRepository(db)
	emp, err := empRepo.FindByID(1, *inv.EmployeeID)
	require.NoError(t, err)
	assert.Equal(t, "张三", emp.Name)
	assert.Equal(t, StatusPending, emp.Status)
	assert.Equal(t, "产品经理", emp.Position)
	assert.Equal(t, "男", emp.Gender) // 从身份证号提取
}

// TestSubmitInvitation_AlreadyUsed 已使用 token 返回错误
func TestSubmitInvitation_AlreadyUsed(t *testing.T) {
	db := setupInvitationTestDB(t)
	svc := newTestInvitationService(db)

	// 创建测试组织
	createInvTestOrg(db, 1, "测试公司")

	// 创建邀请并提交
	resp, err := svc.CreateInvitation(1, 1, &CreateInvitationRequest{Position: "开发"})
	require.NoError(t, err)

	submitReq := &SubmitInvitationRequest{
		Name:     "张三",
		Phone:    "13800008000",
		IDCard:   "110108199001011234",
		Position: "开发",
		HireDate: "2026-01-15",
	}
	err = svc.SubmitInvitation(resp.Token, submitReq)
	require.NoError(t, err)

	// 再次提交同一 token
	submitReq2 := &SubmitInvitationRequest{
		Name:     "李四",
		Phone:    "13900009000",
		IDCard:   "110108199503152345",
		Position: "开发",
		HireDate: "2026-02-01",
	}
	err = svc.SubmitInvitation(resp.Token, submitReq2)
	assert.Error(t, err)
	assert.Equal(t, ErrInvitationUsed, err)
}

// TestSubmitInvitation_Expired 过期 token 返回错误
func TestSubmitInvitation_Expired(t *testing.T) {
	db := setupInvitationTestDB(t)
	invRepo := NewInvitationRepository(db)
	empRepo := NewRepository(db)
	cfg := config.CryptoConfig{AESKey: testAESKey}
	svc := NewInvitationService(invRepo, empRepo, cfg)

	// 手动创建已过期的邀请
	pastTime := time.Now().Add(-24 * time.Hour) // 昨天过期
	inv := &Invitation{
		OrgID:     1,
		Token:     "expired_token_for_test_1234567890abcdef1234567890abcdef12345678",
		Position:  "开发",
		Status:    InvitationStatusPending,
		CreatedBy: 1,
		ExpiresAt: pastTime,
	}
	require.NoError(t, invRepo.Create(inv))

	// 尝试提交
	submitReq := &SubmitInvitationRequest{
		Name:     "张三",
		Phone:    "13800008000",
		IDCard:   "110108199001011234",
		Position: "开发",
		HireDate: "2026-01-15",
	}
	err := svc.SubmitInvitation(inv.Token, submitReq)
	assert.Error(t, err)
	assert.Equal(t, ErrInvitationExpired, err)
}

// TestConfirmOnboarding_PendingToActive status pending -> active
func TestConfirmOnboarding_PendingToActive(t *testing.T) {
	db := setupInvitationTestDB(t)
	svc := newTestInvitationService(db)

	// 创建测试组织
	createInvTestOrg(db, 1, "测试公司")

	// 通过邀请创建员工（status=pending）
	resp, err := svc.CreateInvitation(1, 1, &CreateInvitationRequest{Position: "开发"})
	require.NoError(t, err)

	submitReq := &SubmitInvitationRequest{
		Name:     "张三",
		Phone:    "13800008000",
		IDCard:   "110108199001011234",
		Position: "开发",
		HireDate: "2026-01-15",
	}
	err = svc.SubmitInvitation(resp.Token, submitReq)
	require.NoError(t, err)

	// 获取员工 ID
	invRepo := NewInvitationRepository(db)
	inv, err := invRepo.FindByToken(resp.Token)
	require.NoError(t, err)
	require.NotNil(t, inv.EmployeeID)

	// 确认入职
	err = svc.ConfirmOnboarding(1, *inv.EmployeeID)
	require.NoError(t, err)

	// 验证员工状态已变为 active
	empRepo := NewRepository(db)
	emp, err := empRepo.FindByID(1, *inv.EmployeeID)
	require.NoError(t, err)
	assert.Equal(t, StatusActive, emp.Status)
}

// TestCancelInvitation pending -> cancelled
func TestCancelInvitation(t *testing.T) {
	db := setupInvitationTestDB(t)
	svc := newTestInvitationService(db)

	// 创建邀请
	resp, err := svc.CreateInvitation(1, 1, &CreateInvitationRequest{Position: "测试"})
	require.NoError(t, err)

	// 获取邀请 ID
	invRepo := NewInvitationRepository(db)
	inv, err := invRepo.FindByToken(resp.Token)
	require.NoError(t, err)

	// 取消邀请
	err = svc.CancelInvitation(1, inv.ID)
	require.NoError(t, err)

	// 验证状态
	inv, err = invRepo.FindByToken(resp.Token)
	require.NoError(t, err)
	assert.Equal(t, InvitationStatusCancelled, inv.Status)
}

// TestCancelInvitation_NotPending 非 pending 状态不可取消
func TestCancelInvitation_NotPending(t *testing.T) {
	db := setupInvitationTestDB(t)
	svc := newTestInvitationService(db)

	// 创建测试组织
	createInvTestOrg(db, 1, "测试公司")

	// 创建邀请并使用
	resp, err := svc.CreateInvitation(1, 1, &CreateInvitationRequest{Position: "开发"})
	require.NoError(t, err)

	submitReq := &SubmitInvitationRequest{
		Name:     "张三",
		Phone:    "13800008000",
		IDCard:   "110108199001011234",
		Position: "开发",
		HireDate: "2026-01-15",
	}
	err = svc.SubmitInvitation(resp.Token, submitReq)
	require.NoError(t, err)

	// 获取邀请 ID
	invRepo := NewInvitationRepository(db)
	inv, err := invRepo.FindByToken(resp.Token)
	require.NoError(t, err)

	// 尝试取消已使用的邀请
	err = svc.CancelInvitation(1, inv.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "仅待使用的邀请可以取消")
}

// TestGenerateToken_Uniqueness 多次生成 token 应各不相同
func TestGenerateToken_Uniqueness(t *testing.T) {
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, err := generateToken()
		require.NoError(t, err)
		assert.Len(t, token, 64)
		assert.False(t, tokens[token], "token 不应重复")
		tokens[token] = true
	}
}

// TestGetInvitationDetail_Cancelled 已取消的邀请返回错误
func TestGetInvitationDetail_Cancelled(t *testing.T) {
	db := setupInvitationTestDB(t)
	svc := newTestInvitationService(db)

	// 创建邀请
	resp, err := svc.CreateInvitation(1, 1, &CreateInvitationRequest{Position: "测试"})
	require.NoError(t, err)

	// 获取邀请并取消
	invRepo := NewInvitationRepository(db)
	inv, err := invRepo.FindByToken(resp.Token)
	require.NoError(t, err)
	err = svc.CancelInvitation(1, inv.ID)
	require.NoError(t, err)

	// 尝试获取详情
	_, err = svc.GetInvitationDetail(resp.Token)
	assert.Error(t, err)
	assert.Equal(t, ErrInvitationCancelled, err)
}

// TestListInvitations 分页和状态过滤
func TestListInvitations(t *testing.T) {
	db := setupInvitationTestDB(t)
	svc := newTestInvitationService(db)

	// 创建多个邀请
	for i := 0; i < 5; i++ {
		_, err := svc.CreateInvitation(1, 1, &CreateInvitationRequest{
			Position: fmt.Sprintf("岗位%d", i),
		})
		require.NoError(t, err)
	}

	// 查询全部
	items, total, err := svc.ListInvitations(1, ListInvitationsQuery{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, items, 5)

	// 按状态过滤
	items, total, err = svc.ListInvitations(1, ListInvitationsQuery{Status: InvitationStatusPending, Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)

	// 分页
	items, total, err = svc.ListInvitations(1, ListInvitationsQuery{Page: 1, PageSize: 2})
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, items, 2)
}

// TestSubmitInvitation_PhoneDuplicate 同组织内手机号重复
func TestSubmitInvitation_PhoneDuplicate(t *testing.T) {
	db := setupInvitationTestDB(t)
	svc := newTestInvitationService(db)

	// 创建测试组织
	createInvTestOrg(db, 1, "测试公司")

	// 创建第一个员工
	empSvc := newTestService(db)
	_, err := createTestEmployeeViaService(empSvc, 1, 1, "已存在", "13800008000", "110108199001011234", "开发", "2026-01-15")
	require.NoError(t, err)

	// 创建邀请
	resp, err := svc.CreateInvitation(1, 1, &CreateInvitationRequest{Position: "开发"})
	require.NoError(t, err)

	// 尝试使用相同手机号提交
	submitReq := &SubmitInvitationRequest{
		Name:     "新员工",
		Phone:    "13800008000", // 重复手机号
		IDCard:   "110108199503152345",
		Position: "开发",
		HireDate: "2026-02-01",
	}
	err = svc.SubmitInvitation(resp.Token, submitReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "该手机号已存在")
}

// TestConfirmOnboarding_NotPending 非 pending 员工不可确认
func TestConfirmOnboarding_NotPending(t *testing.T) {
	db := setupInvitationTestDB(t)
	empSvc := newTestService(db)

	// 创建已在职的员工（先创建 pending 再确认）
	resp, err := createTestEmployeeViaService(empSvc, 1, 1, "张三", "13800008000", "110108199001011234", "开发", "2026-01-15")
	require.NoError(t, err)
	// 默认是 pending，先确认一次
	invSvc := newTestInvitationService(db)
	err = invSvc.ConfirmOnboarding(1, resp.ID)
	require.NoError(t, err)

	// 再次确认应失败
	err = invSvc.ConfirmOnboarding(1, resp.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrEmployeeNotPending, err)
}
