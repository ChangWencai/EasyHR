package socialinsurance

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// paymentChannelLabel 缴费渠道中文映射
func paymentChannelLabel(channel string) string {
	switch channel {
	case SIPayChannelSelf:
		return "自主缴费"
	case SIPayChannelAgentNew:
		return "代理新客"
	case SIPayChannelAgentExisting:
		return "代理已合作"
	default:
		return channel
	}
}

// paymentStatusLabel 缴费状态中文映射
func paymentStatusLabel(status PaymentStatus) string {
	switch status {
	case PaymentStatusNormal:
		return "正常"
	case PaymentStatusPending:
		return "待缴"
	case PaymentStatusOverdue:
		return "欠缴"
	case PaymentStatusTransferred:
		return "已转出"
	case PaymentStatusNotTransferred:
		return "未转出"
	default:
		return string(status)
	}
}

// ExportSIRecordsWithDetails 导出参保记录 Excel（含五险分项列，per D-SI-13）
// 列：员工姓名/缴费城市/社保基数/参保月/缴费渠道/状态/6险×2列/合计单位/合计个人/欠缴金额/备注
func ExportSIRecordsWithDetails(c *gin.Context, records []SocialInsuranceRecord, includeDetails bool) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "参保记录"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return fmt.Errorf("create sheet: %w", err)
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	// 表头定义（22列 = 6基础 + 6险×2 + 合计单位 + 合计个人 + 欠缴金额 + 备注）
	headers := []string{
		"员工姓名", "缴费城市", "社保基数", "参保月",
		"缴费渠道", "状态",
	}
	if includeDetails {
		headers = append(headers,
			"养老保险-单位", "养老保险-个人",
			"医疗保险-单位", "医疗保险-个人",
			"失业保险-单位", "失业保险-个人",
			"工伤保险-单位", "工伤保险-个人",
			"生育保险-单位", "生育保险-个人",
			"住房公积金-单位", "住房公积金-个人",
		)
	}
	headers = append(headers,
		"合计单位", "合计个人",
		"欠缴金额", "备注",
	)

	// 表头样式（蓝底白字居中）
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#4472C4"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}
	lastHeaderCol, _ := excelize.CoordinatesToCellName(len(headers), 1)
	f.SetCellStyle(sheetName, "A1", lastHeaderCol, headerStyle)

	// 数字格式（保留两位小数）
	numStyle, _ := f.NewStyle(&excelize.Style{NumFmt: 2})

	// 险种顺序（与表头一致）
	insuranceOrder := []string{"养老保险", "医疗保险", "失业保险", "工伤保险", "生育保险", "住房公积金"}

	// 数据行
	for rowIdx, record := range records {
		row := rowIdx + 2
		col := 1

		// 解析险种明细
		var details []InsuranceAmountDetail
		if record.Details != nil {
			_ = json.Unmarshal(record.Details, &details)
		}
		detailMap := make(map[string]InsuranceAmountDetail)
		for _, d := range details {
			detailMap[d.Name] = d
		}

		// 基本信息列
		cityName := getCityName(record.CityID)
		cell, _ := excelize.CoordinatesToCellName(col, row)
		f.SetCellValue(sheetName, cell, record.EmployeeName)
		col++

		cell, _ = excelize.CoordinatesToCellName(col, row)
		f.SetCellValue(sheetName, cell, cityName)
		col++

		cell, _ = excelize.CoordinatesToCellName(col, row)
		f.SetCellValue(sheetName, cell, record.BaseAmount)
		f.SetCellStyle(sheetName, cell, cell, numStyle)
		col++

		cell, _ = excelize.CoordinatesToCellName(col, row)
		f.SetCellValue(sheetName, cell, record.StartMonth)
		col++

		// 缴费渠道 — 从关联的 SIMonthlyPayment 推导，或使用默认值
		// 导出时直接显示"自主缴费"作为默认渠道
		cell, _ = excelize.CoordinatesToCellName(col, row)
		f.SetCellValue(sheetName, cell, paymentChannelLabel(SIPayChannelSelf))
		col++

		// 状态（参保状态：pending/active/stopped）
		statusText := record.Status
		switch record.Status {
		case "pending":
			statusText = "待参保"
		case "active":
			statusText = "参保中"
		case "stopped":
			statusText = "停缴"
		}
		cell, _ = excelize.CoordinatesToCellName(col, row)
		f.SetCellValue(sheetName, cell, statusText)
		col++

		// 五险分项列（仅 includeDetails=true 时填充）
		if includeDetails {
			for _, name := range insuranceOrder {
				d, ok := detailMap[name]

				// 单位缴纳
				cell, _ = excelize.CoordinatesToCellName(col, row)
				companyAmt := 0.0
				if ok {
					companyAmt = d.CompanyAmount
				}
				f.SetCellValue(sheetName, cell, companyAmt)
				f.SetCellStyle(sheetName, cell, cell, numStyle)
				col++

				// 个人缴纳
				cell, _ = excelize.CoordinatesToCellName(col, row)
				personalAmt := 0.0
				if ok {
					personalAmt = d.PersonalAmount
				}
				f.SetCellValue(sheetName, cell, personalAmt)
				f.SetCellStyle(sheetName, cell, cell, numStyle)
				col++
			}
		}

		// 合计单位
		cell, _ = excelize.CoordinatesToCellName(col, row)
		f.SetCellValue(sheetName, cell, record.TotalCompany)
		f.SetCellStyle(sheetName, cell, cell, numStyle)
		col++

		// 合计个人
		cell, _ = excelize.CoordinatesToCellName(col, row)
		f.SetCellValue(sheetName, cell, record.TotalPersonal)
		f.SetCellStyle(sheetName, cell, cell, numStyle)
		col++

		// 欠缴金额（当状态为 overdue 时显示）
		cell, _ = excelize.CoordinatesToCellName(col, row)
		f.SetCellValue(sheetName, cell, 0.0)
		f.SetCellStyle(sheetName, cell, cell, numStyle)
		col++

		// 备注
		cell, _ = excelize.CoordinatesToCellName(col, row)
		f.SetCellValue(sheetName, cell, "")
	}

	// 合计行
	totalRow := len(records) + 2
	cell, _ := excelize.CoordinatesToCellName(1, totalRow)
	f.SetCellValue(sheetName, cell, "合计")

	// 对数值列使用 SUM 公式
	// 列号映射：C=社保基数(col3)，合计单位、合计个人、欠缴金额
	sumCols := []int{}
	// 社保基数列（col 3）
	sumCols = append(sumCols, 3)
	if includeDetails {
		// 五险分项列（col 7~18，共12列）
		for c := 7; c <= 18; c++ {
			sumCols = append(sumCols, c)
		}
	}
	// 合计单位（col after details or 7）
	companyCol := 7
	personalCol := 8
	overdueCol := 9
	if includeDetails {
		companyCol = 19
		personalCol = 20
		overdueCol = 21
	}
	sumCols = append(sumCols, companyCol, personalCol, overdueCol)

	for _, colNum := range sumCols {
		colName, _ := excelize.ColumnNumberToName(colNum)
		startCell := fmt.Sprintf("%s2", colName)
		endCell := fmt.Sprintf("%s%d", colName, len(records)+1)
		sumFormula := fmt.Sprintf("SUM(%s:%s)", startCell, endCell)
		cell, _ := excelize.CoordinatesToCellName(colNum, totalRow)
		f.SetCellValue(sheetName, cell, sumFormula)
		f.SetCellStyle(sheetName, cell, cell, numStyle)
	}

	// 列宽
	for i := range headers {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheetName, colName, colName, 16)
	}
	f.SetColWidth(sheetName, "A", "A", 12)

	// 写入 HTTP Response
	buf, err := f.WriteToBuffer()
	if err != nil {
		return fmt.Errorf("write excel buffer: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("参保记录_%s.xlsx", timestamp)

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(filename)))
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())

	return nil
}

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
