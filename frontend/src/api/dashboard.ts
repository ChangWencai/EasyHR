import request from '@/api/request'

export interface TodoItem {
  type: string
  title: string
  count: number
  deadline?: string
  priority: number
}

export interface DashboardOverview {
  employee_count: number
  joined_this_month: number
  left_this_month: number
  social_insurance_total: string
  payroll_total: string
}

export interface DashboardResponse {
  todos: TodoItem[]
  overview: DashboardOverview
}

export function fetchDashboard(): Promise<DashboardResponse> {
  return request.get('/dashboard').then((res) => res.data)
}

export interface RingChartStats {
  completed: number
  pending: number
  total: number
  percent: number
}

export interface RingChartResponse {
  stats: RingChartStats
}

export function fetchTodoRingStats(type: 'all' | 'time-limited'): Promise<RingChartResponse> {
  const endpoint = type === 'all' ? '/dashboard/todo-stats' : '/dashboard/time-limited-stats'
  return request.get(endpoint).then((res) => res.data)
}
