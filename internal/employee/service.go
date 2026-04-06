package employee

import (
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/wencai/easyhr/internal/common/config"
	"github.com/wencai/easyhr/internal/common/crypto"
)

// Service 员工业务逻辑层
type Service struct {
	repo     *Repository
	cryptoCfg config.CryptoConfig
}

// NewService 创建员工 Service
func NewService(repo *Repository, cryptoCfg config.CryptoConfig) *Service {
	return &Service{
		repo:     repo,
		cryptoCfg: cryptoCfg,
	}
}

// aesKey 获取 AES 密钥字节
func (s *Service) aesKey() []byte {
	return []byte(s.cryptoCfg.AESKey)
}

// CreateEmployee 创建员工
func (s *Service) CreateEmployee(orgID, userID int64, req *CreateEmployeeRequest) (*EmployeeResponse, error) {
	aesKey := s.aesKey()

	// 手机号加密+哈希
	phoneHash := crypto.HashSHA256(req.Phone)
	phoneEncrypted, err := crypto.Encrypt(req.Phone, aesKey)
	if err != nil {
		return nil, fmt.Errorf("加密手机号失败: %w", err)
	}

	// 身份证号加密+哈希
	idCardHash := crypto.HashSHA256(req.IDCard)
	idCardEncrypted, err := crypto.Encrypt(req.IDCard, aesKey)
	if err != nil {
		return nil, fmt.Errorf("加密身份证号失败: %w", err)
	}

	// 从身份证号提取性别和出生日期
	gender, birthDate, err := extractFromIDCard(req.IDCard)
	if err != nil {
		return nil, fmt.Errorf("身份证号解析失败: %w", err)
	}

	// 解析入职日期
	hireDate, err := time.Parse("2006-01-02", req.HireDate)
	if err != nil {
		return nil, fmt.Errorf("入职日期格式错误: %w", err)
	}

	emp := &Employee{}
	emp.OrgID = orgID
	emp.CreatedBy = userID
	emp.UpdatedBy = userID
	emp.Name = req.Name
	emp.PhoneEncrypted = phoneEncrypted
	emp.PhoneHash = phoneHash
	emp.IDCardEncrypted = idCardEncrypted
	emp.IDCardHash = idCardHash
	emp.Gender = gender
	emp.BirthDate = &birthDate
	emp.Position = req.Position
	emp.HireDate = hireDate
	emp.Status = StatusPending

	// 可选字段：银行卡
	if req.BankName != "" || req.BankAccount != "" {
		emp.BankName = req.BankName
		if req.BankAccount != "" {
			bankEncrypted, err := crypto.Encrypt(req.BankAccount, aesKey)
			if err != nil {
				return nil, fmt.Errorf("加密银行账号失败: %w", err)
			}
			emp.BankAccountEncrypted = bankEncrypted
			emp.BankAccountHash = crypto.HashSHA256(req.BankAccount)
		}
	}

	// 可选字段：紧急联系人
	emp.EmergencyContact = req.EmergencyContact
	if req.EmergencyPhone != "" {
		emergEncrypted, err := crypto.Encrypt(req.EmergencyPhone, aesKey)
		if err != nil {
			return nil, fmt.Errorf("加密紧急联系电话失败: %w", err)
		}
		emp.EmergencyPhoneEncrypted = emergEncrypted
		emp.EmergencyPhoneHash = crypto.HashSHA256(req.EmergencyPhone)
	}

	emp.Address = req.Address
	emp.Remark = req.Remark

	if err := s.repo.Create(emp); err != nil {
		if err == ErrPhoneDuplicate {
			return nil, fmt.Errorf("该手机号已存在")
		}
		if err == ErrIDCardDuplicate {
			return nil, fmt.Errorf("该身份证号已存在")
		}
		return nil, fmt.Errorf("创建员工失败: %w", err)
	}

	return s.toResponse(emp)
}

// ListEmployees 搜索+分页查询员工列表（脱敏）
func (s *Service) ListEmployees(orgID int64, query ListQueryParams) ([]EmployeeResponse, int64, error) {
	params := SearchParams{
		Name:     query.Name,
		Position: query.Position,
		Phone:    query.Phone,
		Status:   query.Status,
	}

	page := query.Page
	pageSize := query.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	employees, total, err := s.repo.List(orgID, params, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("查询员工列表失败: %w", err)
	}

	var responses []EmployeeResponse
	for _, emp := range employees {
		resp, err := s.toResponse(&emp)
		if err != nil {
			return nil, 0, err
		}
		responses = append(responses, *resp)
	}

	return responses, total, nil
}

// GetEmployee 获取员工详情（脱敏）
func (s *Service) GetEmployee(orgID, id int64) (*EmployeeResponse, error) {
	emp, err := s.repo.FindByID(orgID, id)
	if err != nil {
		return nil, fmt.Errorf("员工不存在")
	}
	return s.toResponse(emp)
}

// UpdateEmployee 更新员工信息（部分更新）
func (s *Service) UpdateEmployee(orgID, userID, id int64, req *UpdateEmployeeRequest) (*EmployeeResponse, error) {
	aesKey := s.aesKey()
	updates := make(map[string]interface{})
	updates["updated_by"] = userID

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Position != nil {
		updates["position"] = *req.Position
	}
	if req.HireDate != nil {
		hireDate, err := time.Parse("2006-01-02", *req.HireDate)
		if err != nil {
			return nil, fmt.Errorf("入职日期格式错误: %w", err)
		}
		updates["hire_date"] = hireDate
	}
	if req.BankName != nil {
		updates["bank_name"] = *req.BankName
	}
	if req.EmergencyContact != nil {
		updates["emergency_contact"] = *req.EmergencyContact
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.Remark != nil {
		updates["remark"] = *req.Remark
	}

	// 敏感字段更新：需要重新加密+哈希
	if req.Phone != nil {
		phone := *req.Phone
		phoneHash := crypto.HashSHA256(phone)
		phoneEncrypted, err := crypto.Encrypt(phone, aesKey)
		if err != nil {
			return nil, fmt.Errorf("加密手机号失败: %w", err)
		}
		updates["phone_hash"] = phoneHash
		updates["phone_encrypted"] = phoneEncrypted
	}
	if req.IDCard != nil {
		idCard := *req.IDCard
		idCardHash := crypto.HashSHA256(idCard)
		idCardEncrypted, err := crypto.Encrypt(idCard, aesKey)
		if err != nil {
			return nil, fmt.Errorf("加密身份证号失败: %w", err)
		}
		updates["id_card_hash"] = idCardHash
		updates["id_card_encrypted"] = idCardEncrypted

		// 重新提取性别和出生日期
		gender, birthDate, err := extractFromIDCard(idCard)
		if err == nil {
			updates["gender"] = gender
			updates["birth_date"] = birthDate
		}
	}
	if req.BankAccount != nil {
		account := *req.BankAccount
		if account != "" {
			bankEncrypted, err := crypto.Encrypt(account, aesKey)
			if err != nil {
				return nil, fmt.Errorf("加密银行账号失败: %w", err)
			}
			updates["bank_account_encrypted"] = bankEncrypted
			updates["bank_account_hash"] = crypto.HashSHA256(account)
		} else {
			updates["bank_account_encrypted"] = ""
			updates["bank_account_hash"] = ""
		}
	}
	if req.EmergencyPhone != nil {
		phone := *req.EmergencyPhone
		if phone != "" {
			emergEncrypted, err := crypto.Encrypt(phone, aesKey)
			if err != nil {
				return nil, fmt.Errorf("加密紧急联系电话失败: %w", err)
			}
			updates["emergency_phone_encrypted"] = emergEncrypted
			updates["emergency_phone_hash"] = crypto.HashSHA256(phone)
		} else {
			updates["emergency_phone_encrypted"] = ""
			updates["emergency_phone_hash"] = ""
		}
	}

	if err := s.repo.Update(orgID, id, updates); err != nil {
		return nil, fmt.Errorf("更新员工失败: %w", err)
	}

	return s.GetEmployee(orgID, id)
}

// DeleteEmployee 软删除员工
func (s *Service) DeleteEmployee(orgID, id int64) error {
	if err := s.repo.Delete(orgID, id); err != nil {
		return fmt.Errorf("删除员工失败: %w", err)
	}
	return nil
}

// GetSensitiveInfo 获取员工完整敏感信息（仅 OWNER/ADMIN 可调用）
func (s *Service) GetSensitiveInfo(orgID, id int64) (*SensitiveInfoResponse, error) {
	emp, err := s.repo.FindByID(orgID, id)
	if err != nil {
		return nil, fmt.Errorf("员工不存在")
	}

	aesKey := s.aesKey()
	resp := &SensitiveInfoResponse{}

	resp.Phone, _ = crypto.Decrypt(emp.PhoneEncrypted, aesKey)
	resp.IDCard, _ = crypto.Decrypt(emp.IDCardEncrypted, aesKey)

	if emp.BankAccountEncrypted != "" {
		resp.BankAccount, _ = crypto.Decrypt(emp.BankAccountEncrypted, aesKey)
	}
	if emp.EmergencyPhoneEncrypted != "" {
		resp.EmergencyPhone, _ = crypto.Decrypt(emp.EmergencyPhoneEncrypted, aesKey)
	}

	return resp, nil
}

// ExportExcel 导出员工列表为 Excel（敏感字段脱敏）
func (s *Service) ExportExcel(orgID int64, query ListQueryParams) ([]byte, error) {
	params := SearchParams{
		Name:     query.Name,
		Position: query.Position,
		Phone:    query.Phone,
		Status:   query.Status,
	}

	employees, err := s.repo.FindAllForExport(orgID, params)
	if err != nil {
		return nil, fmt.Errorf("查询导出数据失败: %w", err)
	}

	f := excelize.NewFile()
	sheet := "员工列表"
	f.SetSheetName("Sheet1", sheet)

	// 表头
	headers := []string{"姓名", "手机号", "身份证号", "性别", "岗位", "入职日期", "状态"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// 数据行
	aesKey := s.aesKey()
	for i, emp := range employees {
		row := i + 2

		phone, _ := crypto.Decrypt(emp.PhoneEncrypted, aesKey)
		idCard, _ := crypto.Decrypt(emp.IDCardEncrypted, aesKey)

		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), emp.Name)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), crypto.MaskPhone(phone))
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), crypto.MaskIDCard(idCard))
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), emp.Gender)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), emp.Position)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), emp.HireDate.Format("2006-01-02"))
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), emp.Status)
	}

	// 冻结首行
	f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("生成 Excel 失败: %w", err)
	}

	return buf.Bytes(), nil
}

// toResponse 将 Employee 转换为脱敏的 EmployeeResponse
func (s *Service) toResponse(emp *Employee) (*EmployeeResponse, error) {
	aesKey := s.aesKey()

	phone, _ := crypto.Decrypt(emp.PhoneEncrypted, aesKey)
	idCard, _ := crypto.Decrypt(emp.IDCardEncrypted, aesKey)

	resp := &EmployeeResponse{
		ID:        emp.ID,
		Name:      emp.Name,
		Phone:     crypto.MaskPhone(phone),
		IDCard:    crypto.MaskIDCard(idCard),
		Gender:    emp.Gender,
		BirthDate: emp.BirthDate,
		Position:  emp.Position,
		HireDate:  emp.HireDate,
		Status:    emp.Status,
		Address:   emp.Address,
		Remark:    emp.Remark,
		CreatedAt: emp.CreatedAt,
	}

	// 银行卡信息
	if emp.BankName != "" {
		resp.BankName = emp.BankName
	}
	if emp.BankAccountEncrypted != "" {
		account, err := crypto.Decrypt(emp.BankAccountEncrypted, aesKey)
		if err == nil {
			resp.BankAccount = maskBankAccount(account)
		}
	}

	// 紧急联系人
	if emp.EmergencyContact != "" {
		resp.EmergencyContact = emp.EmergencyContact
	}

	return resp, nil
}

// maskBankAccount 银行账号脱敏：保留后4位，前面用 **** 替代
func maskBankAccount(account string) string {
	if len(account) <= 4 {
		return account
	}
	return "****" + account[len(account)-4:]
}
