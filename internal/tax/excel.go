package tax

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// generateDeclarationExcel 生成个税申报表 Excel
// 格式对齐自然人电子税务局批量导入模板
func generateDeclarationExcel(records []TaxRecord, year, month int) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := fmt.Sprintf("个税申报表_%d年%d月", year, month)
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("create sheet: %w", err)
	}
	f.SetActiveSheet(index)

	// 删除默认 Sheet1
	f.DeleteSheet("Sheet1")

	// 表头行 (15列, 对齐自然人电子税务局格式)
	headers := []string{
		"纳税人姓名", "证件类型", "证件号码(脱敏)", "所得项目",
		"收入额", "基本减除费用", "专项扣除(社保)", "专项附加扣除",
		"其他扣除", "应纳税所得额", "税率", "速算扣除数",
		"应扣税额", "已扣税额", "本期应扣缴税额",
	}

	// 设置表头样式（加粗蓝底白字）
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#4472C4"}},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// 数值格式（保留2位小数）
	numStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 2,
	})

	// 数据行
	for row, record := range records {
		rowNum := row + 2

		// 纳税人姓名
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), record.EmployeeName)
		// 证件类型
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), "居民身份证")
		// 证件号码(脱敏) -- V1.0 简化处理
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), "***")
		// 所得项目
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), "工资薪金所得")

		// 收入额
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), record.GrossIncome)
		f.SetCellStyle(sheetName, fmt.Sprintf("E%d", rowNum), fmt.Sprintf("E%d", rowNum), numStyle)

		// 基本减除费用
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), record.BasicDeduction)
		f.SetCellStyle(sheetName, fmt.Sprintf("F%d", rowNum), fmt.Sprintf("F%d", rowNum), numStyle)

		// 专项扣除(社保)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), record.SIDeduction)
		f.SetCellStyle(sheetName, fmt.Sprintf("G%d", rowNum), fmt.Sprintf("G%d", rowNum), numStyle)

		// 专项附加扣除
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowNum), record.SpecialDeduction)
		f.SetCellStyle(sheetName, fmt.Sprintf("H%d", rowNum), fmt.Sprintf("H%d", rowNum), numStyle)

		// 其他扣除
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", rowNum), 0)
		f.SetCellStyle(sheetName, fmt.Sprintf("I%d", rowNum), fmt.Sprintf("I%d", rowNum), numStyle)

		// 应纳税所得额 (累计值)
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", rowNum), record.CumulativeTaxableIncome)
		f.SetCellStyle(sheetName, fmt.Sprintf("J%d", rowNum), fmt.Sprintf("J%d", rowNum), numStyle)

		// 税率 (转为百分比显示)
		rateStr := fmt.Sprintf("%.0f%%", record.TaxRate*100)
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", rowNum), rateStr)

		// 速算扣除数
		f.SetCellValue(sheetName, fmt.Sprintf("L%d", rowNum), record.QuickDeduction)
		f.SetCellStyle(sheetName, fmt.Sprintf("L%d", rowNum), fmt.Sprintf("L%d", rowNum), numStyle)

		// 应扣税额 (累计税额)
		f.SetCellValue(sheetName, fmt.Sprintf("M%d", rowNum), record.CumulativeTax)
		f.SetCellStyle(sheetName, fmt.Sprintf("M%d", rowNum), fmt.Sprintf("M%d", rowNum), numStyle)

		// 已扣税额 (累计税额 - 本月税额)
		prevTax := record.CumulativeTax - record.MonthlyTax
		f.SetCellValue(sheetName, fmt.Sprintf("N%d", rowNum), prevTax)
		f.SetCellStyle(sheetName, fmt.Sprintf("N%d", rowNum), fmt.Sprintf("N%d", rowNum), numStyle)

		// 本期应扣缴税额
		f.SetCellValue(sheetName, fmt.Sprintf("O%d", rowNum), record.MonthlyTax)
		f.SetCellStyle(sheetName, fmt.Sprintf("O%d", rowNum), fmt.Sprintf("O%d", rowNum), numStyle)
	}

	// 最后一行：合计行（SUM公式）
	totalRow := len(records) + 2
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", totalRow), "合计")

	// 数值列求和 (E-O)
	for _, col := range []string{"E", "F", "G", "H", "I", "J", "L", "M", "N", "O"} {
		startCell := fmt.Sprintf("%s2", col)
		endCell := fmt.Sprintf("%s%d", col, len(records)+1)
		sumFormula := fmt.Sprintf("SUM(%s:%s)", startCell, endCell)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", col, totalRow), sumFormula)
	}

	// 自适应列宽: 所有列宽16
	for i := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheetName, col, col, 16)
	}

	// 写入 buffer
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("write excel: %w", err)
	}
	return buf.Bytes(), nil
}
