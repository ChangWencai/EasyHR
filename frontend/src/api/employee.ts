import request from './request'

export interface Employee {
  id: number
  name: string
  phone: string
  id_number: string
  position: string
  entry_date: string
  status: string
  salary?: number
  probation_salary?: number
  bank_card?: string
  emergency_contact?: string
  emergency_phone?: string
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
    request.get<RegistrationListResponse>('/registrations', { params }),

  create: (data: {
    employee_id?: number
    name: string
    department_id?: number
    position: string
    hire_date: string
  }) => request.post<Registration>('/registrations', data),

  delete: (id: number) => request.delete(`/registrations/${id}`),

  getDetail: (token: string) => request.get<RegistrationDetail>(`/registrations/${token}`),

  submit: (token: string, data: SubmitRegistrationData) =>
    request.post(`/registrations/${token}/submit`, data),
}

export const employeeApi = {
  list: (params: { page: number; page_size?: number; search?: string }) =>
    request.get<EmployeeListResponse>('/employees', { params }),

  get: (id: number) => request.get<Employee>(`/employees/${id}`),

  create: (data: Partial<Employee>) => request.post<Employee>('/employees', data),

  update: (id: number, data: Partial<Employee>) =>
    request.put<Employee>(`/employees/${id}`, data),

  delete: (id: number) => request.delete(`/employees/${id}`),

  invitations: (params?: { page: number; page_size?: number }) =>
    request.get<InvitationListResponse>('/invitations', { params }),

  createInvitation: (data: { name: string; phone: string }) =>
    request.post<{ invite_url: string }>('/invitations', data),

  cancelInvitation: (id: number) => request.delete(`/invitations/${id}`),

  offboardings: (params?: { page: number; page_size?: number; status?: string }) =>
    request.get<OffboardingListResponse>('/offboardings', { params }),

  approveOffboarding: (id: number) =>
    request.put<void>(`/offboardings/${id}/approve`),

  rejectOffboarding: (id: number, reason?: string) =>
    request.put<void>(`/offboardings/${id}/reject`, { reason }),

  completeOffboarding: (id: number) =>
    request.put<void>(`/offboardings/${id}/complete`),

  exportExcel: () =>
    request.get('/employees/export', { responseType: 'blob' }),

  getDashboard: () =>
    request.get<EmployeeDashboard>('/dashboard/employee-dashboard'),
}
