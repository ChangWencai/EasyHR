package tax

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
// 占位实现，将在 Task 2 中完善
func generateTaxCertificatePDF(data *TaxCertificateData) ([]byte, error) {
	// TODO: Task 2 完整实现
	return []byte{}, nil
}
