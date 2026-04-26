package sms_template

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/logger"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	authGroup := rg.Group("")
	authGroup.Use(authMiddleware, middleware.RequireOrg)

	authGroup.POST("/sms-templates", middleware.RequireRole("owner", "admin"), h.Create)
	authGroup.GET("/sms-templates", h.List)
	authGroup.PUT("/sms-templates/:id", middleware.RequireRole("owner", "admin"), h.Update)
	authGroup.DELETE("/sms-templates/:id", middleware.RequireRole("owner", "admin"), h.Delete)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateSmsTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("CreateSmsTemplate: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	tpl, err := h.svc.CreateTemplate(orgID, userID, &req)
	if err != nil {
		logger.SugarLogger.Debugw("CreateSmsTemplate: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusBadRequest, 20500, err.Error())
		return
	}

	response.Success(c, tpl)
}

func (h *Handler) List(c *gin.Context) {
	var query ListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")

	templates, total, err := h.svc.ListTemplates(orgID, query)
	if err != nil {
		logger.SugarLogger.Debugw("ListSmsTemplates: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, 20501, "查询模板列表失败")
		return
	}

	response.PageSuccess(c, templates, total, query.Page, query.PageSize)
}

func (h *Handler) Update(c *gin.Context) {
	var req UpdateSmsTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("UpdateSmsTemplate: 参数错误", "error", err.Error())
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
		logger.SugarLogger.Debugw("UpdateSmsTemplate: 失败", "error", err.Error(), "template_id", id)
		response.Error(c, http.StatusBadRequest, 20502, err.Error())
		return
	}

	response.Success(c, tpl)
}

func (h *Handler) Delete(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板ID")
		return
	}

	if err := h.svc.DeleteTemplate(orgID, id); err != nil {
		logger.SugarLogger.Debugw("DeleteSmsTemplate: 失败", "error", err.Error(), "template_id", id)
		response.Error(c, http.StatusBadRequest, 20503, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "模板已删除"})
}
