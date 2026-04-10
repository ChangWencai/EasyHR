package finance

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
)

// VoucherHandler handles HTTP requests for vouchers.
type VoucherHandler struct {
	svc *VoucherService
}

// NewVoucherHandler creates a new VoucherHandler.
func NewVoucherHandler(svc *VoucherService) *VoucherHandler {
	return &VoucherHandler{svc: svc}
}

// RegisterRoutes registers voucher routes within the given router group.
// Routes:
//   GET    /vouchers           - List vouchers (OWNER, ADMIN)
//   GET    /vouchers/:id       - Get voucher detail (OWNER, ADMIN)
//   POST   /vouchers           - Create voucher (OWNER, ADMIN)
//   POST   /vouchers/submit   - Submit voucher (OWNER, ADMIN)
//   POST   /vouchers/audit    - Audit voucher (OWNER only)
//   POST   /vouchers/reverse  - Reverse voucher (OWNER only)
func (h *VoucherHandler) RegisterRoutes(rg *gin.RouterGroup) {
	vouchers := rg.Group("/vouchers")
	{
		// List and detail: OWNER + ADMIN
		vouchers.GET("", middleware.RequireRole("OWNER", "ADMIN"), h.ListVouchers)
		vouchers.GET("/:id", middleware.RequireRole("OWNER", "ADMIN"), h.GetVoucher)

		// Create and submit: OWNER + ADMIN
		vouchers.POST("", middleware.RequireRole("OWNER", "ADMIN"), h.CreateVoucher)
		vouchers.POST("/submit", middleware.RequireRole("OWNER", "ADMIN"), h.SubmitVoucher)

		// Audit and reverse: OWNER only per D-30
		vouchers.POST("/audit", middleware.RequireRole("OWNER"), h.AuditVoucher)
		vouchers.POST("/reverse", middleware.RequireRole("OWNER"), h.ReverseVoucher)
	}
}

// ListVouchers returns paginated vouchers.
// @Summary 查询凭证列表
// @Tags vouchers
// @Security BearerAuth
// @Param period_id query int false "期间ID"
// @Param keyword query string false "搜索关键词"
// @Param page query int false "页码"
// @Param limit query int false "每页数量"
// @Success 200 {object} response.Success
func (h *VoucherHandler) ListVouchers(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	if orgID == 0 {
		response.BadRequest(c, "无效的企业标识")
		return
	}

	var req ListVoucherRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	vouchers, total, err := h.svc.ListVouchers(orgID, &req)
	if err != nil {
		handleFinanceError(c, err)
		return
	}

	response.PageSuccess(c, vouchers, total, req.Page, req.Limit)
}

// GetVoucher returns a voucher with its journal entries.
// @Summary 获取凭证详情
// @Tags vouchers
// @Security BearerAuth
// @Param id path int true "凭证ID"
// @Success 200 {object} response.Success{data=VoucherResponse}
func (h *VoucherHandler) GetVoucher(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	if orgID == 0 {
		response.BadRequest(c, "无效的企业标识")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的凭证ID")
		return
	}

	voucher, err := h.svc.GetVoucher(orgID, id)
	if err != nil {
		handleFinanceError(c, err)
		return
	}

	resp := toVoucherResponse(voucher)
	response.Success(c, resp)
}

// CreateVoucher creates a new voucher.
// @Summary 创建凭证
// @Tags vouchers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body CreateVoucherRequest true "凭证信息"
// @Success 201 {object} response.Success{data=VoucherResponse}
func (h *VoucherHandler) CreateVoucher(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	if orgID == 0 {
		response.BadRequest(c, "无效的企业标识")
		return
	}

	userID := c.GetInt64("user_id")

	var req CreateVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	voucher, err := h.svc.CreateVoucher(orgID, userID, &req)
	if err != nil {
		handleFinanceError(c, err)
		return
	}

	resp := toVoucherResponse(voucher)
	response.Success(c, resp)
}

// SubmitVoucher submits a draft voucher.
// @Summary 提交凭证
// @Tags vouchers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body SubmitVoucherRequest true "凭证ID"
// @Success 200 {object} response.Success
func (h *VoucherHandler) SubmitVoucher(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	if orgID == 0 {
		response.BadRequest(c, "无效的企业标识")
		return
	}

	var req SubmitVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.SubmitVoucher(orgID, req.VoucherID); err != nil {
		handleFinanceError(c, err)
		return
	}

	response.Success(c, gin.H{"voucher_id": req.VoucherID, "status": "submitted"})
}

// AuditVoucher audits a submitted voucher. OWNER only.
// @Summary 审核凭证
// @Tags vouchers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body AuditVoucherRequest true "凭证ID"
// @Success 200 {object} response.Success
func (h *VoucherHandler) AuditVoucher(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	if orgID == 0 {
		response.BadRequest(c, "无效的企业标识")
		return
	}

	var req AuditVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.AuditVoucher(orgID, req.VoucherID); err != nil {
		handleFinanceError(c, err)
		return
	}

	response.Success(c, gin.H{"voucher_id": req.VoucherID, "status": "audited"})
}

// ReverseVoucher reverses an audited voucher. OWNER only.
// @Summary 红冲凭证
// @Tags vouchers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body ReverseVoucherRequest true "凭证ID"
// @Success 200 {object} response.Success{data=VoucherResponse}
func (h *VoucherHandler) ReverseVoucher(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	if orgID == 0 {
		response.BadRequest(c, "无效的企业标识")
		return
	}

	var req ReverseVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	voucher, err := h.svc.ReverseVoucher(orgID, req.VoucherID)
	if err != nil {
		handleFinanceError(c, err)
		return
	}

	resp := toVoucherResponse(voucher)
	response.Success(c, resp)
}

// toVoucherResponse converts a Voucher model to its API response format.
func toVoucherResponse(v *Voucher) VoucherResponse {
	entries := make([]JournalEntryResponse, 0, len(v.Entries))
	for _, e := range v.Entries {
		entries = append(entries, JournalEntryResponse{
			ID:        e.ID,
			AccountID: e.AccountID,
			DC:        string(e.DC),
			Amount:    e.Amount.String(),
			Summary:   e.Summary,
		})
	}
	return VoucherResponse{
		ID:         v.ID,
		VoucherNo:  v.VoucherNo,
		PeriodID:   v.PeriodID,
		Date:       v.Date.Format("2006-01-02"),
		Status:     v.Status,
		SourceType: v.SourceType,
		SourceID:   v.SourceID,
		Summary:    v.Summary,
		ReversalOf: v.ReversalOf,
		Entries:    entries,
	}
}
