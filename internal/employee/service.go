package employee

import (
	"context"
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/wencai/easyhr/internal/common/config"
	"github.com/wencai/easyhr/internal/common/crypto"
	"github.com/wencai/easyhr/internal/position"
)

// Service 员工业务逻辑层
type Service struct {
	repo        *Repository
	cryptoCfg   config.CryptoConfig
	todoSvc     TodoCreator          // interface to avoid circular import
	positionSvc *position.Service    // 用于按名称查找/创建岗位并关联 PositionID
}

// TodoCreator interface for creating todo items (avoids circular import from todo package)
type TodoCreator interface {
	CreateTodoFromEmployee(orgID int64, title string, todoType string, employeeID *int64, employeeName string, deadline *time.Time, sourceType string, sourceID *int64) error
	ExistsBySource(ctx context.Context, orgID int64, sourceType string, sourceID *int64) (bool, error)
}

// NewService 创建员工 Service
func NewService(repo *Repository, cryptoCfg config.CryptoConfig, todoSvc TodoCreator, positionSvc *position.Service) *Service {
	return &Service{
		repo:        repo,
		cryptoCfg:   cryptoCfg,
		todoSvc:     todoSvc,
		positionSvc: positionSvc,
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
	emp.DepartmentID = req.DepartmentID
	emp.HireDate = hireDate
	emp.Status = StatusPending

	// PositionID 关联：如果请求中已指定则直接使用；否则按 Position 名称查找或创建岗位
	if req.PositionID != nil {
		emp.PositionID = req.PositionID
	} else if req.Position != "" && s.positionSvc != nil {
		posID, err := s.positionSvc.FindOrCreateByName(orgID, userID, req.Position, nil)
		if err == nil {
			emp.PositionID = &posID
		}
		// 查找失败不影响员工创建，仅记录日志
	}

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

	// 创建合同新签限时待办（入职30日内截止）
	if s.todoSvc != nil {
		hireDate := emp.HireDate
		contractDeadline := hireDate.AddDate(0, 0, 30)
		empID := emp.ID
		_ = s.todoSvc.CreateTodoFromEmployee(
			orgID,
			fmt.Sprintf("员工 %s 入职30日内请完成劳动合同签署", emp.Name),
			"contract_new",
			&empID,
			emp.Name,
			&contractDeadline,
			"employee",
			&empID,
		)
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
		// 同时更新 PositionID：按新 Position 文本查找或创建岗位
		if s.positionSvc != nil {
			posID, err := s.positionSvc.FindOrCreateByName(orgID, userID, *req.Position, nil)
			if err == nil {
				updates["position_id"] = posID
			}
		}
	}
	if req.PositionID != nil {
		updates["position_id"] = *req.PositionID
	}
	if req.DepartmentID != nil {
		updates["department_id"] = *req.DepartmentID
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

// PositionCount 岗位员工数量
type PositionCount struct {
	PositionID int64 `json:"position_id"`
	Count     int64 `json:"count"`
}

// ListPositionCounts 获取每个岗位的员工数量
func (s *Service) ListPositionCounts(orgID int64) ([]PositionCount, error) {
	rows, err := s.repo.CountByPositionIDGrouped(orgID)
	if err != nil {
		return nil, err
	}
	counts := make([]PositionCount, len(rows))
	for i, r := range rows {
		counts[i] = PositionCount{PositionID: r.PositionID, Count: r.Count}
	}
	return counts, nil
}

// DeptCount 部门员工数量
type DeptCount struct {
	DepartmentID int64 `json:"department_id"`
	Count       int64 `json:"count"`
}

// ListDeptCounts 获取每个部门的员工数量
func (s *Service) ListDeptCounts(orgID int64) ([]DeptCount, error) {
	rows, err := s.repo.CountByDepartmentIDGrouped(orgID)
	if err != nil {
		return nil, err
	}
	counts := make([]DeptCount, len(rows))
	for i, r := range rows {
		counts[i] = DeptCount{DepartmentID: r.DepartmentID, Count: r.Count}
	}
	return counts, nil
}

// ListRoster 花名册查询（聚合薪资/年限/合同到期/部门/手机号）
func (s *Service) ListRoster(orgID int64, params ListQueryParams) ([]EmployeeRosterItem, int64, error) {
	page := params.Page
	pageSize := params.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	employees, total, err := s.repo.ListRoster(orgID, params, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("查询花名册失败: %w", err)
	}

	if len(employees) == 0 {
		return []EmployeeRosterItem{}, total, nil
	}

	// 收集所有 employeeIDs
	employeeIDs := make([]int64, 0, len(employees))
	for _, emp := range employees {
		employeeIDs = append(employeeIDs, emp.ID)
	}

	// 批量获取关联数据
	salaryMap, _ := s.repo.GetSalaryAmounts(orgID, employeeIDs)
	contractMap, _ := s.repo.GetContractExpiryDays(orgID, employeeIDs)

	// 收集部门 IDs
	deptIDSet := make(map[int64]bool)
	for _, emp := range employees {
		if emp.DepartmentID != nil {
			deptIDSet[*emp.DepartmentID] = true
		}
	}
	deptIDs := make([]int64, 0, len(deptIDSet))
	for id := range deptIDSet {
		deptIDs = append(deptIDs, id)
	}
	deptMap, _ := s.repo.GetDepartmentNames(orgID, deptIDs)

	// 构建花名册项
	aesKey := s.aesKey()
	now := time.Now()
	items := make([]EmployeeRosterItem, 0, len(employees))

	for _, emp := range employees {
		phone, _ := crypto.Decrypt(emp.PhoneEncrypted, aesKey)

		// 计算在职年限
		yearsOfService := calcYearsOfService(emp.HireDate, now)

		// 部门名称
		var deptName string
		if emp.DepartmentID != nil {
			deptName = deptMap[*emp.DepartmentID]
		}

		// 薪资
		var salaryAmount float64
		if amount, ok := salaryMap[emp.ID]; ok {
			salaryAmount = amount
		}

		item := EmployeeRosterItem{
			ID:                 emp.ID,
			Name:               emp.Name,
			Status:             emp.Status,
			Position:           emp.Position,
			DepartmentID:       emp.DepartmentID,
			DepartmentName:     deptName,
			Phone:              crypto.MaskPhone(phone),
			SalaryAmount:       salaryAmount,
			YearsOfService:     yearsOfService,
			ContractExpiryDays: contractMap[emp.ID],
		}
		items = append(items, item)
	}

	return items, total, nil
}

// calcYearsOfService 计算在职年限，格式 "X年Y月"
func calcYearsOfService(hireDate, now time.Time) string {
	years := now.Year() - hireDate.Year()
	months := int(now.Month()) - int(hireDate.Month())

	if months < 0 {
		years--
		months += 12
	}

	// 如果当前日 < 入职日，月份再减1
	if now.Day() < hireDate.Day() {
		months--
		if months < 0 {
			years--
			months += 12
		}
	}

	if years < 0 {
		return "0年0月"
	}

	return fmt.Sprintf("%d年%d月", years, months)
}

// ExportExcel 导出员工列表为 Excel（敏感字段脱敏，含新增列）
func (s *Service) ExportExcel(orgID int64, query ListQueryParams) ([]byte, error) {
	employees, err := s.repo.FindAllForRosterExport(orgID, query)
	if err != nil {
		return nil, fmt.Errorf("查询导出数据失败: %w", err)
	}

	f := excelize.NewFile()
	sheet := "员工花名册"
	f.SetSheetName("Sheet1", sheet)

	// 表头（含新增列）
	headers := []string{"姓名", "状态", "部门", "岗位薪资", "在职年限", "合同到期", "手机号", "身份证号", "性别", "岗位", "入职日期"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// 批量获取关联数据
	employeeIDs := make([]int64, 0, len(employees))
	for _, emp := range employees {
		employeeIDs = append(employeeIDs, emp.ID)
	}
	salaryMap, _ := s.repo.GetSalaryAmounts(orgID, employeeIDs)
	contractMap, _ := s.repo.GetContractExpiryDays(orgID, employeeIDs)

	deptIDSet := make(map[int64]bool)
	for _, emp := range employees {
		if emp.DepartmentID != nil {
			deptIDSet[*emp.DepartmentID] = true
		}
	}
	deptIDs := make([]int64, 0, len(deptIDSet))
	for id := range deptIDSet {
		deptIDs = append(deptIDs, id)
	}
	deptMap, _ := s.repo.GetDepartmentNames(orgID, deptIDs)

	// 创建红色字体样式（用于合同已过期）
	redStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: "FF0000"},
	})

	now := time.Now()
	aesKey := s.aesKey()

	// 数据行
	for i, emp := range employees {
		row := i + 2

		phone, _ := crypto.Decrypt(emp.PhoneEncrypted, aesKey)
		idCard, _ := crypto.Decrypt(emp.IDCardEncrypted, aesKey)

		// 部门名称
		var deptName string
		if emp.DepartmentID != nil {
			deptName = deptMap[*emp.DepartmentID]
		}

		// 薪资格式
		var salaryStr string
		if amount, ok := salaryMap[emp.ID]; ok && amount > 0 {
			salaryStr = fmt.Sprintf("%.2f", amount)
		}

		// 在职年限
		yearsOfService := calcYearsOfService(emp.HireDate, now)

		// 合同到期天数文本
		var contractStr string
		if days, ok := contractMap[emp.ID]; ok {
			if days == nil {
				contractStr = "无固定期限"
			} else if *days > 0 {
				contractStr = fmt.Sprintf("%d天", *days)
			} else if *days == 0 {
				contractStr = "今天到期"
			} else {
				contractStr = fmt.Sprintf("已过期%d天", -*days)
			}
		} else {
			contractStr = "无合同"
		}

		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), emp.Name)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), emp.Status)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), deptName)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), salaryStr)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), yearsOfService)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), contractStr)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), crypto.MaskPhone(phone))
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), crypto.MaskIDCard(idCard))
		f.SetCellValue(sheet, fmt.Sprintf("I%d", row), emp.Gender)
		f.SetCellValue(sheet, fmt.Sprintf("J%d", row), emp.Position)
		f.SetCellValue(sheet, fmt.Sprintf("K%d", row), emp.HireDate.Format("2006-01-02"))

		// 合同到期天数为负数（已过期）时设置红色字体
		if days, ok := contractMap[emp.ID]; ok && days != nil && *days < 0 {
			cell, _ := excelize.CoordinatesToCellName(6, row)
			f.SetCellStyle(sheet, cell, cell, redStyle)
		}
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
		ID:           emp.ID,
		Name:         emp.Name,
		Phone:        crypto.MaskPhone(phone),
		IDCard:       crypto.MaskIDCard(idCard),
		Gender:       emp.Gender,
		BirthDate:    emp.BirthDate,
		Position:     emp.Position,
		PositionID:   emp.PositionID,
		DepartmentID: emp.DepartmentID,
		HireDate:     emp.HireDate,
		Status:       emp.Status,
		Address:      emp.Address,
		Remark:       emp.Remark,
		CreatedAt:    emp.CreatedAt,
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
