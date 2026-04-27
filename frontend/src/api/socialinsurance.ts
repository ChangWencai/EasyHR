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
  city_code: number
  city_name: string
  base_amount: number       // 后端字段: base_amount
  start_month: string
  end_month?: string
  status: 'active' | 'stopped'
  monthly_personal: number
  monthly_company: number
  created_at: string
}

export interface SIDashboardData {
  stats: SIStatItem[]
  overdue_items: OverdueItem[]
}

export interface SIStatItem {
  label: string
  value: string
  trend_percent?: string
  trend_direction: string
}

export interface OverdueItem {
  id: number
  employee_id: number
  employee_name: string
  city: string
  year_month: string
  amount: string
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
  policies: (params?: { city_code?: number; year?: number }) =>
    request.get<SIPolicy[]>('/social-insurance/policies', { params }).then(r => r.data),

  calculate: (data: { city_code: number; year: number; salary: number }) =>
    request.post<SICalculateResult>('/social-insurance/calculate', data),

  dashboard: () =>
    request.get<SIDashboardData>('/social-insurance/dashboard').then(r => r.data),

  records: (params?: { page?: number; page_size?: number; employee_id?: number; status?: string }) =>
    request.get<{ list: SIRecord[]; total: number }>('/social-insurance/records', { params }).then(r => r.data),

  recordDetail: (recordId: number) =>
    request.get<SIRecord>(`/social-insurance/monthly-records/${recordId}`).then(r => r.data),

  changeHistory: (recordId: number) =>
    request.get<SIRecord[]>(`/social-insurance/records/${recordId}/history`).then(r => r.data),

  enrollPreview: (data: { employee_ids: number[]; policy_id: number; salary_base: number }) =>
    request.post<EnrollPreview[]>('/social-insurance/enroll/preview', data).then(r => r.data),

  enroll: (data: { employee_ids: number[]; policy_id: number; start_month: string; salary_base: number }) =>
    request.post<void>('/social-insurance/enroll', data),

  enrollSingle: (data: { employee_id: number; city_code: number; si_base: number; start_year_month?: string }) =>
    request.post<void>('/social-insurance/enroll/single', data),

  stop: (data: { record_ids: number[]; stop_month: string }) =>
    request.post<void>('/social-insurance/stop', data),

  stopSingle: (data: { employee_id: number; stop_year_month: string; reason: string }) =>
    request.post<void>('/social-insurance/stop/single', data),

  exportRecords: (params?: { employee_id?: number; status?: string; start_month?: string; end_month?: string }) =>
    request.get('/social-insurance/records/export', { params, responseType: 'blob' }),
}
