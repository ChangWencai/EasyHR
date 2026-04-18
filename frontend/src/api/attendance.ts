import request from './request'

// === AttendanceRule ===

export interface AttendanceRule {
  id?: number
  mode: 'fixed' | 'scheduled' | 'free'
  work_days: number[]
  work_start: string
  work_end: string
  location: string
  clock_method: 'click' | 'photo'
  holidays: { date: string; name: string }[]
}

export interface Shift {
  id: number
  name: string
  work_start: string
  work_end: string
  work_date_offset: number
}

export interface Schedule {
  employee_id: number
  date: string
  shift_id: number | null
}

// === ClockRecord & ClockLive ===

export interface ClockLiveRecord {
  employee_id: number
  employee_name: string
  department_name: string
  clock_in_time: string
  clock_out_time: string
  status: 'normal' | 'late' | 'absent' | 'no_schedule' | 'not_clocked_in'
  shift_name?: string
}

export interface ClockLiveResponse {
  date: string
  records: ClockLiveRecord[]
  total: number
  page: number
  page_size: number
}

export interface LeaveStats {
  employee_id: number
  employee_name: string
  year_month: string
  leave_days: number
  business_days: number
  outside_days: number
  makeup_count: number
  shift_swap_count: number
  overtime_hours: number
  pending_days: number
  approved_days: number
}

export interface ApprovalRecord {
  id: number
  employee_id: number
  employee_name?: string
  approval_type: string
  type_name: string
  start_time: string
  end_time: string
  duration: number
  leave_type?: string
  reason: string
  status: 'draft' | 'pending' | 'approved' | 'rejected' | 'cancelled' | 'timeout'
  approver_id?: number
  approver_name?: string
  approved_at?: string
  rejected_at?: string
  rejected_note?: string
  cancelled_at?: string
  attachments: string[]
  created_at: string
}

export interface ApprovalListResponse {
  list: ApprovalRecord[]
  total: number
  page: number
}

export interface CreateApprovalRequest {
  approval_type: string
  start_time: string
  end_time: string
  reason?: string
  leave_type?: string
  attachments?: string[]
  cc_user_ids?: number[]
}

export const approvalApi = {
  list: (params: {
    status?: string
    approval_type?: string
    employee_id?: number
    page?: number
    page_size?: number
  }) => request.get<ApprovalListResponse>('/attendance/approvals', { params }),

  create: (data: CreateApprovalRequest) =>
    request.post<ApprovalRecord>('/attendance/approvals', data),

  approve: (id: number) =>
    request.put<ApprovalRecord>(`/attendance/approvals/${id}/approve`),

  reject: (id: number, note?: string) =>
    request.put<ApprovalRecord>(`/attendance/approvals/${id}/reject`, { note }),

  cancel: (id: number) =>
    request.put<ApprovalRecord>(`/attendance/approvals/${id}/cancel`),

  pendingCount: () =>
    request.get<{ pending_count: number }>('/attendance/approvals/pending-count'),
}

export const attendanceApi = {
  // 打卡规则
  getRule: () => request.get<AttendanceRule>('/attendance/rule'),

  saveRule: (data: Partial<AttendanceRule>) =>
    request.put<AttendanceRule>('/attendance/rule', data),

  // 班次
  listShifts: () => request.get<Shift[]>('/attendance/shifts'),

  createShift: (data: Omit<Shift, 'id'>) =>
    request.post<Shift>('/attendance/shifts', data),

  updateShift: (id: number, data: Omit<Shift, 'id'>) =>
    request.put<Shift>(`/attendance/shifts/${id}`, data),

  deleteShift: (id: number) =>
    request.delete(`/attendance/shifts/${id}`),

  // 排班
  listSchedules: (params: { start_date: string; end_date: string; employee_id?: number }) =>
    request.get<Schedule[]>('/attendance/schedules', { params }),

  batchUpsertSchedules: (data: { schedules: Schedule[] }) =>
    request.post('/attendance/schedules', data),

  // 打卡实况
  getClockLive: (params: { date: string; page?: number; page_size?: number }) =>
    request.get<ClockLiveResponse>('/attendance/clock-live', { params }),

  // 创建打卡记录（邀请点签）
  createClockRecord: (data: {
    employee_id: number
    clock_time: string
    clock_type: 'in' | 'out'
    photo_url?: string
  }) => request.post('/attendance/clock-records', data),

  // 假勤统计
  getLeaveStats: (params: { employee_id: number; year_month: string }) =>
    request.get<LeaveStats>('/attendance/leave-stats', { params }),

  // 手动修正假勤统计
  updateLeaveStats: (employeeId: number, yearMonth: string, data: Partial<LeaveStats>) =>
    request.put(`/attendance/leave-stats/${employeeId}?year_month=${yearMonth}`, data),
}
