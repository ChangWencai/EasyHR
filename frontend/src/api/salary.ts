import request from './request'

export interface SalaryTemplateItem {
  id: number
  name: string
  type: 'earning' | 'deduction'
  category: string
  is_enabled: boolean
  is_default: boolean
  description?: string
}

export interface SalaryTemplate {
  id: number
  name: string
  items: SalaryTemplateItem[]
  created_at: string
  updated_at: string
}

export interface EmployeeSalaryItem {
  id: number
  template_item_id: number
  name: string
  type: 'earning' | 'deduction'
  amount: number
  is_active: boolean
}

export interface PayrollItem {
  id: number
  employee_id: number
  employee_name: string
  template_item_id: number
  name: string
  type: 'earning' | 'deduction'
  amount: number
}

export interface Payroll {
  id: number
  year: number
  month: number
  status: 'draft' | 'calculated' | 'confirmed' | 'paid'
  total_gross: number
  total_deduction: number
  total_net: number
  employee_count: number
  items: PayrollItem[]
  created_at: string
  confirmed_at?: string
  paid_at?: string
}

export interface PayrollListResponse {
  list: Payroll[]
  total: number
}

// ========== 薪资看板接口 ==========

export interface StatItem {
  label: string
  value: string
  trend_percent: string | null
  trend_direction: 'up' | 'down' | 'neutral'
}

export interface SalaryDashboardResponse {
  stats: StatItem[]
}

// ========== 调薪接口 ==========

export interface AdjustmentRequest {
  employee_id: number
  effective_month: string
  adjustment_type: 'base_salary' | 'allowance' | 'bonus' | 'year_end_bonus' | 'other'
  adjust_by: 'amount' | 'ratio'
  old_value: number
  new_value: number
}

export interface MassAdjustmentRequest {
  department_ids: number[]
  effective_month: string
  adjustment_type: 'base_salary' | 'allowance' | 'bonus' | 'year_end_bonus' | 'other'
  adjust_by: 'amount' | 'ratio'
  old_value: number
  new_value: number
}

export interface AdjustmentPreviewResponse {
  employee_count: number
  monthly_impact: number
  annual_impact: number
}

// ========== 绩效系数接口 ==========

export interface PerformanceCoefficient {
  employee_id: number
  coefficient: number
}

export const salaryApi = {
  template: () => request.get<SalaryTemplate>('/salary/template'),

  updateTemplate: (data: { items: { id: number; is_enabled: boolean }[] }) =>
    request.put<SalaryTemplate>('/salary/template', data),

  employeeItems: (employeeId: number, month: string) =>
    request.get<EmployeeSalaryItem[]>(`/salary/items/${employeeId}`, {
      params: { month },
    }),

  setEmployeeItems: (
    employeeId: number,
    month: string,
    items: { template_item_id: number; amount: number; is_active: boolean }[],
  ) =>
    request.put<EmployeeSalaryItem[]>(`/salary/items/${employeeId}`, {
      month,
      items,
    }),

  createPayroll: (data: { year: number; month: number; copy_from_month?: string }) =>
    request.post<Payroll>('/salary/payroll', data),

  calculatePayroll: (data: { year: number; month: number }) =>
    request.post<Payroll>('/salary/payroll/calculate', data),

  confirmPayroll: (data: { year: number; month: number }) =>
    request.put<Payroll>('/salary/payroll/confirm', data),

  list: (params: { year: number; month: number; page?: number; page_size?: number }) =>
    request.get<PayrollListResponse>('/salary/payroll', { params }),

  detail: (id: number) => request.get<Payroll>(`/salary/payroll/${id}`),

  recordPayment: (
    id: number,
    data: { method: string; paid_at: string },
  ) => request.put<Payroll>(`/salary/payroll/${id}/pay`, data),

  export: (year: number, month: number) =>
    request.get('/salary/payroll/export', {
      params: { year, month },
      responseType: 'blob',
    }),

  // 薪资看板
  getSalaryDashboard: (year: number, month: number) =>
    request.get<SalaryDashboardResponse>('/salary/dashboard', { params: { year, month } }),

  // 调薪
  createAdjustment: (data: AdjustmentRequest) =>
    request.post('/salary/adjustment', data),

  massAdjustment: (data: MassAdjustmentRequest) =>
    request.post('/salary/mass-adjustment', data),

  getAdjustmentList: (params: { effective_month?: string; page?: number; page_size?: number }) =>
    request.get('/salary/adjustments', { params }),

  // 绩效系数
  getPerformance: (year: number, month: number) =>
    request.get<PerformanceCoefficient[]>('/salary/performance', { params: { year, month } }),

  setPerformance: (data: { coefficients: { employee_id: number; coefficient: number }[]; year: number; month: number }) =>
    request.put('/salary/performance', data),
}
