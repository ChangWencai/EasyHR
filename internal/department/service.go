package department

import (
	"errors"
	"fmt"
	"strings"

	"github.com/wencai/easyhr/internal/employee"
	"github.com/wencai/easyhr/internal/position"
)

var (
	ErrDepartmentNotFound  = errors.New("部门不存在")
	ErrHasChildren         = errors.New("该部门下存在子部门，无法删除")
	ErrHasEmployees        = errors.New("该部门下存在员工，无法删除")
	ErrEmployeeNotFound    = errors.New("员工不存在")
	ErrEmployeeNotInDept   = errors.New("员工不在该部门")
	ErrCircularReference   = errors.New("不能将部门移动到自身或下级部门")
)

// Service 部门业务逻辑层
type Service struct {
	repo         *Repository
	empRepo      *employee.Repository
	positionRepo *position.Repository
}

// NewService 创建部门 Service
func NewService(repo *Repository, empRepo *employee.Repository, positionRepo *position.Repository) *Service {
	return &Service{
		repo:         repo,
		empRepo:      empRepo,
		positionRepo: positionRepo,
	}
}

// CreateDepartment 创建部门
func (s *Service) CreateDepartment(orgID, userID int64, req *CreateDepartmentRequest) (*DepartmentResponse, error) {
	dept := &Department{
		Name:      req.Name,
		ParentID:  req.ParentID,
		SortOrder: req.SortOrder,
	}
	dept.OrgID = orgID
	dept.CreatedBy = userID
	dept.UpdatedBy = userID

	if err := s.repo.Create(dept); err != nil {
		return nil, fmt.Errorf("创建部门失败: %w", err)
	}

	return toResponse(dept), nil
}

// UpdateDepartment 更新部门
func (s *Service) UpdateDepartment(orgID, userID int64, id int64, req *UpdateDepartmentRequest) (*DepartmentResponse, error) {
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.ParentID != nil {
		// 循环引用检测
		if *req.ParentID != 0 {
			hasCycle, err := s.hasCycle(orgID, id, *req.ParentID)
			if err != nil {
				return nil, fmt.Errorf("检查部门层级失败: %w", err)
			}
			if hasCycle {
				return nil, ErrCircularReference
			}
		}
		updates["parent_id"] = *req.ParentID
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}

	if len(updates) == 0 {
		return nil, errors.New("没有需要更新的字段")
	}
	updates["updated_by"] = userID

	if err := s.repo.Update(orgID, id, updates); err != nil {
		if errors.Is(err, errors.New("record not found")) {
			return nil, ErrDepartmentNotFound
		}
		return nil, fmt.Errorf("更新部门失败: %w", err)
	}

	dept, err := s.repo.FindByID(orgID, id)
	if err != nil {
		return nil, ErrDepartmentNotFound
	}

	return toResponse(dept), nil
}

// DeleteDepartment 删除部门（校验无子部门、无员工）
func (s *Service) DeleteDepartment(orgID, id int64) error {
	// 检查是否有子部门
	childCount, err := s.repo.CountChildren(orgID, id)
	if err != nil {
		return fmt.Errorf("检查子部门失败: %w", err)
	}
	if childCount > 0 {
		return ErrHasChildren
	}

	// 检查是否有员工
	empCount, err := s.empRepo.CountByDepartment(orgID, id)
	if err != nil {
		return fmt.Errorf("检查部门员工失败: %w", err)
	}
	if empCount > 0 {
		return ErrHasEmployees
	}

	if err := s.repo.Delete(orgID, id); err != nil {
		return ErrDepartmentNotFound
	}
	return nil
}

// ListDepartments 部门列表（分页）
func (s *Service) ListDepartments(orgID int64, page, pageSize int) ([]DepartmentResponse, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	departments, total, err := s.repo.List(orgID, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("查询部门列表失败: %w", err)
	}

	var resp []DepartmentResponse
	for i := range departments {
		resp = append(resp, *toResponse(&departments[i]))
	}
	return resp, total, nil
}

// GetDepartment 获取部门详情
func (s *Service) GetDepartment(orgID, id int64) (*DepartmentResponse, error) {
	dept, err := s.repo.FindByID(orgID, id)
	if err != nil {
		return nil, ErrDepartmentNotFound
	}
	return toResponse(dept), nil
}

// GetTree 获取组织架构树（部门->岗位->员工三层）
func (s *Service) GetTree(orgID int64) ([]*TreeNode, error) {
	departments, err := s.repo.ListAll(orgID)
	if err != nil {
		return nil, fmt.Errorf("查询部门失败: %w", err)
	}

	employees, err := s.empRepo.ListAllByOrg(orgID)
	if err != nil {
		return nil, fmt.Errorf("查询员工失败: %w", err)
	}

	// 获取岗位列表（BuildTree v2 使用真实岗位节点）
	positions, err := s.positionRepo.ListByOrg(orgID)
	if err != nil {
		return nil, fmt.Errorf("查询岗位失败: %w", err)
	}

	tree := s.BuildTree(departments, employees, positions)
	return tree, nil
}

// SearchTree 搜索组织架构树并高亮匹配节点
func (s *Service) SearchTree(orgID int64, keyword string) ([]*TreeNode, error) {
	tree, err := s.GetTree(orgID)
	if err != nil {
		return nil, err
	}

	kw := strings.ToLower(keyword)
	s.markMatches(tree, kw)
	return tree, nil
}

// BuildTree v2 从扁平记录构建3层树（部门->岗位->员工）
// 使用 Position 表构建真实岗位节点，替代旧版按 Employee.Position 文本分组
func (s *Service) BuildTree(departments []Department, employees []employee.Employee, positions []position.Position) []*TreeNode {
	// 按 parent_id 分组部门
	childrenMap := make(map[int64][]*Department)
	var roots []*Department

	for i := range departments {
		dept := &departments[i]
		if dept.ParentID == nil {
			roots = append(roots, dept)
		} else {
			childrenMap[*dept.ParentID] = append(childrenMap[*dept.ParentID], dept)
		}
	}

	// 按 department_id 分组岗位
	posByDept := make(map[int64][]position.Position)
	var commonPositions []position.Position
	for _, p := range positions {
		if p.DepartmentID == nil {
			commonPositions = append(commonPositions, p)
		} else {
			posByDept[*p.DepartmentID] = append(posByDept[*p.DepartmentID], p)
		}
	}

	// 按 position_id 分组员工
	empByPosID := make(map[int64][]employee.Employee)
	var unassignedEmpsByDept map[int64][]employee.Employee = make(map[int64][]employee.Employee)
	for _, emp := range employees {
		if emp.DepartmentID == nil {
			continue
		}
		if emp.PositionID != nil {
			empByPosID[*emp.PositionID] = append(empByPosID[*emp.PositionID], emp)
		} else {
			unassignedEmpsByDept[*emp.DepartmentID] = append(unassignedEmpsByDept[*emp.DepartmentID], emp)
		}
	}

	// 递归构建树
	var buildNodes func(depts []*Department) []*TreeNode
	buildNodes = func(depts []*Department) []*TreeNode {
		var nodes []*TreeNode
		for _, dept := range depts {
			deptNode := &TreeNode{
				ID:   dept.ID,
				Name: dept.Name,
				Type: "department",
			}

			// 递归添加子部门
			if children, ok := childrenMap[dept.ID]; ok {
				deptNode.Children = append(deptNode.Children, buildNodes(children)...)
			}

			// 添加部门专属岗位节点
			if deptPositions, ok := posByDept[dept.ID]; ok {
				for _, pos := range deptPositions {
					posNode := &TreeNode{
						ID:   pos.ID,
						Name: pos.Name,
						Type: "position",
					}
					// 添加该岗位下的员工
					if posEmps, ok := empByPosID[pos.ID]; ok {
						for _, emp := range posEmps {
							empNode := &TreeNode{
								ID:   emp.ID,
								Name: emp.Name,
								Type: "employee",
							}
							posNode.Children = append(posNode.Children, empNode)
						}
					}
					deptNode.Children = append(deptNode.Children, posNode)
				}
			}

			// 添加通用岗位节点（分配到该部门的员工中引用通用岗位的）
			for _, pos := range commonPositions {
				// 检查该通用岗位下是否有属于此部门的员工
				if posEmps, ok := empByPosID[pos.ID]; ok {
					var deptPosEmps []employee.Employee
					for _, emp := range posEmps {
						if emp.DepartmentID != nil && *emp.DepartmentID == dept.ID {
							deptPosEmps = append(deptPosEmps, emp)
						}
					}
					if len(deptPosEmps) > 0 {
						posNode := &TreeNode{
							ID:   pos.ID,
							Name: pos.Name,
							Type: "position",
						}
						for _, emp := range deptPosEmps {
							empNode := &TreeNode{
								ID:   emp.ID,
								Name: emp.Name,
								Type: "employee",
							}
							posNode.Children = append(posNode.Children, empNode)
						}
						deptNode.Children = append(deptNode.Children, posNode)
					}
				}
			}

			// 添加未分配岗位的员工节点
			if unassigned, ok := unassignedEmpsByDept[dept.ID]; ok && len(unassigned) > 0 {
				posNode := &TreeNode{
					ID:   0,
					Name: "未分配岗位",
					Type: "position",
				}
				for _, emp := range unassigned {
					empNode := &TreeNode{
						ID:   emp.ID,
						Name: emp.Name,
						Type: "employee",
					}
					posNode.Children = append(posNode.Children, empNode)
				}
				deptNode.Children = append(deptNode.Children, posNode)
			}

			nodes = append(nodes, deptNode)
		}
		return nodes
	}

	return buildNodes(roots)
}

// markMatches 递归标记搜索匹配的节点（D-14-08 更新配色）
func (s *Service) markMatches(nodes []*TreeNode, keyword string) bool {
	if nodes == nil {
		return false
	}

	anyChildMatched := false
	for _, node := range nodes {
		childMatched := s.markMatches(node.Children, keyword)

		// 检查当前节点是否匹配
		matched := strings.Contains(strings.ToLower(node.Name), keyword)

		if matched || childMatched {
			anyChildMatched = true
		}

		if matched {
			// 匹配节点高亮蓝色（D-14-08: 使用主色调 #4F6EF7）
			node.ItemStyle = map[string]interface{}{
				"color": "#4F6EF7",
			}
			node.Label = map[string]interface{}{
				"color":      "#4F6EF7",
				"fontWeight": 700,
			}
		} else if !childMatched {
			// 未匹配节点降低透明度（D-14-08: opacity 0.25）
			node.ItemStyle = map[string]interface{}{
				"opacity": 0.25,
			}
			node.Label = map[string]interface{}{
				"opacity": 0.25,
			}
		}
	}

	return anyChildMatched
}

// hasCycle 检测设置 newParentID 为 deptID 的父部门是否会形成循环引用
// BFS 从 newParentID 向上追溯，如果发现 deptID 则存在循环
func (s *Service) hasCycle(orgID, deptID, newParentID int64) (bool, error) {
	if deptID == newParentID {
		return true, nil
	}

	current := newParentID
	visited := make(map[int64]bool)
	for current != 0 {
		if current == deptID {
			return true, nil
		}
		if visited[current] {
			break
		}
		visited[current] = true

		parent, err := s.repo.FindByID(orgID, current)
		if err != nil || parent.ParentID == nil {
			break
		}
		current = *parent.ParentID
	}
	return false, nil
}

// TransferAndDeleteDepartment 转移员工后删除部门（D-14-09 原子操作）
func (s *Service) TransferAndDeleteDepartment(orgID, deptID, targetDeptID int64, employeeIDs []int64) error {
	// 校验目标部门存在
	_, err := s.repo.FindByID(orgID, targetDeptID)
	if err != nil {
		return ErrDepartmentNotFound
	}

	// 校验所有员工属于源部门并执行转移
	for _, empID := range employeeIDs {
		emp, err := s.empRepo.FindByID(orgID, empID)
		if err != nil {
			return ErrEmployeeNotFound
		}
		if emp.DepartmentID == nil || *emp.DepartmentID != deptID {
			return ErrEmployeeNotInDept
		}
		if err := s.empRepo.UpdateDepartmentID(orgID, empID, targetDeptID); err != nil {
			return fmt.Errorf("转移员工 %d 失败: %w", empID, err)
		}
	}

	// 删除部门
	return s.repo.Delete(orgID, deptID)
}

// toResponse 将 Department 转为 DepartmentResponse
func toResponse(dept *Department) *DepartmentResponse {
	return &DepartmentResponse{
		ID:        dept.ID,
		OrgID:     dept.OrgID,
		Name:      dept.Name,
		ParentID:  dept.ParentID,
		SortOrder: dept.SortOrder,
	}
}
