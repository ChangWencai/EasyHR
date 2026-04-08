package salary

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
)

// Handler 工资核算 HTTP 端点
type Handler struct {
	svc *Service
}

// NewHandler 创建工资核算 Handler
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	salary := rg.Group("/salary", authMiddleware)
	{
		salary.GET("/template", h.GetTemplate)
		salary.PUT("/template", middleware.RequireRole("owner", "admin"), h.UpdateTemplate)
		salary.GET("/items", h.GetEmployeeItems)
		salary.PUT("/items/:employee_id", middleware.RequireRole("owner", "admin"), h.SetEmployeeItems)
	}
}

// GetTemplate 获取企业薪资模板
func (h *Handler) GetTemplate(c *gin.Context) {
	orgID := c.GetInt64("org_id")

	tmpl, err := h.svc.GetTemplate(orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, tmpl)
}

// UpdateTemplate 更新企业薪资模板
func (h *Handler) UpdateTemplate(c *gin.Context) {
	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	if err := h.svc.UpdateTemplate(orgID, userID, req.Items); err != nil {
		response.Error(c, http.StatusBadRequest, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetEmployeeItems 获取员工薪资项
func (h *Handler) GetEmployeeItems(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	employeeIDStr := c.Query("employee_id")
	month := c.Query("month")

	if employeeIDStr == "" || month == "" {
		response.BadRequest(c, "employee_id 和 month 为必填参数")
		return
	}

	employeeID, err := strconv.ParseInt(employeeIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "employee_id 格式错误")
		return
	}

	items, err := h.svc.GetEmployeeItems(orgID, employeeID, month)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, items)
}

// SetEmployeeItems 设置员工薪资项
func (h *Handler) SetEmployeeItems(c *gin.Context) {
	employeeIDStr := c.Param("employee_id")
	employeeID, err := strconv.ParseInt(employeeIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "employee_id 格式错误")
		return
	}

	var req SetEmployeeItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	if err := h.svc.SetEmployeeItems(orgID, userID, employeeID, req.Month, req.Items); err != nil {
		response.Error(c, http.StatusBadRequest, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, fmt.Sprintf("员工 %d 薪资项已更新", employeeID))
}
