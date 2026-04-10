import request from './request'

export interface SIPolicy {
  id: number
  city: string
  city_code: string
  year: number
  pension_base_min: number
  pension_base_max: number
  pension_rate: number
  pension_person_rate: number
  pension_company_rate: number
  medical_base_min: number
  medical_base_max: number
  medical_rate: number
  medical_person_rate: number
  medical_company_rate: number
  unemployment_rate: number
  unemployment_person_rate: number
  unemployment_company_rate: number
  maternity_rate: number
  injury_rate: number
  housing_fund_base_min: number
  housing_fund_base_max: number
  housing_fund_rate: number
  housing_fund_person_rate: number
  housing_fund_company_rate: number
  effective_date: string
  created_at: string
}

export interface SIRecord {
  id: number
  employee_id: number
  employee_name: string
  policy_id: number
  city: string
  salary_base: number
  start_month: string
  stop_month?: string
  status: 'active' | 'stopped'
  monthly_personal: number
  monthly_company: number
  created_at: string
}

export interface SICalculateResult {
  pension_personal: number
  pension_company: number
  medical_personal: number
  medical_company: number
  unemployment_personal: number
  unemployment_company: number
  maternity_company: number
  injury_company: number
  housing_fund_personal: number
  housing_fund_company: number
  total_personal: number
  total_company: number
  total: number
}

export interface EnrollPreview {
  employee_name: string
  salary_base: number
  calculation: SICalculateResult
}

export const siApi = {
  policies: (params?: { city?: string; year?: number }) =>
    request.get<SIPolicy[]>('/social-insurance/policies', { params }),

  calculate: (data: { city: string; year: number; salary_base: number }) =>
    request.post<SICalculateResult>('/social-insurance/calculate', data),

  records: (params?: { page?: number; employee_id?: number; status?: string }) =>
    request.get<{ list: SIRecord[]; total: number }>('/social-insurance/records', { params }),

  enrollPreview: (data: { employee_ids: number[]; policy_id: number; salary_base: number }) =>
    request.post<EnrollPreview[]>('/social-insurance/enroll/preview', data),

  enroll: (data: { employee_ids: number[]; policy_id: number; start_month: string; salary_base: number }) =>
    request.post<void>('/social-insurance/enroll', data),

  stop: (data: { record_ids: number[]; stop_month: string }) =>
    request.post<void>('/social-insurance/stop', data),

  changeHistory: (recordId: number) =>
    request.get<SIRecord[]>(`/social-insurance/records/${recordId}/history`),
}
