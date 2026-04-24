import request from './request'

export interface Employee {
  id: number
  name: string
  phone: string
  id_card: string
  position: string
  position_id?: number | null
  department_id?: number | null
  department_name?: string
  hire_date: string
  status: string
  gender?: string
  email?: string
  salary?: number
  probation_salary?: number
  bank_account?: string
  bank_name?: string
  address?: string
  emergency_contact?: string
  emergency_phone?: string
  remark?: string
}

export interface Invitation {
  id: number
  name: string
  phone: string
  status: 'pending' | 'used' | 'expired' | 'cancelled'
  created_at: string
  invite_url: string
  expires_at?: string
}

export interface Offboarding {
  id: number
  employee_id: number
  employee_name: string
  status: 'pending' | 'approved' | 'rejected' | 'completed'
  type: string
  resignation_date: string
  reason: string
  checklist_items: Record<string, unknown>
  completed_at: string | null
  approved_by: number | null
  approved_at: string | null
  created_at: string
}

export interface EmployeeListResponse {
  list: Employee[]
  total: number
}

export interface EmployeeRosterItem {
  id: number
  name: string
  status: string
  position: string
  department_id: number | null
  department_name: string
  phone: string
  salary_amount: number
  years_of_service: string
  contract_expiry_days: number | null
}

export interface InvitationListResponse {
  list: Invitation[]
  total: number
}

export interface OffboardingListResponse {
  list: Offboarding[]
  total: number
}

export interface EmployeeDashboard {
  active_count: number
  joined_this_month: number
  left_this_month: number
  turnover_rate: number
}

export interface Registration {
  id: number
  employee_id: number | null
  token: string
  status: 'pending' | 'used' | 'expired'
  expires_at: string
  used_at: string | null
  created_at: string
  employee_name?: string
  department_name?: string
}

export interface RegistrationListResponse {
  list: Registration[]
  total: number
}

export interface RegistrationDetail {
  name: string
  department_id: number | null
  position: string
  hire_date: string
  status: string
}

export interface SubmitRegistrationData {
  phone?: string
  address?: string
  id_card?: string
  id_card_front_url?: string
  id_card_back_url?: string
  bank_account?: string
  bank_name?: string
  bank_card_front_url?: string
  bank_card_back_url?: string
  education_cert_url?: string
  emergency_contact?: string
  emergency_phone?: string
  emergency_relation?: string
}

export const registrationApi = {
  list: (params: { page: number; page_size: number; status?: string }) =>
    request.get<RegistrationListResponse>('/registrations', { params }).then((r) => r.data),

  create: (data: {
    employee_id?: number
    name: string
    department_id?: number
    position: string
    hire_date: string
  }) => request.post<Registration>('/registrations', data),

  delete: (id: number) => request.delete(`/registrations/${id}`),

  getDetail: (token: string) =>
    request.get<RegistrationDetail>(`/registrations/${token}`).then((r) => r.data),

  submit: (token: string, data: SubmitRegistrationData) =>
    request.post(`/registrations/${token}/submit`, data),

  sendSms: (token: string, phone: string) =>
    request.post('/registrations/send-sms', { token, phone }),
}

export const employeeApi = {
  list: (params: { page: number; page_size?: number; search?: string }) =>
    request.get<EmployeeListResponse>('/employees', { params }).then((r) => r.data),

  get: (id: number) => request.get<Employee>(`/employees/${id}`).then((r) => r.data),

  create: (data: Partial<Employee>) => request.post<Employee>('/employees', data),

  update: (id: number, data: Partial<Employee>) => request.put<Employee>(`/employees/${id}`, data),

  delete: (id: number) => request.delete(`/employees/${id}`),

  invitations: (params?: { page: number; page_size?: number }) =>
    request.get<InvitationListResponse>('/invitations', { params }).then((r) => r.data),

  createInvitation: (data: {
    name: string
    phone: string
    channel: string
    position?: string
    email_template_id?: number
  }) =>
    request.post<{ invite_url: string; channel: string }>('/invitations', data).then((r) => r.data),

  cancelInvitation: (id: number) => request.delete(`/invitations/${id}`),

  offboardings: (params?: { page: number; page_size?: number; status?: string }) =>
    request.get<OffboardingListResponse>('/offboardings', { params }).then((r) => r.data),

  approveOffboarding: (id: number) => request.put<void>(`/offboardings/${id}/approve`),

  rejectOffboarding: (id: number, reason?: string) =>
    request.put<void>(`/offboardings/${id}/reject`, { reason }),

  completeOffboarding: (id: number) => request.put<void>(`/offboardings/${id}/complete`),

  exportExcel: (params?: Record<string, unknown>): Promise<Blob> =>
    request.get('/employees/export', { params, responseType: 'blob' }).then((r) => r.data as Blob),

  getDashboard: () =>
    request.get<EmployeeDashboard>('/dashboard/employee-dashboard').then((r) => r.data),

  getRoster: (params: {
    page: number
    page_size: number
    search?: string
    status?: string
    department_id?: number
  }) =>
    request.get('/employees/roster', { params }).then((r: any) => ({
      list: r.data || [],
      total: r.meta?.total || 0,
    })),

  getSensitiveInfo: (id: number) =>
    request.post<Employee>(`/employees/${id}/sensitive`).then((r) => r.data),

  batchImportEmployees: (
    rows: Record<string, unknown>[],
  ): Promise<{ success: number; failed: number }> =>
    request
      .post('/employees/batch-import', { employees: rows })
      .then((r) => r.data as Promise<{ success: number; failed: number }>),

  confirmOnboarding: (id: number) => request.post(`/employees/${id}/confirm`),
}
