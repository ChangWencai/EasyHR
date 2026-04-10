import request from './request'

export interface TaxBracket {
  level: number
  start: number
  end: number
  rate: number
  quick_deduction: number
}

export interface TaxDeduction {
  id: number
  type: 'housing_loan' | 'housing_rent' | 'elderly_care' | 'children_education' | 'continuing_education' | 'serious_illness' | 'other'
  name: string
  amount: number
  year: number
  max_amount: number
  employee_id?: number
  employee_name?: string
  created_at: string
}

export interface TaxCalculateResult {
  gross_income: number
  tax_free_income: number
  taxable_income: number
  deduction_total: number
  applicable_bracket: TaxBracket
  quick_deduction: number
  tax_amount: number
  net_income: number
}

export interface TaxRecord {
  id: number
  employee_id: number
  employee_name: number
  year: number
  month: number
  gross_income: number
  tax_amount: number
  status: 'pending' | 'declared' | 'paid'
  declared_at?: string
  paid_at?: string
  created_at: string
}

export const taxApi = {
  brackets: (params?: { year?: number }) =>
    request.get<TaxBracket[]>('/tax/brackets', { params }),

  deductions: (params?: { year?: number; employee_id?: number }) =>
    request.get<TaxDeduction[]>('/tax/deductions', { params }),

  createDeduction: (data: Partial<TaxDeduction>) =>
    request.post<TaxDeduction>('/tax/deductions', data),

  updateDeduction: (id: number, data: Partial<TaxDeduction>) =>
    request.put<TaxDeduction>(`/tax/deductions/${id}`, data),

  deleteDeduction: (id: number) =>
    request.delete(`/tax/deductions/${id}`),

  calculate: (data: {
    gross_income: number
    year: number
    month: number
    deduction_ids?: number[]
  }) =>
    request.post<TaxCalculateResult>('/tax/calculate', data),

  records: (params?: { page?: number; year?: number }) =>
    request.get<{ list: TaxRecord[]; total: number }>('/tax/records', { params }),
}
