package wxmp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func newTestRedis(t *testing.T) *redis.Client {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("redis not available")
	}
	return rdb
}

func genHandlerTestToken(userID, employeeID, orgID uint, role string) string {
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
	tokenStr, _ := token.SignedString([]byte("test-secret"))
	return tokenStr
}

func TestHandler_Login_ValidCode(t *testing.T) {
	rdb := newTestRedis(t)
	defer rdb.Close()

	mockRepo := &MockWXMPRepository{}
	mockRepo.Member = &MemberInfo{
		UserID: 1, EmployeeID: 10, OrgID: 100, Name: "张三", Phone: "138****8000", Role: "MEMBER",
	}
	svc := &WXMPService{
		repo:       mockRepo,
		jwtSecret:  []byte("test-secret"),
		accessTTL:  1 * time.Hour,
		refreshTTL: 7 * 24 * time.Hour,
		rdb:        rdb,
		cryptoKey:  []byte("test-crypto-key-32bytes!!!"),
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(svc, nil)
	r.POST("/wxmp/auth/login", h.Login)

	ctx := context.Background()
	rdb.Set(ctx, "sms:code:13800138000", "123456", 5*time.Minute)
	defer rdb.Del(ctx, "sms:code:13800138000")

	body := map[string]string{"phone": "13800138000", "code": "123456"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/wxmp/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	assert.NotNil(t, resp["data"])
}

func TestHandler_Login_InvalidCode(t *testing.T) {
	rdb := newTestRedis(t)
	defer rdb.Close()

	svc := &WXMPService{
		repo:      &MockWXMPRepository{},
		jwtSecret: []byte("test-secret"),
		rdb:       rdb,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(svc, nil)
	r.POST("/wxmp/auth/login", h.Login)

	ctx := context.Background()
	rdb.Set(ctx, "sms:code:13800138000", "123456", 5*time.Minute)
	defer rdb.Del(ctx, "sms:code:13800138000")

	body := map[string]string{"phone": "13800138000", "code": "999999"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/wxmp/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_Payslips_NoToken(t *testing.T) {
	rdb := newTestRedis(t)
	defer rdb.Close()

	svc := &WXMPService{
		repo:      &MockWXMPRepository{},
		jwtSecret: []byte("test-secret"),
		rdb:       rdb,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(svc, nil)
	wxmp := r.Group("/wxmp")
	wxmp.Use(WXMPMemberAuth([]byte("test-secret")))
	wxmp.GET("/payslips", h.ListPayslips)

	req := httptest.NewRequest(http.MethodGet, "/wxmp/payslips", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_Payslips_WithValidToken(t *testing.T) {
	rdb := newTestRedis(t)
	defer rdb.Close()

	mockRepo := &MockWXMPRepository{
		Payslips: []PayslipSummary{
			{ID: 1, Year: 2026, Month: 3, GrossPay: "10000.00", NetPay: "8000.00", Status: "pending"},
		},
	}
	svc := &WXMPService{
		repo:      mockRepo,
		jwtSecret: []byte("test-secret"),
		rdb:       rdb,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(svc, nil)
	wxmp := r.Group("/wxmp")
	wxmp.Use(WXMPMemberAuth([]byte("test-secret")))
	wxmp.GET("/payslips", h.ListPayslips)

	token := genHandlerTestToken(1, 10, 100, "MEMBER")
	req := httptest.NewRequest(http.MethodGet, "/wxmp/payslips", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
}

func TestHandler_Payslips_NonMemberToken(t *testing.T) {
	rdb := newTestRedis(t)
	defer rdb.Close()

	svc := &WXMPService{
		repo:      &MockWXMPRepository{},
		jwtSecret: []byte("test-secret"),
		rdb:       rdb,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(svc, nil)
	wxmp := r.Group("/wxmp")
	wxmp.Use(WXMPMemberAuth([]byte("test-secret")))
	wxmp.GET("/payslips", h.ListPayslips)

	token := genHandlerTestToken(1, 10, 100, "ADMIN")
	req := httptest.NewRequest(http.MethodGet, "/wxmp/payslips", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandler_PayslipDetail_NoVerifyToken(t *testing.T) {
	rdb := newTestRedis(t)
	defer rdb.Close()

	svc := &WXMPService{
		repo:      &MockWXMPRepository{},
		jwtSecret: []byte("test-secret"),
		rdb:       rdb,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(svc, nil)
	wxmp := r.Group("/wxmp")
	wxmp.Use(WXMPMemberAuth([]byte("test-secret")))
	wxmp.GET("/payslips/:id", h.GetPayslipDetail)

	token := genHandlerTestToken(1, 10, 100, "MEMBER")
	req := httptest.NewRequest(http.MethodGet, "/wxmp/payslips/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["message"], "验证")
}

func TestHandler_Expenses_Create(t *testing.T) {
	rdb := newTestRedis(t)
	defer rdb.Close()

	svc := &WXMPService{
		repo:      &MockWXMPRepository{},
		jwtSecret: []byte("test-secret"),
		rdb:       rdb,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(svc, nil)
	wxmp := r.Group("/wxmp")
	wxmp.Use(WXMPMemberAuth([]byte("test-secret")))
	wxmp.POST("/expenses", h.CreateExpense)

	token := genHandlerTestToken(1, 10, 100, "MEMBER")
	body := map[string]interface{}{"type": "travel", "amount": "500.00", "description": "出差"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/wxmp/expenses", bytes.NewReader(jsonBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
}

func TestHandler_Expenses_List(t *testing.T) {
	rdb := newTestRedis(t)
	defer rdb.Close()

	mockRepo := &MockWXMPRepository{
		Expenses: []ExpenseDTO{
			{ID: 1, Type: "travel", Amount: "200.00", Status: "pending", CreatedAt: "2026-04-01"},
		},
	}
	svc := &WXMPService{
		repo:      mockRepo,
		jwtSecret: []byte("test-secret"),
		rdb:       rdb,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(svc, nil)
	wxmp := r.Group("/wxmp")
	wxmp.Use(WXMPMemberAuth([]byte("test-secret")))
	wxmp.GET("/expenses", h.ListExpenses)

	token := genHandlerTestToken(1, 10, 100, "MEMBER")
	req := httptest.NewRequest(http.MethodGet, "/wxmp/expenses", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].([]interface{})
	assert.Len(t, data, 1)
}

func TestHandler_Contracts_List(t *testing.T) {
	rdb := newTestRedis(t)
	defer rdb.Close()

	mockRepo := &MockWXMPRepository{
		Contracts: []ContractDTO{
			{ID: 1, ContractType: "labor", Status: "signed", StartDate: "2025-01-01", EndDate: "2026-01-01"},
		},
	}
	svc := &WXMPService{
		repo:      mockRepo,
		jwtSecret: []byte("test-secret"),
		rdb:       rdb,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(svc, nil)
	wxmp := r.Group("/wxmp")
	wxmp.Use(WXMPMemberAuth([]byte("test-secret")))
	wxmp.GET("/contracts", h.ListContracts)

	token := genHandlerTestToken(1, 10, 100, "MEMBER")
	req := httptest.NewRequest(http.MethodGet, "/wxmp/contracts", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_SocialInsurance_List(t *testing.T) {
	rdb := newTestRedis(t)
	defer rdb.Close()

	mockRepo := &MockWXMPRepository{
		SocialRecords: []SocialInsuranceDTO{
			{PaymentMonth: "2026-03", City: "北京", BaseAmount: "8000.00", TotalPersonal: "880.00"},
		},
	}
	svc := &WXMPService{
		repo:      mockRepo,
		jwtSecret: []byte("test-secret"),
		rdb:       rdb,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(svc, nil)
	wxmp := r.Group("/wxmp")
	wxmp.Use(WXMPMemberAuth([]byte("test-secret")))
	wxmp.GET("/social-insurance", h.GetSocialInsurance)

	token := genHandlerTestToken(1, 10, 100, "MEMBER")
	req := httptest.NewRequest(http.MethodGet, "/wxmp/social-insurance", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
