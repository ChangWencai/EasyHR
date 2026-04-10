package finance

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ReportHandler handles financial report and period management routes.
type ReportHandler struct {
	reportSvc *ReportService
	periodSvc *PeriodService
}

// NewReportHandler creates a new ReportHandler.
func NewReportHandler(reportSvc *ReportService, periodSvc *PeriodService) *ReportHandler {
	return &ReportHandler{
		reportSvc: reportSvc,
		periodSvc: periodSvc,
	}
}

// RegisterRoutes registers the report and period routes.
func (h *ReportHandler) RegisterRoutes(rg *gin.RouterGroup) {
	reports := rg.Group("/reports")
	{
		reports.GET("/balance-sheet", h.GetBalanceSheet)
		reports.GET("/income-statement", h.GetIncomeStatement)
		reports.GET("/multi-period", h.GetMultiPeriodBalanceSheet)
		reports.GET("/vat", h.CalculateVAT)
		reports.GET("/cit", h.CalculateCIT)
		reports.GET("/tax-declaration/export", h.ExportTaxDeclaration)
	}

	periods := rg.Group("/periods")
	{
		periods.GET("", h.ListPeriods)
		periods.POST("/validate", h.ValidateClosing)
		periods.POST("/:id/close", h.ClosePeriod)
		periods.POST("/:id/revert", h.RevertClosing)
	}
}

// GetBalanceSheet returns the balance sheet for a period.
// GET /reports/balance-sheet?period_id=1
func (h *ReportHandler) GetBalanceSheet(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	periodID, err := strconv.ParseInt(c.Query("period_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "period_id is required"})
		return
	}

	result, err := h.reportSvc.GetBalanceSheet(c.Request.Context(), orgID, periodID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetIncomeStatement returns the income statement for a period.
// GET /reports/income-statement?period_id=1
func (h *ReportHandler) GetIncomeStatement(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	periodID, err := strconv.ParseInt(c.Query("period_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "period_id is required"})
		return
	}

	result, err := h.reportSvc.GetIncomeStatement(c.Request.Context(), orgID, periodID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetMultiPeriodBalanceSheet returns multi-period balance sheet comparison.
// GET /reports/multi-period?period_ids=1,2,3
func (h *ReportHandler) GetMultiPeriodBalanceSheet(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	idsParam := c.Query("period_ids")
	if idsParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "period_ids is required (comma-separated)"})
		return
	}

	var periodIDs []int64
	for _, s := range splitAndTrim(idsParam, ",") {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period_ids format"})
			return
		}
		periodIDs = append(periodIDs, id)
	}

	result, err := h.reportSvc.GetMultiPeriodBalanceSheet(c.Request.Context(), orgID, periodIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// CalculateVAT returns monthly VAT calculation.
// GET /reports/vat?year=2026&month=4
func (h *ReportHandler) CalculateVAT(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	year, err := strconv.Atoi(c.Query("year"))
	if err != nil || year < 2000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid year is required"})
		return
	}
	month, err := strconv.Atoi(c.Query("month"))
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid month (1-12) is required"})
		return
	}

	result, err := h.reportSvc.CalculateVAT(c.Request.Context(), orgID, year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// CalculateCIT returns quarterly CIT estimate.
// GET /reports/cit?year=2026&quarter=1
func (h *ReportHandler) CalculateCIT(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	year, err := strconv.Atoi(c.Query("year"))
	if err != nil || year < 2000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid year is required"})
		return
	}
	quarter, err := strconv.Atoi(c.Query("quarter"))
	if err != nil || quarter < 1 || quarter > 4 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid quarter (1-4) is required"})
		return
	}

	result, err := h.reportSvc.CalculateCIT(c.Request.Context(), orgID, year, quarter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ExportTaxDeclaration stub — returns placeholder JSON (Excel export V2.0).
func (h *ReportHandler) ExportTaxDeclaration(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	year, _ := strconv.Atoi(c.Query("year"))
	month, _ := strconv.Atoi(c.Query("month"))

	c.JSON(http.StatusOK, gin.H{
		"message":    "Excel export V2.0",
		"year":       year,
		"month":      month,
		"org_id":     orgID,
		"note":       "Excel export using excelize - to be implemented in V2.0",
	})
}

// ListPeriods returns all periods for the org.
// GET /periods
func (h *ReportHandler) ListPeriods(c *gin.Context) {
	orgID := c.GetInt64("org_id")

	result, err := h.periodSvc.GetPeriods(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ValidateClosing validates whether a period can be closed.
// POST /periods/validate
func (h *ReportHandler) ValidateClosing(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	var req ClosePeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.periodSvc.ValidateClosing(c.Request.Context(), orgID, req.PeriodID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ClosePeriod closes a period.
// POST /periods/:id/close
func (h *ReportHandler) ClosePeriod(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")
	periodID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period id"})
		return
	}

	err = h.periodSvc.ClosePeriod(c.Request.Context(), orgID, periodID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "结账成功"})
}

// RevertClosing reverts a closed period.
// POST /periods/:id/revert
func (h *ReportHandler) RevertClosing(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	periodID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period id"})
		return
	}
	var req RevertPeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !req.Confirm {
		c.JSON(http.StatusBadRequest, gin.H{"error": "confirm must be true"})
		return
	}

	err = h.periodSvc.RevertClosing(c.Request.Context(), orgID, periodID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "反结账成功"})
}

// splitAndTrim splits a string by sep and trims whitespace.
func splitAndTrim(s, sep string) []string {
	var result []string
	for _, part := range strings.Split(s, sep) {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
