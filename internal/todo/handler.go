package todo

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/response"
)

// Handler HTTP 处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建 Handler
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListTodos handles GET /api/v1/todos
func (h *Handler) ListTodos(c *gin.Context) {
	orgIDVal, exists := c.Get("org_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, 40100, "missing org_id")
		return
	}
	orgID := orgIDVal.(int64)

	var req ListTodosRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid query params")
		return
	}

	result, err := h.svc.ListTodos(c.Request.Context(), orgID, &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, "list todos failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

// PinTodo handles PUT /api/v1/todos/:id/pin
func (h *Handler) PinTodo(c *gin.Context) {
	orgIDVal, exists := c.Get("org_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, 40100, "missing org_id")
		return
	}
	orgID := orgIDVal.(int64)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	var req struct {
		Pinned bool `json:"pinned"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid body")
		return
	}

	if err := h.svc.PinTodo(c.Request.Context(), orgID, id, req.Pinned); err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, "pin todo failed: "+err.Error())
		return
	}

	response.Success(c, gin.H{"id": id, "pinned": req.Pinned})
}

// ExportTodos handles GET /api/v1/todos/export
func (h *Handler) ExportTodos(c *gin.Context) {
	orgIDVal, exists := c.Get("org_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, 40100, "missing org_id")
		return
	}
	orgID := orgIDVal.(int64)

	items, err := h.svc.ExportTodos(c.Request.Context(), orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, "export todos failed: "+err.Error())
		return
	}

	if err := ExportTodosExcel(c, items); err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, "generate excel failed: "+err.Error())
		return
	}
}

// ListCarousels handles GET /api/v1/carousels
func (h *Handler) ListCarousels(c *gin.Context) {
	orgIDVal, exists := c.Get("org_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, 40100, "missing org_id")
		return
	}
	orgID := orgIDVal.(int64)

	items, err := h.svc.ListCarousels(c.Request.Context(), orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, "list carousels failed: "+err.Error())
		return
	}

	response.Success(c, items)
}

// InviteTodo handles POST /api/v1/todos/:id/invite
func (h *Handler) InviteTodo(c *gin.Context) {
	orgIDVal, exists := c.Get("org_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, 40100, "missing org_id")
		return
	}
	userIDVal, _ := c.Get("user_id")
	orgID := orgIDVal.(int64)
	userID, _ := userIDVal.(int64)

	idStr := c.Param("id")
	todoID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	result, err := h.svc.InviteTodo(c.Request.Context(), orgID, todoID, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, "invite failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

// TerminateTodo handles PUT /api/v1/todos/:id/terminate
func (h *Handler) TerminateTodo(c *gin.Context) {
	orgIDVal, exists := c.Get("org_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, 40100, "missing org_id")
		return
	}
	orgID := orgIDVal.(int64)

	idStr := c.Param("id")
	todoID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	if err := h.svc.TerminateTodo(c.Request.Context(), orgID, todoID); err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, "terminate failed: "+err.Error())
		return
	}

	response.Success(c, gin.H{"id": todoID})
}

// VerifyInviteToken handles GET /api/v1/todos/invite/:token
// Public endpoint -- no auth required
func (h *Handler) VerifyInviteToken(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		response.Error(c, http.StatusBadRequest, 40000, "missing token")
		return
	}

	result, err := h.svc.VerifyInviteToken(c.Request.Context(), token)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, "verify failed: "+err.Error())
		return
	}

	if !result.Valid {
		if result.Expired {
			response.Error(c, http.StatusGone, 41001, "邀请链接已过期")
			return
		}
		response.Error(c, http.StatusNotFound, 40400, "邀请链接无效")
		return
	}

	response.Success(c, result)
}

// GetInviteTodo handles GET /api/v1/todos/invite/:token/detail -- public
func (h *Handler) GetInviteTodo(c *gin.Context) {
	token := c.Param("token")
	result, err := h.svc.VerifyInviteToken(c.Request.Context(), token)
	if err != nil || !result.Valid {
		response.Error(c, http.StatusNotFound, 40400, "邀请不存在")
		return
	}
	response.Success(c, result)
}

// SubmitInvite handles POST /api/v1/todos/invite/:token/submit
// Public endpoint -- no auth required. Used by InviteFillPage.vue.
func (h *Handler) SubmitInvite(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		response.Error(c, http.StatusBadRequest, 40000, "missing token")
		return
	}

	var req SubmitInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid request: name is required")
		return
	}

	result, err := h.svc.SubmitInvite(c.Request.Context(), token, &req)
	if err != nil {
		// 判断错误类型返回合适状态码
		errMsg := err.Error()
		if errMsg == "invalid token" {
			response.Error(c, http.StatusNotFound, 40400, "邀请链接无效")
			return
		}
		if errMsg == "link already used" {
			response.Error(c, http.StatusConflict, 40900, "该链接已使用过")
			return
		}
		if errMsg == "link expired" {
			response.Error(c, http.StatusGone, 41001, "邀请链接已过期")
			return
		}
		response.Error(c, http.StatusInternalServerError, 50000, "submit failed: "+errMsg)
		return
	}

	response.Success(c, result)
}

// ListAllCarousels handles GET /api/v1/carousels/admin -- all carousels for management
func (h *Handler) ListAllCarousels(c *gin.Context) {
	orgIDVal, exists := c.Get("org_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, 40100, "missing org_id")
		return
	}
	orgID := orgIDVal.(int64)

	items, err := h.svc.ListAllCarousels(c.Request.Context(), orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, "list carousels failed: "+err.Error())
		return
	}

	response.Success(c, gin.H{"data": items})
}

// CreateCarousel handles POST /api/v1/carousels
func (h *Handler) CreateCarousel(c *gin.Context) {
	orgIDVal, exists := c.Get("org_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, 40100, "missing org_id")
		return
	}
	orgID := orgIDVal.(int64)

	var req CarouselRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid request: image_url is required")
		return
	}

	item, err := h.svc.CreateCarousel(c.Request.Context(), orgID, &req)
	if err != nil {
		// Return 400 for max carousel limit error
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}

	response.Success(c, gin.H{"data": item})
}

// UpdateCarousel handles PUT /api/v1/carousels/:id
func (h *Handler) UpdateCarousel(c *gin.Context) {
	orgIDVal, exists := c.Get("org_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, 40100, "missing org_id")
		return
	}
	orgID := orgIDVal.(int64)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	var req CarouselRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid request")
		return
	}

	if err := h.svc.UpdateCarousel(c.Request.Context(), orgID, id, &req); err != nil {
		if err.Error() == "carousel not found" {
			response.Error(c, http.StatusNotFound, 40400, "轮播图不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, 50000, "update carousel failed: "+err.Error())
		return
	}

	response.Success(c, gin.H{"id": id})
}

// DeleteCarousel handles DELETE /api/v1/carousels/:id
func (h *Handler) DeleteCarousel(c *gin.Context) {
	orgIDVal, exists := c.Get("org_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, 40100, "missing org_id")
		return
	}
	orgID := orgIDVal.(int64)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid id")
		return
	}

	if err := h.svc.DeleteCarousel(c.Request.Context(), orgID, id); err != nil {
		if err.Error() == "carousel not found" {
			response.Error(c, http.StatusNotFound, 40400, "轮播图不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, 50000, "delete carousel failed: "+err.Error())
		return
	}

	response.Success(c, gin.H{"id": id})
}
