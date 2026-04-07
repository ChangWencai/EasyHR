package socialinsurance

import (
	"encoding/json"
	"fmt"

	"github.com/xuri/excelize/v2"
)

// generatePaymentDetailExcel 生成缴费明细 Excel
func generatePaymentDetailExcel(records []SocialInsuranceRecord) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "社保缴费明细"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("create sheet: %w", err)
	}
	f.SetActiveSheet(index)

	// 删除默认 Sheet1
	f.DeleteSheet("Sheet1")

	// 表头
	headers := []string{
		"员工姓名", "城市", "参保月份", "基数",
		"养老保险(企业)", "养老保险(个人)",
		"医疗保险(企业)", "医疗保险(个人)",
		"失业保险(企业)", "失业保险(个人)",
		"工伤保险(企业)",
		"生育保险(企业)",
		"住房公积金(企业)", "住房公积金(个人)",
		"企业合计", "个人合计",
	}

	// 设置表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#4472C4"}},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// 数据行
	for row, record := range records {
		rowNum := row + 2

		// 解析险种明细
		var details []InsuranceAmountDetail
		if record.Details != nil {
			_ = json.Unmarshal(record.Details, &details)
		}

		// 基本信息列
		cityName := getCityName(record.CityID)
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), record.EmployeeName)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), cityName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), record.StartMonth)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), record.BaseAmount)

		// 各险种金额列
		detailMap := make(map[string]InsuranceAmountDetail)
		for _, d := range details {
			detailMap[d.Name] = d
		}

		insuranceOrder := []string{"养老保险", "医疗保险", "失业保险", "工伤保险", "生育保险", "住房公积金"}
		// 企业列: 养老E, 医疗G, 失业I, 工伤K, 生育L, 公积金M
		// 个人列: 养老F, 医疗H, 失业J, 公积金N
		companyCols := []string{"E", "G", "I", "K", "L", "M"}
		personalCols := []string{"F", "H", "J", "", "", "N"}

		for idx, name := range insuranceOrder {
			d, ok := detailMap[name]
			if !ok {
				continue
			}
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", companyCols[idx], rowNum), d.CompanyAmount)
			if personalCols[idx] != "" {
				f.SetCellValue(sheetName, fmt.Sprintf("%s%d", personalCols[idx], rowNum), d.PersonalAmount)
			}
		}

		// 合计列
		f.SetCellValue(sheetName, fmt.Sprintf("O%d", rowNum), record.TotalCompany)
		f.SetCellValue(sheetName, fmt.Sprintf("P%d", rowNum), record.TotalPersonal)
	}

	// 最后一行：合计行
	totalRow := len(records) + 2
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", totalRow), "合计")

	// 数值列求和
	for _, col := range []string{"E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P"} {
		startCell := fmt.Sprintf("%s2", col)
		endCell := fmt.Sprintf("%s%d", col, len(records)+1)
		sumFormula := fmt.Sprintf("SUM(%s:%s)", startCell, endCell)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", col, totalRow), sumFormula)
	}

	// 自适应列宽
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
