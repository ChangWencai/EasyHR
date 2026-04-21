package wxmp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wencai/easyhr/internal/city"
	"github.com/wencai/easyhr/internal/common/crypto"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/model"
	"github.com/wencai/easyhr/internal/finance"
	"github.com/wencai/easyhr/internal/salary"
	"github.com/wencai/easyhr/internal/socialinsurance"
	"gorm.io/gorm"
)

// WXMPRepositoryImpl 实现 WXMPRepository 接口
type WXMPRepositoryImpl struct {
	db       *gorm.DB
	crypto   []byte // AES key for phone decryption
	cityRepo *city.Repository
}

// NewRepository 创建 WXMPRepository 实现
func NewRepository(db *gorm.DB, aesKey string) *WXMPRepositoryImpl {
	return &WXMPRepositoryImpl{db: db, crypto: []byte(aesKey), cityRepo: city.NewRepository(db)}
}

// GetMemberByPhone 通过手机号哈希查找会员
// 关联 users 表和 employees 表，返回 MemberInfo
func (r *WXMPRepositoryImpl) GetMemberByPhone(ctx context.Context, phoneHash string) (*MemberInfo, error) {
	var user model.User
	err := r.db.Where("phone_hash = ?", phoneHash).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	var emp struct {
		ID     uint
		Name   string
		OrgID  uint
		UserID uint
	}
	err = r.db.Table("employees").
		Select("id, name, org_id, user_id").
		Where("user_id = ? AND deleted_at IS NULL", user.ID).
		Scan(&emp).Error
	if err != nil || emp.ID == 0 {
		return nil, fmt.Errorf("employee not found for user")
	}

	// 解密并脱敏手机号
	maskedPhone := ""
	if user.Phone != "" {
		decrypted, err := crypto.Decrypt(user.Phone, r.crypto)
		if err == nil {
			maskedPhone = crypto.MaskPhone(decrypted)
		}
	}

	member := &MemberInfo{
		UserID:     uint(user.ID),
		EmployeeID: emp.ID,
		OrgID:      emp.OrgID,
		Name:       emp.Name,
		Phone:      maskedPhone,
		Role:       user.Role,
		HasWechat:  false, // WechatOpenID 字段待 Phase 8 前端部分实现
	}
	return member, nil
}

// BindWechatOpenID 绑定微信 openid
func (r *WXMPRepositoryImpl) BindWechatOpenID(ctx context.Context, userID uint, openID string) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("wechat_openid", openID).Error
}

// ListPayslips 查询员工工资条列表
func (r *WXMPRepositoryImpl) ListPayslips(ctx context.Context, orgID, employeeID uint) ([]PayslipSummary, error) {
	var slips []salary.PayrollSlip
	err := r.db.Scopes(middleware.TenantScope(int64(orgID))).
		Where("employee_id = ? AND deleted_at IS NULL", employeeID).
		Order("created_at DESC").
		Find(&slips).Error
	if err != nil {
		return nil, err
	}

	var summaries []PayslipSummary
	for _, slip := range slips {
		var record salary.PayrollRecord
		if err := r.db.Scopes(middleware.TenantScope(int64(orgID))).
			Where("id = ?", slip.PayrollRecordID).First(&record).Error; err != nil {
			continue
		}
		// 映射 payroll_slip status 到 WXMP status
		status := slip.Status
		if status == salary.SlipStatusSigned {
			status = "confirmed"
		} else if status == salary.SlipStatusSent || status == salary.SlipStatusViewed {
			status = "paid"
		} else {
			status = "pending"
		}

		s := PayslipSummary{
			ID:       uint(slip.ID),
			Year:     record.Year,
			Month:    record.Month,
			GrossPay: fmt.Sprintf("%.2f", record.GrossIncome),
			NetPay:   fmt.Sprintf("%.2f", record.NetIncome),
			Status:   status,
		}
		if slip.SignedAt != nil {
			s.SignedAt = slip.SignedAt.Format("2006-01-02")
		}
		if record.PayDate != nil {
			s.PaidAt = record.PayDate.Format("2006-01-02")
		}
		summaries = append(summaries, s)
	}
	return summaries, nil
}

// GetPayslipByID 查询工资条详情（包含明细项）
func (r *WXMPRepositoryImpl) GetPayslipByID(ctx context.Context, orgID, employeeID, payslipID uint) (*PayslipDetail, error) {
	var slip salary.PayrollSlip
	err := r.db.Scopes(middleware.TenantScope(int64(orgID))).
		Where("id = ? AND employee_id = ? AND deleted_at IS NULL", payslipID, employeeID).
		First(&slip).Error
	if err != nil {
		return nil, fmt.Errorf("payslip not found: %w", err)
	}

	var record salary.PayrollRecord
	if err := r.db.Scopes(middleware.TenantScope(int64(orgID))).
		Where("id = ?", slip.PayrollRecordID).First(&record).Error; err != nil {
		return nil, fmt.Errorf("payroll record not found: %w", err)
	}

	var items []salary.PayrollItem
	r.db.Scopes(middleware.TenantScope(int64(orgID))).
		Where("payroll_record_id = ?", slip.PayrollRecordID).
		Find(&items)

	var payrollItems []PayrollItemDTO
	for _, item := range items {
		payrollItems = append(payrollItems, PayrollItemDTO{
			Name:   item.ItemName,
			Type:   item.ItemType,
			Amount: fmt.Sprintf("%.2f", item.Amount),
		})
	}

		detail := &PayslipDetail{
			ID:           uint(slip.ID),
		Year:         record.Year,
		Month:        record.Month,
		GrossPay:     fmt.Sprintf("%.2f", record.GrossIncome),
		NetPay:       fmt.Sprintf("%.2f", record.NetIncome),
		Items:        payrollItems,
		SocialDeduct: fmt.Sprintf("%.2f", record.SIDeduction),
		TaxDeduct:    fmt.Sprintf("%.2f", record.Tax),
		OtherDeduct:  fmt.Sprintf("%.2f", record.TotalDeductions-record.SIDeduction-record.Tax),
	}
	if slip.SignedAt != nil {
		detail.SignedAt = slip.SignedAt.Format("2006-01-02")
	}
	if record.PayDate != nil {
		detail.PaidAt = record.PayDate.Format("2006-01-02")
	}
	return detail, nil
}

// ListContracts 查询员工合同列表
func (r *WXMPRepositoryImpl) ListContracts(ctx context.Context, orgID, employeeID uint) ([]ContractDTO, error) {
	var contracts []struct {
		ID              uint
		ContractType   string
		Status          string
		StartDate       time.Time
		EndDate         time.Time
		SignedAt        *time.Time
		ContractFileURL string
	}
	err := r.db.Table("contracts").
		Select("contracts.id, contracts.contract_type, contracts.status, contracts.start_date, contracts.end_date, contracts.signed_at, contracts.contract_file_url").
		Joins("JOIN employees ON employees.id = contracts.employee_id").
		Where("employees.id = ? AND employees.org_id = ? AND contracts.deleted_at IS NULL", employeeID, orgID).
		Find(&contracts).Error
	if err != nil {
		return nil, err
	}

	var result []ContractDTO
	for _, c := range contracts {
		dto := ContractDTO{
			ID:           c.ID,
			ContractType: c.ContractType,
			Status:       c.Status,
			StartDate:    c.StartDate.Format("2006-01-02"),
			EndDate:      c.EndDate.Format("2006-01-02"),
			PDFURL:       c.ContractFileURL,
		}
		if c.SignedAt != nil {
			dto.SignedAt = c.SignedAt.Format("2006-01-02")
		}
		result = append(result, dto)
	}
	return result, nil
}

// GetContractByID 查询合同详情
func (r *WXMPRepositoryImpl) GetContractByID(ctx context.Context, orgID, employeeID, contractID uint) (*ContractDetail, error) {
	var contract struct {
		ID              uint
		ContractType   string
		Status          string
		StartDate       time.Time
		EndDate         time.Time
		SignedAt        *time.Time
		ContractFileURL string
	}
	err := r.db.Table("contracts").
		Select("contracts.id, contracts.contract_type, contracts.status, contracts.start_date, contracts.end_date, contracts.signed_at, contracts.contract_file_url").
		Joins("JOIN employees ON employees.id = contracts.employee_id").
		Where("contracts.id = ? AND employees.id = ? AND employees.org_id = ? AND contracts.deleted_at IS NULL", contractID, employeeID, orgID).
		Scan(&contract).Error
	if err != nil {
		return nil, fmt.Errorf("contract not found: %w", err)
	}

	detail := &ContractDetail{
		ID:           contract.ID,
		ContractType: contract.ContractType,
		Status:       contract.Status,
		StartDate:    contract.StartDate.Format("2006-01-02"),
		EndDate:      contract.EndDate.Format("2006-01-02"),
		PDFURL:       contract.ContractFileURL,
	}
	if contract.SignedAt != nil {
		detail.SignedAt = contract.SignedAt.Format("2006-01-02")
	}
	return detail, nil
}

// siDetailItem 解析社保记录 Details JSON 的单个险种
type siDetailItem struct {
	Name           string  `json:"name"`
	PersonalAmount float64 `json:"personal_amount"`
}

// ListSocialInsurance 查询员工社保记录（仅展示个人缴费部分）
func (r *WXMPRepositoryImpl) ListSocialInsurance(ctx context.Context, orgID, employeeID uint) ([]SocialInsuranceDTO, error) {
	var records []socialinsurance.SocialInsuranceRecord
	err := r.db.Scopes(middleware.TenantScope(int64(orgID))).
		Where("employee_id = ? AND deleted_at IS NULL", employeeID).
		Order("start_month DESC").
		Limit(12).
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	var result []SocialInsuranceDTO
	for _, rec := range records {
		// 解析 Details JSON，提取个人缴费明细
		var pension, medical, unemployment float64
		if detailsRaw := rec.Details.String(); detailsRaw != "" && detailsRaw != "null" {
			var details []siDetailItem
			if err := json.Unmarshal([]byte(detailsRaw), &details); err == nil {
				for _, d := range details {
					switch d.Name {
					case "养老保险":
						pension = d.PersonalAmount
					case "医疗保险":
						medical = d.PersonalAmount
					case "失业保险":
						unemployment = d.PersonalAmount
					}
				}
			}
		}

		cityName := r.cityRepo.GetNameByCode(rec.CityCode)
		dto := SocialInsuranceDTO{
			PaymentMonth:  rec.StartMonth,
			City:          cityName,
			BaseAmount:    fmt.Sprintf("%.2f", rec.BaseAmount),
			Pension:       fmt.Sprintf("%.2f", pension),
			Medical:       fmt.Sprintf("%.2f", medical),
			Unemployment:  fmt.Sprintf("%.2f", unemployment),
			TotalPersonal: fmt.Sprintf("%.2f", rec.TotalPersonal),
		}
		result = append(result, dto)
	}
	return result, nil
}

// ListExpenses 查询员工报销单列表
func (r *WXMPRepositoryImpl) ListExpenses(ctx context.Context, orgID, employeeID uint) ([]ExpenseDTO, error) {
	var expenses []finance.ExpenseReimbursement
	err := r.db.Scopes(middleware.TenantScope(int64(orgID))).
		Where("employee_id = ? AND deleted_at IS NULL", employeeID).
		Order("created_at DESC").
		Find(&expenses).Error
	if err != nil {
		return nil, err
	}

	var result []ExpenseDTO
	for _, exp := range expenses {
		// 解析 Attachments JSON 字符串
		var attachments []string
		if exp.Attachments != "" {
			_ = json.Unmarshal([]byte(exp.Attachments), &attachments)
		}

		dto := ExpenseDTO{
			ID:           uint(exp.ID),
			Type:        string(exp.ExpenseType),
			Amount:      exp.Amount.String(),
			Description: exp.Description,
			Status:      string(exp.Status),
			Attachments: attachments,
			CreatedAt:   exp.CreatedAt.Format("2006-01-02"),
		}
		if exp.ApprovedAt != nil {
			dto.ApprovedAt = exp.ApprovedAt.Format("2006-01-02")
		}
		if exp.RejectedAt != nil {
			dto.RejectedAt = exp.RejectedAt.Format("2006-01-02")
			if exp.RejectedNote != "" {
				dto.RejectReason = exp.RejectedNote
			}
		}
		if exp.PaidAt != nil {
			dto.PaidAt = exp.PaidAt.Format("2006-01-02")
		}
		result = append(result, dto)
	}
	return result, nil
}

// GetExpenseByID 查询报销单详情
func (r *WXMPRepositoryImpl) GetExpenseByID(ctx context.Context, orgID, employeeID, expenseID uint) (*ExpenseDTO, error) {
	var exp finance.ExpenseReimbursement
	err := r.db.Scopes(middleware.TenantScope(int64(orgID))).
		Where("id = ? AND employee_id = ? AND deleted_at IS NULL", expenseID, employeeID).
		First(&exp).Error
	if err != nil {
		return nil, fmt.Errorf("expense not found: %w", err)
	}

	var attachments []string
	if exp.Attachments != "" {
		_ = json.Unmarshal([]byte(exp.Attachments), &attachments)
	}

	return &ExpenseDTO{
		ID:           uint(exp.ID),
		Type:        string(exp.ExpenseType),
		Amount:      exp.Amount.String(),
		Description: exp.Description,
		Status:      string(exp.Status),
		Attachments: attachments,
		CreatedAt:   exp.CreatedAt.Format("2006-01-02"),
	}, nil
}

// CreateExpense 创建报销单
func (r *WXMPRepositoryImpl) CreateExpense(ctx context.Context, orgID, employeeID uint, req *ExpenseRequest) (*ExpenseDTO, error) {
	// 验证 attachments 数量
	if len(req.Attachments) > 9 {
		return nil, fmt.Errorf("attachments exceed limit of 9")
	}

	attachmentsJSON := "[]"
	if len(req.Attachments) > 0 {
		if data, err := json.Marshal(req.Attachments); err == nil {
			attachmentsJSON = string(data)
		}
	}

	exp := finance.ExpenseReimbursement{
		BaseModel:   model.BaseModel{OrgID: int64(orgID)},
		EmployeeID:  int64(employeeID),
		Amount:      mustDecimal(req.Amount),
		Description: req.Description,
		ExpenseType:  finance.ExpenseType(req.Type),
		Attachments: attachmentsJSON,
		Status:      finance.ExpenseStatusPending,
	}

	if err := r.db.Create(&exp).Error; err != nil {
		return nil, fmt.Errorf("create expense: %w", err)
	}

	var attachments []string
	if exp.Attachments != "" {
		_ = json.Unmarshal([]byte(exp.Attachments), &attachments)
	}

	return &ExpenseDTO{
		ID:          uint(exp.ID),
		Type:        string(exp.ExpenseType),
		Amount:      exp.Amount.String(),
		Description: exp.Description,
		Status:      string(exp.Status),
		Attachments: attachments,
		CreatedAt:   exp.CreatedAt.Format("2006-01-02"),
	}, nil
}

// SignPayslip 更新工资条签收状态
func (r *WXMPRepositoryImpl) SignPayslip(ctx context.Context, orgID, employeeID, payslipID uint) error {
	now := time.Now()
	result := r.db.Model(&salary.PayrollSlip{}).
		Scopes(middleware.TenantScope(int64(orgID))).
		Where("id = ? AND employee_id = ? AND deleted_at IS NULL", payslipID, employeeID).
		Updates(map[string]interface{}{
			"signed_at": now,
			"status":    salary.SlipStatusSigned,
		})
	if result.Error != nil {
		return fmt.Errorf("sign payslip: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("payslip not found")
	}
	return nil
}

// mustDecimal 将字符串解析为 decimal.Decimal，解析失败则 panic
func mustDecimal(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero
	}
	return d
}
