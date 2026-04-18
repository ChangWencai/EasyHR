package salary

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/logger"
	"github.com/wencai/easyhr/internal/common/response"
)

// GetDashboard 获取薪资看板数据（注册在 Handler 上，由 RegisterRoutes 使用）
func (h *Handler) GetDashboard(c *gin.Context) {
	yearStr := c.Query("year")
	monthStr := c.Query("month")

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		response.BadRequest(c, "year 参数错误")
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		response.BadRequest(c, "month 参数错误")
		return
	}

	orgID := c.GetInt64("org_id")
	logger.SugarLogger.Debugw("GetDashboard: 查询", "org_id", orgID, "year", year, "month", month)

	result, err := h.dashboardSvc.GetDashboard(c.Request.Context(), orgID, year, month)
	if err != nil {
		logger.SugarLogger.Debugw("GetDashboard: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, result)
}
