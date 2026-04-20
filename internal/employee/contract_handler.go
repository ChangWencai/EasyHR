package employee

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
)

// ContractHandler 合同 HTTP 端点
type ContractHandler struct {
	svc *ContractService
}

// NewContractHandler 创建合同 Handler
func NewContractHandler(svc *ContractService) *ContractHandler {
	return &ContractHandler{svc: svc}
}

// RegisterRoutes 注册合同路由
func (h *ContractHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	authGroup := rg.Group("")
	authGroup.Use(authMiddleware, middleware.RequireOrg)

	// 合同 CRUD -- OWNER/ADMIN 可创建/编辑/终止
	authGroup.POST("/employees/:id/contracts", middleware.RequireRole("owner", "admin"), h.CreateContract)
	authGroup.GET("/employees/:id/contracts", h.ListByEmployee) // 所有角色可查看
	authGroup.GET("/contracts/:id", h.GetContract)              // 所有角色可查看
	authGroup.PUT("/contracts/:id", middleware.RequireRole("owner", "admin"), h.UpdateContract)
	authGroup.POST("/contracts/:id/generate-pdf", middleware.RequireRole("owner", "admin"), h.GeneratePDF)
	authGroup.POST("/contracts/:id/upload-signed", middleware.RequireRole("owner", "admin"), h.UploadSigned)
	authGroup.GET("/contracts/:id/upload-url", middleware.RequireRole("owner", "admin"), h.GenerateUploadURL)
	authGroup.PUT("/contracts/:id/terminate", middleware.RequireRole("owner", "admin"), h.TerminateContract)
	authGroup.GET("/contracts", h.ListContracts) // 所有角色可查看企业合同列表
	authGroup.POST("/contracts/:id/send-sign-link", middleware.RequireRole("owner", "admin"), h.SendSignLink) // 老板发起签署
}

// RegisterSignRoutes 注册签署相关端点（员工端，无需认证）
func (h *ContractHandler) RegisterSignRoutes(rg *gin.RouterGroup) {
	signGroup := rg.Group("/contracts/sign")
	signGroup.POST("/send-code", h.SendSignCode)      // 发送验证码
	signGroup.POST("/verify-code", h.VerifySignCode) // 校验验证码
	signGroup.POST("/confirm", h.ConfirmSign)         // 确认签署
}

// CreateContract 创建合同
func (h *ContractHandler) CreateContract(c *gin.Context) {
	employeeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的员工ID")
		return
	}

	var req CreateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	result, err := h.svc.CreateContract(c.Request.Context(), orgID, userID, employeeID, &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 20200, err.Error())
		return
	}

	response.Success(c, result)
}

// ListByEmployee 按员工查询合同列表
func (h *ContractHandler) ListByEmployee(c *gin.Context) {
	employeeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的员工ID")
		return
	}

	var query ContractListQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	contracts, total, err := h.svc.ListByEmployee(c.Request.Context(), orgID, employeeID, query.Page, query.PageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 20201, "查询合同列表失败")
		return
	}

	response.PageSuccess(c, contracts, total, query.Page, query.PageSize)
}

// GetContract 获取合同详情
func (h *ContractHandler) GetContract(c *gin.Context) {
	contractID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的合同ID")
		return
	}

	orgID := c.GetInt64("org_id")
	result, err := h.svc.GetContract(c.Request.Context(), orgID, contractID)
	if err != nil {
		response.Error(c, http.StatusNotFound, 20202, err.Error())
		return
	}

	response.Success(c, result)
}

// UpdateContract 更新合同（仅草稿状态）
func (h *ContractHandler) UpdateContract(c *gin.Context) {
	contractID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的合同ID")
		return
	}

	var req UpdateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	result, err := h.svc.UpdateContract(c.Request.Context(), orgID, userID, contractID, &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 20203, err.Error())
		return
	}

	response.Success(c, result)
}

// GeneratePDF 生成合同 PDF 模板
func (h *ContractHandler) GeneratePDF(c *gin.Context) {
	contractID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的合同ID")
		return
	}

	orgID := c.GetInt64("org_id")
	pdfBytes, err := h.svc.GeneratePDF(c.Request.Context(), orgID, contractID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 20204, err.Error())
		return
	}

	filename := fmt.Sprintf("contract_%d_%s.pdf", contractID, time.Now().Format("20060102150405"))
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// UploadSigned 上传签署扫描件
func (h *ContractHandler) UploadSigned(c *gin.Context) {
	contractID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的合同ID")
		return
	}

	var req UploadSignedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	result, err := h.svc.UploadSigned(c.Request.Context(), orgID, contractID, &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 20205, err.Error())
		return
	}

	response.Success(c, result)
}

// GenerateUploadURL 生成 OSS 签名上传 URL（客户端直传后，再调 UploadSigned 传入 URL）
// 这个功能需要 OSS 客户端，在 contract_handler 中调用 OSS 生成签名 URL
func (h *ContractHandler) GenerateUploadURL(c *gin.Context) {
	contractID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的合同ID")
		return
	}

	orgID := c.GetInt64("org_id")

	// 检查合同存在
	_, err = h.svc.GetContract(c.Request.Context(), orgID, contractID)
	if err != nil {
		response.Error(c, http.StatusNotFound, 20206, err.Error())
		return
	}

	// TODO: 需要在 ContractService 中注入 OSS Client 来生成签名 URL
	// V1.0 先返回提示信息，OSS 上传通过前端直传后调用 UploadSigned 接口
	response.Success(c, gin.H{
		"message":     "请通过 OSS SDK 直接上传签署扫描件，然后调用 UploadSigned 接口传入 URL",
		"contract_id": contractID,
	})
}

// TerminateContract 终止合同
func (h *ContractHandler) TerminateContract(c *gin.Context) {
	contractID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的合同ID")
		return
	}

	var req TerminateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	result, err := h.svc.TerminateContract(c.Request.Context(), orgID, contractID, &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 20207, err.Error())
		return
	}

	response.Success(c, result)
}

// ListContracts 查询企业所有合同
func (h *ContractHandler) ListContracts(c *gin.Context) {
	var query ContractListQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	contracts, total, err := h.svc.ListContracts(c.Request.Context(), orgID, query.Status, query.Page, query.PageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 20208, "查询合同列表失败")
		return
	}

	response.PageSuccess(c, contracts, total, query.Page, query.PageSize)
}

// SendSignCode 发送签署验证码
func (h *ContractHandler) SendSignCode(c *gin.Context) {
	var req SendSignCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.SendSignCode(c.Request.Context(), req.ContractID, req.Phone); err != nil {
		response.Error(c, http.StatusBadRequest, 20212, err.Error())
		return
	}

	response.Success(c, &SendSignCodeResponse{
		Message:   fmt.Sprintf("签署链接已发送至 %s，有效期7天", req.Phone),
		ExpiresIn: int(SignLinkExpiry.Seconds()),
	})
}

// VerifySignCode 校验签署验证码
func (h *ContractHandler) VerifySignCode(c *gin.Context) {
	var req VerifySignCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.svc.VerifySignCode(c.Request.Context(), req.ContractID, req.Phone, req.Code)
	if err != nil {
		if strings.Contains(err.Error(), "过期") {
			response.Error(c, http.StatusBadRequest, 20213, "验证码已过期，请重新获取")
		} else {
			response.Error(c, http.StatusBadRequest, 20214, "验证码错误，请重新输入")
		}
		return
	}

	response.Success(c, result)
}

// ConfirmSign 确认签署
func (h *ContractHandler) ConfirmSign(c *gin.Context) {
	var req ConfirmSignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.svc.ConfirmSign(c.Request.Context(), req.ContractID, req.SignToken)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 20215, err.Error())
		return
	}

	response.Success(c, result)
}

// SendSignLink 老板发起签署（生成PDF + 上传OSS + 发送短信）
func (h *ContractHandler) SendSignLink(c *gin.Context) {
	contractID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的合同ID")
		return
	}

	orgID := c.GetInt64("org_id")
	if err := h.svc.SendSignLink(c.Request.Context(), orgID, contractID); err != nil {
		response.Error(c, http.StatusBadRequest, 20217, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "签署链接发送成功"})
}

// GetSignedPdf 获取已签PDF（员工端，通过 SignToken 访问）
func (h *ContractHandler) GetSignedPdf(c *gin.Context) {
	contractID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的合同ID")
		return
	}

	url, err := h.svc.GetSignedPdfURL(c.Request.Context(), contractID)
	if err != nil {
		response.Error(c, http.StatusNotFound, 20216, "合同不存在或尚未签署")
		return
	}

	response.Success(c, &GetSignedPdfResponse{URL: url})
}
