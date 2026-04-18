import request from '@/api/request'

export interface TodoItem {
  id: number
  org_id: number
  title: string
  type: string
  content?: string
  employee_id?: number
  employee_name?: string
  created_by: number
  creator_name: string
  deadline?: string
  is_time_limited: boolean
  urgency_status: 'normal' | 'overdue' | 'expired'
  status: 'pending' | 'completed' | 'terminated'
  source_type?: string
  source_id?: number
  is_pinned: boolean
  sort_order: number
  created_at: string
  updated_at: string
}

export interface CarouselItem {
  id: number
  org_id: number
  image_url: string
  link_url?: string
  sort_order: number
  active: boolean
  start_at?: string
  end_at?: string
}

export interface ListTodosResponse {
  items: TodoItem[]
  total: number
  page: number
  page_size: number
}

export interface ListTodosParams {
  keyword?: string
  start_date?: string
  end_date?: string
  status?: '' | 'pending' | 'completed' | 'terminated'
  page?: number
  page_size?: number
}

export function listTodos(params: ListTodosParams): Promise<ListTodosResponse> {
  return request.get('/todos', { params }).then((res) => res.data)
}

export function pinTodo(id: number, pinned: boolean): Promise<void> {
  return request.put(`/todos/${id}/pin`, { pinned }).then((res) => res.data)
}

export function listCarousels(): Promise<CarouselItem[]> {
  return request.get('/carousels').then((res) => res.data)
}

export function exportTodos(): void {
  const token = localStorage.getItem('token')
  const baseURL = import.meta.env.VITE_API_BASE_URL || '/api/v1'
  const url = `${baseURL}/todos/export`
  fetch(url, {
    headers: { Authorization: `Bearer ${token}` },
  })
    .then((res) => res.blob())
    .then((blob) => {
      const objectUrl = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = objectUrl
      link.download = `待办事项_${new Date().toISOString().slice(0, 10)}.xlsx`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.URL.revokeObjectURL(objectUrl)
    })
}

export interface InviteResult {
  url: string
}

export interface VerifyResult {
  valid: boolean
  expired: boolean
  title: string
  todo_id: number
}

export function inviteTodo(todoId: number): Promise<InviteResult> {
  return request.post(`/todos/${todoId}/invite`).then((res) => res.data)
}

export function terminateTodo(todoId: number): Promise<void> {
  return request.put(`/todos/${todoId}/terminate`).then((res) => res.data)
}

export function verifyInviteToken(token: string): Promise<VerifyResult> {
  return request.get(`/todos/invite/${token}`).then((res) => res.data)
}
