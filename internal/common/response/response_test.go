package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestSuccess(t *testing.T) {
	c, w := setupTestContext()
	Success(c, gin.H{"id": 1})

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	assert.Equal(t, "success", resp["message"])
	assert.NotNil(t, resp["data"])
}

func TestError(t *testing.T) {
	c, w := setupTestContext()
	Error(c, http.StatusBadRequest, 10001, "user not found")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(10001), resp["code"])
	assert.Equal(t, "user not found", resp["message"])
	assert.Nil(t, resp["data"])
}

func TestPageSuccess(t *testing.T) {
	c, w := setupTestContext()
	PageSuccess(c, []gin.H{{"id": 1}}, 10, 1, 20)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	meta := resp["meta"].(map[string]interface{})
	assert.Equal(t, float64(10), meta["total"])
	assert.Equal(t, float64(1), meta["page"])
	assert.Equal(t, float64(20), meta["page_size"])
}

func TestUnauthorized(t *testing.T) {
	c, w := setupTestContext()
	Unauthorized(c, "invalid token")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40100), resp["code"])
}

func TestForbidden(t *testing.T) {
	c, w := setupTestContext()
	Forbidden(c, "permission denied")

	assert.Equal(t, http.StatusForbidden, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40300), resp["code"])
}

func TestBadRequest(t *testing.T) {
	c, w := setupTestContext()
	BadRequest(c, "invalid phone number")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(40000), resp["code"])
}
