package email_template

import (
	"errors"
	"fmt"
)

// Service 邮箱模板业务逻辑层
type Service struct {
	repo *Repository
}

// NewService 创建邮箱模板 Service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateTemplate 创建模板
func (s *Service) CreateTemplate(orgID, userID int64, req *CreateTemplateRequest) (*TemplateResponse, error) {
	// 检查名称唯一性
	exists, err := s.repo.FindByName(orgID, req.Name)
	if err == nil && exists != nil {
		return nil, errors.New("该模板名称已存在")
	}

	// 如果设为默认，先清除其他默认
	if req.IsDefault {
		if err := s.repo.ClearDefault(orgID); err != nil {
			return nil, fmt.Errorf("清除默认模板失败: %w", err)
		}
	}

	tpl := &EmailTemplate{
		Name:      req.Name,
		Subject:   req.Subject,
		Content:   req.Content,
		IsDefault: req.IsDefault,
	}
	tpl.OrgID = orgID
	tpl.CreatedBy = userID
	tpl.UpdatedBy = userID

	if err := s.repo.Create(tpl); err != nil {
		return nil, fmt.Errorf("创建模板失败: %w", err)
	}

	return toResponse(tpl), nil
}

// UpdateTemplate 更新模板
func (s *Service) UpdateTemplate(orgID, userID, id int64, req *UpdateTemplateRequest) (*TemplateResponse, error) {
	// 检查名称唯一性
	if req.Name != nil {
		existing, err := s.repo.FindByName(orgID, *req.Name)
		if err == nil && existing != nil && existing.ID != id {
			return nil, errors.New("该模板名称已存在")
		}
	}

	// 如果设为默认，先清除其他默认
	if req.IsDefault != nil && *req.IsDefault {
		if err := s.repo.ClearDefault(orgID); err != nil {
			return nil, fmt.Errorf("清除默认模板失败: %w", err)
		}
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Subject != nil {
		updates["subject"] = *req.Subject
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.IsDefault != nil {
		updates["is_default"] = *req.IsDefault
	}

	if len(updates) == 0 {
		return nil, errors.New("没有需要更新的字段")
	}
	updates["updated_by"] = userID

	if err := s.repo.Update(orgID, id, updates); err != nil {
		return nil, errors.New("模板不存在")
	}

	tpl, err := s.repo.FindByID(orgID, id)
	if err != nil {
		return nil, errors.New("模板不存在")
	}

	return toResponse(tpl), nil
}

// ListTemplates 查询模板列表
func (s *Service) ListTemplates(orgID int64, query ListQuery) ([]TemplateResponse, int64, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}

	templates, total, err := s.repo.List(orgID, query.Page, query.PageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("查询模板列表失败: %w", err)
	}

	resp := make([]TemplateResponse, 0, len(templates))
	for i := range templates {
		resp = append(resp, *toResponse(&templates[i]))
	}

	return resp, total, nil
}

// DeleteTemplate 删除模板
func (s *Service) DeleteTemplate(orgID, id int64) error {
	if err := s.repo.Delete(orgID, id); err != nil {
		return errors.New("模板不存在")
	}
	return nil
}

// toResponse 将 EmailTemplate 转为 TemplateResponse
func toResponse(tpl *EmailTemplate) *TemplateResponse {
	return &TemplateResponse{
		ID:        tpl.ID,
		OrgID:     tpl.OrgID,
		Name:      tpl.Name,
		Subject:   tpl.Subject,
		Content:   tpl.Content,
		IsDefault: tpl.IsDefault,
		CreatedBy: tpl.CreatedBy,
		CreatedAt: tpl.CreatedAt,
		UpdatedBy: tpl.UpdatedBy,
		UpdatedAt: tpl.UpdatedAt,
	}
}
