package department

import (
	"errors"
	"fmt"
	"strings"

	"github.com/wencai/easyhr/internal/employee"
)

// Service 部门业务逻辑层
type Service struct {
	repo    *Repository
	empRepo *employee.Repository
}

// NewService 创建部门 Service
func NewService(repo *Repository, empRepo *employee.Repository) *Service {
	return &Service{
		repo:    repo,
		empRepo: empRepo,
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

	tree := s.BuildTree(departments, employees)
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

// BuildTree 从扁平记录构建3层树（部门->岗位->员工）
func (s *Service) BuildTree(departments []Department, employees []employee.Employee) []*TreeNode {
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

	// 按 department_id 分组员工
	empByDept := make(map[int64][]employee.Employee)
	for _, emp := range employees {
		if emp.DepartmentID != nil {
			empByDept[*emp.DepartmentID] = append(empByDept[*emp.DepartmentID], emp)
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

			// 按岗位分组添加员工
			if emps, ok := empByDept[dept.ID]; ok {
				posGroups := make(map[string][]employee.Employee)
				for _, emp := range emps {
					posName := emp.Position
					if posName == "" {
						posName = "未分配岗位"
					}
					posGroups[posName] = append(posGroups[posName], emp)
				}

				for posName, posEmps := range posGroups {
					posNode := &TreeNode{
						ID:   0, // 岗位节点为虚拟节点
						Name: posName,
						Type: "position",
					}
					for _, emp := range posEmps {
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

			nodes = append(nodes, deptNode)
		}
		return nodes
	}

	return buildNodes(roots)
}

// markMatches 递归标记搜索匹配的节点
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
			// 匹配节点高亮蓝色
			node.ItemStyle = map[string]interface{}{
				"color": "#1677FF",
			}
			node.Label = map[string]interface{}{
				"color":      "#1677FF",
				"fontWeight": 700,
			}
		} else if !childMatched {
			// 未匹配节点降低透明度
			node.ItemStyle = map[string]interface{}{
				"color": "#D9D9D9",
			}
			node.Label = map[string]interface{}{
				"color": "#BFBFBF",
			}
		}
	}

	return anyChildMatched
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
