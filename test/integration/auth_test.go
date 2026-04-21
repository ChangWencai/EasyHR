package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wencai/easyhr/internal/audit"
	"github.com/wencai/easyhr/internal/city"
	"github.com/wencai/easyhr/internal/common/config"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/user"
	"github.com/wencai/easyhr/pkg/jwt"
	"github.com/wencai/easyhr/pkg/sms"
	"github.com/wencai/easyhr/test/testutil"
)

const testJWTSecret = "test-jwt-secret-for-integration-testing-must-be-long-enough-32chars"

func setupIntegrationTest(t *testing.T) (*gin.Engine, *redis.Client) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	t.Cleanup(func() { testutil.CleanupTestDB(db) })

	rdb := testutil.SetupTestRedis()
	if !testutil.WaitForRedis(rdb) {
		t.Skip("redis not available for integration test")
	}
	t.Cleanup(func() { testutil.CleanupTestRedis(rdb) })

	smsClient, _ := sms.NewClient(sms.Config{
		AccessKeyID:     "test",
		AccessKeySecret: "test",
		SignName:        "Test",
		TemplateCode:    "SMS_TEST",
	})

	userRepo := user.NewRepository(db)
	userSvc := user.NewService(userRepo, rdb, smsClient, config.JWTConfig{
		Secret:     testJWTSecret,
		AccessTTL:  time.Hour,
		RefreshTTL: 24 * time.Hour,
	}, config.CryptoConfig{
		AESKey: "0123456789abcdef0123456789abcdef",
	})
	userHandler := user.NewHandler(userSvc)
	authMiddleware := middleware.Auth(testJWTSecret, rdb)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RequestLogger())

	v1 := r.Group("/api/v1")
	{
		userHandler.RegisterRoutes(v1, authMiddleware)
		city.NewHandler(db).RegisterRoutes(v1)
		audit.NewHandler(audit.NewRepository(db)).RegisterRoutes(v1)
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
	}

	return r, rdb
}

func TestIntegrationAuthFlow(t *testing.T) {
	r, rdb := setupIntegrationTest(t)
	ctx := context.Background()

	phone := "13800138000"
	code := "123456"
	rdb.Set(ctx, "sms:code:"+phone, code, 5*time.Minute)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/v1/auth/login",
		strings.NewReader(`{"phone":"`+phone+`","code":"`+code+`"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	r.HandleContext(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"access_token"`)
	assert.Contains(t, w.Body.String(), `"onboarding_required":true`)
}

func TestIntegrationCityList(t *testing.T) {
	r, _ := setupIntegrationTest(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/cities", nil)
	r.HandleContext(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"name":"北京"`)
}

func TestIntegrationAuthRequired(t *testing.T) {
	r, _ := setupIntegrationTest(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/v1/auth/logout", nil)
	r.HandleContext(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestIntegrationRBACMemberCannotCreateUser(t *testing.T) {
	r, rdb := setupIntegrationTest(t)

	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	defer testutil.CleanupTestDB(db)

	org, err := testutil.CreateTestOrg(db, "TestOrg", "91110000MA00000001", "北京")
	require.NoError(t, err)
	member, err := testutil.CreateTestUser(db, org.ID, "Member", "hash2", "member")
	require.NoError(t, err)

	memberToken, err := jwt.GenerateAccessToken(member.ID, org.ID, "member", testJWTSecret, time.Hour)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/v1/users",
		strings.NewReader(`{"phone":"13900139000","name":"NewUser","role":"admin"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Authorization", "Bearer "+memberToken)
	r.HandleContext(c)

	assert.Equal(t, http.StatusForbidden, w.Code)

	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request, _ = http.NewRequest("GET", "/api/v1/users", nil)
	c2.Request.Header.Set("Authorization", "Bearer "+memberToken)
	r.HandleContext(c2)

	assert.Equal(t, http.StatusOK, w2.Code)
	_ = rdb
}

func TestIntegrationOwnerCanCreateUser(t *testing.T) {
	r, rdb := setupIntegrationTest(t)

	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	defer testutil.CleanupTestDB(db)

	org, err := testutil.CreateTestOrg(db, "TestOrg", "91110000MA00000002", "上海")
	require.NoError(t, err)
	owner, err := testutil.CreateTestUser(db, org.ID, "Owner", "hash3", "owner")
	require.NoError(t, err)

	ownerToken, err := jwt.GenerateAccessToken(owner.ID, org.ID, "owner", testJWTSecret, time.Hour)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/v1/users",
		strings.NewReader(`{"phone":"13900139000","name":"NewAdmin","role":"admin"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Authorization", "Bearer "+ownerToken)
	r.HandleContext(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"message":"子账号创建成功"`)
	_ = rdb
}
