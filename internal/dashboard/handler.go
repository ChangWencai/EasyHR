package dashboard

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/response"
)

// Handler handles HTTP requests for dashboard endpoints.
type Handler struct {
	svc ServiceInterface
}

// NewHandler creates a new dashboard Handler.
func NewHandler(svc ServiceInterface) *Handler {
	return &Handler{svc: svc}
}

// GetDashboard handles GET /api/v1/dashboard.
// It extracts org_id from the JWT context and returns the dashboard result.
func (h *Handler) GetDashboard(c *gin.Context) {
	orgIDVal, exists := c.Get("org_id")
	if !exists {
		response.Unauthorized(c, "missing org_id in context")
		return
	}
	orgID, ok := orgIDVal.(uint)
	if !ok {
		response.Unauthorized(c, "invalid org_id type")
		return
	}

	result, err := h.svc.GetDashboard(c.Request.Context(), orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, "failed to get dashboard: "+err.Error())
		return
	}

	response.Success(c, result)
}

// RegisterDashboardRouter registers the dashboard router group.
func RegisterDashboardRouter(rg *gin.RouterGroup, svc ServiceInterface, authMiddleware gin.HandlerFunc) {
	handler := NewHandler(svc)
	rg.Use(authMiddleware)
	rg.GET("", handler.GetDashboard)
}
