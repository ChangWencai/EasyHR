package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/wencai/easyhr/pkg/jwt"
)

const testAuthSecret = "test-secret-key-for-auth-middleware-testing-must-be-long"

func setupAuthTest() (*gin.Engine, *httptest.ResponseRecorder, *gin.Context) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	return r, w, c
}

func newTestRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func TestAuthMissingHeader(t *testing.T) {
	r, w, c := setupAuthTest()
	r.Use(Auth(testAuthSecret, newTestRedis()))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	c.Request, _ = http.NewRequest("GET", "/test", nil)
	r.HandleContext(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthInvalidFormat(t *testing.T) {
	r, w, c := setupAuthTest()
	r.Use(Auth(testAuthSecret, newTestRedis()))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat token")
	c.Request = req
	r.HandleContext(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthValidToken(t *testing.T) {
	token, _ := jwt.GenerateAccessToken(1, 100, "owner", testAuthSecret, time.Hour)

	rdb := newTestRedis()
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("redis not available")
	}

	r, _, c := setupAuthTest()
	var capturedUserID, capturedOrgID int64
	var capturedRole string
	r.Use(Auth(testAuthSecret, rdb))
	r.GET("/test", func(c *gin.Context) {
		capturedUserID = c.GetInt64("user_id")
		capturedOrgID = c.GetInt64("org_id")
		capturedRole = c.GetString("role")
		c.String(200, "ok")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	c.Request = req
	r.HandleContext(c)

	assert.Equal(t, int64(1), capturedUserID)
	assert.Equal(t, int64(100), capturedOrgID)
	assert.Equal(t, "owner", capturedRole)
}

func TestAuthExpiredToken(t *testing.T) {
	token, _ := jwt.GenerateAccessToken(1, 100, "owner", testAuthSecret, -time.Hour)

	r, w, c := setupAuthTest()
	r.Use(Auth(testAuthSecret, newTestRedis()))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	c.Request = req
	r.HandleContext(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthTamperedToken(t *testing.T) {
	token, _ := jwt.GenerateAccessToken(1, 100, "owner", testAuthSecret, time.Hour)
	tampered := token[:len(token)-5] + "xxxxx"

	r, w, c := setupAuthTest()
	r.Use(Auth(testAuthSecret, newTestRedis()))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tampered)
	c.Request = req
	r.HandleContext(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthBlacklistedToken(t *testing.T) {
	token, _ := jwt.GenerateAccessToken(1, 100, "owner", testAuthSecret, time.Hour)
	claims, _ := jwt.ParseToken(token, testAuthSecret)

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx := context.Background()
	rdb.Set(ctx, "token:blacklist:"+claims.ID, "1", time.Hour)
	defer rdb.Del(ctx, "token:blacklist:"+claims.ID)

	r, w, c := setupAuthTest()
	r.Use(Auth(testAuthSecret, rdb))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	c.Request = req
	r.HandleContext(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
