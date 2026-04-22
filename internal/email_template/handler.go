package email_template

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/logger"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
)

// Handler 邮箱模板 HTTP 端点
type Handler struct {
	svc *Service
}

// NewHandler 创建邮箱模板 Handler
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	authGroup := rg.Group("")
	authGroup.Use(authMiddleware, middleware.RequireOrg)

	authGroup.POST("/email-templates", middleware.RequireRole("owner", "admin"), h.Create)
	authGroup.GET("/email-templates", h.List)
	authGroup.PUT("/email-templates/:id", middleware.RequireRole("owner", "admin"), h.Update)
	authGroup.DELETE("/email-templates/:id", middleware.RequireRole("owner", "admin"), h.Delete)
}

// Create 创建模板
func (h *Handler) Create(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("CreateTemplate: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	tpl, err := h.svc.CreateTemplate(orgID, userID, &req)
	if err != nil {
		logger.SugarLogger.Debugw("CreateTemplate: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusBadRequest, 20400, err.Error())
		return
	}

	response.Success(c, tpl)
}

// List 模板列表
func (h *Handler) List(c *gin.Context) {
	var query ListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")

	templates, total, err := h.svc.ListTemplates(orgID, query)
	if err != nil {
		logger.SugarLogger.Debugw("ListTemplates: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, 20401, "查询模板列表失败")
		return
	}

	response.PageSuccess(c, templates, total, query.Page, query.PageSize)
}

// Update 更新模板
func (h *Handler) Update(c *gin.Context) {
	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("UpdateTemplate: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板ID")
		return
	}

	tpl, err := h.svc.UpdateTemplate(orgID, userID, id, &req)
	if err != nil {
		logger.SugarLogger.Debugw("UpdateTemplate: 失败", "error", err.Error(), "template_id", id)
		response.Error(c, http.StatusBadRequest, 20402, err.Error())
		return
	}

	response.Success(c, tpl)
}

// Delete 删除模板
func (h *Handler) Delete(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板ID")
		return
	}

	if err := h.svc.DeleteTemplate(orgID, id); err != nil {
		logger.SugarLogger.Debugw("DeleteTemplate: 失败", "error", err.Error(), "template_id", id)
		response.Error(c, http.StatusBadRequest, 20403, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "模板已删除"})
}
