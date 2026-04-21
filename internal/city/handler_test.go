package city

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/wencai/easyhr/test/testutil"
)

func TestCityList(t *testing.T) {
	db, err := testutil.SetupTestDB()
	if err != nil {
		t.Fatal(err)
	}
	db.AutoMigrate(&AreaCode{})
	// 插入测试数据
	db.Create(&AreaCode{Code: 110000, Name: "北京", Level: 1, Pcode: 0, Category: 1})
	db.Create(&AreaCode{Code: 440100, Name: "广州", Level: 2, Pcode: 440000, Category: 1})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	h := NewHandler(db)
	h.RegisterRoutes(r.Group("/api/v1"))

	c.Request, _ = http.NewRequest("GET", "/api/v1/cities", nil)
	r.HandleContext(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"name":"北京"`)
	assert.Contains(t, w.Body.String(), `"code":110000`)
}

func TestCityListByLevel(t *testing.T) {
	db, err := testutil.SetupTestDB()
	if err != nil {
		t.Fatal(err)
	}
	db.AutoMigrate(&AreaCode{})
	db.Create(&AreaCode{Code: 110000, Name: "北京", Level: 1, Pcode: 0, Category: 1})
	db.Create(&AreaCode{Code: 440100, Name: "广州", Level: 2, Pcode: 440000, Category: 1})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	h := NewHandler(db)
	h.RegisterRoutes(r.Group("/api/v1"))

	c.Request, _ = http.NewRequest("GET", "/api/v1/cities?level=1", nil)
	r.HandleContext(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"name":"北京"`)
	assert.NotContains(t, w.Body.String(), `"name":"广州"`)
}
