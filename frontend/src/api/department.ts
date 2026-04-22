import request from './request'

export interface Department {
  id: number
  name: string
  parent_id: number | null
  sort_order: number
  employee_count?: number
}

export interface PaginatedResponse<T> {
  code: number
  message: string
  data: T
  meta: {
    total: number
    page: number
    page_size: number
  }
}

export interface TreeNode {
  id: number
  name: string
  type: 'department' | 'position' | 'employee'
  children?: TreeNode[]
  itemStyle?: Record<string, unknown>
  label?: Record<string, unknown>
}

export const departmentApi = {
  list: () => request.get<Department[]>('/departments').then(r => r.data),

  getTree: () => request.get<TreeNode[]>('/departments/tree'),

  searchTree: (keyword: string) =>
    request.get<TreeNode[]>('/departments/search', { params: { keyword } }),

  create: (data: { name: string; parent_id?: number | null; sort_order?: number }) =>
    request.post<Department>('/departments', data),

  update: (id: number, data: Partial<Pick<Department, 'name' | 'parent_id' | 'sort_order'>>) =>
    request.put<Department>(`/departments/${id}`, data),

  delete: (id: number) => request.delete(`/departments/${id}`),

  transferDelete: (id: number, data: { target_department_id: number; employee_ids: number[] }) =>
    request.delete<void>(`/departments/${id}/transfer`, { data }),
}
