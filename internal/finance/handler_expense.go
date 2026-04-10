package finance

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
)

// ExpenseHandler handles HTTP requests for expense reimbursements.
type ExpenseHandler struct {
	expenseSvc *ExpenseService
	voucherSvc *VoucherService
}

// NewExpenseHandler creates a new ExpenseHandler.
func NewExpenseHandler(expenseSvc *ExpenseService, voucherSvc *VoucherService) *ExpenseHandler {
	return &ExpenseHandler{
		expenseSvc: expenseSvc,
		voucherSvc: voucherSvc,
	}
}

// RegisterRoutes registers expense reimbursement routes.
func (h *ExpenseHandler) RegisterRoutes(rg *gin.RouterGroup) {
	expenses := rg.Group("/expenses")
	{
		// Submit expense — OWNER+ADMIN+MEMBER
		expenses.POST("", middleware.RequireRole("OWNER", "ADMIN", "MEMBER"), h.SubmitExpense)
		// List expenses — OWNER+ADMIN
		expenses.GET("", middleware.RequireRole("OWNER", "ADMIN"), h.ListExpenses)
		// Get expense detail — OWNER+ADMIN
		expenses.GET("/:id", middleware.RequireRole("OWNER", "ADMIN"), h.GetExpense)
		// Approve expense — OWNER+ADMIN
		expenses.POST("/:id/approve", middleware.RequireRole("OWNER", "ADMIN"), h.ApproveExpense)
		// Reject expense — OWNER+ADMIN
		expenses.POST("/:id/reject", middleware.RequireRole("OWNER", "ADMIN"), h.RejectExpense)
		// Mark as paid — OWNER+ADMIN
		expenses.POST("/:id/mark-paid", middleware.RequireRole("OWNER", "ADMIN"), h.MarkPaid)
	}
}

// SubmitExpense creates and submits a new expense reimbursement (MEMBER can submit).
func (h *ExpenseHandler) SubmitExpense(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	if orgID == 0 {
		response.BadRequest(c, "无效的企业标识")
		return
	}
	userID := c.GetInt64("user_id")

	var req CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// For MEMBER role: employee_id comes from the user's employee record
	// For OWNER/ADMIN: they can submit for any employee_id
	role := c.GetString("role")
	if role == "MEMBER" {
		// MEMBER can only submit for themselves (employee_id from request must match their employee record)
		// In Phase 8, we'll look up employee_id from user-employee mapping
		// For now, accept the employee_id from request
	}

	expense, err := h.expenseSvc.CreateExpense(orgID, userID, &req)
	if err != nil {
		handleFinanceError(c, err)
		return
	}

	response.Success(c, toExpenseResponse(expense, ""))
}

// ListExpenses lists expense reimbursements with optional filters (OWNER+ADMIN).
func (h *ExpenseHandler) ListExpenses(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	if orgID == 0 {
		response.BadRequest(c, "无效的企业标识")
		return
	}

	var req ListExpenseRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	expenses, total, err := h.expenseSvc.ListExpenses(orgID, &req)
	if err != nil {
		handleFinanceError(c, err)
		return
	}

	items := make([]ExpenseResponse, len(expenses))
	for i, exp := range expenses {
		items[i] = toExpenseResponse(&exp, "")
	}

	response.PageSuccess(c, items, total, req.Page, req.Limit)
}

// GetExpense returns an expense by ID (OWNER+ADMIN).
func (h *ExpenseHandler) GetExpense(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	if orgID == 0 {
		response.BadRequest(c, "无效的企业标识")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的报销单ID")
		return
	}

	expense, err := h.expenseSvc.GetExpense(orgID, id)
	if err != nil {
		handleFinanceError(c, err)
		return
	}

	var voucherNo string
	if expense.VoucherID != nil && *expense.VoucherID > 0 {
		if v, err := h.voucherSvc.GetVoucher(orgID, *expense.VoucherID); err == nil {
			voucherNo = v.VoucherNo
		}
	}

	response.Success(c, toExpenseResponse(expense, voucherNo))
}

// ApproveExpense approves a pending expense and auto-generates an expense voucher (OWNER+ADMIN).
func (h *ExpenseHandler) ApproveExpense(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	if orgID == 0 {
		response.BadRequest(c, "无效的企业标识")
		return
	}
	userID := c.GetInt64("user_id")

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的报销单ID")
		return
	}

	var req ApproveExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body
		req = ApproveExpenseRequest{}
	}

	expense, voucher, err := h.expenseSvc.ApproveExpense(orgID, userID, id, req.Note)
	if err != nil {
		handleFinanceError(c, err)
		return
	}

	response.Success(c, gin.H{
		"expense":   toExpenseResponse(expense, voucher.VoucherNo),
		"voucher":   toVoucherResponse(voucher),
		"message":   "报销单审批通过，已生成费用凭证",
	})
}

// RejectExpense rejects a pending expense reimbursement (OWNER+ADMIN).
func (h *ExpenseHandler) RejectExpense(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	if orgID == 0 {
		response.BadRequest(c, "无效的企业标识")
		return
	}
	userID := c.GetInt64("user_id")

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的报销单ID")
		return
	}

	var req RejectExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "驳回原因不能为空")
		return
	}

	expense, err := h.expenseSvc.RejectExpense(orgID, userID, id, req.Note)
	if err != nil {
		handleFinanceError(c, err)
		return
	}

	response.Success(c, toExpenseResponse(expense, ""))
}

// MarkPaid marks an approved expense as paid and generates a payment voucher (OWNER+ADMIN).
func (h *ExpenseHandler) MarkPaid(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	if orgID == 0 {
		response.BadRequest(c, "无效的企业标识")
		return
	}
	userID := c.GetInt64("user_id")

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的报销单ID")
		return
	}

	var req MarkPaidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req = MarkPaidRequest{}
	}

	expense, voucher, err := h.expenseSvc.MarkExpensePaid(orgID, userID, id, req.Note)
	if err != nil {
		handleFinanceError(c, err)
		return
	}

	response.Success(c, gin.H{
		"expense": toExpenseResponse(expense, voucher.VoucherNo),
		"voucher": toVoucherResponse(voucher),
		"message": "报销单已标记为已支付",
	})
}

// toExpenseResponse converts an ExpenseReimbursement to ExpenseResponse.
func toExpenseResponse(exp *ExpenseReimbursement, voucherNo string) ExpenseResponse {
	var approvedAt, rejectedAt, paidAt *string
	if exp.ApprovedAt != nil {
		s := exp.ApprovedAt.Format("2006-01-02 15:04:05")
		approvedAt = &s
	}
	if exp.RejectedAt != nil {
		s := exp.RejectedAt.Format("2006-01-02 15:04:05")
		rejectedAt = &s
	}
	if exp.PaidAt != nil {
		s := exp.PaidAt.Format("2006-01-02 15:04:05")
		paidAt = &s
	}

	return ExpenseResponse{
		ID:           exp.ID,
		EmployeeID:   exp.EmployeeID,
		Amount:       exp.Amount.String(),
		ExpenseType:  exp.ExpenseType,
		Description:  exp.Description,
		Attachments:  exp.AttachmentURLs(),
		Status:       exp.Status,
		ApproverID:   exp.ApproverID,
		ApprovedAt:   approvedAt,
		ApprovedNote: exp.ApprovedNote,
		RejectedAt:   rejectedAt,
		RejectedNote: exp.RejectedNote,
		PaidAt:       paidAt,
		PaidNote:     exp.PaidNote,
		VoucherID:    exp.VoucherID,
		VoucherNo:    voucherNo,
		CreatedAt:    exp.CreatedAt,
	}
}
