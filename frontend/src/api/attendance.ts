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

export interface MonthlyReportItem {
  employee_id: number
  employee_name: string
  department_name: string
  actual_days: number
  required_days: number
  overtime_hours: number
  absent_days: number
  leave_days: number
  business_days: number
  attendance_rate: number
  late_count: number
}

export interface MonthlyStats {
  total_actual_days: number
  total_required_days: number
  total_overtime_hours: number
  total_absent_days: number
}

export interface MonthlyReportResponse {
  year_month: string
  stats: MonthlyStats
  list: MonthlyReportItem[]
  total: number
  page: number
  page_size: number
}

export interface DailyRecord {
  date: string
  clock_in: string
  clock_out: string
  status: string
  is_holiday: boolean
  is_weekend: boolean
  symbol: string
}

export interface DailyRecordsResponse {
  employee_id: number
  year_month: string
  records: DailyRecord[]
}

// === Compliance Reports (COMP-05~COMP-08) ===

export interface OvertimeItem {
  employee_id: number
  employee_name: string
  department_name: string
  holiday_hours: number
  weekday_hours: number
  weekend_hours: number
  total_hours: number
}

export interface ComplianceOvertimeStats {
  total_holiday_hours: number
  total_weekday_hours: number
  total_weekend_hours: number
}

export interface ComplianceOvertimeResponse {
  year_month: string
  stats: ComplianceOvertimeStats
  list: OvertimeItem[]
  total: number
  page: number
  page_size: number
}

export interface LeaveItem {
  employee_id: number
  employee_name: string
  department_name: string
  annual_quota: number
  annual_used: number
  annual_left: number
  sick_days: number
  personal_days: number
}

export interface ComplianceLeaveStats {
  annual_quota_employee_count: number
  total_annual_used: number
  total_sick_days: number
  total_personal_days: number
}

export interface ComplianceLeaveResponse {
  year_month: string
  stats: ComplianceLeaveStats
  list: LeaveItem[]
  total: number
  page: number
  page_size: number
}

export interface AnomalyItem {
  employee_id: number
  employee_name: string
  department_name: string
  late_count: number
  early_leave_count: number
  absent_days: number
  anomaly_count: number
  is_anomaly: boolean
}

export interface ComplianceAnomalyStats {
  anomaly_employee_count: number
  total_late_count: number
  total_absent_days: number
}

export interface ComplianceAnomalyResponse {
  year_month: string
  stats: ComplianceAnomalyStats
  list: AnomalyItem[]
  total: number
  page: number
  page_size: number
}

export interface MonthlyComplianceItem {
  employee_id: number
  employee_name: string
  department_name: string
  required_days: number
  actual_days: number
  late_count: number
  early_leave_count: number
  absent_days: number
  overtime_hours: number
  annual_leave_days: number
  sick_leave_days: number
  personal_leave_days: number
  is_anomaly: boolean
}

export interface ComplianceMonthlyStats {
  total_required_days: number
  total_actual_days: number
  total_overtime_hours: number
  total_absent_days: number
  total_anomaly_count: number
}

export interface ComplianceMonthlyResponse {
  year_month: string
  stats: ComplianceMonthlyStats
  list: MonthlyComplianceItem[]
  total: number
  page: number
  page_size: number
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

  // 出勤月报
  getMonthlyReport: (params: { year_month: string; page?: number; page_size?: number }) =>
    request.get<MonthlyReportResponse>('/attendance/monthly', { params }),

  exportMonthlyExcel: (params: { year_month: string }) =>
    request.get('/attendance/monthly/export', { params, responseType: 'blob' }),

  getDailyRecords: (params: { employee_id: number; year_month: string }) =>
    request.get<DailyRecordsResponse>('/attendance/daily-records', { params }),

  // 合规报表 - 加班统计
  getComplianceOvertime: (params: { year_month: string; dept_ids?: string; page?: number; page_size?: number }) =>
    request.get<ComplianceOvertimeResponse>('/attendance/compliance/overtime', { params }),

  // 合规报表 - 请假合规
  getComplianceLeave: (params: { year_month: string; dept_ids?: string; page?: number; page_size?: number }) =>
    request.get<ComplianceLeaveResponse>('/attendance/compliance/leave', { params }),

  // 合规报表 - 出勤异常
  getComplianceAnomaly: (params: { year_month: string; dept_ids?: string; page?: number; page_size?: number }) =>
    request.get<ComplianceAnomalyResponse>('/attendance/compliance/anomaly', { params }),

  // 合规报表 - 月度汇总
  getComplianceMonthly: (params: { year_month: string; dept_ids?: string; page?: number; page_size?: number }) =>
    request.get<ComplianceMonthlyResponse>('/attendance/compliance/monthly', { params }),

  // 合规报表 - 月度汇总导出
  exportComplianceMonthly: (params: { year_month: string; dept_ids?: string }) =>
    request.get('/attendance/compliance/monthly/export', { params, responseType: 'blob' }),
}
