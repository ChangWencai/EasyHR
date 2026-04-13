package wxmp

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/response"
	"github.com/wencai/easyhr/internal/user"
)

// Handler HTTP 处理器
type Handler struct {
	svc     *WXMPService
	userSvc *user.Service
}

// NewHandler 创建 Handler
func NewHandler(svc *WXMPService, userSvc *user.Service) *Handler {
	return &Handler{svc: svc, userSvc: userSvc}
}

// RegisterRoutes 注册所有 wxmp 路由
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	// 认证路由（无需登录）
	auth := rg.Group("/auth")
	{
		auth.POST("/send-code", h.SendCode)
		auth.POST("/login", h.Login)
		auth.POST("/wechat/bind", h.BindWechat)
	}

	// 会员路由（需要 MEMBER JWT）
	member := rg.Group("")
	member.Use(WXMPMemberAuth(nil)) // jwtSecret 由 router 注入
	{
		// 工资条
		member.GET("/payslips", h.ListPayslips)
		member.POST("/payslips/:id/verify", h.VerifyPayslip)
		member.GET("/payslips/:id", h.GetPayslipDetail)
		member.POST("/payslips/:id/sign", h.SignPayslip)

		// 合同
		member.GET("/contracts", h.ListContracts)
		member.GET("/contracts/:id/pdf", h.GetContractPDF)

		// 社保
		member.GET("/social-insurance", h.GetSocialInsurance)

		// 报销
		member.POST("/expenses", h.CreateExpense)
		member.GET("/expenses", h.ListExpenses)
		member.GET("/expenses/:id", h.GetExpenseDetail)

		// OSS
		member.POST("/oss/upload-url", h.GetOssUploadURL)
	}
}

// Login POST /wxmp/auth/login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	resp, err := h.svc.LoginMember(c.Request.Context(), req.Phone, req.Code)
	if err != nil {
		// 区分错误类型
		errMsg := err.Error()
		if errors.Is(err, errors.New("验证码已过期")) ||
			errors.Is(err, errors.New("验证码错误")) ||
			errors.Is(err, errors.New("该手机号未关联员工账号")) {
			response.Error(c, http.StatusBadRequest, 40001, errMsg)
			return
		}
		response.Error(c, http.StatusInternalServerError, 50000, errMsg)
		return
	}
	response.Success(c, resp)
}

// SendCode POST /wxmp/auth/send-code
func (h *Handler) SendCode(c *gin.Context) {
	var req user.SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if err := h.userSvc.SendCode(c.Request.Context(), req.Phone); err != nil {
		errMsg := err.Error()
		// 限流等业务错误返回 429
		if errMsg == "发送过于频繁，请稍后再试" || errMsg == "短信发送失败，请稍后再试" {
			response.Error(c, http.StatusTooManyRequests, 10002, errMsg)
			return
		}
		response.Error(c, http.StatusInternalServerError, 90001, errMsg)
		return
	}
	response.Success(c, gin.H{"message": "验证码已发送"})
}

// BindWechat POST /wxmp/auth/wechat/bind
func (h *Handler) BindWechat(c *gin.Context) {
	// TODO: 实现微信授权 code 换取 openid 的逻辑
	// Phase 8 后续迭代处理
	response.Error(c, http.StatusNotImplemented, 50101, "微信绑定功能开发中")
}

// ListPayslips GET /wxmp/payslips
func (h *Handler) ListPayslips(c *gin.Context) {
	employeeID := c.GetUint("employee_id")
	orgID := c.GetUint("org_id")

	payslips, err := h.svc.GetPayslips(c.Request.Context(), orgID, employeeID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	if payslips == nil {
		payslips = []PayslipSummary{}
	}
	response.Success(c, payslips)
}

// VerifyPayslip POST /wxmp/payslips/:id/verify
func (h *Handler) VerifyPayslip(c *gin.Context) {
	payslipID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid payslip id")
		return
	}
	employeeID := c.GetUint("employee_id")
	orgID := c.GetUint("org_id")

	var req VerifyPayslipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 获取用户手机号用于验证（从 context 或再次查询）
	// 小程序端需要传手机号，我们这里简化处理：从 JWT 中获取或要求前端传
	phone := c.GetHeader("X-WXMP-Phone")
	if phone == "" {
		response.BadRequest(c, "缺少手机号")
		return
	}

	resp, err := h.svc.VerifyPayslipAccess(c.Request.Context(), orgID, employeeID, uint(payslipID), req.Code, phone)
	if err != nil {
		errMsg := err.Error()
		if errors.Is(err, errors.New("验证码已过期")) ||
			errors.Is(err, errors.New("验证码错误")) {
			response.Error(c, http.StatusBadRequest, 40001, errMsg)
			return
		}
		response.Error(c, http.StatusInternalServerError, 50000, errMsg)
		return
	}
	response.Success(c, resp)
}

// GetPayslipDetail GET /wxmp/payslips/:id
func (h *Handler) GetPayslipDetail(c *gin.Context) {
	payslipID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid payslip id")
		return
	}
	employeeID := c.GetUint("employee_id")
	orgID := c.GetUint("org_id")
	verifyToken := c.GetHeader("X-Verify-Token")

	detail, err := h.svc.GetPayslipDetail(c.Request.Context(), orgID, employeeID, uint(payslipID), verifyToken)
	if err != nil {
		errMsg := err.Error()
		if errors.Is(err, errors.New("请先验证身份")) ||
			errors.Is(err, errors.New("验证已过期，请重新验证")) {
			response.Error(c, http.StatusForbidden, 40301, "请先验证身份")
			return
		}
		response.Error(c, http.StatusInternalServerError, 50000, errMsg)
		return
	}
	response.Success(c, detail)
}

// SignPayslip POST /wxmp/payslips/:id/sign
func (h *Handler) SignPayslip(c *gin.Context) {
	payslipID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid payslip id")
		return
	}
	employeeID := c.GetUint("employee_id")
	orgID := c.GetUint("org_id")

	if err := h.svc.SignPayslip(c.Request.Context(), orgID, employeeID, uint(payslipID)); err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "签收成功"})
}

// ListContracts GET /wxmp/contracts
func (h *Handler) ListContracts(c *gin.Context) {
	employeeID := c.GetUint("employee_id")
	orgID := c.GetUint("org_id")

	contracts, err := h.svc.GetContracts(c.Request.Context(), orgID, employeeID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	if contracts == nil {
		contracts = []ContractDTO{}
	}
	response.Success(c, contracts)
}

// GetContractPDF GET /wxmp/contracts/:id/pdf
func (h *Handler) GetContractPDF(c *gin.Context) {
	contractID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid contract id")
		return
	}
	employeeID := c.GetUint("employee_id")
	orgID := c.GetUint("org_id")

	pdfURL, err := h.svc.GetContractPDF(c.Request.Context(), orgID, employeeID, uint(contractID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	response.Success(c, gin.H{"pdf_url": pdfURL})
}

// GetSocialInsurance GET /wxmp/social-insurance
func (h *Handler) GetSocialInsurance(c *gin.Context) {
	employeeID := c.GetUint("employee_id")
	orgID := c.GetUint("org_id")

	records, err := h.svc.GetSocialInsurance(c.Request.Context(), orgID, employeeID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	if records == nil {
		records = []SocialInsuranceDTO{}
	}
	response.Success(c, records)
}

// CreateExpense POST /wxmp/expenses
func (h *Handler) CreateExpense(c *gin.Context) {
	employeeID := c.GetUint("employee_id")
	orgID := c.GetUint("org_id")

	var req ExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	exp, err := h.svc.CreateExpense(c.Request.Context(), orgID, employeeID, &req)
	if err != nil {
		errMsg := err.Error()
		if err.Error() == "attachments exceed limit of 9" {
			response.Error(c, http.StatusBadRequest, 40002, "附件最多9张")
			return
		}
		response.Error(c, http.StatusInternalServerError, 50000, errMsg)
		return
	}
	response.Success(c, exp)
}

// ListExpenses GET /wxmp/expenses
func (h *Handler) ListExpenses(c *gin.Context) {
	employeeID := c.GetUint("employee_id")
	orgID := c.GetUint("org_id")

	expenses, err := h.svc.GetExpenses(c.Request.Context(), orgID, employeeID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	if expenses == nil {
		expenses = []ExpenseDTO{}
	}
	response.Success(c, expenses)
}

// GetExpenseDetail GET /wxmp/expenses/:id
func (h *Handler) GetExpenseDetail(c *gin.Context) {
	expenseID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid expense id")
		return
	}
	employeeID := c.GetUint("employee_id")
	orgID := c.GetUint("org_id")

	expense, err := h.svc.GetExpenseDetail(c.Request.Context(), orgID, employeeID, uint(expenseID))
	if err != nil {
		if err.Error() == "expense not found" {
			response.Error(c, http.StatusNotFound, 40401, "报销单不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	response.Success(c, expense)
}

// GetOssUploadURL POST /wxmp/oss/upload-url
func (h *Handler) GetOssUploadURL(c *gin.Context) {
	// 此功能需要 OSS client，Phase 8 前端部分实现时处理
	// 后端仅返回占位响应
	response.Success(c, gin.H{
		"upload_url": "",
		"file_key":   "",
		"message":    "OSS上传功能开发中",
	})
}
