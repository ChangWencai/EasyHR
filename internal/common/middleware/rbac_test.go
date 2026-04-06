package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequireRoleAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.Use(func(ctx *gin.Context) {
		ctx.Set("role", "admin")
		ctx.Next()
	})
	r.Use(RequireRole("owner", "admin"))
	r.GET("/test", func(ctx *gin.Context) {
		ctx.String(200, "ok")
	})

	c.Request, _ = http.NewRequest("GET", "/test", nil)
	r.HandleContext(c)

	assert.Equal(t, 200, w.Code)
}

func TestRequireRoleForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.Use(func(ctx *gin.Context) {
		ctx.Set("role", "member")
		ctx.Next()
	})
	r.Use(RequireRole("owner"))
	r.GET("/test", func(ctx *gin.Context) {
		ctx.String(200, "ok")
	})

	c.Request, _ = http.NewRequest("GET", "/test", nil)
	r.HandleContext(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireRoleNoRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.Use(RequireRole("owner"))
	r.GET("/test", func(ctx *gin.Context) {
		ctx.String(200, "ok")
	})

	c.Request, _ = http.NewRequest("GET", "/test", nil)
	r.HandleContext(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
