package position

import (
	"errors"
	"fmt"
)

var (
	// ErrPositionNotFound 岗位不存在
	ErrPositionNotFound = errors.New("岗位不存在")
	// ErrPositionDuplicate 同一部门内岗位名称重复
	ErrPositionDuplicate = errors.New("同一部门内该岗位名称已存在")
	// ErrPositionInUse 岗位下有员工，无法删除
	ErrPositionInUse = errors.New("该岗位下有员工，无法删除")
)

// Service 岗位业务逻辑层
type Service struct {
	repo *Repository
}

// NewService 创建岗位 Service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetPositionByID 根据 ID 获取岗位（不校验租户）
func (s *Service) GetPositionByID(id int64) (*Position, error) {
	return s.repo.FindByIDWithoutTenant(id)
}

// CreatePosition 创建岗位（含去重校验）
func (s *Service) CreatePosition(orgID, userID int64, req *CreatePositionRequest) (*PositionResponse, error) {
	// 检查同部门同名岗位
	exists, err := s.repo.ExistsByNameAndDept(orgID, req.DepartmentID, req.Name)
	if err != nil {
		return nil, fmt.Errorf("检查岗位失败: %w", err)
	}
	if exists {
		return nil, ErrPositionDuplicate
	}

	pos := &Position{
		Name:         req.Name,
		DepartmentID: req.DepartmentID,
		SortOrder:    req.SortOrder,
	}
	pos.OrgID = orgID
	pos.CreatedBy = userID
	pos.UpdatedBy = userID

	if err := s.repo.Create(pos); err != nil {
		return nil, fmt.Errorf("创建岗位失败: %w", err)
	}

	return toPositionResponse(pos), nil
}

// UpdatePosition 更新岗位
func (s *Service) UpdatePosition(orgID, userID, id int64, req *UpdatePositionRequest) (*PositionResponse, error) {
	// 如果修改了名称或部门，检查唯一性
	if req.Name != nil {
		// 查找当前岗位以获取当前 department_id
		current, err := s.repo.FindByID(orgID, id)
		if err != nil {
			return nil, ErrPositionNotFound
		}
		deptID := current.DepartmentID
		if req.DepartmentID != nil {
			deptID = req.DepartmentID
		}
		exists, err := s.repo.ExistsByNameAndDept(orgID, deptID, *req.Name)
		if err != nil {
			return nil, fmt.Errorf("检查岗位失败: %w", err)
		}
		if exists {
			return nil, ErrPositionDuplicate
		}
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.DepartmentID != nil {
		updates["department_id"] = req.DepartmentID
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}

	if len(updates) == 0 {
		return nil, errors.New("没有需要更新的字段")
	}
	updates["updated_by"] = userID

	if err := s.repo.Update(orgID, id, updates); err != nil {
		return nil, ErrPositionNotFound
	}

	pos, err := s.repo.FindByID(orgID, id)
	if err != nil {
		return nil, ErrPositionNotFound
	}

	return toPositionResponse(pos), nil
}

// DeletePosition 删除岗位（检查是否有员工引用）
func (s *Service) DeletePosition(orgID, id int64) error {
	// 检查是否有员工使用该岗位
	count, err := s.repo.CountByPositionID(orgID, id)
	if err != nil {
		return fmt.Errorf("检查岗位员工失败: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("%w（%d名员工）", ErrPositionInUse, count)
	}

	if err := s.repo.Delete(orgID, id); err != nil {
		return ErrPositionNotFound
	}
	return nil
}

// ListPositions 获取岗位列表
func (s *Service) ListPositions(orgID int64) ([]PositionResponse, error) {
	positions, err := s.repo.ListByOrg(orgID)
	if err != nil {
		return nil, fmt.Errorf("查询岗位列表失败: %w", err)
	}

	var resp []PositionResponse
	for i := range positions {
		resp = append(resp, *toPositionResponse(&positions[i]))
	}
	return resp, nil
}

// GetSelectOptions 获取岗位下拉选项（按部门分组）
// 部门专属岗位 + 通用岗位 + 未分配选项
func (s *Service) GetSelectOptions(orgID int64, deptID *int64) (*PositionSelectOptions, error) {
	positions, err := s.repo.ListByOrg(orgID)
	if err != nil {
		return nil, err
	}

	var deptPositions, commonPositions []PositionOption
	for _, p := range positions {
		opt := PositionOption{ID: &p.ID, Name: p.Name}
		if p.DepartmentID == nil {
			commonPositions = append(commonPositions, opt)
		} else if deptID != nil && *p.DepartmentID == *deptID {
			deptPositions = append(deptPositions, opt)
		} else if deptID == nil {
			deptPositions = append(deptPositions, opt)
		}
	}

	return &PositionSelectOptions{
		DeptPositions:   deptPositions,
		CommonPositions: commonPositions,
		Unassigned:      PositionOption{ID: nil, Name: "未分配岗位"},
	}, nil
}

// MigrateFromEmployeePositions 迁移现有员工 position 文本到 Position 表
// 仅在 Position 表为空时执行，幂等操作。employees 由调用方传入，避免循环依赖
func (s *Service) MigrateFromEmployeePositions(orgID int64, employees []struct {
	Name     string
	Position string
}) error {
	// 检查是否已迁移
	existing, err := s.repo.ListByOrg(orgID)
	if err != nil {
		return fmt.Errorf("检查已有岗位失败: %w", err)
	}
	if len(existing) > 0 {
		return nil // 已迁移，跳过
	}

	// 收集唯一岗位名称
	posNames := make(map[string]bool)
	for _, emp := range employees {
		if emp.Position != "" {
			posNames[emp.Position] = true
		}
	}

	// 为每个唯一名称创建通用岗位记录
	for name := range posNames {
		pos := &Position{
			Name:         name,
			DepartmentID: nil, // 迁移的岗位默认为通用岗位
			SortOrder:    0,
		}
		pos.OrgID = orgID
		if err := s.repo.Create(pos); err != nil {
			return fmt.Errorf("创建岗位 %q 失败: %w", name, err)
		}
	}

	return nil
}

// FindOrCreateByName 按岗位名称查找或创建岗位，并返回其 ID
// 如果同名岗位已存在（无论部门归属），直接返回；否则创建新的通用岗位
func (s *Service) FindOrCreateByName(orgID, userID int64, name string, deptID *int64) (int64, error) {
	if name == "" {
		return 0, errors.New("岗位名称不能为空")
	}

	// 查找同名岗位（跨部门唯一性）
	positions, err := s.repo.ListByOrg(orgID)
	if err != nil {
		return 0, fmt.Errorf("查询岗位列表失败: %w", err)
	}
	for _, p := range positions {
		if p.Name == name {
			return p.ID, nil
		}
	}

	// 不存在：创建通用岗位（department_id=nil）
	pos := &Position{
		Name:         name,
		DepartmentID: nil, // 始终创建为通用岗位
		SortOrder:    0,
	}
	pos.OrgID = orgID
	pos.CreatedBy = userID
	pos.UpdatedBy = userID

	if err := s.repo.Create(pos); err != nil {
		return 0, fmt.Errorf("创建岗位失败: %w", err)
	}
	return pos.ID, nil
}

// toPositionResponse 将 Position 转为 PositionResponse
func toPositionResponse(pos *Position) *PositionResponse {
	return &PositionResponse{
		ID:           pos.ID,
		OrgID:        pos.OrgID,
		Name:         pos.Name,
		DepartmentID: pos.DepartmentID,
		SortOrder:    pos.SortOrder,
	}
}
