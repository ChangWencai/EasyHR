package salary

import (
	"fmt"
	"strings"

	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// TaxUploadRow 解析后的个税 Excel 行
type TaxUploadRow struct {
	Name       string
	TaxAmount  float64 // 个税金额
	Adjustment float64 // 应补/应退额（正=补，负=退）
	RowNumber  int     // Excel 行号（从1开始）
}

// TaxUploadResult 上传结果
type TaxUploadResult struct {
	TotalRows     int
	MatchedCount  int
	MatchedRows   []TaxMatchedRow
	UnmatchedRows []UnmatchedRow
}

// TaxMatchedRow 匹配成功的行
type TaxMatchedRow struct {
	RowNumber   int
	Name        string
	EmployeeID  int64
	EmployeeName string
	TaxAmount   float64
	Adjustment  float64
}

// UnmatchedRow 无法匹配的行
type UnmatchedRow struct {
	RowNumber int
	Name      string
	Reason    string // "未找到匹配员工" / "存在多个匹配"
}

// 列名别名映射
var (
	taxNameAliases       = []string{"姓名", "员工姓名", "纳税人姓名"}
	taxAmountAliases     = []string{"个税", "税额", "个人所得税", "本期应扣缴税额"}
	taxAdjustmentAliases = []string{"应补/应退额", "应补退额", "应补（退）额", "应补", "应退"}
)

// UploadTaxFile 解析 Excel 并返回匹配结果（不更新数据库）
func (s *Service) UploadTaxFile(orgID int64, year, month int, file []byte) (*TaxUploadResult, error) {
	// 1. 解析 Excel
	rows, err := parseTaxExcel(file)
	if err != nil {
		return nil, fmt.Errorf("解析 Excel 文件失败: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("Excel 文件没有数据行")
	}

	// 2. 获取在职员工列表
	employees, err := s.empProvider.GetActiveEmployees(orgID)
	if err != nil {
		return nil, fmt.Errorf("获取员工列表失败: %w", err)
	}

	// 3. 逐行匹配
	result := &TaxUploadResult{TotalRows: len(rows)}
	for _, row := range rows {
		matched := matchEmployee(row.Name, employees)
		if len(matched) == 1 {
			result.MatchedCount++
			result.MatchedRows = append(result.MatchedRows, TaxMatchedRow{
				RowNumber:    row.RowNumber,
				Name:         row.Name,
				EmployeeID:   matched[0].ID,
				EmployeeName: matched[0].Name,
				TaxAmount:    row.TaxAmount,
				Adjustment:   row.Adjustment,
			})
		} else if len(matched) > 1 {
			result.UnmatchedRows = append(result.UnmatchedRows, UnmatchedRow{
				RowNumber: row.RowNumber,
				Name:      row.Name,
				Reason:    "存在多个匹配员工",
			})
		} else {
			result.UnmatchedRows = append(result.UnmatchedRows, UnmatchedRow{
				RowNumber: row.RowNumber,
				Name:      row.Name,
				Reason:    "未找到匹配员工",
			})
		}
	}

	return result, nil
}

// ConfirmTaxUpload 确认上传并更新工资记录
func (s *Service) ConfirmTaxUpload(orgID, userID int64, year, month int, matchedRows []TaxMatchedRow) error {
	return s.repo.db.Transaction(func(tx *gorm.DB) error {
		for _, row := range matchedRows {
			// 查找该员工该月的工资记录
			var record PayrollRecord
			if err := tx.Scopes(middleware.TenantScope(orgID)).
				Where("employee_id = ? AND year = ? AND month = ?", row.EmployeeID, year, month).
				First(&record).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					continue // 该员工没有工资记录，跳过
				}
				return fmt.Errorf("查询工资记录失败: %w", err)
			}

			// 更新个税金额
			newTax := record.Tax + row.TaxAmount + row.Adjustment
			if newTax < 0 {
				newTax = 0
			}
			// 重算扣除合计 = 原扣除合计 - 原个税 + 新个税
			newDeductions := record.TotalDeductions - record.Tax + newTax
			// 重算实发工资
			newNetIncome := record.GrossIncome - newDeductions

			if err := tx.Model(&record).Updates(map[string]interface{}{
				"tax":             newTax,
				"total_deductions": newDeductions,
				"net_income":      newNetIncome,
				"updated_by":      userID,
			}).Error; err != nil {
				return fmt.Errorf("更新员工 %s 个税失败: %w", record.EmployeeName, err)
			}
		}

		// 将状态重置为 draft（如果当前是 calculated 或 confirmed）
		if err := tx.Model(&PayrollRecord{}).
			Scopes(middleware.TenantScope(orgID)).
			Where("year = ? AND month = ? AND status IN ?", year, month, []string{PayrollStatusCalculated, PayrollStatusConfirmed}).
			Update("status", PayrollStatusDraft).Error; err != nil {
			return fmt.Errorf("重置工资状态失败: %w", err)
		}

		return nil
	})
}

// parseTaxExcel 解析个税 Excel 文件
func parseTaxExcel(file []byte) ([]TaxUploadRow, error) {
	f, err := excelize.OpenReader(strings.NewReader(string(file)))
	if err != nil {
		return nil, fmt.Errorf("打开 Excel 文件失败: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("Excel 文件没有工作表")
	}
	allRows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("读取 Excel 行失败: %w", err)
	}

	if len(allRows) == 0 {
		return nil, fmt.Errorf("Excel 文件没有数据")
	}

	// 解析表头，找到列索引
	header := allRows[0]
	nameCol, taxCol, adjCol := findTaxColumns(header)

	if nameCol < 0 {
		return nil, fmt.Errorf("未找到姓名列（支持的列名: %s）", strings.Join(taxNameAliases, "、"))
	}

	var result []TaxUploadRow
	for i, row := range allRows {
		if i == 0 {
			continue // 跳过表头
		}

		name := getCellValue(row, nameCol)
		if strings.TrimSpace(name) == "" {
			continue // 空行跳过
		}

		tr := TaxUploadRow{
			Name:      strings.TrimSpace(name),
			RowNumber: i + 1, // Excel 行号从1开始
		}

		if taxCol >= 0 && getCellValue(row, taxCol) != "" {
			fmt.Sscanf(strings.TrimSpace(getCellValue(row, taxCol)), "%f", &tr.TaxAmount)
		}
		if adjCol >= 0 && getCellValue(row, adjCol) != "" {
			fmt.Sscanf(strings.TrimSpace(getCellValue(row, adjCol)), "%f", &tr.Adjustment)
		}

		result = append(result, tr)
	}

	return result, nil
}

// findTaxColumns 从表头中找到姓名、个税、应补/应退额的列索引
func findTaxColumns(header []string) (nameCol, taxCol, adjCol int) {
	nameCol, taxCol, adjCol = -1, -1, -1

	for i, h := range header {
		trimmed := strings.TrimSpace(h)
		if nameCol < 0 && containsAlias(trimmed, taxNameAliases) {
			nameCol = i
		}
		if taxCol < 0 && containsAlias(trimmed, taxAmountAliases) {
			taxCol = i
		}
		if adjCol < 0 && containsAlias(trimmed, taxAdjustmentAliases) {
			adjCol = i
		}
	}

	return nameCol, taxCol, adjCol
}

// containsAlias 检查表头是否匹配某个别名
func containsAlias(header string, aliases []string) bool {
	for _, alias := range aliases {
		if header == alias {
			return true
		}
	}
	return false
}

// getCellValue 安全获取行中的单元格值
func getCellValue(row []string, col int) string {
	if col < len(row) {
		return row[col]
	}
	return ""
}

// matchEmployee 匹配员工：精确匹配 -> 模糊匹配 -> 无匹配
func matchEmployee(name string, employees []EmployeeInfo) []EmployeeInfo {
	// 1. 精确匹配
	for i := range employees {
		if employees[i].Name == name {
			return []EmployeeInfo{employees[i]}
		}
	}

	// 2. 模糊匹配（name 是员工姓名的子串）
	var matched []EmployeeInfo
	for i := range employees {
		if strings.Contains(employees[i].Name, name) || strings.Contains(name, employees[i].Name) {
			matched = append(matched, employees[i])
		}
	}

	return matched
}

