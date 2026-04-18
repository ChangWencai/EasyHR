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
