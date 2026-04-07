package tax

import (
	"bytes"
	"fmt"
	"time"

	"github.com/go-pdf/fpdf"
)

// TaxCertificateData 个税凭证 PDF 填充数据
type TaxCertificateData struct {
	EmployeeName     string
	Year, Month      int
	GrossIncome      float64
	TotalDeduction   float64
	TaxRate          float64
	MonthlyTax       float64
	CumulativeIncome float64
	CumulativeTax    float64
	OrgName          string
}

// generateTaxCertificatePDF 生成个税凭证 PDF
// V1.0 使用内置 Helvetica 字体，内容以英文标注
func generateTaxCertificatePDF(data *TaxCertificateData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// 标题
	pdf.SetFont("Helvetica", "B", 18)
	pdf.CellFormat(0, 15, "Individual Income Tax Certificate", "", 0, "C", false, 0, "")
	pdf.Ln(12)

	// 企业名称
	pdf.SetFont("Helvetica", "", 12)
	if data.OrgName != "" {
		pdf.CellFormat(0, 8, fmt.Sprintf("Organization: %s", data.OrgName), "", 1, "C", false, 0, "")
	}
	pdf.Ln(5)

	lineHeight := 8.0

	// 员工信息区
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, lineHeight, "Employee Information", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("  Name: %s", data.EmployeeName), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("  Tax Year: %d     Tax Month: %d", data.Year, data.Month), "", 1, "L", false, 0, "")

	pdf.Ln(8)

	// 个税明细表格
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, lineHeight, "Tax Details", "", 1, "L", false, 0, "")
	pdf.Ln(3)

	// 表格参数
	colWidth := 90.0
	labelWidth := 55.0
	valueWidth := colWidth - labelWidth

	// 左列和右列数据
	leftItems := []struct {
		label string
		value string
	}{
		{"Monthly Income:", fmt.Sprintf("%.2f CNY", data.GrossIncome)},
		{"Total Deduction:", fmt.Sprintf("%.2f CNY", data.TotalDeduction)},
		{"Tax Rate:", fmt.Sprintf("%.0f%%", data.TaxRate*100)},
	}

	rightItems := []struct {
		label string
		value string
	}{
		{"Monthly Tax:", fmt.Sprintf("%.2f CNY", data.MonthlyTax)},
		{"Cumulative Income:", fmt.Sprintf("%.2f CNY", data.CumulativeIncome)},
		{"Cumulative Tax:", fmt.Sprintf("%.2f CNY", data.CumulativeTax)},
	}

	pdf.SetFont("Helvetica", "", 11)

	// 表格边框
	tableTop := pdf.GetY()
	tableHeight := float64(len(leftItems)) * lineHeight

	// 绘制数据行
	for i := 0; i < len(leftItems); i++ {
		y := tableTop + float64(i)*lineHeight

		// 左列
		pdf.SetXY(10, y)
		pdf.SetFont("Helvetica", "", 11)
		pdf.CellFormat(labelWidth, lineHeight, leftItems[i].label, "1", 0, "L", false, 0, "")
		pdf.CellFormat(valueWidth, lineHeight, leftItems[i].value, "1", 0, "R", false, 0, "")

		// 右列
		pdf.SetFont("Helvetica", "", 11)
		pdf.CellFormat(labelWidth, lineHeight, rightItems[i].label, "1", 0, "L", false, 0, "")
		pdf.CellFormat(valueWidth, lineHeight, rightItems[i].value, "1", 0, "R", false, 0, "")
	}

	pdf.Ln(tableHeight + 10)

	// 打印日期
	pdf.SetFont("Helvetica", "", 10)
	printDate := time.Now().Format("2006-01-02")
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("Print Date: %s", printDate), "", 1, "R", false, 0, "")

	// 输出到 buffer
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("generate PDF: %w", err)
	}
	return buf.Bytes(), nil
}
