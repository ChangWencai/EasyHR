package finance

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
)

// InvoiceHandler handles HTTP requests for invoices.
type InvoiceHandler struct {
	invoiceSvc *InvoiceService
	voucherSvc *VoucherService
}

// NewInvoiceHandler creates a new InvoiceHandler.
func NewInvoiceHandler(invoiceSvc *InvoiceService, voucherSvc *VoucherService) *InvoiceHandler {
	return &InvoiceHandler{
		invoiceSvc: invoiceSvc,
		voucherSvc: voucherSvc,
	}
}

// RegisterRoutes registers invoice routes.
func (h *InvoiceHandler) RegisterRoutes(rg *gin.RouterGroup) {
	invoices := rg.Group("/invoices")
	{
		// List invoices — OWNER+ADMIN
		invoices.GET("", middleware.RequireRole("OWNER", "ADMIN"), h.ListInvoices)
		// Create invoice — OWNER+ADMIN
		invoices.POST("", middleware.RequireRole("OWNER", "ADMIN"), h.CreateInvoice)
		// Get invoice detail — OWNER+ADMIN
		invoices.GET("/:id", middleware.RequireRole("OWNER", "ADMIN"), h.GetInvoice)
		// Update invoice — OWNER+ADMIN
		invoices.PUT("/:id", middleware.RequireRole("OWNER", "ADMIN"), h.UpdateInvoice)
		// Link invoice to voucher — OWNER+ADMIN
		invoices.POST("/:id/link-voucher", middleware.RequireRole("OWNER", "ADMIN"), h.LinkToVoucher)
		// Monthly tax summary — OWNER+ADMIN
		invoices.GET("/monthly-summary", middleware.RequireRole("OWNER", "ADMIN"), h.GetMonthlySummary)
	}
}

// ListInvoices lists invoices with optional filters.
func (h *InvoiceHandler) ListInvoices(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	if orgID == 0 {
		response.BadRequest(c, "无效的企业标识")
		return
	}

	var req ListInvoiceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	invoices, total, err := h.invoiceSvc.ListInvoices(orgID, &req)
	if err != nil {
		handleFinanceError(c, err)
		return
	}

	items := make([]InvoiceResponse, len(invoices))
	for i, inv := range invoices {
		items[i] = toInvoiceResponse(&inv, "")
	}

	response.PageSuccess(c, items, total, req.Page, req.Limit)
}

// CreateInvoice creates a new invoice.
func (h *InvoiceHandler) CreateInvoice(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	if orgID == 0 {
		response.BadRequest(c, "无效的企业标识")
		return
	}
	userID := c.GetInt64("user_id")

	var req CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	invoice, err := h.invoiceSvc.CreateInvoice(orgID, userID, &req)
	if err != nil {
		handleFinanceError(c, err)
		return
	}

	response.Success(c, toInvoiceResponse(invoice, ""))
}

// GetInvoice returns an invoice by ID.
func (h *InvoiceHandler) GetInvoice(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	if orgID == 0 {
		response.BadRequest(c, "无效的企业标识")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的发票ID")
		return
	}

	invoice, err := h.invoiceSvc.GetInvoice(orgID, id)
	if err != nil {
		handleFinanceError(c, err)
		return
	}

	var voucherNo string
	if invoice.VoucherID != nil && *invoice.VoucherID > 0 {
		if v, err := h.voucherSvc.GetVoucher(orgID, *invoice.VoucherID); err == nil {
			voucherNo = v.VoucherNo
		}
	}

	response.Success(c, toInvoiceResponse(invoice, voucherNo))
}

// UpdateInvoice updates an existing invoice.
func (h *InvoiceHandler) UpdateInvoice(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	if orgID == 0 {
		response.BadRequest(c, "无效的企业标识")
		return
	}
	userID := c.GetInt64("user_id")

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的发票ID")
		return
	}

	var req UpdateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	invoice, err := h.invoiceSvc.UpdateInvoice(orgID, userID, id, &req)
	if err != nil {
		handleFinanceError(c, err)
		return
	}

	response.Success(c, toInvoiceResponse(invoice, ""))
}

// LinkToVoucher links an invoice to a voucher.
func (h *InvoiceHandler) LinkToVoucher(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	if orgID == 0 {
		response.BadRequest(c, "无效的企业标识")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的发票ID")
		return
	}

	var req LinkVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// Verify voucher exists and belongs to org
	if _, err := h.voucherSvc.GetVoucher(orgID, req.VoucherID); err != nil {
		handleFinanceError(c, err)
		return
	}

	if err := h.invoiceSvc.LinkToVoucher(orgID, id, req.VoucherID); err != nil {
		handleFinanceError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "发票关联凭证成功"})
}

// GetMonthlySummary returns monthly VAT summary.
func (h *InvoiceHandler) GetMonthlySummary(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	if orgID == 0 {
		response.BadRequest(c, "无效的企业标识")
		return
	}

	yearStr := c.Query("year")
	monthStr := c.Query("month")
	if yearStr == "" || monthStr == "" {
		response.BadRequest(c, "year 和 month 参数必须提供")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 {
		response.BadRequest(c, "无效的 year 参数")
		return
	}
	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		response.BadRequest(c, "无效的 month 参数")
		return
	}

	summary, err := h.invoiceSvc.GetMonthlySummary(orgID, year, month)
	if err != nil {
		handleFinanceError(c, err)
		return
	}

	response.Success(c, summary)
}

// toInvoiceResponse converts an Invoice to InvoiceResponse.
func toInvoiceResponse(inv *Invoice, voucherNo string) InvoiceResponse {
	dateStr := ""
	if !inv.Date.IsZero() {
		dateStr = inv.Date.Format("2006-01-02")
	}
	return InvoiceResponse{
		ID:          inv.ID,
		InvoiceType: inv.InvoiceType,
		Code:        inv.Code,
		Number:      inv.Number,
		Date:        dateStr,
		Amount:      inv.Amount.String(),
		TaxRate:     inv.TaxRate.String(),
		TaxAmount:   inv.TaxAmount.String(),
		Status:      inv.Status,
		VoucherID:   inv.VoucherID,
		VoucherNo:   voucherNo,
		Remark:      inv.Remark,
		CreatedAt:   inv.CreatedAt,
	}
}
