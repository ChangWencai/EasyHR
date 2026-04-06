package employee

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
)

// Handler 员工 HTTP 端点
type Handler struct {
	svc *Service
}

// NewHandler 创建员工 Handler
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	authGroup := rg.Group("")
	authGroup.Use(authMiddleware)

	authGroup.POST("/employees", middleware.RequireRole("owner", "admin"), h.CreateEmployee)
	authGroup.GET("/employees", h.ListEmployees)
	authGroup.GET("/employees/export", middleware.RequireRole("owner", "admin"), h.ExportExcel)
	authGroup.GET("/employees/:id", h.GetEmployee)
	authGroup.PUT("/employees/:id", middleware.RequireRole("owner", "admin"), h.UpdateEmployee)
	authGroup.DELETE("/employees/:id", middleware.RequireRole("owner", "admin"), h.DeleteEmployee)
	authGroup.POST("/employees/:id/sensitive", middleware.RequireRole("owner", "admin"), h.GetSensitiveInfo)
}

// CreateEmployee 创建员工
func (h *Handler) CreateEmployee(c *gin.Context) {
	var req CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	emp, err := h.svc.CreateEmployee(orgID, userID, &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 20100, err.Error())
		return
	}

	response.Success(c, emp)
}

// ListEmployees 员工列表
func (h *Handler) ListEmployees(c *gin.Context) {
	var query ListQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	employees, total, err := h.svc.ListEmployees(orgID, query)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 20101, "查询员工列表失败")
		return
	}

	response.PageSuccess(c, employees, total, query.Page, query.PageSize)
}

// GetEmployee 获取员工详情
func (h *Handler) GetEmployee(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的员工ID")
		return
	}

	emp, err := h.svc.GetEmployee(orgID, id)
	if err != nil {
		response.Error(c, http.StatusNotFound, 20102, err.Error())
		return
	}

	response.Success(c, emp)
}

// UpdateEmployee 更新员工信息
func (h *Handler) UpdateEmployee(c *gin.Context) {
	var req UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的员工ID")
		return
	}

	emp, err := h.svc.UpdateEmployee(orgID, userID, id, &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 20103, err.Error())
		return
	}

	response.Success(c, emp)
}

// DeleteEmployee 删除员工
func (h *Handler) DeleteEmployee(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的员工ID")
		return
	}

	if err := h.svc.DeleteEmployee(orgID, id); err != nil {
		response.Error(c, http.StatusBadRequest, 20104, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "员工已删除"})
}

// ExportExcel 导出员工列表为 Excel
func (h *Handler) ExportExcel(c *gin.Context) {
	var query ListQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	data, err := h.svc.ExportExcel(orgID, query)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 20105, "导出失败")
		return
	}

	filename := fmt.Sprintf("employees_%s.xlsx", time.Now().Format("20060102150405"))
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// GetSensitiveInfo 获取员工完整敏感信息
func (h *Handler) GetSensitiveInfo(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的员工ID")
		return
	}

	info, err := h.svc.GetSensitiveInfo(orgID, id)
	if err != nil {
		response.Error(c, http.StatusNotFound, 20106, err.Error())
		return
	}

	response.Success(c, info)
}
