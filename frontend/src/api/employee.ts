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
  status: 'pending_review' | 'approved' | 'completed'
  resign_reason: string
  last_workday: string
  checklist: {
    items_returned: boolean
    handover_done: boolean
    final_settlement: boolean
  }
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
    request.post<void>(`/offboardings/${id}/approve`),

  completeOffboarding: (id: number) =>
    request.post<void>(`/offboardings/${id}/complete`),

  exportExcel: () =>
    request.get('/employees/export', { responseType: 'blob' }),

  getDashboard: () =>
    request.get<EmployeeDashboard>('/dashboard/employee-dashboard'),
}
