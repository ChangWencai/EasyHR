package salary

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/logger"
	"github.com/wencai/easyhr/internal/common/response"
)

// SalaryListHandler 薪资列表 HTTP 端点
type SalaryListHandler struct {
	svc *Service
}

// NewSalaryListHandler 创建薪资列表 Handler
func NewSalaryListHandler(svc *Service) *SalaryListHandler {
	return &SalaryListHandler{svc: svc}
}

// RegisterRoutes 注册薪资列表路由
func (h *SalaryListHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	list := rg.Group("/salary", authMiddleware)
	{
		list.GET("/list", h.ListSalary)
		list.GET("/export", h.ExportSalaryList)
	}
}

// ListSalary 查询薪资列表
func (h *SalaryListHandler) ListSalary(c *gin.Context) {
	orgID := c.GetInt64("org_id")

	yearStr := c.Query("year")
	monthStr := c.Query("month")
	deptStr := c.Query("department_id")
	keyword := c.Query("keyword")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		response.BadRequest(c, "year 参数错误")
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		response.BadRequest(c, "month 参数错误")
		return
	}

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var deptID *int64
	if deptStr != "" {
		id, err := strconv.ParseInt(deptStr, 10, 64)
		if err == nil {
			deptID = &id
		}
	}

	filter := SalaryListFilter{
		OrgID:        orgID,
		Year:         year,
		Month:        month,
		DepartmentID: deptID,
		Keyword:      keyword,
		Page:         page,
		PageSize:     pageSize,
	}

	logger.SugarLogger.Debugw("ListSalary: 查询", "org_id", orgID, "year", year, "month", month)

	records, total, err := h.svc.ListSalaryRecords(filter)
	if err != nil {
		logger.SugarLogger.Debugw("ListSalary: 失败", "error", err.Error())
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, err.Error())
		return
	}

	response.PageSuccess(c, records, total, page, pageSize)
}

// ExportSalaryList 导出薪资列表（支持含明细列选项）
func (h *SalaryListHandler) ExportSalaryList(c *gin.Context) {
	orgID := c.GetInt64("org_id")

	yearStr := c.Query("year")
	monthStr := c.Query("month")
	includeDetailsStr := c.DefaultQuery("include_details", "false")

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		response.BadRequest(c, "year 参数错误")
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		response.BadRequest(c, "month 参数错误")
		return
	}

	includeDetails := includeDetailsStr == "true" || includeDetailsStr == "1"

	logger.SugarLogger.Debugw("ExportSalaryList: 查询", "org_id", orgID, "year", year, "month", month, "include_details", includeDetails)

	// 查询工资记录
	records, err := h.svc.repo.FindPayrollRecordsByMonth(orgID, year, month)
	if err != nil {
		logger.SugarLogger.Debugw("ExportSalaryList: 查询工资记录失败", "error", err.Error())
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, "查询工资记录失败")
		return
	}

	// 查询工资明细
	var recordsWithItems []PayrollRecordWithItems
	for _, record := range records {
		items, err := h.svc.repo.FindPayrollItemsByRecord(orgID, record.ID)
		if err != nil {
			continue
		}
		recordsWithItems = append(recordsWithItems, PayrollRecordWithItems{
			Record: record,
			Items:  items,
		})
	}

	// 生成 Excel
	data, err := ExportPayrollExcelWithDetails(recordsWithItems, year, month, includeDetails)
	if err != nil {
		logger.SugarLogger.Debugw("ExportSalaryList: 生成 Excel 失败", "error", err.Error())
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, "生成 Excel 失败")
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=工资条_%d_%02d.xlsx", year, month))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
