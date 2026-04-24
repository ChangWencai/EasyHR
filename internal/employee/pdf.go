package employee

import (
	"bytes"
	"embed"
	"fmt"
	"time"

	"github.com/go-pdf/fpdf"
)

//go:embed fonts/NotoSansSC-Regular.ttf
//go:embed fonts/NotoSansSC-Bold.ttf
var fontFiles embed.FS

// ContractPDFData 合同 PDF 填充数据
type ContractPDFData struct {
	OrgName         string
	CreditCode      string
	EmployeeName    string
	IDCard          string
	Position        string
	City            string
	Salary          float64
	ProbationMonths int
	ProbationSalary  float64
	StartDate       time.Time
	EndDate         *time.Time
	ContractType    string
	SignDate        time.Time
}

// GenerateContractPDF 生成劳动合同 PDF（中文内容 + 中文字体）
// 支持3种合同类型：fixed_term（劳动合同）、intern（实习协议）、indefinite（兼职合同）
func GenerateContractPDF(data *ContractPDFData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetMargins(20, 15, 20)
	pdf.SetAutoPageBreak(true, 15)

	// 注册中文字体
	regularFont, err := fontFiles.ReadFile("fonts/NotoSansSC-Regular.ttf")
	if err != nil {
		return nil, fmt.Errorf("read regular font: %w", err)
	}
	pdf.AddUTF8FontFromBytes("NotoSansSC", "", regularFont)
	pdf.AddUTF8FontFromBytes("NotoSansSC", "B", regularFont)

	// 根据合同类型选择模板
	switch data.ContractType {
	case ContractTypeIntern:
		generateInternPDF(pdf, data)
	case ContractTypeIndefinite:
		generateIndefinitePDF(pdf, data)
	default:
		generateFixedTermPDF(pdf, data)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("generate PDF: %w", err)
	}
	return buf.Bytes(), nil
}

// generateFixedTermPDF 生成固定期限劳动合同（劳动合同）
func generateFixedTermPDF(pdf *fpdf.Fpdf, data *ContractPDFData) {
	// 标题
	pdf.SetFont("NotoSansSC", "B", 18)
	pdf.SetX(20)
	pdf.CellFormat(170, 12, "劳动合同", "", 0, "C", false, 0, "")
	pdf.Ln(14)

	// 副标题
	pdf.SetFont("NotoSansSC", "", 10)
	pdf.SetX(20)
	pdf.CellFormat(170, 6, "（固定期限）", "", 0, "C", false, 0, "")
	pdf.Ln(12)

	// 甲乙双方信息
	pdf.SetFont("NotoSansSC", "B", 11)
	pdf.SetX(20)
	pdf.CellFormat(170, 7, "甲方（用人单位）", "", 0, "L", false, 0, "")
	pdf.Ln(7)

	pdf.SetFont("NotoSansSC", "", 11)
	lineHeight := 7.0

	pdf.SetX(25)
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("用人单位名称：%s", data.OrgName), "", 1, "L", false, 0, "")
	pdf.SetX(25)
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("统一社会信用代码：%s", data.CreditCode), "", 1, "L", false, 0, "")
	pdf.Ln(3)

	pdf.SetFont("NotoSansSC", "B", 11)
	pdf.SetX(20)
	pdf.CellFormat(170, 7, "乙方（劳动者）", "", 0, "L", false, 0, "")
	pdf.Ln(7)

	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("姓    名：%s", data.EmployeeName), "", 1, "L", false, 0, "")
	pdf.SetX(25)
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("身份证号码：%s", data.IDCard), "", 1, "L", false, 0, "")
	pdf.Ln(6)

	// 合同条款
	pdf.SetFont("NotoSansSC", "B", 11)
	pdf.SetX(20)
	pdf.CellFormat(170, 7, "根据《中华人民共和国劳动合同法》及相关法律法规，甲乙双方遵循合法、公平、平等自愿、协商一致、诚实信用的原则，订立本合同。", "", 1, "L", false, 0, "")
	pdf.Ln(3)

	addArticleHeading(pdf, "第一条", "合同期限")
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	text := fmt.Sprintf("本合同期限自 %s 起至 %s 止。",
		data.StartDate.Format("2006年01月02日"),
		formatEndDate(data.EndDate))
	pdf.MultiCell(165, lineHeight, text, "", "L", false)
	pdf.Ln(2)

	addArticleHeading(pdf, "第二条", "工作内容和工作地点")
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.MultiCell(165, lineHeight,
		fmt.Sprintf("乙方同意在 %s 从事 %s 工作，具体工作内容和要求由双方另行约定。", data.City, data.Position),
		"", "L", false)
	pdf.Ln(2)

	addArticleHeading(pdf, "第三条", "工作时间和休息休假")
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.MultiCell(165, lineHeight,
		"乙方的工作时间、休息休假按照国家有关规定执行。",
		"", "L", false)
	pdf.Ln(2)

	addArticleHeading(pdf, "第四条", "劳动报酬")
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.MultiCell(165, lineHeight,
		fmt.Sprintf("乙方月工资标准为 %.2f 元（人民币）。", data.Salary),
		"", "L", false)
	pdf.SetX(25)
	pdf.MultiCell(165, lineHeight,
		"工资支付方式及日期按照甲方工资支付制度执行。",
		"", "L", false)
	pdf.Ln(2)

addArticleHeading(pdf, "第五条", "试用期")
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	if data.ProbationMonths > 0 {
		pdf.MultiCell(165, lineHeight,
			fmt.Sprintf("试用期自合同生效之日起%d个月。", data.ProbationMonths),
			"", "L", false)
		if data.ProbationSalary > 0 {
			pdf.MultiCell(165, lineHeight,
				fmt.Sprintf("试用期工资按照%.2f元/月执行，不低于合同约定工资的80%%。", data.ProbationSalary),
				"", "L", false)
		} else {
			pdf.MultiCell(165, lineHeight,
				fmt.Sprintf("试用期工资按照%.2f元/月执行。", data.Salary),
				"", "L", false)
		}
	} else {
		pdf.MultiCell(165, lineHeight,
			"本合同不设试用期。",
			"", "L", false)
	}
	pdf.Ln(2)

	addArticleHeading(pdf, "第六条", "社会保险")
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.MultiCell(165, lineHeight,
		"甲方依法为乙方缴纳社会保险，双方按国家规定各自承担相应费用。",
		"", "L", false)
	pdf.Ln(2)

	addArticleHeading(pdf, "第八条", "劳动保护和工作条件")
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.MultiCell(165, lineHeight,
		"甲方为乙方提供符合国家规定的劳动安全卫生条件和必要的劳动防护用品。",
		"", "L", false)
	pdf.Ln(2)

	addArticleHeading(pdf, "第九条", "合同生效")
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.MultiCell(165, lineHeight,
		"本合同一式两份，甲乙双方各执一份，自双方签字（或盖章）之日起生效。",
		"", "L", false)
	pdf.Ln(10)

	// 签章区域
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.CellFormat(70, lineHeight, fmt.Sprintf("甲方（盖章）：________________"), "", 0, "L", false, 0, "")
	pdf.CellFormat(60, lineHeight, "", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("乙方（签字）：________________"), "", 1, "L", false, 0, "")
	pdf.Ln(4)
	pdf.SetX(25)
	pdf.CellFormat(70, lineHeight, fmt.Sprintf("签订日期：%s", data.SignDate.Format("2006年01月02日")), "", 0, "L", false, 0, "")
	pdf.Ln(15)

	// 页脚
	addFooter(pdf)
}

// generateInternPDF 生成实习协议
func generateInternPDF(pdf *fpdf.Fpdf, data *ContractPDFData) {
	// 标题
	pdf.SetFont("NotoSansSC", "B", 18)
	pdf.SetX(20)
	pdf.CellFormat(170, 12, "实习协议", "", 0, "C", false, 0, "")
	pdf.Ln(14)

	pdf.SetFont("NotoSansSC", "", 10)
	pdf.SetX(20)
	pdf.CellFormat(170, 6, "（实习生）", "", 0, "C", false, 0, "")
	pdf.Ln(12)

	lineHeight := 7.0

	// 甲乙双方
	pdf.SetFont("NotoSansSC", "B", 11)
	pdf.SetX(20)
	pdf.CellFormat(170, 7, "甲方（用人单位）", "", 0, "L", false, 0, "")
	pdf.Ln(7)
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("用人单位名称：%s", data.OrgName), "", 1, "L", false, 0, "")
	pdf.SetX(25)
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("统一社会信用代码：%s", data.CreditCode), "", 1, "L", false, 0, "")
	pdf.Ln(3)

	pdf.SetFont("NotoSansSC", "B", 11)
	pdf.SetX(20)
	pdf.CellFormat(170, 7, "乙方（实习生）", "", 0, "L", false, 0, "")
	pdf.Ln(7)
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("姓    名：%s", data.EmployeeName), "", 1, "L", false, 0, "")
	pdf.SetX(25)
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("身份证号码：%s", data.IDCard), "", 1, "L", false, 0, "")
	pdf.Ln(6)

	pdf.SetFont("NotoSansSC", "B", 11)
	pdf.SetX(20)
	pdf.CellFormat(170, 7, "根据相关法律法规，甲乙双方经协商一致，签订本实习协议。", "", 1, "L", false, 0, "")
	pdf.Ln(3)

	addArticleHeading(pdf, "一", "实习期限")
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.MultiCell(165, lineHeight,
		fmt.Sprintf("实习期自 %s 起至 %s 止。",
			data.StartDate.Format("2006年01月02日"),
			formatEndDate(data.EndDate)),
		"", "L", false)
	pdf.Ln(2)

	addArticleHeading(pdf, "二", "实习内容")
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.MultiCell(165, lineHeight,
		fmt.Sprintf("乙方在 %s 从事 %s 岗位实习，具体实习内容由甲方安排。", data.City, data.Position),
		"", "L", false)
	pdf.Ln(2)

	addArticleHeading(pdf, "三", "实习补贴")
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.MultiCell(165, lineHeight,
		fmt.Sprintf("实习期间，乙方享有实习补贴 %.2f 元/月。", data.Salary),
		"", "L", false)
	pdf.Ln(2)

	addArticleHeading(pdf, "四", "工作时间与保险")
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.MultiCell(165, lineHeight,
		"实习期间的工作时间按照国家相关规定执行。甲方为乙方购买人身意外伤害保险。",
		"", "L", false)
	pdf.Ln(2)

	addArticleHeading(pdf, "五", "协议解除")
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.MultiCell(165, lineHeight,
		"甲乙双方均可提前三日通知对方解除本协议。",
		"", "L", false)
	pdf.Ln(2)

	addArticleHeading(pdf, "六", "其他约定")
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.MultiCell(165, lineHeight,
		"本协议一式两份，甲乙双方各执一份，自双方签字之日起生效。",
		"", "L", false)
	pdf.Ln(10)

	// 签章区域
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.CellFormat(70, lineHeight, fmt.Sprintf("甲方（盖章）：________________"), "", 0, "L", false, 0, "")
	pdf.CellFormat(60, lineHeight, "", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("乙方（签字）：________________"), "", 1, "L", false, 0, "")
	pdf.Ln(4)
	pdf.SetX(25)
	pdf.CellFormat(70, lineHeight, fmt.Sprintf("签订日期：%s", data.SignDate.Format("2006年01月02日")), "", 0, "L", false, 0, "")
	pdf.Ln(15)

	addFooter(pdf)
}

// generateIndefinitePDF 生成兼职合同（非全日制）
func generateIndefinitePDF(pdf *fpdf.Fpdf, data *ContractPDFData) {
	// 标题
	pdf.SetFont("NotoSansSC", "B", 18)
	pdf.SetX(20)
	pdf.CellFormat(170, 12, "兼职劳动合同", "", 0, "C", false, 0, "")
	pdf.Ln(14)

	pdf.SetFont("NotoSansSC", "", 10)
	pdf.SetX(20)
	pdf.CellFormat(170, 6, "（非全日制用工）", "", 0, "C", false, 0, "")
	pdf.Ln(12)

	lineHeight := 7.0

	// 甲乙双方
	pdf.SetFont("NotoSansSC", "B", 11)
	pdf.SetX(20)
	pdf.CellFormat(170, 7, "甲方（用人单位）", "", 0, "L", false, 0, "")
	pdf.Ln(7)
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("用人单位名称：%s", data.OrgName), "", 1, "L", false, 0, "")
	pdf.SetX(25)
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("统一社会信用代码：%s", data.CreditCode), "", 1, "L", false, 0, "")
	pdf.Ln(3)

	pdf.SetFont("NotoSansSC", "B", 11)
	pdf.SetX(20)
	pdf.CellFormat(170, 7, "乙方（劳动者）", "", 0, "L", false, 0, "")
	pdf.Ln(7)
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("姓    名：%s", data.EmployeeName), "", 1, "L", false, 0, "")
	pdf.SetX(25)
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("身份证号码：%s", data.IDCard), "", 1, "L", false, 0, "")
	pdf.Ln(6)

	pdf.SetFont("NotoSansSC", "B", 11)
	pdf.SetX(20)
	pdf.CellFormat(170, 7, "根据《中华人民共和国劳动合同法》及相关法律法规，甲乙双方经协商一致，签订本合同。", "", 1, "L", false, 0, "")
	pdf.Ln(3)

	addArticleHeading(pdf, "一", "工作内容")
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.MultiCell(165, lineHeight,
		fmt.Sprintf("乙方同意在 %s 从事 %s 工作。", data.City, data.Position),
		"", "L", false)
	pdf.Ln(2)

	addArticleHeading(pdf, "二", "工作时间")
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.MultiCell(165, lineHeight,
		"非全日制用工，每日工作时间不超过4小时，每周工作时间累计不超过24小时。",
		"", "L", false)
	pdf.Ln(2)

	addArticleHeading(pdf, "三", "劳务报酬")
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.MultiCell(165, lineHeight,
		fmt.Sprintf("劳务报酬按 %.2f 元/小时标准计算，按日或按周支付。", data.Salary),
		"", "L", false)
	pdf.Ln(2)

	addArticleHeading(pdf, "四", "合同期限")
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.MultiCell(165, lineHeight,
		fmt.Sprintf("本合同为非全日制劳动合同，期限自 %s 起。", data.StartDate.Format("2006年01月02日")),
		"", "L", false)
	pdf.Ln(2)

	addArticleHeading(pdf, "五", "社会保险")
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.MultiCell(165, lineHeight,
		"甲方应当为乙方缴纳工伤保险，其他社会保险由乙方自行缴纳。",
		"", "L", false)
	pdf.Ln(2)

	addArticleHeading(pdf, "六", "合同终止")
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.MultiCell(165, lineHeight,
		"甲乙双方可随时通知对方终止用工，终止时甲方应当支付乙方相应的报酬。",
		"", "L", false)
	pdf.Ln(2)

	addArticleHeading(pdf, "七", "其他约定")
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.MultiCell(165, lineHeight,
		"本协议一式两份，甲乙双方各执一份，自双方签字（或盖章）之日起生效。",
		"", "L", false)
	pdf.Ln(10)

	// 签章区域
	pdf.SetFont("NotoSansSC", "", 11)
	pdf.SetX(25)
	pdf.CellFormat(70, lineHeight, fmt.Sprintf("甲方（盖章）：________________"), "", 0, "L", false, 0, "")
	pdf.CellFormat(60, lineHeight, "", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, lineHeight, fmt.Sprintf("乙方（签字）：________________"), "", 1, "L", false, 0, "")
	pdf.Ln(4)
	pdf.SetX(25)
	pdf.CellFormat(70, lineHeight, fmt.Sprintf("签订日期：%s", data.SignDate.Format("2006年01月02日")), "", 0, "L", false, 0, "")
	pdf.Ln(15)

	addFooter(pdf)
}

// addArticleHeading 添加条款标题
func addArticleHeading(pdf *fpdf.Fpdf, num, title string) {
	pdf.SetFont("NotoSansSC", "B", 11)
	pdf.SetX(20)
	pdf.CellFormat(0, 7, fmt.Sprintf("%s %s", num, title), "", 1, "L", false, 0, "")
}

// formatEndDate 格式化结束日期，无固定期限返回"无固定期限"
func formatEndDate(endDate *time.Time) string {
	if endDate == nil {
		return "无固定期限"
	}
	return endDate.Format("2006年01月02日")
}

// addFooter 添加页脚
func addFooter(pdf *fpdf.Fpdf) {
	pdf.SetFont("NotoSansSC", "", 9)
	pdf.SetX(20)
	pdf.CellFormat(170, 5, "本合同由易人事（EasyHR）电子合同系统生成，仅用于合同存档参考，具体条款以双方书面协议为准。", "", 0, "C", false, 0, "")
}
