package employee

import (
	"bytes"
	"fmt"
	"time"

	"github.com/go-pdf/fpdf"
)

// ContractPDFData 合同 PDF 填充数据
type ContractPDFData struct {
	OrgName      string
	CreditCode   string
	EmployeeName string
	IDCard       string
	Position     string
	City         string
	Salary       float64
	StartDate    time.Time
	EndDate      *time.Time
	ContractType string
	SignDate     time.Time
}

// GenerateContractPDF 生成劳动合同 PDF
// V1.0 使用内置 Helvetica 字体，中文内容以拼音/英文标注
// TODO: 后续需注册中文字体（如思源黑体 TTF ~16MB），通过 pdf.AddUTF8Font() 实现
func GenerateContractPDF(data *ContractPDFData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 20)

	// 标题
	pdf.CellFormat(0, 15, "Labor Contract", "", 0, "C", false, 0, "")
	pdf.Ln(20)

	pdf.SetFont("Helvetica", "", 12)
	lineHeight := 8.0

	// 甲方信息（企业）
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, lineHeight, "Party A (Employer)", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 12)
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("  Company: %s", data.OrgName), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("  Credit Code: %s", data.CreditCode), "", 1, "L", false, 0, "")

	pdf.Ln(5)

	// 乙方信息（员工）
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, lineHeight, "Party B (Employee)", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 12)
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("  Name: %s", data.EmployeeName), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("  ID Card: %s", data.IDCard), "", 1, "L", false, 0, "")

	pdf.Ln(5)

	// 合同条款
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, lineHeight, "Contract Terms", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 12)

	// 合同类型
	contractTypeStr := "Fixed Term"
	if data.ContractType == ContractTypeIndefinite {
		contractTypeStr = "Indefinite Term"
	} else if data.ContractType == ContractTypeIntern {
		contractTypeStr = "Internship"
	}
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("  Type: %s", contractTypeStr), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("  Position: %s", data.Position), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("  Work Location: %s", data.City), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("  Monthly Salary: %.2f CNY", data.Salary), "", 1, "L", false, 0, "")

	// 合同期限
	contractPeriod := fmt.Sprintf("From %s", data.StartDate.Format("2006-01-02"))
	if data.EndDate != nil {
		contractPeriod += fmt.Sprintf(" to %s", data.EndDate.Format("2006-01-02"))
	} else {
		contractPeriod += " (Indefinite term)"
	}
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("  Contract Period: %s", contractPeriod), "", 1, "L", false, 0, "")

	pdf.Ln(15)

	// 签名区域
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, lineHeight, "Signatures", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 12)
	pdf.CellFormat(0, lineHeight, "Party A Seal: ____________    Date: ____________", "", 1, "L", false, 0, "")
	pdf.Ln(5)
	pdf.CellFormat(0, lineHeight, "Party B Signature: ____________    Date: ____________", "", 1, "L", false, 0, "")

	// 输出到 buffer
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("generate PDF: %w", err)
	}
	return buf.Bytes(), nil
}
