package employee

import (
	"context"
	"fmt"
	"time"

	"github.com/wencai/easyhr/internal/common/config"
	"github.com/wencai/easyhr/internal/common/crypto"
	"github.com/wencai/easyhr/internal/common/model"
	"gorm.io/gorm"
)

// ContractService 合同业务逻辑层
type ContractService struct {
	contractRepo *ContractRepository
	empRepo      *Repository
	db           *gorm.DB // 用于查询 Organization 信息
	cryptoCfg    config.CryptoConfig
	todoSvc      TodoCreator // interface to avoid circular import
}

// NewContractService 创建合同 Service
func NewContractService(
	contractRepo *ContractRepository,
	empRepo *Repository,
	db *gorm.DB,
	cryptoCfg config.CryptoConfig,
	todoSvc TodoCreator,
) *ContractService {
	return &ContractService{
		contractRepo: contractRepo,
		empRepo:      empRepo,
		db:           db,
		cryptoCfg:    cryptoCfg,
		todoSvc:      todoSvc,
	}
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
	contract.Salary = req.Salary
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

	// 构建 PDF 数据
	data := &ContractPDFData{
		OrgName:      org.Name,
		CreditCode:   org.CreditCode,
		EmployeeName: emp.Name,
		IDCard:       idCard,
		Position:     emp.Position,
		City:         org.City,
		Salary:       contract.Salary,
		StartDate:    contract.StartDate,
		EndDate:      contract.EndDate,
		ContractType: contract.ContractType,
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

	if contract.Status != ContractStatusDraft {
		return nil, fmt.Errorf("仅草稿状态合同可编辑")
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

	if err := s.contractRepo.Update(orgID, contractID, updates); err != nil {
		return nil, fmt.Errorf("更新合同失败: %w", err)
	}

	return s.GetContract(ctx, orgID, contractID)
}

// toContractResponse 将 Contract 转换为 ContractResponse
func (s *ContractService) toContractResponse(c *Contract, empName string) *ContractResponse {
	resp := &ContractResponse{
		ID:              c.ID,
		EmployeeID:      c.EmployeeID,
		EmployeeName:    empName,
		ContractType:    c.ContractType,
		StartDate:       c.StartDate,
		EndDate:         c.EndDate,
		Salary:          c.Salary,
		Status:          c.Status,
		PDFURL:          c.PDFURL,
		SignedPDFURL:    c.SignedPDFURL,
		SignDate:        c.SignDate,
		TerminateDate:   c.TerminateDate,
		TerminateReason: c.TerminateReason,
		CreatedAt:       c.CreatedAt,
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
