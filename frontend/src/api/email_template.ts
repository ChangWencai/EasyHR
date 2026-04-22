import request from './request'

export interface EmailTemplate {
  id: number
  org_id: number
  name: string
  subject: string
  content: string
  is_default: boolean
  created_by: number
  created_at: string
  updated_by?: number
  updated_at?: string
}

export interface CreateTemplateRequest {
  name: string
  subject: string
  content: string
  is_default?: boolean
}

export interface UpdateTemplateRequest {
  name?: string
  subject?: string
  content?: string
  is_default?: boolean
}

export const emailTemplateApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    request.get<{ list: EmailTemplate[]; total: number }>('/email-templates', { params }).then(r => r.data),

  create: (data: CreateTemplateRequest) =>
    request.post<EmailTemplate>('/email-templates', data),

  update: (id: number, data: UpdateTemplateRequest) =>
    request.put<EmailTemplate>(`/email-templates/${id}`, data),

  delete: (id: number) =>
    request.delete(`/email-templates/${id}`),
}
