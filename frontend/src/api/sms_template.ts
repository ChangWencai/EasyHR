import request from './request'

export interface SmsTemplate {
  id: number
  org_id: number
  name: string
  scene: string
  template_code: string
  content: string
  is_default: boolean
  created_by: number
  created_at: string
  updated_by?: number
  updated_at?: string
}

export interface CreateSmsTemplateRequest {
  name: string
  scene: string
  template_code: string
  content: string
  is_default?: boolean
}

export interface UpdateSmsTemplateRequest {
  name?: string
  scene?: string
  template_code?: string
  content?: string
  is_default?: boolean
}

export const sceneLabels: Record<string, string> = {
  verification_code: '验证码',
  contract_sign: '合同签署',
  registration: '入职邀请',
  salary_slip: '工资条通知',
  salary_unlock: '薪资解锁',
}

export const smsTemplateApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    request.get<{ list: SmsTemplate[]; total: number }>('/sms-templates', { params }).then(r => r.data),

  create: (data: CreateSmsTemplateRequest) =>
    request.post<SmsTemplate>('/sms-templates', data),

  update: (id: number, data: UpdateSmsTemplateRequest) =>
    request.put<SmsTemplate>(`/sms-templates/${id}`, data),

  delete: (id: number) =>
    request.delete(`/sms-templates/${id}`),
}
