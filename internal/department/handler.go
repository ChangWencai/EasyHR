package department

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/logger"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
)

// DepartmentHandler 部门 HTTP 端点
type DepartmentHandler struct {
	svc *Service
}

// NewDepartmentHandler 创建部门 Handler
func NewDepartmentHandler(svc *Service) *DepartmentHandler {
	return &DepartmentHandler{svc: svc}
}

// RegisterRoutes 注册路由
func (h *DepartmentHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	authGroup := rg.Group("")
	authGroup.Use(authMiddleware, middleware.RequireOrg)

	authGroup.POST("/departments", middleware.RequireRole("owner", "admin"), h.CreateDepartment)
	authGroup.GET("/departments", h.ListDepartments)
	authGroup.GET("/departments/tree", h.GetTree)
	authGroup.GET("/departments/search", h.SearchTree)
	authGroup.GET("/departments/:id", h.GetDepartment)
	authGroup.PUT("/departments/:id", middleware.RequireRole("owner", "admin"), h.UpdateDepartment)
	authGroup.DELETE("/departments/:id", middleware.RequireRole("owner", "admin"), h.DeleteDepartment)
}

// CreateDepartment 创建部门
func (h *DepartmentHandler) CreateDepartment(c *gin.Context) {
	var req CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("CreateDepartment: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	dept, err := h.svc.CreateDepartment(orgID, userID, &req)
	if err != nil {
		logger.SugarLogger.Debugw("CreateDepartment: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusBadRequest, 20200, err.Error())
		return
	}

	response.Success(c, dept)
}

// ListDepartments 部门列表
func (h *DepartmentHandler) ListDepartments(c *gin.Context) {
	var query DepartmentListQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		logger.SugarLogger.Debugw("ListDepartments: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")

	departments, total, err := h.svc.ListDepartments(orgID, query.Page, query.PageSize)
	if err != nil {
		logger.SugarLogger.Debugw("ListDepartments: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, 20201, "查询部门列表失败")
		return
	}

	response.PageSuccess(c, departments, total, query.Page, query.PageSize)
}

// GetDepartment 获取部门详情
func (h *DepartmentHandler) GetDepartment(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的部门ID")
		return
	}

	dept, err := h.svc.GetDepartment(orgID, id)
	if err != nil {
		logger.SugarLogger.Debugw("GetDepartment: 失败", "error", err.Error(), "department_id", id)
		response.Error(c, http.StatusNotFound, 20202, err.Error())
		return
	}

	response.Success(c, dept)
}

// UpdateDepartment 更新部门
func (h *DepartmentHandler) UpdateDepartment(c *gin.Context) {
	var req UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("UpdateDepartment: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的部门ID")
		return
	}

	dept, err := h.svc.UpdateDepartment(orgID, userID, id, &req)
	if err != nil {
		logger.SugarLogger.Debugw("UpdateDepartment: 失败", "error", err.Error(), "department_id", id)
		response.Error(c, http.StatusBadRequest, 20203, err.Error())
		return
	}

	response.Success(c, dept)
}

// DeleteDepartment 删除部门
func (h *DepartmentHandler) DeleteDepartment(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的部门ID")
		return
	}

	if err := h.svc.DeleteDepartment(orgID, id); err != nil {
		logger.SugarLogger.Debugw("DeleteDepartment: 失败", "error", err.Error(), "department_id", id)
		response.Error(c, http.StatusBadRequest, 20204, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "部门已删除"})
}

// GetTree 获取组织架构树
func (h *DepartmentHandler) GetTree(c *gin.Context) {
	orgID := c.GetInt64("org_id")

	tree, err := h.svc.GetTree(orgID)
	if err != nil {
		logger.SugarLogger.Debugw("GetTree: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, 20205, "获取组织架构失败")
		return
	}

	response.Success(c, tree)
}

// SearchTree 搜索组织架构树
func (h *DepartmentHandler) SearchTree(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	keyword := c.Query("keyword")
	if keyword == "" {
		response.BadRequest(c, "搜索关键字不能为空")
		return
	}

	tree, err := h.svc.SearchTree(orgID, keyword)
	if err != nil {
		logger.SugarLogger.Debugw("SearchTree: 失败", "error", err.Error(), "org_id", orgID, "keyword", keyword)
		response.Error(c, http.StatusInternalServerError, 20206, "搜索组织架构失败")
		return
	}

	response.Success(c, tree)
}
