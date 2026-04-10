package finance

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wencai/easyhr/internal/common/model"
)

// InvoiceService handles business logic for invoices.
type InvoiceService struct {
	invoiceRepo  *InvoiceRepository
	voucherRepo  *VoucherRepository
}

// NewInvoiceService creates a new InvoiceService.
func NewInvoiceService(invoiceRepo *InvoiceRepository, voucherRepo *VoucherRepository) *InvoiceService {
	return &InvoiceService{
		invoiceRepo: invoiceRepo,
		voucherRepo: voucherRepo,
	}
}

// CreateInvoice creates a new invoice.
func (s *InvoiceService) CreateInvoice(orgID, userID int64, req *CreateInvoiceRequest) (*Invoice, error) {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, &FinanceError{Code: 60230, Err: fmt.Errorf("金额格式错误: %s", req.Amount)}
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, &FinanceError{Code: 60230, Err: fmt.Errorf("金额必须大于零")}
	}

	taxRate, err := decimal.NewFromString(req.TaxRate)
	if err != nil {
		return nil, &FinanceError{Code: 60231, Err: fmt.Errorf("税率格式错误: %s", req.TaxRate)}
	}

	// Compute tax amount: amount / (1 + tax_rate) * tax_rate
	taxAmount := amount.Div(taxRate.Add(decimal.NewFromInt(1))).Mul(taxRate)

	invoiceDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, &FinanceError{Code: 60232, Err: fmt.Errorf("日期格式错误，请使用 YYYY-MM-DD")}
	}

	status := InvoiceStatus(req.Status)
	if status == "" {
		status = InvoiceStatusUnverified
	}

	invoice := &Invoice{
		BaseModel:   model.BaseModel{OrgID: orgID, CreatedBy: userID, UpdatedBy: userID},
		InvoiceType: req.InvoiceType,
		Code:        req.Code,
		Number:      req.Number,
		Date:        invoiceDate,
		Amount:      amount,
		TaxRate:     taxRate,
		TaxAmount:   taxAmount,
		Status:      status,
		Remark:      req.Remark,
	}

	if err := s.invoiceRepo.Create(invoice); err != nil {
		return nil, fmt.Errorf("保存发票失败: %w", err)
	}
	return invoice, nil
}

// UpdateInvoice updates an existing invoice.
func (s *InvoiceService) UpdateInvoice(orgID, userID, invoiceID int64, req *UpdateInvoiceRequest) (*Invoice, error) {
	invoice, err := s.invoiceRepo.GetByID(orgID, invoiceID)
	if err != nil {
		return nil, &FinanceError{Code: CodeVoucherNotFound, Err: fmt.Errorf("发票不存在或无权访问")}
	}

	if req.Code != nil {
		invoice.Code = *req.Code
	}
	if req.Number != nil {
		invoice.Number = *req.Number
	}
	if req.Date != nil {
		dt, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			return nil, &FinanceError{Code: 60232, Err: fmt.Errorf("日期格式错误，请使用 YYYY-MM-DD")}
		}
		invoice.Date = dt
	}
	if req.Amount != nil {
		amount, err := decimal.NewFromString(*req.Amount)
		if err != nil {
			return nil, &FinanceError{Code: 60230, Err: fmt.Errorf("金额格式错误: %s", *req.Amount)}
		}
		invoice.Amount = amount
		// Recalculate tax amount
		invoice.TaxAmount = amount.Div(invoice.TaxRate.Add(decimal.NewFromInt(1))).Mul(invoice.TaxRate)
	}
	if req.TaxRate != nil {
		taxRate, err := decimal.NewFromString(*req.TaxRate)
		if err != nil {
			return nil, &FinanceError{Code: 60231, Err: fmt.Errorf("税率格式错误: %s", *req.TaxRate)}
		}
		invoice.TaxRate = taxRate
		// Recalculate tax amount
		invoice.TaxAmount = invoice.Amount.Div(taxRate.Add(decimal.NewFromInt(1))).Mul(taxRate)
	}
	if req.Status != nil {
		invoice.Status = *req.Status
	}
	if req.Remark != nil {
		invoice.Remark = *req.Remark
	}

	invoice.UpdatedBy = userID
	if err := s.invoiceRepo.Update(invoice); err != nil {
		return nil, fmt.Errorf("更新发票失败: %w", err)
	}
	return invoice, nil
}

// GetInvoice returns an invoice by ID.
func (s *InvoiceService) GetInvoice(orgID, invoiceID int64) (*Invoice, error) {
	return s.invoiceRepo.GetByID(orgID, invoiceID)
}

// LinkToVoucher links an invoice to a voucher.
func (s *InvoiceService) LinkToVoucher(orgID int64, invoiceID, voucherID int64) error {
	invoice, err := s.invoiceRepo.GetByID(orgID, invoiceID)
	if err != nil {
		return &FinanceError{Code: CodeVoucherNotFound, Err: fmt.Errorf("发票不存在或无权访问")}
	}
	if invoice.VoucherID != nil && *invoice.VoucherID != 0 {
		return &FinanceError{Code: CodeInvalidStatus, Err: fmt.Errorf("发票已关联凭证，不能重复关联")}
	}
	return s.invoiceRepo.LinkVoucher(orgID, invoiceID, voucherID)
}

// ListInvoices returns paginated invoices with filters.
func (s *InvoiceService) ListInvoices(orgID int64, req *ListInvoiceRequest) ([]Invoice, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}
	return s.invoiceRepo.List(orgID, req)
}

// GetMonthlySummary returns VAT summary for a given year/month.
// Groups by invoice_type and sums tax_amount for verified invoices (D-22).
func (s *InvoiceService) GetMonthlySummary(orgID int64, year, month int) (*MonthlyTaxSummary, error) {
	outputTax, inputTax, outputAmount, inputAmount, err := s.invoiceRepo.GetMonthlyTaxSummary(orgID, year, month)
	if err != nil {
		return nil, fmt.Errorf("查询月度汇总失败: %w", err)
	}
	netVAT := outputTax.Sub(inputTax)
	return &MonthlyTaxSummary{
		Year:            year,
		Month:           month,
		OutputTaxSum:    outputTax.String(),
		InputTaxSum:     inputTax.String(),
		OutputAmountSum: outputAmount.String(),
		InputAmountSum:  inputAmount.String(),
		NetVAT:          netVAT.String(),
	}, nil
}
