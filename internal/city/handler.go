package city

import (
	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/response"
	"gorm.io/gorm"
)

type Handler struct {
	repo *Repository
}

// NewHandler 创建 Handler（依赖注入 db）
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{repo: NewRepository(db)}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/cities", h.List)
}

func (h *Handler) List(c *gin.Context) {
	level := c.Query("level")
	province := c.Query("province")  // 省名（兼容旧接口）
	cityParam := c.Query("city")    // 城市名

	var areas []AreaCode
	var err error

	switch {
	case level != "":
		// 按层级查询：level=1(省) level=2(城市) level=3(区县)
		l := 1
		if level == "2" {
			l = 2
		} else if level == "3" {
			l = 3
		}
		areas, err = h.repo.GetByLevel(l)
	case province != "":
		// 按省名筛选城市（level=2 且父编码对应该省）
		// 先找省编码，再找子级城市
		provinces, err := h.repo.GetByLevel(LevelProvince)
		if err != nil {
			response.Error(c, 500, 50000, "查询失败")
			return
		}
		var provinceCode int64
		for _, p := range provinces {
			if p.Name == province {
				provinceCode = p.Code
				break
			}
		}
		if provinceCode == 0 {
			response.Success(c, []AreaCode{})
			return
		}
		areas, err = h.repo.GetByLevelAndPcode(LevelCity, provinceCode)
	case cityParam != "":
		// 按城市名筛选区县（level=3 且父编码对应该城市）
		cities, err := h.repo.GetByLevel(LevelCity)
		if err != nil {
			response.Error(c, 500, 50000, "查询失败")
			return
		}
		var cityCode int64
		for _, city := range cities {
			if city.Name == cityParam {
				cityCode = city.Code
				break
			}
		}
		if cityCode == 0 {
			response.Success(c, []AreaCode{})
			return
		}
		areas, err = h.repo.GetByLevelAndPcode(LevelDistrict, cityCode)
	default:
		// 默认返回省份列表
		areas, err = h.repo.GetByLevel(LevelProvince)
	}

	if err != nil {
		response.Error(c, 500, 50000, "查询失败")
		return
	}

	response.Success(c, areas)
}