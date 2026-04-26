package sms_template

import (
	"errors"
	"fmt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateTemplate(orgID, userID int64, req *CreateSmsTemplateRequest) (*SmsTemplateResponse, error) {
	exists, err := s.repo.FindByName(orgID, req.Name)
	if err == nil && exists != nil {
		return nil, errors.New("该模板名称已存在")
	}

	tpls, _, _ := s.repo.List(orgID, 1, 1)
	if len(tpls) == 0 {
		s.repo.SeedPresets(orgID)
	}

	if req.IsDefault {
		if err := s.repo.ClearDefault(orgID); err != nil {
			return nil, fmt.Errorf("清除默认模板失败: %w", err)
		}
	}

	tpl := &SmsTemplate{
		Name:         req.Name,
		Scene:        req.Scene,
		TemplateCode: req.TemplateCode,
		Content:      req.Content,
		IsDefault:    req.IsDefault,
	}
	tpl.OrgID = orgID
	tpl.CreatedBy = userID
	tpl.UpdatedBy = userID

	if err := s.repo.Create(tpl); err != nil {
		return nil, fmt.Errorf("创建模板失败: %w", err)
	}

	return toResponse(tpl), nil
}

func (s *Service) UpdateTemplate(orgID, userID, id int64, req *UpdateSmsTemplateRequest) (*SmsTemplateResponse, error) {
	if req.Name != nil {
		existing, err := s.repo.FindByName(orgID, *req.Name)
		if err == nil && existing != nil && existing.ID != id {
			return nil, errors.New("该模板名称已存在")
		}
	}

	if req.IsDefault != nil && *req.IsDefault {
		if err := s.repo.ClearDefault(orgID); err != nil {
			return nil, fmt.Errorf("清除默认模板失败: %w", err)
		}
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Scene != nil {
		updates["scene"] = *req.Scene
	}
	if req.TemplateCode != nil {
		updates["template_code"] = *req.TemplateCode
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

func (s *Service) ListTemplates(orgID int64, query ListQuery) ([]SmsTemplateResponse, int64, error) {
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

	resp := make([]SmsTemplateResponse, 0, len(templates))
	for i := range templates {
		resp = append(resp, *toResponse(&templates[i]))
	}

	return resp, total, nil
}

func (s *Service) DeleteTemplate(orgID, id int64) error {
	if err := s.repo.Delete(orgID, id); err != nil {
		return errors.New("模板不存在")
	}
	return nil
}

func toResponse(tpl *SmsTemplate) *SmsTemplateResponse {
	return &SmsTemplateResponse{
		ID:           tpl.ID,
		OrgID:        tpl.OrgID,
		Name:         tpl.Name,
		Scene:        tpl.Scene,
		TemplateCode: tpl.TemplateCode,
		Content:      tpl.Content,
		IsDefault:    tpl.IsDefault,
		CreatedBy:    tpl.CreatedBy,
		CreatedAt:    tpl.CreatedAt,
		UpdatedBy:    tpl.UpdatedBy,
		UpdatedAt:    tpl.UpdatedAt,
	}
}
