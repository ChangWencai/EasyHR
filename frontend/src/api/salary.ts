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
  adjust_value: number
  old_value?: number
  new_value?: number
}

export interface MassAdjustmentRequest {
  department_ids: number[]
  effective_month: string
  adjustment_type: 'base_salary' | 'allowance' | 'bonus' | 'year_end_bonus' | 'other'
  adjust_by: 'amount' | 'ratio'
  adjust_value: number
  old_value?: number
  new_value?: number
}

export interface AdjustmentPreviewRequest {
  employee_id?: number
  department_ids?: number[]
  effective_month: string
  adjustment_type: 'base_salary' | 'allowance' | 'bonus' | 'year_end_bonus' | 'other'
  adjust_by: 'amount' | 'ratio'
  adjust_value: number
}

export interface AdjustmentPreviewResponse {
  employee_count: number
  department_count?: number
  monthly_impact: number
  annual_impact: number
  effective_month: string
}

// ========== 绩效系数接口 ==========

export interface PerformanceCoefficient {
  employee_id: number
  coefficient: number
}

// ========== 个税上传接口 ==========

export interface TaxUploadRow {
  row_number: number
  name: string
  employee_id: number
  employee_name: string
  tax_amount: number
  adjustment: number
}

export interface UnmatchedRow {
  row_number: number
  name: string
  reason: string
}

export interface TaxUploadResult {
  total_rows: number
  matched_count: number
  matched_rows: TaxUploadRow[]
  unmatched_rows: UnmatchedRow[]
}

// ========== 工资条发送接口 ==========

export interface SlipSendLog {
  id: number
  payroll_record_id: number
  employee_id: number
  channel: 'miniapp' | 'sms' | 'h5'
  status: 'pending' | 'sending' | 'sent' | 'failed'
  error_message?: string
  sent_at?: string
  confirmed_at?: string
  created_at: string
}

export const salaryApi = {
  template: () => request.get<SalaryTemplate>('/salary/template').then(r => r.data),

  updateTemplate: (data: { items: { id: number; is_enabled: boolean }[] }) =>
    request.put<SalaryTemplate>('/salary/template', data),

  employeeItems: (employeeId: number, month: string) =>
    request.get('/salary/items', {
      params: { employee_id: employeeId, month },
    }).then((r: any) => r.data || []),

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
    request.get<PayrollListResponse>('/salary/payroll', { params }).then(r => r.data),

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
    request.get<SalaryDashboardResponse>('/salary/dashboard', { params: { year, month } }).then(r => r.data),

  // 调薪
  createAdjustment: (data: AdjustmentRequest) =>
    request.post('/salary/adjustment', data),

  massAdjustment: (data: MassAdjustmentRequest) =>
    request.post('/salary/mass-adjustment', data),

  previewAdjustment: (data: AdjustmentPreviewRequest) =>
    request.post<AdjustmentPreviewResponse>('/salary/adjustment/preview', data),

  getAdjustmentList: (params: { effective_month?: string; page?: number; page_size?: number }) =>
    request.get('/salary/adjustments', { params }),

  // 绩效系数
  getPerformance: (year: number, month: number) =>
    request.get<PerformanceCoefficient[]>('/salary/performance', { params: { year, month } }).then(r => r.data),

  setPerformance: (data: { coefficients: { employee_id: number; coefficient: number }[]; year: number; month: number }) =>
    request.put('/salary/performance', data),

  // 个税上传
  uploadTax: (year: number, month: number, file: File) => {
    const form = new FormData()
    form.append('file', file)
    return request.post<TaxUploadResult>('/salary/tax-upload', form, { params: { year, month } }).then(r => r.data)
  },

  confirmTaxUpload: (data: { year: number; month: number; matched_rows: TaxUploadRow[] }) =>
    request.post('/salary/tax-upload/confirm', data),

  // 工资条发送
  sendSlipAll: (data: { year: number; month: number; employee_ids?: number[]; channel?: string }) =>
    request.post('/salary/slip/send-all', data),

  getSlipLogs: (params: { year?: number; month?: number; page?: number; page_size?: number }) =>
    request.get<{ logs: SlipSendLog[]; total: number }>('/salary/slip/logs', { params }).then(r => r.data),

  // 薪资列表
  getSalaryList: (params: { year: number; month: number; department_id?: number; keyword?: string; page?: number; page_size?: number }) =>
    request.get<PayrollListResponse>('/salary/payroll', { params }).then(r => r.data),

  // 解锁
  sendUnlockCode: (data: { phone: string }) =>
    request.post('/salary/unlock/send-code', data),

  unlockRecord: (data: { record_id: number; sms_code: string }) =>
    request.post('/salary/unlock', data),

  // 导出
  exportWithDetails: (year: number, month: number): Promise<Blob> =>
    request.get('/salary/payroll/export', {
      params: { year, month },
      responseType: 'blob',
    }).then(r => r.data as Blob),

  // 考勤导入
  importAttendance: (year: number, month: number, file: File) => {
    const form = new FormData()
    form.append('file', file)
    return request.post<AttendanceImportResult>('/salary/attendance/import', form, {
      params: { year, month },
    }).then(r => r.data)
  },
}

export interface AttendanceImportResult {
  matched_count: number
  error_rows?: Array<{ row_number: number; name: string; error: string }>
}

export interface SalaryRecord {
  id: number
  employee_id: number
  employee_name: string
  department_name: string
  gross_income: number
  total_deductions: number
  tax: number
  si_deduction: number
  net_income: number
  status: string
}
