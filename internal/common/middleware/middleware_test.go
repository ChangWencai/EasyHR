package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTenantScope(t *testing.T) {
	scope := TenantScope(42)
	assert.NotNil(t, scope)
	// Scope is a function that modifies *gorm.DB — verify it's callable
	// Full integration test requires real DB, unit test verifies function signature
}

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.Use(CORS())
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	c.Request, _ = http.NewRequest("GET", "/test", nil)
	r.HandleContext(c)

	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
}

func TestCORSOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.Use(CORS())
	r.OPTIONS("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	c.Request, _ = http.NewRequest("OPTIONS", "/test", nil)
	r.HandleContext(c)

	assert.Equal(t, 204, w.Code)
}

func TestRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
}

func TestRequestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.Use(RequestLogger())
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	c.Request, _ = http.NewRequest("GET", "/test", nil)
	r.HandleContext(c)

	assert.Equal(t, 200, w.Code)
}
