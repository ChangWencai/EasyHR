package audit

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/audit-logs", middleware.RequireRole("owner", "admin"), h.List)
}

func (h *Handler) List(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	module := c.Query("module")
	action := c.Query("action")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "20")

	var p, ps int
	if _, err := fmt.Sscanf(page, "%d", &p); err != nil || p < 1 {
		p = 1
	}
	if _, err := fmt.Sscanf(pageSize, "%d", &ps); err != nil || ps < 1 || ps > 100 {
		ps = 20
	}

	logs, total, err := h.repo.List(orgID, module, action, p, ps)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 90001, "查询审计日志失败")
		return
	}
	response.PageSuccess(c, logs, total, p, ps)
}
