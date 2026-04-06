package city

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCityList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	h := NewHandler()
	h.RegisterRoutes(r.Group("/api/v1"))

	c.Request, _ = http.NewRequest("GET", "/api/v1/cities", nil)
	r.HandleContext(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"name":"北京"`)
	assert.Contains(t, w.Body.String(), `"province":"广东"`)
}

func TestCityListByProvince(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	h := NewHandler()
	h.RegisterRoutes(r.Group("/api/v1"))

	c.Request, _ = http.NewRequest("GET", "/api/v1/cities?province=广东", nil)
	r.HandleContext(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"name":"广州"`)
	assert.Contains(t, w.Body.String(), `"name":"深圳"`)
	assert.NotContains(t, w.Body.String(), `"name":"北京"`)
}
