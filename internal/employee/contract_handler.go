package employee

import (
	"fmt"
	"net/http"
	"strconv"
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
	authGroup.Use(authMiddleware)

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
