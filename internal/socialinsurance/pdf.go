package socialinsurance

import (
	"bytes"
	"fmt"

	"github.com/go-pdf/fpdf"
)

// EnrollmentPDFData 参保材料 PDF 填充数据
type EnrollmentPDFData struct {
	EmployeeName  string
	CityName      string
	BaseAmount    float64
	StartMonth    string
	Items         []InsuranceAmountDetail
	TotalCompany  float64
	TotalPersonal float64
}

// generateEnrollmentPDF 生成参保材料 PDF
// V1.0 使用内置 Helvetica 字体，内容以英文标注
func generateEnrollmentPDF(data *EnrollmentPDFData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// 标题
	pdf.SetFont("Helvetica", "B", 18)
	pdf.CellFormat(0, 15, "Social Insurance Enrollment Form", "", 0, "C", false, 0, "")
	pdf.Ln(20)

	lineHeight := 8.0

	// 员工信息区
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, lineHeight, "Employee Information", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("  Name: %s", data.EmployeeName), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("  City: %s", data.CityName), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("  Base Amount: %.2f CNY", data.BaseAmount), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("  Start Month: %s", data.StartMonth), "", 1, "L", false, 0, "")

	pdf.Ln(10)

	// 各险种明细表
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, lineHeight, "Insurance Details", "", 1, "L", false, 0, "")
	pdf.Ln(3)

	// 表头
	pdf.SetFont("Helvetica", "B", 9)
	colWidths := []float64{35, 22, 22, 28, 22, 28, 28}
	headers := []string{"Type", "Base", "Co. Rate", "Co. Amount", "Per. Rate", "Per. Amount", "Subtotal"}

	for i, h := range headers {
		pdf.CellFormat(colWidths[i], lineHeight, h, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(lineHeight)

	// 数据行
	pdf.SetFont("Helvetica", "", 9)
	for _, item := range data.Items {
		subtotal := item.CompanyAmount + item.PersonalAmount
		cells := []string{
			item.Name,
			fmt.Sprintf("%.2f", item.Base),
			fmt.Sprintf("%.2f%%", item.CompanyRate*100),
			fmt.Sprintf("%.2f", item.CompanyAmount),
			fmt.Sprintf("%.2f%%", item.PersonalRate*100),
			fmt.Sprintf("%.2f", item.PersonalAmount),
			fmt.Sprintf("%.2f", subtotal),
		}
		for i, c := range cells {
			pdf.CellFormat(colWidths[i], lineHeight, c, "1", 0, "C", false, 0, "")
		}
		pdf.Ln(lineHeight)
	}

	// 合计行
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(colWidths[0], lineHeight, "TOTAL", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colWidths[1], lineHeight, "", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colWidths[2], lineHeight, "", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colWidths[3], lineHeight, fmt.Sprintf("%.2f", data.TotalCompany), "1", 0, "C", false, 0, "")
	pdf.CellFormat(colWidths[4], lineHeight, "", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colWidths[5], lineHeight, fmt.Sprintf("%.2f", data.TotalPersonal), "1", 0, "C", false, 0, "")
	totalAll := data.TotalCompany + data.TotalPersonal
	pdf.CellFormat(colWidths[6], lineHeight, fmt.Sprintf("%.2f", totalAll), "1", 0, "C", false, 0, "")
	pdf.Ln(lineHeight)

	// 输出到 buffer
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("generate PDF: %w", err)
	}
	return buf.Bytes(), nil
}
