package wxmp

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// newTestWXMPService creates a service with a mock repo and real/skip Redis
func newTestWXMPService(t *testing.T) (*WXMPService, *redis.Client) {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("redis not available")
	}

	mockRepo := &MockWXMPRepository{}
	svc := &WXMPService{
		repo:       mockRepo,
		jwtSecret:  []byte("test-secret"),
		accessTTL:  1 * time.Hour,
		refreshTTL: 7 * 24 * time.Hour,
		rdb:        rdb,
		cryptoKey:  []byte("test-crypto-key-32bytes!!!"),
	}
	return svc, rdb
}

func genTestWXMPToken(userID, employeeID, orgID uint, role, secret string) string {
	claims := WXMPClaims{
		UserID:     userID,
		EmployeeID: employeeID,
		OrgID:      orgID,
		Role:       role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        "test-jti",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(secret))
	return tokenStr
}

func TestLoginMember_ValidCode(t *testing.T) {
	svc, rdb := newTestWXMPService(t)
	defer rdb.Close()

	phone := "13800138000"
	code := "123456"
	mockRepo := svc.repo.(*MockWXMPRepository)
	mockRepo.Member = &MemberInfo{
		UserID: 1, EmployeeID: 10, OrgID: 100, Name: "张三", Phone: "138****8000", Role: "MEMBER",
	}

	ctx := context.Background()
	rdb.Set(ctx, "sms:code:"+phone, code, 5*time.Minute)
	defer rdb.Del(ctx, "sms:code:"+phone)

	resp, err := svc.LoginMember(ctx, phone, code)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)

	// 验证 JWT claims
	token, err := jwt.ParseWithClaims(resp.AccessToken, &WXMPClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	assert.NoError(t, err)
	claims := token.Claims.(*WXMPClaims)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, uint(10), claims.EmployeeID)
	assert.Equal(t, uint(100), claims.OrgID)
	assert.Equal(t, "MEMBER", claims.Role)
}

func TestLoginMember_InvalidCode(t *testing.T) {
	svc, rdb := newTestWXMPService(t)
	defer rdb.Close()

	phone := "13800138000"
	ctx := context.Background()
	rdb.Set(ctx, "sms:code:"+phone, "123456", 5*time.Minute)
	defer rdb.Del(ctx, "sms:code:"+phone)

	_, err := svc.LoginMember(ctx, phone, "999999")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "验证码错误")
}

func TestLoginMember_ExpiredCode(t *testing.T) {
	svc, rdb := newTestWXMPService(t)
	defer rdb.Close()

	phone := "13800138000"
	ctx := context.Background()
	// 不预存验证码，模拟已过期

	_, err := svc.LoginMember(ctx, phone, "123456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "验证码已过期")
}

func TestLoginMember_NonExistentPhone(t *testing.T) {
	svc, rdb := newTestWXMPService(t)
	defer rdb.Close()

	phone := "13800138000"
	code := "123456"
	mockRepo := svc.repo.(*MockWXMPRepository)
	mockRepo.Err = assert.AnError

	ctx := context.Background()
	rdb.Set(ctx, "sms:code:"+phone, code, 5*time.Minute)
	defer rdb.Del(ctx, "sms:code:"+phone)

	_, err := svc.LoginMember(ctx, phone, code)
	assert.Error(t, err)
}

func TestGetPayslips_ReturnsList(t *testing.T) {
	svc, rdb := newTestWXMPService(t)
	defer rdb.Close()

	mockRepo := svc.repo.(*MockWXMPRepository)
	mockRepo.Payslips = []PayslipSummary{
		{ID: 1, Year: 2026, Month: 3, GrossPay: "10000.00", NetPay: "8000.00", Status: "pending"},
		{ID: 2, Year: 2026, Month: 2, GrossPay: "9500.00", NetPay: "7600.00", Status: "paid"},
	}

	ctx := context.Background()
	result, err := svc.GetPayslips(ctx, 100, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, uint(1), result[0].ID)
}

func TestGetPayslips_EmptyList(t *testing.T) {
	svc, rdb := newTestWXMPService(t)
	defer rdb.Close()

	mockRepo := svc.repo.(*MockWXMPRepository)
	mockRepo.Payslips = nil

	ctx := context.Background()
	result, err := svc.GetPayslips(ctx, 100, 10)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestGetPayslipDetail_NotVerified(t *testing.T) {
	svc, rdb := newTestWXMPService(t)
	defer rdb.Close()

	ctx := context.Background()
	_, err := svc.GetPayslipDetail(ctx, 100, 10, 1, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "请先验证身份")
}

func TestGetPayslipDetail_InvalidToken(t *testing.T) {
	svc, rdb := newTestWXMPService(t)
	defer rdb.Close()

	ctx := context.Background()
	_, err := svc.GetPayslipDetail(ctx, 100, 10, 1, "wrong-token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "验证已过期")
}

func TestVerifyPayslipAccess_ValidCode(t *testing.T) {
	svc, rdb := newTestWXMPService(t)
	defer rdb.Close()

	phone := "13800138000"
	code := "123456"
	ctx := context.Background()
	rdb.Set(ctx, "sms:code:"+phone, code, 5*time.Minute)
	defer rdb.Del(ctx, "sms:code:"+phone)

	resp, err := svc.VerifyPayslipAccess(ctx, 100, 10, 1, code, phone)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.VerifyToken)
	assert.NotEmpty(t, resp.ExpiresAt)
}

func TestVerifyPayslipAccess_InvalidCode(t *testing.T) {
	svc, rdb := newTestWXMPService(t)
	defer rdb.Close()

	phone := "13800138000"
	ctx := context.Background()
	rdb.Set(ctx, "sms:code:"+phone, "123456", 5*time.Minute)
	defer rdb.Del(ctx, "sms:code:"+phone)

	_, err := svc.VerifyPayslipAccess(ctx, 100, 10, 1, "999999", phone)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "验证码错误")
}

func TestSignPayslip_Success(t *testing.T) {
	svc, rdb := newTestWXMPService(t)
	defer rdb.Close()

	mockRepo := svc.repo.(*MockWXMPRepository)
	mockRepo.SignPayslipErr = nil

	ctx := context.Background()
	err := svc.SignPayslip(ctx, 100, 10, 1)
	assert.NoError(t, err)
}

func TestSignPayslip_NotFound(t *testing.T) {
	svc, rdb := newTestWXMPService(t)
	defer rdb.Close()

	mockRepo := svc.repo.(*MockWXMPRepository)
	mockRepo.SignPayslipErr = assert.AnError

	ctx := context.Background()
	err := svc.SignPayslip(ctx, 100, 10, 999)
	assert.Error(t, err)
}

func TestGetContracts_ReturnsContracts(t *testing.T) {
	svc, rdb := newTestWXMPService(t)
	defer rdb.Close()

	mockRepo := svc.repo.(*MockWXMPRepository)
	mockRepo.Contracts = []ContractDTO{
		{ID: 1, ContractType: "labor", Status: "signed", StartDate: "2025-01-01", EndDate: "2026-01-01"},
	}

	ctx := context.Background()
	result, err := svc.GetContracts(ctx, 100, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "labor", result[0].ContractType)
}

func TestGetSocialInsurance_ReturnsRecords(t *testing.T) {
	svc, rdb := newTestWXMPService(t)
	defer rdb.Close()

	mockRepo := svc.repo.(*MockWXMPRepository)
	mockRepo.SocialRecords = []SocialInsuranceDTO{
		{PaymentMonth: "2026-03", City: "北京", BaseAmount: "8000.00", TotalPersonal: "880.00"},
	}

	ctx := context.Background()
	result, err := svc.GetSocialInsurance(ctx, 100, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "北京", result[0].City)
}

func TestCreateExpense_Success(t *testing.T) {
	svc, rdb := newTestWXMPService(t)
	defer rdb.Close()

	ctx := context.Background()
	req := &ExpenseRequest{
		Type:        "travel",
		Amount:      "500.00",
		Description: "出差报销",
		Attachments: []string{"oss://file1.jpg"},
	}
	exp, err := svc.CreateExpense(ctx, 100, 10, req)
	assert.NoError(t, err)
	assert.NotNil(t, exp)
	assert.Equal(t, "pending", exp.Status)
}

func TestGetExpenses_ReturnsList(t *testing.T) {
	svc, rdb := newTestWXMPService(t)
	defer rdb.Close()

	mockRepo := svc.repo.(*MockWXMPRepository)
	mockRepo.Expenses = []ExpenseDTO{
		{ID: 1, Type: "travel", Amount: "200.00", Status: "pending", CreatedAt: "2026-04-01"},
		{ID: 2, Type: "traffic", Amount: "50.00", Status: "approved", CreatedAt: "2026-03-15"},
	}

	ctx := context.Background()
	result, err := svc.GetExpenses(ctx, 100, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, uint(1), result[0].ID)
}
