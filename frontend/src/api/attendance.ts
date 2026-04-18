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
}
