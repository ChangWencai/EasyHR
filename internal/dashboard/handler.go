package dashboard

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/response"
)

// Handler handles HTTP requests for the dashboard.
type Handler struct {
	svc ServiceInterface
}

// NewHandler creates a new DashboardHandler.
func NewHandler(svc ServiceInterface) *Handler {
	return &Handler{svc: svc}
}

// GetDashboard handles GET /api/v1/dashboard.
// org_id is extracted from the JWT context (set by auth middleware).
func (h *Handler) GetDashboard(c *gin.Context) {
	orgIDVal, exists := c.Get("org_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, 40100, "missing org_id in context")
		return
	}
	orgID, ok := orgIDVal.(int64)
	if !ok {
		response.Error(c, http.StatusUnauthorized, 40100, "invalid org_id type")
		return
	}

	result, err := h.svc.GetDashboard(c.Request.Context(), orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, "failed to get dashboard: "+err.Error())
		return
	}

	response.Success(c, result)
}

// GetEmployeeDashboard handles GET /api/v1/dashboard/employee-dashboard.
func (h *Handler) GetEmployeeDashboard(c *gin.Context) {
	orgIDVal, exists := c.Get("org_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, 40100, "missing org_id in context")
		return
	}
	orgID, ok := orgIDVal.(int64)
	if !ok {
		response.Error(c, http.StatusUnauthorized, 40100, "invalid org_id type")
		return
	}

	result, err := h.svc.GetEmployeeDashboard(c.Request.Context(), orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, "failed to get employee dashboard: "+err.Error())
		return
	}

	response.Success(c, result)
}

// GetTodoStats handles GET /api/v1/dashboard/todo-stats.
func (h *Handler) GetTodoStats(c *gin.Context) {
	orgIDVal, exists := c.Get("org_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, 40100, "missing org_id in context")
		return
	}
	orgID, ok := orgIDVal.(int64)
	if !ok {
		response.Error(c, http.StatusUnauthorized, 40100, "invalid org_id type")
		return
	}

	result, err := h.svc.GetTodoStats(c.Request.Context(), orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, "failed to get todo stats: "+err.Error())
		return
	}

	response.Success(c, result)
}

// GetTimeLimitedStats handles GET /api/v1/dashboard/time-limited-stats.
func (h *Handler) GetTimeLimitedStats(c *gin.Context) {
	orgIDVal, exists := c.Get("org_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, 40100, "missing org_id in context")
		return
	}
	orgID, ok := orgIDVal.(int64)
	if !ok {
		response.Error(c, http.StatusUnauthorized, 40100, "invalid org_id type")
		return
	}

	result, err := h.svc.GetTimeLimitedStats(c.Request.Context(), orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, "failed to get time-limited stats: "+err.Error())
		return
	}

	response.Success(c, result)
}
