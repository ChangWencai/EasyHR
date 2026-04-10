package finance

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
)

// AccountHandler handles HTTP requests for accounting accounts.
type AccountHandler struct {
	svc *AccountService
}

// NewAccountHandler creates a new AccountHandler.
func NewAccountHandler(svc *AccountService) *AccountHandler {
	return &AccountHandler{svc: svc}
}

// RegisterRoutes registers account routes within the given router group.
// Routes:
//   GET  /accounts         - Get account tree (OWNER, ADMIN)
//   POST /accounts          - Create custom account (OWNER, ADMIN)
//   PUT  /accounts/:id      - Update account (OWNER, ADMIN)
//
// MEMBER role has no access to finance accounts per D-30.
func (h *AccountHandler) RegisterRoutes(rg *gin.RouterGroup) {
	accounts := rg.Group("/accounts")
	accounts.Use(middleware.RequireRole("OWNER", "ADMIN"))
	{
		accounts.GET("", h.GetTree)
		accounts.POST("", h.CreateAccount)
		accounts.PUT("/:id", h.UpdateAccount)
	}
}

// GetTree returns the full account tree.
// @Summary 获取会计科目树
// @Tags accounts
// @Security BearerAuth
// @Success 200 {object} response.Success{data=[]AccountTreeResponse}
func (h *AccountHandler) GetTree(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	if orgID == 0 {
		response.BadRequest(c, "无效的企业标识")
		return
	}

	// Ensure preset accounts exist
	_ = h.svc.SeedIfEmpty(orgID)

	tree, err := h.svc.GetTree(orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 60000, "获取科目树失败: "+err.Error())
		return
	}

	response.Success(c, tree)
}

// CreateAccount creates a new custom account.
// @Summary 创建自定义会计科目
// @Tags accounts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body CreateAccountRequest true "科目信息"
// @Success 201 {object} response.Success{data=Account}
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	if orgID == 0 {
		response.BadRequest(c, "无效的企业标识")
		return
	}

	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	account, err := h.svc.CreateCustomAccount(orgID, &req)
	if err != nil {
		handleFinanceError(c, err)
		return
	}

	response.Success(c, account)
}

// UpdateAccount updates an existing account.
// @Summary 更新会计科目
// @Tags accounts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "科目ID"
// @Param body body UpdateAccountRequest true "更新信息"
// @Success 200 {object} response.Success{data=Account}
func (h *AccountHandler) UpdateAccount(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	if orgID == 0 {
		response.BadRequest(c, "无效的企业标识")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的科目ID")
		return
	}

	var req UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	account, err := h.svc.UpdateAccount(orgID, id, &req)
	if err != nil {
		handleFinanceError(c, err)
		return
	}

	response.Success(c, account)
}

// handleFinanceError maps FinanceError to appropriate HTTP responses.
func handleFinanceError(c *gin.Context, err error) {
	if fe, ok := err.(*FinanceError); ok {
		response.Error(c, http.StatusBadRequest, fe.Code, fe.Error())
		return
	}
	response.Error(c, http.StatusInternalServerError, 60000, err.Error())
}
