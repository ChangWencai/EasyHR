package tax

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
)

// Handler 个税 HTTP 端点
type Handler struct {
	svc *Service
}

// NewHandler 创建个税 Handler
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	authGroup := rg.Group("")
	authGroup.Use(authMiddleware)

	// 税率表 -- 所有角色可查询，OWNER 可初始化
	authGroup.GET("/tax/brackets", h.ListTaxBrackets)
	authGroup.POST("/tax/brackets/seed", middleware.RequireRole("owner"), h.SeedBrackets)

	// 专项附加扣除 -- OWNER/ADMIN 管理，MEMBER 只读
	authGroup.POST("/tax/deductions", middleware.RequireRole("owner", "admin"), h.CreateDeduction)
	authGroup.GET("/tax/deductions", h.ListDeductions)
	authGroup.GET("/tax/deductions/:id", h.GetDeduction)
	authGroup.PUT("/tax/deductions/:id", middleware.RequireRole("owner", "admin"), h.UpdateDeduction)
	authGroup.DELETE("/tax/deductions/:id", middleware.RequireRole("owner", "admin"), h.DeleteDeduction)

	// 个税计算 -- OWNER/ADMIN
	authGroup.POST("/tax/calculate", middleware.RequireRole("owner", "admin"), h.CalculateTax)
	authGroup.POST("/tax/calculate-contract", middleware.RequireRole("owner", "admin"), h.CalculateTaxFromContract)

	// 个税记录查询 -- 所有角色，MEMBER 只看自己
	authGroup.GET("/tax/records", h.ListTaxRecords)
	authGroup.GET("/tax/records/:id", h.GetTaxRecord)
	authGroup.GET("/tax/my-records", h.GetMyTaxRecords)

	// 申报管理 -- OWNER/ADMIN 管理操作
	authGroup.GET("/tax/declarations", h.ListDeclarations)
	authGroup.GET("/tax/declarations/current", h.GetCurrentDeclaration)
	authGroup.PUT("/tax/declarations/:id/declare", middleware.RequireRole("owner", "admin"), h.MarkAsDeclared)

	// 导出 -- OWNER/ADMIN
	authGroup.GET("/tax/declarations/export-excel", middleware.RequireRole("owner", "admin"), h.ExportDeclarationExcel)
	authGroup.GET("/tax/records/:id/export-pdf", middleware.RequireRole("owner", "admin"), h.ExportTaxCertificatePDF)
}

// ========== 税率表端点 ==========

// ListTaxBrackets 查询税率表列表
func (h *Handler) ListTaxBrackets(c *gin.Context) {
	var query TaxBracketListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	year := query.EffectiveYear
	if year == 0 {
		year = time.Now().Year()
	}

	brackets, total, err := h.svc.ListTaxBrackets(year, query.Page, query.PageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 40001, "查询税率表失败")
		return
	}

	response.PageSuccess(c, brackets, total, 1, len(brackets))
}

// SeedBrackets 初始化税率表种子数据
func (h *Handler) SeedBrackets(c *gin.Context) {
	var req struct {
		Year int `json:"year"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.ShouldBindQuery(&req)
	}

	year := req.Year
	if year == 0 {
		year = time.Now().Year()
	}

	if err := h.svc.SeedDefaultBrackets(year); err != nil {
		response.Error(c, http.StatusInternalServerError, 40001, "初始化税率表失败")
		return
	}

	response.Success(c, gin.H{"message": "税率表初始化成功"})
}

// ========== 专项附加扣除端点 ==========

// CreateDeduction 创建专项附加扣除
func (h *Handler) CreateDeduction(c *gin.Context) {
	var req CreateDeductionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	result, err := h.svc.CreateDeduction(orgID, userID, &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}

	response.Success(c, result)
}

// ListDeductions 查询专项附加扣除列表
func (h *Handler) ListDeductions(c *gin.Context) {
	var params DeductionListQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")

	deductions, total, err := h.svc.ListDeductions(orgID, params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 40002, "查询扣除列表失败")
		return
	}

	response.PageSuccess(c, deductions, total, params.Page, params.PageSize)
}

// GetDeduction 获取单个扣除详情
func (h *Handler) GetDeduction(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的扣除ID")
		return
	}

	orgID := c.GetInt64("org_id")

	result, err := h.svc.GetDeduction(orgID, id)
	if err != nil {
		response.Error(c, http.StatusNotFound, 40002, err.Error())
		return
	}

	response.Success(c, result)
}

// UpdateDeduction 更新专项附加扣除
func (h *Handler) UpdateDeduction(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的扣除ID")
		return
	}

	var req UpdateDeductionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	if err := h.svc.UpdateDeduction(orgID, userID, id, &req); err != nil {
		response.Error(c, http.StatusBadRequest, 40002, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "扣除更新成功"})
}

// DeleteDeduction 删除专项附加扣除
func (h *Handler) DeleteDeduction(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的扣除ID")
		return
	}

	orgID := c.GetInt64("org_id")

	if err := h.svc.DeleteDeduction(orgID, id); err != nil {
		response.Error(c, http.StatusBadRequest, 40002, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "扣除删除成功"})
}

// ========== 个税计算端点 ==========

// CalculateTax 计算个税（手动输入收入）
func (h *Handler) CalculateTax(c *gin.Context) {
	var req struct {
		EmployeeID int64   `json:"employee_id" binding:"required"`
		Year       int     `json:"year" binding:"required"`
		Month      int     `json:"month" binding:"required"`
		GrossIncome float64 `json:"gross_income" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")

	result, err := h.svc.CalculateTax(orgID, req.EmployeeID, req.Year, req.Month, req.GrossIncome)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40003, err.Error())
		return
	}

	response.Success(c, result)
}

// CalculateTaxFromContract 计算个税（从合同获取薪资）
func (h *Handler) CalculateTaxFromContract(c *gin.Context) {
	var req struct {
		EmployeeID int64 `json:"employee_id" binding:"required"`
		Year       int   `json:"year" binding:"required"`
		Month      int   `json:"month" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")

	result, err := h.svc.CalculateTaxFromContract(orgID, req.EmployeeID, req.Year, req.Month)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40003, err.Error())
		return
	}

	response.Success(c, result)
}

// ========== 个税记录查询端点 ==========

// ListTaxRecords 查询个税记录列表
func (h *Handler) ListTaxRecords(c *gin.Context) {
	var params TaxRecordListQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")

	records, total, err := h.svc.ListTaxRecords(orgID, params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 40004, "查询个税记录失败")
		return
	}

	response.PageSuccess(c, records, total, params.Page, params.PageSize)
}

// GetTaxRecord 获取个税记录详情
func (h *Handler) GetTaxRecord(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的记录ID")
		return
	}

	orgID := c.GetInt64("org_id")

	result, err := h.svc.GetTaxRecord(orgID, id)
	if err != nil {
		response.Error(c, http.StatusNotFound, 40004, err.Error())
		return
	}

	response.Success(c, result)
}

// GetMyTaxRecords MEMBER 角色查看自己的个税记录
func (h *Handler) GetMyTaxRecords(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	records, err := h.svc.GetMyTaxRecords(orgID, userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, 40004, err.Error())
		return
	}

	response.Success(c, records)
}

// ========== 申报管理端点 ==========

// ListDeclarations 查询申报列表
func (h *Handler) ListDeclarations(c *gin.Context) {
	var query DeclarationListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")

	declarations, total, err := h.svc.ListDeclarations(orgID, query.Year, query.Page, query.PageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 40005, "查询申报列表失败")
		return
	}

	response.PageSuccess(c, declarations, total, query.Page, query.PageSize)
}

// GetCurrentDeclaration 获取或创建当月申报
func (h *Handler) GetCurrentDeclaration(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	now := time.Now()

	decl, err := h.svc.GetOrCreateDeclaration(orgID, now.Year(), int(now.Month()))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 40005, "获取申报记录失败")
		return
	}

	response.Success(c, decl)
}

// MarkAsDeclared 标记为已申报
func (h *Handler) MarkAsDeclared(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的申报ID")
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	if err := h.svc.MarkAsDeclared(orgID, userID, id); err != nil {
		response.Error(c, http.StatusBadRequest, 40005, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "标记申报成功"})
}

// ========== 导出端点 ==========

// ExportDeclarationExcel 导出个税申报表 Excel
func (h *Handler) ExportDeclarationExcel(c *gin.Context) {
	yearStr := c.Query("year")
	monthStr := c.Query("month")

	year, _ := strconv.Atoi(yearStr)
	month, _ := strconv.Atoi(monthStr)

	if year == 0 {
		year = time.Now().Year()
	}
	if month == 0 {
		month = int(time.Now().Month())
	}

	orgID := c.GetInt64("org_id")

	data, err := h.svc.ExportDeclarationExcel(orgID, year, month)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 40006, "导出Excel失败: "+err.Error())
		return
	}

	filename := fmt.Sprintf("tax_declaration_%d_%02d.xlsx", year, month)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// ExportTaxCertificatePDF 导出个税凭证 PDF
func (h *Handler) ExportTaxCertificatePDF(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的记录ID")
		return
	}

	orgID := c.GetInt64("org_id")

	data, err := h.svc.ExportTaxCertificatePDF(orgID, id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 40006, "导出PDF失败: "+err.Error())
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=tax_certificate.pdf")
	c.Data(http.StatusOK, "application/pdf", data)
}
