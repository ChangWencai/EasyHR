package salary

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// AttendanceRow 考勤导入行
type AttendanceRow struct {
	Name               string
	SickLeaveDays      float64
	PersonalLeaveDays  float64
	Remark             string
}

// parseAttendanceExcel 解析考勤 Excel 文件
// 模板格式：员工姓名 | 事假(天) | 病假(天) | 备注
func parseAttendanceExcel(file []byte) ([]AttendanceRow, error) {
	f, err := excelize.OpenReader(strings.NewReader(string(file)))
	if err != nil {
		return nil, fmt.Errorf("打开 Excel 文件失败: %w", err)
	}
	defer f.Close()

	// 读取第一个 sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("Excel 文件没有工作表")
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("读取 Excel 行失败: %w", err)
	}

	var result []AttendanceRow
	for i, row := range rows {
		if i == 0 {
			continue // 跳过表头
		}
		if len(row) < 1 || strings.TrimSpace(row[0]) == "" {
			continue // 空行跳过
		}

		ar := AttendanceRow{
			Name: strings.TrimSpace(row[0]),
		}

		if len(row) > 1 && strings.TrimSpace(row[1]) != "" {
			var sickDays float64
			fmt.Sscanf(strings.TrimSpace(row[1]), "%f", &sickDays)
			ar.SickLeaveDays = sickDays
		}
		if len(row) > 2 && strings.TrimSpace(row[2]) != "" {
			var personalDays float64
			fmt.Sscanf(strings.TrimSpace(row[2]), "%f", &personalDays)
			ar.PersonalLeaveDays = personalDays
		}
		if len(row) > 3 {
			ar.Remark = strings.TrimSpace(row[3])
		}

		result = append(result, ar)
	}

	return result, nil
}

// PayrollRecordWithItems 工资记录及明细（用于导出）
type PayrollRecordWithItems struct {
	Record PayrollRecord
	Items  []PayrollItem
}

// ExportPayrollExcel 导出工资条 Excel
// 导出格式：员工姓名 | 基本工资 | 绩效 | 补贴合计 | 事假扣款 | 病假扣款 | 其他扣款 | 税前收入 | 社保个人 | 个税 | 实发工资
func ExportPayrollExcel(records []PayrollRecordWithItems, year, month int) ([]byte, error) {
	f := excelize.NewFile()
	sheetName := "工资条"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("创建工作表失败: %w", err)
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	// 定义表头
	headers := []string{
		"员工姓名", "基本工资", "绩效", "补贴合计", "事假扣款",
		"病假扣款", "其他扣款", "税前收入", "社保个人", "个税", "实发工资",
	}

	// 设置表头样式（蓝底白字）
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return nil, fmt.Errorf("创建表头样式失败: %w", err)
	}

	// 写入表头
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return nil, fmt.Errorf("设置表头失败: %w", err)
		}
	}
	// 应用表头样式
	if err := f.SetCellStyle(sheetName, "A1", "K1", headerStyle); err != nil {
		return nil, fmt.Errorf("应用表头样式失败: %w", err)
	}

	// 写入数据行
	row := 2
	var totalGross, totalSI, totalTax, totalNet float64

	for _, recordWithItems := range records {
		rec := recordWithItems.Record
		items := recordWithItems.Items

		// 解析薪资项
		baseSalary := getSalaryItemAmount(items, "基本工资")
		performance := getSalaryItemAmount(items, "绩效工资")
		subsidies := getSalaryItemSum(items, []string{"岗位补贴", "餐补", "交通补", "通讯补", "其他补贴"})
		personalLeave := getSalaryItemAmount(items, "事假扣款")
		sickLeave := getSalaryItemAmount(items, "病假扣款")
		otherDeduction := getSalaryItemAmount(items, "其他扣款")

		// 写入行数据
		values := []interface{}{
			rec.EmployeeName,
			baseSalary,
			performance,
			subsidies,
			personalLeave,
			sickLeave,
			otherDeduction,
			rec.GrossIncome,
			rec.SIDeduction,
			rec.Tax,
			rec.NetIncome,
		}

		for i, val := range values {
			cell, _ := excelize.CoordinatesToCellName(i+1, row)
			if err := f.SetCellValue(sheetName, cell, val); err != nil {
				return nil, fmt.Errorf("设置单元格值失败: %w", err)
			}
		}

		totalGross += rec.GrossIncome
		totalSI += rec.SIDeduction
		totalTax += rec.Tax
		totalNet += rec.NetIncome
		row++
	}

	// 写入合计行
合计Row := row
	totalValues := []interface{}{
		"合计", "", "", "", "", "", "",
		totalGross, totalSI, totalTax, totalNet,
	}

	for i, val := range totalValues {
		cell, _ := excelize.CoordinatesToCellName(i+1, 合计Row)
		if err := f.SetCellValue(sheetName, cell, val); err != nil {
			return nil, fmt.Errorf("设置合计值失败: %w", err)
		}
	}

	// 设置数字格式（保留两位小数）
	numStyle, err := f.NewStyle(&excelize.Style{
		NumFmt: 2, // 0.00
	})
	if err == nil {
		// 应用到所有数据列（B到K）
		for col := 'B'; col <= 'K'; col++ {
			for row := 2; row <= 合计Row; row++ {
				cell := fmt.Sprintf("%c%d", col, row)
				_ = f.SetCellStyle(sheetName, cell, cell, numStyle)
			}
		}
	}

	// 设置列宽
	f.SetColWidth(sheetName, "A", "A", 12)
	f.SetColWidth(sheetName, "B", "K", 14)

	// 生成文件
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("生成 Excel 文件失败: %w", err)
	}

	return buffer.Bytes(), nil
}

// getSalaryItemAmount 获取指定薪资项金额
func getSalaryItemAmount(items []PayrollItem, name string) float64 {
	for _, item := range items {
		if item.ItemName == name {
			return item.Amount
		}
	}
	return 0
}

// getSalaryItemSum 获取多个薪资项总和
func ExportPayrollExcelWithDetails(records []PayrollRecordWithItems, year, month int, includeDetails bool) ([]byte, error) {
	f := excelize.NewFile()
	sheetName := "工资条"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("创建工作表失败: %w", err)
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	// 基础表头
	headers := []string{
		"员工姓名", "税前收入", "社保个人", "公积金个人", "个税",
		"扣除合计", "实发工资", "状态",
	}
	detailHeaders := []string{} // 明细列（当 includeDetails=true 时填充）

	if includeDetails {
		// 收集所有明细项名称作为表头
		itemNameMap := make(map[string]bool)
		var itemNames []string
		for _, rec := range records {
			for _, item := range rec.Items {
				if !itemNameMap[item.ItemName] {
					itemNameMap[item.ItemName] = true
					itemNames = append(itemNames, item.ItemName)
				}
			}
		}
		detailHeaders = itemNames
	}

	// 合并表头
	fullHeaders := append(headers, detailHeaders...)

	// 设置表头样式
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return nil, fmt.Errorf("创建表头样式失败: %w", err)
	}

	for i, header := range fullHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return nil, fmt.Errorf("设置表头失败: %w", err)
		}
	}
	lastCol, _ := excelize.CoordinatesToCellName(len(fullHeaders), 1)
	if err := f.SetCellStyle(sheetName, "A1", lastCol, headerStyle); err != nil {
		return nil, fmt.Errorf("应用表头样式失败: %w", err)
	}

	// 写入数据行
	row := 2
	var totalGross, totalSI, totalTax, totalNet float64

	for _, recordWithItems := range records {
		rec := recordWithItems.Record
		items := recordWithItems.Items

		// 基础列
		rowData := []interface{}{
			rec.EmployeeName,
			rec.GrossIncome,
			rec.SIDeduction,
			0.0, // 公积金单独列（目前合并在 si_deduction 中）
			rec.Tax,
			rec.TotalDeductions,
			rec.NetIncome,
			statusText(rec.Status),
		}

		// 明细列
		if includeDetails {
			for _, name := range detailHeaders {
				amount := getSalaryItemAmount(items, name)
				rowData = append(rowData, amount)
			}
		}

		for i, val := range rowData {
			cell, _ := excelize.CoordinatesToCellName(i+1, row)
			if err := f.SetCellValue(sheetName, cell, val); err != nil {
				return nil, fmt.Errorf("设置单元格值失败: %w", err)
			}
		}

		totalGross += rec.GrossIncome
		totalSI += rec.SIDeduction
		totalTax += rec.Tax
		totalNet += rec.NetIncome
		row++
	}

	// 合计行
	合计Row := row
	totalRowData := []interface{}{
		"合计", totalGross, totalSI, "", totalTax, "", totalNet, "",
	}
	if includeDetails {
		for range detailHeaders {
			totalRowData = append(totalRowData, "")
		}
	}
	for i, val := range totalRowData {
		cell, _ := excelize.CoordinatesToCellName(i+1, 合计Row)
		if err := f.SetCellValue(sheetName, cell, val); err != nil {
			return nil, fmt.Errorf("设置合计值失败: %w", err)
		}
	}

	// 数字格式
	totalCols := len(fullHeaders)
	numStyle, err := f.NewStyle(&excelize.Style{NumFmt: 2})
	if err == nil {
		// 根据表头数量计算最后一列
		for colIdx := 2; colIdx <= totalCols; colIdx++ {
			colName, _ := excelize.CoordinatesToCellName(colIdx, 1)
			for r := 2; r <= 合计Row; r++ {
				cell := fmt.Sprintf("%s%d", colName[:len(colName)-1], r)
				_ = f.SetCellStyle(sheetName, cell, cell, numStyle)
			}
		}
	}

	lastColName, _ := excelize.CoordinatesToCellName(totalCols, 1)
	// 列宽
	f.SetColWidth(sheetName, "A", "A", 12)
	f.SetColWidth(sheetName, "B", lastColName[:len(lastColName)-1], 14)

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("生成 Excel 文件失败: %w", err)
	}

	return buffer.Bytes(), nil
}

// statusText 返回状态中文文本
func statusText(status string) string {
	switch status {
	case "draft":
		return "草稿"
	case "calculated":
		return "已核算"
	case "confirmed":
		return "已确认"
	case "paid":
		return "已发放"
	default:
		return status
	}
}

// getSalaryItemSum 获取多个薪资项总和
func getSalaryItemSum(items []PayrollItem, names []string) float64 {
	var sum float64
	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}
	for _, item := range items {
		if nameMap[item.ItemName] {
			sum += item.Amount
		}
	}
	return sum
}
