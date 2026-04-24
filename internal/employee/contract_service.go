package employee

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/wencai/easyhr/internal/common/config"
	"github.com/wencai/easyhr/internal/common/crypto"
	"github.com/wencai/easyhr/internal/common/model"
	"github.com/wencai/easyhr/pkg/oss"
	"github.com/wencai/easyhr/pkg/sms"
	"gorm.io/gorm"
)

// ContractService 合同业务逻辑层
type ContractService struct {
	contractRepo *ContractRepository
	empRepo      *Repository
	db           *gorm.DB // 用于查询 Organization 信息
	cryptoCfg    config.CryptoConfig
	todoSvc      TodoCreator // interface to avoid circular import
	smsClient    *sms.Client
	ossClient    *oss.Client
}

// NewContractService 创建合同 Service
func NewContractService(
	contractRepo *ContractRepository,
	empRepo *Repository,
	db *gorm.DB,
	cryptoCfg config.CryptoConfig,
	todoSvc TodoCreator,
	smsClient *sms.Client,
	ossClient *oss.Client,
) *ContractService {
	return &ContractService{
		contractRepo: contractRepo,
		empRepo:      empRepo,
		db:           db,
		cryptoCfg:    cryptoCfg,
		todoSvc:      todoSvc,
		smsClient:    smsClient,
		ossClient:    ossClient,
	}
}

// generateSignCode 生成6位纯数字验证码
func generateSignCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// aesKey 获取 AES 密钥字节
func (s *ContractService) aesKey() []byte {
	return []byte(s.cryptoCfg.AESKey)
}

// CreateContract 创建合同（status=draft）
func (s *ContractService) CreateContract(ctx context.Context, orgID, userID, employeeID int64, req *CreateContractRequest) (*ContractResponse, error) {
	// 验证员工存在
	emp, err := s.empRepo.FindByID(orgID, employeeID)
	if err != nil {
		return nil, fmt.Errorf("员工不存在")
	}

	// 解析开始日期
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("开始日期格式错误: %w", err)
	}

	// 解析结束日期（无固定期限为 nil）
	var endDate *time.Time
	if req.EndDate != "" {
		parsed, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("结束日期格式错误: %w", err)
		}
		endDate = &parsed
	}

	contract := &Contract{}
	contract.OrgID = orgID
	contract.CreatedBy = userID
	contract.UpdatedBy = userID
	contract.EmployeeID = employeeID
	contract.ContractType = req.ContractType
	contract.StartDate = startDate
	contract.EndDate = endDate

	// 薪资：优先使用请求中的薪资，否则使用员工档案中的正式薪资
	if req.Salary != nil && *req.Salary > 0 {
		contract.Salary = *req.Salary
	} else if emp.Salary != nil {
		contract.Salary = *emp.Salary
	}

	// 试用期薪资：优先使用请求中的，否则使用员工档案中的试用期薪资
	if req.ProbationSalary != nil && *req.ProbationSalary > 0 {
		contract.ProbationSalary = *req.ProbationSalary
	} else if emp.ProbationSalary != nil {
		contract.ProbationSalary = *emp.ProbationSalary
	}

	// 试用期月数
	if req.ProbationMonths != nil {
		contract.ProbationMonths = *req.ProbationMonths
	}

	contract.Status = ContractStatusDraft

	if err := s.contractRepo.Create(contract); err != nil {
		return nil, fmt.Errorf("创建合同失败: %w", err)
	}

	return s.toContractResponse(contract, emp.Name), nil
}

// GeneratePDF 生成合同 PDF 模板
func (s *ContractService) GeneratePDF(ctx context.Context, orgID, contractID int64) ([]byte, error) {
	// 查询合同
	contract, err := s.contractRepo.FindByID(orgID, contractID)
	if err != nil {
		return nil, fmt.Errorf("合同不存在")
	}

	// 查询员工信息（含解密身份证号）
	emp, err := s.empRepo.FindByID(orgID, contract.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("关联员工不存在")
	}

	// 查询企业信息
	var org model.Organization
	if err := s.db.Where("id = ?", orgID).First(&org).Error; err != nil {
		return nil, fmt.Errorf("企业信息查询失败: %w", err)
	}

	// 解密身份证号用于 PDF
	aesKey := s.aesKey()
	idCard, _ := crypto.Decrypt(emp.IDCardEncrypted, aesKey)

	// 薪资：合同薪资 > 0 则用合同薪资，否则用员工薪资
	salary := contract.Salary
	if salary <= 0 && emp.Salary != nil {
		salary = *emp.Salary
	}
	probationSalary := contract.ProbationSalary
	if probationSalary <= 0 && emp.ProbationSalary != nil {
		probationSalary = *emp.ProbationSalary
	}

	// 构建 PDF 数据
	data := &ContractPDFData{
		OrgName:         org.Name,
		CreditCode:      org.CreditCode,
		EmployeeName:    emp.Name,
		IDCard:          idCard,
		Position:        emp.Position,
		City:            org.City,
		Salary:          salary,
		ProbationMonths: contract.ProbationMonths,
		ProbationSalary:  probationSalary,
		StartDate:       contract.StartDate,
		EndDate:         contract.EndDate,
		ContractType:    contract.ContractType,
		SignDate:        emp.HireDate,
	}

	// 生成 PDF
	pdfBytes, err := GenerateContractPDF(data)
	if err != nil {
		return nil, fmt.Errorf("生成PDF失败: %w", err)
	}

	// 更新合同状态为 pending_sign
	if err := s.contractRepo.Update(orgID, contractID, map[string]interface{}{
		"status":     ContractStatusPendingSign,
		"updated_by": contract.CreatedBy,
	}); err != nil {
		return nil, fmt.Errorf("更新合同状态失败: %w", err)
	}

	return pdfBytes, nil
}

// UploadSigned 上传签署扫描件
func (s *ContractService) UploadSigned(ctx context.Context, orgID, contractID int64, req *UploadSignedRequest) (*ContractResponse, error) {
	// 查询合同
	contract, err := s.contractRepo.FindByID(orgID, contractID)
	if err != nil {
		return nil, fmt.Errorf("合同不存在")
	}

	// 解析签署日期
	signDate, err := time.Parse("2006-01-02", req.SignDate)
	if err != nil {
		return nil, fmt.Errorf("签署日期格式错误: %w", err)
	}

	// 判断状态：如果签署日期在合同期限内（或无固定期限），状态为 active
	status := ContractStatusSigned
	if contract.EndDate == nil || !signDate.After(*contract.EndDate) {
		status = ContractStatusActive
	}

	updates := map[string]interface{}{
		"signed_pdf_url": req.SignedPDFURL,
		"sign_date":      signDate,
		"status":         status,
		"updated_by":     contract.UpdatedBy,
	}

	if err := s.contractRepo.Update(orgID, contractID, updates); err != nil {
		return nil, fmt.Errorf("更新合同失败: %w", err)
	}

	// 重新查询
	updated, err := s.contractRepo.FindByID(orgID, contractID)
	if err != nil {
		return nil, err
	}

	return s.toContractResponse(updated, ""), nil
}

// TerminateContract 终止合同
func (s *ContractService) TerminateContract(ctx context.Context, orgID, contractID int64, req *TerminateContractRequest) (*ContractResponse, error) {
	// 查询合同
	contract, err := s.contractRepo.FindByID(orgID, contractID)
	if err != nil {
		return nil, fmt.Errorf("合同不存在")
	}

	// 解析终止日期
	terminateDate, err := time.Parse("2006-01-02", req.TerminateDate)
	if err != nil {
		return nil, fmt.Errorf("终止日期格式错误: %w", err)
	}

	updates := map[string]interface{}{
		"terminate_date":   terminateDate,
		"terminate_reason": req.TerminateReason,
		"status":           ContractStatusTerminated,
		"updated_by":       contract.UpdatedBy,
	}

	if err := s.contractRepo.Update(orgID, contractID, updates); err != nil {
		return nil, fmt.Errorf("终止合同失败: %w", err)
	}

	// 重新查询
	updated, err := s.contractRepo.FindByID(orgID, contractID)
	if err != nil {
		return nil, err
	}

	return s.toContractResponse(updated, ""), nil
}

// GetContract 获取合同详情
func (s *ContractService) GetContract(ctx context.Context, orgID, contractID int64) (*ContractResponse, error) {
	contract, err := s.contractRepo.FindByID(orgID, contractID)
	if err != nil {
		return nil, fmt.Errorf("合同不存在")
	}

	// 获取员工姓名
	empName := ""
	emp, err := s.empRepo.FindByID(orgID, contract.EmployeeID)
	if err == nil {
		empName = emp.Name
	}

	return s.toContractResponse(contract, empName), nil
}

// ListByEmployee 按员工查询合同列表
func (s *ContractService) ListByEmployee(ctx context.Context, orgID, employeeID int64, page, pageSize int) ([]ContractResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	contracts, total, err := s.contractRepo.ListByEmployee(orgID, employeeID, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("查询合同列表失败: %w", err)
	}

	var responses []ContractResponse
	for _, c := range contracts {
		empName := ""
		emp, err := s.empRepo.FindByID(orgID, c.EmployeeID)
		if err == nil {
			empName = emp.Name
		}
		responses = append(responses, *s.toContractResponse(&c, empName))
	}

	return responses, total, nil
}

// ListContracts 查询企业所有合同
func (s *ContractService) ListContracts(ctx context.Context, orgID int64, status string, page, pageSize int) ([]ContractResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	contracts, total, err := s.contractRepo.List(orgID, status, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("查询合同列表失败: %w", err)
	}

	var responses []ContractResponse
	for _, c := range contracts {
		empName := ""
		emp, err := s.empRepo.FindByID(orgID, c.EmployeeID)
		if err == nil {
			empName = emp.Name
		}
		responses = append(responses, *s.toContractResponse(&c, empName))
	}

	return responses, total, nil
}

// UpdateContract 更新合同信息（仅草稿状态可编辑）
func (s *ContractService) UpdateContract(ctx context.Context, orgID, userID, contractID int64, req *UpdateContractRequest) (*ContractResponse, error) {
	contract, err := s.contractRepo.FindByID(orgID, contractID)
	if err != nil {
		return nil, fmt.Errorf("合同不存在")
	}

	if contract.Status != ContractStatusDraft && contract.Status != ContractStatusPendingSign {
		return nil, fmt.Errorf("仅草稿或待签状态合同可编辑")
	}

	updates := make(map[string]interface{})
	updates["updated_by"] = userID

	if req.ContractType != nil {
		updates["contract_type"] = *req.ContractType
	}
	if req.StartDate != nil {
		startDate, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("开始日期格式错误: %w", err)
		}
		updates["start_date"] = startDate
	}
	if req.EndDate != nil {
		if *req.EndDate == "" {
			updates["end_date"] = nil
		} else {
			endDate, err := time.Parse("2006-01-02", *req.EndDate)
			if err != nil {
				return nil, fmt.Errorf("结束日期格式错误: %w", err)
			}
			updates["end_date"] = endDate
		}
	}
	if req.Salary != nil {
		updates["salary"] = *req.Salary
	}
	if req.ProbationMonths != nil {
		updates["probation_months"] = *req.ProbationMonths
	}
	if req.ProbationSalary != nil {
		updates["probation_salary"] = *req.ProbationSalary
	}

	if err := s.contractRepo.Update(orgID, contractID, updates); err != nil {
		return nil, fmt.Errorf("更新合同失败: %w", err)
	}

	return s.GetContract(ctx, orgID, contractID)
}

// DeleteContract 删除合同（仅草稿/待签状态可删除）
func (s *ContractService) DeleteContract(ctx context.Context, orgID, contractID int64) error {
	contract, err := s.contractRepo.FindByID(orgID, contractID)
	if err != nil {
		return fmt.Errorf("合同不存在")
	}

	if contract.Status != ContractStatusDraft && contract.Status != ContractStatusPendingSign {
		return fmt.Errorf("仅草稿或待签状态的合同可删除")
	}

	return s.contractRepo.Delete(orgID, contractID)
}

// formatDate converts time.Time to YYYY-MM-DD string
func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// formatDatePtr converts *time.Time to *string (YYYY-MM-DD)
func formatDatePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}

// toContractResponse 将 Contract 转换为 ContractResponse
func (s *ContractService) toContractResponse(c *Contract, empName string) *ContractResponse {
	resp := &ContractResponse{
		ID:               c.ID,
		EmployeeID:       c.EmployeeID,
		EmployeeName:     empName,
		ContractType:     c.ContractType,
		StartDate:        formatDate(c.StartDate),
		EndDate:          formatDatePtr(c.EndDate),
		Salary:           c.Salary,
		ProbationMonths:  c.ProbationMonths,
		ProbationSalary:  c.ProbationSalary,
		Status:           c.Status,
		PDFURL:           c.PDFURL,
		SignedPDFURL:     c.SignedPDFURL,
		SignDate:         formatDatePtr(c.SignDate),
		TerminateDate:    formatDatePtr(c.TerminateDate),
		TerminateReason:  c.TerminateReason,
		CreatedAt:        c.CreatedAt.Format(time.RFC3339),
	}
	return resp
}

// CheckContractRenewalReminders scans contracts expiring within 30 days and creates renewal todos.
// Called by todo scheduler daily at 02:05 CST.
func (s *ContractService) CheckContractRenewalReminders(ctx context.Context) error {
	now := time.Now()
	deadline := now.AddDate(0, 1, 0) // 30 days ahead

	contracts, err := s.FindContractsExpiringSoon(ctx, now, deadline)
	if err != nil {
		return fmt.Errorf("find expiring contracts: %w", err)
	}

	for _, contract := range contracts {
		if s.todoSvc != nil {
			contractID := contract.ID
			exists, _ := s.todoSvc.ExistsBySource(ctx, contract.OrgID, "contract", &contractID)
			if exists {
				continue
			}
			emp, _ := s.empRepo.FindByID(contract.OrgID, contract.EmployeeID)
			empName := ""
			var empID *int64
			if emp != nil {
				empName = emp.Name
				id := emp.ID
				empID = &id
			}

			_ = s.todoSvc.CreateTodoFromEmployee(
				contract.OrgID,
				fmt.Sprintf("员工 %s 的劳动合同即将到期（%s），请及时处理续签", empName, contract.EndDate.Format("2006-01-02")),
				"contract_renew",
				empID,
				empName,
				contract.EndDate,
				"contract",
				&contractID,
			)
		}
	}

	return nil
}

// FindContractsExpiringSoon finds contracts expiring within the given date range.
func (s *ContractService) FindContractsExpiringSoon(ctx context.Context, now, deadline time.Time) ([]Contract, error) {
	var contracts []Contract
	err := s.db.Model(&Contract{}).
		Where("status = ? AND end_date IS NOT NULL AND end_date >= ? AND end_date <= ?",
			ContractStatusActive, now, deadline).
		Find(&contracts).Error
	return contracts, err
}

// SendSignCode 生成签署验证码并发送短信
func (s *ContractService) SendSignCode(ctx context.Context, contractID int64, phone string) error {
	// 租户隔离：先通过手机号找到员工，再确认合同属于该员工
	emp, err := s.empRepo.FindByPhoneHashGlobal(crypto.HashSHA256(phone))
	if err != nil || emp == nil {
		return fmt.Errorf("该手机号未关联任何员工")
	}

	// 验证合同存在且属于该员工
	contract, err := s.contractRepo.FindByID(emp.OrgID, contractID)
	if err != nil {
		return fmt.Errorf("合同不存在")
	}
	if contract.EmployeeID != emp.ID {
		return fmt.Errorf("合同不属于该员工")
	}
	if contract.Status != ContractStatusDraft && contract.Status != ContractStatusPendingSign {
		return fmt.Errorf("合同状态不允许发起签署")
	}

	// 生成6位验证码
	code, err := generateSignCode()
	if err != nil {
		return fmt.Errorf("生成验证码失败: %w", err)
	}

	// 存储验证码（覆盖同合同+手机号的旧记录）
	signCode := &ContractSignCode{
		ContractID: contractID,
		Phone:     phone,
		Code:      code,
		ExpiresAt: time.Now().Add(SignCodeExpiry),
	}
	if err := s.contractRepo.UpsertSignCode(signCode); err != nil {
		return fmt.Errorf("存储验证码失败: %w", err)
	}

	// 发送短信
	if s.smsClient != nil {
		templateCode := os.Getenv("ALIYUN_SMS_CONTRACT_TEMPLATE_CODE")
		if templateCode != "" {
			signToken, _ := generateToken()
			signLink := fmt.Sprintf("%s/sign/%d?token=%s",
				os.Getenv("APP_BASE_URL"), contractID, signToken)
			templateParam := fmt.Sprintf(`{"name":"%s","link":"%s","days":"7"}`, emp.Name, signLink)
			_ = s.smsClient.SendTemplateMessage(ctx, phone, templateCode, templateParam)
		}
	}

	return nil
}

// VerifySignCode 校验签署验证码
func (s *ContractService) VerifySignCode(ctx context.Context, contractID int64, phone, code string) (*VerifySignCodeResponse, error) {
	// 租户隔离：先通过手机号找到员工
	emp, err := s.empRepo.FindByPhoneHashGlobal(crypto.HashSHA256(phone))
	if err != nil || emp == nil {
		return nil, fmt.Errorf("该手机号未关联任何员工")
	}

	// 查询最新验证码记录
	signCode, err := s.contractRepo.FindLatestSignCode(contractID, phone)
	if err != nil {
		return nil, fmt.Errorf("验证码错误，请重新获取")
	}

	// 校验有效期
	if time.Now().After(signCode.ExpiresAt) {
		return nil, fmt.Errorf("验证码已过期，请重新获取")
	}

	// 校验验证码
	if signCode.Code != code {
		return nil, fmt.Errorf("验证码错误，请重新输入")
	}

	// 生成 SignToken（用于 ConfirmSign）
	signToken, _ := generateToken()
	signCode.SignToken = signToken
	signCode.Verified = true
	signCode.ExpiresAt = time.Now().Add(SignTokenExpiry)
	if err := s.contractRepo.UpdateSignCode(signCode); err != nil {
		return nil, fmt.Errorf("更新验证码失败: %w", err)
	}

	// 获取合同详情用于前端展示
	contract, _ := s.contractRepo.FindByID(emp.OrgID, contractID)
	foundEmp, _ := s.empRepo.FindByID(emp.OrgID, contract.EmployeeID)
	var org model.Organization
	s.db.Where("id = ?", contract.OrgID).First(&org)

	endDateStr := ""
	if contract.EndDate != nil {
		endDateStr = contract.EndDate.Format("2006-01-02")
	}

	return &VerifySignCodeResponse{
		SignToken:    signToken,
		ExpiresIn:    int(SignTokenExpiry.Seconds()),
		EmployeeName: foundEmp.Name,
		ContractType: contract.ContractType,
		StartDate:    contract.StartDate.Format("2006-01-02"),
		EndDate:      endDateStr,
		OrgName:      org.Name,
	}, nil
}

// ConfirmSign 确认签署（通过 SignToken 验证，无需再验证手机号）
func (s *ContractService) ConfirmSign(ctx context.Context, contractID int64, signToken string) (*ConfirmSignResponse, error) {
	// 查找 SignToken
	signCode, err := s.contractRepo.FindBySignToken(signToken)
	if err != nil {
		return nil, fmt.Errorf("签署验证失败，请重新验证")
	}
	if signCode.ContractID != contractID {
		return nil, fmt.Errorf("签署验证失败")
	}
	if time.Now().After(signCode.ExpiresAt) {
		return nil, fmt.Errorf("签署已超时，请重新验证")
	}

	// 查询合同（通过 phone 关联查询员工 orgID）
	emp, _ := s.empRepo.FindByPhoneHashGlobal(crypto.HashSHA256(signCode.Phone))
	orgID := int64(0)
	if emp != nil {
		orgID = emp.OrgID
	}
	contract, err := s.contractRepo.FindByID(orgID, contractID)
	if err != nil {
		return nil, fmt.Errorf("合同不存在")
	}
	if contract.Status == ContractStatusSigned || contract.Status == ContractStatusActive {
		return nil, fmt.Errorf("该合同已完成签署")
	}

	// 更新合同状态
	now := time.Now()
	signedPdfUrl := contract.PDFURL

	status := ContractStatusSigned
	if contract.EndDate == nil || !now.After(*contract.EndDate) {
		status = ContractStatusActive
	}

	if err := s.contractRepo.Update(orgID, contractID, map[string]interface{}{
		"status":         status,
		"sign_date":       now,
		"signed_pdf_url": signedPdfUrl,
	}); err != nil {
		return nil, fmt.Errorf("更新合同状态失败: %w", err)
	}

	return &ConfirmSignResponse{
		SignedPDFURL: signedPdfUrl,
		Message:     "签署成功",
	}, nil
}

// FindPendingSignContracts 查找已发起签署但3天内未签的合同
func (s *ContractService) FindPendingSignContracts(ctx context.Context) ([]Contract, error) {
	var contracts []Contract
	threeDaysAgo := time.Now().Add(-3 * 24 * time.Hour)
	err := s.db.Model(&Contract{}).
		Where("status = ? AND created_at <= ?",
			ContractStatusPendingSign, threeDaysAgo).
		Find(&contracts).Error
	return contracts, err
}

// CheckPendingSignReminders 扫描 pending_sign 超3天未签的合同，给老板发待办提醒
func (s *ContractService) CheckPendingSignReminders(ctx context.Context) error {
	contracts, err := s.FindPendingSignContracts(ctx)
	if err != nil {
		return err
	}

	for _, contract := range contracts {
		if s.todoSvc != nil {
			contractID := contract.ID
			exists, _ := s.todoSvc.ExistsBySource(ctx, contract.OrgID, "contract_pending_sign", &contractID)
			if exists {
				continue // 已提醒过
			}
			emp, _ := s.empRepo.FindByID(contract.OrgID, contract.EmployeeID)
			empName := ""
			var empID *int64
			if emp != nil {
				empName = emp.Name
				id := emp.ID
				empID = &id
			}
			// 计算发起天数
			daysPending := int(time.Since(contract.CreatedAt).Hours() / 24)

			_ = s.todoSvc.CreateTodoFromEmployee(
				contract.OrgID,
				fmt.Sprintf("员工 %s 的合同已发起签署 %d 天，员工尚未签署，请跟进", empName, daysPending),
				"contract_pending_sign",
				empID,
				empName,
				nil,
				"contract",
				&contractID,
			)
		}
	}
	return nil
}

// GetSignedPdfURL 获取已签合同 PDF URL
func (s *ContractService) GetSignedPdfURL(ctx context.Context, contractID int64) (string, error) {
	contract, err := s.contractRepo.FindByID(0, contractID)
	if err != nil {
		return "", fmt.Errorf("合同不存在或尚未签署")
	}
	url := contract.SignedPDFURL
	if url == "" {
		url = contract.PDFURL
	}
	if url == "" {
		return "", fmt.Errorf("合同尚未签署")
	}
	// 如果是 OSS key，生成签名 URL
	if s.ossClient != nil && !strings.HasPrefix(url, "http") {
		signedUrl, err := s.ossClient.GenerateGetURL(ctx, url, 1*time.Hour)
		if err == nil {
			return signedUrl, nil
		}
	}
	return url, nil
}

// SendSignLink 老板发起签署：生成PDF + 上传OSS + 发送短信
func (s *ContractService) SendSignLink(ctx context.Context, orgID, contractID int64) error {
	// 生成 PDF
	pdfBytes, err := s.GeneratePDF(ctx, orgID, contractID)
	if err != nil {
		return fmt.Errorf("生成PDF失败: %w", err)
	}

	// 上传到 OSS
	pdfUrl, err := s.uploadPdfToOss(ctx, orgID, contractID, pdfBytes)
	if err != nil {
		return fmt.Errorf("上传PDF失败: %w", err)
	}

	// 更新合同 PDF URL
	contract, _ := s.contractRepo.FindByID(orgID, contractID)
	s.contractRepo.Update(orgID, contractID, map[string]interface{}{
		"pdf_url": pdfUrl,
	})

	// 获取员工手机号
	emp, _ := s.empRepo.FindByID(orgID, contract.EmployeeID)
	if emp == nil {
		return fmt.Errorf("员工不存在")
	}
	empPhone, _ := crypto.Decrypt(emp.PhoneEncrypted, s.aesKey())
	if empPhone == "" {
		return fmt.Errorf("员工手机号为空")
	}

	// 生成签署链接
	signToken, _ := generateToken()
	signLink := fmt.Sprintf("%s/sign/%d?token=%s",
		os.Getenv("APP_BASE_URL"), contractID, signToken)

	// 发送短信
	if s.smsClient != nil {
		templateCode := os.Getenv("ALIYUN_SMS_CONTRACT_TEMPLATE_CODE")
		if templateCode != "" {
			templateParam := fmt.Sprintf(`{"name":"%s","link":"%s","days":"7"}`, emp.Name, signLink)
			_ = s.smsClient.SendTemplateMessage(ctx, empPhone, templateCode, templateParam)
		}
	}

	return nil
}

// uploadPdfToOss 上传合同 PDF 到 OSS
func (s *ContractService) uploadPdfToOss(ctx context.Context, orgID, contractID int64, pdfBytes []byte) (string, error) {
	if s.ossClient == nil {
		objectKey := fmt.Sprintf("contracts/org_%d/contract_%d_%d.pdf",
			orgID, contractID, time.Now().Unix())
		return objectKey, nil
	}
	objectKey := fmt.Sprintf("contracts/org_%d/contract_%d_%d.pdf",
		orgID, contractID, time.Now().Unix())
	putURL, err := s.ossClient.GeneratePutURL(ctx, "contract", orgID, objectKey, int64(len(pdfBytes)), "application/pdf", 30*time.Minute)
	if err != nil {
		return "", fmt.Errorf("生成签署链接失败: %w", err)
	}
	_ = putURL // 前端直传，或后端上传
	return objectKey, nil
}
