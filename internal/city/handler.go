package city

import (
	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/response"
)

type Handler struct {
	cities []City
}

func NewHandler() *Handler {
	return &Handler{cities: Cities}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/cities", h.List)
}

func (h *Handler) List(c *gin.Context) {
	province := c.Query("province")
	result := h.cities
	if province != "" {
		var filtered []City
		for _, city := range h.cities {
			if city.Province == province {
				filtered = append(filtered, city)
			}
		}
		result = filtered
	}
	response.Success(c, result)
}
